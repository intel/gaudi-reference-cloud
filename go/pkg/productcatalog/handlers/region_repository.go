// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	queryAddRegion = `
		INSERT INTO region (
			name, friendly_name, type, subnet, availability_zone, prefix, is_default, api_dns, vnet, admin_name
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9 ,$10
		)
	`

	queryUpdateRegionBase = "UPDATE region SET "
	queryDeleteRegion     = "DELETE FROM region WHERE name = $1"
	queryGetRegionByName  = `
		SELECT name, friendly_name, type, subnet, availability_zone, prefix, is_default, api_dns, vnet, admin_name, created_at, updated_at 
		FROM region 
		WHERE name = $1
	`
	queryGetRegionsBase = `
		SELECT 
			name, friendly_name, type, subnet, availability_zone, 
			prefix, is_default, api_dns, vnet, admin_name, created_at, updated_at 
		FROM 
			region
		WHERE 1=1
	`
	queryGetUserRegionsBase = `
		SELECT 
			r.name, r.friendly_name, r.type, r.subnet, r.availability_zone, 
			r.prefix, r.is_default, r.api_dns, r.vnet
		FROM 
			region r
		LEFT JOIN 
			cloud_accounts_region_access cara 
		ON 
			r.name = cara.region_name 
			AND cara.cloudaccount_id = $1
		WHERE 
			(r.type = 'open' OR cara.cloudaccount_id IS NOT NULL)
			AND r.name != 'global'
	`
	queryResetDefaultRegion = "UPDATE region SET is_default = FALSE WHERE name != $1"
)

type RegionRepository struct {
	db *sql.DB
}

func NewRegionRepository(db *sql.DB) (*RegionRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}

	return &RegionRepository{
		db: db,
	}, nil
}

func (repo *RegionRepository) AddRegion(ctx context.Context, region *pb.AddRegionRequest) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionRepository.AddRegion").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	userEmail, err := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if err != nil {
		logger.Error(err, "failed to extract email")
		return fmt.Errorf("failed to extract email")
	}

	// If the request sets the region as default, check if the region type is "General Availability"
	if region.IsDefault {
		if region.Type != "open" {
			logger.Error(fmt.Errorf("region type is not 'open'"), "invalid region type for default")
			return status.Error(codes.InvalidArgument, "only regions with type 'open' can be set as default")
		}

		// Update all other regions to be non-default
		err = repo.ResetDefaultRegion(ctx, region.Name)
		if err != nil {
			logger.Error(err, "error resetting default region flag in database")
			return status.Errorf(codes.Internal, "resetting default region flag failed")
		}
	}

	_, err = repo.db.ExecContext(ctx, queryAddRegion,
		region.Name, region.FriendlyName, region.Type, region.Subnet,
		region.AvailabilityZone, region.Prefix, region.IsDefault, region.ApiDns, region.Vnet, userEmail,
	)
	if err != nil {
		logger.Error(err, "failed to execute add region query")
		return fmt.Errorf("failed to execute add region query")
	}
	return nil
}

func (repo *RegionRepository) UpdateRegion(ctx context.Context, region *pb.UpdateRegionRequest) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionRepository.UpdateRegion").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	query := queryUpdateRegionBase
	var params []interface{}
	var setClauses []string
	paramIndex := 1

	// If the request sets the region as default, check if the region type is "General Availability"
	if region.IsDefault != nil && *region.IsDefault {
		var regionType string
		query := "SELECT type FROM region WHERE name = $1"
		err := repo.db.QueryRowContext(ctx, query, region.Name).Scan(&regionType)
		if err != nil {
			logger.Error(err, "error fetching region type from database")
			return status.Errorf(codes.Internal, "fetching region type failed")
		}

		if regionType != "open" {
			logger.Error(fmt.Errorf("region type is not 'open'"), "invalid region type for default")
			return status.Error(codes.InvalidArgument, "only regions with type 'open' can be set as default")
		}

		// Update all other regions to be non-default
		err = repo.ResetDefaultRegion(ctx, region.Name)
		if err != nil {
			logger.Error(err, "error resetting default region flag in database")
			return status.Errorf(codes.Internal, "resetting default region flag failed")
		}
	}

	// Helper function to append set clauses and parameters
	appendSetClause := func(column string, value interface{}) {
		if value != nil {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, paramIndex))
			params = append(params, value)
			paramIndex++
		}
	}

	// Validate and append set clauses if they are present
	if region.FriendlyName != nil {
		appendSetClause("friendly_name", region.FriendlyName)
	}
	if region.Type != nil {
		appendSetClause("type", region.Type)
	}
	if region.Subnet != nil {
		appendSetClause("subnet", region.Subnet)
	}
	if region.AvailabilityZone != nil {
		appendSetClause("availability_zone", region.AvailabilityZone)
	}
	if region.Prefix != nil {
		appendSetClause("prefix", region.Prefix)
	}
	if region.IsDefault != nil {
		appendSetClause("is_default", region.IsDefault)
	}
	if region.ApiDns != nil {
		appendSetClause("api_dns", region.ApiDns)
	}
	if region.Vnet != nil {
		appendSetClause("vnet", region.Vnet)
	}

	userEmail, err := grpcutil.ExtractClaimFromCtx(ctx, false, grpcutil.EmailClaim)
	if err != nil {
		logger.Error(err, "failed to extract email")
		return fmt.Errorf("failed to extract email")
	}
	appendSetClause("admin_name", userEmail)

	// Always update the updated_at field
	setClauses = append(setClauses, "updated_at = NOW()")

	query += strings.Join(setClauses, ", ")
	query += fmt.Sprintf(" WHERE name = $%d", paramIndex)
	params = append(params, region.Name)

	_, err = repo.db.ExecContext(ctx, query, params...)
	if err != nil {
		logger.Error(err, "failed to execute update region query")
		return fmt.Errorf("failed to execute update region query")
	}
	return nil
}

func (repo *RegionRepository) DeleteRegion(ctx context.Context, name string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionRepository.DeleteRegion").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	_, err := repo.db.ExecContext(ctx, queryDeleteRegion, name)
	if err != nil {
		logger.Error(err, "failed to execute delete region query")
		return fmt.Errorf("failed to execute delete region query")
	}
	return nil
}

func (repo *RegionRepository) GetRegionByName(ctx context.Context, name string) (*pb.Region, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionRepository.GetRegionByName").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	row := repo.db.QueryRowContext(ctx, queryGetRegionByName, name)

	var region pb.Region
	var createdAt, updatedAt time.Time
	err := row.Scan(
		&region.Name, &region.FriendlyName, &region.Type, &region.Subnet,
		&region.AvailabilityZone, &region.Prefix, &region.IsDefault, &region.ApiDns, &region.Vnet,
		&region.AdminName, &createdAt, &updatedAt,
	)
	if err != nil {
		logger.Error(err, "failed to scan region by name")
		return nil, fmt.Errorf("failed to scan region by name")
	}
	region.CreatedAt = timestamppb.New(createdAt)
	region.UpdatedAt = timestamppb.New(updatedAt)
	return &region, nil
}

func (repo *RegionRepository) GetRegions(ctx context.Context, filter *pb.RegionFilter) ([]*pb.Region, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionRepository.GetRegions").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	var queryBuilder strings.Builder
	if _, err := queryBuilder.WriteString(queryGetRegionsBase); err != nil {
		logger.Error(err, "error writing base query string")
		return nil, fmt.Errorf("error writing base query string")
	}

	var params []interface{}
	paramIndex := 1

	// Helper function to append query and parameters
	appendFilter := func(condition string, value interface{}) {
		if value != nil {
			if _, err := fmt.Fprintf(&queryBuilder, " AND %s = $%d", condition, paramIndex); err != nil {
				logger.Error(err, "error writing filter condition", "condition", condition)
				return
			}
			params = append(params, value)
			paramIndex++
		}
	}

	// Validate and append filters if they are present
	if filter != nil {
		if filter.Name != nil {
			appendFilter("name", filter.Name)
		}
		if filter.FriendlyName != nil {
			appendFilter("friendly_name", filter.FriendlyName)
		}
		if filter.Type != nil {
			appendFilter("type", filter.Type)
		}
		if filter.Subnet != nil {
			appendFilter("subnet", filter.Subnet)
		}
		if filter.AvailabilityZone != nil {
			appendFilter("availability_zone", filter.AvailabilityZone)
		}
		if filter.Prefix != nil {
			appendFilter("prefix", filter.Prefix)
		}
		if filter.IsDefault != nil {
			appendFilter("is_default", filter.IsDefault)
		}
		if filter.ApiDns != nil {
			appendFilter("api_dns", filter.ApiDns)
		}
		if filter.AdminName != nil {
			appendFilter("admin_name", filter.AdminName)
		}
		if filter.Vnet != nil {
			appendFilter("vnet", filter.Vnet)
		}
	}

	query := queryBuilder.String()
	rows, err := repo.db.QueryContext(ctx, query, params...)
	logger.Info("Query", "query", query)
	logger.Info("Params", "params", params)
	if err != nil {
		logger.Error(err, "failed executing query", "query", query, "params", params)
		return nil, fmt.Errorf("failed executing query")
	}
	defer rows.Close()

	var regions []*pb.Region
	for rows.Next() {
		var createdAt, updatedAt time.Time
		var region pb.Region
		if err := rows.Scan(
			&region.Name, &region.FriendlyName, &region.Type, &region.Subnet,
			&region.AvailabilityZone, &region.Prefix, &region.IsDefault, &region.ApiDns, &region.Vnet,
			&region.AdminName, &createdAt, &updatedAt,
		); err != nil {
			logger.Error(err, "error scanning row")
			return nil, fmt.Errorf("error scanning row")
		}
		region.CreatedAt = timestamppb.New(createdAt)
		region.UpdatedAt = timestamppb.New(updatedAt)
		regions = append(regions, &region)
	}

	if err := rows.Err(); err != nil {
		logger.Error(err, "error iterating over rows")
		return nil, fmt.Errorf("error iterating over rows")
	}

	return regions, nil
}

func (repo *RegionRepository) GetUserRegions(ctx context.Context, filter *pb.RegionUserFilter) ([]*pb.RegionUser, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionRepository.GetUserRegions").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	var queryBuilder strings.Builder
	if _, err := queryBuilder.WriteString(queryGetUserRegionsBase); err != nil {
		logger.Error(err, "error writing base query string")
		return nil, fmt.Errorf("error writing base query string")
	}

	var params []interface{}
	params = append(params, filter.CloudaccountId)

	if filter.RegionFilter != nil {
		paramIndex := 2

		// Helper function to append query and parameters
		appendFilter := func(condition string, value interface{}) {
			if value != nil {
				if _, err := fmt.Fprintf(&queryBuilder, " AND %s = $%d", condition, paramIndex); err != nil {
					logger.Error(err, "error writing filter condition", "condition", condition)
					return
				}
				params = append(params, value)
				paramIndex++
			}
		}

		// Validate and append filters if they are present
		if filter.RegionFilter.Name != nil {
			appendFilter("r.name", filter.RegionFilter.Name)
		}
		if filter.RegionFilter.FriendlyName != nil {
			appendFilter("r.friendly_name", filter.RegionFilter.FriendlyName)
		}
		if filter.RegionFilter.Type != nil {
			appendFilter("r.type", filter.RegionFilter.Type)
		}
		if filter.RegionFilter.Subnet != nil {
			appendFilter("r.subnet", filter.RegionFilter.Subnet)
		}
		if filter.RegionFilter.AvailabilityZone != nil {
			appendFilter("r.availability_zone", filter.RegionFilter.AvailabilityZone)
		}
		if filter.RegionFilter.Prefix != nil {
			appendFilter("r.prefix", filter.RegionFilter.Prefix)
		}
		if filter.RegionFilter.IsDefault != nil {
			appendFilter("r.is_default", filter.RegionFilter.IsDefault)
		}
		if filter.RegionFilter.ApiDns != nil {
			appendFilter("r.api_dns", filter.RegionFilter.ApiDns)
		}
		if filter.RegionFilter.AdminName != nil {
			appendFilter("r.admin_name", filter.RegionFilter.AdminName)
		}
		if filter.RegionFilter.Vnet != nil {
			appendFilter("r.vnet", filter.RegionFilter.Vnet)
		}
	}

	query := queryBuilder.String()
	rows, err := repo.db.QueryContext(ctx, query, params...)
	if err != nil {
		logger.Error(err, "failed executing query", "query", query, "params", params)
		return nil, fmt.Errorf("failed executing query")
	}
	defer rows.Close()

	var regions []*pb.RegionUser
	for rows.Next() {
		var region pb.RegionUser
		if err := rows.Scan(
			&region.Name, &region.FriendlyName, &region.Type, &region.Subnet,
			&region.AvailabilityZone, &region.Prefix, &region.IsDefault, &region.ApiDns,
			&region.Vnet,
		); err != nil {
			logger.Error(err, "error scanning row")
			return nil, fmt.Errorf("error scanning row")
		}
		regions = append(regions, &region)
	}

	if err := rows.Err(); err != nil {
		logger.Error(err, "error iterating over rows")
		return nil, fmt.Errorf("error iterating over rows")
	}

	return regions, nil
}

func (repo *RegionRepository) ResetDefaultRegion(ctx context.Context, name string) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("RegionRepository.ResetDefaultRegion").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	_, err := repo.db.ExecContext(ctx, queryResetDefaultRegion, name)
	if err != nil {
		logger.Error(err, "failed to execute reset default region query")
		return fmt.Errorf("failed to execute reset default region query")
	}
	return nil
}
