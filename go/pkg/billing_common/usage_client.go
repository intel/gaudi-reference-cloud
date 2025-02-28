// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
)

type UsageSvcClient struct {
	UsageClient pb.UsageServiceClient
}

type BillingUsage struct {
	CloudAccountId   string
	ProductId        string
	TransactionId    string
	ResourceId       string
	ResourceName     string
	Quantity         float64
	Amount           float64
	UsageRate        float64
	AmountUnReported float64
	UsageId          string
	RegionName       string
	UsageDate        time.Time
	UsageUnitType    string
}

func NewUsageServiceClient(ctx context.Context, resolver grpcutil.Resolver) (*UsageSvcClient, error) {
	logger := log.FromContext(ctx).WithName("NewUsageServiceClient")
	var usageConn *grpc.ClientConn
	usageAddr, err := resolver.Resolve(ctx, "usage")
	if err != nil {
		logger.Error(err, "grpc resolver not able to resolve", "addr", usageAddr)
		return nil, err
	}
	usageConn, err = grpcConnect(ctx, usageAddr)
	if err != nil {
		return nil, err
	}
	usage := pb.NewUsageServiceClient(usageConn)
	return &UsageSvcClient{UsageClient: usage}, nil
}

func (usageSvc *UsageSvcClient) GetResourceUsageOverStream(ctx context.Context, cloudAcct *pb.CloudAccount) (*pb.ResourceUsages, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageSvcClient.GetResourceUsageOverStream").WithValues("cloudAccountId", cloudAcct.GetId()).Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	reported := false
	resourceUsagesFilter := &pb.ResourceUsagesFilter{
		CloudAccountId: &cloudAcct.Id,
		Reported:       &reported,
	}
	resourceUsagesStream, err := usageSvc.UsageClient.StreamSearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		logger.Error(err, "failed to get resource usages for cloud accounts", "cloudAcct.Id", cloudAcct.Id)
		return nil, err
	}
	resourceUsages := &pb.ResourceUsages{}
	for {
		resourceUsage, err := resourceUsagesStream.Recv()
		if err == io.EOF {
			logger.Info("resourceUsagesStream eof")
			break
		}

		if err != nil {
			logger.Error(err, "failed to receive resource usages for cloud account", "cloudAcct.Id", cloudAcct.Id)
			return nil, err
		}
		logger.V(9).Info("received unreported resource usage", "usageId", resourceUsage.Id)
		resourceUsages.ResourceUsages = append(resourceUsages.ResourceUsages, resourceUsage)
	}

	return resourceUsages, nil
}

func (usageSvc *UsageSvcClient) GetProductUsageReportOverStream(ctx context.Context, cloudAcct *pb.CloudAccount) ([]*pb.ReportProductUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.GetProductUsageReportOverStream").WithValues("cloudAccountId", cloudAcct.GetId()).Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	reported := false
	resourceUsagesFilter := &pb.ProductUsagesReportFilter{
		CloudAccountId: &cloudAcct.Id,
		Reported:       &reported,
	}

	productUsagesStream, err := usageSvc.UsageClient.SearchProductUsagesReport(ctx, resourceUsagesFilter)
	if err != nil {
		logger.Error(err, "unable to SearchProductUsagesUnreported")
		return nil, err
	}
	productUsagesToReport := []*pb.ReportProductUsage{}
	for {
		reportProductUsage, err := productUsagesStream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			logger.Error(err, "received error in SearchProductUsagesUnreported")
			return nil, err
		}
		logger.V(9).Info("received unreported product usage", "usageId", reportProductUsage.Id)
		productUsagesToReport = append(productUsagesToReport, reportProductUsage)
	}

	return productUsagesToReport, nil
}

func (usageSvc *UsageSvcClient) UpdateAndMarkProductRecordsReported(ctx context.Context, productUsagesReported []*BillingUsage, cloudAcct *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageSvcClient.UpdateAndMarkProductRecordsReported").Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	for _, billingProductUsageReported := range productUsagesReported {
		if billingProductUsageReported.AmountUnReported == 0 {
			_, err := usageSvc.UsageClient.MarkProductUsageAsReported(ctx, &pb.ReportProductUsageId{Id: billingProductUsageReported.UsageId})
			if err != nil {
				logger.Error(err, "failed to mark usage as reported for ", "usageId", billingProductUsageReported.UsageId, "cloudAccountId", billingProductUsageReported.CloudAccountId)
			}
		} else {
			_, err := usageSvc.UsageClient.UpdateProductUsageReport(ctx,
				&pb.ReportProductUsageUpdate{
					ProductUsageReportId: billingProductUsageReported.UsageId,
					UnReportedQuantity:   billingProductUsageReported.AmountUnReported / billingProductUsageReported.UsageRate,
				})
			if err != nil {
				logger.Error(err, "failed to update the unreported usage for ", "usageId", billingProductUsageReported.UsageId, "cloudAccountId", billingProductUsageReported.CloudAccountId)
			}
		}
	}
	return nil
}

func (usageSvc *UsageSvcClient) UpdateAndMarkResourceRecordsReported(ctx context.Context, billingUsage []*BillingUsage, cloudAcct *pb.CloudAccount) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("UsageSvcClient.UpdateAndMarkResourceRecordsReported").Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	for _, billingUsageReported := range billingUsage {
		if billingUsageReported.AmountUnReported == 0 {
			_, err := usageSvc.UsageClient.MarkResourceUsageAsReported(ctx, &pb.ResourceUsageId{Id: billingUsageReported.UsageId})
			if err != nil {
				logger.Error(err, "failed to mark resource usage as reported for ", "usageId", billingUsageReported.UsageId, "cloudAccountId", cloudAcct.GetId())
			}
		} else {
			_, err := usageSvc.UsageClient.UpdateResourceUsage(ctx,
				&pb.ResourceUsageUpdate{
					ResourceUsageId:    billingUsageReported.UsageId,
					UnReportedQuantity: billingUsageReported.AmountUnReported / billingUsageReported.UsageRate,
				})
			if err != nil {
				logger.Error(err, "failed to update the unreported resource usage for ", "usageId", billingUsageReported.UsageId, "cloudAccountId", cloudAcct.GetId())
			}
		}
	}
	return nil
}
