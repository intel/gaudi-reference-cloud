// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package address_translation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/client-go/util/retry"
	"math/rand"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AddressTranslationPrivateService struct {
	pb.UnimplementedAddressTranslationPrivateServiceServer
	db                        *sql.DB
	cfg                       config.Config
	cloudAccountServiceClient pb.CloudAccountServiceClient
	sqlTransformer            *AddressTranslationSqlTransformer
}

func NewAddressTranslationPrivateService(
	db *sql.DB,
	config config.Config,
	cloudAccountServiceClient pb.CloudAccountServiceClient,
) (*AddressTranslationPrivateService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &AddressTranslationPrivateService{
		db:                        db,
		cfg:                       config,
		cloudAccountServiceClient: cloudAccountServiceClient,
		sqlTransformer:            NewAddressTranslationSqlTransformer(),
	}, nil
}

func (s *AddressTranslationPrivateService) PingPrivate(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("AddressTranslationService.PingPrivate")
	log.Info("PingPrivate")
	return &emptypb.Empty{}, nil
}

// Private API: Create a new Address Translation record.
func (s *AddressTranslationPrivateService) CreatePrivate(ctx context.Context, req *pb.AddressTranslationCreatePrivateRequest) (*pb.AddressTranslationPrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AddressTranslationService.Create").Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.AddressTranslationPrivate, error) {
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

		translationType, err := IsValidTranslationType(req.Spec.TranslationType)
		if err != nil {
			return nil, err
		}

		// Duplication validation is done based on db constraint (See error handling in create function)

		// TODO: Validate portId

		// TODO: get real values
		profileId := uuid.New().String()
		ipAddress := fmt.Sprintf("192.168.1.%d", rand.Intn(255))
		macAddress := fmt.Sprintf("02:00:5e:%02x:%02x:%02x", rand.Intn(256), rand.Intn(256), rand.Intn(256))

		resourceId, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}

		addressTranslation := &pb.AddressTranslationPrivate{
			Metadata: &pb.AddressTranslationMetadataPrivate{
				CloudAccountId: cloudAccountId,
				ResourceId:     resourceId.String(),
			},
			Spec: &pb.AddressTranslationSpecPrivate{
				TranslationType: string(translationType),
				PortId:          req.Spec.PortId,
				ProfileId:       profileId,
				IpAddress:       ipAddress,
				MacAddress:      macAddress,
			},
			Status: &pb.AddressTranslationStatus{
				Phase:   pb.AddressTranslationPhase_AddressTranslationPhase_Provisioning,
				Message: "AddressTranslation is provisioning",
			},
		}

		if err := s.create(ctx, addressTranslation); err != nil {
			return nil, err
		}

		// Query database and return response.
		return s.get(ctx, cloudAccountId, addressTranslation.Metadata.ResourceId)
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Private API: Get an Address Translation record.
func (s *AddressTranslationPrivateService) GetPrivate(ctx context.Context, req *pb.AddressTranslationGetPrivateRequest) (*pb.AddressTranslationPrivate, error) {
	// Validate input.
	if req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AddressTranslationService.GetPrivate").WithValues(logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)

	cloudAccountId := req.Metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	// Validate resource id is a valid UUID.
	resourceId := req.Metadata.GetResourceId()
	_, err := uuid.Parse(resourceId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	return s.get(ctx, cloudAccountId, resourceId)
}

// Private API: List Address Translation records. Optionally filter by column
func (s *AddressTranslationPrivateService) ListPrivate(ctx context.Context, req *pb.AddressTranslationListPrivateRequest) (*pb.AddressTranslationPrivateList, error) {
	if req.Spec == nil {
		return nil, status.Error(codes.InvalidArgument, "missing spec")
	}
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AddressTranslationService.ListPrivate").Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	if err := cloudaccount.CheckValidId(req.Metadata.CloudAccountId); err != nil {
		return nil, err
	}

	resp, err := s.list(ctx, req)
	log.LogResponseOrError(logger, req, resp, err)

	return resp, err
}

// Private API: Delete an Address Translation record.
func (s *AddressTranslationPrivateService) DeletePrivate(ctx context.Context, req *pb.AddressTranslationGetPrivateRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("DeleteService.Delete").WithValues(logkeys.ResourceId, req.Metadata.GetResourceId()).Start()

	if req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	defer span.End()
	log.Info("Request", logkeys.Request, req)

	cloudAccountId := req.Metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	updateFunc := func(addressTranslation *pb.AddressTranslationPrivate) error {
		if addressTranslation.Metadata.DeletionTimestamp != nil {
			return nil
		}

		deletionTime := timestamppb.Now()
		addressTranslation.Metadata.DeletionTimestamp = deletionTime
		addressTranslation.Status.Phase = pb.AddressTranslationPhase_AddressTranslationPhase_Deleting
		addressTranslation.Status.Message = "AddressTranslation is deleting"
		return nil
	}

	if err := s.update(ctx, cloudAccountId, req.Metadata.ResourceId, req.Metadata.ResourceVersion, updateFunc); err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
		return nil, err
	}

	log.Info("Response", logkeys.Response, "deleted successfully")

	return &emptypb.Empty{}, nil
}

func (s *AddressTranslationPrivateService) UpdateStatusPrivate(ctx context.Context, req *pb.AddressTranslationUpdateStatusPrivateRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("AddressTranslationPrivateService.UpdateStatus").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()
	log.Info("Request", logkeys.Request, req)
	if req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}
	if req.Status == nil {
		return nil, status.Error(codes.InvalidArgument, "missing status")
	}
	cloudAccountId := req.Metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	updateFunc := func(addressTranslation *pb.AddressTranslationPrivate) error {
		addressTranslation.Status = req.Status
		if req.Metadata.DeletedTimestamp != nil {
			addressTranslation.Metadata.DeletedTimestamp = req.Metadata.DeletedTimestamp
		}
		return nil
	}

	err := s.update(ctx, cloudAccountId, req.Metadata.ResourceId, req.Metadata.ResourceVersion, updateFunc)
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
		return nil, err
	}

	log.Info("Response", logkeys.Response, "status updated successfully")

	return &emptypb.Empty{}, nil
}

func (s *AddressTranslationPrivateService) get(ctx context.Context, cloudAccountId string, resourceId string) (*pb.AddressTranslationPrivate, error) {
	rows, err := s.selectAddressTranslation(ctx, cloudAccountId, resourceId)
	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		return nil, err
	}
	return s.rowToAddressTranslation(ctx, rows)
}

// Read a database row into a public Address Translation. Used for public APIs.
func (s *AddressTranslationPrivateService) rowToAddressTranslation(ctx context.Context, rows *sql.Rows) (*pb.AddressTranslationPrivate, error) {
	log := log.FromContext(ctx).WithName("NetworkService.rowToAddressTranslation")
	addressTranslationPrivate, err := s.sqlTransformer.FromRow(ctx, rows)
	if err != nil {
		return nil, fmt.Errorf("rowToAddressTranslation: %w", err)
	}
	log.V(9).Info("Read from database", logkeys.ADDRESS_TRANSLATION, addressTranslationPrivate)
	return addressTranslationPrivate, nil
}

// Caller must close the returned sql.Rows.
func (s *AddressTranslationPrivateService) selectAddressTranslation(ctx context.Context, cloudAccountId string, resourceId string) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		select %s
		from   address_translation
		where  cloud_account_id = $1
		and resource_id = $2
		and  deleted_timestamp = $3
	`, s.sqlTransformer.ColumnsForFromRow())

	rows, err := s.db.QueryContext(ctx, query, cloudAccountId, resourceId, common.TimestampInfinityStr)
	if err != nil {
		return nil, fmt.Errorf("selectAddressTranslation: %w", err)
	}
	if !rows.Next() {
		defer rows.Close()
		return nil, status.Error(codes.NotFound, "resource not found")
	}
	return rows, nil
}

func (s *AddressTranslationPrivateService) create(ctx context.Context, addressTranslation *pb.AddressTranslationPrivate) error {
	log := log.FromContext(ctx).WithName("AddressTranslationService.create")
	log.Info("Create", logkeys.ADDRESS_TRANSLATION, addressTranslation)

	// Flatten addressTranslation into columns.
	flattened, err := s.sqlTransformer.Flatten(ctx, addressTranslation)
	if err != nil {
		return err
	}

	// Insert into database.
	query := fmt.Sprintf(`insert into address_translation (resource_id, cloud_account_id, %s) values ($1, $2, %s)`,
		flattened.GetColumnsString(), flattened.GetInsertValuesString(3))
	args := append([]any{addressTranslation.Metadata.ResourceId, addressTranslation.Metadata.CloudAccountId}, flattened.Values...)

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == common.KErrUniqueViolation {
			errMsg := fmt.Sprintf("AddressTranslation with the same CloudAccountId (%s), PortId (%s), and TranslationType (%s) already exists", addressTranslation.Metadata.CloudAccountId, addressTranslation.Spec.PortId, addressTranslation.Spec.TranslationType)
			return status.Error(codes.AlreadyExists, errMsg)
		}
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}

func (s *AddressTranslationPrivateService) list(ctx context.Context, addressTranslation *pb.AddressTranslationListPrivateRequest) (*pb.AddressTranslationPrivateList, error) {
	log := log.FromContext(ctx).WithName("AddressTranslationService.list")
	log.Info("List. Query Params:", logkeys.ADDRESS_TRANSLATION, addressTranslation)
	// Flatten addressTranslation into columns.
	flattened := s.sqlTransformer.FlattenForList(ctx, addressTranslation)

	whereString := flattened.GetWhereString(2)

	query := fmt.Sprintf(`
			select %s
			from   address_translation
			where cloud_account_id = $1  
			and %s
		`, s.sqlTransformer.ColumnsForFromRow(), whereString)

	args := append([]any{addressTranslation.Metadata.CloudAccountId}, flattened.Values...)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*pb.AddressTranslationPrivate
	for rows.Next() {
		item, err := s.rowToAddressTranslation(ctx, rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	resp := &pb.AddressTranslationPrivateList{
		Items: items,
	}
	return resp, nil
}

func (s *AddressTranslationPrivateService) update(ctx context.Context, cloudAccountId string, resourceId string, resourceVersion string, updateFunc func(*pb.AddressTranslationPrivate) error) error {
	log := log.FromContext(ctx).WithName("AddressTranslationService.update")
	log.Info("Update", logkeys.CloudAccountId, cloudAccountId, logkeys.ADDRESS_TRANSLATION, resourceId)

	isRetryable := func(err error) bool { return resourceVersion == "" && status.Code(err) == codes.FailedPrecondition }
	return retry.OnError(retry.DefaultRetry, isRetryable, func() error {
		addressTranslation, err := s.get(ctx, cloudAccountId, resourceId)
		if err != nil {
			return err
		}

		if resourceVersion != "" && resourceVersion != addressTranslation.Metadata.ResourceVersion {
			return status.Error(codes.FailedPrecondition, "resource version mismatch")
		}

		if err := updateFunc(addressTranslation); err != nil {
			return err
		}

		deletedTimestamp := ""
		if addressTranslation.Metadata.DeletedTimestamp != nil {
			deletedTimestamp = "deleted_timestamp = '" + addressTranslation.Metadata.DeletedTimestamp.AsTime().Format(time.RFC3339) + "',"
		}

		// Flatten addressTranslation into columns.
		flattened, err := s.sqlTransformer.Flatten(ctx, addressTranslation)
		if err != nil {
			return err
		}

		args := append([]any{addressTranslation.Metadata.CloudAccountId, addressTranslation.Metadata.ResourceId, common.TimestampInfinityStr}, flattened.Values...)
		query := fmt.Sprintf(`
		update address_translation
		set resource_version = nextval('address_translation_resource_version_seq'),
		    %s
			%s
		where cloud_account_id = $1 
		and resource_id = $2
		and deleted_timestamp = $3
	`, deletedTimestamp, flattened.GetUpdateSetString(4))
		result, err := s.db.ExecContext(ctx, query, args...)

		if err != nil {
			return fmt.Errorf("updateStatus: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("rows affected: %w", err)
		}

		if rowsAffected < 1 {
			return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
		}

		return nil
	})
}
