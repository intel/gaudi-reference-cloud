// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"testing"

	"github.com/google/uuid"
	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"gotest.tools/assert"
)

var (
	idcTestSyncProducts []*pb.Product
	idcTestSyncVendors  []*pb.Vendor
)

const (
	computeProductVMSmallXeon4Name      = "compute-xeon4-small-vm"
	computeProductVMSmallXeon4MatchExpr = "billUsage && service == compute && instanceType == vm.xeon4.small"
)

func BeforTestSetupProductsAndVendors(ctx context.Context, testName string, t *testing.T) {
	idcComputeProductFamilies := make([]*pb.ProductFamily, 0)
	idcTestSyncVendors = make([]*pb.Vendor, 0)
	idcTestSyncProducts = make([]*pb.Product, 0)
	vendorId := uuid.NewString()
	idcComputeProductFamilyId := uuid.NewString()
	idcComputeProductFamily := billingCommon.GetIdcComputeProductFamily(idcComputeProductFamilyId)
	idcComputeProductFamilies = append(idcComputeProductFamilies, idcComputeProductFamily)
	vendor := billingCommon.GetIdcVendor(vendorId, idcComputeProductFamilies)
	productId := uuid.NewString()
	idcProduct := billingCommon.GetProduct(computeProductVMSmallXeon4Name, productId, vendorId, idcComputeProductFamilyId, uuid.NewString(), computeProductVMSmallXeon4MatchExpr)
	idcTestSyncVendors = append(idcTestSyncVendors, vendor)
	idcTestSyncProducts = append(idcTestSyncProducts, idcProduct)
}

func TestSyncProducts(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("SyncProductsTest")
	logger.Info("BEGIN")
	defer logger.Info("End")
	BeforTestSetupProductsAndVendors(ctx, "SyncProductsTest", t)
	testProductClient := billingCommon.NewProductTestClient(idcTestSyncVendors, idcTestSyncProducts)
	ariaController := NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	productController := NewProductController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials(), testProductClient, ariaController)
	err := productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}
	ariaPlan := common.GetAriaPlanClient()
	clientPlanId := common.GetTestClientId(idcTestSyncProducts[0].Id)
	getPlanDetailResp, err := ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}
	assert.Equal(t, clientPlanId, getPlanDetailResp.ClientPlanId)
	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{getPlanDetailResp.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
	//Create an invalid plan and sync
	invalidClientPlanId := uuid.NewString()
	resp, err := ariaPlan.CreateTestPlan(ctx, "Test Service", "Test Schedule", "Test Plan Name", invalidClientPlanId)
	if err != nil {
		t.Fatalf("error in creating test plan %v", err)
	}
	controller := NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	promoClient := common.GetPromoClient()
	if err := controller.InitAria(ctx); err != nil {
		t.Fatal(err)
	}
	err = promoClient.AddPlansToPromo(ctx, []string{invalidClientPlanId})
	if err != nil {
		t.Fatalf("error in adding plnas to promo %v", err)
	}
	err = productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}
	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{resp.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}

func TestSyncProductsEnterprise(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("SyncProductsTest")
	logger.Info("BEGIN")
	defer logger.Info("End")
	BeforTestSetupProductsAndVendors(ctx, "SyncProductsTest", t)
	testProductClient := billingCommon.NewProductTestClient(idcTestSyncVendors, idcTestSyncProducts)
	ariaController := NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	productController := NewProductController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials(), testProductClient, ariaController)

	err := productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	ariaPlan := common.GetAriaPlanClient()
	clientPlanId := common.GetTestClientId(idcTestSyncProducts[0].Id)
	getPlanDetailResp, err := ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}
	assert.Equal(t, clientPlanId, getPlanDetailResp.ClientPlanId)

	// Updating Enterprise Rates
	idcTestSyncProducts[0].Rates[2].Rate = "300"
	idcTestSyncProducts[0].Metadata["displayName"] = "test the sync flow"

	// Rates getting updated in the Aria Plans
	err = productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{getPlanDetailResp.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}

}

func TestHasDiffPlanDetailSyncProducts(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("HasDiffPlanDetailSyncProductsTest")
	logger.Info("BEGIN")
	defer logger.Info("End")
	BeforTestSetupProductsAndVendors(ctx, "HasDiffPlanDetailSyncProductsTest", t)
	testProductClient := billingCommon.NewProductTestClient(idcTestSyncVendors, idcTestSyncProducts)
	ariaController := NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	productController := NewProductController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials(), testProductClient, ariaController)
	ariaPlan := common.GetAriaPlanClient()

	err := productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	// Change plan name in the plan detail
	idcTestSyncProducts[0].Name = "New Test plan"
	err = productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	clientPlanId := common.GetTestClientId(idcTestSyncProducts[0].Id)
	getPlanDetailResp, err := ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}

	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{getPlanDetailResp.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}

	// Change plan ID in the plan detail, this will result in creation of new plan
	idcTestSyncProducts[0].Id = uuid.NewString()
	err = productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	// Change plan description in the plan detail
	idcTestSyncProducts[0].Description = "New test plan description"
	err = productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	clientPlanId = common.GetTestClientId(idcTestSyncProducts[0].Id)
	getPlanDetailResp, err = ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}

	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{getPlanDetailResp.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}

func TestHasDiffPlanServiceRateSyncProducts(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("HasDiffPlanServiceRateSyncProductsTest")
	logger.Info("BEGIN")
	defer logger.Info("End")
	BeforTestSetupProductsAndVendors(ctx, "HasDiffPlanServiceRateSyncProductsTest", t)
	testProductClient := billingCommon.NewProductTestClient(idcTestSyncVendors, idcTestSyncProducts)
	ariaController := NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	productController := NewProductController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials(), testProductClient, ariaController)

	err := productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	ariaPlan := common.GetAriaPlanClient()
	clientPlanId := common.GetTestClientId(idcTestSyncProducts[0].Id)
	getPlanDetailResp, err := ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}
	assert.Equal(t, clientPlanId, getPlanDetailResp.ClientPlanId)

	// Updating Plan Service Rate
	idcTestSyncProducts[0].Rates[1].Rate = "11"

	// Rates getting updated in the Aria Plans
	err = productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{getPlanDetailResp.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}

}

func TestHasDiffPlanServiceSyncProducts(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("HasDiffPlanServiceSyncProductsTest")
	logger.Info("BEGIN")
	defer logger.Info("End")
	BeforTestSetupProductsAndVendors(ctx, "HasDiffPlanServiceSyncProductsTest", t)
	testProductClient := billingCommon.NewProductTestClient(idcTestSyncVendors, idcTestSyncProducts)
	ariaController := NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	productController := NewProductController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials(), testProductClient, ariaController)

	err := productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	ariaPlan := common.GetAriaPlanClient()
	clientPlanId := common.GetTestClientId(idcTestSyncProducts[0].Id)
	getPlanDetailResp, err := ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}
	assert.Equal(t, clientPlanId, getPlanDetailResp.ClientPlanId)

	// Updating Plan Service Name
	idcTestSyncProducts[0].Metadata["displayName"] = "Test Service Name"

	// Rates getting updated in the Aria Plans
	err = productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{getPlanDetailResp.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}

}

func TestHasDiffPlanSupplFieldSyncProducts(t *testing.T) {

	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("HasDiffPlanSupplFieldSyncProductsTest")
	logger.Info("BEGIN")
	defer logger.Info("End")
	BeforTestSetupProductsAndVendors(ctx, "HasDiffPlanSupplFieldSyncProductsTest", t)
	testProductClient := billingCommon.NewProductTestClient(idcTestSyncVendors, idcTestSyncProducts)
	ariaController := NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	productController := NewProductController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials(), testProductClient, ariaController)

	err := productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	ariaPlan := common.GetAriaPlanClient()
	clientPlanId := common.GetTestClientId(idcTestSyncProducts[0].Id)
	getPlanDetailResp, err := ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}
	assert.Equal(t, clientPlanId, getPlanDetailResp.ClientPlanId)

	// Updating SupplementalObjectField
	idcTestSyncProducts[0].Pcq = "Test PCQ"

	err = productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync products %v", err)
	}

	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{getPlanDetailResp.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}

}

func getStorageProductsAndVendors(ctx context.Context, testName string, t *testing.T) ([]*pb.Vendor, []*pb.Product) {
	idcFileStorageProductFamilies := make([]*pb.ProductFamily, 0)
	idcVendors := make([]*pb.Vendor, 0)
	idcFileStorageProducts := make([]*pb.Product, 0)
	vendorId := uuid.NewString()
	idcFileStorageProductFamilyId := uuid.NewString()
	idcFileStorageProductFamily := billingCommon.GetIdcStorageProductFamily(idcFileStorageProductFamilyId)
	idcFileStorageProductFamilies = append(idcFileStorageProductFamilies, idcFileStorageProductFamily)
	vendor := billingCommon.GetIdcVendor(vendorId, idcFileStorageProductFamilies)
	productId := uuid.NewString()
	idcFileStorageProduct := billingCommon.GetIdcFileStorageProduct(billingCommon.IdcFileStorageProductName, productId, vendorId, idcFileStorageProductFamilyId, uuid.NewString(), billingCommon.IdcFileStroageProductMatchExpr)
	idcVendors = append(idcVendors, vendor)
	idcFileStorageProducts = append(idcFileStorageProducts, idcFileStorageProduct)
	return idcVendors, idcFileStorageProducts
}

func TestSyncStorageProducts(t *testing.T) {
	if config.Cfg.AriaSystem.AuthKey == "" {
		t.Log("aria system auth key is not set")
		return
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestSyncStorageProducts")
	logger.Info("BEGIN")
	defer logger.Info("End")
	idcVendors, idcFileStorageProducts := getStorageProductsAndVendors(ctx, "TestSyncStorageProducts", t)
	testProductClient := billingCommon.NewProductTestClient(idcVendors, idcFileStorageProducts)
	ariaController := NewAriaController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials())
	productController := NewProductController(common.GetAriaClient(), common.GetAriaAdminClient(), common.GetAriaCredentials(), testProductClient, ariaController)
	err := productController.SyncProducts(ctx)
	if err != nil {
		t.Fatalf("error in sync storage products %v", err)
	}
	ariaPlan := common.GetAriaPlanClient()
	clientPlanId := common.GetTestClientId(idcFileStorageProducts[0].Id)
	getPlanDetailResp, err := ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}
	assert.Equal(t, clientPlanId, getPlanDetailResp.ClientPlanId)
	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{getPlanDetailResp.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}
