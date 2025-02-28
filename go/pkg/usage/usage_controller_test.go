// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestCalculateUsagesProductDoesntMatch(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestCalculateUsagesProductDoesntMatch")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	transactionId := "transaction-id"

	// push a metering record
	meteringRecordTimeMetric := 60
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, "invalidInstanceType", transactionId, float32(meteringRecordTimeMetric))

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)
	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	if len(resourceUsages.ResourceUsages) != 0 {
		t.Fatalf("invalid length of resource usages")
	}

	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	if len(productUsages.ProductUsages) != 0 {
		t.Fatalf("invalid length of product usages")
	}

	invalidMeteringRecords := GetAllInvalidMeteringRecords(ctx, t)

	if len(invalidMeteringRecords) != 1 {
		t.Fatalf("invalid length of invalid metering records")
	}

	if invalidMeteringRecords[0].CloudAccountId != premiumCloudAcct.Id {
		t.Fatalf("cloud account id does not match")
	}

	verifyAllMeteringReported(ctx, t, premiumCloudAcct.Id, resourceId, true)
}

func GetAllInvalidMeteringRecords(ctx context.Context, t *testing.T) []*pb.InvalidMeteringRecord {
	meteringServiceClient := pb.NewMeteringServiceClient(meteringConn)
	invalidMeteringRecordFilter := &pb.InvalidMeteringRecordFilter{}

	invalidMeteringSearchClient, err := meteringServiceClient.SearchInvalid(ctx, invalidMeteringRecordFilter)
	if err != nil {
		t.Fatalf("failed to get invalid metering search client: %v", err)
	}

	var invalidMeteringRecords []*pb.InvalidMeteringRecord
	for {
		invalidMeteringRecordR, err := invalidMeteringSearchClient.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				t.Fatalf("invalid error received for searching records: %v", err)
			}
		}
		invalidMeteringRecords = append(invalidMeteringRecords, invalidMeteringRecordR)
	}

	return invalidMeteringRecords
}

// todo: this test can be broken down and refactored and both.
func TestCalculateUsages(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestCalculateUsages")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	transactionId := "transaction-id"

	// push a metering record
	meteringRecordTimeMetric := 60
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType, transactionId, float32(meteringRecordTimeMetric))

	mappedProducts := getComputeMappedProducts(ctx, t, xeon3SmallInstanceType)

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)
	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	mappedResourceUsages := map[string]*pb.ResourceUsage{}
	// verify resource usages
	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	for _, resourceUsage := range resourceUsages.ResourceUsages {
		mappedResourceUsages[resourceUsage.Id] = resourceUsage
	}

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	lengthOfMappedProducts := len(mappedProducts)
	// todo: need to make sure the product catalog is populated as per needed
	// to verify the expected size of the resource usages.
	if len(resourceUsages.ResourceUsages) != lengthOfMappedProducts {
		t.Fatalf("invalid length of resource usages")
	}

	for _, resourceUsage := range resourceUsages.ResourceUsages {
		verifyResourceUsage(t, resourceUsage, premiumCloudAcct.Id, resourceId, float64(meteringRecordTimeMetric/60), acctType,
			DefaultServiceRegion, mappedProducts)
		if resourceUsage.UnReportedQuantity != float64(meteringRecordTimeMetric/60) {
			t.Fatalf("invalid unreported amount")
		}
	}

	mappedProductUsages := map[string]*pb.ProductUsage{}
	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		mappedProductUsages[productUsageRecord.Id] = productUsageRecord
	}

	if len(productUsages.ProductUsages) != lengthOfMappedProducts {
		t.Fatalf("invalid length of product usages")
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		verifyProductUsage(t, productUsageRecord, premiumCloudAcct.Id, acctType, float64(meteringRecordTimeMetric/60),
			DefaultServiceRegion, mappedProducts)
	}

	// verify metering record for resource.
	meteringRecordForResource, err := usageData.GetMeteringForResource(ctx, resourceId)
	if err != nil {
		t.Fatalf("failed to get metering record for resource")
	}

	verifyResourceMeteringRecord(t, meteringRecordForResource, premiumCloudAcct.Id, resourceId, transactionId, DefaultServiceRegion)

	// verify metering records.
	verifyAllMeteringReported(ctx, t, premiumCloudAcct.Id, resourceId, true)

	// run again and nothing should change as metering is already reported..
	usageController.CalculateUsages(ctx)

	resourceUsagesRepeat, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	if len(resourceUsagesRepeat.ResourceUsages) != lengthOfMappedProducts {
		t.Fatalf("invalid length of resource usages upon rerun")
	}

	for _, resourceUsageRecordRepeat := range resourceUsagesRepeat.ResourceUsages {
		if resourceUsageRecord, hasResourceUsageRecord := mappedResourceUsages[resourceUsageRecordRepeat.Id]; hasResourceUsageRecord {

			if resourceUsageRecordRepeat.CloudAccountId != resourceUsageRecord.CloudAccountId ||
				resourceUsageRecordRepeat.TransactionId != resourceUsageRecord.TransactionId {
				t.Fatalf("resource usages don't match upon rerun")
			}
		} else {
			t.Fatalf("resource usages don't match upon rerun")
		}
	}

	productUsagesRepeat, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	if len(productUsagesRepeat.ProductUsages) != lengthOfMappedProducts {
		t.Fatalf("invalid length of product usages upon rerun")
	}

	for _, productUsageRecordRepeat := range productUsagesRepeat.ProductUsages {
		if productUsageRecord, hasProductUsageRecord := mappedProductUsages[productUsageRecordRepeat.Id]; hasProductUsageRecord {

			if productUsageRecordRepeat.CloudAccountId != productUsageRecord.CloudAccountId ||
				productUsageRecordRepeat.ProductName != productUsageRecord.ProductName {
				t.Fatalf("product usages don't match upon rerun")
			}
		} else {
			t.Fatalf("product usages don't match upon rerun")
		}
	}

}

func TestCalculateUsagesPrevMeteringReported(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestCalculateUsagesPrevMeteringReported")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	transactionId := "transaction-id"

	// push a metering record
	meteringRecordTimeMetric := 60

	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType, transactionId, float32(meteringRecordTimeMetric))

	mappedProducts := getComputeMappedProducts(ctx, t, xeon3SmallInstanceType)

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)
	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	mappedResourceUsages := map[string]*pb.ResourceUsage{}
	// verify resource usages
	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		mappedResourceUsages[resourceUsageRecord.Id] = resourceUsageRecord
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		verifyResourceUsage(t, resourceUsageRecord, premiumCloudAcct.Id, resourceId, float64(meteringRecordTimeMetric/60), acctType,
			DefaultServiceRegion, mappedProducts)
		if resourceUsageRecord.UnReportedQuantity != float64(meteringRecordTimeMetric/60) {
			t.Fatalf("invalid unreported amount")
		}
	}

	mappedProductUsages := map[string]*pb.ProductUsage{}
	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		mappedProductUsages[productUsageRecord.Id] = productUsageRecord
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		verifyProductUsage(t, productUsageRecord, premiumCloudAcct.Id, acctType, float64(meteringRecordTimeMetric/60),
			DefaultServiceRegion, mappedProducts)
	}
}

func TestCalculateUsagesMultMeteringRecords(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestCalculateUsagesMultMeteringRecords")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	transactionId := "transaction-id"
	transactionId1 := "transaction-id-1"

	// push a metering record
	meteringRecordTimeMetric := 60
	meteringRecordAmount1 := 120
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType, transactionId, float32(meteringRecordTimeMetric))
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType, transactionId1, float32(meteringRecordAmount1))

	mappedProducts := getComputeMappedProducts(ctx, t, xeon3SmallInstanceType)

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	// verify resource usages
	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	lengthOfMappedProducts := len(mappedProducts)
	// todo: need to make sure the product catalog is populated as per needed
	// to verify the expected size of the resource usages.
	if len(resourceUsages.ResourceUsages) != lengthOfMappedProducts {
		t.Fatalf("invalid length of resource usages")
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		verifyResourceUsage(t, resourceUsageRecord, premiumCloudAcct.Id, resourceId, float64(meteringRecordAmount1/60), acctType,
			DefaultServiceRegion, mappedProducts)
		if resourceUsageRecord.UnReportedQuantity != float64(meteringRecordAmount1/60) {
			t.Fatalf("invalid unreported amount")
		}
	}

	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	if len(productUsages.ProductUsages) != lengthOfMappedProducts {
		t.Fatalf("invalid length of product usages")
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		verifyProductUsage(t, productUsageRecord, premiumCloudAcct.Id, acctType, float64(meteringRecordAmount1/60),
			DefaultServiceRegion, mappedProducts)
	}

	// verify metering record for resource.
	meteringRecordForResource, err := usageData.GetMeteringForResource(ctx, resourceId)
	if err != nil {
		t.Fatalf("failed to get metering record for resource")
	}

	verifyResourceMeteringRecord(t, meteringRecordForResource, premiumCloudAcct.Id, resourceId, transactionId1, DefaultServiceRegion)

	// verify metering records.
	verifyAllMeteringReported(ctx, t, premiumCloudAcct.Id, resourceId, true)
}

func TestCalculateUsagesMultiple(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestCalculateUsagesMultiple")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	resourceMeteringTransactionId := "resource-transaction-id"
	resourceMeteringTransactionId1 := "resource-transaction-id-1"

	resource1Id := "resource1-id"
	resource1MeteringTransactionId := "resource1-transaction-id"
	resource1MeteringTransactionId1 := "resource1-transaction-id-1"

	// push a metering record
	meteringRecordTimeMetric := 60
	meteringRecordTimeMetric1 := 120
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType,
		resourceMeteringTransactionId, float32(meteringRecordTimeMetric))
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType,
		resourceMeteringTransactionId1, float32(meteringRecordTimeMetric1))

	resource1MeteringRecordAmount := 180
	resource1MeteringRecordAmount1 := 240
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resource1Id, xeon3SmallInstanceType,
		resource1MeteringTransactionId, float32(resource1MeteringRecordAmount))
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resource1Id, xeon3SmallInstanceType,
		resource1MeteringTransactionId1, float32(resource1MeteringRecordAmount1))

	mappedProducts := getComputeMappedProducts(ctx, t, xeon3SmallInstanceType)

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	// verify resource usages
	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	lengthOfMappedProducts := len(mappedProducts)
	// todo: need to make sure the product catalog is populated as per needed
	// to verify the expected size of the resource usages.
	if len(resourceUsages.ResourceUsages) != 2*lengthOfMappedProducts {
		t.Fatalf("invalid length of resource usages")
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		var resourceIdToVerify string
		var amountToVerify float64
		var unreportedAmountToVerify float64
		if resourceId == resourceUsageRecord.ResourceId {
			resourceIdToVerify = resourceId
			amountToVerify = float64(meteringRecordTimeMetric1 / 60)
			unreportedAmountToVerify = float64(meteringRecordTimeMetric1 / 60)
		} else {
			resourceIdToVerify = resource1Id
			amountToVerify = float64(resource1MeteringRecordAmount1 / 60)
			unreportedAmountToVerify = float64(resource1MeteringRecordAmount1 / 60)
		}
		verifyResourceUsage(t, resourceUsageRecord, premiumCloudAcct.Id, resourceIdToVerify,
			amountToVerify, acctType,
			DefaultServiceRegion, mappedProducts)
		if resourceUsageRecord.UnReportedQuantity != unreportedAmountToVerify {
			t.Fatalf("invalid unreported amount")
		}
	}

	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	if len(productUsages.ProductUsages) != 2*lengthOfMappedProducts {
		t.Fatalf("invalid length of product usages")
	}

	// verify metering records.
	verifyAllMeteringReported(ctx, t, premiumCloudAcct.Id, resourceId, true)
}

func TestCalculateUsagesMultipleRuns(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestCalculateUsagesMultiple")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	resourceMeteringTransactionId := "resource-transaction-id"
	resourceMeteringTransactionId1 := "resource-transaction-id-1"

	// push a metering record
	meteringRecordTimeMetric := 60
	meteringRecordTimeMetric1 := 120
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType,
		resourceMeteringTransactionId, float32(meteringRecordTimeMetric))

	mappedProducts := getComputeMappedProducts(ctx, t, xeon3SmallInstanceType)

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType,
		resourceMeteringTransactionId1, float32(meteringRecordTimeMetric1))

	usageController.CalculateUsages(ctx)

	// verify resource usages
	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	lengthOfMappedProducts := len(mappedProducts)
	// todo: need to make sure the product catalog is populated as per needed
	// to verify the expected size of the resource usages.
	if len(resourceUsages.ResourceUsages) != 2*lengthOfMappedProducts {
		t.Fatalf("invalid length of resource usages")
	}

	verifyResourceUsage(t, resourceUsages.ResourceUsages[0], premiumCloudAcct.Id, resourceId,
		float64(meteringRecordTimeMetric/60), acctType, DefaultServiceRegion, mappedProducts)
	if resourceUsages.ResourceUsages[0].UnReportedQuantity != float64(meteringRecordTimeMetric/60) {
		t.Fatalf("invalid unreported amount")
	}

	verifyResourceUsage(t, resourceUsages.ResourceUsages[1], premiumCloudAcct.Id, resourceId,
		float64((meteringRecordTimeMetric1-meteringRecordTimeMetric)/60), acctType, DefaultServiceRegion, mappedProducts)
	if resourceUsages.ResourceUsages[0].UnReportedQuantity !=
		float64((meteringRecordTimeMetric1-meteringRecordTimeMetric)/60) {
		t.Fatalf("invalid unreported amount")
	}

	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	if len(productUsages.ProductUsages) != 2*lengthOfMappedProducts {
		t.Fatalf("invalid length of product usages")
	}

	// verify metering records.
	verifyAllMeteringReported(ctx, t, premiumCloudAcct.Id, resourceId, true)
}

func TestCalculateUsagesAcrossTypes(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestCalculateUsagesIntel")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	cloudAcctTypes := []pb.AccountType{pb.AccountType_ACCOUNT_TYPE_INTEL,
		pb.AccountType_ACCOUNT_TYPE_STANDARD, pb.AccountType_ACCOUNT_TYPE_ENTERPRISE}

	for _, acctType := range cloudAcctTypes {
		// verify for single cloud account and single resource
		// create premium cloud account
		user := uuid.NewString() + "@example.com"
		resourceId := "resource-id"
		transactionId := "transaction-id"

		if acctType == pb.AccountType_ACCOUNT_TYPE_INTEL {
			user = "intel-" + user
			resourceId = "intel-" + resourceId
			transactionId = "intel-" + transactionId
		} else if acctType == pb.AccountType_ACCOUNT_TYPE_STANDARD {
			user = "standard-" + user
			resourceId = "standard-" + resourceId
			transactionId = "standard-" + transactionId
		} else if acctType == pb.AccountType_ACCOUNT_TYPE_ENTERPRISE {
			user = "enterprise-" + user
			resourceId = "enterprise-" + resourceId
			transactionId = "enterprise-" + transactionId
		}
		cloudAcct := CreateAndGetCloudAccount(t, ctx, user, acctType)

		// push a metering record
		meteringRecordTimeMetric := 60
		CreateComputeMeteringRecord(ctx, t, cloudAcct.Id, resourceId, xeon3SmallInstanceType, transactionId, float32(meteringRecordTimeMetric))

		mappedProducts := getComputeMappedProducts(ctx, t, xeon3SmallInstanceType)

		usageData := NewUsageData(usageDb)
		usageRecordData := NewUsageRecordData(usageDb)

		usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
		// call the scheduled job.
		usageController.CalculateUsages(ctx)

		// verify resource usages
		resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &cloudAcct.Id})

		if err != nil {
			t.Fatalf("failed to search resource usages: %v", err)
		}

		lengthOfMappedProducts := len(mappedProducts)
		// todo: need to make sure the product catalog is populated as per needed
		// to verify the expected size of the resource usages.
		if len(resourceUsages.ResourceUsages) != lengthOfMappedProducts {
			t.Fatalf("invalid length of resource usages")
		}

		for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
			verifyResourceUsage(t, resourceUsageRecord, cloudAcct.Id, resourceId, float64(meteringRecordTimeMetric/60), acctType,
				DefaultServiceRegion, mappedProducts)
			if resourceUsageRecord.UnReportedQuantity != float64(meteringRecordTimeMetric/60) {
				t.Fatalf("invalid unreported amount")
			}
		}

		// verify product usages.
		productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &cloudAcct.Id})

		if err != nil {
			t.Fatalf("failed to search product usages: %v", err)
		}

		if len(productUsages.ProductUsages) != lengthOfMappedProducts {
			t.Fatalf("invalid length of product usages")
		}

		for _, productUsageRecord := range productUsages.ProductUsages {
			verifyProductUsage(t, productUsageRecord, cloudAcct.Id, acctType, float64(meteringRecordTimeMetric/60),
				DefaultServiceRegion, mappedProducts)
		}

		// verify metering record for resource.
		meteringRecordForResource, err := usageData.GetMeteringForResource(ctx, resourceId)
		if err != nil {
			t.Fatalf("failed to get metering record for resource")
		}

		verifyResourceMeteringRecord(t, meteringRecordForResource, cloudAcct.Id, resourceId, transactionId, DefaultServiceRegion)

		// verify metering records.
		verifyAllMeteringReported(ctx, t, cloudAcct.Id, resourceId, true)
	}
}

func verifyAllProductUsageRecordsReported(t *testing.T, ctx context.Context, cloudAcctId string, reported bool) {

	productUsageRecordsFilter := &pb.ProductUsageRecordsFilter{
		CloudAccountId: &cloudAcctId,
	}

	usageRecordServiceClient := pb.NewUsageRecordServiceClient(clientConn)
	stream, _ := usageRecordServiceClient.SearchProductUsageRecords(ctx, productUsageRecordsFilter)

	for {

		productUsageRecord, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("unexpected error for searching product usage records %v", err)
		}

		if productUsageRecord.Reported != reported {
			t.Fatalf("value of reported does not match")
		}
	}

}

func verifyAllMeteringReported(ctx context.Context, t *testing.T, cloudAcctId string, resourceId string, reported bool) {
	meteringServiceClient := pb.NewMeteringServiceClient(meteringConn)
	meteringSearchClient, err := meteringServiceClient.Search(ctx,
		&pb.UsageFilter{
			CloudAccountId: &cloudAcctId,
			ResourceId:     &resourceId,
		})

	if err != nil {
		t.Fatalf("failed to get metering search client: %v", err)
	}

	for {
		meteringRecord, err := meteringSearchClient.Recv()
		if err == io.EOF {
			meteringSearchClient.CloseSend()
			return
		}
		if err != nil {
			t.Fatalf("failed to read metering records: %v", err)
		}
		if meteringRecord.Reported != reported {
			t.Fatalf("value of reported does not match")
		}
	}
}

func getComputeMappedProducts(ctx context.Context, t *testing.T, instanceType string) map[string]*pb.Product {
	mappedProducts := map[string]*pb.Product{}
	products, err := productClient.GetProductCatalogProducts(ctx)
	if err != nil {
		t.Fatalf("failed to get products: %v", err)
	}
	for _, product := range products {
		// use a simple matching.. have to make sure product catalog matches are not complex.
		if strings.Contains(product.MatchExpr, instanceType) {
			mappedProducts[product.Id] = product
		}
	}
	return mappedProducts
}

func getStorageMappedProducts(ctx context.Context, t *testing.T, serviceType string) map[string]*pb.Product {
	mappedProducts := map[string]*pb.Product{}
	products, err := productClient.GetProductCatalogProducts(ctx)
	if err != nil {
		t.Fatalf("failed to get products: %v", err)
	}
	for _, product := range products {
		// use a simple matching.. have to make sure product catalog matches are not complex.
		if strings.Contains(product.MatchExpr, serviceType) {
			mappedProducts[product.Id] = product
		}
	}
	return mappedProducts
}

func verifyResourceMeteringRecord(t *testing.T, meteringRecordForResource *ResourceMetering,
	cloudAccountId string, resourceId string, transactionId string, region string) {
	if meteringRecordForResource.CloudAccountId != cloudAccountId {
		t.Fatalf("cloud account id does not match")
	}
	if meteringRecordForResource.ResourceId != resourceId {
		t.Fatalf("resource id does not match")
	}
	if meteringRecordForResource.TransactionId != transactionId {
		t.Fatalf("transaction id does not match")
	}
	if meteringRecordForResource.Region != region {
		t.Fatalf("region does not match")
	}
}

func verifyResourceUsageNoQty(t *testing.T,
	resourceUsage *pb.ResourceUsage, cloudAccountId string,
	resourceId string, cloudAccountType pb.AccountType, region string, mappedProducts map[string]*pb.Product) {
	if resourceUsage.CloudAccountId != cloudAccountId {
		t.Fatalf("cloud account id does not match")
	}
	if resourceUsage.ResourceId != resourceId {
		t.Fatalf("resource id does not match")
	}
	if resourceUsage.Region != region {
		t.Fatalf("region does not match")
	}
	if product, hasProduct := mappedProducts[resourceUsage.ProductId]; hasProduct {
		if resourceUsage.ProductName != product.Name {
			t.Fatalf("product name does not match")
		}
		// there are too many rates because of the call in productClient.GetProductsForAccountTypes
		// it keeps appending the rates.
		for _, rate := range product.Rates {
			if rate.AccountType == cloudAccountType {
				if fmt.Sprint(resourceUsage.Rate) != rate.Rate {
					t.Fatalf("rate does not match")
				}
				if resourceUsage.UsageUnitType != rate.Unit.String() {
					t.Fatalf("usage unit type does not match")
				}
			}
		}
	} else {
		t.Fatalf("invalid product id")
	}
}

func verifyResourceUsage(t *testing.T,
	resourceUsage *pb.ResourceUsage, cloudAccountId string,
	resourceId string, expectedQuantity float64, cloudAccountType pb.AccountType, region string, mappedProducts map[string]*pb.Product) {
	verifyResourceUsageNoQty(t,
		resourceUsage, cloudAccountId,
		resourceId, cloudAccountType, region, mappedProducts)

	if resourceUsage.Quantity != expectedQuantity {
		t.Fatalf("quantity does not match")
	}
}

func verifyProductUsageNoQty(t *testing.T,
	productUsage *pb.ProductUsage, cloudAccountId string,
	cloudAccountType pb.AccountType, region string, mappedProducts map[string]*pb.Product) {
	if productUsage.CloudAccountId != cloudAccountId {
		t.Fatalf("cloud account id does not match")
	}
	if productUsage.Region != region {
		t.Fatalf("region does not match")
	}
	if product, hasProduct := mappedProducts[productUsage.ProductId]; hasProduct {
		if productUsage.ProductName != product.Name {
			t.Fatalf("product name does not match")
		}
		// there are too many rates because of the call in productClient.GetProductsForAccountTypes
		// it keeps appending the rates.
		for _, rate := range product.Rates {
			if rate.AccountType == cloudAccountType {
				if fmt.Sprint(productUsage.Rate) != rate.Rate {
					t.Fatalf("rate does not match")
				}
				if productUsage.UsageUnitType != rate.Unit.String() {
					t.Fatalf("usage unit type does not match")
				}
			}
		}
	} else {
		t.Fatalf("invalid product id")
	}
}

func verifyProductUsage(t *testing.T,
	productUsage *pb.ProductUsage, cloudAccountId string,
	cloudAccountType pb.AccountType, expectedQuantity float64, region string, mappedProducts map[string]*pb.Product) {
	verifyProductUsageNoQty(t,
		productUsage, cloudAccountId,
		cloudAccountType, region, mappedProducts)
	if productUsage.Quantity != expectedQuantity {
		t.Fatalf("quantity does not match")
	}

}

func TestCalculateUsagesMultipleRuns1(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCalculateUsagesMultiple")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	resourceMeteringTransactionId := "resource-transaction-id"
	resourceMeteringTransactionId1 := "resource-transaction-id-1"
	resourceMeteringTransactionId2 := "resource-transaction-id-2"

	// push a metering record
	meteringRecordTimeMetric := 60
	meteringRecordTimeMetric1 := 180
	meteringRecordTimeMetric2 := 360
	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType,
		resourceMeteringTransactionId, float32(meteringRecordTimeMetric))

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType,
		resourceMeteringTransactionId1, float32(meteringRecordTimeMetric1))

	usageController.CalculateUsages(ctx)

	CreateComputeMeteringRecord(ctx, t, premiumCloudAcct.Id, resourceId, xeon3SmallInstanceType,
		resourceMeteringTransactionId2, float32(meteringRecordTimeMetric2))

	usageController.CalculateUsages(ctx)

	// verify resource usages
	_, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

}

func TestStorageCalcUsagesSingleMeteringRec(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestStorageCalcUsagesSingleMeteringRec")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	transactionId := "transaction-id"

	// push a metering record
	meteringRecordTimeMetric := 60
	meteringRecordStorageMetric := 2
	CreateStorageMeteringRecord(ctx, t, premiumCloudAcct.Id, FileStorageServiceType, resourceId, transactionId,
		float32(meteringRecordTimeMetric), float32(meteringRecordStorageMetric))

	mappedProducts := getStorageMappedProducts(ctx, t, FileStorageServiceType)

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	mappedResourceUsages := map[string]*pb.ResourceUsage{}
	// verify resource usages
	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		mappedResourceUsages[resourceUsageRecord.Id] = resourceUsageRecord
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		verifyResourceUsage(t, resourceUsageRecord, premiumCloudAcct.Id, resourceId,
			float64(meteringRecordTimeMetric*meteringRecordStorageMetric), acctType,
			DefaultServiceRegion, mappedProducts)
		if resourceUsageRecord.UnReportedQuantity != float64(meteringRecordTimeMetric*meteringRecordStorageMetric) {
			t.Fatalf("invalid unreported amount")
		}
	}

	mappedProductUsages := map[string]*pb.ProductUsage{}
	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		mappedProductUsages[productUsageRecord.Id] = productUsageRecord
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		verifyProductUsage(t, productUsageRecord, premiumCloudAcct.Id, acctType, float64(meteringRecordTimeMetric*meteringRecordStorageMetric),
			DefaultServiceRegion, mappedProducts)
	}
}

func TestStorageCalcUsagesTwoMeteringRecSameStorageQty(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestStorageCalcUsagesTwoMeteringRecSameStorageQty")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	transactionId := "transaction-id"
	transactionId1 := "transaction-id-1"

	// push a metering record
	meteringRecordTimeMetric := 60
	meteringRecordTimeMetric1 := meteringRecordTimeMetric * 2
	meteringRecordStorageMetric := 2
	CreateStorageMeteringRecord(ctx, t, premiumCloudAcct.Id, FileStorageServiceType, resourceId, transactionId,
		float32(meteringRecordTimeMetric), float32(meteringRecordStorageMetric))

	CreateStorageMeteringRecord(ctx, t, premiumCloudAcct.Id, FileStorageServiceType, resourceId, transactionId1,
		float32(meteringRecordTimeMetric1), float32(meteringRecordStorageMetric))
	mappedProducts := getStorageMappedProducts(ctx, t, FileStorageServiceType)

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	mappedResourceUsages := map[string]*pb.ResourceUsage{}
	// verify resource usages
	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		mappedResourceUsages[resourceUsageRecord.Id] = resourceUsageRecord
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		verifyResourceUsage(t, resourceUsageRecord, premiumCloudAcct.Id, resourceId,
			float64(meteringRecordTimeMetric*meteringRecordStorageMetric), acctType,
			DefaultServiceRegion, mappedProducts)
		if resourceUsageRecord.UnReportedQuantity != float64(meteringRecordTimeMetric*meteringRecordStorageMetric) {
			t.Fatalf("invalid unreported amount")
		}
	}

	mappedProductUsages := map[string]*pb.ProductUsage{}
	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		mappedProductUsages[productUsageRecord.Id] = productUsageRecord
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		verifyProductUsage(t, productUsageRecord, premiumCloudAcct.Id, acctType, float64(meteringRecordTimeMetric*meteringRecordStorageMetric),
			DefaultServiceRegion, mappedProducts)
	}
}

func TestStorageCalcUsagesTwoMeteringRecDiffStorageQty(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestStorageCalcUsagesTwoMeteringRecDiffStorageQty")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"
	transactionId := "transaction-id"
	transactionId1 := "transaction-id-1"

	// push a metering record
	meteringRecordTimeMetric := 60
	meteringRecordTimeMetric1 := meteringRecordTimeMetric * 2
	meteringRecordStorageMetric := 2
	meteringRecordStorageMetric1 := 3
	CreateStorageMeteringRecord(ctx, t, premiumCloudAcct.Id, FileStorageServiceType, resourceId, transactionId,
		float32(meteringRecordTimeMetric), float32(meteringRecordStorageMetric))

	CreateStorageMeteringRecord(ctx, t, premiumCloudAcct.Id, FileStorageServiceType, resourceId, transactionId1,
		float32(meteringRecordTimeMetric1), float32(meteringRecordStorageMetric1))

	mappedProducts := getStorageMappedProducts(ctx, t, FileStorageServiceType)

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateUsages(ctx)

	mappedResourceUsages := map[string]*pb.ResourceUsage{}
	// verify resource usages
	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		mappedResourceUsages[resourceUsageRecord.Id] = resourceUsageRecord
	}

	for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
		verifyResourceUsage(t, resourceUsageRecord, premiumCloudAcct.Id, resourceId,
			float64(meteringRecordTimeMetric*meteringRecordStorageMetric), acctType,
			DefaultServiceRegion, mappedProducts)
		if resourceUsageRecord.UnReportedQuantity != float64(meteringRecordTimeMetric*meteringRecordStorageMetric) {
			t.Fatalf("invalid unreported amount")
		}
	}

	mappedProductUsages := map[string]*pb.ProductUsage{}
	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		mappedProductUsages[productUsageRecord.Id] = productUsageRecord
	}

	for _, productUsageRecord := range productUsages.ProductUsages {
		verifyProductUsage(t, productUsageRecord, premiumCloudAcct.Id, acctType, float64(meteringRecordTimeMetric*meteringRecordStorageMetric),
			DefaultServiceRegion, mappedProducts)
	}
}

func TestStorageCalcUsagesMultipleMeteringRecSameStorageQty(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestStorageCalcUsagesMultipleMeteringRecSameStorageQty")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)

	resourceId := "resource-id"

	meteringRecordTimeMetric := 60
	meteringRecordStorageMetric := 1

	for count := 3; count <= 5; count++ {
		// do test initialization
		InitializeTests(ctx, t)

		for meteringRecordCount := 1; meteringRecordCount <= count; meteringRecordCount++ {
			transactionId := fmt.Sprintf("transaction-id-%d", meteringRecordCount)
			CreateStorageMeteringRecord(ctx, t, premiumCloudAcct.Id, FileStorageServiceType, resourceId, transactionId,
				float32(meteringRecordTimeMetric*meteringRecordCount), float32(meteringRecordStorageMetric))
		}

		mappedProducts := getStorageMappedProducts(ctx, t, FileStorageServiceType)
		usageData := NewUsageData(usageDb)
		usageRecordData := NewUsageRecordData(usageDb)

		usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
		// call the scheduled job.
		usageController.CalculateUsages(ctx)

		mappedResourceUsages := map[string]*pb.ResourceUsage{}
		// verify resource usages
		resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

		if err != nil {
			t.Fatalf("failed to search resource usages: %v", err)
		}

		for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
			mappedResourceUsages[resourceUsageRecord.Id] = resourceUsageRecord
		}

		for _, resourceUsageRecord := range resourceUsages.ResourceUsages {
			verifyResourceUsage(t, resourceUsageRecord, premiumCloudAcct.Id, resourceId,
				float64((count-1)*meteringRecordTimeMetric*meteringRecordStorageMetric), acctType,
				DefaultServiceRegion, mappedProducts)
			if resourceUsageRecord.UnReportedQuantity != float64((count-1)*meteringRecordTimeMetric*meteringRecordStorageMetric) {
				t.Fatalf("invalid unreported amount")
			}
		}

		mappedProductUsages := map[string]*pb.ProductUsage{}
		// verify product usages.
		productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

		if err != nil {
			t.Fatalf("failed to search product usages: %v", err)
		}

		for _, productUsageRecord := range productUsages.ProductUsages {
			mappedProductUsages[productUsageRecord.Id] = productUsageRecord
		}

		for _, productUsageRecord := range productUsages.ProductUsages {
			verifyProductUsage(t, productUsageRecord, premiumCloudAcct.Id, acctType,
				float64((count-1)*meteringRecordTimeMetric*meteringRecordStorageMetric), DefaultServiceRegion, mappedProducts)
		}
	}

}

func TestCalculateUsagesUsingProductUsageRecords(t *testing.T) {
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestCalculateUsagesUsingProductUsageRecords")
	logger.Info("BEGIN")
	defer logger.Info("End")

	// do test initialization
	InitializeTests(ctx, t)

	// verify for single cloud account and single resource
	// create premium cloud account
	premiumUser := "premium-" + uuid.NewString() + "@example.com"
	acctType := pb.AccountType_ACCOUNT_TYPE_PREMIUM
	premiumCloudAcct := CreateAndGetCloudAccount(t, ctx, premiumUser, acctType)
	productName := GetXeon3SmallInstanceType()
	region := DefaultServiceRegion

	transactionId := "transaction-id"

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	productUsageRecordCreate := GetProductUsageRecordCreate(premiumCloudAcct.Id, &productName, region,
		transactionId, defaultProductUsageRecordQty,
		map[string]string{
			"availabilityZone": region,
			"instanceType":     GetXeon3SmallInstanceType(),
			"service":          GetIdcComputeServiceName(),
		})

	err := usageRecordData.StoreProductUsageRecord(ctx, productUsageRecordCreate)

	if err != nil {
		t.Fatalf("unexpected error for storing product usage record %v", err)
	}

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	// call the scheduled job.
	usageController.CalculateProductUsages(ctx)

	resourceUsages, err := usageData.SearchResourceUsages(ctx, &pb.ResourceUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search resource usages: %v", err)
	}

	if len(resourceUsages.ResourceUsages) != 0 {
		t.Fatalf("invalid length of resource usages")
	}

	// verify product usages.
	productUsages, err := usageData.SearchProductUsages(ctx, &pb.ProductUsagesFilter{CloudAccountId: &premiumCloudAcct.Id})

	if err != nil {
		t.Fatalf("failed to search product usages: %v", err)
	}

	if len(productUsages.ProductUsages) != 1 {
		t.Fatalf("invalid length of product usages")
	}

	if productUsages.ProductUsages[0].Quantity != defaultProductUsageRecordQty ||
		productUsages.ProductUsages[0].CloudAccountId != premiumCloudAcct.Id {
		t.Fatalf("expected values do not match")
	}
	verifyAllProductUsageRecordsReported(t, ctx, premiumCloudAcct.Id, true)
}
