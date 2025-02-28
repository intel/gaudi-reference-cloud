// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ProductAccessService struct {
	pb.UnimplementedProductAccessServiceServer
	dbClient *sql.DB
}

var fieldOpts []protodb.FieldOptions = []protodb.FieldOptions{
	{Name: "cloudaccountId", StoreEmptyStringAsNull: false},
	{Name: "productId", StoreEmptyStringAsNull: false},
	{Name: "vendorId", StoreEmptyStringAsNull: false},
	{Name: "familyId", StoreEmptyStringAsNull: false},
	{Name: "adminName", StoreEmptyStringAsNull: false},
	{Name: "created", StoreEmptyStringAsNull: true},
}

func NewProductAccessService(db *sql.DB) (*ProductAccessService, error) {
	return &ProductAccessService{
		dbClient: db,
	}, nil
}

func (srv *ProductAccessService) ReadAccess(ctx context.Context, access *pb.GetAccessRequest) (*pb.GetAccessResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductAccessService.ReadAccess").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	obj := pb.GetAccessResponse{}
	err := func() error {
		var rows *sql.Rows
		var err error
		req := pb.ProductAccessRequest{}
		filterParams := protodb.NewProtoToSql(&req, fieldOpts...)
		readParams := protodb.NewSqlToProto(&req, fieldOpts...)
		query := fmt.Sprintf("SELECT %v FROM cloud_accounts_product_access %v", readParams.GetNamesString(), filterParams.GetFilter())
		rows, err = srv.dbClient.QueryContext(ctx, query)
		if err != nil {
			logger.Error(err, "failed to read cloud account prodcut access ", "filterParams", filterParams, "readParams", readParams, "context", "QueryContext")
			return err
		}

		defer rows.Close()

		for rows.Next() {
			resp := pb.ProductAccessResponse{}
			var created time.Time

			if err := rows.Scan(&resp.CloudaccountId, &resp.VendorId, &resp.FamilyId, &resp.ProductId, &resp.AdminName, &created); err != nil {
				return err
			}
			resp.Created = timestamppb.New(created)
			obj.Acl = append(obj.Acl, &resp)
		}
		return err
	}()
	return &obj, err
}

func (srv *ProductAccessService) AddAccess(ctx context.Context, request *pb.ProductAccessRequest) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductAccessService.AddAccess").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	params := protodb.NewProtoToSql(request, fieldOpts...)
	vals := params.GetValues()
	tx, err := srv.dbClient.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error starting db transaction", "params", params, "vals", vals, "context", "BeginTx")
		return nil, err
	}
	defer tx.Rollback()
	// Execute insert query
	query := fmt.Sprintf("INSERT INTO cloud_accounts_product_access (%v) VALUES(%v)", params.GetNamesString(), params.GetParamsString())
	logger.V(9).Info(query)
	if _, err = tx.ExecContext(ctx, query, vals...); err != nil {
		logger.Error(err, "error inserting account into db", "query", query, "context", "ExecContext")
		return nil, status.Errorf(codes.Internal, "access record insertion failed")
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err, "error committing db transaction", "params", params, "vals", vals, "context", "Commit")
		return nil, err
	}
	logger.Info("Database transaction completed")
	return &emptypb.Empty{}, nil
}

func (srv *ProductAccessService) RemoveAccess(ctx context.Context, request *pb.DeleteAccessRequest) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductAccessService.RemoveAccess").WithValues("cloudAccountId", request.CloudaccountId, "productId", request.ProductId, "vendorId", request.VendorId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	logger.Info("Delete request for :", "cloudaccountid", "productId", "vendorId", request.CloudaccountId, request.ProductId, request.VendorId)
	queryDelete := `
			delete from cloud_accounts_product_access where cloudaccount_id = $1 AND product_id = $2 AND vendor_id = $3
			`
	_, err := srv.dbClient.ExecContext(ctx, queryDelete, request.CloudaccountId, request.ProductId, request.VendorId)
	if err != nil {
		logger.Error(err, "Failed to delete Cloud Account with productId, vendorId invoked", "cloudAccountId", "productId", "vendorId", request.CloudaccountId, request.ProductId, request.VendorId)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (srv *ProductAccessService) CheckProductAccess(ctx context.Context, request *pb.ProductAccessCheckRequest) (*wrapperspb.BoolValue, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductAccessService.CheckProductAccess").Start()
	defer span.End()
	logger.V(9).Info("begin")
	defer logger.V(9).Info("end")

	logger.Info("checking product access", "cloudAccountId", request.CloudaccountId, "productId", request.ProductId)

	query := `
        SELECT 1
        FROM cloud_accounts_product_access
        WHERE cloudaccount_id = $1 AND product_id = $2
        LIMIT 1
    `

	var exists bool
	err := srv.dbClient.QueryRowContext(ctx, query, request.CloudaccountId, request.ProductId).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info("no access found for cloudaccount_id and product_id", "cloudAccountId", request.CloudaccountId, "productId", request.ProductId)
			return &wrapperspb.BoolValue{Value: false}, nil
		}
		logger.Error(err, "failed to check cloud account product access", "cloudAccountId", request.CloudaccountId, "productId", request.ProductId)
		return nil, err
	}

	return &wrapperspb.BoolValue{Value: true}, nil
}
