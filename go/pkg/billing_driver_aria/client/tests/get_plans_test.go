// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestGetPlansByClientPlanId(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetPlansByClientPlanId")
	logger.Info("testing get plans by client plan id")

	ariaPlan := common.GetAriaPlanClient()
	_, err := ariaPlan.GetAriaPlanDetailsAllForClientPlanId(context.Background(), GetClientPlanId())
	if err != nil {
		t.Fatalf("failed to get all client plans: %v", err)
	}
}

func TestGetClientPlanServiceRates(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetClientPlanServiceRates")
	logger.Info("testing get plan service rates")

	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	createRespBody, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, GetUsageType(t))
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	planServiceRatesResp, err := ariaPlan.GetClientPlanServiceRates(context.Background(), common.GetTestClientId(product.Id), common.GetTestClientId(productFamily.Id))
	if err != nil {
		t.Fatalf("failed to get client plan service rates: %v", err)
	}
	assert.Equal(t, common.GetTestClientId(product.Id+"."+"premium"), planServiceRatesResp.PlanServiceRates[0].ClientRateScheduleId)

	_, err = ariaPlan.DeletePlans(context.Background(), []int{createRespBody.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}

func TestGetPlanDetails(t *testing.T) {
	logger := log.FromContext(context.Background()).WithName("TestGetPlanDetails")
	logger.Info("testing aria service")

	ariaPlan := common.GetAriaPlanClient()
	product := GetProduct()
	productFamily := GetProductFamily()
	createRespBody, err := ariaPlan.CreatePlan(context.Background(), product, productFamily, GetUsageType(t))
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	planDetailResp, err := ariaPlan.GetAriaPlanDetails(context.Background(), common.GetTestClientId(product.Id))
	if err != nil {
		t.Fatalf("failed to get  client plan details: %v", err)
	}
	assert.Equal(t, product.Name, planDetailResp.PlanName)
	assert.Equal(t, "Recurring", planDetailResp.PlanType)
	assert.Equal(t, common.GetTestClientId(product.Id), planDetailResp.ClientPlanId)
	assert.Equal(t, common.GetTestClientId(productFamily.Id), planDetailResp.Services[0].ClientServiceId)
	_, err = ariaPlan.DeletePlans(context.Background(), []int{createRespBody.PlanNo})
	if err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
}
