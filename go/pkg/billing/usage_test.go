// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"testing"

	"github.com/google/uuid"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	meteringQuery "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/db/query"
	meteringServer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/usage"
)

func initializeTestsGetUsage(t *testing.T, ctx context.Context) {
	err := meteringQuery.DeleteAllRecords(ctx, meteringServer.MeteringDb)
	if err != nil {
		t.Fatalf("failed to delete all metering records: %v", err)
	}

	usageData := usage.NewUsageData(usage.GetUsageDb())
	err = usageData.DeleteAllUsages(ctx)
	if err != nil {
		t.Fatalf("could not delete all records")
	}
}

func createUsagesForGettingUsages(t *testing.T, ctx context.Context, cloudAcct *pb.CloudAccount, region string) {

	resourceId := uuid.NewString()
	transactionId := uuid.NewString()

	// use the usage tests here as this API is eventually going to be a part of usage service.
	meteringServiceClient := pb.NewMeteringServiceClient(meteringClientConn)
	_, err := meteringServiceClient.Create(context.Background(),
		usage.GetComputeUsageRecordCreate(cloudAcct.Id, resourceId, transactionId,
			region, usage.GetIdcComputeServiceName(), usage.GetXeon3SmallInstanceType(), 10))

	if err != nil {
		t.Fatalf("failed to create metering record: %v", err)
	}

	cloudAccountClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	usageData := usage.NewUsageData(usage.GetUsageDb())
	usageRecordData := usage.NewUsageRecordData(usage.GetUsageDb())
	usageController := usage.NewUsageController(*billingCommon.NewCloudAccountClientForTest(cloudAccountClient), productClient,
		billingCommon.NewMeteringClientForTest(meteringServiceClient), usageData, usageRecordData)
	usageController.CalculateUsages(ctx)

}

func TestBillingUsage(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestBillingUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	initializeTestsGetUsage(t, ctx)

	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)

	createUsagesForGettingUsages(t, ctx, premiumCloudAcct, usage.DefaultServiceRegion)

	billingUsageClient := pb.NewBillingUsageServiceClient(clientConn)
	billingUsageResponse, err := billingUsageClient.Read(ctx, &pb.BillingUsageFilter{CloudAccountId: premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to get billing usage response: %v", err)
	}

	verifyBillingUsages(t, ctx, premiumCloudAcct.Id, billingUsageResponse)
}

func TestBillingMultipleUsages(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestBillingMultipleUsages")
	logger.Info("BEGIN")
	defer logger.Info("End")

	initializeTestsGetUsage(t, ctx)

	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)

	createUsagesForGettingUsages(t, ctx, premiumCloudAcct, usage.DefaultServiceRegion)
	createUsagesForGettingUsages(t, ctx, premiumCloudAcct, "region-2")

	billingUsageClient := pb.NewBillingUsageServiceClient(clientConn)
	billingUsageResponse, err := billingUsageClient.Read(ctx, &pb.BillingUsageFilter{CloudAccountId: premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to get billing usage response: %v", err)
	}

	if len(billingUsageResponse.Usages) != 2 {
		t.Fatalf("should have had two usages")
	}

	verifyBillingUsages(t, ctx, premiumCloudAcct.Id, billingUsageResponse)
}

func verifyBillingUsages(t *testing.T, ctx context.Context, cloudAcctId string, billingUsageResponse *pb.BillingUsageResponse) {
	usageData := usage.NewUsageData(usage.GetUsageDb())
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &cloudAcctId})

	if err != nil {
		t.Fatalf("failed to search product usage records: %v", err)
	}

	var productUsageRecordsAmount float64 = 0
	var productUsageRecordsQty float64 = 0

	mappedProductUsages := map[string][]*pb.ProductUsage{}

	for _, productUsage := range productUsages.ProductUsages {
		productUsageRecordsAmount += productUsage.Rate * productUsage.Quantity
		mappedProductUsages[productUsage.ProductName] = append(mappedProductUsages[productUsage.ProductId],
			productUsage)
		productUsageRecordsQty += productUsage.Quantity
	}

	if billingUsageResponse.TotalAmount != productUsageRecordsAmount {
		t.Fatalf("incorrect amount")
	}

	if billingUsageResponse.TotalUsage != productUsageRecordsQty {
		t.Fatalf("incorrect quantity")
	}

	for _, billingUsageResponseUsage := range billingUsageResponse.Usages {
		if prodUsageRecords, hasProduct := mappedProductUsages[billingUsageResponseUsage.ProductType]; hasProduct {
			region := billingUsageResponseUsage.RegionName
			var prodUsageRecordsAmount float64 = 0
			var prodUsageRecordsQty float64 = 0
			var prodUsageRecordsUnitType string
			for _, prodUsageRecord := range prodUsageRecords {
				if region == prodUsageRecord.Region {
					prodUsageRecordsAmount += prodUsageRecord.Rate * prodUsageRecord.Quantity
					prodUsageRecordsQty += prodUsageRecord.Quantity
					prodUsageRecordsUnitType = prodUsageRecord.UsageUnitType
				}
			}
			if billingUsageResponseUsage.Amount != prodUsageRecordsAmount {
				t.Fatalf("incorrect product usage amount")
			}
			if billingUsageResponseUsage.MinsUsed != prodUsageRecordsQty {
				t.Fatalf("incorrect product usage qty")
			}
			if Cfg.GetFeaturesBillingUsageMetrics() {
				if billingUsageResponseUsage.BillingUsageMetrics.UsageQuantity != prodUsageRecordsQty {
					t.Fatalf("incorrect product usage quantity")
				}
				if billingUsageResponseUsage.BillingUsageMetrics.UsageUnitName != prodUsageRecordsUnitType {
					t.Fatalf("incorrect product usage unit type")
				}
			}

		} else {
			t.Fatalf("incorrect product type")
		}
	}
}

func createStorageUsagesForGettingUsages(t *testing.T, ctx context.Context, cloudAcct *pb.CloudAccount, region string) {

	resourceId := uuid.NewString()
	transactionId := uuid.NewString()
	timeQuantity := float32(10.0)
	storageQuantity := float32(10.0)
	// use the usage tests here as this API is eventually going to be a part of usage service.
	meteringServiceClient := pb.NewMeteringServiceClient(meteringClientConn)
	_, err := meteringServiceClient.Create(context.Background(),
		usage.GetStorageUsageRecordCreate(cloudAcct.Id, usage.GetIdcFileStoragServiceType(), resourceId, transactionId,
			region, usage.GetIdcFileStorageServiceName(), timeQuantity, storageQuantity))

	if err != nil {
		t.Fatalf("failed to create storage metering record: %v", err)
	}

	cloudAccountClient := pb.NewCloudAccountServiceClient(cloudAccountConn)
	usageData := usage.NewUsageData(usage.GetUsageDb())
	usageRecordData := usage.NewUsageRecordData(usage.GetUsageDb())
	usageController := usage.NewUsageController(*billingCommon.NewCloudAccountClientForTest(cloudAccountClient), productClient,
		billingCommon.NewMeteringClientForTest(meteringServiceClient), usageData, usageRecordData)
	usageController.CalculateUsages(ctx)

}

func TestBillingStorageUsage(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestBillingStorageUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")

	initializeTestsGetUsage(t, ctx)

	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, pb.AccountType_ACCOUNT_TYPE_PREMIUM)

	createStorageUsagesForGettingUsages(t, ctx, premiumCloudAcct, usage.DefaultServiceRegion)

	billingUsageClient := pb.NewBillingUsageServiceClient(clientConn)
	billingUsageResponse, err := billingUsageClient.Read(ctx, &pb.BillingUsageFilter{CloudAccountId: premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to get billing usage response: %v", err)
	}

	verifyBillingUsages(t, ctx, premiumCloudAcct.Id, billingUsageResponse)
}
