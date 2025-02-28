// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	clientTestCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	meteringQuery "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/db/query"
	meteringServer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/usage"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var cloudAcctId string

func getMappedProducts(ctx context.Context, t *testing.T, instanceType string) map[string]*pb.Product {
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

func initializeTests(t *testing.T, ctx context.Context) {
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

// Todo: Fix the test to have a more cleaner setup..
func BeforeUsageControllerTest(ctx context.Context, testName string, t *testing.T) {
	premiumUser := "premium-" + uuid.NewString() + "@example.com"

	cloudAccountClient := AriaService.cloudAccountClient
	cloudAccountId, err := cloudAccountClient.Create(ctx,
		&pb.CloudAccountCreate{
			Name:  premiumUser,
			Owner: premiumUser,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}
	cloudAcctId = cloudAccountId.GetId()

	resource1Id := uuid.NewString()
	transaction1Id := uuid.NewString()

	meteringServiceClient := pb.NewMeteringServiceClient(meteringClientConn)
	_, err = meteringServiceClient.Create(context.Background(),
		usage.GetComputeUsageRecordCreate(cloudAcctId, resource1Id, transaction1Id,
			usage.DefaultServiceRegion, usage.GetIdcComputeServiceName(), usage.GetXeon3SmallInstanceType(), 10))

	if err != nil {
		t.Fatalf("failed to create metering record: %v", err)
	}

	usageData := usage.NewUsageData(usage.GetUsageDb())
	usageRecordData := usage.NewUsageRecordData(usage.GetUsageDb())
	uC := usage.NewUsageController(*billingCommon.NewCloudAccountClientForTest(cloudAccountClient), productClient,
		billingCommon.NewMeteringClientForTest(meteringServiceClient), usageData, usageRecordData)
	uC.CalculateUsages(ctx)

	createPlan := clientTestCommon.GetAriaPlanClient()
	usageTypeClient := clientTestCommon.GetAriaUsageTypeClient()
	ariaAccountClient := clientTestCommon.GetAriaAccountClient()

	// Do the pre-requisites of the testing
	ariaController := NewAriaController(clientTestCommon.GetAriaClient(), clientTestCommon.GetAriaAdminClient(), clientTestCommon.GetAriaCredentials())
	err = ariaController.InitAria(ctx)
	if err != nil {
		t.Fatalf("failed to initialize aria: %v", err)
	}
	// get the usage type of the type minutes
	usageType, err := usageTypeClient.GetMinutesUsageType(context.Background())
	if err != nil {
		t.Fatalf("failed to get usage type: %v", err)
	}
	_, err = ariaAccountClient.CreateAriaAccount(ctx, client.GetAccountClientId(cloudAccountId.Id), client.GetDefaultPlanClientId(), client.ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	_, _, productFamilies := usage.BuildStaticProductsAndVendors()

	products := getMappedProducts(ctx, t, usage.GetXeon3SmallInstanceType())

	for _, product := range products {
		_, err = createPlan.CreatePlan(ctx, product, productFamilies[0], usageType)
		if err != nil {
			t.Fatalf("failed to create compute plan: %v", err)
		}
	}

	usageServiceClient := pb.NewUsageServiceClient(usageConn)
	usageController := NewUsageController(clientTestCommon.GetAriaCredentials(), cloudAccountClient, usageServiceClient,
		clientTestCommon.GetUsageClient(), clientTestCommon.GetAriaAccountClient())

	usageController.ReportAllUsage(ctx)
}

func TestReportUsageToAria(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestUsage")
	logger.Info("BEGIN")
	defer logger.Info("End")
	initializeTests(t, ctx)
	BeforeUsageControllerTest(ctx, "TestReportUsageToAria", t)

}
