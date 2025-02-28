// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	//aria "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestPromo(t *testing.T) {
	promoClient := common.GetPromoClient()
	ctx := context.Background()
	if err := InitAriaForTesting(ctx); err != nil {
		t.Fatal(err)
	}

	ariaPlan := common.GetAriaPlanClient()
	productFamily := GetProductFamily()
	usageType := GetUsageType(t)

	type prodElem struct {
		found bool
		prod  *pb.Product
	}

	planMap := map[string]*prodElem{}
	planNos := []int{}
	planIds := []string{}
	for ii := 0; ii < 5; ii++ {
		prod := GetProduct()
		prod.Id = uuid.NewString()
		prod.Name = fmt.Sprintf("promo test %v", ii)
		resp, err := ariaPlan.CreatePlan(ctx, prod, productFamily, usageType)
		if err != nil {
			t.Error(err)
			break
		}
		prodClientId := client.GetPlanClientId(prod.Id)
		planMap[prodClientId] = &prodElem{found: false, prod: prod}
		planNos = append(planNos, resp.PlanNo)

		planIds = append(planIds, prodClientId)
	}
	err := promoClient.AddPlansToPromo(ctx, planIds)
	if err != nil {
		t.Error(err)
	}

	plans, err := ariaPlan.GetAllClientPlansForPromoCode(ctx, client.GetPromoCode())
	if err != nil {
		t.Error(err)
	} else {
		for _, plan := range plans.AllClientPlanDtls {
			if elem, ok := planMap[plan.ClientPlanId]; ok {
				elem.found = true
			}
		}
	}

	for _, elem := range planMap {
		if !elem.found {
			t.Errorf("Missing plan for product %v", elem.prod.Id)
		}
	}

	if len(planNos) > 0 {
		_, err = ariaPlan.DeletePlans(ctx, planNos)
		if err != nil {
			t.Error(err)
		}
	}
}
