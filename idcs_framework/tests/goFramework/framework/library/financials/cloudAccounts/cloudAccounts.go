package cloudAccounts

import (
	"bytes"
	//"compute/framework_pkg/service_apis"
	"encoding/json"
	"fmt"
	"goFramework/framework/common/http_client"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/testsetup"
	"goFramework/utils"

	//"github.com/nsf/jsondiff"
	//"github.com/google/uuid"
	"strconv"

	"github.com/tidwall/gjson"

	// "reflect"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func Validate_Get_CloudAccounts_Response_Struct(data []byte) bool {
	var structResponse GetCloudAccountResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Validate_Create_CAcc_Response_Struct(data []byte) bool {
	var structResp CreateCloudAccountResponse
	flag := utils.CompareJSONToStruct(data, structResp)
	return flag
}

func Validate_CAcc_Response2_Struct(data []byte) bool {
	var structResp GetCAccResponse2
	flag := utils.CompareJSONToStruct(data, structResp)
	return flag
}

func Validate_GetMemCAcc_Response_Struct(data []byte) bool {
	var structResp GetMemberByIdResponse
	flag := utils.CompareJSONToStruct(data, structResp)
	return flag
}

func Validate_Create_CAcc_Enroll_Response_Struct(data []byte) bool {
	var structResponse CreateCAccEnrollResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func Get_CAcc_responce(url string, expected_status_code int) (string, int) {
	apiActualResponse, responseCode := http_client.Get(url, expected_status_code)
	if apiActualResponse == "Failed" {
		return "null", responseCode
	} else {
		return apiActualResponse, responseCode
	}
}

func Get_CAcc_by_type(url string, params string, expected_status_code int) (string, int) {
	cloudaccount_endpoing_url := url + "?" + "type=" + params
	get_response_body, get_response_status := Get_CAcc_responce(cloudaccount_endpoing_url, expected_status_code)
	return get_response_body, get_response_status
}

func CAcc_RandomPayload_gen() (string, string, string, string, string) {
	Name := utils.GenerateString(5) + "@fnf.com"
	Oid := utils.GenerateString(12)
	Owner := Name
	ParentID := utils.GenerateString(12)
	Tid := utils.GenerateString(12)
	return Name, Oid, Owner, ParentID, Tid
}

func CreateCloudAccountWithOIDC(name string, oid string, owner string, parentId string, personId string, tid string, enrolled bool, lowcredits bool,
	delinquent bool, terminateMsg bool, termpaid bool, paidServAllow bool, billing_acc bool, acc_type string,
	expected_status_code int, token string, base_url string) (string, CreateCloudAccountOIDCStruct, int) {
	url := base_url + "/v1/cloudaccounts"
	createCloudAcc := CreateCloudAccountOIDCStruct{
		BillingAccountCreated:  billing_acc,
		CountryCode:            "US",
		CreditsDepleted:        "2024-03-15T17:05:58.841Z",
		Delinquent:             delinquent,
		Enrolled:               enrolled,
		LowCredits:             lowcredits,
		Name:                   name,
		Oid:                    oid,
		Owner:                  owner,
		ParentID:               parentId,
		PersonID:               personId,
		TerminateMessageQueued: terminateMsg,
		TerminatePaidServices:  termpaid,
		PaidServicesAllowed:    paidServAllow,
		Tid:                    tid,
		Type:                   acc_type,
	}
	jsonPayload, _ := json.Marshal(createCloudAcc)
	req := []byte(jsonPayload)
	reqBodyCreate := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.PostOIDC(url, reqBodyCreate, expected_status_code, token)
	fmt.Println("JSON RESPONSE....", jsonStr)
	flag := Validate_Create_CAcc_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	if !flag {
		return "False", createCloudAcc, respCode
	} else {
		cloudAccountId := gjson.Get(jsonStr, "id").String()
		return cloudAccountId, createCloudAcc, respCode
	}
}

func CreateCloudAccount(name string, oid string, owner string, parentId string, tid string, enrolled bool, lowcredits bool,
	delinquent bool, terminateMsg bool, termpaid bool, paidServAllow bool, billing_acc bool, acc_type string,
	expected_status_code int) (string, CreateCloudAccountStruct) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts"
	createCloudAcc := CreateCloudAccountStruct{
		BillingAccountCreated:  false,
		CreditsDepleted:        "2024-03-15T17:05:58.841Z",
		Delinquent:             true,
		Enrolled:               enrolled,
		LowCredits:             lowcredits,
		Name:                   name,
		Oid:                    oid,
		Owner:                  owner,
		ParentID:               "",
		PersonID:               "",
		TerminateMessageQueued: terminateMsg,
		TerminatePaidServices:  termpaid,
		PaidServicesAllowed:    paidServAllow,
		Tid:                    tid,
		Type:                   acc_type,
		CountryCode:            "IN",
		Restricted:             false,
		TradeRestricted:        false,
	}
	jsonPayload, _ := json.Marshal(createCloudAcc)
	req := []byte(jsonPayload)
	reqBodyCreate := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBodyCreate, expected_status_code)
	flag := Validate_Create_CAcc_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	if !flag {
		return "False", createCloudAcc
	} else {
		cloudAccountId := gjson.Get(jsonStr, "id").String()
		return cloudAccountId, createCloudAcc
	}
}

func DeleteCloudAccountResources(cloudAccountId string) {
	// Delete all ssh keys first
	base_url := utils.Get_ComputeUrl() + "/v1/cloudaccounts/" + cloudAccountId
	token := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccountId, utils.Get_Base_Url1(), token)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	sshkey_endpoint := base_url + "/sshpublickeys"
	get_allsshkeys_response_status, get_allsshkeys_response_body := frisby.GetAllSSHKey(sshkey_endpoint, userToken)
	if get_allsshkeys_response_status != 200 {
		logger.Logf.Info("Failed to get ssh keys")
		return
	}
	result := gjson.Get(get_allsshkeys_response_body, "items")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		sshkey_name := gjson.Get(data, "name").String()
		// Delete the key
		ssh_status, _ := frisby.DeleteSSHKeyByName(sshkey_endpoint, userToken, sshkey_name)
		logger.Logf.Info("SSH key %s deleted with status %d ", sshkey_name, ssh_status)
		return true // keep iterating
	})

	instance_endpoint := base_url + "/instances"
	response_status, response_body := frisby.GetAllInstance(instance_endpoint, userToken)
	if response_status != 200 {
		logger.Logf.Info("Failed to get instance")
		return
	}
	result = gjson.Get(response_body, "items")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		instance_id_created := gjson.Get(data, "metadata.resourceId").String()
		// Delete the key
		status, _ := frisby.DeleteInstanceById(instance_endpoint, userToken, instance_id_created)
		logger.Logf.Info("Instance %d deleted with status %d ", instance_id_created, status)
		return true // keep iterating
	})

}

func DeleteCloudAccount(cloudAccount_id string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/id/" + cloudAccount_id
	// Delete all ssh keys, vnets, instances associated with cloud account
	if os.Getenv("MultiSuites") == "" {
		DeleteCloudAccountResources(cloudAccount_id)
	}
	jsonStr, _ := http_client.Delete(url, expected_status_code)
	if jsonStr != "{}" {
		return "False", jsonStr
	} else {
		return "True", jsonStr
	}

}

func GetCAccById(cloudAccount_id string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/id/" + cloudAccount_id
	jsonStr, _ := http_client.Get(url, expected_status_code)
	flag := Validate_Get_CloudAccounts_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	if !flag {
		return "False", jsonStr
	} else {
		return "True", jsonStr
	}
}

func GetCAccByName(cloudAccount_name string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/name/" + cloudAccount_name
	jsonStr, _ := http_client.Get(url, expected_status_code)
	flag := Validate_Get_CloudAccounts_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	if !flag {
		return "False", jsonStr
	} else {
		return "True", jsonStr
	}
}

func GetCAccByAccType(cloudAccount_type string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts?type=" + cloudAccount_type
	jsonStr, statusCode := http_client.Get(url, expected_status_code)
	response := string(jsonStr)
	responses := strings.Split(response, "\n")
	var pickedResponse string
	if len(responses) > 0 {
		pickedResponse = responses[0]
		logger.Logf.Info("pickedResponse", pickedResponse)
	}
	CAcc_type := gjson.Get(pickedResponse, "result.type").String()
	flag := Validate_CAcc_Response2_Struct([]byte(pickedResponse))
	logger.Logf.Info("Flag ", flag)
	if CAcc_type != cloudAccount_type || statusCode != expected_status_code {
		return "False", jsonStr
	} else if !flag {
		return "False", jsonStr
	} else {
		return "True", jsonStr
	}
}

func GetCAccEUEnrolled(base_url string, token string, cloudAccount_type string, PUenrolled bool, expected_status_code int) (string, string) {
	url := base_url + "/v1/cloudaccounts?type=" + cloudAccount_type + "&enrolled=" + strconv.FormatBool(PUenrolled)
	logger.Log.Info("url" + url)
	jsonStr, _ := http_client.GetOIDC(url, token, expected_status_code)
	response := string(jsonStr)
	responses := strings.Split(response, "\n")
	var pickedResponse string
	if len(responses) > 0 {
		pickedResponse = responses[0]
		logger.Logf.Info("pickedResponse", pickedResponse)
	}
	CAcc_type := gjson.Get(pickedResponse, "result.type").String()
	enrolled_val := gjson.Get(pickedResponse, "result.enrolled").String()
	Conv_enroll, _ := strconv.ParseBool(enrolled_val)
	flag := Validate_CAcc_Response2_Struct([]byte(pickedResponse))
	logger.Logf.Info("Flag ", flag)
	if !flag {
		return "False", jsonStr
	} else if CAcc_type != "ACCOUNT_TYPE_ENTERPRISE" || Conv_enroll != PUenrolled {
		return "False", jsonStr
	} else {
		return "True", jsonStr
	}
}

func GetCAccPUEnrolled(cloudAccount_type string, PUenrolled bool, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts?type=" + cloudAccount_type + "&enrolled=" + strconv.FormatBool(PUenrolled)
	logger.Log.Info("url" + url)
	jsonStr, _ := http_client.Get(url, expected_status_code)
	response := string(jsonStr)
	responses := strings.Split(response, "\n")
	var pickedResponse string
	if len(responses) > 0 {
		pickedResponse = responses[0]
		logger.Logf.Info("pickedResponse", pickedResponse)
	}
	CAcc_type := gjson.Get(pickedResponse, "result.type").String()
	enrolled_val := gjson.Get(pickedResponse, "result.enrolled").String()
	Conv_enroll, _ := strconv.ParseBool(enrolled_val)
	flag := Validate_CAcc_Response2_Struct([]byte(pickedResponse))
	logger.Logf.Info("Flag ", flag)
	if !flag {
		return "False", jsonStr
	} else if CAcc_type != "ACCOUNT_TYPE_PREMIUM" || Conv_enroll != PUenrolled {
		return "False", jsonStr
	} else {
		return "True", jsonStr
	}
}

func AddMemberstoCAcc(count int, cloudAccount_Id string, expected_status_code int) (string, MembersCAccStruct) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/id/" + cloudAccount_Id + "/members/add"
	var addMembers MembersCAccStruct
	for i := 1; i <= count; i++ {
		addMembers = MembersCAccStruct{
			Members: []string{utils.GenerateInt(12)},
		}
		jsonPayload, _ := json.Marshal(addMembers)
		req := []byte(jsonPayload)
		reqBody := bytes.NewBuffer(req)
		jsonStr, _ := http_client.Post(url, reqBody, expected_status_code)
		if jsonStr != "{}" {
			return "False", addMembers
		}
	}
	return "True", addMembers
}

func GetCAccMembersById(cloudAccount_Id string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/id/" + cloudAccount_Id + "/members"
	jsonStr, _ := http_client.Get(url, expected_status_code)
	flag := Validate_GetMemCAcc_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	if !flag {
		return "False", jsonStr
	} else {
		return "True", jsonStr
	}
}

func GetCAccMembersByName(cloudAccount_name string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/name/" + cloudAccount_name + "/members"
	jsonStr, _ := http_client.Get(url, expected_status_code)
	flag := Validate_GetMemCAcc_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	if !flag {
		return "False", jsonStr
	} else {
		return "True", jsonStr
	}
}

func GetCAccByOwner(cloudAccount_owner string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts?owner=" + cloudAccount_owner
	jsonStr, statusCode := http_client.Get(url, expected_status_code)
	if statusCode == expected_status_code {
		return "True", jsonStr
	} else {
		return "False", jsonStr
	}
}

func GetCAccByTid(cloudAccount_tid string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts?tid=" + cloudAccount_tid
	jsonStr, statusCode := http_client.Get(url, expected_status_code)
	if statusCode == expected_status_code {
		return "True", jsonStr
	} else {
		return "False", jsonStr
	}
}

func GetCAccByOid(cloudAccount_oid string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts?oid=" + cloudAccount_oid
	jsonStr, statusCode := http_client.Get(url, expected_status_code)
	if statusCode == expected_status_code {
		return "True", jsonStr
	} else {
		return "False", jsonStr
	}
}

func GetCAccByTid_Oid(cloudAccount_tid string, cloudAccount_oid string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts?tid=" + cloudAccount_tid + "&oid=" + cloudAccount_oid
	jsonStr, statusCode := http_client.Get(url, expected_status_code)
	if statusCode == expected_status_code {
		return "True", jsonStr
	} else {
		return "False", jsonStr
	}
}

func DeleteMembersofCAcc(cloudAccount_id string, MemberId []string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/id/" + cloudAccount_id + "/members/delete"
	deleteMembers := MembersCAccStruct{
		Members: MemberId,
	}
	jsonPayload, _ := json.Marshal(deleteMembers)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr != "{}" {
		return "False", jsonStr
	}
	cloudAccount_id = gjson.Get(jsonStr, "id").String()
	return cloudAccount_id, jsonStr
}

func GetCAccByAnyString(cloudAccount_parentId string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts?parentId=" + cloudAccount_parentId
	jsonStr, statusCode := http_client.Get(url, expected_status_code)
	if statusCode == expected_status_code {
		return "True", jsonStr
	} else {
		return "Flase", jsonStr
	}
}

func CAcc_check_BillingAcc(cloudAccount_id string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/id/" + cloudAccount_id
	jsonStr, _ := http_client.Get(url, expected_status_code)
	CAcc_billing := gjson.Get(jsonStr, "billingAccountCreated").Bool()
	CAcc_paidServices := gjson.Get(jsonStr, "paidServicesAllowed").Bool()
	CAcc_type := gjson.Get(jsonStr, "type").String()
	CAcc_types1 := []string{"ACCOUNT_TYPE_ENTERPRISE", "ACCOUNT_TYPE_PREMIUM"}
	CAcc_types2 := []string{"ACCOUNT_TYPE_STANDARD", "ACCOUNT_TYPE_INTEL"}
	logger.Logf.Info("CAcc_billing ", CAcc_billing)
	check := utils.Contains(CAcc_type, CAcc_types1)
	check1 := utils.Contains(CAcc_type, CAcc_types2)
	if !CAcc_billing && !CAcc_paidServices && check {
		logger.Logf.Info("Failed. This cloud account doesn't have a billing account created. This account type is ", CAcc_type)
		return "False", jsonStr
	} else if !CAcc_billing && !CAcc_paidServices && !check1 {
		logger.Logf.Info("Failed. This cloud account have a billing account created. This account type is ", CAcc_type)
		return "False", jsonStr
	} else {
		logger.Log.Info("Test Case Passed.")
		return "True", jsonStr
	}
}

func GetCAccByInvalidEndpoint(value string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts?parendId=" + value
	jsonStr, statusCode := http_client.Get(url, expected_status_code)
	if statusCode == expected_status_code {
		return "True", jsonStr
	} else {
		return "False", jsonStr
	}
}

func Get_JWT_Token(tid string, username string, enterpriseId string, idp string) string {
	token_base_url := utils.Get_CA_OIDC_Url() + "/token?"
	var groups = "DevCloud%20Console%20Standard"
	url := token_base_url + "tid=" + tid + "&enterpriseId=" + enterpriseId + "&email=" + username + "&groups=" + groups + "&idp=" + idp
	logger.Log.Info("CloudAccountTokenUrl : " + url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Error("JWT Genreation Failed with Error " + err.Error())
		panic("Failed to obtain JWT token, Hence Stopping test execution")
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Error("JWT Generation Failed with Error " + err.Error())
		panic("Failed to obtain JWT token, Hence Stopping test execution")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Error("JWT Generation Failed with Error " + err.Error())
		panic("Failed to obtain JWT token, Hence Stopping test execution")
	}
	// Log the request body
	jsonStr := string(body)
	return jsonStr
}

// Random payload generation for jwt token
func Rand_token_payload_gen() string {
	//enterpriseId := utils.GenerateString(12)
	tid := utils.GenerateString(12)
	///username := utils.GenerateString(10) + "@example.com"
	return tid
}

func CreateCAccwithEnroll(acctype string, tid string, username string, enterpriseId string, premium_val bool,
	expected_status_code int) (string, string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/enroll"
	var idp string
	if acctype == "enterprise" {
		idp = "https://login.microsoftonline.com/24d2eec2-c04e-4c44-830d-f374a7b9559e/v2.0"
	} else if acctype == "standard" || acctype == "premium" || acctype == "intel" {
		idp = "intelcorpintb2c.onmicrosoft.com"
	}
	// Add azure token generation,
	tokenType := utils.Get_Token_Type()
	var token string
	if tokenType == "azure" {
		token, _ = auth.Get_Azure_Bearer_Token(username)
	}
	if tokenType == "oidc" {
		token = Get_JWT_Token(tid, username, enterpriseId, idp)
	}
	var termStatus bool = true
	os.Setenv("cloudAccTest", "True")
	os.Setenv("cloudAccToken", token)
	createCAcc := CreateCloudAccountEnrollStruct{
		Premium:     premium_val,
		TermsStatus: termStatus,
	}
	jsonPayload, _ := json.Marshal(createCAcc)
	req := []byte(jsonPayload)
	reqBodyCreate := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.Post(url, reqBodyCreate, expected_status_code)
	os.Unsetenv("cloudAccTest")
	flag := Validate_Create_CAcc_Enroll_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	cAccId := gjson.Get(jsonStr, "cloudAccountId").String()
	Action := gjson.Get(jsonStr, "action").String()
	Registered := gjson.Get(jsonStr, "registered").String()
	HaveBillingAcc := gjson.Get(jsonStr, "haveBillingAccount").String()
	HaveCloudAcc := gjson.Get(jsonStr, "haveCloudAccount").String()
	Enrolled := gjson.Get(jsonStr, "enrolled").String()
	AccType := gjson.Get(jsonStr, "cloudAccountType").String()
	if expected_status_code != respCode {
		logger.Logf.Info("Flag1 ", flag)
		return "False", AccType, jsonStr
	}
	if acctype == "premium" && Action == "ENROLL_ACTION_COUPON_OR_CREDIT_CARD" {
		return cAccId, AccType, jsonStr
	}
	if cAccId == "" || Action != "ENROLL_ACTION_NONE" || Registered != "true" || HaveBillingAcc != "true" ||
		HaveCloudAcc != "true" || Enrolled != "true" {
		logger.Logf.Info("Flag2 ", flag)
		return "False", AccType, jsonStr
	}
	if !flag {
		logger.Logf.Info("Flag3 ", flag)
		return cAccId, AccType, jsonStr
	} else {
		if acctype == "enterprise" || acctype == "premium" {
			ret, _ := billing.GetAriaAccountDetailsAllForClientId(cAccId, 200)
			if ret == false {
				logger.Logf.Info("Flag4 ", flag)
				return "False", AccType, jsonStr
			}
		}

		return cAccId, AccType, jsonStr
	}
}

func UpdateCAccById(create_tag string, cloudaccount_id string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/id/" + cloudaccount_id
	jsonData := utils.Get_cloudAccounts_Create_Payload(create_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Patch(url, reqBody, expected_status_code)
	if jsonStr != "{}" {
		return "False", "not working"
	}
	return "True", "working"
}

func UpdateCAccDuplicateOidById(oid string, cloudaccount_id string, expected_status_code int) (string, DuplicateOidStruct) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/id/" + cloudaccount_id
	UpdateOid := DuplicateOidStruct{
		Oid: oid,
	}
	jsonPayload, _ := json.Marshal(UpdateOid)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Patch(url, reqBody, expected_status_code)
	logger.Logf.Info("Result: ", jsonStr)
	if jsonStr != "{}" {
		return "False", UpdateOid
	}
	return "True", UpdateOid
}

func CAccEnsure(payload CreateCloudAccountStruct, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/ensure"
	jsonPayload, _ := json.Marshal(payload)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.Post(url, reqBody, expected_status_code)
	flag := Validate_Get_CloudAccounts_Response_Struct([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	if expected_status_code != respCode {
		return "False", jsonStr
	}
	if !flag {
		return "False", jsonStr
	} else {
		return "True", jsonStr
	}
}
