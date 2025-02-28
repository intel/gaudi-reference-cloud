package billing

import (
	"bytes"
	"encoding/json"

	"goFramework/framework/common/http_client"
	"goFramework/utils"
	"strconv"

	request "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	response "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"

	"github.com/tidwall/gjson"
)

// Todo: This is a very initial implementation

func Validate_Get_Plans_Response(data []byte) bool {
	var structResponse *response.GetClientPlansAllMResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Validate_Unapplied_Credit_Response(data []byte) bool {
	//var structResponse *response.GetUnappliedServiceCreditsMResponse
	var structResponse AriaUnappliedCredits
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Validate_Usage_Type_Response(data []byte) bool {
	var structResponse *response.GetUsageTypeDetails
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Validate_Billing_Account_Get_Response(data []byte) bool {
	var structResponse *response.GetAcctDetailsAllMResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func MakeAriaCall(jsonPayload []byte, url string, expected_status_code int) (bool, string) {
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	if expected_status_code != 0 {
		return true, jsonStr
	}
	return true, jsonStr
}

func GetAriaAccountDetailsAllForClientId(clientAccountID string, expected_status_code int) (bool, string) {
	url := utils.Get_Aria_Base_Url()
	clientNum, authKey := utils.Get_Aria_Config()
	clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	getAccountDetailsAllRequest := request.GetAcctDetailsAllMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_details_all_m"},
		OutputFormat: "json",
		ClientNo:     clientNo,
		AuthKey:      authKey,
		ClientAcctId: "idc." + clientAccountID,
		AltCallerId:  "IDC Tester",
	}
	jsonPayload, _ := json.Marshal(getAccountDetailsAllRequest)
	ret, jsonStr := MakeAriaCall(jsonPayload, url, expected_status_code)
	if ret == false {
		return ret, jsonStr
	}
	//flag := Validate_Billing_Account_Get_Response([]byte(jsonStr))
	return true, jsonStr

}
func Get_plans(expected_status_code int) (string, string) {
	url := utils.Get_Aria_Base_Url()
	clientNum, authKey := utils.Get_Aria_Config()
	clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	getPlansRequest := request.GetAllClientPlans{
		AriaRequest: request.AriaRequest{
			RestCall: "get_client_plans_all_m"},
		OutputFormat: "json",
		ClientNo:     clientNo,
		AuthKey:      authKey,
		ClientPlanId: "idc.master",
	}
	jsonPayload, _ := json.Marshal(getPlansRequest)
	ret, jsonStr := MakeAriaCall(jsonPayload, url, expected_status_code)
	if ret == false {
		return "", jsonStr
	}
	// flag := Validate_Get_Plans_Response([]byte(jsonStr))
	// if flag != true {
	// 	return "", jsonStr
	// }
	planName := gjson.Get(jsonStr, "all_client_plan_dtls.plan_name").String()
	return planName, jsonStr

}

func GetUnappliedServiceCredits(clientAccountId string, expected_status_code int, expected_credits string) (bool, string) {
	url := utils.Get_Aria_Base_Url()
	clientNum, authKey := utils.Get_Aria_Config()
	clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	getCreditDetailsAllRequest := request.GetUnappliedServiceCreditsMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_unapplied_service_credits_m"},
		OutputFormat: "json",
		ClientNo:     clientNo,
		AuthKey:      authKey,
		ClientAcctId: "idc." + clientAccountId,
	}

	jsonPayload, _ := json.Marshal(getCreditDetailsAllRequest)
	ret, jsonStr := MakeAriaCall(jsonPayload, url, expected_status_code)
	if ret == false {
		return ret, jsonStr
	}
	flag := Validate_Unapplied_Credit_Response([]byte(jsonStr))
	amount := 0
	arr := gjson.Get(jsonStr, "..#.unapplied_service_credits_details.initial_amount")
	for _, v := range arr.Array() {
		amount = amount + int(v.Num)
	}
	if strconv.Itoa(amount) == expected_credits {
		return true, jsonStr
	}
	return flag, jsonStr

}

func GetUsageTypeDetails(usageTypeCode string, expected_status_code int) (bool, string) {
	url := utils.Get_Aria_Base_Url()
	clientNum, authKey := utils.Get_Aria_Config()
	clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	getUsageTypeDetailsRequest := request.GetUsageTypeDetails{
		AriaRequest: request.AriaRequest{
			RestCall: "get_usage_type_details_m"},
		OutputFormat:  "json",
		ClientNo:      clientNo,
		AuthKey:       authKey,
		AltCallerId:   AriaClientId,
		UsageTypeCode: usageTypeCode,
	}
	jsonPayload, _ := json.Marshal(getUsageTypeDetailsRequest)
	ret, jsonStr := MakeAriaCall(jsonPayload, url, expected_status_code)
	if ret == false {
		return ret, jsonStr
	}
	flag := Validate_Usage_Type_Response([]byte(jsonStr))
	return flag, jsonStr

}
