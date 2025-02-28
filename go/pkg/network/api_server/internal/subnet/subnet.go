// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package subnet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/vpc"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/pbconvert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/transformer"
	networkutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/utils"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/client-go/util/retry"
)

const (
	subnetIdPrefix = "subnet-"
	subnetIdKey    = "resource_id"
	subnetNameKey  = "name"
)

type SubnetService struct {
	pb.UnimplementedSubnetServiceServer
	pb.UnimplementedSubnetPrivateServiceServer
	db                        *sql.DB
	cfg                       config.Config
	cloudAccountServiceClient pb.CloudAccountServiceClient
	sqlTransformer            *SubnetSQLTransformer
	pbConverter               *pbconvert.PbConverter
	vpcService                *vpc.VPCService
	availabilityZones         []string
}

func NewSubnetService(
	db *sql.DB,
	config config.Config,
	cloudAccountServiceClient pb.CloudAccountServiceClient,
	vpcService *vpc.VPCService,
	availabilityZones []string,
) (*SubnetService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &SubnetService{
		db:                        db,
		cfg:                       config,
		cloudAccountServiceClient: cloudAccountServiceClient,
		sqlTransformer:            NewSubnetSQLTransformer(),
		pbConverter:               pbconvert.NewPbConverter(),
		vpcService:                vpcService,
		availabilityZones:         availabilityZones,
	}, nil
}

func (s *SubnetService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("SubnetService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (s *SubnetService) PingPrivate(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("SubnetService.PingPrivate")
	log.Info("PingPrivate")
	return &emptypb.Empty{}, nil
}

// Private API: Create a new Subnet
func (s *SubnetService) CreatePrivate(ctx context.Context, req *pb.SubnetCreatePrivateRequest) (*pb.SubnetPrivate, error) {
	return nil, nil
}

// Public API: Create a new Subnet
func (s *SubnetService) Create(ctx context.Context, req *pb.SubnetCreateRequest) (*pb.VPCSubnet, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("SubnetService.Create").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VPCSubnet, error) {
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

		// Validate
		if req.Spec.AvailabilityZone == "" {
			return nil, status.Error(codes.InvalidArgument, "missing availabilityZone")
		}

		if !slices.Contains(s.availabilityZones, req.Spec.AvailabilityZone) {
			return nil, status.Error(codes.InvalidArgument, "invalid availabilityZone")
		}

		// Transcode SubnetSpec to SubnetSpecPrivate.
		// Unmatched fields will remain at their default.
		subnetSpecPrivate := &pb.SubnetSpecPrivate{}
		if err := s.pbConverter.Transcode(req.Spec, subnetSpecPrivate); err != nil {
			return nil, fmt.Errorf("unable to transcode subnet spec: %w", err)
		}

		subnet := &pb.SubnetPrivate{
			Metadata: &pb.SubnetMetadataPrivate{
				CloudAccountId: cloudAccountId,
				Name:           req.Metadata.Name,
				Labels:         req.Metadata.Labels,
			},
			Spec: subnetSpecPrivate,
			Status: &pb.SubnetStatusPrivate{
				Phase:   pb.SubnetPhase_SubnetPhase_Provisioning,
				Message: "Subnet is provisioning",
			},
		}

		if err := s.create(ctx, subnet); err != nil {
			return nil, err
		}

		// Query database and return response.
		return s.get(ctx, cloudAccountId, subnetIdKey, subnet.Metadata.ResourceId)
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Public API.
func (s *SubnetService) Update(ctx context.Context, req *pb.SubnetUpdateRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("SubnetService.Update").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := s.updateSubnet(ctx, req.Metadata)
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Public API.
func (s *SubnetService) Get(ctx context.Context, req *pb.SubnetGetRequest) (*pb.VPCSubnet, error) {

	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SubnetService.Get").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VPCSubnet, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		return s.get(ctx, cloudAccountId, subnetIdKey, req.Metadata.GetResourceId())
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

// Public API.
func (s *SubnetService) Search(ctx context.Context, req *pb.SubnetSearchRequest) (*pb.SubnetSearchResponse, error) {

	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SubnetService.Search").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.SubnetSearchResponse, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		if _, err := uuid.Parse(req.Spec.VpcId); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid vpc id")
		}

		if err := utils.ValidateLabels(req.Metadata.Labels); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		flattenedObject := protodb.Flattened{
			Columns: []string{"cloud_account_id", "deleted_timestamp", "value->'spec'->>'vpcId'"},
			Values:  []any{cloudAccountId, common.TimestampInfinityStr, req.Spec.VpcId},
		}

		labels := req.Metadata.Labels
		for key, value := range labels {
			column := fmt.Sprintf("value->'metadata'->'labels'->>'%s'", key)
			flattenedObject.Add(column, value)
		}

		whereString := flattenedObject.GetWhereString(1)

		query := fmt.Sprintf(`
			select %s
			from   subnet
			where  %s
			order by name
		`, transformer.ColumnsForFromRow(), whereString)

		args := flattenedObject.Values
		rows, err := s.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var items []*pb.VPCSubnet
		for rows.Next() {
			item, err := s.rowToSubnet(ctx, rows)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		resp := &pb.SubnetSearchResponse{
			Items: items,
		}
		return resp, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

// Public API: Delete a Subnet
func (s *SubnetService) Delete(ctx context.Context, req *pb.SubnetDeleteRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("SubnetService.Delete").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := s.deleteSubnet(ctx, req)
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Private API
func (s *SubnetService) UpdatePrivate(ctx context.Context, req *pb.SubnetUpdatePrivateRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("SubnetService.UpdatePrivate").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := s.updateSubnet(ctx, req.Metadata)
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Allow update of:
//   - Status
//
// Private API.
func (s *SubnetService) UpdateStatus(ctx context.Context, req *pb.SubnetUpdateStatusRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SubnetService.UpdateStatus").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
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
		updateFunc := func(subnet *pb.SubnetPrivate) error {
			subnet.Status = req.Status
			if req.Metadata.DeletedTimestamp != nil {
				subnet.Metadata.DeletedTimestamp = req.Metadata.DeletedTimestamp
			}
			return nil
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		if err := s.update(ctx, cloudAccountId, req.Metadata.ResourceId, req.Metadata.ResourceVersion, updateFunc); err != nil {
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

// Private API.
func (s *SubnetService) GetPrivate(ctx context.Context, req *pb.SubnetGetPrivateRequest) (*pb.SubnetPrivate, error) {

	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SubnetService.GetPrivate").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.SubnetPrivate, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		return s.getPrivate(ctx, cloudAccountId, subnetIdKey, req.Metadata.GetResourceId())
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

func (s *SubnetService) get(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*pb.VPCSubnet, error) {
	rows, err := s.selectSubnet(ctx, cloudAccountId, argName, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.rowToSubnet(ctx, rows)
}

func (s *SubnetService) getPrivate(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*pb.SubnetPrivate, error) {
	rows, err := s.selectSubnet(ctx, cloudAccountId, argName, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.sqlTransformer.FromRow(ctx, rows)
}

func (s *SubnetService) updateSubnet(ctx context.Context, metadata *pb.SubnetMetadata) (*emptypb.Empty, error) {

	resp, err := func() (*emptypb.Empty, error) {
		if metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		if metadata.Name == "" {
			return nil, status.Error(codes.InvalidArgument, "missing name")
		}
		// Validate name
		if err := networkutils.ValidateSubnetName(metadata.Name); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if err := utils.ValidateLabels(metadata.Labels); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		updateFunc := func(subnet *pb.SubnetPrivate) error {
			subnet.Metadata.Name = metadata.Name
			subnet.Metadata.Labels = metadata.Labels
			return nil
		}

		cloudAccountId := metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}
		if err := s.update(ctx, cloudAccountId, metadata.ResourceId, metadata.ResourceVersion, updateFunc); err != nil {
			return nil, err
		}

		return &emptypb.Empty{}, nil
	}()

	return resp, err
}

func (s *SubnetService) deleteSubnet(ctx context.Context, req *pb.SubnetDeleteRequest) (*emptypb.Empty, error) {
	if req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	cloudAccountId := req.Metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	if req.Spec == nil {
		return nil, status.Error(codes.InvalidArgument, "missing spec")
	}

	if _, err := s.vpcService.ValidateVPC(ctx, cloudAccountId, req.Spec.VpcId); err != nil {
		return nil, err
	}

	updateFunc := func(subnet *pb.SubnetPrivate) error {
		if subnet.Metadata.DeletionTimestamp == nil {
			subnet.Metadata.DeletionTimestamp = timestamppb.Now()
			subnet.Status.Phase = pb.SubnetPhase_SubnetPhase_Deleting
			subnet.Status.Message = "Subnet is deleting"
		}
		return nil
	}

	if err := s.update(ctx, cloudAccountId, req.Metadata.GetResourceId(), req.Metadata.ResourceVersion, updateFunc); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Update a subnet record using the user-provided updateFunc to update the subnet.
// This uses optimistic concurrency control to ensure that the record has not been updated between the select and update.
// Additionally, if the caller provides a resource version, optimistic concurrency control can be extended to
// previous get or search calls.
func (s *SubnetService) update(
	ctx context.Context,
	cloudAccountId string,
	resourceId string,
	resourceVersion string,
	updateFunc func(*pb.SubnetPrivate) error) error {

	query := fmt.Sprintf(`
		select %s
		from   subnet
		where  cloud_account_id = $1
			and  %s = $2
			and  deleted_timestamp = $3
	`, transformer.ColumnsForFromRow(), subnetIdKey)

	// Retry on conflict if caller did not provide resourceVersion.
	isRetryable := func(err error) bool {
		return resourceVersion == "" && status.Code(err) == codes.FailedPrecondition
	}

	err := retry.OnError(retry.DefaultRetry, isRetryable, func() error {
		rows, err := s.db.QueryContext(ctx, query, cloudAccountId, resourceId, common.TimestampInfinityStr)
		if err != nil {
			return err
		}
		defer rows.Close()
		if !rows.Next() {
			return status.Error(codes.NotFound, "resource not found")
		}
		subnet, err := s.sqlTransformer.FromRow(ctx, rows)
		if err != nil {
			return err
		}
		metadata := subnet.Metadata

		// If resource version was provided, ensure that stored version matches.
		if resourceVersion != "" && resourceVersion != metadata.ResourceVersion {
			return status.Error(codes.FailedPrecondition, "stored resource version does not match requested resource version")
		}

		// Update Subnet object.
		if err := updateFunc(subnet); err != nil {
			return err
		}

		// Flatten subnet into columns.
		flattened, err := s.sqlTransformer.Flatten(ctx, subnet)
		if err != nil {
			return err
		}

		args := append([]any{metadata.CloudAccountId, metadata.ResourceId, metadata.ResourceVersion}, flattened.Values...)

		deletedTimestamp := ""
		if metadata.DeletedTimestamp != nil {
			deletedTimestamp = "deleted_timestamp = '" + metadata.DeletedTimestamp.AsTime().Format(time.RFC3339) + "',"
		}

		// Update database.
		updateQuery := fmt.Sprintf(`
		update subnet
		set    resource_version = nextval('subnet_resource_version_seq'),
			   name = '%s',
			   %s
			   %s
		where  cloud_account_id = $1
		and    resource_id = $2
		and    resource_version = $3
		`, metadata.Name, deletedTimestamp, flattened.GetUpdateSetString(4))
		sqlResult, err := s.db.ExecContext(ctx, updateQuery, args...)
		if err != nil {
			pgErr := &pgconn.PgError{}
			if errors.As(err, &pgErr) && pgErr.Code == common.KErrUniqueViolation {
				return status.Error(codes.AlreadyExists, "update: subnet name already exists")
			}

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

// Validates and sets defaults in the provided Subnet object and stores it in the database.
func (s *SubnetService) create(ctx context.Context, subnet *pb.SubnetPrivate) error {
	ctx, _, span := obs.LogAndSpanFromContext(ctx).WithName("NetworkService.create").WithValues(logkeys.CloudAccountId, subnet.Metadata.CloudAccountId,
		logkeys.ResourceId, subnet.Metadata.GetResourceId()).Start()
	defer span.End()

	// TODO: Check quota

	// Validate
	if err := utils.ValidateLabels(subnet.Metadata.Labels); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Validate name
	if err := networkutils.ValidateSubnetName(subnet.Metadata.Name); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// validate subnet cidr.
	if err := networkutils.ValidateCIDR(subnet.Spec.CidrBlock); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Validate the VPCId passed is valid for this cloud account
	vpc, err := s.vpcService.ValidateVPC(ctx, subnet.Metadata.CloudAccountId, subnet.Spec.VpcId)
	if err != nil {
		if status.Code(err) == codes.NotFound || status.Code(err) == codes.InvalidArgument {
			return status.Error(codes.InvalidArgument, "invalid vpcId")
		}
		return status.Error(codes.Internal, err.Error())
	}

	// Validate subnet is within VPC CIDR
	if isSubnetWithinVPC, err := networkutils.IsCIDRWithinCIDR(vpc.Spec.CidrBlock, subnet.Spec.CidrBlock); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	} else if !isSubnetWithinVPC {
		return status.Error(codes.InvalidArgument, "subnet CIDR is not within VPC CIDR.")
	}

	// Use single transaction for:
	// - Locking DB row (Prevents race condition between requests on same or different instances of this service)
	// - Verify no overlaps with existing CIDRS
	// - Insert into database
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("BeginTx: %w", err)
	}
	defer tx.Rollback()

	// Lock the vpc row of this subnet
	err = s.lockVPCRow(ctx, tx, subnet.Metadata.CloudAccountId, subnet.Spec.VpcId)
	if err != nil {
		return err
	}

	if overlapsExistingCIDR, err := networkutils.OverlapsExistingCIDRs(ctx, tx, subnet); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	} else if overlapsExistingCIDR {
		return status.Error(codes.InvalidArgument, "subnet CIDR overlaps with existing subnet CIDR within the VPC.")
	}

	// Calculate resourceId
	resourceId, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	subnet.Metadata.ResourceId = resourceId.String()

	// Calculate name if not provided.
	if subnet.Metadata.Name == "" {
		subnet.Metadata.Name = subnet.Metadata.ResourceId
	}
	name := subnet.Metadata.Name

	// Flatten instance into columns.
	flattened, err := s.sqlTransformer.Flatten(ctx, subnet)
	if err != nil {
		return err
	}

	// Insert into database.
	query := fmt.Sprintf(`insert into subnet (resource_id, cloud_account_id, name, %s) values ($1, $2, $3, %s)`,
		flattened.GetColumnsString(), flattened.GetInsertValuesString(4))
	args := append([]any{subnet.Metadata.ResourceId, subnet.Metadata.CloudAccountId, name}, flattened.Values...)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == common.KErrUniqueViolation {
			return status.Error(codes.AlreadyExists, "insert: subnet "+name+" already exists")
		}
		return fmt.Errorf("insert: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// Caller must close the returned sql.Rows.
func (s *SubnetService) selectSubnet(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		select %s
		from   subnet
		where  cloud_account_id = $1
		  and  %s = $2
		  and  deleted_timestamp = $3
	`, transformer.ColumnsForFromRow(), argName)

	rows, err := s.db.QueryContext(ctx, query, cloudAccountId, arg, common.TimestampInfinityStr)
	if err != nil {
		return nil, fmt.Errorf("selectSubnet: %w", err)
	}
	if !rows.Next() {
		defer rows.Close()
		return nil, status.Error(codes.NotFound, "resource not found")
	}
	return rows, nil
}

// Read a database row into a public Subnet. Used for public APIs.
func (s *SubnetService) rowToSubnet(ctx context.Context, rows *sql.Rows) (*pb.VPCSubnet, error) {
	log := log.FromContext(ctx).WithName("NetworkService.rowToSubnet")
	subnetPrivate, err := s.sqlTransformer.FromRow(ctx, rows)
	if err != nil {
		return nil, fmt.Errorf("rowToSubnet: %w", err)
	}
	subnet := &pb.VPCSubnet{}
	if err := s.pbConverter.Transcode(subnetPrivate, subnet); err != nil {
		return nil, fmt.Errorf("rowToSubnet: %w", err)
	}
	log.V(9).Info("Read from database", logkeys.Subnet, subnet)
	return subnet, nil
}

func (s *SubnetService) lockVPCRow(ctx context.Context, tx *sql.Tx, cloudAccountId string, resourceId string) error {
	// This function locks a vpc row as long as the transaction is open
	query := `
		select *
		from   vpc
		where  cloud_account_id = $1
		  and  resource_id = $2
		  and  deleted_timestamp = $3
		for update
	`

	rows, err := tx.QueryContext(ctx, query, cloudAccountId, resourceId, common.TimestampInfinityStr)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return fmt.Errorf("lockVPCRow: %w", err)
	}
	if !rows.Next() {
		return status.Error(codes.NotFound, "resource not found")
	}

	return nil
}
