// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"fmt"
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/pborman/uuid"
)

var planInstanceNos = []int64{
	2140218,
}
var ariaAccountMap = map[string]response.GetAcctDetailsAllMResponse{}

// Create Aria Account
func CreateAriaAccountHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.CreateAcctCompleteMRequest](w, body)

	for _, val := range reqStruct.Acct {
		//applied only master plan
		for _, value := range val.MasterPlansDetail {
			if value.BillingGroupIdx != GroupIdx["BillingGroupIdx"] ||
				value.DunningGroupIdx != GroupIdx["DunningGroupIdx"] ||
				value.PlanInstanceIdx != GroupIdx["PlanInstanceIdx"] ||
				value.PlanInstanceStatus != GroupIdx["PlanInstanceStatus"] ||
				value.PlanInstanceUnits != float32(GroupIdx["PlanInstanceUnits"]) {
				MockResponseWriter(w, errorResp{
					ErrorCode: 1004,
					ErrorMsg:  "Values not equal or clientPlanId already Exists",
				})
				return
			}
		}
	}

	if lookupService(reqStruct.Acct[0].ClientAcctId, clientAcctId) || lookupService(reqStruct.Acct[0].Userid, userIds) {
		// http.Error(w, "UserID is already in use", http.StatusCreated)
		MockResponseWriter(w, errorResp{
			ErrorCode: 1004,
			ErrorMsg:  "UserID is already in use",
		})
		return
	}

	clientAcctId = append(clientAcctId, reqStruct.Acct[0].ClientAcctId)
	userIds = append(userIds, reqStruct.Acct[0].Userid)
	var resp response.CreateAcctCompleteMResponse
	acctNo := randomDigitGen(8)
	if strictValidation {

		resp = response.CreateAcctCompleteMResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
			OutAcct: []data.OutAcct{
				{
					AcctNo:                  acctNo,
					Userid:                  reqStruct.Acct[0].Userid,
					ClientAcctId:            reqStruct.Acct[0].ClientAcctId,
					AcctLocaleNo:            randomDigitGen(5),
					AcctLocaleName:          "System_US_English_locale",
					StatementContactDetails: GetDefaultStatementContactDetails(),
					BillingErrors:           GetDefaultBillingError(),
					MasterPlansAssigned:     GetDefaultMasterPlanAssigned(),
					InvoiceInfo:             GetDefaultInvoiceInfo(),
				},
			},
			ChiefAcctInfo: GetDefaultChiefAcctInfo(),
		}
	} else {
		resp = response.CreateAcctCompleteMResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
		}
		resp.OutAcct = make([]data.OutAcct, 1)
		resp.OutAcct[0].AcctNo = acctNo
	}
	acctNos = append(acctNos, acctNo)
	for _, val := range reqStruct.Acct {
		ariaAccountMap[val.ClientAcctId] = GetDefaultGetAcctResponse(val.ClientAcctId)
	}
	MockResponseWriter(w, resp)
}

// Get Aria Account
// This function has only client_id ,auth_key and acct_no as required parameters
func GetAriaAccountHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetAcctDetailsAllMRequest](w, body)
	if !lookupService(reqStruct.AcctNo, acctNos) &&
		!lookupService(reqStruct.ClientAcctId, clientAcctId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1004,
			ErrorMsg:  "Error checking attributes:- Acct no or ClientAcctId not present in the records",
		})
		return
	}
	var resp response.GetAcctDetailsAllMResponse
	if strictValidation {
		resp = ariaAccountMap[reqStruct.ClientAcctId]
	} else {
		resp = response.GetAcctDetailsAllMResponse{
			AriaResponse: response.AriaResponse{
				ErrorCode: 0,
				ErrorMsg:  "OK",
			},
		}
	}
	MockResponseWriter(w, resp)
}

func SetAccountNotifyTemplateGroupHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.AccountNotifyTemplateGroup](w, body)
	if !lookupService(reqStruct.AcctNo, acctNos) &&
		!lookupService(reqStruct.ClientAcctId, clientAcctId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "account does not exist - error fetching AcctNo or ClientAcctId",
		})
		return
	}
	resp := response.AriaResponse{
		ErrorCode: 0,
		ErrorMsg:  "OK",
	}
	MockResponseWriter(w, resp)
}

var clientBillingGroupIds = []string{}
var billingGroupMap = map[string]data.BillingGroupsInfo{}

func CreateBillingGroupHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.CreateBillingGroupRequest](w, body)
	if !lookupService(reqStruct.AcctNo, acctNos) && !lookupService(reqStruct.ClientAcctId, clientAcctId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "account does not exist - error fetching AcctNo or ClientAcctId",
		})
		return
	}
	billingGroupNo2 := randomDigitGen(8)
	resp := response.CreateBillingGroupResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		BillingGroupNo2: billingGroupNo2,
	}
	billingGroupMap[reqStruct.ClientBillingGroupId] = data.BillingGroupsInfo{
		BillingGroupNo:          billingGroupNo2,
		BillingGroupName:        uuid.New()[:30],
		BillingGroupDescription: StringGen(50),
		ClientBillingGroupId:    reqStruct.ClientBillingGroupId,
		MasterPlanSummary:       GetDefaultMasterPlanSummary(),
	}
	billingGroupNos = append(billingGroupNos, billingGroupNo2)
	clientBillingGroupIds = append(clientBillingGroupIds, reqStruct.ClientBillingGroupId)
	MockResponseWriter(w, resp)
}

var clientDunningGroupIds = []string{}
var dunningGroupMap = map[string]data.DunningGroupsDetails{}

func CreateDunningGroupHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.CreateDunningGroupRequest](w, body)
	if !lookupService(reqStruct.AcctNo, acctNos) && !lookupService(reqStruct.ClientAcctId, clientAcctId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "account does not exist - error fetching AcctNo or ClientAcctId",
		})
	}
	resp := response.CreateDunningGroupResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
	}
	dunningGroupMap[reqStruct.ClientDunningGroupId] = data.DunningGroupsDetails{
		DunningGroupNo:          randomDigitGen(8),
		DunningGroupName:        uuid.New()[:30],
		DunningGroupDescription: StringGen(50),
		ClientDunningGroupId:    reqStruct.ClientDunningGroupId,
		DunningProcessNo:        0,
		ClientDunningProcessId:  reqStruct.ClientDunningProcessId,
		Status:                  0,
		MasterPlanSummary:       GetDefaultMasterPlanSummary(),
	}
	clientDunningGroupIds = append(clientDunningGroupIds, reqStruct.ClientDunningGroupId)
	MockResponseWriter(w, resp)
}

func GetBillingGroupHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetGroupRequest](w, body)
	if !lookupService(reqStruct.ClientAcctId, clientAcctId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "account does not exist - error fetching AcctNo or ClientAcctId",
		})
	}
	resp := response.GetBillingGroupResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
	}
	if strictValidation {
		for _, val := range billingGroupMap {
			resp.BillingGroupDetails = append(resp.BillingGroupDetails, val)
		}
	}
	MockResponseWriter(w, resp)
}

func GetDunningGroupHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetGroupRequest](w, body)
	if !lookupService(reqStruct.ClientAcctId, clientAcctId) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "account does not exist - error fetching AcctNo or ClientAcctId",
		})
	}
	resp := response.GetDunningGroupResponse{
		AriaResponse: response.AriaResponse{ErrorCode: 0, ErrorMsg: "OK"},
	}
	if strictValidation {
		for _, val := range dunningGroupMap {
			resp.DunningGroupDetails = append(resp.DunningGroupDetails, val)
		}
	}
	MockResponseWriter(w, resp)
}

func GetAccountNoFromUserIdHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetAccountNoFromUserIdMRequest](w, body)
	if !lookupService(reqStruct.UserId, userIds) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "User ID does not exist",
		})
	}
	var acctNo int64
	if len(acctNos) == 0 {
		acctNo = randomDigitGen(8)
		acctNos = append(acctNos, acctNo)
	} else {
		acctNo = acctNos[len(acctNos)-1]
	}
	resp := response.GetAccountNoFromUserIdMResponse{
		AriaResponse: response.AriaResponse{ErrorCode: 0, ErrorMsg: "OK"},
		AcctNo:       acctNo,
	}
	MockResponseWriter(w, resp)
}

var assignPlanToAcct = []data.AllAcctPlansM{}

func AssignPlanToAccountHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.AssignAcctPlanMRequest](w, body)
	if !lookupService(reqStruct.AcctNo, acctNos) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "Acct No does not exist",
		})
	}
	assignPlanToAcct = append(assignPlanToAcct, data.AllAcctPlansM{
		ClientPlanInstanceId: reqStruct.ClientPlanInstanceId,
		PlanInstanceNo:       intcnv(reqStruct.ClientPlanInstanceId),
	})
	resp := response.AssignAcctPlanMResponse{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		PlanInstanceNo: intcnv(reqStruct.ClientPlanInstanceId),
	}
	MockResponseWriter(w, resp)
}

func GetAccountCreditHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetAccountCredits](w, body)
	fmt.Println(acctNos)
	if !lookupService(reqStruct.AcctNo, acctNos) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "Acct No does not exist",
		})
	}
	resp := response.GetAccountCredits{
		AriaResponse: response.AriaResponse{
			ErrorCode: 0,
			ErrorMsg:  "OK",
		},
		AllCredits: GetDefaultAcctCredits(reqStruct.AcctNo),
	}
	MockResponseWriter(w, resp)
}

func GetAccountCreditDetailsHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetCreditDetails](w, body)
	if !lookupService(reqStruct.AcctNo, acctNos) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "Acct No does not exist",
		})
	}
	resp := GetDefaultCreditDetails(reqStruct.ClientNo)
	MockResponseWriter(w, resp)
}

func GetAcctPlansHandler(w http.ResponseWriter, body []byte) {
	reqStruct := requestAriaMock[*request.GetAcctPlansAllMRequest](w, body)
	if !lookupService(reqStruct.AcctNo, acctNos) {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1009,
			ErrorMsg:  "Acct No does not exist",
		})
	}
	resp := response.GetAcctPlansAllMResponse{
		AriaResponse:  response.AriaResponse{ErrorCode: 0, ErrorMsg: "OK"},
		RecordCount:   1,
		AllAcctPlansM: []data.AllAcctPlansM{},
	}
	MockResponseWriter(w, resp)
}
