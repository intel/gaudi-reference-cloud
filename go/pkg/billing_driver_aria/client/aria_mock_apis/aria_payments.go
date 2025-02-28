// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"net/http"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/pborman/uuid"
)

const (
	PROC_STATUS_CODE string = "100"
	PROC_STATUS_TEXT string = "pass"
)

// Get Payment Method
func GetPaymentMethodHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetPaymentMethodsRequest](w, body)

	if !lookupService(reqStruct.ClientAccountId, clientAcctId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "account does not exist - error fetching AcctNo",
		})
	}
	var resp response.GetPaymentMethodsMResponse
	if strictValidation {
		resp = response.GetPaymentMethodsMResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			AccountPaymentMethods: GetDefaultAccountPaymentMethods(),
		}
	} else {
		resp = response.GetPaymentMethodsMResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
		}
	}

	MockResponseWriter(w, resp)
}

// assign collections
var assignIds = []string{}
var assignGroupIds = []string{}

func AssignCollectionsAccountGroupHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.AssignCollectionsAccountGroupRequest](w, body)
	if lookupService(reqStruct.ClientAccountId, assignIds) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 12004,
			ErrorMsg:  "Account already assigned to this group",
		})
	}
	if lookupService(reqStruct.ClientAccountId, clientAcctId) {
		assignIds = append(assignIds, reqStruct.ClientAccountId)
		assignGroupIds = append(assignGroupIds, reqStruct.ClientAcctGroupId)
	}

	resp := response.AssignCollectionsAccountGroupMResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
	}
	MockResponseWriter(w, resp)
}

var billingGroupNos = []int64{
	971875,
	953361,
}

// Add account payment method
func AddAccountPaymentMethodHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.UpdateAccountBillingGroupRequest](w, body)
	currTime := time.Now()
	if !lookupService(reqStruct.ClientBillingGroupId, clientBillingGroupIds) ||
		reqStruct.CCExpireYear < currTime.Year() ||
		(reqStruct.CCExpireMonth < 1 || reqStruct.CCExpireMonth > 12) ||
		(reqStruct.CCV < 100 || reqStruct.CCV > 1000) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1004,
			ErrorMsg:  "Credentials not valid or Billing Group No doesn't exist",
		})
		return
	}
	stmtContactNo := randomDigitGen(8)
	billingGroupNo := randomDigitGen(6)
	procAuthCode := strcnv(randomDigitGen(10))
	var resp response.UpdateAccountBillingGroupMResponse
	if strictValidation {
		resp = response.UpdateAccountBillingGroupMResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			ProcStatusCode:       PROC_STATUS_CODE,
			ProcStatusText:       PROC_STATUS_TEXT,
			ProcpaymentId:        procAuthCode + "." + strcnv(randomDigitGen(4)),
			ProcAuthCode:         procAuthCode,
			ProcMerchantComments: uuid.New(),
			BillingGroupNo:       strcnv(billingGroupNo),
			BillingGroupNo2:      billingGroupNo,
			StmtContactNo:        stmtContactNo,         // this should be string
			StmtContactNo2:       strcnv(stmtContactNo), // this should be int
			BillingContactInfo:   GetDefaultBillingContactInfo(),
		}
	} else {
		resp = response.UpdateAccountBillingGroupMResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
		}
	}

	MockResponseWriter(w, resp)
}
