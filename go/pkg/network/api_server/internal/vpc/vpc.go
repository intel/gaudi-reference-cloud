// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vpc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/pbconvert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/config"
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
	vpcIdPrefix = "vpc-"
)

type VPCService struct {
	pb.UnimplementedVPCServiceServer
	pb.UnimplementedVPCPrivateServiceServer
	db                        *sql.DB
	cfg                       config.Config
	cloudAccountServiceClient pb.CloudAccountServiceClient
	sqlTransformer            *VPCSqlTransformer
	pbConverter               *pbconvert.PbConverter
}

func NewVPCService(
	db *sql.DB,
	config config.Config,
	cloudAccountServiceClient pb.CloudAccountServiceClient,
) (*VPCService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &VPCService{
		db:                        db,
		cfg:                       config,
		cloudAccountServiceClient: cloudAccountServiceClient,
		sqlTransformer:            NewVPCSqlTransformer(),
		pbConverter:               pbconvert.NewPbConverter(),
	}, nil
}

func (s *VPCService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("VPCService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (s *VPCService) PingPrivate(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("VPCService.PingPrivate")
	log.Info("PingPrivate")
	return &emptypb.Empty{}, nil
}

// Private API: Create a new VPC
func (s *VPCService) CreatePrivate(ctx context.Context, req *pb.VPCCreatePrivateRequest) (*pb.VPCPrivate, error) {
	return nil, nil
}

// Public API: Create a new VPC
func (s *VPCService) Create(ctx context.Context, req *pb.VPCCreateRequest) (*pb.VPC, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("VPCService.Create").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VPC, error) {
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

		// Transcode VPCSpec to VPCSpecPrivate.
		// Unmatched fields will remain at their default.
		vpcSpecPrivate := &pb.VPCSpecPrivate{}
		if err := s.pbConverter.Transcode(req.Spec, vpcSpecPrivate); err != nil {
			return nil, fmt.Errorf("unable to transcode vpc spec: %w", err)
		}

		vpc := &pb.VPCPrivate{
			Metadata: &pb.VPCMetadataPrivate{
				CloudAccountId: cloudAccountId,
				Name:           req.Metadata.Name,
				Labels:         req.Metadata.Labels,
			},
			Spec: vpcSpecPrivate,
			Status: &pb.VPCStatusPrivate{
				Phase:   pb.VPCPhase_VPCPhase_Provisioning,
				Message: "VPC is provisioning",
			},
		}

		if err := s.create(ctx, vpc); err != nil {
			return nil, err
		}

		// Query database and return response.
		return s.get(ctx, cloudAccountId, "resource_id", vpc.Metadata.ResourceId)
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Public API.
func (s *VPCService) Get(ctx context.Context, req *pb.VPCGetRequest) (*pb.VPC, error) {

	// Validate input.
	if req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("VPCService.Get").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VPC, error) {

		argName, arg, err := common.ResourceUniqueColumnAndValue(req.Metadata.GetResourceId(), req.Metadata.GetName())
		if err != nil {
			return nil, err
		}

		// Validate the vpcid is a valid resource id
		if argName == "resource_id" {
			_, err := uuid.Parse(req.Metadata.GetResourceId())
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid resource id")
			}
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		return s.get(ctx, cloudAccountId, argName, arg)
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

// Public API.
func (s *VPCService) Search(ctx context.Context, req *pb.VPCSearchRequest) (*pb.VPCSearchResponse, error) {

	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("VPCService.Search").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VPCSearchResponse, error) {
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
			from   vpc
			where  %s
			order by name
		`, s.sqlTransformer.ColumnsForFromRow(), whereString)

		args := flattenedObject.Values
		rows, err := s.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var items []*pb.VPC
		for rows.Next() {
			item, err := s.rowToVPC(ctx, rows)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		resp := &pb.VPCSearchResponse{
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

// Public API: Delete a VPC
func (s *VPCService) Delete(ctx context.Context, req *pb.VPCDeleteRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("VPCService.Delete").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()
	log.Info("Request", logkeys.Request, req)

	resp, err := s.deleteVPC(ctx, req.Metadata)
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

// Private API.
func (s *VPCService) GetPrivate(ctx context.Context, req *pb.VPCGetPrivateRequest) (*pb.VPCPrivate, error) {

	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("VPCService.GetPrivate").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VPCPrivate, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		return s.getPrivate(ctx, cloudAccountId, "resource_id", req.Metadata.GetResourceId())
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}

	return resp, err
}

// Allows update of:
//   - Name
//   - Labels
//
// Public API
func (s *VPCService) Update(ctx context.Context, req *pb.VPCUpdateRequest) (*emptypb.Empty, error) {
	if req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("VPCService.Update").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*emptypb.Empty, error) {

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		updateFunc := func(vpc *pb.VPCPrivate) error {
			vpc.Metadata.Name = req.Metadata.GetName()
			vpc.Metadata.Labels = req.Metadata.Labels
			return nil
		}

		if err := s.update(ctx, cloudAccountId, req.Metadata.GetResourceId(), req.Metadata.GetName(), req.Metadata.ResourceVersion, updateFunc); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

// Allow update of:
//   - Status
//
// Private API.
func (s *VPCService) UpdateStatus(ctx context.Context, req *pb.VPCUpdateStatusRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("VPCService.UpdateStatus").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
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
		updateFunc := func(vpc *pb.VPCPrivate) error {
			vpc.Status = req.Status

			return nil
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		if err := s.update(ctx, cloudAccountId, req.Metadata.ResourceId, "", req.Metadata.ResourceVersion, updateFunc); err != nil {
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

func (s *VPCService) get(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*pb.VPC, error) {
	rows, err := s.selectVPC(ctx, cloudAccountId, argName, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.rowToVPC(ctx, rows)
}

func (s *VPCService) getPrivate(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*pb.VPCPrivate, error) {
	rows, err := s.selectVPC(ctx, cloudAccountId, argName, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.sqlTransformer.FromRow(ctx, rows)
}

func (s *VPCService) deleteVPC(ctx context.Context, metadata *pb.VPCMetadataReference) (*emptypb.Empty, error) {
	if metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	cloudAccountId := metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	updateFunc := func(vpc *pb.VPCPrivate) error {
		if vpc.Metadata.DeletionTimestamp == nil {
			vpc.Metadata.DeletionTimestamp = timestamppb.Now()
		}
		return nil
	}

	if err := s.update(ctx, cloudAccountId, metadata.GetResourceId(), "", "", updateFunc); err != nil { //metadata.ResourceVersion
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Update a vpc record using the user-provided updateFunc to update the vpc.
// This uses optimistic concurrency control to ensure that the record has not been updated between the select and update.
// Additionally, if the caller provides a resource version, optimistic concurrency control can be extended to
// previous get or search calls.
func (s *VPCService) update(
	ctx context.Context,
	cloudAccountId string,
	resourceId string,
	name string,
	resourceVersion string,
	updateFunc func(*pb.VPCPrivate) error) error {
	argName, arg, err := common.ResourceUniqueColumnAndValue(resourceId, name)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	query := fmt.Sprintf(`
		select %s
		from   vpc
		where  cloud_account_id = $1
			and  %s = $2
			and  deleted_timestamp = $3
	`, s.sqlTransformer.ColumnsForFromRow(), argName)

	// Retry on conflict if caller did not provide resourceVersion.
	isRetryable := func(err error) bool {
		return resourceVersion == "" && status.Code(err) == codes.FailedPrecondition
	}

	err = retry.OnError(retry.DefaultRetry, isRetryable, func() error {
		rows, err := s.db.QueryContext(ctx, query, cloudAccountId, arg, common.TimestampInfinityStr)
		if err != nil {
			return err
		}
		defer rows.Close()
		if !rows.Next() {
			return status.Error(codes.NotFound, "resource not found")
		}
		vpc, err := s.sqlTransformer.FromRow(ctx, rows)
		if err != nil {
			return err
		}
		metadata := vpc.Metadata

		// If resource version was provided, ensure that stored version matches.
		if resourceVersion != "" && resourceVersion != metadata.ResourceVersion {
			return status.Error(codes.FailedPrecondition, "stored resource version does not match requested resource version")
		}

		// Update VPC object.
		if err := updateFunc(vpc); err != nil {
			return err
		}

		// Flatten vpc into columns.
		flattened, err := s.sqlTransformer.Flatten(ctx, vpc)
		if err != nil {
			return err
		}

		args := append([]any{metadata.CloudAccountId, metadata.ResourceId, metadata.ResourceVersion}, flattened.Values...)

		// Update database.
		updateQuery := fmt.Sprintf(`
		update vpc
		set    resource_version = nextval('vpc_resource_version_seq'),
			   %s
		where  cloud_account_id = $1
		and    resource_id = $2
		and    resource_version = $3
		`, flattened.GetUpdateSetString(4))
		sqlResult, err := s.db.ExecContext(ctx, updateQuery, args...)
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

// Validates and sets defaults in the provided VPC object and stores it in the database.
func (s *VPCService) create(ctx context.Context, vpc *pb.VPCPrivate) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("NetworkService.create").WithValues(logkeys.CloudAccountId, vpc.Metadata.CloudAccountId,
		logkeys.ResourceId, vpc.Metadata.GetResourceId()).Start()
	defer span.End()

	log.V(9).Info("hi")

	// TODO: Check quota

	// Validate
	if err := utils.ValidateLabels(vpc.Metadata.Labels); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if err := networkutils.ValidateCIDR(vpc.Spec.CidrBlock); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Calculate resourceId
	resourceId, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	vpc.Metadata.ResourceId = resourceId.String()

	// Calculate name if not provided.
	if vpc.Metadata.Name == "" {
		vpc.Metadata.Name = vpc.Metadata.ResourceId
	}
	name := vpc.Metadata.Name

	// Insert into database in a single transaction.
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("BeginTx: %w", err)
	}
	defer tx.Rollback()

	// Flatten vpc into columns.
	flattened, err := s.sqlTransformer.Flatten(ctx, vpc)
	if err != nil {
		return err
	}

	// Insert into database.
	query := fmt.Sprintf(`insert into vpc (resource_id, cloud_account_id, name, %s) values ($1, $2, $3, %s)`,
		flattened.GetColumnsString(), flattened.GetInsertValuesString(4))
	args := append([]any{vpc.Metadata.ResourceId, vpc.Metadata.CloudAccountId, name}, flattened.Values...)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == common.KErrUniqueViolation {
			return status.Error(codes.AlreadyExists, "insert: vpc "+name+" already exists")
		}
		return fmt.Errorf("insert: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// Caller must close the returned sql.Rows.
func (s *VPCService) selectVPC(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		select %s
		from   vpc
		where  cloud_account_id = $1
		  and  %s = $2
		  and  deleted_timestamp = $3
	`, s.sqlTransformer.ColumnsForFromRow(), argName)

	rows, err := s.db.QueryContext(ctx, query, cloudAccountId, arg, common.TimestampInfinityStr)
	if err != nil {
		return nil, fmt.Errorf("selectVPC: %w", err)
	}
	if !rows.Next() {
		defer rows.Close()
		return nil, status.Error(codes.NotFound, "resource not found")
	}
	return rows, nil
}

// Read a database row into a public VPC. Used for public APIs.
func (s *VPCService) rowToVPC(ctx context.Context, rows *sql.Rows) (*pb.VPC, error) {
	log := log.FromContext(ctx).WithName("NetworkService.rowToVPC")
	vpcPrivate, err := s.sqlTransformer.FromRow(ctx, rows)
	if err != nil {
		return nil, fmt.Errorf("rowToVPC: %w", err)
	}
	vpc := &pb.VPC{}
	if err := s.pbConverter.Transcode(vpcPrivate, vpc); err != nil {
		return nil, fmt.Errorf("rowToVPC: %w", err)
	}
	log.V(9).Info("Read from database", logkeys.VPC, vpc)
	return vpc, nil
}
