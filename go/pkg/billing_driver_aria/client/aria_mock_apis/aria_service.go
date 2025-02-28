// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"fmt"
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
)

var serviceNosMap = map[int64]response.GetServiceDetails{}
var serviceNos = make([]int, 0)
var clientServiceId = make([]string, 0)

// Create IDC Service
func CreateAriaServiceHandler(w http.ResponseWriter, req map[string]interface{}) {

	var reqStruct request.CreateService
	reqStruct.ServiceName = strcnv(req["service_name"])
	reqStruct.ServiceType = strcnv(req["service_type"])
	reqStruct.GlCd = strcnv(req["gl_cd"])
	reqStruct.ClientServiceId = strcnv(req["client_service_id"])
	if lookupService(reqStruct.ServiceName, serviceName) ||
		!lookupService(reqStruct.ServiceType, serviceType) ||
		reqStruct.GlCd != "1" ||
		lookupService(reqStruct.ClientServiceId, clientServiceId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "Attributes service_name/service_type already exists!",
		})
		return
	}
	serviceNum := int(randomDigitGen(8))
	resp := response.CreateService{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		ServiceNo: serviceNum,
	}
	serviceNos = append(serviceNos, serviceNum)
	clientServiceId = append(clientServiceId, reqStruct.ClientServiceId)
	if strictValidation {
		serviceNosMap[int64(serviceNum)] = response.GetServiceDetails{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			ServiceNo:       strcnv(serviceNos[len(serviceNos)-1]),
			ClientServiceId: strcnv(serviceNos[len(clientServiceId)-1]),
			ServiceName:     StringGen(8),
			ServiceType:     StringGen(8),
			GlCd:            reqStruct.GlCd,
		}
	} else {
		serviceNosMap[int64(serviceNum)] = response.GetServiceDetails{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
		}
	}
	MockResponseWriter(w, resp)
}

// Get IDC Service
// This function has only client_id and auth_key as required parameters
func GetAriaServiceDetailsHandler(w http.ResponseWriter, req map[string]interface{}) {
	for _, val := range serviceNosMap {
		resp := val
		MockResponseWriter(w, resp)
	}
}

func UpdateAriaServicehandler(w http.ResponseWriter, req map[string]interface{}) {

	var reqStruct request.CreateService
	reqStruct.ServiceName = fmt.Sprint(req["service_name"])
	reqStruct.ServiceType = fmt.Sprint(req["service_type"])
	reqStruct.GlCd = fmt.Sprint(req["gl_cd"])
	if !lookupService(reqStruct.ServiceName, serviceName) ||
		!lookupService(reqStruct.ServiceType, serviceType) ||
		reqStruct.GlCd != "1" {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "Attributes service_name/service_type already exists!",
		})
		return
	}
	serviceNum := int(randomDigitGen(8))
	resp := response.CreateService{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		ServiceNo: serviceNum,
	}
	MockResponseWriter(w, resp)
}

func CreateBulkUsageRecord(w http.ResponseWriter, body []byte) {
	var reqStruct = requestAriaMock[*request.BulkRecordUsageMRequest](w, body)
	for _, val := range reqStruct.UsageRecs {
		if !lookupService(val.ClientAcctId, clientAcctId) {
			MockResponseWriter(w, errorResp{
				ErrorCode: 1009,
				ErrorMsg:  "Account Doesn't exist",
			})
		}
	}
	resp := response.BulkRecordUsageMResponse{
		AriaResponse: response.AriaResponse{ErrorCode: 0, ErrorMsg: "OK"},
		ErrorCode:    0,
		ErrorMsg:     "OK",
	}
	MockResponseWriter(w, resp)
}
