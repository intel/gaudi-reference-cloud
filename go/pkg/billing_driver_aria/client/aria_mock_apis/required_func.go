// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

var strictValidation bool = false

type errorResp struct {
	ErrorCode int64  `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

var serviceType = []string{
	"Activation",

	"Recurring",

	"Usage-Based",

	"Order-Based",

	"Cancellation",

	"Minimum Fee",

	"Recurring Arrears",
}
var GroupIdx = map[string]int64{

	"BillingGroupIdx": 1,

	"DunningGroupIdx": 1,

	"PlanInstanceIdx": 1,

	"PlanInstanceStatus": 1,

	"PlanInstanceUnits": 1,
}
var serviceName = []string{}

var planName = []string{
	"Test Product/Plan Name",
}

var planType = []string{

	"Master Recurring Plan",

	"Master Pre-paid Plan",

	"Master Free Plan",

	"Supplemental Recurring Plan",

	"Supplemental Free Plan",

	"Recurring",
}
var Currency = []string{
	"usd",
	"inr",
	"gbp",
}

const (
	UsageTypeNo1            int    = 2099999401
	UsageTypeNo2            int    = 2099999332
	UsageTypeNo3            int    = 24787429023
	PromoSetNo1             int    = 10184784
	PromoSetName1           string = "IDC Plan Set"
	PromoSetDesc1           string = "IDC Plan Set"
	PromoCodeDesc1          string = "IDC Plans"
	ClientPromoSetId1       string = "IDC_Plan_Set"
	PlanDesc1               string = "Generic master plan that will be used for initial account creation. Updated"
	UsageTypeDesc1          string = "Invoice line items total value usage"
	UsageTypeName1          string = "Invoice line items total value usage"
	UsageTypeDesc2          string = "Total other status accounts usage"
	UsageTypeName2          string = "Total other status accounts usage"
	UsageTypeDesc3          string = "Minutes"
	UsageTypeName3          string = "Minutes"
	UsageTypeName4          string = "Unit"
	UsageTypeDesc4          string = "Unit"
	UsageTypeCd4            string = "Unit"
	ReasonCd1               int64  = 9999
	ReasonText1             string = "Coupon Application"
	Comments1               string = "API Testing"
	creditId2               int64  = 7584741
	ServiceNameToApply      string = "Account Credit"
	statusCd                int64  = 1
	notifyMethod            int64  = 10
	invoiceApprovalRequired int64  = 1
	billingInd              int64  = 1
	displayInd              int64  = 1
	newAcctStatus           int64  = 1
	UsageUnitTypeNo1        int    = 52
	UsageUnitTypeNo2        int    = 1
	UsageUnitTypeNo3        int    = 8
	UsageUnitTypeNo4        int    = 18
	UsageUnitTypeNo5        int    = 50
	UsageUnitTypeDesc1      string = "Call"
	UsageUnitTypeDesc2      string = "Unit"
	UsageUnitTypeDesc3      string = "Attachment"
	UsageUnitTypeDesc4      string = "Cent"
	UsageUnitTypeDesc5      string = "minutes"
	clientMasterPlanId      string = "Customer_Account_Plan"
	one                     int64  = 1
)

// some tests require a masterPlan before hand
var masterPlanNos = []int64{11306378}

var acctNos = []int64{}

var altCallerId = []string{"IDC.driver"}
var clientAcctId = []string{}
var clientPlanId = []string{}
var userIds = []string{}

// Checker Functions
func MockResponseWriter(w http.ResponseWriter, resp any) {
	errResp := errorResp{
		ErrorCode: 1001,
		ErrorMsg:  "",
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		errResp.ErrorMsg = "Unexpected: Unmarshalling Error"
		errorN, err := json.Marshal(errResp)
		if err != nil {
			errResp.ErrorMsg = "Failed to marshal json"
			err := errResp
			err_str := fmt.Sprintf("%#v", err)
			http.Error(w, err_str, http.StatusBadRequest)
			return
		}
		jsonErr, err := json.Marshal(errorN)
		if err != nil {
			errResp.ErrorMsg = "Failed to marshal json"
			err := errResp
			err_str := fmt.Sprintf("%#v", err)
			http.Error(w, err_str, http.StatusBadRequest)
			return
		}
		_, err = w.Write(jsonErr)
		if err != nil {
			errResp.ErrorMsg = "Failed to write json error"
			err := errResp
			err_str := fmt.Sprintf("%#v", err)
			http.Error(w, err_str, http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonResp)
	if err != nil {
		errResp.ErrorMsg = "Failed to write the body"
		err := errResp
		err_str := fmt.Sprintf("%#v", err)
		http.Error(w, err_str, http.StatusBadRequest)
		return
	}
}

func lookupService[T comparable](lookup T, lookupArr []T) bool {
	if len(lookupArr) == 0 {
		return false
	}
	for _, val := range lookupArr {
		if val == lookup {
			return true
		}
	}
	return false
}
func strcnv(s any) string {
	return fmt.Sprint(s)
}

func intcnv(s string) int64 {
	ans, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return ans
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func StringGen(n int) string {
	b := make([]byte, n)
	l := len(letters)
	for i := range b {
		b[i] = letters[rand.Intn(l)]
	}
	return string(b)
}

func unmarshalError(w http.ResponseWriter) {
	err := errorResp{
		ErrorCode: 1001,
		ErrorMsg:  "Unmarshalling error",
	}
	errJson, cerr := json.Marshal(err)
	if cerr != nil {
		return
	}

	if _, cerr := w.Write(errJson); cerr != nil {
		return
	}
}

func requestAriaMock[T request.Request](w http.ResponseWriter, body []byte) T {
	var reqStruct T
	err := json.Unmarshal(body, &reqStruct)
	if err != nil {
		unmarshalError(w)
	}
	return reqStruct
}

func requestAriaAdminMock[T request.Request](w http.ResponseWriter, query map[string]any) T {
	// Convert parsed query to JSON wo we can call requestAriaMock
	body, err := json.Marshal(query)
	if err != nil {
		unmarshalError(w)
	}
	return requestAriaMock[T](w, body)
}

// Auth Checker
const ClientNo = 5025576
const AuthKey = "MockAuthKey"

func AuthChecker(req map[string]interface{}, w http.ResponseWriter) bool {
	reqClientNo := int64(req["client_no"].(float64))
	auth_key := fmt.Sprint(req["auth_key"])
	if reqClientNo != ClientNo {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1004,
			ErrorMsg:  "Authentication Error: Client No is not valid/provided",
		})
		return false
	}
	if auth_key != AuthKey {
		MockResponseWriter(w, errorResp{
			ErrorCode: 1004,
			ErrorMsg:  "Authentication Error: Auth Key is not valid/provided",
		})
		return false
	}
	return true
}

// Number generator
func randomDigitGen(dig int) int64 {
	return rand.Int63n(int64(math.Pow(10, float64(dig))))
}

func CreateOrUpdateUsageTypeHandler(usageTypeCode string, usageTypeNo int) {
	usageTypeCodes = append(usageTypeCodes, usageTypeCode)

	usageTypes[usageTypeCode] = data.UsageType{
		UsageTypeNo:   usageTypeNo,
		UsageTypeDesc: "Dummy Usage Type",
		UsageUnitType: "",
		UsageTypeName: "Dummy",
		IsEditable:    false,
	}
}
