// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	productFamilyDescriptionKey = "product.family.description"
	usageQuantityUnitKey        = "usage.quantity.unit"
	usageUnitKey                = "usage.unit"
)

type BillingUsageService struct {
	usageServiceClient pb.UsageServiceClient
	pb.UnimplementedBillingUsageServiceServer
}

func (s *BillingUsageService) Read(ctx context.Context, in *pb.BillingUsageFilter) (*pb.BillingUsageResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingUsageService.Read").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	billingUsageResponse := &pb.BillingUsageResponse{}

	startTime, endTime, err := billingCommon.DetermineUsageDates(ctx, in.SearchStart, in.SearchEnd)

	if err != nil {
		logger.Error(err, "failed to determine start time and end time for retrieving usages")
		return nil, status.Errorf(codes.Internal, GetBillingError(FailedToParseFilter, err).Error())
	}

	usages, err := s.usageServiceClient.SearchUsages(ctx,
		&pb.UsagesFilter{
			CloudAccountId: &in.CloudAccountId,
			StartTime:      startTime,
			EndTime:        endTime,
			Region:         in.RegionName,
		})

	if err != nil {
		logger.Error(err, "failed to get usages")
		return nil, status.Errorf(codes.Internal, GetBillingError(FailedToGetUsages, err).Error())
	}

	for _, usage := range usages.Usages {
		billingUsage := &pb.BillingUsage{
			// need to add service name
			ServiceName: "",
			// product type is not supported - we should change it to product name.
			ProductType: usage.ProductName,
			Start:       usage.Start,
			End:         usage.End,
			// we are going to soon stop supporting mins used and change it to quantity.
			MinsUsed:   usage.Quantity,
			Amount:     usage.Amount,
			Rate:       fmt.Sprintf("%g", usage.Rate),
			RegionName: usage.Region,
		}

		if Cfg.GetFeaturesBillingUsageMetrics() {
			filter := pb.ProductFilter{Name: &usage.ProductName}
			products, err := productClient.GetProductCatalogProductsWithFilter(ctx, &filter)
			if err != nil && len(products) != 1 {
				logger.Error(err, "error reading product catalog entries")
				return nil, status.Errorf(codes.Internal, GetBillingError(FailedToGetUsagesProduct, err).Error())
			}
			if len(products) == 0 || products[0] == nil || products[0].Metadata == nil {
				logger.Error(err, "error getting usage product from product catalog entries")
				continue
			}
			productFamily := ""
			if tProductFamily, ok := products[0].Metadata[productFamilyDescriptionKey]; ok {
				productFamily = tProductFamily
			}
			billingUsage.ProductFamily = productFamily
			usageQuantityUnitName := ""
			if quantityUnit, ok := products[0].Metadata[usageQuantityUnitKey]; ok {
				usageQuantityUnitName = quantityUnit
			}
			//To support old record
			usageUnitType := usage.UsageUnitType
			if usageUnit, ok := products[0].Metadata[usageUnitKey]; ok {
				usageUnitType = usageUnit
			}
			billingUsageMetrics := &pb.BillingUsageMetrics{
				UsageQuantity: usage.Quantity,
				// change if type is different from name for UI
				UsageUnitName:         usageUnitType,
				UsageQuantityUnitName: usageQuantityUnitName,
			}
			billingUsage.BillingUsageMetrics = billingUsageMetrics
		}

		billingUsageResponse.Usages = append(billingUsageResponse.Usages, billingUsage)
	}

	billingUsageResponse.TotalAmount = usages.TotalAmount
	billingUsageResponse.TotalUsage = usages.TotalQuantity
	billingUsageResponse.LastUpdated = usages.LastUpdated
	billingUsageResponse.Period = usages.Period

	return billingUsageResponse, nil
}
