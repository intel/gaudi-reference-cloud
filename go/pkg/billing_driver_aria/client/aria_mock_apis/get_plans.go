// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"fmt"
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/pborman/uuid"
)

// Get Plans Handler
// This function has only client_id and auth_key as required parameters
const (
	ACTIVE_ONE  int = 1
	ACTIVE_ZERO int = 0
)

func GetPlansHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetAllClientPlans](w, body)

	var resp response.GetClientPlansAllMResponse
	//Please comment the below code if you are testing get first
	if !lookupService(reqStruct.ClientPlanId, clientPlanId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1004,
			ErrorMsg:  "clientPlanId doesn't Exist",
		})
		return
	}
	if strictValidation {
		resp = response.GetClientPlansAllMResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			AllClientPlanDtls: GetDefaultAllClientPlanDtl(reqStruct.PromoCode),
		}
		if reqStruct.PromoCode != "" {
			resp.AllClientPlanDtls[0].PromotionalPlanSets = GetDefaultPromotionalPlanSet(reqStruct.PromoCode)
		}
	} else {
		resp = response.GetClientPlansAllMResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			AllClientPlanDtls: []data.AllClientPlanDtl{
				{
					ClientPlanId: reqStruct.ClientPlanId,
					PlanName:     planName[len(planName)-1],
					CurrencyCd:   Currency[0],
				},
			},
		}
	}

	MockResponseWriter(w, resp)
}

func GetPlanDetailsHandler(w http.ResponseWriter, body map[string]interface{}) {

	var resp response.GetPlanDetailResponse
	if strictValidation {
		resp = response.GetPlanDetailResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			PlanNo:        planNos[len(planNos)-1],
			PlanName:      planName[len(planName)-1],
			PlanDesc:      StringGen(15),
			PlanType:      planType[0],
			PlanGroups:    []data.PlanGroup{},
			PlanGroupsIds: []data.PlanGroupId{},
			CurrencyCd:    Currency[0],
			ClientPlanId:  clientPlanId[len(clientPlanId)-1],
			ActiveInd:     ACTIVE_ONE,
		}
	} else {
		resp = response.GetPlanDetailResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			PlanType:     planType[5],
			PlanName:     planName[len(planName)-1],
			ClientPlanId: clientPlanId[len(clientPlanId)-1],
			PlanNo:       planNos[len(planNos)-1],
		}

	}
	resp.Services = make([]data.Service, 1)
	clientServiceId = append(clientServiceId, config.Cfg.ClientIdPrefix+".Test Service Plan "+uuid.New()[:12])
	resp.Services[0].ClientServiceId = clientServiceId[len(clientServiceId)-1]
	fmt.Println(clientServiceId)

	MockResponseWriter(w, resp)
}

func GetClientPlanServiceRatesHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetClientPlanServiceRates](w, body)
	if lookupService(reqStruct.ClientPlanId, clientPlanId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1004,
			ErrorMsg:  "clientPlanId doesn't Exist",
		})
		return
	}
	resp := response.GetClientPlanServiceRates{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		PlanServiceRates: GetDefaultPlanServiceRate(),
	}
	MockResponseWriter(w, resp)
}
