// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcredits

import (
	"context"
	"database/sql"
	"time"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

var (
	cloudCreditsUsageReportLinkChannel = make(chan bool)
	cloudCreditsUsageReportTicker      *time.Ticker
)

func StartCloudCreditUsageReportScheduler(ctx context.Context, creditUsageReport CloudCreditUsageReport) {
	logger := log.FromContext(ctx).WithName("StartCloudCreditUsageReportScheduler")
	logger.Info("BEGIN")
	logger.Info("cfg", "StartCloudCreditUsageReportScheduler", config.Cfg.CreditUsageReportSchedulerInterval)
	defer logger.Info("END")
	cloudCreditsUsageReportTicker = time.NewTicker(time.Duration(config.Cfg.CreditUsageReportSchedulerInterval) * time.Second)
	go reportCloudCreditUsageLoop(context.Background(), &creditUsageReport)
}

func StopCloudCreditUsageReportScheduler() {
	if cloudCreditsUsageReportLinkChannel != nil {
		close(cloudCreditsUsageReportLinkChannel)
		cloudCreditsUsageReportLinkChannel = nil
	}
}

func reportCloudCreditUsageLoop(ctx context.Context, creditUsageReport *CloudCreditUsageReport) {
	ctx, logger, _ := obs.LogAndSpanFromContext(ctx).WithName("reportCloudCreditUsageLoop").Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	for {
		err := creditUsageReport.ReportCloudCreditResourceUsages(ctx)
		if err != nil {
			logger.Error(err, "failed to report resource credit usages")
		}
		err = creditUsageReport.ReportCloudCreditProductUsages(ctx)
		if err != nil {
			logger.Error(err, "failed to report product credit usages")
		}
		select {
		case <-cloudCreditsUsageReportLinkChannel:
			return
		case tm := <-cloudCreditsUsageReportTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

type CloudCreditUsageReport struct {
	cloudCreditsDb     *sql.DB
	cloudAccountClient *billingCommon.CloudAccountSvcClient
	usageServiceClient *billingCommon.UsageSvcClient
}

func NewCreditUsageReportScheduler(cloudCreditsDb *sql.DB, cloudAccountClient *billingCommon.CloudAccountSvcClient,
	usageServiceClient *billingCommon.UsageSvcClient) *CloudCreditUsageReport {
	return &CloudCreditUsageReport{
		cloudCreditsDb:     cloudCreditsDb,
		cloudAccountClient: cloudAccountClient,
		usageServiceClient: usageServiceClient,
	}
}

func (creditUsageReport *CloudCreditUsageReport) ReportCloudCreditResourceUsages(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.ReportCloudCreditResourcetUsages").Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	cloudAccts := creditUsageReport.cloudAccountClient.GetStandardAndIntelAccounts(ctx)
	for _, cloudAcct := range cloudAccts {

		resourceUsages, err := creditUsageReport.usageServiceClient.GetResourceUsageOverStream(ctx, cloudAcct)
		if err != nil {
			logger.Error(err, "failed to get resource usages for cloud account", "cloudAccountId", cloudAcct.GetId())
			return err
		}

		if resourceUsages == nil || len(resourceUsages.ResourceUsages) == 0 {
			logger.Info("no resource usages to report", "cloudAccountId", cloudAcct.GetId())
			continue
		}

		billingUsagesToReport := []*billingCommon.BillingUsage{}
		for _, resourceUsage := range resourceUsages.ResourceUsages {
			billingUsagesToReport = append(billingUsagesToReport, &billingCommon.BillingUsage{
				CloudAccountId:   resourceUsage.CloudAccountId,
				ProductId:        resourceUsage.ProductId,
				ResourceId:       resourceUsage.ResourceId,
				ResourceName:     resourceUsage.ResourceName,
				Quantity:         resourceUsage.Quantity,
				Amount:           resourceUsage.UnReportedQuantity * resourceUsage.Rate,
				AmountUnReported: resourceUsage.UnReportedQuantity * resourceUsage.Rate,
				UsageRate:        resourceUsage.Rate,
				TransactionId:    resourceUsage.Id,
				UsageId:          resourceUsage.Id,
				RegionName:       resourceUsage.Region,
				UsageDate:        resourceUsage.Timestamp.AsTime(),
				UsageUnitType:    resourceUsage.UsageUnitType,
			})
		}
		billingUsagesReported, err := creditUsageReport.ReportCloudCreditUsage(ctx, billingUsagesToReport)

		if err != nil {
			logger.Error(err, "failed to report cloud credit usage for cloud account", "cloudAccountId", cloudAcct.GetId())
		}

		err = creditUsageReport.usageServiceClient.UpdateAndMarkResourceRecordsReported(ctx, billingUsagesReported, cloudAcct)
		if err != nil {
			logger.Error(err, "failed to update resource usage records reported", "cloudAccountId", cloudAcct.GetId())
		}

	}

	return nil
}

func (creditUsageReport *CloudCreditUsageReport) GetSortedCredits(ctx context.Context, cloudAccountId string) ([]Credit, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.GetSortedCredits").WithValues("cloudAccountId", cloudAccountId).Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	defer span.End()

	query := "SELECT id,coupon_code,remaining_amount " +
		"FROM cloud_credits " +
		"WHERE cloud_account_id=$1 AND expiry >= NOW() AND remaining_amount != 0 " +
		"ORDER BY created_at ASC "

	logger.V(9).Info("get sorted credit", "query", query, "cloudAccountId", cloudAccountId)
	var credits []Credit

	rows, err := creditUsageReport.cloudCreditsDb.QueryContext(ctx, query, cloudAccountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		obj := Credit{}
		if err := rows.Scan(&obj.CreditId, &obj.Code, &obj.RemainingAmount); err != nil {
			return nil, err
		}
		logger.V(9).Info("credit", "creditId", obj.CreditId, "remainingAmount", obj.RemainingAmount)
		credits = append(credits, obj)
	}

	return credits, nil
}

func GetRemainingAmount(credits []Credit) float64 {
	totalAmt := float64(0)
	for _, obj := range credits {

		totalAmt += obj.RemainingAmount
	}
	return totalAmt
}

func (creditUsageReport *CloudCreditUsageReport) ReportCloudCreditUsage(ctx context.Context, billingUsagesToReport []*billingCommon.BillingUsage) ([]*billingCommon.BillingUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.ReportCloudCreditUsage").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	billingUsagesReported := []*billingCommon.BillingUsage{}
	for _, billingUsageToReport := range billingUsagesToReport {

		creditList, err := creditUsageReport.GetSortedCredits(ctx, billingUsageToReport.CloudAccountId)
		if err != nil {
			logger.Error(err, "failed to get a sorted list of credits", "cloudAccountId", billingUsageToReport.CloudAccountId)
			return nil, err
		}

		if len(creditList) == 0 {
			logger.Info("no available credit for cloud account", "cloudAccountId", billingUsageToReport.CloudAccountId)
			return nil, err
		}

		amountToReport := billingUsageToReport.Amount

		tx, err := creditUsageReport.cloudCreditsDb.BeginTx(ctx, nil)
		if err != nil {
			logger.Error(err, "error starting db transaction for reporting usage", "cloudAccountId", billingUsageToReport.CloudAccountId)
			return nil, err
		}
		defer tx.Rollback()

		// Update remainingamount query
		query := "UPDATE cloud_credits " +
			"SET remaining_amount=$1, updated_at=NOW()::timestamp " +
			"WHERE cloud_account_id=$2 AND id=$3"

		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			logger.Error(err, "failed to prepare query for updating remaining amount for cloud credits", "cloudAccountId", billingUsageToReport.CloudAccountId)
			return nil, err
		}
		defer stmt.Close()

		for _, creditObj := range creditList {
			logger.Info("Consuming", "creditId", creditObj.CreditId)
			var remainingAmount float64

			if creditObj.RemainingAmount >= amountToReport {
				remainingAmount = creditObj.RemainingAmount - amountToReport
				amountToReport = 0
			} else {
				remainingAmount = 0
				amountToReport = amountToReport - creditObj.RemainingAmount
			}

			_, err = stmt.ExecContext(ctx, remainingAmount, billingUsageToReport.CloudAccountId, creditObj.CreditId)
			if err != nil {
				logger.Error(err, "error updating db for the remaining amount of cloud credit", "creditId", creditObj.CreditId, "cloudAccountId", billingUsageToReport.CloudAccountId)
				return nil, err
			}

			if amountToReport == 0 {
				break
			}
		}

		if err := tx.Commit(); err != nil {
			logger.Error(err, "error committing db transaction", "cloudAccountId", billingUsageToReport.CloudAccountId)
			return nil, err
		}

		billingUsage := &billingCommon.BillingUsage{
			CloudAccountId: billingUsageToReport.CloudAccountId,
			ProductId:      billingUsageToReport.ProductId,
			ResourceId:     billingUsageToReport.ResourceId,
			ResourceName:   billingUsageToReport.ResourceName,
			Quantity:       billingUsageToReport.Quantity,
			Amount:         billingUsageToReport.Amount,
			UsageRate:      billingUsageToReport.UsageRate,
			TransactionId:  billingUsageToReport.TransactionId,
			UsageId:        billingUsageToReport.UsageId,
			RegionName:     billingUsageToReport.RegionName,
			UsageDate:      billingUsageToReport.UsageDate,
		}

		if amountToReport == 0 {
			billingUsage.AmountUnReported = 0
		} else {
			billingUsage.AmountUnReported = amountToReport
		}
		billingUsagesReported = append(billingUsagesReported, billingUsage)
		// there is no point in continuing any further..
		break
	}
	return billingUsagesReported, nil
}

func (creditUsageReport *CloudCreditUsageReport) ReportCloudCreditProductUsages(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.ReportCloudCreditProductUsages").Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	cloudAccts := creditUsageReport.cloudAccountClient.GetStandardAndIntelAccounts(ctx)

	for _, cloudAcct := range cloudAccts {
		reportProductUsages, err := creditUsageReport.usageServiceClient.GetProductUsageReportOverStream(ctx, cloudAcct)
		if err != nil {
			logger.Error(err, "failed to report cloud credit usages for cloud account", "cloudAccountId", cloudAcct.GetId())
			return err
		}

		if len(reportProductUsages) == 0 || reportProductUsages == nil {
			logger.Info("no product usages to report", "cloudAccountId", cloudAcct.GetId())
			continue
		}

		billingUsagesToReport := []*billingCommon.BillingUsage{}
		for _, reportProductUsage := range reportProductUsages {
			billingUsagesToReport = append(billingUsagesToReport, &billingCommon.BillingUsage{
				CloudAccountId:   reportProductUsage.CloudAccountId,
				ProductId:        reportProductUsage.ProductId,
				Quantity:         reportProductUsage.Quantity,
				Amount:           reportProductUsage.UnReportedQuantity * reportProductUsage.Rate,
				AmountUnReported: reportProductUsage.UnReportedQuantity * reportProductUsage.Rate,
				UsageRate:        reportProductUsage.Rate,
				TransactionId:    reportProductUsage.Id,
				UsageId:          reportProductUsage.Id,
				UsageDate:        reportProductUsage.Timestamp.AsTime(),
				UsageUnitType:    reportProductUsage.UsageUnitType,
			})
		}

		productUsagesReported, err := creditUsageReport.ReportCloudCreditUsage(ctx, billingUsagesToReport)
		if err != nil {
			logger.Error(err, "failed to report cloud credit usage for cloud account", "cloudAccountId", cloudAcct.GetId())
		}

		if len(productUsagesReported) == 0 || productUsagesReported == nil {
			logger.Info("received no productUsagesReported", "cloudAccountId", cloudAcct.GetId())
			continue
		}

		err = creditUsageReport.usageServiceClient.UpdateAndMarkProductRecordsReported(ctx, productUsagesReported, cloudAcct)
		if err != nil {
			logger.Error(err, "failed to update usage records", "cloudAccountId", cloudAcct.GetId())
		}

	}
	return nil
}
