// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
)

func CreateServiceCreditsHandler(w http.ResponseWriter, body []byte) {
	var reqStruct = requestAriaMock[*request.CreateAdvancedServiceCreditMRequest](w, body)

	if !lookupService(reqStruct.ClientAcctId, clientAcctId) {
		http.Error(w, "Error checking attributes:- invalid type", http.StatusInternalServerError)
		return
	}
	resp := response.CreateAdvancedServiceCreditMResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
	}
	MockResponseWriter(w, resp)
}

func GetUnappliedServiceCreditsHandler(w http.ResponseWriter, body []byte) {
	var reqStruct = requestAriaMock[*request.GetUnappliedServiceCreditsMRequest](w, body)
	if !lookupService(reqStruct.AcctNo, acctNos) || !lookupService(reqStruct.AltCallerId, altCallerId) {
		http.Error(w, "Error checking attributes:- invalid type", http.StatusInternalServerError)
		return
	}
	resp := response.GetUnappliedServiceCreditsMResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		UnappliedServiceCreditsDetails: GetDefaultUnappliedServiceCreditsDetail(),
	}
	MockResponseWriter(w, resp)
}
