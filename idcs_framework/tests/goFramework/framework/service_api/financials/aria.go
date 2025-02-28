package financials

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/frisby_client"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/utils"
	"strconv"

	request "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	response "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
)

type CreditCardDetails struct {
	CCNumber      int64
	CCExpireMonth int
	CCExpireYear  int
	CCV           int
}

// Todo: This is a very initial implementation

// Aria structs

type AutoAprovalWorkFlow struct {
	request.AriaRequest
	OutputFormat       string `json:"output_format"`
	ClientNo           int64  `json:"client_no"`
	AuthKey            string `json:"auth_key"`
	ClientAcctID       string `json:"client_acct_id"`
	SuppFieldName      string `json:"supp_field_name"`
	SuppFieldValue     string `json:"supp_field_value"`
	SuppFieldDirective string `json:"supp_field_directive"`
}

func Validate_Get_Plans_Response(data []byte) bool {
	var structResponse *response.GetClientPlansAllMResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Validate_Unapplied_Credit_Response(data []byte) bool {
	var structResponse *response.GetUnappliedServiceCreditsMResponse
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

func MakeAriaCall(payload map[string]interface{}) (int, string) {
	url := financials_utils.GetAriaBaseUrl()
	crt_path, key_path := financials_utils.GetAriaCertKeyFilePath()
	frisby_response := frisby_client.AriaPost(url, "", payload, crt_path, key_path)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "ARIA POST API")
	return responseCode, responseBody
}

func GetAriaClientNo() int64 {
	ariaclientId, _ := utils.Get_Aria_Config()
	clientNo, _ := strconv.ParseInt(ariaclientId, 10, 64)
	return clientNo
}

func GetAriaAuth() string {
	_, ariaAuth := utils.Get_Aria_Config()
	return ariaAuth
}

func GetAriaAccountDetailsAllForClientId(clientAccountID string, clientNum string, authKey string) (int, string) {
	fmt.Println("Arai call: GetAriaAccountDetailsAllForClientId", clientNum, authKey)
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
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	fmt.Println("Arai call: Pyaload", payload)

	payload["client_no"] = "3760759"
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody

}

func GetAriaAccountDetails(clientAccountID string) (int, string) {
	getAccountDetailsAllRequest := request.GetAcctDetailsAllMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_details_all_m"},
		OutputFormat: "json",
		ClientNo:     GetAriaClientNo(),
		AuthKey:      GetAriaAuth(),
		ClientAcctId: "idc." + clientAccountID,
		AltCallerId:  "IDC Tester",
	}
	jsonPayload, _ := json.Marshal(getAccountDetailsAllRequest)
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	fmt.Println("Arai call: Pyaload", payload)

	payload["client_no"] = "3760759"
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody

}

func SetAutoApprovalToFalse(clientAccountID string, clientNum string, authKey string) (int, string) {
	fmt.Println("Arai call: GetAriaAccountDetailsAllForClientId", clientNum, authKey)
	clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	autoApprovalRequest := AutoAprovalWorkFlow{
		AriaRequest: request.AriaRequest{
			RestCall: "modify_acct_supp_fields_m"},
		OutputFormat:       "json",
		ClientNo:           clientNo,
		AuthKey:            authKey,
		ClientAcctID:       "idc." + clientAccountID,
		SuppFieldName:      "Auto approval by workflow",
		SuppFieldValue:     "False",
		SuppFieldDirective: "2",
	}
	jsonPayload, _ := json.Marshal(autoApprovalRequest)
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	fmt.Println("Arai call: Pyaload", payload)

	payload["client_no"] = clientNum
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody

}

func GetAriaPendingInvoiceNumberForClientId(clientAccountID, clientNum, authKey string) (int, string) {
	// clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	fmt.Println("Arai call: GetAriaPendingInvoiceNumberForClientId", clientNum, authKey)
	getInvoiceRequest := request.Invoice{
		AriaRequest: request.AriaRequest{
			RestCall: "get_pending_invoice_no_m"},
		OutputFormat: "json",
		ClientNo:     3760759,
		AuthKey:      "G9eKbQnFJNpbXQnRykXB8uKfDTqaXVnX",
		ClientAcctId: "idc." + clientAccountID,
		AltCallerId:  "IDC Tester",
	}

	jsonPayload, _ := json.Marshal(getInvoiceRequest)
	var invoicePayload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &invoicePayload)
	logger.Logf.Infof("Arai call: GetAriaPendingInvoiceNumberForClientId, Pyaload %s", invoicePayload)
	invoicePayload["client_no"] = "3760759"
	responseCode, responseBody := MakeAriaCall(invoicePayload)
	return responseCode, responseBody
}

func ManageAriaPendingInvoiceForClientId(clientAccountID, invoice_id, clientNum, authKey string, directive int64) (int, string) {
	// clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	invoiceId, _ := strconv.ParseInt(invoice_id, 10, 64)
	fmt.Print("Arai call: ManageAriaPendingInvoiceForClientId", clientNum, authKey, invoiceId)
	getManageRequest := request.ManageInvoice{
		AriaRequest: request.AriaRequest{
			RestCall: "manage_pending_invoice_m"},
		ActionDirective: directive,
		InvoiceNo:       invoiceId,
		OutputFormat:    "json",
		ClientNo:        3760759,
		AuthKey:         "G9eKbQnFJNpbXQnRykXB8uKfDTqaXVnX",
		ClientAcctId:    "idc." + clientAccountID,
	}
	jsonPayload, _ := json.Marshal(getManageRequest)
	var managePayload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &managePayload)

	logger.Logf.Infof("Arai call: ManageAriaPendingInvoiceForClientId, Pyaload %s", managePayload)
	managePayload["client_no"] = "3760759"
	responseCode, responseBody := MakeAriaCall(managePayload)
	return responseCode, responseBody

}

func GenerateAriaInvoiceForClientId(clientAccountID, clientNum, authKey string) (int, string) {
	// clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	fmt.Print("Arai call: GenerateAriaInvoiceForClientId", clientNum, authKey)
	getManageRequest := request.Invoice{
		AriaRequest: request.AriaRequest{
			RestCall: "gen_invoice_m"},
		// CombineInvoices: 1,
		OutputFormat: "json",
		ClientNo:     3760759,
		AuthKey:      "G9eKbQnFJNpbXQnRykXB8uKfDTqaXVnX",
		ClientAcctId: "idc." + clientAccountID,
	}
	jsonPayload, _ := json.Marshal(getManageRequest)
	var managePayload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &managePayload)

	logger.Logf.Infof("Arai call: GenerateAriaInvoiceForClientId, Pyaload %s", managePayload)
	responseCode, responseBody := MakeAriaCall(managePayload)
	return responseCode, responseBody
}

func Get_plans(clientNum string, authKey string) (int, string) {
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
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody

}

func Get_Client_Plans(clientNum string, authKey string, cloudAccid string) (int, string) {
	clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	getPlansRequest := request.GetAcctPlansMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_plans_m"},
		OutputFormat: "json",
		ClientNo:     clientNo,
		AuthKey:      authKey,
		ClientAcctId: "idc." + cloudAccid,
	}
	jsonPayload, _ := json.Marshal(getPlansRequest)
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody

}

func GetUnappliedServiceCredits(clientAccountID string, clientNum string, authKey string) (int, string) {
	clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	getCreditDetailsAllRequest := request.GetUnappliedServiceCreditsMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_unapplied_service_credits_m"},
		OutputFormat: "json",
		ClientNo:     clientNo,
		AuthKey:      authKey,
		ClientAcctId: clientAccountID,
	}

	jsonPayload, _ := json.Marshal(getCreditDetailsAllRequest)
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody

}

func GetUsageTypeDetails(usageTypeCode string, AriaClientId string, clientNum string, authKey string) (int, string) {
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
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody

}

func UpdateAccountBillingGroup(clientAccountID string, clientBillingGroupId string, primaryPaymentMethodNo int64) (int, string) {
	updateAccountBillingGroupRequest := request.UpdateAccountBillingGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "update_acct_billing_group_m"},
		OutputFormat:           "json",
		ClientNo:               GetAriaClientNo(),
		AuthKey:                GetAriaAuth(),
		ClientAccountId:        "idc." + clientAccountID,
		ClientBillingGroupId:   clientBillingGroupId,
		PrimaryPaymentMethodNo: primaryPaymentMethodNo,
	}
	jsonPayload, _ := json.Marshal(updateAccountBillingGroupRequest)
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody
}

func AddAccountPaymentMethod(clientAccountId string, clientPaymentMethodId string, clientBillingGroupId string, payMethodType int, creditCardDetails CreditCardDetails) (int, string) {
	updateAccountBillingGroupRequest := request.UpdateAccountBillingGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "update_acct_billing_group_m"},
		OutputFormat:                 "json",
		ClientNo:                     GetAriaClientNo(),
		AuthKey:                      GetAriaAuth(),
		ClientAccountId:              "idc." + clientAccountId,
		ClientBillingGroupId:         clientBillingGroupId,
		ClientPrimaryPaymentMethodId: clientPaymentMethodId,
		ClientPaymentMethodId:        clientPaymentMethodId,
		PayMethodType:                payMethodType,
		CCNumber:                     creditCardDetails.CCNumber,
		CCExpireMonth:                creditCardDetails.CCExpireMonth,
		CCExpireYear:                 creditCardDetails.CCExpireYear,
		CCV:                          creditCardDetails.CCV,
	}

	jsonPayload, _ := json.Marshal(updateAccountBillingGroupRequest)
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody
}

func RemovePaymentMethod(clientAccountID string, paymentMethodNo int64, clientNum string, authKey string) (int, string) {
	clientNo, _ := strconv.ParseInt(clientNum, 10, 64)
	removePaymentmethodRequest := request.RemovePaymentMethodRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "remove_acct_payment_method_m"},
		OutputFormat:    "json",
		ClientNo:        clientNo,
		AuthKey:         authKey,
		ClientAccountId: clientAccountID,
		PaymentMethodNo: paymentMethodNo,
	}

	jsonPayload, _ := json.Marshal(removePaymentmethodRequest)
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody
}

func GetPaymentMethods(clientAccountId string) (int, string) {
	getPaymentmethodsRequest := request.GetPaymentMethodsRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_payment_methods_and_terms_m"},
		OutputFormat:     "json",
		ClientNo:         GetAriaClientNo(),
		AuthKey:          GetAriaAuth(),
		ClientAccountId:  "idc." + clientAccountId,
		PaymentsReturned: 3,
		FilterStatus:     1,
	}

	jsonPayload, _ := json.Marshal(getPaymentmethodsRequest)
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody
}

// Assigning a specified account to a collections account group required to add the payment method
func AssignCollectionsAccountGroup(clientAccountId string, clientAcctGroupId string) (int, string) {
	assignCollectionsAccountGroupRequest := request.AssignCollectionsAccountGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "assign_collections_acct_group_m"},
		OutputFormat:      "json",
		ClientNo:          GetAriaClientNo(),
		AuthKey:           GetAriaAuth(),
		ClientAccountId:   "idc." + clientAccountId,
		ClientAcctGroupId: clientAcctGroupId,
	}

	jsonPayload, _ := json.Marshal(assignCollectionsAccountGroupRequest)
	var payload map[string]interface{}
	_ = json.Unmarshal(jsonPayload, &payload)
	responseCode, responseBody := MakeAriaCall(payload)
	return responseCode, responseBody
}
