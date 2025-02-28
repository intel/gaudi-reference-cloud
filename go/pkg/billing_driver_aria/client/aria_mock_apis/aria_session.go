// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/pborman/uuid"
)

var sessionIds = make([]string, 0)

func SetSessionHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.SetSessionMRequest](w, body)

	if !lookupService(reqStruct.ClientAccountId, clientAcctId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1002,
			ErrorMsg:  "Session already set",
		})
	}
	sessionIds = append(sessionIds, uuid.New()[:30])
	resp := response.SetSessionMResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		SessionId: sessionIds[len(sessionIds)-1],
	}
	MockResponseWriter(w, resp)
}
