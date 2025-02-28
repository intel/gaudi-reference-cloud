// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UsageRecordService struct {
	pb.UnimplementedUsageRecordServiceServer
	config          *Config
	session         *sql.DB
	usageRecordData *UsageRecordData
}

func NewUsageRecordService(config *Config, session *sql.DB) *UsageRecordService {
	return &UsageRecordService{
		config:          config,
		session:         session,
		usageRecordData: NewUsageRecordData(session),
	}
}

func (svc UsageRecordService) CreateProductUsageRecord(ctx context.Context, productUsageRecordCreate *pb.ProductUsageRecordCreate) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageRecordService.CreateProductUsageRecord").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// todo: add all the validation checks
	if utils.IsValidCloudAccountId(productUsageRecordCreate.CloudAccountId) ||
		utils.IsValidTransactionId(productUsageRecordCreate.TransactionId) ||
		utils.IsValidRegion(productUsageRecordCreate.Region) ||
		utils.IsValidTimestamp(productUsageRecordCreate.Timestamp) ||
		(productUsageRecordCreate.StartTime != nil && utils.IsValidTimestamp(productUsageRecordCreate.StartTime)) ||
		(productUsageRecordCreate.EndTime != nil && utils.IsValidTimestamp(productUsageRecordCreate.EndTime)) {
		logger.Error(errors.New(InvalidInputArguments), "invalid input arguments, ignoring product usage record creation")
		return nil, status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring product usage record creation.")
	}

	err := svc.usageRecordData.StoreProductUsageRecord(ctx, productUsageRecordCreate)

	if err != nil {
		logger.Error(err, "failed to create product usage record")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (svc UsageRecordService) SearchProductUsageRecords(f *pb.ProductUsageRecordsFilter, rs pb.UsageRecordService_SearchProductUsageRecordsServer) error {
	logger := log.FromContext(rs.Context()).WithName("UsageRecordService.SearchProductUsageRecords")

	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if (f.CloudAccountId != nil && utils.IsValidCloudAccountId(*f.CloudAccountId)) ||
		(f.Region != nil && utils.IsValidRegion(*f.Region)) ||
		(f.TransactionId != nil && utils.IsValidTransactionId(*f.TransactionId)) ||
		(f.StartTime != nil && utils.IsValidTimestamp(f.StartTime)) ||
		(f.EndTime != nil && utils.IsValidTimestamp(f.EndTime)) {
		logger.Error(errors.New(InvalidInputArguments), "invalid input arguments, ignoring product usage records search")
		return status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring product usage records search")
	}

	if err := svc.usageRecordData.SearchProductUsageRecords(rs, f); err != nil {
		logger.Error(err, "error completing product usage records search query with input filters")
		return err
	}

	return nil
}

func (svc UsageRecordService) SearchInvalidProductUsageRecords(f *pb.InvalidProductUsageRecordsFilter, rs pb.UsageRecordService_SearchInvalidProductUsageRecordsServer) error {
	logger := log.FromContext(rs.Context()).WithName("UsageRecordService.SearchInvalidProductUsageRecords")

	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if (f.CloudAccountId != nil && utils.IsValidCloudAccountId(*f.CloudAccountId)) ||
		(f.RecordId != nil && utils.IsValidRegion(*f.RecordId)) ||
		(f.Region != nil && utils.IsValidRegion(*f.Region)) ||
		(f.TransactionId != nil && utils.IsValidTransactionId(*f.TransactionId)) ||
		(f.StartTime != nil && utils.IsValidTimestamp(f.StartTime)) ||
		(f.EndTime != nil && utils.IsValidTimestamp(f.EndTime)) {
		logger.Error(errors.New(InvalidInputArguments), "invalid input arguments, ignoring invalid product usage records search")
		return status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring invalid product usage records search")
	}

	if err := svc.usageRecordData.SearchInvalidProductUsageRecords(rs, f); err != nil {
		logger.Error(err, "error completing invalid product usage records search query with input filters")
		return err
	}

	return nil
}

func (svc UsageRecordService) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("UsageRecordService.Ping")
	logger.Info("Ping")
	return &emptypb.Empty{}, nil
}
