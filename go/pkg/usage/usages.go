// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const InvalidInputArguments = "INVALID_INPUT_ARGS"

type UsageService struct {
	pb.UnimplementedUsageServiceServer
	config    *Config
	session   *sql.DB
	usageData *UsageData
}

func NewUsageService(config *Config, session *sql.DB) *UsageService {
	return &UsageService{
		config:    config,
		session:   session,
		usageData: NewUsageData(session),
	}
}

func (svc UsageService) PostBulkUploadResourceUsages(ctx context.Context, bulkUploadResourceUsages *pb.BulkUploadResourceUsages) (*pb.BulkUploadResourceUsagesFailed, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.PostBulkUploadResourceUsages").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	bulkUploadResourceUsagesFailed, err := svc.usageData.BulkUpload(ctx, bulkUploadResourceUsages)

	if err != nil {
		return bulkUploadResourceUsagesFailed, status.Errorf(codes.Internal, err.Error())
	}

	return bulkUploadResourceUsagesFailed, nil

}

func (svc UsageService) CreateResourceUsage(ctx context.Context, resourceUsageCreate *pb.ResourceUsageCreate) (*pb.ResourceUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.CreateResourceUsage").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// todo: add all the validation checks
	if utils.IsValidCloudAccountId(resourceUsageCreate.CloudAccountId) ||
		utils.IsValidResourceId(resourceUsageCreate.ResourceId) {
		logger.Error(errors.New(InvalidInputArguments), "invalid input arguments, ignoring usage creation")
		return nil, status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring usage creation.")
	}

	resourceUsage, err := svc.usageData.StoreResourceUsage(ctx, resourceUsageCreate, time.Now())

	if err != nil {
		logger.Error(err, "failed to create resource usage")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	err = svc.usageData.StoreProductUsage(ctx, &ProductUsageCreate{
		Id:             uuid.NewString(),
		CloudAccountId: resourceUsageCreate.CloudAccountId,
		ProductId:      resourceUsageCreate.ProductId,
		ProductName:    resourceUsageCreate.ProductName,
		Region:         resourceUsageCreate.Region,
		Quantity:       resourceUsageCreate.Quantity,
		Rate:           resourceUsageCreate.Rate,
		UsageUnitType:  resourceUsageCreate.UsageUnitType,
		StartTime:      resourceUsageCreate.StartTime.AsTime(),
		EndTime:        resourceUsageCreate.EndTime.AsTime(),
	}, time.Now())

	if err != nil {
		logger.Error(err, "failed to create product usage as a part of creating resource usage")
		err = svc.usageData.DeleteResourceUsage(ctx, resourceUsage.Id)
		if err != nil {
			logger.Error(err, "failed to delete resource usage upon failed creation of product usage for",
				"id", resourceUsage.Id)
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return resourceUsage, nil
}

func (svc UsageService) UpdateResourceUsage(ctx context.Context, resourceUsageRecordUpdate *pb.ResourceUsageUpdate) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.UpdateResourceUsage").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	err := svc.usageData.UpdateResourceUsage(ctx, resourceUsageRecordUpdate.ResourceUsageId,
		&ResourceUsageUpdate{UnReportedQuantity: resourceUsageRecordUpdate.UnReportedQuantity})

	if err != nil {
		logger.Error(err, "failed to update resource usage")
		return &emptypb.Empty{}, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (svc UsageService) GetResourceUsageById(ctx context.Context, resourceUsageId *pb.ResourceUsageId) (*pb.ResourceUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.GetResourceUsageById").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	resourceUsageRecord, err := svc.usageData.GetResourceUsageById(ctx, resourceUsageId.Id)

	if err != nil {
		logger.Error(err, "failed to get resource usage by id")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return resourceUsageRecord, nil
}

func (svc UsageService) MarkResourceUsageAsReported(ctx context.Context, resourceUsageId *pb.ResourceUsageId) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.MarkResourceUsageAsReported").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	err := svc.usageData.MarkResourceUsageAsReported(ctx, resourceUsageId.Id)

	if err != nil {
		logger.Error(err, "failed to mark resource usage as reported")
		return &emptypb.Empty{}, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (svc UsageService) GetProductUsageById(ctx context.Context, productUsageId *pb.ProductUsageId) (*pb.ProductUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.GetProductUsageById").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	productUsageRecord, err := svc.usageData.GetProductUsageById(ctx, productUsageId.Id)

	if err != nil {
		logger.Error(err, "failed to get product usage by id")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return productUsageRecord, nil
}

func (svc UsageService) SearchResourceUsages(ctx context.Context, resourceUsagesFilter *pb.ResourceUsagesFilter) (*pb.ResourceUsages, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.SearchResourceUsages").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// todo: add all the validation checks
	if (resourceUsagesFilter.CloudAccountId != nil && utils.IsValidCloudAccountId(*resourceUsagesFilter.CloudAccountId)) ||
		(resourceUsagesFilter.ResourceId != nil && utils.IsValidResourceId(*resourceUsagesFilter.ResourceId)) ||
		(resourceUsagesFilter.StartTime != nil && utils.IsValidTimestamp(resourceUsagesFilter.StartTime)) ||
		(resourceUsagesFilter.EndTime != nil && utils.IsValidTimestamp(resourceUsagesFilter.EndTime)) {
		logger.Error(errors.New(InvalidInputArguments), "invalid search arguments")
		return nil, status.Errorf(codes.InvalidArgument, "invalid search arguments")
	}

	resourceUsages, err := svc.usageData.SearchResourceUsages(ctx, resourceUsagesFilter)

	if err != nil {
		logger.Error(err, "failed to get search resource usages")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return resourceUsages, nil
}

func (svc UsageService) StreamSearchResourceUsages(resourceUsagesFilter *pb.ResourceUsagesFilter, resourceUsagesStream pb.UsageService_StreamSearchResourceUsagesServer) error {
	ctx, logger, span := obs.LogAndSpanFromContext(resourceUsagesStream.Context()).WithName("UsageService.StreamSearchResourceUsages").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	if (resourceUsagesFilter.CloudAccountId != nil && utils.IsValidCloudAccountId(*resourceUsagesFilter.CloudAccountId)) ||
		(resourceUsagesFilter.ResourceId != nil && utils.IsValidResourceId(*resourceUsagesFilter.ResourceId)) ||
		(resourceUsagesFilter.StartTime != nil && utils.IsValidTimestamp(resourceUsagesFilter.StartTime)) ||
		(resourceUsagesFilter.EndTime != nil && utils.IsValidTimestamp(resourceUsagesFilter.EndTime)) {
		logger.Error(errors.New(InvalidInputArguments), "invalid search arguments")
		return status.Errorf(codes.InvalidArgument, "invalid search arguments")
	}
	if err := svc.usageData.SendResourceUsages(ctx, resourceUsagesFilter, resourceUsagesStream); err != nil {
		logger.Error(err, "failed to send search resource usages")
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func (svc UsageService) SearchUsages(ctx context.Context, usagesFilter *pb.UsagesFilter) (*pb.Usages, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.SearchUsages").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	// todo: add all the validation checks
	if (usagesFilter.CloudAccountId != nil && utils.IsValidCloudAccountId(*usagesFilter.CloudAccountId)) ||
		(usagesFilter.Region != nil && utils.IsValidRegion(*usagesFilter.Region)) ||
		(usagesFilter.StartTime != nil && utils.IsValidTimestamp(usagesFilter.StartTime)) ||
		(usagesFilter.EndTime != nil && utils.IsValidTimestamp(usagesFilter.EndTime)) {
		logger.Error(errors.New(InvalidInputArguments), "invalid search arguments")
		return nil, status.Errorf(codes.InvalidArgument, "invalid search arguments")
	}

	logger.Info("get usage times:", "startTimeValue", usagesFilter.StartTime.AsTime().Format("January 2, 2006"),
		"endTimeValue", usagesFilter.EndTime.AsTime().Format("January 2, 2006"))
	productUsagesFilter := &pb.ProductUsagesFilter{
		CloudAccountId: usagesFilter.CloudAccountId,
		Region:         usagesFilter.Region,
		StartTime:      usagesFilter.StartTime,
		EndTime:        usagesFilter.EndTime,
	}

	productUsages, err := svc.usageData.SearchProductUsages(ctx, productUsagesFilter)

	if err != nil {
		logger.Error(err, "failed to get search product usages")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	usageMap := map[string]*pb.UsageV2{}

	for _, productUsageRecord := range productUsages.ProductUsages {

		// this is because the same product can exist in multiple regions and the usage is scoped to the region.
		usageMapKey := fmt.Sprintf("%s:%s", productUsageRecord.Region, productUsageRecord.ProductId)
		if usage, ok := usageMap[usageMapKey]; ok {
			usage.Amount += productUsageRecord.Quantity * productUsageRecord.Rate
			usage.Quantity += productUsageRecord.Quantity
			if usage.Start.AsTime().After(productUsageRecord.StartTime.AsTime()) {
				usage.Start = productUsageRecord.StartTime
			}
			if usage.End.AsTime().Before(productUsageRecord.EndTime.AsTime()) {
				usage.End = productUsageRecord.EndTime
			}
		} else {
			usageMap[usageMapKey] = &pb.UsageV2{
				// service name needs to be added.
				ServiceName:   "",
				ProductName:   productUsageRecord.ProductName,
				Start:         productUsageRecord.StartTime,
				End:           productUsageRecord.EndTime,
				Quantity:      productUsageRecord.Quantity,
				Amount:        productUsageRecord.Quantity * productUsageRecord.Rate,
				Rate:          productUsageRecord.Rate,
				Region:        productUsageRecord.Region,
				UsageUnitType: productUsageRecord.UsageUnitType,
			}
		}
	}

	usages := &pb.Usages{}

	periodStart := time.Now()
	periodEnd := time.Now()

	for _, usageR := range usageMap {
		if periodStart.After(usageR.Start.AsTime()) {
			periodStart = usageR.Start.AsTime()
		}
		if periodEnd.Before(usageR.End.AsTime()) {
			periodEnd = usageR.End.AsTime()
		}
		usages.TotalQuantity += usageR.Quantity
		usages.TotalAmount += usageR.Amount
		usages.Usages = append(usages.Usages, usageR)
	}

	usages.Period = fmt.Sprintf("%s, %d", periodStart.Month().String(), periodStart.Year())
	usages.LastUpdated = timestamppb.New(periodEnd)

	return usages, nil
}

func (svc UsageService) SearchProductUsages(ctx context.Context, productUsagesFilter *pb.ProductUsagesFilter) (*pb.ProductUsages, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.SearchProductUsages").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// todo: add all the validation checks
	if (productUsagesFilter.CloudAccountId != nil && utils.IsValidCloudAccountId(*productUsagesFilter.CloudAccountId)) ||
		(productUsagesFilter.StartTime != nil && utils.IsValidTimestamp(productUsagesFilter.StartTime)) ||
		(productUsagesFilter.EndTime != nil && utils.IsValidTimestamp(productUsagesFilter.EndTime)) {
		logger.Error(errors.New(InvalidInputArguments), "invalid search arguments")
		return nil, status.Errorf(codes.InvalidArgument, "invalid search arguments")
	}

	productUsages, err := svc.usageData.SearchProductUsages(ctx, productUsagesFilter)

	if err != nil {
		logger.Error(err, "failed to get search product usages")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return productUsages, nil
}

func (svc UsageService) StreamSearchProductUsages(productUsagesFilter *pb.ProductUsagesFilter, productUsagesStream pb.UsageService_StreamSearchProductUsagesServer) error {
	ctx, logger, span := obs.LogAndSpanFromContext(productUsagesStream.Context()).WithName("UsageService.StreamSearchProductUsages").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	// todo: add all the validation checks
	if (productUsagesFilter.CloudAccountId != nil && utils.IsValidCloudAccountId(*productUsagesFilter.CloudAccountId)) ||
		(productUsagesFilter.StartTime != nil && utils.IsValidTimestamp(productUsagesFilter.StartTime)) ||
		(productUsagesFilter.EndTime != nil && utils.IsValidTimestamp(productUsagesFilter.EndTime)) {
		logger.Error(errors.New(InvalidInputArguments), "invalid search arguments")
		return status.Errorf(codes.InvalidArgument, "invalid search arguments")
	}

	if err := svc.usageData.SendProductUsages(ctx, productUsagesFilter, productUsagesStream); err != nil {
		logger.Error(err, "failed to send product usages")
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func (svc UsageService) Ping(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.Ping").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	return &emptypb.Empty{}, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

func (svc UsageService) SearchProductUsagesReport(productUsagesReportFilter *pb.ProductUsagesReportFilter, rs pb.UsageService_SearchProductUsagesReportServer) error {
	_, logger, span := obs.LogAndSpanFromContext(rs.Context()).WithName("UsageService.SearchProductUsagesReport").Start()

	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if (productUsagesReportFilter.CloudAccountId != nil && utils.IsValidCloudAccountId(*productUsagesReportFilter.CloudAccountId)) ||
		(productUsagesReportFilter.StartTime != nil && utils.IsValidTimestamp(productUsagesReportFilter.StartTime)) ||
		(productUsagesReportFilter.EndTime != nil && utils.IsValidTimestamp(productUsagesReportFilter.EndTime)) {
		logger.Error(errors.New(InvalidInputArguments), "invalid search arguments")
		return status.Errorf(codes.InvalidArgument, "invalid search arguments")
	}

	if err := svc.usageData.SearchProductUsagesReport(productUsagesReportFilter, rs); err != nil {
		logger.Error(err, "error completing search query for product usages unreported")
		return err
	}

	return nil
}

func (svc UsageService) MarkProductUsageAsReported(ctx context.Context, reportProductUsageId *pb.ReportProductUsageId) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.MarkProductUsageAsReported").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	err := svc.usageData.MarkProductUsageAsReported(ctx, reportProductUsageId.Id)

	if err != nil {
		logger.Error(err, "failed to mark product usage as reported")
		return &emptypb.Empty{}, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (svc UsageService) UpdateProductUsageReport(ctx context.Context, reportProductUsageUpdate *pb.ReportProductUsageUpdate) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageService.UpdateProductUsageReport").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	err := svc.usageData.UpdateProductUsageReport(ctx, reportProductUsageUpdate)

	if err != nil {
		logger.Error(err, "failed to update product usage report")
		return &emptypb.Empty{}, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
