// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	cloudCreditUsageReportLinkChannel = make(chan bool)
	cloudCreditUsageReportTicker      *time.Ticker
)

func StartCloudCreditUsageReportScheduler(ctx context.Context, creditUsageReport *CloudCreditUsageReport) {
	logger := log.FromContext(ctx).WithName("StartCloudCreditUsageReportScheduler")
	logger.V(9).Info("BEGIN")
	logger.Info("cfg", "StartCloudCreditUsageReportScheduler", Cfg.CloudCreditUsageReportSchedulerInterval)
	defer logger.Info("END")
	cloudCreditUsageReportTicker = time.NewTicker(time.Duration(Cfg.CloudCreditUsageReportSchedulerInterval) * time.Second)
	go reportCloudCreditUsageLoop(context.Background(), creditUsageReport)
}

func StopCloudCreditUsageReportScheduler() {
	if cloudCreditUsageReportLinkChannel != nil {
		close(cloudCreditUsageReportLinkChannel)
		cloudCreditUsageReportLinkChannel = nil
	}
}

func reportCloudCreditUsageLoop(ctx context.Context, creditUsageReport *CloudCreditUsageReport) {
	ctx, logger, _ := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.reportCloudCreditUsageLoop").Start()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	for {
		err := creditUsageReport.ReportCloudCreditUsages(ctx)
		if err != nil {
			logger.Error(err, "failed to report credit usages")
		}
		select {
		case <-cloudCreditUsageReportLinkChannel:
			return
		case tm := <-cloudCreditUsageReportTicker.C:
			if tm.IsZero() {
				return
			}
		}
	}
}

type CloudCreditUsageReport struct {
	billingDb          *sql.DB
	cloudAccountClient pb.CloudAccountServiceClient
	usageServiceClient pb.UsageServiceClient
}

func NewCloudCreditUsageReport(billingDb *sql.DB, cloudAccountClient pb.CloudAccountServiceClient,
	usageServiceClient pb.UsageServiceClient) *CloudCreditUsageReport {
	return &CloudCreditUsageReport{
		billingDb:          billingDb,
		cloudAccountClient: cloudAccountClient,
		usageServiceClient: usageServiceClient,
	}
}

// One could move this code to billing common but we cannot fail if one cloud account cannot be retrieved.
// Such custom behavior unfortunately leads to custom code that cannot be moved to common functionality.
func (creditUsageReport *CloudCreditUsageReport) getCloudAccts(ctx context.Context) []*pb.CloudAccount {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.getCloudAccts").Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	acctTypes := []pb.AccountType{pb.AccountType_ACCOUNT_TYPE_STANDARD, pb.AccountType_ACCOUNT_TYPE_INTEL}
	cloudAccts := []*pb.CloudAccount{}

	for _, acctType := range acctTypes {
		cloudAccountSearchClient, err :=
			creditUsageReport.cloudAccountClient.Search(ctx, &pb.CloudAccountFilter{Type: &acctType})
		if err != nil {
			logger.Error(err, "failed to get cloud account client for searching on", "accountType", acctType.String())
			continue
		}

		for {
			cloudAccount, err := cloudAccountSearchClient.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				logger.Error(err, "failed to get cloud account")
				break
			}
			cloudAccts = append(cloudAccts, cloudAccount)
		}
	}

	return cloudAccts
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

// some of these fields will not be used.
// leaving the old fields as they were in case there is some value.
func (creditUsageReport *CloudCreditUsageReport) ReportCloudCreditUsages(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.ReportCloudCreditUsages").Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")

	cloudAccts := creditUsageReport.getCloudAccts(ctx)

	for _, cloudAcct := range cloudAccts {
		resourceUsages, err := creditUsageReport.GetResourceUsageOverStream(ctx, cloudAcct)
		if err != nil {
			logger.Error(err, "failed to report cloud credit usages for cloud account", "cloudAccountId", cloudAcct.GetId())
		}

		if resourceUsages == nil {
			logger.Info("no resource usages", "cloudAccountId", cloudAcct.GetId())
			continue
		}

		billingUsagesToReport := []*BillingUsage{}
		for _, resourceUsage := range resourceUsages.ResourceUsages {
			billingUsagesToReport = append(billingUsagesToReport, &BillingUsage{
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
		for _, billingUsageReported := range billingUsagesReported {
			if billingUsageReported.AmountUnReported == 0 {
				_, err = creditUsageReport.usageServiceClient.MarkResourceUsageAsReported(ctx, &pb.ResourceUsageId{Id: billingUsageReported.UsageId})
				if err != nil {
					logger.Error(err, "failed to mark usage as reported for ", "usageId", billingUsageReported.UsageId, "cloudAccountId", cloudAcct.GetId())
				}
			} else {
				_, err = creditUsageReport.usageServiceClient.UpdateResourceUsage(ctx,
					&pb.ResourceUsageUpdate{
						ResourceUsageId:    billingUsageReported.UsageId,
						UnReportedQuantity: billingUsageReported.AmountUnReported / billingUsageReported.UsageRate,
					})
				if err != nil {
					logger.Error(err, "failed to update the unreported usage for ", "usageId", billingUsageReported.UsageId, "cloudAccountId", cloudAcct.GetId())
				}
			}
		}
	}

	return nil
}

func (creditUsageReport *CloudCreditUsageReport) GetResourceUsage(ctx context.Context, cloudAcct *pb.CloudAccount) (*pb.ResourceUsages, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.GetResourceUsage").WithValues("cloudAccountId", cloudAcct.GetId()).Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")
	reported := false
	resourceUsages, err := creditUsageReport.usageServiceClient.SearchResourceUsages(ctx,
		&pb.ResourceUsagesFilter{
			CloudAccountId: &cloudAcct.Id,
			Reported:       &reported,
		})
	if err != nil {
		logger.Error(err, "failed to get usages for reporting to cloud credits")
		return nil, err
	}
	return resourceUsages, nil
}

func (creditUsageReport *CloudCreditUsageReport) GetResourceUsageOverStream(ctx context.Context, cloudAcct *pb.CloudAccount) (*pb.ResourceUsages, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.GetResourceUsageOverStream").WithValues("cloudAccountId", cloudAcct.GetId()).Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")
	reported := false
	resourceUsagesFilter := &pb.ResourceUsagesFilter{
		CloudAccountId: &cloudAcct.Id,
		Reported:       &reported,
	}
	resourceUsagesStream, err := creditUsageReport.usageServiceClient.StreamSearchResourceUsages(ctx, resourceUsagesFilter)
	if err != nil {
		logger.Error(err, "failed to get resource usages for cloud account")
		return nil, err
	}
	resourceUsages := &pb.ResourceUsages{}
	for {

		resourceUsage, err := resourceUsagesStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			logger.Error(err, "failed to recv resource usages for cloud account")
			return nil, err
		}

		resourceUsages.ResourceUsages = append(resourceUsages.ResourceUsages, resourceUsage)
	}
	return resourceUsages, nil
}

type Credit struct {
	CreditId        uint64
	Code            string
	RemainingAmount float64
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

	rows, err := creditUsageReport.billingDb.QueryContext(ctx, query, cloudAccountId)
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

func (creditUsageReport *CloudCreditUsageReport) ReportCloudCreditUsage(ctx context.Context, billingUsagesToReport []*BillingUsage) ([]*BillingUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditUsageReport.ReportCloudCreditUsage").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	billingUsagesReported := []*BillingUsage{}
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

		tx, err := creditUsageReport.billingDb.BeginTx(ctx, nil)
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
			// no need to proceed if the amount reported is 0
			// add this check to the for loop
			if amountToReport == 0 {
				break
			}
		}

		if err := tx.Commit(); err != nil {
			logger.Error(err, "error committing db transaction", "cloudAccountId", billingUsageToReport.CloudAccountId)
			return nil, err
		}

		if amountToReport == 0 {
			billingUsagesReported = append(billingUsagesReported, &BillingUsage{
				CloudAccountId:   billingUsageToReport.CloudAccountId,
				ProductId:        billingUsageToReport.ProductId,
				ResourceId:       billingUsageToReport.ResourceId,
				ResourceName:     billingUsageToReport.ResourceName,
				Quantity:         billingUsageToReport.Quantity,
				Amount:           billingUsageToReport.Amount,
				AmountUnReported: 0,
				UsageRate:        billingUsageToReport.UsageRate,
				TransactionId:    billingUsageToReport.TransactionId,
				UsageId:          billingUsageToReport.UsageId,
				RegionName:       billingUsageToReport.RegionName,
				UsageDate:        billingUsageToReport.UsageDate,
			})
			// add this check to the for loop
		} else {
			billingUsagesReported = append(billingUsagesReported, &BillingUsage{
				CloudAccountId:   billingUsageToReport.CloudAccountId,
				ProductId:        billingUsageToReport.ProductId,
				ResourceId:       billingUsageToReport.ResourceId,
				ResourceName:     billingUsageToReport.ResourceName,
				Quantity:         billingUsageToReport.Quantity,
				Amount:           billingUsageToReport.Amount,
				AmountUnReported: amountToReport,
				UsageRate:        billingUsageToReport.UsageRate,
				TransactionId:    billingUsageToReport.TransactionId,
				UsageId:          billingUsageToReport.UsageId,
				RegionName:       billingUsageToReport.RegionName,
				UsageDate:        billingUsageToReport.UsageDate,
			})
			// there is no point in continuing any further..
			break
		}
	}
	return billingUsagesReported, nil
}
