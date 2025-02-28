// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"errors"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/usages"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

// AriaBulkRecordUsageClient - Bulk Record Usage client API to create bulk usage records for a specified client
type AriaUsageClient struct {
	ariaClient      *AriaClient
	ariaCredentials *AriaCredentials
}

type BillingUsage struct {
	CloudAccountId string
	ProductId      string
	TransactionId  string
	ResourceId     string
	ResourceName   string
	Amount         float64
	RecordId       string
	UsageId        string
	RegionName     string
	UsageDate      time.Time
	UsageUnitType  string
}

func NewAriaUsageClient(ariaClient *AriaClient, ariaCredentials *AriaCredentials) *AriaUsageClient {
	return &AriaUsageClient{
		ariaClient:      ariaClient,
		ariaCredentials: ariaCredentials,
	}
}

func (usageClient *AriaUsageClient) CreateBulkUsageRecords(ctx context.Context, usage []*BillingUsage) (*response.BulkRecordUsageMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaClient.CallAriaIgnoreAriaFailure").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	bulkRecordUsageRequest := request.BulkRecordUsageMRequest{
		RestCall:     "bulk_record_usage_m",
		OutputFormat: "json",
		ClientNo:     usageClient.ariaCredentials.clientNo,
		AuthKey:      usageClient.ariaCredentials.authKey,
		AltCallerId:  AriaClientId,
	}

	err := validateIncomingSequence(usage)
	if err != nil {
		return nil, err
	}

	bulkRecordUsageRequest.UsageRecs = make([]usages.UsageRec, len(usage))
	for i := 0; i < len(bulkRecordUsageRequest.UsageRecs); i++ {
		usageAriaDate := DateTimeToAriaFormat(usage[i].UsageDate)
		usageType := GetUsageUnitTypeCode(usage[i].UsageUnitType)
		bulkRecordUsageRequest.UsageRecs[i].UsageTypeCode = usageType
		bulkRecordUsageRequest.UsageRecs[i].ClientAcctId = GetAccountClientId(usage[i].CloudAccountId)
		bulkRecordUsageRequest.UsageRecs[i].ClientMasterPlanInstanceId = GetClientMasterPlanInstanceId(usage[i].CloudAccountId, usage[i].ProductId)
		bulkRecordUsageRequest.UsageRecs[i].UsageUnits = GetLimitedDecimalUsageUnit(usage[i].Amount)
		bulkRecordUsageRequest.UsageRecs[i].ClientRecordId = usage[i].TransactionId
		bulkRecordUsageRequest.UsageRecs[i].UsageDate = usageAriaDate
		// removed for product usages
		// if usage[i].RegionName != "" {
		// 	bulkRecordUsageRequest.UsageRecs[i].Qualifier1 = usage[i].RegionName + ":" + usage[i].ResourceName
		// } else {
		// 	bulkRecordUsageRequest.UsageRecs[i].Qualifier1 = usage[i].ResourceName
		// }
	}
	return CallAriaIgnoreAriaFailure[response.BulkRecordUsageMResponse](ctx, usageClient.ariaClient, &bulkRecordUsageRequest, FailedToCreateAriaBulkUsageRecordError)
}

func (usageClient *AriaUsageClient) CreateUsageRecord(ctx context.Context, usageType string, usage *BillingUsage) (*response.RecordUsage, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageClient.CreateUsageRecord").
		WithValues("cloudAccountId", usage.CloudAccountId, "productId", usage.ProductId, "transactionId", usage.TransactionId, "amount", usage.Amount).
		Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	recordUsageRequest := request.RecordUsage{
		RestCall:     "record_usage_m",
		OutputFormat: "json",
		ClientNo:     usageClient.ariaCredentials.clientNo,
		AuthKey:      usageClient.ariaCredentials.authKey,
		AltCallerId:  AriaClientId,
	}

	err := validateIncomingSequence([]*BillingUsage{usage})
	if err != nil {
		return nil, GetErrorForParamter(recordUsageRequest.RestCall, usage, err)
	}
	usageAriaDate := DateTimeToAriaFormat(usage.UsageDate)
	recordUsageRequest.UsageTypeCode = usageType
	recordUsageRequest.ClientAcctId = GetAccountClientId(usage.CloudAccountId)
	recordUsageRequest.ClientMasterPlanInstanceId = GetClientMasterPlanInstanceId(usage.CloudAccountId, usage.ProductId)
	recordUsageRequest.UsageUnits = GetLimitedDecimalUsageUnit(usage.Amount)
	recordUsageRequest.ClientRecordId = usage.TransactionId
	recordUsageRequest.UsageDate = usageAriaDate

	return CallAriaIgnoreAriaFailure[response.RecordUsage](ctx, usageClient.ariaClient, &recordUsageRequest, FailedToCreateAriaUsageRecordError)
}

func validateIncomingSequence(usage []*BillingUsage) error {
	for _, u := range usage {
		switch {
		case u.CloudAccountId == "":
			return errors.New("cloud account id is empty")
		case u.ProductId == "":
			return errors.New("product id is empty")
		case u.TransactionId == "":
			return errors.New("transaction id is empty")
		// case u.ResourceId == "":
		// 	return errors.New("resource id is empty")
		case u.Amount < 0:
			return errors.New("amount is invalid")
		}
	}
	return nil
}

func (usageClient *AriaUsageClient) GetUnbilledUsageSummary(ctx context.Context, clientAccountID string, clientMasterPlanInstanceId string) (*response.GetUnbilledUsageSummaryMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageClient.GetUnbilledUsageSummary").WithValues("clientAccountId", clientAccountID, "clientMasterPlanInstanceId", clientMasterPlanInstanceId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getUnbilledRequest := request.GetUnbilledUsageSummaryMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_unbilled_usage_summary_m",
		},
		OutputFormat:               "json",
		ClientNo:                   usageClient.ariaCredentials.clientNo,
		AuthKey:                    usageClient.ariaCredentials.authKey,
		ClientAcctId:               clientAccountID,
		ClientMasterPlanInstanceId: clientMasterPlanInstanceId,
	}
	return CallAria[response.GetUnbilledUsageSummaryMResponse](ctx, usageClient.ariaClient, &getUnbilledRequest, FailedToGetUnbilledUsageSummary)
}

func (usageClient *AriaUsageClient) GetUsageSummaryByType(ctx context.Context, clientAccountID string, clientMasterPlanInstanceId string, startDate string, startTime string, endDate string, endTime string) (*response.GetUsageSummaryByType, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageClient.GetUsageSummaryByType").WithValues("clientAccountId", clientAccountID, "clientMasterPlanInstanceId", clientMasterPlanInstanceId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getUnbilledRequest := request.GetUsageSummaryByType{
		AriaRequest: request.AriaRequest{
			RestCall: "get_usage_summary_by_type_m",
		},
		OutputFormat:               "json",
		ClientNo:                   usageClient.ariaCredentials.clientNo,
		AuthKey:                    usageClient.ariaCredentials.authKey,
		ClientAcctId:               clientAccountID,
		ClientMasterPlanInstanceId: clientMasterPlanInstanceId,
		DateFilterStartDate:        startDate,
		DateFilterStartTime:        startTime,
		DateFilterEndDate:          endDate,
		DateFilterEndTime:          endTime,
	}
	return CallAria[response.GetUsageSummaryByType](ctx, usageClient.ariaClient, &getUnbilledRequest, FailedToGetUsageSummaryByType)
}

func (usageClient *AriaUsageClient) GetUsageHistory(ctx context.Context, clientAccountID string, clientMasterPlanInstanceId string, startDate string, endDate string) (*response.GetUsageHistory, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaUsageClient.GetUsageHistory").
		WithValues("clientAccountId", clientAccountID, "clientMasterPlanInstanceId", clientMasterPlanInstanceId, "startDate", startDate, "endDate", endDate).
		Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getUnbilledRequest := request.GetUsageHistory{
		AriaRequest: request.AriaRequest{
			RestCall: "get_usage_history_m",
		},
		OutputFormat:               "json",
		ClientNo:                   usageClient.ariaCredentials.clientNo,
		AuthKey:                    usageClient.ariaCredentials.authKey,
		ClientAcctId:               clientAccountID,
		ClientMasterPlanInstanceId: clientMasterPlanInstanceId,
		DateRangeStart:             startDate,
		DateRangeEnd:               endDate,
	}
	return CallAria[response.GetUsageHistory](ctx, usageClient.ariaClient, &getUnbilledRequest, FailedToGetUsageHistory)
}
