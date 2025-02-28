// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
)

const (
	NUMBER_OF_DAYS                  = 1
	ACCOUNT_TYPE_PREMIUM            = "premium"
	ACCOUNT_TYPE_ENTERPRISE_PENDING = "enterprise_pending"
	ACCOUNT_TYPE_ENTERPRISE         = "enterprise"
	DATE_LAYOUT                     = "2006-01-02"
	ACTION_DIRECTIVE_REGENERATE     = 3
	ACTION_DIRECTIVE_APPROVE        = 1
)

// please use this for initializing Aria and nothing in Aria controller.
func InitAriaForTesting(ctx context.Context) error {
	if err := EnableMinutesUsageType(); err != nil {
		return err
	}

	if err := EnableDefaultPlan(ctx); err != nil {
		return err
	}

	if err := common.GetPromoClient().EnsurePlanSet(ctx); err != nil {
		return err
	}

	if err := common.GetPromoClient().EnsurePromo(ctx); err != nil {
		return err
	}

	return nil
}

func EnableDefaultPlan(ctx context.Context) error {
	err := common.GetAriaPlanClient().CreateDefaultPlan(ctx)
	if err != nil && !strings.Contains(err.Error(), "error code:1001") {
		return err
	}
	return nil
}

func CreateAccountWithDefaultPlan(ctx context.Context, clientAcctId string) (*response.CreateAcctCompleteMResponse, error) {
	err := EnableDefaultPlan(ctx)
	if err != nil {
		return nil, err
	}
	ariaAccountClient := common.GetAriaAccountClient()
	clientPlanId := client.GetDefaultPlanClientId()
	createAcctResponse, err := ariaAccountClient.CreateAriaAccount(ctx, clientAcctId, clientPlanId, client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		return nil, err
	}
	return createAcctResponse, nil
}

func CreateAcctWithPlanDefaultProd(ctx context.Context, clientAcctId string) (*response.CreateAcctCompleteMResponse, string, error) {
	createPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	planId := product.Id
	err := EnableMinutesUsageType()
	if err != nil {
		return nil, "", err
	}
	usageType, err := GetMinutesUsageType()
	if err != nil {
		return nil, "", err
	}
	createAcctResponse, err := CreateAccountWithDefaultPlan(ctx, clientAcctId)
	if err != nil {
		return nil, "", err
	}
	_, err = createPlan.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		return nil, "", err
	}
	err = AssignPlanToAccount(ctx, clientAcctId, GetTestClientPlanId(planId), ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		return nil, "", err
	}
	return createAcctResponse, GetTestClientPlanId(planId), nil
}

func CreateTestUsageRecord(ctx context.Context, clientAccountId string, clientPlanId string, usageTypeCode string, amount float64) error {
	ariaUsageClient := common.GetUsageClient()
	currentDate := time.Now()
	prevDate := currentDate.AddDate(0, 0, -NUMBER_OF_DAYS)
	usages := []*client.BillingUsage{
		{
			CloudAccountId: GetCloudAcctIdFromClientAcctId(clientAccountId),
			ProductId:      GetProdIdFromClientPlanId(clientPlanId),
			TransactionId:  uuid.New().String(),
			ResourceId:     uuid.New().String(),
			ResourceName:   uuid.New().String(),
			RecordId:       uuid.NewString(),
			UsageDate:      prevDate,
			Amount:         amount,
			UsageUnitType:  "RATE_UNIT_DOLLARS_PER_MINUTE",
		},
	}
	usageResp, err := ariaUsageClient.CreateBulkUsageRecords(ctx, usages)
	if client.IsPayloadEmpty(usageResp) {
		return fmt.Errorf("create usage record error: %v", usageResp)
	}
	if !strings.Contains(string(usageResp.ErrorMsg), "Bulk usage loader completed with some errors") &&
		!strings.Contains(string(usageResp.ErrorMsg), "OK") {
		return fmt.Errorf("bulk usage loader did not complete sucessfully: %v", usageResp)
	}
	//TO make sure get usages data
	usages[0].RecordId = uuid.NewString()
	if _, err2 := ariaUsageClient.CreateUsageRecord(ctx, usageTypeCode, usages[0]); err2 != nil {
		return fmt.Errorf("create usage record error: %v", err2)
	}

	return err
}

func CreateTestUsageRecordDefaultAmount(ctx context.Context, clientAccountId string, clientPlanId string, usageTypeCode string) error {
	return CreateTestUsageRecord(ctx, clientAccountId, clientPlanId, usageTypeCode, DefaultUsageAmount)
}

func AssignPlanToAccount(ctx context.Context, clientAccountId string, clientPlanId string, acctType string) error {
	cloudAcctId := GetCloudAcctIdFromClientAcctId(clientAccountId)
	productId := GetProdIdFromClientPlanId(clientPlanId)
	clientMasterPlanInstanceId := client.GetClientMasterPlanInstanceId(cloudAcctId, productId)
	clientRateScheduleId := client.GetRateScheduleClientId(productId, acctType)
	ariaAccountClient := common.GetAriaAccountClient()
	_, err := ariaAccountClient.AssignPlanToAccountWithBillingAndDunningGroup(ctx, clientAccountId, clientPlanId, clientMasterPlanInstanceId, clientRateScheduleId, client.BILL_LAG_DAYS, 0)
	return err
}

func AssignPlanToEnterpriseAccount(ctx context.Context, clientAccountId string, clientPlanId string, overrideBillThroughDate string, parentClientMasterPlanInstanceId string) error {
	cloudAcctId := GetCloudAcctIdFromClientAcctId(clientAccountId)
	productId := GetProdIdFromClientPlanId(clientPlanId)
	clientMasterPlanInstanceId := client.GetClientMasterPlanInstanceId(cloudAcctId, productId)
	clientRateScheduleId := client.GetRateScheduleClientId(productId, ACCOUNT_TYPE_ENTERPRISE)
	ariaAccountClient := common.GetAriaAccountClient()
	_, err := ariaAccountClient.AssignPlanToEnterpriseChildAccount(ctx, clientAccountId, clientPlanId, clientMasterPlanInstanceId, clientRateScheduleId, client.BILL_LAG_DAYS, 0, overrideBillThroughDate, client.PRORATE_FIRST_INVOICE, client.PARENT_PAY_FOR_CHILD_ACCOUNT, parentClientMasterPlanInstanceId)
	return err
}

func AssignCreditsToAccount(ctx context.Context, clientAccountId string, amount float64) (*response.CreateAdvancedServiceCreditMResponse, error) {
	createCreditsResponse, err := common.GetServiceCreditClient().CreateServiceCredits(ctx,
		clientAccountId,
		amount,
		kFixMeWeHaventImplementedReasonCodeYet, DefaultCreditExpirationDate, DefaultCommentsForCredits)
	return createCreditsResponse, err
}

func EnableMinutesUsageType() error {
	usageTypeClient := common.GetAriaUsageTypeClient()
	ctx := context.Background()
	_, err := usageTypeClient.GetUsageTypeDetails(ctx, client.GetMinsUsageTypeCode())
	if err != nil && strings.Contains(err.Error(), "error code:1010") {
		usageUnitTypes, err := usageTypeClient.GetMinuteUsageUnitType(ctx)
		if err != nil {
			return err
		}
		_, err = usageTypeClient.CreateUsageType(ctx, client.USAGE_TYPE_NAME, client.USAGE_TYPE_DESC, usageUnitTypes.UsageUnitTypeNo, client.GetMinsUsageTypeCode())
		if err != nil {
			return err
		}
	}
	return nil
}

func GetMinutesUsageType() (*data.UsageType, error) {
	usageTypeClient := common.GetAriaUsageTypeClient()
	usageType, err := usageTypeClient.GetMinutesUsageType(context.Background())
	if err != nil {
		return nil, err
	}
	return usageType, nil
}

func GetTestClientPlanId(id string) string {
	return config.Cfg.ClientIdPrefix + "." + id
}

func CreateAcctWithUsage(ctx context.Context, clientAcctId string) error {
	_, clientPlanId, err := CreateAcctWithPlanDefaultProd(ctx, clientAcctId)
	if err != nil {
		return err
	}
	err = CreateTestUsageRecord(ctx, clientAcctId, clientPlanId, client.GetMinsUsageTypeCode(), DefaultUsageAmount)
	if err != nil {
		return err
	}
	return nil
}

// this method is exactly the same as in product controller.
// it cannot be a util unless we make all clients or specific clients global context.
// however making clients global context will effect the injection of deps and hence not making it.
func GetPlanServices(ctx context.Context, resp *response.GetPlanDetailResponse, clientPlanId string) ([]data.PlanService, error) {
	planServices := make([]data.PlanService, 0, len(resp.Services))
	var err error
	for _, service := range resp.Services {
		//Map service to rates
		// Note: This is different than design - Atanu to address.
		clientPlanServiceRates, err := common.GetAriaPlanClient().GetClientPlanServiceRates(ctx, clientPlanId, service.ClientServiceId)
		if err != nil {
			return nil, err
		}

		dService := data.PlanService{
			ServiceNo:        int64(service.ServiceNo),
			ClientServiceId:  service.ClientServiceId,
			PlanServiceRates: clientPlanServiceRates.PlanServiceRates,
		}
		planServices = append(planServices, dService)
	}
	return planServices, err
}

func ManageInvoice(ctx context.Context, t *testing.T, cloudAcctId string) int64 {
	ariaInvoiceClient := common.GetAriaInvoiceClient()
	clientAccountId := client.GetAccountClientId(cloudAcctId)
	invoiceId := int64(0)
	getPendingInovice, err := ariaInvoiceClient.GetPendingInvoiceNo(ctx, clientAccountId)
	if err != nil {
		t.Fatalf("failed to get pending invoice: %v", err)
	}
	for _, pendingInvoice := range getPendingInovice.PendingInvoice {
		_, err := ariaInvoiceClient.ManagePendingInvoiceWithInoviceNo(ctx, clientAccountId, pendingInvoice.InvoiceNo, ACTION_DIRECTIVE_REGENERATE)
		if err != nil {
			t.Fatalf("failed to regenerate invoice: %v", err)
		}
	}
	getPendingInovice, err = ariaInvoiceClient.GetPendingInvoiceNo(ctx, clientAccountId)
	if err != nil {
		t.Fatalf("failed to get pending invoice to approve: %v", err)
	}
	for _, pendingInvoice := range getPendingInovice.PendingInvoice {
		_, err := ariaInvoiceClient.ManagePendingInvoiceWithInoviceNo(ctx, clientAccountId, pendingInvoice.InvoiceNo, ACTION_DIRECTIVE_APPROVE)
		if err != nil {
			t.Fatalf("failed to approve invoice: %v", err)
		}
		if strings.Contains(pendingInvoice.ClientMasterPlanInstanceId, clientAccountId) {
			invoiceId = pendingInvoice.InvoiceNo
		}
	}
	return invoiceId
}

func EnableStorageUsageType() error {
	usageTypeClient := common.GetAriaUsageTypeClient()
	ctx := context.Background()
	_, err := usageTypeClient.GetUsageTypeDetails(ctx, client.GetStorageUsageUnitTypeCode())
	if err != nil && strings.Contains(err.Error(), "error code:1010") {
		usageUnitTypes, err := usageTypeClient.GetStorageUsageUnitType(ctx)
		if err != nil {
			return err
		}
		_, err = usageTypeClient.CreateUsageType(ctx, client.USAGE_TYPE_NAME, client.USAGE_TYPE_DESC, usageUnitTypes.UsageUnitTypeNo, client.GetStorageUsageUnitTypeCode())
		if err != nil {
			return err
		}
	}
	return nil
}

func GetStorageUsageType() (*data.UsageType, error) {
	usageTypeClient := common.GetAriaUsageTypeClient()
	usageType, err := usageTypeClient.GetStorageUsageType(context.Background())
	if err != nil {
		return nil, err
	}
	return usageType, nil
}

func CreateTestStorageUsageRecord(ctx context.Context, clientAccountId string, clientPlanId string, usageTypeCode string, amount float64, usageUnitType string) error {
	ariaUsageClient := common.GetUsageClient()
	currentDate := time.Now()
	prevDate := currentDate.AddDate(0, 0, -NUMBER_OF_DAYS)
	usages := []*client.BillingUsage{
		{
			CloudAccountId: GetCloudAcctIdFromClientAcctId(clientAccountId),
			ProductId:      GetProdIdFromClientPlanId(clientPlanId),
			TransactionId:  uuid.New().String(),
			ResourceId:     uuid.New().String(),
			ResourceName:   uuid.New().String(),
			RecordId:       uuid.NewString(),
			UsageDate:      prevDate,
			Amount:         amount,
			UsageUnitType:  usageUnitType,
		},
	}
	usageResp, err := ariaUsageClient.CreateBulkUsageRecords(ctx, usages)
	if client.IsPayloadEmpty(usageResp) {
		return fmt.Errorf("create usage record error: %v", usageResp)
	}
	if !strings.Contains(string(usageResp.ErrorMsg), "Bulk usage loader completed with some errors") &&
		!strings.Contains(string(usageResp.ErrorMsg), "OK") {
		return fmt.Errorf("bulk usage loader did not complete sucessfully: %v", usageResp)
	}
	//TO make sure get usages data
	usages[0].RecordId = uuid.NewString()
	if _, err2 := ariaUsageClient.CreateUsageRecord(ctx, usageTypeCode, usages[0]); err2 != nil {
		return fmt.Errorf("create usage record error: %v", err2)
	}

	return err
}

func GetTestUsageUnitTypeCode(usageType string) (*data.UsageType, error) {
	if usageType == "RATE_UNIT_DOLLARS_PER_TB_PER_HOUR" || usageType == "per TB per Hour" {
		return GetStorageUsageType()
	}
	return GetMinutesUsageType()
}
