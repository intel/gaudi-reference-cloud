// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"

	//aria "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/stretchr/testify/assert"
)

func GetUsageType(t *testing.T) *data.UsageType {
	usageTypeClient := common.GetAriaUsageTypeClient()
	usageType, err := usageTypeClient.GetMinutesUsageType(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return usageType
}

func TestCreatePlan(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreatePlan")
	logger.Info("testing create plan")

	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)
	createRespBody, err := ariaPlan.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(ctx, GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("failed to get client plan: %v", err)
	}

	hasDiffPlan := client.HasDiffPlanDetail(ctx, resp.AllClientPlanDtls[0], product)
	assert.Equal(t, false, hasDiffPlan)

	hasDiffPlanService := client.HasDiffPlanService(ctx, resp.AllClientPlanDtls[0].PlanServices[0], usageType, product)
	assert.Equal(t, false, hasDiffPlanService)

	hasDiffPlanServiceRate := client.HasDiffPlanServiceRate(ctx, resp.AllClientPlanDtls[0].PlanServices[0].PlanServiceRates, product)
	assert.Equal(t, false, hasDiffPlanServiceRate)

	hasDiffPlanSupplField := client.HasDiffPlanSupplField(ctx, resp.AllClientPlanDtls[0].PlanSuppFields, product)
	assert.Equal(t, false, hasDiffPlanSupplField)

	_, err = ariaPlan.DeletePlans(context.Background(), []int{createRespBody.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}

func TestEditPlan(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestEditPlan")
	logger.Info("testing edit plan")

	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)
	createRespBody, err := ariaPlan.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}
	product.Name = "EditProductName"

	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(ctx, GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("failed to get client plan: %v", err)
	}
	hasDiffPlan := client.HasDiffPlanDetail(ctx, resp.AllClientPlanDtls[0], product)
	assert.Equal(t, true, hasDiffPlan)

	product.Metadata["displayName"] = "New Testing ServiceName"
	hasDiffPlanService := client.HasDiffPlanService(ctx, resp.AllClientPlanDtls[0].PlanServices[0], usageType, product)
	assert.Equal(t, true, hasDiffPlanService)

	for _, rate := range product.Rates {
		rate.Rate = ".6"
	}
	hasDiffPlanServiceRate := client.HasDiffPlanServiceRate(ctx, resp.AllClientPlanDtls[0].PlanServices[0].PlanServiceRates, product)
	assert.Equal(t, true, hasDiffPlanServiceRate)

	_, err = ariaPlan.EditPlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("failed to edit plan: %v", err)
	}
	serviceClient := common.GetAriaServiceClient()
	_, err = serviceClient.UpdateAriaService(ctx, client.GetPlanClientId(product.Id), product.Metadata["displayName"], usageType.UsageTypeNo)
	if err != nil {
		t.Fatalf("failed to update service plan: %v", err)
	}
	planDetails, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(ctx, GetTestClientPlanId(product.GetId()))
	if err != nil {
		t.Fatalf("failed to get client plan: %v", err)
	}
	hasDiffPlan = client.HasDiffPlanDetail(ctx, planDetails.AllClientPlanDtls[0], product)
	assert.Equal(t, false, hasDiffPlan)

	hasDiffPlanService = client.HasDiffPlanService(ctx, planDetails.AllClientPlanDtls[0].PlanServices[0], usageType, product)
	assert.Equal(t, false, hasDiffPlanService)

	hasDiffPlanServiceRate = client.HasDiffPlanServiceRate(ctx, planDetails.AllClientPlanDtls[0].PlanServices[0].PlanServiceRates, product)
	assert.Equal(t, false, hasDiffPlanServiceRate)

	_, err = ariaPlan.DeletePlans(ctx, []int{createRespBody.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}

// Negative test case for creating the plan with same clientPlanId
func TestCreatePlanTwice(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestCreatePlanTwice")
	logger.Info("testing create plan service, trying to connect with existing client plan id")

	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	createRespBody, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, GetUsageType(t))
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	// Calling Create Plan twice
	_, err = ariaPlan.CreatePlan(context.Background(), product, productFamily, GetUsageType(t))
	if err == nil {
		t.Fatalf("create plan was successfull with existing client_plan_id which should have failed: %v", err)
	}

	_, err = ariaPlan.DeletePlans(context.Background(), []int{createRespBody.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}

}

func TestDeactivatePlan(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestDeactivatePlan")
	logger.Info("testing deactivate plan")

	ariaPlan := common.GetAriaPlanClient()

	product := GetProduct()
	productFamily := GetProductFamily()

	// Create Plan
	createRespBody, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, GetUsageType(t))
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	// Deactivate Plan
	_, err = ariaPlan.DeactivatePlan(context.Background(), product, productFamily, GetUsageType(t))
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{createRespBody.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}

func TestDeactivatePlanFromClientPlanDetail(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestDeactivatePlanFromClientPlanDetail")
	logger.Info("testing deactivate plan from client plan detail")

	ariaPlan := common.GetAriaPlanClient()

	product := GetProduct()
	productFamily := GetProductFamily()

	usageType := GetUsageType(t)
	// Create Plan
	createRespBody, err := ariaPlan.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("failed to Create plan: %v", err)
	}

	resp, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(ctx, common.GetTestClientId(product.Id))
	if err != nil {
		t.Fatalf("failed to get all client plans: %v", err)
	}

	// Deactivate Plan
	_, err = ariaPlan.DeactivatePlanFromClientPlanDetail(ctx, &resp.AllClientPlanDtls[0], usageType)
	if err != nil {
		t.Fatalf("failed to deactivate plan: %v", err)
	}

	getPlanDetailResp, err := ariaPlan.GetAriaPlanDetails(ctx, common.GetTestClientId(product.Id))
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}
	assert.Equal(t, 0, getPlanDetailResp.ActiveInd)

	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{createRespBody.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}

func TestDeactivatePlanFromGetClientPlanDetail(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestDeactivatePlanFromClientPlanDetail")
	logger.Info("testing deactivate plan from get client plan detail")

	ariaPlan := common.GetAriaPlanClient()

	product := GetProduct()
	productFamily := GetProductFamily()

	usageType := GetUsageType(t)

	// Create Plan
	createRespBody, err := ariaPlan.CreatePlan(ctx, product, productFamily, usageType)
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}
	clientPlanId := common.GetTestClientId(product.Id)
	resp, err := ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plans: %v", err)
	}

	planServices, err := GetPlanServices(ctx, resp, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan services: %v", err)
	}

	clientPlanDetails := client.MapResponseToClientPlanDetail(resp, planServices)

	// Deactivate Plan
	_, err = ariaPlan.DeactivatePlanFromClientPlanDetail(ctx, clientPlanDetails, usageType)
	if err != nil {
		t.Fatalf("failed to deactivate plan: %v", err)
	}

	getPlanDetailResp, err := ariaPlan.GetAriaPlanDetails(ctx, clientPlanId)
	if err != nil {
		t.Fatalf("failed to get all client plan detail: %v", err)
	}
	assert.Equal(t, 0, getPlanDetailResp.ActiveInd)

	// Delete Plan
	_, err = ariaPlan.DeletePlans(context.Background(), []int{createRespBody.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}

func TestDeletePlan(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestDeactivatePlan")
	logger.Info("testing delete plan")
	ariaPlan := common.GetAriaPlanClient()

	product := GetProduct()
	productFamily := GetProductFamily()
	// Creation of the plan for deleting the same plan.
	createRespBody, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, GetUsageType(t))
	if err != nil {
		t.Fatalf("failed to Create plan: %v", err)
	}

	// Deletion of the plan created above
	_, err = ariaPlan.DeletePlans(context.Background(), []int{createRespBody.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}

	// TODO: Call DeactivatePlan API when DeletePlan fails for those plans associated with accounts.
}
