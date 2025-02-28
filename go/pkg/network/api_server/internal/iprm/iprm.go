// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iprm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/subnet"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/transformer"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/client-go/util/retry"
)

const (
	portIdKey = "resource_id"
)

type IPRMService struct {
	//pb.UnimplementedIPResourceManagerServiceServer
	pb.UnimplementedIPRMPrivateServiceServer
	db                        *sql.DB
	cfg                       config.Config
	cloudAccountServiceClient pb.CloudAccountServiceClient
	sqlTransformer            *PortSQLTransformer
	subnetService             *subnet.SubnetService
}

func NewIPRMService(
	db *sql.DB,
	config config.Config,
	cloudAccountServiceClient pb.CloudAccountServiceClient,
	subnetService *subnet.SubnetService,
) (*IPRMService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}

	return &IPRMService{
		db:                        db,
		cfg:                       config,
		cloudAccountServiceClient: cloudAccountServiceClient,
		sqlTransformer:            NewPortSQLTransformer(),
		subnetService:             subnetService,
	}, nil
}

func (s *IPRMService) PingPrivate(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("IPRMService.PingPrivate")
	log.Info("PingPrivate")
	return &emptypb.Empty{}, nil
}

// create a port in sdn, if its not exists.
func (s *IPRMService) ReservePort(ctx context.Context, req *pb.ReservePortRequest) (*pb.PortPrivate, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("IPRMService.ReservePort").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*pb.PortPrivate, error) {
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

		// validate subnet.
		_, err := s.subnetService.Get(ctx, &pb.SubnetGetRequest{
			Metadata: &pb.SubnetMetadataReference{
				CloudAccountId: cloudAccountId,
				NameOrId:       &pb.SubnetMetadataReference_ResourceId{ResourceId: req.Spec.SubnetId},
			},
		})

		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid subnet")
		}
		port := &pb.PortPrivate{
			Metadata: &pb.PortMetadataPrivate{
				CloudAccountId: cloudAccountId,
			},
			Spec: &pb.PortSpecPrivate{
				SubnetId:        req.Spec.SubnetId,
				IpuSerialNumber: req.Spec.IpuSerialNumber,
				ChassisId:       req.Spec.ChassisId,
				IpAddress:       req.Spec.IpAddress,
				MacAddress:      req.Spec.MacAddress,
				SshEnabled:      req.Spec.SshEnabled,
				InternetAccess:  req.Spec.InternetAccess,
			},
			Status: &pb.PortStatusPrivate{
				Phase:   pb.PortPhase_PortPhase_Provisioning,
				Message: "Port is provisioning",
			},
		}
		if err := s.create(ctx, port); err != nil {
			pgErr := &pgconn.PgError{}
			// Unique violation means that the port already exists.
			if errors.As(err, &pgErr) && pgErr.Code == common.KErrUniqueViolation {
				// Return the port that already exists.
				return s.get(ctx, cloudAccountId, map[string]interface{}{
					"value->'spec'->>'ipuSerialNumber'": port.Spec.IpuSerialNumber,
					"value->'spec'->>'chassisId'":       port.Spec.ChassisId,
					"value->'spec'->>'macAddress'":      port.Spec.MacAddress,
				})
			}
			return nil, err
		}
		// Query database and return response.
		return s.get(ctx, cloudAccountId, map[string]interface{}{
			portIdKey: port.Metadata.ResourceId,
		})
	}()

	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Remove a port from sdn.
func (s *IPRMService) ReleasePort(ctx context.Context, req *pb.ReleasePortRequest) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("IPRMService.ReleasePort").Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	// TODO: validate port id.
	if req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	cloudAccountId := req.Metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	updateFunc := func(port *pb.PortPrivate) error {
		// TODO: use status only ?!
		if port.Metadata.DeletionTimestamp == nil {
			port.Metadata.DeletionTimestamp = timestamppb.Now()
			port.Status = &pb.PortStatusPrivate{
				Phase:   pb.PortPhase_PortPhase_Deleting,
				Message: "Port deleting",
			}
		}
		return nil
	}

	if err := s.update(ctx, cloudAccountId, req.Metadata.GetResourceId(), req.Metadata.ResourceVersion, updateFunc); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Allow update of:
//   - Status
//
// Private API.
func (s *IPRMService) UpdateStatus(ctx context.Context, req *pb.PortUpdateStatusRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IPRMService.UpdateStatus").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
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
		updateFunc := func(port *pb.PortPrivate) error {
			port.Status = req.Status

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

func (s *IPRMService) GetPortPrivate(ctx context.Context, req *pb.GetPortPrivateRequest) (*pb.PortPrivate, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IPRMService.GetPrivate").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.PortPrivate, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		return s.get(ctx, cloudAccountId, map[string]interface{}{
			portIdKey: req.Metadata.GetResourceId(),
		})
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

func (s *IPRMService) get(ctx context.Context, cloudAccountId string, args map[string]interface{}) (*pb.PortPrivate, error) {
	rows, err := s.selectPort(ctx, cloudAccountId, args)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.rowToPort(ctx, rows)
}

func (s *IPRMService) rowToPort(ctx context.Context, rows *sql.Rows) (*pb.PortPrivate, error) {
	log := log.FromContext(ctx).WithName("NetworkService.rowToPort")
	port, err := s.sqlTransformer.FromRow(ctx, rows)
	if err != nil {
		return nil, fmt.Errorf("FromRow: %w", err)
	}
	log.V(9).Info("Read from database", logkeys.PORT, port)
	return port, nil
}

// Caller must close the returned sql.Rows.
func (s *IPRMService) selectPort(ctx context.Context, cloudAccountId string, args map[string]interface{}) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		select %s
		from   port
		where  cloud_account_id = $1
		and  deleted_timestamp = $2
	`, transformer.ColumnsForFromRow())

	values := []interface{}{
		cloudAccountId,
		common.TimestampInfinityStr,
	}
	i := 3

	for k, v := range args {
		query += fmt.Sprintf(" and %s = $%d", k, i)
		values = append(values, v)
		i += 1
	}

	rows, err := s.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("selectPort: %w", err)
	}
	if !rows.Next() {
		defer rows.Close()
		return nil, status.Error(codes.NotFound, "resource not found")
	}
	return rows, nil
}

func (s *IPRMService) create(ctx context.Context, port *pb.PortPrivate) error {
	ctx, _, span := obs.LogAndSpanFromContext(ctx).WithName("IPRMService.create").WithValues(logkeys.CloudAccountId, port.Metadata.CloudAccountId).Start()
	defer span.End()

	// create resourceId
	resourceId, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	port.Metadata.ResourceId = resourceId.String()

	name := port.Metadata.Name

	// Insert into database in a single transaction.
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("BeginTx: %w", err)
	}
	defer tx.Rollback()

	// Flatten instance into columns.
	flattened, err := s.sqlTransformer.Flatten(ctx, port)
	if err != nil {
		return err
	}

	// Insert into database.
	query := fmt.Sprintf(`insert into port (resource_id, cloud_account_id, name, %s) values ($1, $2, $3, %s)`,
		flattened.GetColumnsString(), flattened.GetInsertValuesString(4))
	args := append([]any{port.Metadata.ResourceId, port.Metadata.CloudAccountId, name}, flattened.Values...)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// Update a port record using the user-provided updateFunc to update the port.
// This uses optimistic concurrency control to ensure that the record has not been updated between the select and update.
// Additionally, if the caller provides a resource version, optimistic concurrency control can be extended to
// previous get or search calls.
func (s *IPRMService) update(
	ctx context.Context,
	cloudAccountId string,
	resourceId string,
	resourceVersion string,
	updateFunc func(*pb.PortPrivate) error) error {

	query := fmt.Sprintf(`
		select %s
		from   port
		where  cloud_account_id = $1
			and  %s = $2
			and  deleted_timestamp = $3
	`, transformer.ColumnsForFromRow(), portIdKey)

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

		port, err := s.sqlTransformer.FromRow(ctx, rows)
		if err != nil {
			return err
		}
		metadata := port.Metadata

		// If resource version was provided, ensure that stored version matches.
		if resourceVersion != "" && resourceVersion != metadata.ResourceVersion {
			return status.Error(codes.FailedPrecondition, "stored resource version does not match requested resource version")
		}

		// Update Port object.
		if err := updateFunc(port); err != nil {
			return err
		}

		// Flatten port into columns.
		flattened, err := s.sqlTransformer.Flatten(ctx, port)
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
		update port
		set    resource_version = nextval('subnet_resource_version_seq'),
			   %s
			   %s
			   , name = (value::jsonb->'metadata'->>'name')::text
		where  cloud_account_id = $1
		and    resource_id = $2
		and    resource_version = $3
		`, deletedTimestamp, flattened.GetUpdateSetString(4))

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
