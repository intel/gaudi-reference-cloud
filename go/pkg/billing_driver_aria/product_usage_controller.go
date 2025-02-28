// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	productUsageControllerLinkChannel = make(chan bool)
	productUsageControllerTicker      *time.Ticker
)

func StartProductUsageReportScheduler(ctx context.Context, usageController *UsageController) {
	_, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.StartProductUsageReportScheduler").Start()
	logger.Info("BEGIN")
	defer logger.Info("END")
	productUsageControllerTicker = time.NewTicker(time.Duration(config.Cfg.ReportProductUsageSchedulerInterval) * time.Second)
	go reportProductUsageLoop(context.Background(), usageController)
}

func StopProductReportUsageScheduler() {
	if productUsageControllerLinkChannel != nil {
		close(productUsageControllerLinkChannel)
		productUsageControllerLinkChannel = nil
	}
}

func reportProductUsageLoop(ctx context.Context, usageController *UsageController) {
	ctx, logger, _ := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.reportProductUsageLoop").Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	for {
		err := usageController.ReportAllProductUsage(ctx)
		if err != nil {
			logger.Error(err, "failed to report all product usages", "context", "ReportAllUsage")
		}
		select {
		case <-productUsageControllerLinkChannel:
			return
		case tm := <-productUsageControllerTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

func NewProductUsageController(ariaCredentials *client.AriaCredentials, cloudAccountClient pb.CloudAccountServiceClient,
	usageServiceClient pb.UsageServiceClient, ariaUsageClient *client.AriaUsageClient, ariaAccountClient *client.AriaAccountClient) *UsageController {
	return &UsageController{
		cloudAccountClient: cloudAccountClient,
		usageServiceClient: usageServiceClient,
		ariaUsageClient:    ariaUsageClient,
		ariaAccountClient:  ariaAccountClient}
}

func (usageController *UsageController) ReportAllProductUsage(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("UsageController.ReportAllProductUsage").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	cloudAccts := usageController.getCloudAccts(ctx)

	for _, cloudAcct := range cloudAccts {
		productUsages, err := usageController.GetProductUsageOverStream(ctx, cloudAcct)
		if err != nil {
			logger.Error(err, "failed to report usages for cloud account", "cloudAccountId", cloudAcct.Id, "cloudAccountType", &cloudAcct.Type, "context", "SearchResourceUsages")
			span.RecordError(fmt.Errorf("failed to report usages for cloud account %w", err))
			return err
		}

		if productUsages == nil {
			logger.Info("no product usages", "cloudAcct", cloudAcct.Id)
			continue
		}

		billingUsagesToUpload := []*client.BillingUsage{}
		for _, productUsageRecord := range productUsages {
			billingUsagesToUpload = append(billingUsagesToUpload, &client.BillingUsage{
				CloudAccountId: productUsageRecord.CloudAccountId,
				ProductId:      productUsageRecord.ProductId,
				Amount:         productUsageRecord.Quantity,
				TransactionId:  productUsageRecord.Id,
				UsageId:        productUsageRecord.Id,
				UsageDate:      productUsageRecord.EndTime.AsTime(),
				UsageUnitType:  productUsageRecord.UsageUnitType,
			})
		}
		billingUsagesUploaded, err := usageController.ReportUsageToAria(ctx, billingUsagesToUpload)

		if err != nil {
			span.RecordError(fmt.Errorf("failed to upload product usage for cloud account %w", err))
			logger.Error(err, "failed to upload product usage for cloud account", "cloudAccountId", cloudAcct.Id, "cloudAccountType", cloudAcct.Type, "context", "ReportUsageToAria")
		}
		for _, billingUsageUploaded := range billingUsagesUploaded {
			_, err := usageController.usageServiceClient.MarkProductUsageAsReported(ctx, &pb.ReportProductUsageId{Id: billingUsageUploaded.UsageId})
			if err != nil {
				span.RecordError(fmt.Errorf("failed to mark product usage as reported for %w", err))
				logger.Error(err, "failed to mark product usage as reported for ", "usageId", billingUsageUploaded.UsageId)
			}
		}
	}

	return nil
}

func (usageController *UsageController) GetProductUsageOverStream(ctx context.Context, cloudAcct *pb.CloudAccount) ([]*pb.ReportProductUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.GetProductUsageOverStream").WithValues("cloudAccountId", cloudAcct.GetId()).Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	reported := false
	resourceUsagesFilter := &pb.ProductUsagesReportFilter{
		CloudAccountId: &cloudAcct.Id,
		Reported:       &reported,
	}

	productUsagesStream, err := usageController.usageServiceClient.SearchProductUsagesReport(ctx, resourceUsagesFilter)
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
			logger.Error(err, "recv error in SearchProductUsagesUnreported")
			return nil, err
		}
		productUsagesToReport = append(productUsagesToReport, reportProductUsage)
	}
	logger.Info("unreported product usage", "usages", productUsagesToReport)
	return productUsagesToReport, nil
}
