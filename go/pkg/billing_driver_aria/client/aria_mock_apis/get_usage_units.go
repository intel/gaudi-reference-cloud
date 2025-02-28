// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

// Get Usage Unit types
// This function has only client_id and auth_key as required parameters
func GetUsageUnitsHandler(w http.ResponseWriter, body map[string]interface{}) {

	var resp response.GetUsageUnitTypes
	if strictValidation {
		resp = response.GetUsageUnitTypes{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			UsageUnitTypes: []data.UsageUnitType{
				GetDefaultUsageUnitTypes(0),
				GetDefaultUsageUnitTypes(1),
				GetDefaultUsageUnitTypes(2),
				GetDefaultUsageUnitTypes(3),
			},
		}
	} else {
		resp = response.GetUsageUnitTypes{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			UsageUnitTypes: []data.UsageUnitType{
				GetDefaultUsageUnitTypes(4),
			},
		}
	}
	MockResponseWriter(w, resp)
}
