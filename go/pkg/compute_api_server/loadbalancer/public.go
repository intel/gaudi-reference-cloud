package loadbalancer

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/instance"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/pbconvert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedLoadBalancerServiceServer
	pb.UnimplementedLoadBalancerPrivateServiceServer
	db             *sql.DB
	cfg            config.Config
	pbConverter    *pbconvert.PbConverter
	sqlTransformer *LoadBalancerSqlTransformer

	instanceService *instance.InstanceService
}

func NewLoadBalancerService(db *sql.DB, cfg config.Config, instanceService *instance.InstanceService) (*Service, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &Service{
		db:              db,
		cfg:             cfg,
		pbConverter:     pbconvert.NewPbConverter(),
		sqlTransformer:  NewLoadBalancerSqlTransformer(),
		instanceService: instanceService,
	}, nil
}

// Launch a load balancer.
// Public API.
func (lb *Service) Create(ctx context.Context, req *pb.LoadBalancerCreateRequest) (*pb.LoadBalancer, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerService.Create").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.LoadBalancer, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Spec == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec")
		}

		cloudAccountId := req.Metadata.CloudAccountId

		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		// Transcode LoadBalancerSpec to LoadBalancerSpecPrivate.
		// Unmatched fields will remain at their default.
		loadbalancerSpecPrivate := &pb.LoadBalancerSpecPrivate{}
		if err := lb.pbConverter.Transcode(req.Spec, loadbalancerSpecPrivate); err != nil {
			return nil, status.Errorf(codes.Internal, "unable to transcode load balancer spec: %v", err)
		}

		ts := timestamppb.Now()

		loadbalancer := &pb.LoadBalancerPrivate{
			Metadata: &pb.LoadBalancerMetadataPrivate{
				CloudAccountId:    cloudAccountId,
				Name:              req.Metadata.Name,
				Labels:            req.Metadata.Labels,
				CreationTimestamp: ts,
			},
			Spec: loadbalancerSpecPrivate,
			Status: &pb.LoadBalancerStatusPrivate{
				Conditions: nil,
				Listeners:  nil,
				State:      "Pending",
				Vip:        "",
			},
		}

		if err := lb.create(ctx, loadbalancer); err != nil {
			return nil, err
		}

		// Query database and return response.
		return lb.get(ctx, cloudAccountId, "resource_id", loadbalancer.Metadata.ResourceId)
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, utils.SanitizeError(err)
}

// Public API.
func (lb *Service) Get(ctx context.Context, req *pb.LoadBalancerGetRequest) (*pb.LoadBalancer, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerService.Get").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.LoadBalancer, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		argName, arg, err := common.ResourceUniqueColumnAndValue(req.Metadata.GetResourceId(), req.Metadata.GetName())
		if err != nil {
			return nil, err
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		return lb.get(ctx, cloudAccountId, argName, arg)
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, utils.SanitizeError(err)
}

// Public API.
func (lb *Service) Update(ctx context.Context, req *pb.LoadBalancerUpdateRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerService.Update").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Spec == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec")
		}
		if req.Spec.Security == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec.security")
		}
		if req.Spec.Listeners == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec.listeners")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		updateFunc := func(loadbalancer *pb.LoadBalancerPrivate) error {
			loadbalancer.Spec.Listeners = req.Spec.Listeners
			loadbalancer.Spec.Security = req.Spec.Security
			return nil
		}
		if err := lb.update(ctx, cloudAccountId, req.Metadata.GetResourceId(), req.Metadata.GetName(), req.Metadata.ResourceVersion, updateFunc, false); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, utils.SanitizeError(err)
}

// Allow update of:
//   - Status
//
// Private API.
func (lb *Service) UpdateStatus(ctx context.Context, req *pb.LoadBalancerUpdateStatusRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerService.UpdateStatus").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Status == nil {
			return nil, status.Error(codes.InvalidArgument, "missing status")
		}
		updateFunc := func(loadbalancer *pb.LoadBalancerPrivate) error {
			loadbalancer.Status = req.Status
			return nil
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		if err := lb.update(ctx, cloudAccountId, req.Metadata.ResourceId, "", req.Metadata.ResourceVersion, updateFunc, true); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

// Delete indicates a public API user's request to delete a load balancer. It sets the deletionTimestamp of the record's "value" field.
// Public API.
func (lb *Service) Delete(ctx context.Context, req *pb.LoadBalancerDeleteRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerService.Delete").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()
	log.Info("Request", logkeys.Request, req)

	resp, err := lb.deleteLoadBalancer(ctx, req.Metadata)
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	}
	log.Info("Response", logkeys.Response, resp)

	return resp, utils.SanitizeError(err)
}

// RemoveFinalizer marks a record as hidden from users and controllers. It sets the record's deleted_timestamp.
// Private API.
func (lb *Service) RemoveFinalizer(ctx context.Context, req *pb.LoadBalancerRemoveFinalizerRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerService.RemoveFinalizer").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		query := `
			update loadbalancer
			set    deleted_timestamp = current_timestamp,
			       resource_version = nextval('loadbalancer_resource_version_seq')
			where  cloud_account_id = $1
			  and  resource_id = $2
			  and  deleted_timestamp = $3
		`
		if _, err := lb.db.ExecContext(ctx, query, req.Metadata.CloudAccountId, req.Metadata.ResourceId, common.TimestampInfinityStr); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

// Public API.
func (lb *Service) Search(ctx context.Context, req *pb.LoadBalancerSearchRequest) (*pb.LoadBalancerSearchResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadBalancerService.Search").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.LoadBalancerSearchResponse, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		if err := utils.ValidateLabels(req.Metadata.Labels); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		flattenedObject := protodb.Flattened{
			Columns: []string{"cloud_account_id", "deleted_timestamp"},
			Values:  []any{cloudAccountId, common.TimestampInfinityStr},
		}

		labels := req.Metadata.Labels
		for key, value := range labels {
			column := fmt.Sprintf("value->'metadata'->'labels'->>'%s'", key)
			flattenedObject.Add(column, value)
		}

		whereString := flattenedObject.GetWhereString(1)

		query := fmt.Sprintf(`
			select %s
			from   loadbalancer
			where  %s
			order by name
		`, lb.sqlTransformer.ColumnsForFromRow(), whereString)

		args := flattenedObject.Values
		rows, err := lb.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var items []*pb.LoadBalancer
		for rows.Next() {
			item, err := lb.rowToLoadBalancer(ctx, rows)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		resp := &pb.LoadBalancerSearchResponse{
			Items: items,
		}
		return resp, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, utils.SanitizeError(err)
}

// This function verifies that the combined total of existing and requested load balancer does not surpass the quota limit for the cloud account id.
//
// Parameters:
// - ctx: A context.Context instance for carrying deadlines, cancellation signals, and other request-scoped values.
// - cloudAccountId: A string representing the unique identifier of the user.
func (lb *Service) validateQuotaLimits(ctx context.Context, cloudAccountId string) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("LoadbalancerService.validateQuotaLimits").WithValues(logkeys.CloudAccountId, cloudAccountId).Start()
	defer span.End()
	log.Info("validateQuotaLimits", logkeys.CloudAccountId, cloudAccountId)
	loadbalancers, err := lb.Search(ctx, &pb.LoadBalancerSearchRequest{
		Metadata: &pb.LoadBalancerMetadataSearch{
			CloudAccountId: cloudAccountId,
		},
	})
	if err != nil {
		log.Error(err, "error in search call")
		return status.Error(codes.Internal, "error in quota processing")
	}

	// Set to the default, then look if a cloud account overrides.
	allowedQuota := lb.cfg.CloudAccountQuota.DefaultLoadBalancerQuota

	if cloudAccountQuota, found := lb.cfg.CloudAccountQuota.CloudAccountIDQuotas[cloudAccountId]; found {
		allowedQuota = cloudAccountQuota.LoadbalancerQuota
	}
	log.Info("validateQuotaLimits", logkeys.AllowedQuota, allowedQuota, logkeys.CurrentLoadBalancerCount, len(loadbalancers.Items))
	if allowedQuota <= len(loadbalancers.Items) {
		return status.Error(codes.OutOfRange, "Your account has reached the maximum allowed load balancer limit")
	}

	return nil
}

// Validate the number of listeners configure does not exceed the max limit.
func (lb *Service) validateListenerCount(cloudAccountId string, numListeners int) error {

	allowedListeners := lb.cfg.CloudAccountQuota.DefaultLoadBalancerListenerQuota

	// Lookup to see if cloudaccount has defined specific configuration for loadbalancer listener quota.
	if cloudAccountQuota, found := lb.cfg.CloudAccountQuota.CloudAccountIDQuotas[cloudAccountId]; found {
		allowedListeners = cloudAccountQuota.LoadbalancerListenerQuota
	}

	if numListeners > allowedListeners {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("too many listeners configured, max allowed is %d", allowedListeners))
	}
	return nil
}

// Validate the number of sourceIPs configure does not exceed the max limit.
func (lb *Service) validateSourceIPsCount(cloudAccountId string, numSourceIPs int) error {

	allowedSourceIPs := lb.cfg.CloudAccountQuota.DefaultLoadBalancerSourceIPQuota

	// Lookup to see if cloudaccount has defined specific configuration for loadbalancer source IP quota.
	if cloudAccountQuota, found := lb.cfg.CloudAccountQuota.CloudAccountIDQuotas[cloudAccountId]; found {
		allowedSourceIPs = cloudAccountQuota.LoadbalancerSourceIPQuota
	}

	if numSourceIPs > allowedSourceIPs {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("too many source IPs configured, max allowed is %d", allowedSourceIPs))
	}

	return nil
}

func (lb *Service) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("LoadBalancerService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (lb *Service) PingPrivate(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("LoadBalancerService.PingPrivate")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}
