// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	queryAddAccess = `
		INSERT INTO cloud_accounts_region_access (
			cloudaccount_id, region_name, admin_name
		) VALUES (
			$1, $2, $3
		)
	`

	queryRemoveAccess = `
		DELETE FROM cloud_accounts_region_access 
		WHERE cloudaccount_id = $1 AND region_name = $2
	`

	queryCheckRegionAccess = `
		SELECT EXISTS (
			SELECT 1 
			FROM cloud_accounts_region_access 
			WHERE cloudaccount_id = $1 AND region_name = $2
		)
	`

	querySelect = `
		SELECT cloudaccount_id, region_name, admin_name, created_at
		FROM cloud_accounts_region_access
		WHERE cloudaccount_id = $1
	`

	querySelectAll = `
		SELECT car.cloudaccount_id, car.region_name, car.admin_name, car.created_at
		FROM cloud_accounts_region_access car
		JOIN region r ON car.region_name = r.name
	`

	queryCheckRegionType = `
		SELECT type
		FROM region
		WHERE name = $1
		LIMIT 1
	`
)

type RegionAccessRepository struct {
	db *sql.DB
}

func NewRegionAccessRepository(db *sql.DB) (*RegionAccessRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}

	return &RegionAccessRepository{
		db: db,
	}, nil
}

func (repo *RegionAccessRepository) AddAccess(ctx context.Context, cloudAccountId, regionName string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionAccessRepository.AddAccess").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	// Check if the region type is "controlled" or "open"
	var regionType string
	err := repo.db.QueryRowContext(ctx, queryCheckRegionType, regionName).Scan(&regionType)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error(err, "region not found", "regionName", regionName)
			return status.Errorf(codes.Internal, "region not found: %s", regionName)
		}
		logger.Error(err, "error fetching region type from database", "regionName", regionName)
		return status.Errorf(codes.Internal, "error fetching region type from database")
	}

	if regionType != "controlled" {
		logger.Error(nil, "only regions with type 'controlled' can be accessed", "regionType", regionType)
		return status.Errorf(codes.Internal, "only regions with type 'controlled' can be accessed: %s", regionType)
	}

	userEmail, err := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if err != nil {
		logger.Error(err, "failed to extract email")
		return status.Errorf(codes.Internal, "failed to extract email")
	}

	_, err = repo.db.ExecContext(ctx, queryAddAccess, cloudAccountId, regionName, userEmail)
	if err != nil {
		logger.Error(err, "failed to execute add access query", "cloudAccountId", cloudAccountId, "regionName", regionName)
		return status.Errorf(codes.Internal, "failed to execute add access query")
	}
	return nil
}

func (repo *RegionAccessRepository) RemoveAccess(ctx context.Context, cloudAccountId, regionName string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionAccessRepository.RemoveAccess").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	_, err := repo.db.ExecContext(ctx, queryRemoveAccess, cloudAccountId, regionName)
	if err != nil {
		logger.Error(err, "failed to execute remove access query", "cloudAccountId", cloudAccountId, "regionName", regionName)
		return status.Errorf(codes.Internal, "failed to execute remove access query")
	}
	return nil
}

func (repo *RegionAccessRepository) CheckRegionAccess(ctx context.Context, cloudAccountId, regionName string) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionAccessRepository.CheckRegionAccess").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	var exists bool
	err := repo.db.QueryRowContext(ctx, queryCheckRegionAccess, cloudAccountId, regionName).Scan(&exists)
	if err != nil {
		logger.Error(err, "failed to check region access", "cloudAccountId", cloudAccountId, "regionName", regionName)
		return false, fmt.Errorf("failed to check region access")
	}
	return exists, nil
}

func (repo *RegionAccessRepository) ReadAccess(ctx context.Context, cloudAccountId, regionAccessType *string) ([]*pb.RegionAccessResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionAccessRepository.ReadAccess").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	query := querySelectAll
	var args []interface{}
	var conditions []string

	if cloudAccountId != nil && *cloudAccountId != "" {
		conditions = append(conditions, "car.cloudaccount_id = $"+fmt.Sprint(len(args)+1))
		args = append(args, *cloudAccountId)
	}

	if regionAccessType != nil && *regionAccessType != "" {
		conditions = append(conditions, "r.type = $"+fmt.Sprint(len(args)+1))
		args = append(args, *regionAccessType)
	}

	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions, " AND ")
	}

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error(err, "failed to execute read access query", "query", query, "args", args)
		return nil, fmt.Errorf("failed to execute read access query")
	}
	defer rows.Close()

	var accessList []*pb.RegionAccessResponse
	for rows.Next() {
		access, err := scanRegionAccessRow(rows)
		if err != nil {
			logger.Error(err, "error scanning region access row")
			return nil, fmt.Errorf("error scanning region access row")
		}
		accessList = append(accessList, access)
	}

	if err := rows.Err(); err != nil {
		logger.Error(err, "error iterating over rows")
		return nil, fmt.Errorf("error iterating over rows")
	}

	return accessList, nil
}

func joinConditions(conditions []string, separator string) string {
	if len(conditions) == 1 {
		return conditions[0]
	}
	return fmt.Sprint(conditions[0], separator, conditions[1])
}

func scanRegionAccessRow(rows *sql.Rows) (*pb.RegionAccessResponse, error) {
	var access pb.RegionAccessResponse
	var createdAt time.Time
	err := rows.Scan(&access.CloudaccountId, &access.RegionName, &access.AdminName, &createdAt)
	if err != nil {
		return nil, err
	}
	access.Created = timestamppb.New(createdAt)
	return &access, nil
}
