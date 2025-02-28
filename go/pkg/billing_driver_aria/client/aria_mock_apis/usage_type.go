// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

var usageTypes = map[string]data.UsageType{}
var usageTypeCodes = []string{}

// Create Usage Type Handler
func CreateUsageTypehandler(w http.ResponseWriter, body map[string]interface{}) {
	usageTypeCode := strcnv(body["usage_type_code"])
	usageTypeNo := int(randomDigitGen(11))
	CreateOrUpdateUsageTypeHandler(usageTypeCode, usageTypeNo)
	resp := response.CreateUsageType{
		AriaResponse: response.AriaResponse{ErrorCode: 0, ErrorMsg: "OK"},
		UsageTypeNo:  int(usageTypeNo),
	}

	MockResponseWriter(w, resp)
}

// Update Usage Type Handler
func UpdateUsageTypeHandler(w http.ResponseWriter, body map[string]interface{}) {
	key := strcnv(body["usage_type_code"])
	usageTypes[key] = data.UsageType{
		UsageTypeNo:   int(intcnv(strcnv(body["usage_type_no"]))),
		UsageTypeDesc: strcnv(body["usage_type_desc"]),
		UsageTypeName: strcnv(body["usage_type_name"]),
		UsageUnitType: strcnv(body["usage_unit_type"]),
	}
	resp := response.CreateUsageType{
		AriaResponse: response.AriaResponse{ErrorCode: 0, ErrorMsg: "OK"},
		UsageTypeNo:  int(usageTypes[strcnv(body["usage_type_code"])].UsageTypeNo),
	}
	MockResponseWriter(w, resp)
}
