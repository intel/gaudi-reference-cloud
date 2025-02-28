package loadbalancer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/client-go/util/retry"
)

func (lb *Service) create(ctx context.Context, loadbalancer *pb.LoadBalancerPrivate) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerService.create").WithValues(logkeys.CloudAccountId, loadbalancer.Metadata.CloudAccountId,
		logkeys.ResourceId, loadbalancer.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.LoadBalancer, loadbalancer)

	// Validate input.
	if loadbalancer.Metadata == nil {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	if loadbalancer.Spec == nil {
		return status.Error(codes.InvalidArgument, "missing spec")
	}

	if loadbalancer.Spec.Security == nil {
		return status.Error(codes.InvalidArgument, "missing security configuration")
	}

	if len(loadbalancer.Spec.Listeners) == 0 {
		return status.Error(codes.InvalidArgument, "missing listeners")
	}

	// Calculate resourceId if not provided.
	if loadbalancer.Metadata.ResourceId == "" {
		resourceId, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		loadbalancer.Metadata.ResourceId = resourceId.String()
	}

	// Validate resourceId.
	if _, err := uuid.Parse(loadbalancer.Metadata.ResourceId); err != nil {
		return status.Error(codes.InvalidArgument, "invalid resourceId")
	}

	// Calculate name if not provided.
	if loadbalancer.Metadata.Name == "" {
		loadbalancer.Metadata.Name = loadbalancer.Metadata.ResourceId
	}
	name := loadbalancer.Metadata.Name

	if err := validateLoadBalancerName(name); err != nil {
		return err
	}

	if err := lb.baseValidations(ctx, loadbalancer); err != nil {
		return err
	}

	// Ensure that a load balancer with the same name does not already exist in the database.
	_, err := lb.get(ctx, loadbalancer.Metadata.CloudAccountId, "name", name)
	if err == nil {
		return status.Errorf(codes.AlreadyExists, "load balancer with name %s already exists", name)
	}
	if status.Code(err) != codes.NotFound {
		return err
	}

	// Insert into database in a single transaction.
	tx, err := lb.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return status.Errorf(codes.Internal, "BeginTx: %v", err)
	}
	defer tx.Rollback()

	// Flatten loadbalancer into columns.
	flattened, err := lb.sqlTransformer.Flatten(ctx, loadbalancer)
	if err != nil {
		return err
	}

	// Check for quota limits on the requested account. If the account is over quota an error
	// will be returned and the LB will not be created.
	if err := lb.validateQuotaLimits(ctx, loadbalancer.Metadata.CloudAccountId); err != nil {
		return err
	}

	// Insert into database.
	query := fmt.Sprintf(`insert into loadbalancer (resource_id, cloud_account_id, name, %s) values ($1, $2, $3, %s)`,
		flattened.GetColumnsString(), flattened.GetInsertValuesString(4))
	args := append([]any{loadbalancer.Metadata.ResourceId, loadbalancer.Metadata.CloudAccountId, name}, flattened.Values...)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == common.KErrUniqueViolation {
			return status.Error(codes.AlreadyExists, "insert: load balancer "+name+" already exists")
		}
		return status.Errorf(codes.Internal, "insert: %v", err)
	}
	if err := tx.Commit(); err != nil {
		return status.Errorf(codes.Internal, "commit: %v", err)
	}
	return nil
}

func (lb *Service) get(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*pb.LoadBalancer, error) {
	if argName == "resource_id" {
		// Validate the resource id is a valid guid
		if _, err := uuid.Parse(arg.(string)); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid resourceId")
		}
	}

	rows, err := lb.selectLoadBalancer(ctx, cloudAccountId, argName, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return lb.rowToLoadBalancer(ctx, rows)
}

// Update an load balancer record using the user-provided updateFunc to update the load balancer.
// This uses optimistic concurrency control to ensure that the record has not been updated between the select and update.
// Additionally, if the caller provides a resource version, optimistic concurrency control can be extended to
// previous get or search calls.
func (lb *Service) update(
	ctx context.Context,
	cloudAccountId string,
	resourceId string,
	name string,
	resourceVersion string,
	updateFunc func(*pb.LoadBalancerPrivate) error,
	isStatusUpdate bool) error {
	argName, arg, err := common.ResourceUniqueColumnAndValue(resourceId, name)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	if argName == "resource_id" {
		// Validate the resource id is a valid guid
		if _, err := uuid.Parse(resourceId); err != nil {
			return status.Error(codes.InvalidArgument, "invalid resourceId")
		}
	}
	query := fmt.Sprintf(`
		select %s
		from   loadbalancer
		where  cloud_account_id = $1
			and  %s = $2
			and  deleted_timestamp = $3
	`, lb.sqlTransformer.ColumnsForFromRow(), argName)

	// Retry on conflict if caller did not provide resourceVersion.
	isRetryable := func(err error) bool {
		return resourceVersion == "" && status.Code(err) == codes.FailedPrecondition
	}

	err = retry.OnError(retry.DefaultRetry, isRetryable, func() error {
		rows, err := lb.db.QueryContext(ctx, query, cloudAccountId, arg, common.TimestampInfinityStr)
		if err != nil {
			return err
		}
		defer rows.Close()
		if !rows.Next() {
			return status.Error(codes.NotFound, "resource not found")
		}
		loadbalancer, err := lb.sqlTransformer.FromRow(ctx, rows)
		if err != nil {
			return err
		}
		metadata := loadbalancer.Metadata
		// If resource version was provided, ensure that stored version matches.
		if resourceVersion != "" && resourceVersion != metadata.ResourceVersion {
			return status.Error(codes.FailedPrecondition, "stored resource version does not match requested resource version")
		}

		// Update load balancer object.
		if err := updateFunc(loadbalancer); err != nil {
			return err
		}

		// Only validate resources that are active, not deleted, and it not a status update.
		if loadbalancer.Metadata.DeletionTimestamp == nil && !isStatusUpdate {
			// Validate the updated resource
			if err := lb.baseValidations(ctx, loadbalancer); err != nil {
				return err
			}
		}

		// Flatten loadbalancer into columns.
		flattened, err := lb.sqlTransformer.Flatten(ctx, loadbalancer)
		if err != nil {
			return err
		}

		args := append([]any{metadata.CloudAccountId, metadata.ResourceId, metadata.ResourceVersion}, flattened.Values...)

		// Update database.
		updateQuery := fmt.Sprintf(`
		update loadbalancer
		set    resource_version = nextval('loadbalancer_resource_version_seq'),
			   %s
		where  cloud_account_id = $1
		and    resource_id = $2
		and    resource_version = $3
		`, flattened.GetUpdateSetString(4))

		sqlResult, err := lb.db.ExecContext(ctx, updateQuery, args...)
		if err != nil {
			return err
		}
		rowsAffected, err := sqlResult.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected < 1 {
			return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
		}
		return nil
	})
	if err != nil {
		st, _ := status.FromError(err)
		return status.Error(st.Code(), "update: "+err.Error())
	}
	return err
}

func (lb *Service) deleteLoadBalancer(ctx context.Context, metadata *pb.LoadBalancerMetadataReference) (*emptypb.Empty, error) {
	if metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	cloudAccountId := metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	if err := lb.deleteLoadBalancerInternal(ctx, cloudAccountId, metadata.GetResourceId(), metadata.GetName(), metadata.ResourceVersion); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (lb *Service) deleteLoadBalancerInternal(ctx context.Context, cloudAccountId string, resourceId string, name string, resourceVersion string) error {
	updateFunc := func(loadbalancer *pb.LoadBalancerPrivate) error {
		if loadbalancer.Metadata.DeletionTimestamp == nil {
			loadbalancer.Metadata.DeletionTimestamp = timestamppb.Now()
		}
		return nil
	}
	if err := lb.update(ctx, cloudAccountId, resourceId, name, resourceVersion, updateFunc, false); err != nil {
		return err
	}
	return nil
}

// Caller must close the returned sql.Rows.
func (lb *Service) selectLoadBalancer(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		select %s
		from  loadbalancer
		where  cloud_account_id = $1
		  and  %s = $2
		  and  deleted_timestamp = $3
	`, lb.sqlTransformer.ColumnsForFromRow(), argName)
	rows, err := lb.db.QueryContext(ctx, query, cloudAccountId, arg, common.TimestampInfinityStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "selectLoadBalancer: %v", err)
	}
	if !rows.Next() {
		defer rows.Close()
		return nil, status.Error(codes.NotFound, "resource not found")
	}
	return rows, nil
}

func (lb *Service) baseValidations(ctx context.Context, loadbalancer *pb.LoadBalancerPrivate) error {
	if err := utils.ValidateLabels(loadbalancer.Metadata.Labels); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if len(loadbalancer.Spec.Listeners) == 0 {
		return status.Error(codes.InvalidArgument, "listeners are required")
	}

	if len(loadbalancer.Spec.Security.Sourceips) == 0 {
		return status.Error(codes.InvalidArgument, "source ips are required")
	}

	if err := lb.validateSourceIPsCount(loadbalancer.Metadata.CloudAccountId, len(loadbalancer.Spec.Security.Sourceips)); err != nil {
		return err
	}

	if err := validateSourceIPs(loadbalancer.Spec.Security.Sourceips); err != nil {
		return err
	}

	if err := lb.validateListenerCount(loadbalancer.Metadata.CloudAccountId, len(loadbalancer.Spec.Listeners)); err != nil {
		return err
	}

	if err := lb.validateListenerPoolMembers(ctx, loadbalancer.Metadata.CloudAccountId, loadbalancer.Spec); err != nil {
		return err
	}

	configuredListenerPorts := make(map[int32]int)

	// Iterate over each listener validating
	for _, listener := range loadbalancer.Spec.Listeners {

		if err := validateInstanceSelector(listener); err != nil {
			return err
		}

		if err := validateLBPort(listener.Port); err != nil {
			return err
		}

		// check if the port is already in use, if so return error
		if _, found := configuredListenerPorts[listener.Port]; found {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("port %d is already in use by another listener", listener.Port))
		}

		// add the port to the list of configured port
		configuredListenerPorts[listener.Port] = 0

		// Ensure that the healthcheck monitor type is configured. If not defined the by the user
		// then default to "TCP".
		if err := validateLBMonitoringType(listener.Pool.Monitor); err != nil {
			return err
		}

		// Ensure that the load balancing type is configured. If not defined the by the user
		// then default to "round-robin".
		if err := validateLBMModeType(listener.Pool.LoadBalancingMode); err != nil {
			return err
		}
	}
	return nil
}

// Read a database row into a public LoadBalancer. Used for public APIs.
func (lb *Service) rowToLoadBalancer(ctx context.Context, rows *sql.Rows) (*pb.LoadBalancer, error) {
	log := log.FromContext(ctx).WithName("LoadBalancerService.rowToLoadBalancer")
	loadBalancerPrivate, err := lb.sqlTransformer.FromRow(ctx, rows)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "rowToLoadBalancer: %v", err)
	}
	loadbalancer := &pb.LoadBalancer{}
	if err := lb.pbConverter.Transcode(loadBalancerPrivate, loadbalancer); err != nil {
		return nil, status.Errorf(codes.Internal, "rowToLoadBalancer: %v", err)
	}
	log.V(9).Info("Read from database", logkeys.LoadBalancer, loadbalancer)
	return loadbalancer, nil
}
