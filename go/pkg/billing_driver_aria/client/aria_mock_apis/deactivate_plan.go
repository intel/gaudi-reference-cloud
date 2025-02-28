// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"fmt"
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
)

// Deactivate Plan Handler
var deactivatePlanArr = map[int]int{}

func DeactivatePlanHandler(w http.ResponseWriter, body map[string]interface{}) {
	var reqStruct request.PlanRequest
	reqStruct.PlanName = strcnv(body["plan_name"])
	reqStruct.PlanType = strcnv(body["plan_type"])
	reqStruct.Currency = strcnv(body["currency"])
	reqStruct.Active = int(intcnv(strcnv(body["active"])))
	reqStruct.ClientPlanId = strcnv(body["client_plan_id"])
	reqStruct.EditDirectives = int(intcnv(strcnv(body["edit_directives"])))
	fmt.Println(planName, " ", reqStruct.PlanName)
	if !lookupService(reqStruct.PlanName, planName) ||
		!lookupService(reqStruct.PlanType, planType) ||
		!lookupService(reqStruct.Currency, Currency) ||
		reqStruct.Active == 1 ||
		reqStruct.EditDirectives != 2 ||
		!lookupService(reqStruct.ClientPlanId, clientPlanId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "Mandatory Attributes are not correct: PlanName/ClientPlanID/Active/Edit Directives",
		})
		return
	}
	resp := response.PlanResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		PlanNo: planNos[len(planNos)-1],
	}
	deactivatePlanArr[planNos[len(planNos)-1]] = 1
	MockResponseWriter(w, resp)
}
