// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
)

// Create Plan Client
func CreatePlanClientHandler(w http.ResponseWriter, body map[string]interface{}) {

	//check validity of attributes - test conditions dependent on preset values
	var reqStruct request.PlanRequest
	reqStruct.PlanName = strcnv(body["plan_name"])
	reqStruct.PlanType = strcnv(body["plan_type"])
	reqStruct.Currency = strcnv(body["currency"])
	reqStruct.Active = int(intcnv(strcnv(body["active"])))
	reqStruct.ClientPlanId = strcnv(body["client_plan_id"])
	if lookupService(reqStruct.PlanName, planName) ||
		lookupService(reqStruct.ClientPlanId, clientPlanId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "PlanName/ClientPlanId Already Exists",
		})
		return
	}
	if !lookupService(reqStruct.PlanType, planType) ||
		!lookupService(reqStruct.Currency, Currency) ||
		reqStruct.Active == 0 {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "PlanType/Currency value is not correct:",
		})
	}
	clientPlanId = append(clientPlanId, reqStruct.ClientPlanId)
	planName = append(planName, reqStruct.PlanName)
	planNo := randomDigitGen(8)
	resp := response.PlanResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		PlanNo: int(planNo),
	}
	planNos = append(planNos, int(planNo))
	MockResponseWriter(w, resp)
}
