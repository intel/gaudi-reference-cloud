// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"fmt"
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

// Get Usage Type
// This function has only client_id and auth_key as required parameters
func GetUsageTypeHandler(w http.ResponseWriter, body map[string]interface{}) {

	var resp response.GetUsageTypes
	if strictValidation {
		resp = response.GetUsageTypes{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			UsageTypes: []data.UsageType{
				GetDefaultUsageTypes(0),
				GetDefaultUsageTypes(1),
				GetDefaultUsageTypes(2),
			},
		}
		//If data is present in map , will be pushed in th response :- Example for map usage
		for _, val := range usageTypes {
			resp.UsageTypes = append(resp.UsageTypes, val)
		}
	} else {
		resp = response.GetUsageTypes{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			UsageTypes: []data.UsageType{
				GetDefaultUsageTypes(2)}}
	}

	MockResponseWriter(w, resp)

}

func GetUsageTypeDetailsHandler(w http.ResponseWriter, body map[string]interface{}) {

	var reqStruct request.GetUsageTypeDetails
	reqStruct.UsageTypeCode = strcnv(body["usage_type_code"])
	usageTypeNo := int(randomDigitGen(11))
	fmt.Println(reqStruct.UsageTypeCode)
	if !lookupService(reqStruct.UsageTypeCode, usageTypeCodes) {
		CreateOrUpdateUsageTypeHandler(reqStruct.UsageTypeCode, usageTypeNo)
		// MockResponseWriter(w, error_resp{
		// 	ErrorCode: 1004,
		// 	ErrorMsg:  "Usage Type doesn't Exist",
		// })
		// return
	}

	resp := response.GetUsageTypeDetails{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		UsageTypeNo:            usageTypes[reqStruct.UsageTypeCode].UsageTypeNo,
		UsageTypeName:          usageTypes[reqStruct.UsageTypeCode].UsageTypeName,
		UsageTypeDesc:          usageTypes[reqStruct.UsageTypeCode].UsageTypeDesc,
		UsageTypeDisplayString: "Invoice line items total value usage",
		UsageUnitType:          "Unit",
		UsageTypeCode:          reqStruct.UsageTypeCode,
		IsEditable:             false,
	}

	MockResponseWriter(w, resp)
}
