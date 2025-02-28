// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package sync

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/client-go/rest"
)

type ProductSyncService struct {
	pb.UnimplementedProductSyncServiceServer
	restClient     *rest.RESTClient
	dbClient       *sql.DB
	defaultRegions []pb.DefaultRegionSpec
}

func NewProductSyncService(restClient *rest.RESTClient, db *sql.DB, defaultRegions []pb.DefaultRegionSpec) (*ProductSyncService, error) {
	if restClient == nil {
		return nil, fmt.Errorf("k8s client is required")
	}

	return &ProductSyncService{
		restClient:     restClient,
		dbClient:       db,
		defaultRegions: defaultRegions,
	}, nil
}

func (srv *ProductSyncService) Put(ctx context.Context, product *pb.DefaultProduct) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductSyncService.Put").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	if err := product.Validate(); err != nil {
		logger.Error(err, "validation failed for product")
		return nil, status.Error(codes.InvalidArgument, "validation failed for product")
	}

	// Begin a transaction
	tx, err := srv.dbClient.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error beginning product sync transaction")
		return nil, status.Errorf(codes.Internal, "error beginning product sync transaction: %v", err)
	}

	// Upsert service
	err = insertRegions(ctx, tx, srv.defaultRegions)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error(rbErr, "error rolling back transaction after inserting regions failure")
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error committing product sync transaction")
		return nil, status.Errorf(codes.Internal, "error committing product sync transaction: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func insertRegions(ctx context.Context, tx *sql.Tx, defaultRegions []pb.DefaultRegionSpec) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductSyncService.insertRegion").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	for i := range defaultRegions {
		region := &defaultRegions[i]

		// Ensure api_dns has protocol and path
		region.ApiDns = ensureProtocolAndPath(region.ApiDns)

		// Check if the region exists in the region table
		var regionExists bool
		err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM region WHERE name = $1)", region.Name).Scan(&regionExists)
		if err != nil {
			logger.Info("region already exists in data base", "region name", region.Name)
			continue
		}

		// If the region does not exist, insert it
		if !regionExists {
			insertRegionQuery := `
			INSERT INTO region (name, friendly_name, type, subnet, availability_zone, prefix, is_default, api_dns, vnet, admin_name) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`
			vnetValue := region.Name + "-default"

			logger.Info("executing", "query", insertRegionQuery)
			logger.Info("with arguments", "name", region.Name)

			if _, err := tx.ExecContext(ctx, insertRegionQuery, region.Name, region.FriendlyName, region.Type, region.Subnet, region.AvailabilityZone, region.Prefix, region.IsDefault, region.ApiDns, vnetValue, "system"); err != nil {
				logger.Error(err, "error upserting rate set")
				return status.Errorf(codes.Internal, "error upserting rate set")
			}
		}
	}
	return nil
}

func ensureProtocolAndPath(apiDns string) string {
	if apiDns == "" || apiDns == "NA" {
		return apiDns
	}

	// Check if the URL already has a protocol
	if !strings.HasPrefix(apiDns, "http://") && !strings.HasPrefix(apiDns, "https://") {
		apiDns = "https://" + apiDns // Default to https
	}

	// Check if the URL already has a path
	if !strings.HasSuffix(apiDns, "/v1") {
		apiDns = strings.TrimRight(apiDns, "/") + "/v1"
	}

	return apiDns
}

func (srv *ProductSyncService) SearchStream(req *pb.DefaultProductSearchRequest, stream pb.ProductSyncService_SearchStreamServer) error {
	ctx := stream.Context()
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductSyncService.SearchStream").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// Prepare the SQL query to select all products with the new fields
	productQuery := `
		SELECT 
			DISTINCT ON (id) id, name
		FROM 
			region 
		ORDER BY id,  name ASC
	`

	// Execute the product query
	rows, err := srv.dbClient.QueryContext(ctx, productQuery)
	if err != nil {
		logger.Error(err, "error querying product records")
		return status.Errorf(codes.Internal, "error querying product records: %v", err)
	}
	defer rows.Close()

	// Map to hold products and their associated metadata and rates
	productMap := make(map[string]*pb.DefaultProduct)

	for rows.Next() {
		var (
			productId, productName string
		)

		if err := rows.Scan(&productId, &productName); err != nil {
			logger.Error(err, "error scanning product record")
			return status.Errorf(codes.Internal, "error scanning product record: %v", err)
		}

		// Check if the product already exists in the map
		product, exists := productMap[productId]
		if !exists {
			product = &pb.DefaultProduct{
				Metadata: &pb.DefaultProduct_Metadata{
					Name: productName,
				},
			}
			productMap[productId] = product
		}
	}

	// Stream the products to the client
	for _, product := range productMap {
		if err := stream.Send(product); err != nil {
			logger.Error(err, "error sending product record")
			return status.Errorf(codes.Internal, "error sending product record: %v", err)
		}
	}

	return nil
}

func (srv *ProductSyncService) Delete(ctx context.Context, product *pb.DefaultProductDeleteRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductSyncService.Delete").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	return nil, nil
}
