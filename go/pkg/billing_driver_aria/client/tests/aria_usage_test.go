// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

var (
	ClientAccountId string
	ProductId       string
)

func BeforeUsageTest(ctx context.Context, testName string, t *testing.T) {
	logger := log.FromContext(ctx).WithName("BeforeUsageTest")
	logger.Info("Running usage setup tasks before test", "testName", testName)
	logger.Info("usage setup tasks", "ClientAccountId", ClientAccountId, "ProductId", ProductId)
	if ClientAccountId == "" && ProductId == "" {
		createPlan := common.GetAriaPlanClient()
		product := GetProduct()
		productFamily := GetProductFamily()
		ProductId = product.GetId()
		err := EnableMinutesUsageType()
		if err != nil {
			t.Fatalf("failed to create usage type: %v", err)
		}
		usageType, err := GetMinutesUsageType()
		if err != nil {
			t.Fatalf("failed to get usage type: %v", err)
		}
		clientAcctId := GetClientAccountId()
		ClientAccountId = clientAcctId
		_, err = CreateAccountWithDefaultPlan(ctx, clientAcctId)
		if err != nil {
			t.Fatalf("failed to create account: %v", err)
		}
		_, err = createPlan.CreatePlan(ctx, product, productFamily, usageType)
		if err != nil {
			t.Fatalf("failed to create plan: %v", err)
		}
		clientPlanId := GetTestClientPlanId(product.GetId())
		err = AssignPlanToAccount(ctx, clientAcctId, clientPlanId, ACCOUNT_TYPE_PREMIUM)
		if err != nil {
			t.Fatalf("failed to assign plan to account: %v", err)
		}
		err = CreateTestUsageRecordDefaultAmount(ctx, clientAcctId, clientPlanId, usageType.UsageTypeCode)
		if err != nil {
			t.Fatalf("failed to create bulk usage record: %v", err)
		}
	}

}

func TestGetUnbilledUsage(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetUnappliedUsageSummary")
	logger.Info("testing unapplied usage summary")
	ctx := context.Background()
	BeforeUsageTest(ctx, "TestGetUnbilledUsage", t)
	ariaUsageClient := common.GetUsageClient()
	cloudAcctId := GetCloudAcctIdFromClientAcctId(ClientAccountId)
	_, err := ariaUsageClient.GetUnbilledUsageSummary(ctx, ClientAccountId, client.GetClientMasterPlanInstanceId(cloudAcctId, ProductId))
	if err != nil {
		t.Fatalf("Failed to get unbilled usage summary: %v", err)
	}
}

func TestGetUsageSummaryByType(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetUsageSummaryByType")
	logger.Info("testing usage summary by type")
	ctx := context.Background()
	BeforeUsageTest(ctx, "TestGetUsageSummaryByType", t)
	currentTime := time.Now()
	startDateTime := currentTime.AddDate(0, 0, -2)
	endDateTime := currentTime.AddDate(0, 0, 1)
	startDate, startTime := client.SplitDateTimeToAriaFormat(startDateTime)
	endDate, endTime := client.SplitDateTimeToAriaFormat(endDateTime)
	ariaUsageClient := common.GetUsageClient()
	cloudAcctId := GetCloudAcctIdFromClientAcctId(ClientAccountId)
	resp, err := ariaUsageClient.GetUsageSummaryByType(ctx, ClientAccountId, client.GetClientMasterPlanInstanceId(cloudAcctId, ProductId), startDate, startTime, endDate, endTime)
	logger.Info("usage summary ", "resp", resp)
	if err != nil {
		t.Fatalf("Failed to get usage summary: %v", err)
	}
}

func TestGetUsageHistory(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetUsageHistory")
	logger.Info("testing usage history")
	ctx := context.Background()
	BeforeUsageTest(ctx, "TestGetUsageHistory", t)
	currentTime := time.Now()
	layout := "2006-01-02 15:04:05"
	startDateTime := currentTime.AddDate(0, 0, -30).Format(layout)
	endDateTime := currentTime.AddDate(0, 0, 1).Format(layout)
	ariaUsageClient := common.GetUsageClient()
	cloudAcctId := GetCloudAcctIdFromClientAcctId(ClientAccountId)
	resp, err := ariaUsageClient.GetUsageHistory(ctx, ClientAccountId, client.GetClientMasterPlanInstanceId(cloudAcctId, ProductId), startDateTime, endDateTime)
	logger.Info("usage history ", "resp", resp)
	if err != nil {
		t.Fatalf("Failed to get usage history: %v", err)
	}
}

func TestGetUsageHistoryWithLimitedDecmial(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetUsageHistoryWithLimitedDecmial")
	logger.Info("testing usage history with limit decimal")
	ctx := context.Background()
	BeforeUsageTest(ctx, "TestGetUsageHistoryWithLimitedDecmial", t)
	currentTime := time.Now()
	layout := "2006-01-02 15:04:05"
	startDateTime := currentTime.AddDate(0, 0, -30).Format(layout)
	endDateTime := currentTime.AddDate(0, 0, 1).Format(layout)
	ariaUsageClient := common.GetUsageClient()
	cloudAcctId := GetCloudAcctIdFromClientAcctId(ClientAccountId)
	clientPlanId := GetTestClientPlanId(ProductId)
	err := CreateTestUsageRecord(ctx, ClientAccountId, clientPlanId, client.GetMinsUsageTypeCode(), 10.1267567809)
	if err != nil {
		t.Fatalf("Failed to create usage record: %v", err)
	}
	resp, err := ariaUsageClient.GetUsageHistory(ctx, ClientAccountId, client.GetClientMasterPlanInstanceId(cloudAcctId, ProductId), startDateTime, endDateTime)
	logger.Info("usage history ", "resp", resp)
	if err != nil {
		t.Fatalf("Failed to get usage history: %v", err)
	}
	found := false
	for _, usageRecord := range resp.UsageHistoryRecs {
		if usageRecord.Units == 10.13 {
			found = true
		}
	}
	if !found {
		t.Errorf(" invalid usage %v", resp)
	}
}

func setupStorgePlanUsage(ctx context.Context, testName string, t *testing.T, usageUnitType string) (string, string) {
	logger := log.FromContext(ctx).WithName("setupStorgePlanUsage")
	logger.Info("running storage usage setup tasks before test", "testName", testName)
	createPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	productId := product.GetId()
	err := EnableStorageUsageType()
	if err != nil {
		t.Fatalf("failed to create usage type: %v", err)
	}
	usageType, err := GetTestUsageUnitTypeCode(usageUnitType)
	if err != nil {
		t.Fatalf("failed to get usage type: %v", err)
	}
	clientAcctId := GetClientAccountId()
	logger.Info("storage usage setup tasks", "ClientAccountId", clientAcctId, "ProductId", productId)
	_, err = CreateAccountWithDefaultPlan(ctx, clientAcctId)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	_, err = createPlan.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}
	clientPlanId := GetTestClientPlanId(product.GetId())
	err = AssignPlanToAccount(ctx, clientAcctId, clientPlanId, ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to assign plan to account: %v", err)
	}
	err = CreateTestStorageUsageRecord(ctx, clientAcctId, clientPlanId, usageType.UsageTypeCode, 100, usageUnitType)
	if err != nil {
		t.Fatalf("failed to create bulk usage record: %v", err)
	}
	return clientAcctId, productId
}

func TestGetStorgePlanUsageHistory(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetStorgePlanUsageHistory")
	logger.Info("testing storage plan usage history")
	ctx := context.Background()
	clientAcctId, productId := setupStorgePlanUsage(ctx, "TestGetStorgePlanUsageHistory", t, "RATE_UNIT_TERABYTE_PER_HOUR")
	currentTime := time.Now()
	layout := "2006-01-02 15:04:05"
	startDateTime := currentTime.AddDate(0, 0, -30).Format(layout)
	endDateTime := currentTime.AddDate(0, 0, 1).Format(layout)
	ariaUsageClient := common.GetUsageClient()
	cloudAcctId := GetCloudAcctIdFromClientAcctId(clientAcctId)
	resp, err := ariaUsageClient.GetUsageHistory(ctx, clientAcctId, client.GetClientMasterPlanInstanceId(cloudAcctId, productId), startDateTime, endDateTime)
	logger.Info("usage history ", "resp", resp)
	if err != nil {
		t.Fatalf("Failed to get usage history: %v", err)
	}
}

func TestGetStorgePlanUnitTypeUsageHistory(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetStorgePlanUsageHistory")
	logger.Info("testing storage plan usage history")
	ctx := context.Background()
	clientAcctId, productId := setupStorgePlanUsage(ctx, "TestGetStorgePlanUnitUsageHistory", t, "per TB per Hour")
	currentTime := time.Now()
	layout := "2006-01-02 15:04:05"
	startDateTime := currentTime.AddDate(0, 0, -30).Format(layout)
	endDateTime := currentTime.AddDate(0, 0, 1).Format(layout)
	ariaUsageClient := common.GetUsageClient()
	cloudAcctId := GetCloudAcctIdFromClientAcctId(clientAcctId)
	resp, err := ariaUsageClient.GetUsageHistory(ctx, clientAcctId, client.GetClientMasterPlanInstanceId(cloudAcctId, productId), startDateTime, endDateTime)
	logger.Info("usage history ", "resp", resp)
	if err != nil {
		t.Fatalf("Failed to get usage history: %v", err)
	}
}
