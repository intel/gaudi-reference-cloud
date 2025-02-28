package billing

import (
	"bytes"
	"fmt"
	"os"

	"strings"
	"sync"

	//"crypto/tls"

	"encoding/json"
	"goFramework/framework/common/http_client"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"

	"google.golang.org/protobuf/types/known/timestamppb"

	"strconv"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var cloId string

type MaaSProperties struct {
	ServiceType    string `json:"serviceType"`
	ProcessingType string `json:"processingType"`
}

type MaasCreateUsage struct {
	CloudAccountID string `json:"cloudAccountId"`
	EndTime        string `json:"endTime"`
	Properties     MaaSProperties
	Quantity       int    `json:"quantity"`
	Region         string `json:"region"`
	StartTime      string `json:"startTime"`
	Timestamp      string `json:"timestamp"`
	TransactionID  string `json:"transactionId"`
}

type MaasSearchUsage struct {
	CloudAccountID string `json:"cloudAccountId"`
	EndTime        string `json:"endTime"`
	Properties     MaaSProperties
	Quantity       int    `json:"quantity"`
	Region         string `json:"region"`
	StartTime      string `json:"startTime"`
	Timestamp      string `json:"timestamp"`
	TransactionID  string `json:"transactionId"`
}

func ValidateCreateBillingAccountResponse(data []byte) bool {
	var structResponse CreateCloudAccount1Response
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func ValidateUsageResponse(data []byte) bool {
	var structResponse UsageResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func ValidateInvoiceResponse(data []byte) bool {
	var structResponse InvoiceResponse
	flag := utils.CompareJSONToStruct(data, structResponse)
	return flag
}

func CreateCloudAccount(acc_type string, expected_status_code int) (string, CreateCloudAccountStruct) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts"
	createCloudAcc := CreateCloudAccountStruct{
		BillingAccountCreated:  false,
		CreditsDepleted:        "2023-03-15T17:05:58.841Z",
		Enrolled:               true,
		LowCredits:             false,
		Name:                   utils.GenerateString(10),
		Oid:                    utils.GenerateString(12),
		Owner:                  utils.GenerateString(10),
		ParentID:               utils.GenerateString(12),
		TerminateMessageQueued: true,
		TerminatePaidServices:  true,
		Tid:                    utils.GenerateString(12),
		Type:                   acc_type,
	}
	jsonPayload, _ := json.Marshal(createCloudAcc)
	req := []byte(jsonPayload)
	reqBodyCreate := bytes.NewBuffer(req)
	jsonStr, _ := http_client.Post(url, reqBodyCreate, expected_status_code)
	flag := ValidateCreateBillingAccountResponse([]byte(jsonStr))
	logger.Logf.Info("Flag ", flag)
	if !flag {
		return "False", createCloudAcc
	} else {
		cloudAccountId := gjson.Get(jsonStr, "id").String()
		return cloudAccountId, createCloudAcc
	}
}

func CreateBillingAccountWithSpecificCloudAccountIdOIDC(base_url string, cloudAccountId string, expected_status_code int, token string) (bool, string) {
	// Read Config file
	url := base_url + "/v1/billing/accounts"
	createBillingAccount := CreateBillingAccountWithCloudAccIdStruct{
		CloudAccountID: cloudAccountId,
	}
	jsonPayload, _ := json.Marshal(createBillingAccount)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.PostOIDC(url, reqBody, expected_status_code, token)

	fmt.Println("result....", jsonStr)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	//ret, _ := GetAriaAccountDetailsAllForClientId(cloudAccountId, 200)
	return true, jsonStr
}

func CreateBillingAccountWithSpecificCloudAccountId(cloudAccountId string, expected_status_code int) (bool, string) {
	// Read Config file
	url := utils.Get_Billing_Base_Url() + "/accounts"
	createBillingAccount := CreateBillingAccountWithCloudAccIdStruct{
		CloudAccountID: cloudAccountId,
	}
	jsonPayload, _ := json.Marshal(createBillingAccount)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	ret, _ := GetAriaAccountDetailsAllForClientId(cloudAccountId, 200)
	return ret, jsonStr
}

func CreateBillingAccountWithCloudAccountId(acc_type string, expected_status_code int) (bool, string) {
	// Read Config file
	var cloudAccountId string
	if acc_type == "WrongTypeCloudAccId" {
		cloudAccountId, _ = CreateCloudAccount("TEST", 200)
	} else {
		cloudAccountId, _ = CreateCloudAccount("ACCOUNT_TYPE_PREMIUM", 200)
	}

	url := utils.Get_Billing_Base_Url() + "/accounts"
	cloId = cloudAccountId
	createBillingAccount := CreateBillingAccountWithCloudAccIdStruct{
		CloudAccountID: cloudAccountId,
	}
	jsonPayload, _ := json.Marshal(createBillingAccount)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	enrolled := gjson.Get(jsonStr, "enrolled").String()
	action := gjson.Get(jsonStr, "action").String()
	if enrolled == "false" || action == "ENROLL_ACTION_RETRY" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	ret, _ := GetAriaAccountDetailsAllForClientId(cloudAccountId, 200)

	return ret, jsonStr
}

func CreateCloudCredits(acc_type string, create_tag string, expected_status_code int) (bool, string) {
	// Read Config file
	url := utils.Get_Credits_Base_Url() + "/credit"
	jsonData := utils.Get_Cloud_Credit_Create_Payload(create_tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	var newJsonPayload string
	if create_tag != "MissingCloudAccId" && create_tag != "nonexistingCloudAccId" && create_tag != "InvalidCloudAccId" && create_tag != "InvalidCloudAccIdLessLength" {
		CreateBillingAccountWithCloudAccountId(acc_type, 200)
		newJsonPayload, _ = sjson.Set(jsonPayload, "cloudAccountId", cloId)
	} else {

		newJsonPayload = jsonPayload
	}
	req := []byte(newJsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	// Expected json
	// Response: {"code":13,"message":"billing api error:FAILED_TO_CREATE_BILLING_ACCT,service error:rpc error: code = Internal desc = aria driver api error:FAILED_TO_CREATE_BILLING_ACCOUNT,aria controller error::context:FAILED_TO_UPDATE_ACCOUNT_CONTACT,aria api:update_contact_m,aria error code:1016,aria error message:invalid input","details":[]}
	if create_tag == "WrongReason" {
		if strings.Contains(jsonStr, "invalid input") {
			return true, jsonStr
		}
	}
	if create_tag == "InvalidReason" {
		if strings.Contains(jsonStr, "invalid input") {
			return true, jsonStr
		}

	}
	if create_tag == "InvalidReason" {
		if strings.Contains(jsonStr, "invalid input") {
			return true, jsonStr
		}
	}
	if create_tag == "MissingReason" {
		if strings.Contains(jsonStr, "invalid input") {
			return true, jsonStr
		}
	}
	if create_tag == "PreviousExpirationDate" {
		if strings.Contains(jsonStr, "invalid input") {
			return true, jsonStr
		}
	}

	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}

	return true, jsonStr
}

func GetBillingAccountId(cloudAccountName string, expected_status_code int) string {
	url := utils.Get_CloudAccountUrl() + "/name/" + cloudAccountName
	logger.Logf.Info("Find Usage Record URL  : ", url)
	jsonStr, _ := http_client.Get(url, expected_status_code)
	cloudAccountId := gjson.Get(jsonStr, "id").String()
	logger.Logf.Infof("Get response is %s:", cloudAccountId)
	return cloudAccountId
}

func ApplyCloudCreditsToBillingAccount(creditData CreateCloudCreditsStruct, expected_status_code int, expected_credits string) (bool, string) {
	url := utils.Get_Credits_Base_Url() + "/credit"
	jsonPayload, _ := json.Marshal(creditData)
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}

	res := GetUnappliedCloudCreditsNegative(creditData.CloudAccountID, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").String()
	logger.Logf.Info("Unapplied credits : ", unappliedCredits)
	logger.Logf.Info("Expected credits : ", expected_credits)
	if unappliedCredits == expected_credits {
		ariaResponse, _ := GetUnappliedServiceCredits(creditData.CloudAccountID, 200, expected_credits)
		if ariaResponse == true {
			return true, jsonStr
		}
		return true, jsonStr
	}
	return false, jsonStr

}

func SyncProductPCe2e(plan_name string, tag string, expected_status_code int, billing_base_url string) (bool, string) {
	url := billing_base_url + "/v1/billing" + "/sync"
	jsonData := utils.Get_PC_Sync_Payload(tag)
	fmt.Print(url)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.PostPC(url, reqBody, expected_status_code)
	logger.Log.Info("Response from sync" + jsonStr)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	planName, jsonStr := Get_plans(200)
	if planName != plan_name {
		return false, jsonStr
	}
	return true, jsonStr
}

func SyncProductCatalog(plan_name string, tag string, expected_status_code int) (bool, string) {
	url := utils.Get_Billing_Base_Url() + "/sync"
	jsonData := utils.Get_PC_Sync_Payload(tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.Post(url, reqBody, expected_status_code)
	logger.Log.Info("Response from sync" + jsonStr)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	planName, jsonStr := Get_plans(200)
	if planName != plan_name {
		return false, jsonStr
	}
	return true, jsonStr

}

func GetUnappliedCloudCredits(tag string, expected_status_code int) string {
	jsonData := utils.Get_Cloud_Credit_Create_Payload(tag)
	cloudAccountId := GetBillingAccountId(gjson.Get(jsonData, "payload.cloudAccountName").String(), 200)
	url := utils.Get_Credits_Base_Url() + "/credit" + "/unapplied?cloudAccountId=" + cloudAccountId
	logger.Logf.Info("Find Usage Record URL  : ", url)
	jsonStr, _ := http_client.Get(url, expected_status_code)
	return jsonStr
}

func GetUnappliedCloudCreditsNegative(cloudAccountId string, expected_status_code int) string {
	url := utils.Get_Credits_Base_Url() + "/credit"
	logger.Logf.Info("Find Usage Record URL  : ", url)
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccountId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	token := "Bearer " + userToken
	if cloudAccountId == "" {
		user := utils.Get_UserName("Premium")
		token, _ = auth.Get_Azure_Bearer_Token(user)
		token = "Bearer " + token
	}
	status, response := financials.GetUnappliedCredits(url, token, cloudAccountId)
	logger.Logf.Infof("Unapplied Credit Status : %s ", status)
	return response
}

func GetCloudCredits(cloudAccountId string, expected_status_code int) (bool, string) {
	url := utils.Get_Credits_Base_Url() + "/credit" + "?cloudAccountId=" + cloudAccountId
	logger.Logf.Info("Get Cloud Credits URL  : %s ", url)
	jsonStr, respCode := http_client.Get(url, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	if jsonStr == "{}" {
		return true, jsonStr
	}
	return true, jsonStr

}

func CreateBillingAccWithCloudAccId(tag string, expected_status_code int) (bool, string) {
	url := utils.Get_CloudAccountUrl()
	jsonData := utils.Get_Cloud_Credit_Create_Payload(tag)
	jsonPayload := gjson.Get(jsonData, "payload").String()
	req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(req)
	jsonStr, respCode := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	return true, jsonStr

}

func GetExpirationTime(creationTime string) string {
	layout := "2006-01-02T15:04:05.99.282805528Z"
	t, err := time.Parse(layout, creationTime)
	//t, err := time.Parse(time.RFC3339nano, creationTime)
	if err != nil {
		panic(err)
	}
	pb := timestamppb.New(t)
	logger.Logf.Infof("New time %s", pb.AsTime().AddDate(0, 0, 0).UTC().Format(time.RFC3339))
	ninetyDaysFromCreation := t.AddDate(0, 0, 90)
	expires := timestamppb.New(ninetyDaysFromCreation)
	creation_timestamp := t.UTC().Format(time.RFC3339)
	expiration_timestamp := expires.AsTime().UTC().Format(time.RFC3339)
	logger.Logf.Infof("Creation time %s", creation_timestamp)
	logger.Logf.Infof("Expiration time %s", expiration_timestamp)
	return expiration_timestamp
}

func GetCreationExpirationTime() (string, string) {
	time.Sleep(1 * time.Second)
	oneMinFromCurrentTime := timestamppb.Now().AsTime().Add(1 * time.Minute)
	creationTime := timestamppb.New(oneMinFromCurrentTime)
	ninetyDaysFromCreation := creationTime.AsTime().AddDate(0, 0, 90)
	expires := timestamppb.New(ninetyDaysFromCreation)
	creation_timestamp := creationTime.AsTime().UTC().Format(time.RFC3339)
	expiration_timestamp := expires.AsTime().UTC().Format(time.RFC3339)
	logger.Logf.Infof("Creation time %s", creation_timestamp)
	logger.Logf.Infof("Expiration time %s", expiration_timestamp)
	return creation_timestamp, expiration_timestamp
}

func GetExpirationInOneMinute() (string, string) {
	time.Sleep(1 * time.Second)
	oneMinFromCurrentTime := timestamppb.Now().AsTime().Add(1 * time.Minute)
	creationTime := timestamppb.New(oneMinFromCurrentTime)
	oneMinFromCreation := creationTime.AsTime().Add(1 * time.Minute)
	expires := timestamppb.New(oneMinFromCreation)
	creation_timestamp := creationTime.AsTime().UTC().Format(time.RFC3339)
	expiration_timestamp := expires.AsTime().UTC().Format(time.RFC3339)
	logger.Logf.Infof("Creation time %s", creation_timestamp)
	logger.Logf.Infof("Expiration time %s", expiration_timestamp)
	return creation_timestamp, expiration_timestamp
}

func GetExpirationInThreeMinute() (string, string) {
	time.Sleep(1 * time.Second)
	var lastCreditExpiry = time.Now().AddDate(-100, 0, 0)
	oneMinFromCurrentTime := timestamppb.Now().AsTime().Add(1 * time.Minute)
	creationTime := timestamppb.New(oneMinFromCurrentTime)
	oneMinFromCreation := creationTime.AsTime().Add(3 * time.Minute)
	expires := timestamppb.New(oneMinFromCreation)
	creation_timestamp := creationTime.AsTime().UTC().Format(time.RFC3339)
	expiration_timestamp := expires.AsTime().UTC().Format(time.RFC3339)
	logger.Logf.Infof("Creation time %s", creation_timestamp)
	logger.Logf.Infof("Expiration time %s", expiration_timestamp)
	logger.Logf.Infof("lastCreditExpiry %s", lastCreditExpiry)
	return creation_timestamp, expiration_timestamp
}

func GetExpirationInTime(t time.Duration) (string, string) {
	time.Sleep(1 * time.Second)
	var lastCreditExpiry = time.Now().AddDate(-100, 0, 0)
	oneMinFromCurrentTime := timestamppb.Now().AsTime().Add(1 * time.Minute)
	creationTime := timestamppb.New(oneMinFromCurrentTime)
	oneMinFromCreation := creationTime.AsTime().Add(t * time.Minute)
	expires := timestamppb.New(oneMinFromCreation)
	creation_timestamp := creationTime.AsTime().UTC().Format(time.RFC3339)
	expiration_timestamp := expires.AsTime().UTC().Format(time.RFC3339)
	logger.Logf.Infof("Creation time %s", creation_timestamp)
	logger.Logf.Infof("Expiration time %s", expiration_timestamp)
	logger.Logf.Infof("lastCreditExpiry %s", lastCreditExpiry)
	return creation_timestamp, expiration_timestamp
}

func GetPremiumCouponExpiry(creation_time string, shorterExpiry bool) string {
	created, _ := time.Parse(time.RFC3339, creation_time)
	var creditExpiry string
	if shorterExpiry {
		after := created.AddDate(0, 0, 38).Format(time.RFC3339)
		expiryTS, _ := time.Parse(time.RFC3339, after)
		expiryTS1 := expiryTS.Format("2006-01-02")
		creditExpiry = expiryTS1 + "T00:00:00Z"
	} else {
		after := created.AddDate(0, 0, 128).Format(time.RFC3339)
		expiryTS, _ := time.Parse(time.RFC3339, after)
		expiryTS1 := expiryTS.Format("2006-01-02")
		creditExpiry = expiryTS1 + "T00:00:00Z"

	}

	return creditExpiry

}

func GetCouponExpiry(creation_time string) string {
	created, _ := time.Parse(time.RFC3339, creation_time)
	var creditExpiry string
	after := created.AddDate(0, 0, 90).Format(time.RFC3339)
	expiryTS, _ := time.Parse(time.RFC3339, after)
	expiryTS1 := expiryTS.Format("2006-01-02")
	creditExpiry = expiryTS1 + "T00:00:00Z"
	return creditExpiry
}

func ExpireCoupon(minutes time.Duration) (string, string) {
	time.Sleep(1 * time.Second)

	var lastCreditExpiry = time.Now().AddDate(-100, 0, 0)
	oneMinFromCurrentTime := timestamppb.Now().AsTime().Add(1 * time.Minute)
	creationTime := timestamppb.New(oneMinFromCurrentTime)
	oneMinFromCreation := creationTime.AsTime().Add(minutes * time.Minute)
	expires := timestamppb.New(oneMinFromCreation)
	creation_timestamp := creationTime.AsTime().UTC().Format(time.RFC3339)
	expiration_timestamp := expires.AsTime().UTC().Format(time.RFC3339)
	logger.Logf.Infof("Creation time %s", creation_timestamp)
	logger.Logf.Infof("Expiration time %s", expiration_timestamp)
	logger.Logf.Infof("lastCreditExpiry %s", lastCreditExpiry)
	return creation_timestamp, expiration_timestamp
}

func GetLesserExpirationTime() (string, string) {
	oneMinFromCurrentTime := timestamppb.Now().AsTime().Add(time.Second)
	creationTime := timestamppb.New(oneMinFromCurrentTime)
	ninetyDaysFromCreation := creationTime.AsTime().AddDate(0, 0, 2)
	expires := timestamppb.New(ninetyDaysFromCreation)
	creation_timestamp := creationTime.AsTime().UTC().Format(time.RFC3339)
	expiration_timestamp := expires.AsTime().UTC().Format(time.RFC3339)
	logger.Logf.Infof("Creation time %s", creation_timestamp)
	logger.Logf.Infof("Expiration time %s", expiration_timestamp)
	return creation_timestamp, expiration_timestamp
}

func CreateCoupon(couponPayload []byte, expected_status_code int) (bool, string) {
	url := utils.Get_Credits_Base_Url() + "/coupons"
	//jsonPayload, _ := json.Marshal(couponPayload)
	//req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(couponPayload)
	jsonStr, respCode := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	return true, jsonStr

}

func RedeemCoupon(couponPayload []byte, expected_status_code int) (bool, string) {
	var jsonString RedeemCouponStruct
	base_url := utils.Get_Base_Url1()
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	// jsonPayload, _ := json.Marshal(couponPayload)
	// req := []byte(jsonPayload)
	if err := json.Unmarshal([]byte(couponPayload), &jsonString); err != nil {
		return false, err.Error()
	}
	userCloudAccId := jsonString.CloudAccountID
	code := jsonString.Code
	userName, _ := testsetup.GetCloudAccountUserName(userCloudAccId, base_url, authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	token := "Bearer " + userToken
	redeem_coupon_endpoint := utils.Get_Credits_Base_Url() + "/coupons/redeem"
	redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), code, userCloudAccId)
	coupon_redeem_status, jsonStr := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
	logger.Logf.Info("RedeemCoupon response  : %s", jsonStr)
	if coupon_redeem_status != 200 {
		return true, jsonStr
	}
	return true, jsonStr

}

func DisableCoupon(couponPayload []byte, expected_status_code int) (bool, string) {
	url := utils.Get_Credits_Base_Url() + "/coupons/disable"
	// jsonPayload, _ := json.Marshal(couponPayload)
	// req := []byte(jsonPayload)
	reqBody := bytes.NewBuffer(couponPayload)
	jsonStr, respCode := http_client.Post(url, reqBody, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	return true, jsonStr

}

func GetCoupons(couponCode string, expected_status_code int) (bool, string) {
	url := utils.Get_Credits_Base_Url() + "/coupons" + "?code=" + couponCode
	logger.Logf.Info("Get Coupons URL  : ", url)
	jsonStr, respCode := http_client.Get(url, expected_status_code)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	if jsonStr == "{}" {
		return true, jsonStr
	}
	return true, jsonStr
}

func GetUsage(cloudAccId string, expected_status_code int) (bool, string) {
	usage_url := utils.Get_Billing_Base_Url() + "/usages?cloudAccountId=" + cloudAccId
	logger.Logf.Info("Get Usage URL  : ", usage_url)
	jsonStr, respCode := http_client.Get(usage_url, 200)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	if jsonStr == "{}" {
		return true, jsonStr
	}
	flag := ValidateUsageResponse([]byte(jsonStr))
	return flag, jsonStr

}

func GetUsageHistory(cloudAccId string, startDate string, endDate string, expected_status_code int) (bool, string) {
	usage_url := utils.Get_Billing_Base_Url() + "/usages?cloudAccountId=" + cloudAccId + "&searchStart=" + startDate + "&searchEnd=" + endDate
	logger.Logf.Info("Get Usage URL  : ", usage_url)
	jsonStr, respCode := http_client.Get(usage_url, 200)
	logger.Logf.Infof("Get response is %s:", jsonStr)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	if jsonStr == "{}" {
		return true, jsonStr
	}
	flag := ValidateUsageResponse([]byte(jsonStr))
	return flag, jsonStr

}

func CalculateandValidateUsage(runningSecs string, rate string) float64 {
	runFactor, _ := strconv.ParseFloat(runningSecs, 64)
	usage := runFactor / 60
	rateFactor, _ := strconv.ParseFloat(rate, 64)
	calcAmount := usage * rateFactor
	return calcAmount
}

func GetInvoice(cloudAccId string, expected_status_code int) (bool, string) {
	usage_url := utils.Get_Billing_Base_Url() + "/invoices?cloudAccountId=" + cloudAccId
	logger.Logf.Info("Get Usage URL  : ", usage_url)
	jsonStr, respCode := http_client.Get(usage_url, 200)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	if jsonStr == "{}" {
		return true, jsonStr
	}
	flag := ValidateInvoiceResponse([]byte(jsonStr))
	return flag, jsonStr

}

func GetInvoicesHistory(cloudAccId string, startDate string, endDate string, expected_status_code int) (bool, string) {
	usage_url := utils.Get_Billing_Base_Url() + "/invoices?cloudAccountId=" + cloudAccId + "&searchStart=" + startDate + "&searchEnd=" + endDate
	logger.Logf.Info("Get Usage URL  : ", usage_url)
	jsonStr, respCode := http_client.Get(usage_url, 200)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	if jsonStr == "{}" {
		return true, jsonStr
	}
	flag := ValidateInvoiceResponse([]byte(jsonStr))
	return flag, jsonStr

}

func GetInvoicesById(cloudAccId string, id string, expected_status_code int) (bool, string) {
	usage_url := utils.Get_Billing_Base_Url() + "/invoices?cloudAccountId=" + cloudAccId + "&id=" + id
	logger.Logf.Info("Get Usage URL  : ", usage_url)
	jsonStr, respCode := http_client.Get(usage_url, 200)
	if jsonStr == "Failed" {
		return false, jsonStr
	}
	logger.Logf.Infof("")
	if expected_status_code != respCode {
		return false, jsonStr
	}
	if expected_status_code != 200 {
		return true, jsonStr
	}
	if jsonStr == "{}" {
		return true, jsonStr
	}
	flag := ValidateInvoiceResponse([]byte(jsonStr))
	return flag, jsonStr

}

func Create_Redeem_Coupon(userType string, couponAmount int64, no_of_user int64, cloudAccId string) error {
	creation_time, expirationtime := GetCreationExpirationTime()
	var isStandard bool = false
	//isStandard = false
	if userType == "Standard" {
		isStandard = true
	}
	createCoupon := StandardCreateCouponStruct{
		Amount:     couponAmount,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    no_of_user,
		IsStandard: isStandard,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := CreateCoupon(req, 200)
	if !ret_value {
		return fmt.Errorf("create Coupon Failed")
	}
	couponCode := gjson.Get(data, "code").String()
	if createCoupon.Amount != gjson.Get(data, "amount").Int() {
		return fmt.Errorf("expected Amount from coupon Creation : %d, But Got : %d", createCoupon.Amount, gjson.Get(data, "amount").Int())
	}

	if createCoupon.Creator != gjson.Get(data, "creator").String() {
		return fmt.Errorf("expected Coupon Creator Name  : %s, But Got : %s", createCoupon.Creator, gjson.Get(data, "creator").String())
	}

	if createCoupon.Expires != gjson.Get(data, "expires").String() {
		return fmt.Errorf("expected Coupon Expiry Time  : %s, But Got : %s", createCoupon.Expires, gjson.Get(data, "expires").String())
	}

	if createCoupon.NumUses != gjson.Get(data, "numUses").Int() {
		return fmt.Errorf("expected Coupon Number of Uses : %d, But Got : %d", createCoupon.NumUses, gjson.Get(data, "numUses").Int())
	}

	if createCoupon.Start != gjson.Get(data, "start").String() {
		return fmt.Errorf("expected Coupon Start date : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
	}

	// Get coupon and validate
	getret_value, getdata := GetCoupons(couponCode, 200)
	if !getret_value {
		return fmt.Errorf("get on Coupons Failed")
	}
	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {

		if createCoupon.Amount != gjson.Get(val.String(), "amount").Int() {
			return fmt.Errorf("expected Amount from Get coupon Creation Time : %d, But Got : %d", createCoupon.Amount, gjson.Get(data, "amount").Int())
		}

		if createCoupon.Creator != gjson.Get(val.String(), "creator").String() {
			return fmt.Errorf("expected Coupon Get Creator Name  : %s, But Got : %s", createCoupon.Creator, gjson.Get(data, "creator").String())
		}

		if createCoupon.Expires != gjson.Get(val.String(), "expires").String() {
			return fmt.Errorf("expected Get Coupon Expiry Time  : %s, But Got : %s", createCoupon.Expires, gjson.Get(data, "expires").String())
		}

		if createCoupon.NumUses != gjson.Get(val.String(), "numUses").Int() {
			return fmt.Errorf("expected Get Coupon Number of Uses : %d, But Got : %d", createCoupon.NumUses, gjson.Get(data, "numUses").Int())
		}

		if createCoupon.Start != gjson.Get(val.String(), "start").String() {
			return fmt.Errorf("expected Get Coupon Start date : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
		}

		if gjson.Get(val.String(), "numRedeemed").String() != "0" {
			return fmt.Errorf("expected Get numRedeemed : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
		}

	}

	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(1 * time.Minute)
	os.Setenv("cloudAccId", cloudAccId)
	redeemCoupon := RedeemCouponStruct{
		CloudAccountID: cloudAccId,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	ff, str := RedeemCoupon(req, 200)
	logger.Logf.Infof("Redeem coupon bool : %t", ff)
	logger.Logf.Infof("Redeem coupon response : %s ", str)

	// Get coupon and validate
	// _, getdata = GetCoupons(couponCode, 200)
	// couponData = gjson.Get(getdata, "coupons")
	// redemptions := gjson.Get(getdata, "result.redemptions")
	// for _, val := range redemptions.Array() {
	// 	if cloudAccId != gjson.Get(val.String(), "cloudAccountId").String() {
	// 		return fmt.Errorf("get Cloud Acc Id from Coupon Data after redemption Expected : %s, But Got : %s", cloudAccId, gjson.Get(val.String(), "cloudAccountId").String())
	// 	}

	// 	if couponCode != gjson.Get(val.String(), "code").String() {
	// 		return fmt.Errorf("get couponCode from Coupon Data after redemption Expected : %s, But Got : %s", couponCode, gjson.Get(val.String(), "code").String())
	// 	}

	// 	if gjson.Get(val.String(), "installed").String() != "true" {
	// 		return fmt.Errorf("get installed from Coupon Data after redemption Expected : %s, But Got : %s", "true", gjson.Get(val.String(), "installed").String())
	// 	}

	// }
	// for _, val := range couponData.Array() {
	// 	if gjson.Get(val.String(), "numRedeemed").String() != "1" {
	// 		return fmt.Errorf("get numRedeemed from Coupon Data after redemption Expected : %s, But Got : %s", "1", gjson.Get(val.String(), "numRedeemed").String())
	// 	}

	// }
	return nil
}

func Credit_Migrate(cloudAccId string, token string) error {
	base_url := utils.Get_Credits_Base_Url() + "/credit/creditmigrate"
	payload := financials_utils.EnrichCreditMigratePayload(financials_utils.GetCreditMigratePayload(), cloudAccId)
	logger.Logf.Infof("Credit Migrate Url : %s", base_url)
	logger.Logf.Infof("Credit Migrate Payload : %s ", payload)
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	response_status, responseBody := financials.UpgradeWithCoupon(base_url, userToken, payload)
	if response_status != 200 {
		return fmt.Errorf("failed to migrate credits after upgrade for  cloud account : %s, error : %s", cloudAccId, responseBody)
	}
	return nil

}

func CreateCouponToRedeem(userType string, couponAmount int64, no_of_user int64, cloudAccId string, token string) (string, error) {
	creation_time, expirationtime := GetCreationExpirationTime()
	var isStandard bool = false
	//isStandard = false
	if userType == "Standard" {
		isStandard = true
	}
	createCoupon := StandardCreateCouponStruct{
		Amount:     couponAmount,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    no_of_user,
		IsStandard: isStandard,
	}

	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := CreateCoupon(req, 200)
	if !ret_value {
		return "", fmt.Errorf("create Coupon Failed")
	}
	couponCode := gjson.Get(data, "code").String()
	if createCoupon.Amount != gjson.Get(data, "amount").Int() {
		return "", fmt.Errorf("expected Amount from coupon Creation : %d, But Got : %d", createCoupon.Amount, gjson.Get(data, "amount").Int())
	}

	if createCoupon.Creator != gjson.Get(data, "creator").String() {
		return "", fmt.Errorf("expected Coupon Creator Name  : %s, But Got : %s", createCoupon.Creator, gjson.Get(data, "creator").String())
	}

	if createCoupon.Expires != gjson.Get(data, "expires").String() {
		return "", fmt.Errorf("expected Coupon Expiry Time  : %s, But Got : %s", createCoupon.Expires, gjson.Get(data, "expires").String())
	}

	if createCoupon.NumUses != gjson.Get(data, "numUses").Int() {
		return "", fmt.Errorf("expected Coupon Number of Uses : %d, But Got : %d", createCoupon.NumUses, gjson.Get(data, "numUses").Int())
	}

	if createCoupon.Start != gjson.Get(data, "start").String() {
		return "", fmt.Errorf("expected Coupon Start date : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
	}

	// Get coupon and validate
	getret_value, getdata := GetCoupons(couponCode, 200)
	if !getret_value {
		return "", fmt.Errorf("get on Coupons Failed")
	}
	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		if createCoupon.Amount != gjson.Get(val.String(), "amount").Int() {
			return couponCode, fmt.Errorf("expected Amount from Get coupon Creation Time : %d, But Got : %d", createCoupon.Amount, gjson.Get(data, "amount").Int())
		}

		if createCoupon.Creator != gjson.Get(val.String(), "creator").String() {
			return couponCode, fmt.Errorf("expected Coupon Get Creator Name  : %s, But Got : %s", createCoupon.Creator, gjson.Get(data, "creator").String())
		}

		if createCoupon.Expires != gjson.Get(val.String(), "expires").String() {
			return couponCode, fmt.Errorf("expected Get Coupon Expiry Time  : %s, But Got : %s", createCoupon.Expires, gjson.Get(data, "expires").String())
		}

		if createCoupon.NumUses != gjson.Get(val.String(), "numUses").Int() {
			return couponCode, fmt.Errorf("expected Get Coupon Number of Uses : %d, But Got : %d", createCoupon.NumUses, gjson.Get(data, "numUses").Int())
		}

		if createCoupon.Start != gjson.Get(val.String(), "start").String() {
			return couponCode, fmt.Errorf("expected Get Coupon Start date : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
		}

		if gjson.Get(val.String(), "numRedeemed").String() != "0" {
			return couponCode, fmt.Errorf("expected Get numRedeemed : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
		}

	}

	return couponCode, nil

}

func Standard_to_premium_upgrade_with_coupon(couponType string, couponAmount int64, no_of_user int64, cloudAccId string, token string, cloudAccountUpgradeToType string) error {
	couponCode, err := CreateCouponToRedeem(couponType, couponAmount, no_of_user, cloudAccId, token)
	if err != nil {
		return err
	}
	//Redeem coupon to upgrade to premium
	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(1 * time.Minute)
	base_url := utils.Get_Base_Url1() + "/v1/cloudaccounts/upgrade"
	Coupon_api_payload := financials_utils.EnrichUpgradeCouponPayload(financials_utils.GetUpgradeCouponPayload(), cloudAccId, cloudAccountUpgradeToType, couponCode)
	logger.Logf.Infof("Upgrade coupon payload : %s ", Coupon_api_payload)
	response_status, responseBody := financials.UpgradeWithCoupon(base_url, token, Coupon_api_payload)
	if response_status != 200 {
		return fmt.Errorf("failed to upgrade cloud account : %s, error : %s", cloudAccId, responseBody)
	}

	// Get coupon and validate
	_, getdata := GetCoupons(couponCode, 200)
	couponData := gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		if cloudAccId != gjson.Get(val.String(), "cloudAccountId").String() {
			return fmt.Errorf("get Cloud Acc Id from Coupon Data after redemption Expected : %s, But Got : %s", cloudAccId, gjson.Get(val.String(), "cloudAccountId").String())
		}

		if couponCode != gjson.Get(val.String(), "code").String() {
			return fmt.Errorf("get couponCode from Coupon Data after redemption Expected : %s, But Got : %s", couponCode, gjson.Get(val.String(), "code").String())
		}

		if gjson.Get(val.String(), "installed").String() != "true" {
			return fmt.Errorf("get installed from Coupon Data after redemption Expected : %s, But Got : %s", "true", gjson.Get(val.String(), "installed").String())
		}

	}
	for _, val := range couponData.Array() {
		if gjson.Get(val.String(), "numRedeemed").String() != "1" {
			return fmt.Errorf("get numRedeemed from Coupon Data after redemption Expected : %s, But Got : %s", "1", gjson.Get(val.String(), "numRedeemed").String())
		}

	}
	return nil
}

func Standard_to_premium_upgrade_with_cc(creditCardPayload string, userName string, password string, consoleurl string, replaceUrl string, token string) error {
	financials.UpgradeThroughCreditCard(creditCardPayload, userName, password, consoleurl, replaceUrl)
	base_url := utils.Get_Base_Url1()
	cloudaccId, _ := testsetup.GetCloudAccountId(userName, base_url, token)
	bilingOptions, bilingOptions1 := financials.GetBillingOptions(base_url, token, cloudaccId)
	logger.Log.Info("bilingOptions1" + bilingOptions1)
	logger.Logf.Infof("bilingOptions", bilingOptions)
	// get_CAcc_id := cloudaccId

	// assert.Equal(suite.T(), bilingOptions, 200, "Failed: Failed to get billing options")
	// assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.suffix").String(), "1111", "Failed to validate credit card suffix")
	// assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.expiration").String(), "10/2028", "Failed to validate credit card expiration")
	// assert.Equal(suite.T(), gjson.Get(bilingOptions1, "creditCard.type").String(), "Visa", "Failed to validate credit card type")
	// assert.Equal(suite.T(), gjson.Get(bilingOptions1, "email").String(), "testuservisa@premium.com", "Failed to validate user email")
	// assert.Equal(suite.T(), gjson.Get(bilingOptions1, "firstName").String(), "Test Visa", "Failed to validate First name")
	// assert.Equal(suite.T(), gjson.Get(bilingOptions1, "lastName").String(), "Visa User", "Failed to validate Last name")
	// assert.Equal(suite.T(), gjson.Get(bilingOptions1, "paymentType").String(), "PAYMENT_CREDIT_CARD", "Failed to validate payment type")
	// assert.Equal(suite.T(), gjson.Get(bilingOptions1, "cloudAccountId").String(), cloudaccId, "Failed to validate cloud Account Id")

	if bilingOptions != 200 {
		return fmt.Errorf("response code from billing options is not correct Expected : 200, But Got : %d", bilingOptions)
	}
	// cardNumber := gjson.Get(creditCardPayload, "CardnumberInput").String()
	// suffix := cardNumber[len(cardNumber)-4:]
	// if gjson.Get(bilingOptions1, "creditCard.suffix").String() != suffix {
	// 	return fmt.Errorf("expected suffix from billing options  : %s, But Got : %s", suffix, gjson.Get(bilingOptions1, "creditCard.suffix").String())
	// }

	return nil
}

func Create_Redeem_Coupon_With_Shrt_Expiry(userType string, couponAmount int64, no_of_user int64, cloudAccId string, expiration time.Duration) error {
	creation_time, expirationtime := ExpireCoupon(expiration)
	var isStandard bool = false
	//isStandard = false
	if userType == "Standard" {
		isStandard = true
	}
	createCoupon := StandardCreateCouponStruct{
		Amount:     couponAmount,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    no_of_user,
		IsStandard: isStandard,
	}
	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := CreateCoupon(req, 200)
	if !ret_value {
		return fmt.Errorf("create Coupon Failed")
	}
	couponCode := gjson.Get(data, "code").String()
	if createCoupon.Amount != gjson.Get(data, "amount").Int() {
		return fmt.Errorf("expected Amount from coupon Creation : %d, But Got : %d", createCoupon.Amount, gjson.Get(data, "amount").Int())
	}

	if createCoupon.Creator != gjson.Get(data, "creator").String() {
		return fmt.Errorf("expected Coupon Creator Name  : %s, But Got : %s", createCoupon.Creator, gjson.Get(data, "creator").String())
	}

	if createCoupon.Expires != gjson.Get(data, "expires").String() {
		return fmt.Errorf("expected Coupon Expiry Time  : %s, But Got : %s", createCoupon.Expires, gjson.Get(data, "expires").String())
	}

	if createCoupon.NumUses != gjson.Get(data, "numUses").Int() {
		return fmt.Errorf("expected Coupon Number of Uses : %d, But Got : %d", createCoupon.NumUses, gjson.Get(data, "numUses").Int())
	}

	if createCoupon.Start != gjson.Get(data, "start").String() {
		return fmt.Errorf("expected Coupon Start date : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
	}

	// Get coupon and validate
	getret_value, getdata := GetCoupons(couponCode, 200)
	if !getret_value {
		return fmt.Errorf("get on Coupons Failed")
	}
	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {

		if createCoupon.Amount != gjson.Get(val.String(), "amount").Int() {
			return fmt.Errorf("expected Amount from Get coupon Creation Time : %d, But Got : %d", createCoupon.Amount, gjson.Get(data, "amount").Int())
		}

		if createCoupon.Creator != gjson.Get(val.String(), "creator").String() {
			return fmt.Errorf("expected Coupon Get Creator Name  : %s, But Got : %s", createCoupon.Creator, gjson.Get(data, "creator").String())
		}

		if createCoupon.Expires != gjson.Get(val.String(), "expires").String() {
			return fmt.Errorf("expected Get Coupon Expiry Time  : %s, But Got : %s", createCoupon.Expires, gjson.Get(data, "expires").String())
		}

		if createCoupon.NumUses != gjson.Get(val.String(), "numUses").Int() {
			return fmt.Errorf("expected Get Coupon Number of Uses : %d, But Got : %d", createCoupon.NumUses, gjson.Get(data, "numUses").Int())
		}

		if createCoupon.Start != gjson.Get(val.String(), "start").String() {
			return fmt.Errorf("expected Get Coupon Start date : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
		}

		if gjson.Get(val.String(), "numRedeemed").String() != "0" {
			return fmt.Errorf("expected Get numRedeemed : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
		}

	}

	//Redeem coupon
	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(1 * time.Minute)
	redeemCoupon := RedeemCouponStruct{
		CloudAccountID: cloudAccId,
		Code:           couponCode,
	}
	jsonPayload, _ = json.Marshal(redeemCoupon)
	req = []byte(jsonPayload)
	_, _ = RedeemCoupon(req, 200)

	// Get coupon and validate
	_, getdata = GetCoupons(couponCode, 200)
	couponData = gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		if cloudAccId != gjson.Get(val.String(), "cloudAccountId").String() {
			return fmt.Errorf("get Cloud Acc Id from Coupon Data after redemption Expected : %s, But Got : %s", cloudAccId, gjson.Get(val.String(), "cloudAccountId").String())
		}

		if couponCode != gjson.Get(val.String(), "code").String() {
			return fmt.Errorf("get couponCode from Coupon Data after redemption Expected : %s, But Got : %s", couponCode, gjson.Get(val.String(), "code").String())
		}

		if gjson.Get(val.String(), "installed").String() != "true" {
			return fmt.Errorf("get installed from Coupon Data after redemption Expected : %s, But Got : %s", "true", gjson.Get(val.String(), "installed").String())
		}

	}
	for _, val := range couponData.Array() {
		if gjson.Get(val.String(), "numRedeemed").String() != "1" {
			return fmt.Errorf("get numRedeemed from Coupon Data after redemption Expected : %s, But Got : %s", "1", gjson.Get(val.String(), "numRedeemed").String())
		}

	}
	return nil
}

// Compute Util

func Create_Vm_Instance(cloudAccId string, vm_type string, userToken string, expected_create_response int) (error, string, bool) {
	computeUrl := utils.Get_Compute_Base_Url()

	logger.Log.Info("Compute Url" + computeUrl)

	// Create an ssh key  for the user

	fmt.Println("Starting the SSH-Public-Key Creation via API...")
	// form the endpoint and payload
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/" + "sshpublickeys"
	sshPublicKey := utils.GetSSHKey()
	fmt.Println("SSH key is" + sshPublicKey)
	sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
	fmt.Println("SSH  end point ", ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)

	if sshkey_creation_status != 200 {
		return fmt.Errorf("failed to create ssh key with payload: %s, But Got Response : %s", ssh_publickey_payload, sshkey_creation_body), "", false
	}

	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()

	if sshkey_name != ssh_publickey_name_created {
		return fmt.Errorf("failed to validate ssh key name Expected: %s, But Got Response : %s", sshkey_name, ssh_publickey_name_created), "", false
	}

	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/vnets"
	vnet_name := compute_utils.GetVnetName()
	vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
	// hit the api
	fmt.Println("Vnet end point ", vnet_endpoint)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
	vnet_created := gjson.Get(vnet_creation_body, "metadata.name").String()
	if vnet_creation_status != 200 {
		return fmt.Errorf("failed to create vnet with payload: %s, But Got Response : %s", vnet_payload, vnet_creation_body), "", false
	}

	fmt.Println("Starting the Instance Creation via API...")
	// form the endpoint and payload
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	vm_name := "fnf-auto-vm" + utils.GenerateSSHKeyName(5)
	logger.Log.Info("Automation VM Name" + vm_name)
	imageName := "ubuntu-2204-jammy-v20230122"
	if vm_type == "bm-spr" {
		imageName = "ubuntu-22.04-spr-metal-cloudimg-amd64-v20240115"
	}
	if vm_type == "bm-icp-gaudi2" {
		imageName = "ubuntu-20.04-gaudi-metal-cloudimg-amd64-v20231013"
	}
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, vm_type, imageName, ssh_publickey_name_created, vnet_created)
	fmt.Println("instance_payload", instance_payload)

	// hit the api
	create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	logger.Log.Info("create_response_body" + create_response_body)

	if create_response_status == 429 || create_response_status == 403 {
		if create_response_status == 403 {
			message := gjson.Get(create_response_body, "message").String()
			if message == "product not found" {
				return fmt.Errorf("skipping test as product not found , But Create Instance  Response : %s", create_response_body), "", true
			}
		} else {
			return fmt.Errorf("skipping test as instances are in high demand , But Create Instance  Response : %s", create_response_body), "", true
		}

	}

	if create_response_status != expected_create_response {
		return fmt.Errorf("failed to create instance , Create Instance Response Expected: %d, but got : %d", expected_create_response, create_response_status), create_response_body, false
	}

	if create_response_status != 200 {
		if create_response_status != expected_create_response {
			return fmt.Errorf("failed to create instance , Create Instance Response Expected: %d, but got : %d", expected_create_response, create_response_status), create_response_body, false
		} else {
			return nil, create_response_body, false
		}
	}

	VMName := gjson.Get(create_response_body, "metadata.name").String()

	if VMName != vm_name {
		return fmt.Errorf("failed validate vm name , Expected : %s, But Got : %s ", vm_name, VMName), create_response_body, false
	}

	startTime := time.Now()
	// Validate instance powered up

	instance_id_created := gjson.Get(create_response_body, "metadata.resourceId").String()
	for counter := 0; counter <= 18; counter++ {
		get_response_byid_status, get_response_byid_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		if get_response_byid_status != 200 {
			return fmt.Errorf("failed to get instance by id : %s", get_response_byid_body), create_response_body, true
		}
		instancePhase := gjson.Get(get_response_byid_body, "status.phase").String()
		logger.Logf.Info("instancePhase: ", instancePhase)
		if instancePhase != "Ready" {
			if instancePhase == "Failed" {
				logger.Logf.Info("Instance is in Failed state")
				return fmt.Errorf("vm is in failed state, hence skipping test,  vm name  : %s, Expected State: Ready, But Got vm state: %s ", vm_name, instancePhase), create_response_body, true
			}

			if counter > 59 {
				return fmt.Errorf("vm is not in ready state after timeout, hence skipping test,  vm name : %s, Expected State: Ready, But Got vm state: %s ", vm_name, instancePhase), create_response_body, true
			}

			logger.Logf.Info("Instance is not in ready state")

		} else {
			logger.Logf.Info("Instance is in ready state")
			elapsedTime := time.Since(startTime)
			logger.Logf.Info("Time took for instance to get to ready state: ", elapsedTime)
			return nil, create_response_body, false
		}
		time.Sleep(10 * time.Second)
	}

	return nil, create_response_body, false
}

func ValidateUsage(cloudAccId string, expectedUsageAmt float64, rateExpected float64, instanceType string, authToken string) error {
	base_url := utils.Get_Base_Url1()
	var err_str string = ""
	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, userToken)
	if usage_response_status != 200 {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected ResponseCode: %d, Actual ResponseCode: %d ", cloudAccId, 200, usage_response_status)

	}

	logger.Logf.Info("usage_response_body: ", usage_response_body)
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")

	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "productType").String() == instanceType {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			if actualAMount < expectedUsageAmt || actualAMount != expectedUsageAmt {
				s := strconv.FormatFloat(expectedUsageAmt, 'f', -1, 64)
				v := strconv.FormatFloat(actualAMount, 'f', -1, 64)
				err_str = "failed validate usage for cloud account " + cloudAccId + ", Expected : " + s + ", Actual : actualAMount " + v
			}

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			if rateFactor != rateExpected {
				s := strconv.FormatFloat(rateExpected, 'f', -1, 64)
				v := strconv.FormatFloat(rateFactor, 'f', -1, 64)
				err_str = "failed validate rate for cloud account " + cloudAccId + ", Expected : " + s + ", Actual : actualAMount " + v
			}

		}
		return true // keep iterating
	})
	if err_str != "" {
		return fmt.Errorf(err_str)
	}
	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	if total_amount_from_response < expectedUsageAmt {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected : %f, Actual : %f ", cloudAccId, expectedUsageAmt, total_amount_from_response)
	}
	return nil
}

func ValidateUsageDateRange(cloudAccId string, startDate string, endDate string, expectedUsageAmt float64, rateExpected float64, instanceType string, authToken string) error {
	base_url := utils.Get_Base_Url1()
	var err_str string = ""
	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId + "&searchStart=" + startDate + "&searchEnd=" + endDate
	logger.Logf.Info("usage_url : ", usage_url)
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, userToken)
	logger.Logf.Info("usage_response_body: ", usage_response_body)

	if usage_response_status != 200 {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected ResponseCode: %d, Actual ResponseCode: %d ", cloudAccId, 200, usage_response_status)

	}

	logger.Logf.Info("usage_response_body: ", usage_response_body)
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")

	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "productType").String() == instanceType {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)

			if expectedUsageAmt == float64(0) && actualAMount != expectedUsageAmt {
				s := strconv.FormatFloat(expectedUsageAmt, 'f', -1, 64)
				v := strconv.FormatFloat(actualAMount, 'f', -1, 64)
				err_str = "failed validate usage for cloud account " + cloudAccId + ", Expected : " + s + ", Actual : actualAMount " + v

			}
			if actualAMount < expectedUsageAmt || (actualAMount == float64(0) && expectedUsageAmt != float64(0)) {
				s := strconv.FormatFloat(expectedUsageAmt, 'f', -1, 64)
				v := strconv.FormatFloat(actualAMount, 'f', -1, 64)
				err_str = "failed validate usage for cloud account " + cloudAccId + ", Expected : " + s + ", Actual : actualAMount " + v
			}

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			if rateFactor != rateExpected {
				s := strconv.FormatFloat(rateExpected, 'f', -1, 64)
				v := strconv.FormatFloat(rateFactor, 'f', -1, 64)
				err_str = "failed validate rate for cloud account " + cloudAccId + ", Expected : " + s + ", Actual : actualAMount " + v
			}

		}
		return true // keep iterating
	})
	if err_str != "" {
		return fmt.Errorf(err_str)
	}
	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	if total_amount_from_response < expectedUsageAmt || total_amount_from_response == float64(0) {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected : %f, Actual : %f ", cloudAccId, expectedUsageAmt, total_amount_from_response)
	}
	return nil
}

func ValidateZeroUsage(cloudAccId string, rateExpected float64, instanceType string, authToken string) error {
	base_url := utils.Get_Base_Url1()
	var err_str string = ""
	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, userToken)
	if usage_response_status != 200 {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected ResponseCode: %d, Actual ResponseCode: %d ", cloudAccId, 200, usage_response_status)

	}

	logger.Logf.Info("usage_response_body: ", usage_response_body)
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")

	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "productType").String() == instanceType {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			if actualAMount != float64(0) {
				v := strconv.FormatFloat(actualAMount, 'f', -1, 64)
				err_str = "failed validate usage for cloud account " + cloudAccId + ", Expected :  0 " + ", Actual : actualAMount " + v
			}

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			if rateFactor != rateExpected {
				s := strconv.FormatFloat(rateExpected, 'f', -1, 64)
				v := strconv.FormatFloat(rateFactor, 'f', -1, 64)
				err_str = "failed validate rate for cloud account " + cloudAccId + ", Expected : " + s + ", Actual : actualAMount " + v
			}

		}
		return true // keep iterating
	})
	if err_str != "" {
		return fmt.Errorf(err_str)
	}
	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	if total_amount_from_response != float64(0) {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected : 0, Actual : %f ", cloudAccId, total_amount_from_response)
	}
	return nil
}
func ValidateUsageNotZero(cloudAccId string, rateExpected float64, instanceType string, authToken string) error {
	base_url := utils.Get_Base_Url1()
	var err_str string = ""
	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, userToken)
	if usage_response_status != 200 {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected ResponseCode: %d, Actual ResponseCode: %d ", cloudAccId, 200, usage_response_status)

	}

	logger.Logf.Info("usage_response_body: ", usage_response_body)
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")

	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		logger.Logf.Infof("Usage Data : %s ", data)
		if gjson.Get(data, "productType").String() == instanceType {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			if actualAMount == float64(0) {
				v := strconv.FormatFloat(actualAMount, 'f', -1, 64)
				err_str = "failed validate usage for cloud account " + cloudAccId + ", Expected : >  0 " + ", Actual : actualAMount " + v
			}

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			if rateFactor != rateExpected {
				s := strconv.FormatFloat(rateExpected, 'f', -1, 64)
				v := strconv.FormatFloat(rateFactor, 'f', -1, 64)
				err_str = "failed validate rate for cloud account " + cloudAccId + ", Expected : " + s + ", Actual : actualAMount " + v
			}

		}
		return true // keep iterating
	})
	if err_str != "" {
		return fmt.Errorf(err_str)
	}
	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	if total_amount_from_response == float64(0) {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected > 0, Actual : %f ", cloudAccId, total_amount_from_response)
	}
	return nil
}

func ValidateUsageinRange(cloudAccId string, rateExpected float64, instanceType string, authToken string, minVal float64, maxVal float64) error {
	base_url := utils.Get_Base_Url1()
	var err_str string = ""
	usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + cloudAccId
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	usage_response_status, usage_response_body := financials.GetUsage(usage_url, userToken)
	if usage_response_status != 200 {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected ResponseCode: %d, Actual ResponseCode: %d ", cloudAccId, 200, usage_response_status)

	}

	logger.Logf.Info("usage_response_body: ", usage_response_body)
	result := gjson.Parse(usage_response_body)
	arr := gjson.Get(result.String(), "usages")

	arr.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		if gjson.Get(data, "productType").String() == instanceType {
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			if actualAMount < minVal || actualAMount > maxVal {
				v := strconv.FormatFloat(actualAMount, 'f', -1, 64)
				mi := strconv.FormatFloat(minVal, 'f', -1, 64)
				ma := strconv.FormatFloat(maxVal, 'f', -1, 64)
				err_str = "failed validate usage for cloud account " + cloudAccId + ", Expected : in Range min : " + mi + " and max :  " + ma + ", Actual : actualAMount " + v
			}

			rate := gjson.Get(data, "rate").String()
			rateFactor, _ := strconv.ParseFloat(rate, 64)
			if rateFactor != rateExpected {
				s := strconv.FormatFloat(rateExpected, 'f', -1, 64)
				v := strconv.FormatFloat(rateFactor, 'f', -1, 64)
				err_str = "failed validate rate for cloud account " + cloudAccId + ", Expected : " + s + ", Actual : actualAMount " + v
			}

		}
		return true // keep iterating
	})
	if err_str != "" {
		return fmt.Errorf(err_str)
	}
	total_amount_from_response := gjson.Get(usage_response_body, "totalAmount").Float()
	if total_amount_from_response == float64(0) {
		return fmt.Errorf("failed validate usage for cloud account %s, Expected : 0, Actual : %f ", cloudAccId, total_amount_from_response)
	}
	return nil
}

func ValidateCreditsNonZeroDepletion(cloudAccId string, redeemAmount float64, authToken string) error {
	base_url := utils.Get_Credits_Base_Url() + "/credit"
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	response_status, responseBody := financials.GetCredits(base_url, userToken, cloudAccId)
	if response_status != 200 {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected ResponseCode: %d, Actual ResponseCode: %d ", cloudAccId, 200, response_status)

	}
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	//usedAmount = testsetup.RoundFloat(usedAmount, 0)
	if usedAmount == float64(0) {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected UsedCredits: > 0, Actual UsedCredits: %f ", cloudAccId, usedAmount)

	}

	if remainingAmount == redeemAmount {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected Remaining Amount: < %f, Actual Remaining Amount: %f ", cloudAccId, redeemAmount, remainingAmount)

	}

	res := GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	origVal := unappliedCredits
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 1)
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	if unappliedCredits == redeemAmount {
		if origVal == redeemAmount {
			return fmt.Errorf("failed validate unapplied credits for cloud account %s, Expected Unapplied Credits:<  %f, Actual Unapplied Credits: %f ", cloudAccId, redeemAmount, unappliedCredits)
		}

	}
	return nil
}

func ValidateUsageCreditsinRange(cloudAccId string, expectedMinUsedCreditAmt float64, expectedMaxUsedCreditAmt float64, authToken string, expectedMinRemainingAmt float64, expectedMaxRemainingAmt float64, expectedMinUnappliedCredits float64, expectedMaxUnappliedCredits float64, greaterThan float64) error {
	base_url := utils.Get_Credits_Base_Url() + "/credit"
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	response_status, responseBody := financials.GetCredits(base_url, userToken, cloudAccId)
	if response_status != 200 {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected ResponseCode: %d, Actual ResponseCode: %d ", cloudAccId, 200, response_status)

	}
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	if usedAmount <= greaterThan {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected UsedCredits: %f, Actual UsedCredits: %f ", cloudAccId, expectedMaxUsedCreditAmt, usedAmount)
	}
	if usedAmount < expectedMinUsedCreditAmt || usedAmount > expectedMaxUsedCreditAmt {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected UsedCredits: %f, Actual UsedCredits: %f ", cloudAccId, expectedMaxUsedCreditAmt, usedAmount)

	}

	if remainingAmount < expectedMinRemainingAmt || remainingAmount > expectedMaxRemainingAmt {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected UsedCredits: %f, Actual UsedCredits: %f ", cloudAccId, expectedMaxRemainingAmt, remainingAmount)

	}

	res := GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 1)
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	if unappliedCredits < expectedMinUnappliedCredits || unappliedCredits > expectedMaxUnappliedCredits {
		return fmt.Errorf("failed validate unapplied credits for cloud account %s, Expected UsedCredits: %f, Actual UsedCredits: %f ", cloudAccId, expectedMaxUnappliedCredits, unappliedCredits)

	}
	return nil
}

func ValidateCredits(cloudAccId string, expectedUsedCreditAmt float64, authToken string, expectedRemainingAmt float64, expectedUnappliedCredits float64, greaterThan float64) error {
	base_url := utils.Get_Credits_Base_Url() + "/credit"
	userName, _ := testsetup.GetCloudAccountUserName(cloudAccId, utils.Get_Base_Url1(), authToken)
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	response_status, responseBody := financials.GetCredits(base_url, userToken, cloudAccId)
	logger.Logf.Info("Credit Response for cloud Account %s is %s: ", cloudAccId, responseBody)
	if response_status != 200 {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected ResponseCode: %d, Actual ResponseCode: %d ", cloudAccId, 200, response_status)

	}
	usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
	remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
	usedAmount = testsetup.RoundFloat(usedAmount, 0)
	if !(usedAmount >= (expectedUsedCreditAmt-0.5) && usedAmount <= expectedUsedCreditAmt+0.5) {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected UsedCredits: %f, Actual UsedCredits: %f ", cloudAccId, expectedUsedCreditAmt, usedAmount)

	}

	if !(remainingAmount >= (expectedRemainingAmt-0.5) && remainingAmount <= expectedRemainingAmt+0.5) || remainingAmount < greaterThan {
		return fmt.Errorf("failed validate credits for cloud account %s, Expected Remaining Amount: %f, Actual Remaining Amount: %f ", cloudAccId, expectedRemainingAmt, remainingAmount)

	}

	res := GetUnappliedCloudCreditsNegative(cloudAccId, 200)
	unappliedCredits := gjson.Get(res, "unappliedAmount").Float()
	unappliedCredits = testsetup.RoundFloat(unappliedCredits, 1)
	logger.Logf.Info("Unapplied credits After credits1 coupon: ", unappliedCredits)
	if !(unappliedCredits >= (expectedUnappliedCredits-0.5) && unappliedCredits <= expectedUnappliedCredits+0.5) {
		return fmt.Errorf("failed validate unapplied credits for cloud account %s, Expected Unapplied Credits: %f, Actual Unapplied credits: %f ", cloudAccId, expectedUnappliedCredits, unappliedCredits)

	}
	return nil
}

func GetCAccById(cloudAccount_id string, expected_status_code int) (string, string) {
	url := utils.Get_CA_Base_Url() + "/v1/cloudaccounts/id/" + cloudAccount_id
	jsonStr, _ := http_client.Get(url, expected_status_code)
	return "True", jsonStr
}

func Upgrade_to_Premium_with_coupon(cloudAccId string, authToken string, userToken string, couponAmount int64) error {
	ret_value1, responsePayload := GetCAccById(cloudAccId, 200)
	if ret_value1 != "True" {
		return fmt.Errorf("failed to get cloud account details by id : %s", cloudAccId)
	}

	if gjson.Get(responsePayload, "type").String() != "ACCOUNT_TYPE_STANDARD" {
		return fmt.Errorf("account type for cloud account %s is not ACCOUNT_TYPE_STANDARD : ", cloudAccId)
	}

	if gjson.Get(responsePayload, "upgradedToPremium").String() != "UPGRADE_NOT_INITIATED" {
		return fmt.Errorf("validation failed on cloud account attribute : upgradedToPremium, expetced : UPGRADE_NOT_INITIATED, actual :%s ", gjson.Get(responsePayload, "upgradedToPremium").String())
	}

	if gjson.Get(responsePayload, "upgradedToEnterprise").String() != "UPGRADE_NOT_INITIATED" {
		return fmt.Errorf("validation failed on cloud account attribute : upgradedToEnterprise, expetced : UPGRADE_NOT_INITIATED, actual :%s ", gjson.Get(responsePayload, "upgradedToEnterprise").String())
	}

	// Upgrade standard account using coupon
	upgrade_err := Standard_to_premium_upgrade_with_coupon("Premium", couponAmount, 2, cloudAccId, userToken, "ACCOUNT_TYPE_PREMIUM")

	if upgrade_err != nil {
		logger.Logf.Infof("upgrade from standard to premium with coupon failed, with error : ", upgrade_err)
		return upgrade_err
	}

	migrate_err := Credit_Migrate(cloudAccId, authToken)

	if migrate_err != nil {
		logger.Logf.Infof("upgrade from standard to premium with coupon failed in credit migration , with error : ", migrate_err)
		return migrate_err
	}

	// Check cloud account attributes after upgrade

	ret_value1, responsePayload = GetCAccById(cloudAccId, 200)

	if ret_value1 != "True" {
		return fmt.Errorf("failed to get cloud account details by id : %s", cloudAccId)
	}

	if gjson.Get(responsePayload, "type").String() != "ACCOUNT_TYPE_PREMIUM" {
		return fmt.Errorf("account type for cloud account %s is not ACCOUNT_TYPE_PREMIUM : ", cloudAccId)
	}

	if gjson.Get(responsePayload, "upgradedToPremium").String() != "UPGRADE_COMPLETE" {
		return fmt.Errorf("validation failed on cloud account attribute : upgradedToPremium, expetced : UPGRADE_COMPLETE, actual :%s ", gjson.Get(responsePayload, "upgradedToPremium").String())
	}

	if gjson.Get(responsePayload, "upgradedToEnterprise").String() != "UPGRADE_NOT_INITIATED" {
		return fmt.Errorf("validation failed on cloud account attribute : upgradedToEnterprise, expetced : UPGRADE_NOT_INITIATED, actual :%s ", gjson.Get(responsePayload, "upgradedToEnterprise").String())
	}

	if gjson.Get(responsePayload, "paidServicesAllowed").String() != "true" {
		return fmt.Errorf("validation failed on cloud account attribute : paidServicesAllowed, expetced : true, actual :%s ", gjson.Get(responsePayload, "paidServicesAllowed").String())
	}

	if gjson.Get(responsePayload, "lowCredits").String() != "false" {
		return fmt.Errorf("validation failed on cloud account attribute : lowCredits, expetced : false, actual :%s ", gjson.Get(responsePayload, "lowCredits").String())
	}

	if gjson.Get(responsePayload, "terminatePaidServices").String() != "false" {
		return fmt.Errorf("validation failed on cloud account attribute : terminatePaidServices, expetced : false, actual :%s ", gjson.Get(responsePayload, "terminatePaidServices").String())
	}

	return nil

}

func CreateCouponToRedeemExpiry(userType string, couponAmount int64, no_of_user int64, cloudAccId string, token string, expiration time.Duration) (string, error) {
	creation_time, expirationtime := ExpireCoupon(expiration)
	var isStandard bool = false
	//isStandard = false
	if userType == "Standard" {
		isStandard = true
	}
	createCoupon := StandardCreateCouponStruct{
		Amount:     couponAmount,
		Creator:    "idc_billing@intel.com",
		Expires:    expirationtime,
		Start:      creation_time,
		NumUses:    no_of_user,
		IsStandard: isStandard,
	}

	jsonPayload, _ := json.Marshal(createCoupon)
	req := []byte(jsonPayload)
	ret_value, data := CreateCoupon(req, 200)
	if !ret_value {
		return "", fmt.Errorf("create Coupon Failed")
	}
	couponCode := gjson.Get(data, "code").String()
	if createCoupon.Amount != gjson.Get(data, "amount").Int() {
		return "", fmt.Errorf("expected Amount from coupon Creation : %d, But Got : %d", createCoupon.Amount, gjson.Get(data, "amount").Int())
	}

	if createCoupon.Creator != gjson.Get(data, "creator").String() {
		return "", fmt.Errorf("expected Coupon Creator Name  : %s, But Got : %s", createCoupon.Creator, gjson.Get(data, "creator").String())
	}

	if createCoupon.Expires != gjson.Get(data, "expires").String() {
		return "", fmt.Errorf("expected Coupon Expiry Time  : %s, But Got : %s", createCoupon.Expires, gjson.Get(data, "expires").String())
	}

	if createCoupon.NumUses != gjson.Get(data, "numUses").Int() {
		return "", fmt.Errorf("expected Coupon Number of Uses : %d, But Got : %d", createCoupon.NumUses, gjson.Get(data, "numUses").Int())
	}

	if createCoupon.Start != gjson.Get(data, "start").String() {
		return "", fmt.Errorf("expected Coupon Start date : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
	}

	// Get coupon and validate
	getret_value, getdata := GetCoupons(couponCode, 200)
	if !getret_value {
		return "", fmt.Errorf("get on Coupons Failed")
	}
	couponData := gjson.Get(getdata, "coupons")
	for _, val := range couponData.Array() {
		if createCoupon.Amount != gjson.Get(val.String(), "amount").Int() {
			return couponCode, fmt.Errorf("expected Amount from Get coupon Creation Time : %d, But Got : %d", createCoupon.Amount, gjson.Get(data, "amount").Int())
		}

		if createCoupon.Creator != gjson.Get(val.String(), "creator").String() {
			return couponCode, fmt.Errorf("expected Coupon Get Creator Name  : %s, But Got : %s", createCoupon.Creator, gjson.Get(data, "creator").String())
		}

		if createCoupon.Expires != gjson.Get(val.String(), "expires").String() {
			return couponCode, fmt.Errorf("expected Get Coupon Expiry Time  : %s, But Got : %s", createCoupon.Expires, gjson.Get(data, "expires").String())
		}

		if createCoupon.NumUses != gjson.Get(val.String(), "numUses").Int() {
			return couponCode, fmt.Errorf("expected Get Coupon Number of Uses : %d, But Got : %d", createCoupon.NumUses, gjson.Get(data, "numUses").Int())
		}

		if createCoupon.Start != gjson.Get(val.String(), "start").String() {
			return couponCode, fmt.Errorf("expected Get Coupon Start date : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
		}

		if gjson.Get(val.String(), "numRedeemed").String() != "0" {
			return couponCode, fmt.Errorf("expected Get numRedeemed : %s, But Got : %s", createCoupon.Start, gjson.Get(data, "start").String())
		}

	}

	return couponCode, nil

}

func Upgrade_Coupon_With_Shrt_Expiry(userType string, couponAmount int64, no_of_user int64, cloudAccId string, token string, cloudAccountUpgradeToType string, expiration time.Duration) error {
	couponCode, err := CreateCouponToRedeemExpiry(userType, couponAmount, no_of_user, cloudAccId, token, expiration)
	if err != nil {
		return err
	}
	//Redeem coupon to upgrade to premium
	logger.Log.Info("Starting Billing Test : Redeem coupon for Standard cloud account")
	time.Sleep(1 * time.Minute)
	base_url := utils.Get_Base_Url1() + "/v1/cloudaccounts/upgrade"
	Coupon_api_payload := financials_utils.EnrichUpgradeCouponPayload(financials_utils.GetUpgradeCouponPayload(), cloudAccId, cloudAccountUpgradeToType, couponCode)
	logger.Logf.Infof("Upgrade coupon payload", Coupon_api_payload)
	response_status, responseBody := financials.UpgradeWithCoupon(base_url, token, Coupon_api_payload)
	if response_status != 200 {
		return fmt.Errorf("failed to upgrade cloud account : %s, error : %s", cloudAccId, responseBody)
	}

	// Get coupon and validate
	_, getdata := GetCoupons(couponCode, 200)
	couponData := gjson.Get(getdata, "coupons")
	redemptions := gjson.Get(getdata, "result.redemptions")
	for _, val := range redemptions.Array() {
		if cloudAccId != gjson.Get(val.String(), "cloudAccountId").String() {
			return fmt.Errorf("get Cloud Acc Id from Coupon Data after redemption Expected : %s, But Got : %s", cloudAccId, gjson.Get(val.String(), "cloudAccountId").String())
		}

		if couponCode != gjson.Get(val.String(), "code").String() {
			return fmt.Errorf("get couponCode from Coupon Data after redemption Expected : %s, But Got : %s", couponCode, gjson.Get(val.String(), "code").String())
		}

		if gjson.Get(val.String(), "installed").String() != "true" {
			return fmt.Errorf("get installed from Coupon Data after redemption Expected : %s, But Got : %s", "true", gjson.Get(val.String(), "installed").String())
		}

	}
	for _, val := range couponData.Array() {
		if gjson.Get(val.String(), "numRedeemed").String() != "1" {
			return fmt.Errorf("get numRedeemed from Coupon Data after redemption Expected : %s, But Got : %s", "1", gjson.Get(val.String(), "numRedeemed").String())
		}

	}
	return nil
}

func CreateOTP(adminAccountId string, memberEmail string, admintoken string) (string, error) {
	create_otp_url := utils.Get_Base_Url1() + "/v1/otp/create"
	create_otp_payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"memberEmail": "%s"
	}`, adminAccountId, memberEmail)
	logger.Logf.Infof("Create OTP url :%s", create_otp_url)
	logger.Logf.Infof("Create OTP Payload :%s", create_otp_payload)
	responseCode, responseBody := financials.CreateOTP(create_otp_url, admintoken, create_otp_payload)
	if responseCode != 200 {
		return responseBody, fmt.Errorf("create otp failed for admin account : %s, with response : %s", adminAccountId, responseBody)
	}
	return responseBody, nil
}

func GetOTP(adminAccountId string, memberEmail string) (string, string) {
	pg_password := financials_utils.GetCloudAccDBPAssword()
	otpCode, err := financials.GetOTP(adminAccountId, memberEmail, pg_password)
	return otpCode, err

}

func DeleteOTP(adminAccountId string, memberEmail string) (string, string) {
	pg_password := financials_utils.GetCloudAccDBPAssword()
	otpCode, err := financials.GetOTP(adminAccountId, memberEmail, pg_password)
	return otpCode, err

}

func VerifyOTP(adminAccountId string, memberEmail string, otpCode string, admintoken string) (string, error) {
	verify_otp_url := utils.Get_Base_Url1()
	verify_otp_payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
		"memberEmail": "%s",
		"otpCode": "%s"
	}`, adminAccountId, memberEmail, otpCode)
	logger.Logf.Infof("Verify OTP url :%s", verify_otp_url)
	logger.Logf.Infof("Verify OTP Payload :%s", verify_otp_payload)

	rescode, response := financials.VerifyOTP(verify_otp_url, admintoken, verify_otp_payload)
	if rescode != 200 {
		return response, fmt.Errorf("create OTP failed with error code : %d ", rescode)
	}
	logger.Logf.Infof("Verify OTP response :%s", rescode)
	logger.Logf.Infof("Verify OTP response :%s", response)
	return response, nil

}

func SendInviteParallel(memberEmails []string) {
	type SendInvite struct {
		CloudAccountID string `json:"cloudAccountId"`
		Invites        []struct {
			Expiry      string `json:"expiry"`
			MemberEmail string `json:"memberEmail"`
			Note        string `json:"note"`
		} `json:"invites"`
	}

	Accounts := utils.Get_numAccounts()
	numAccounts := int(Accounts)
	const interval = 3 * time.Minute // Interval time
	var group sync.WaitGroup
	for i := 1; i <= numAccounts; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()

		}(i)
	}
	group.Wait()
}

func SendInvite(adminAccountId string, memberEmail string, admintoken string, membertoken string, idcAdminToken string) (string, error) {
	invite_url := utils.Get_Base_Url1() + "/v1/cloudaccounts/invitations/create"
	base_url := utils.Get_Base_Url1()
	code, body := financials.CreateInviteCode(invite_url, admintoken, adminAccountId, memberEmail)
	logger.Logf.Infof("SendInvite response code  :%s", code)
	logger.Logf.Infof("SendInvite response body :%s", body)
	if code != 200 {
		return body, fmt.Errorf("error creating invitation for member : %s", memberEmail)
	}

	// Send invite code to member

	payload := fmt.Sprintf(`{
		"adminAccountId": "%s",
		"memberEmail": "%s"
	}`, adminAccountId, memberEmail)

	code, body = financials.SendInviteCode(base_url, membertoken, payload)
	if code != 200 {
		return body, fmt.Errorf("error creating invitation code for member : %s", memberEmail)
	}

	// Verify Member is part of admins members list

	return "", nil

}

func GetInviteCode(adminAccountId string, memberEmail string) (string, error) {
	// Retrieve OTP Code for Invited Member
	pg_password := financials_utils.GetCloudAccDBPAssword()
	time.Sleep(20 * time.Second)
	inviteCode, errStr := financials.GetInviteCode(adminAccountId, memberEmail, pg_password)
	if errStr != "" {
		return "", fmt.Errorf("failed to create invite, OTP code not found")
	}
	if inviteCode == "" {
		return "", fmt.Errorf("failed to create invite, OTP code not found")
	}

	return inviteCode, nil
}

func VerifyInviteCode(adminAccountId string, memberEmail string, inviteCode string, membertoken string) (string, error) {
	// Verify OTP for Member and join to group
	base_url := utils.Get_Base_Url1()
	payload := fmt.Sprintf(`{
		"adminCloudAccountId": "%s",
		"inviteCode": "%s",
		"memberEmail": "%s"
	}`, adminAccountId, inviteCode, memberEmail)
	fmt.Println("Payload: ", payload)
	code, body := financials.VerifyInviteCode(base_url, membertoken, payload)
	logger.Logf.Infof("VerifyInviteCode response code  :%s", code)
	logger.Logf.Infof("VerifyInviteCode response body :%s", body)
	if code != 200 {
		return body, fmt.Errorf("failed to verify invite code: %s", memberEmail)
	}
	return body, nil
}

func ReadInvitations(adminAccountId string, memberEmail string, admintoken string) (string, error) {
	base_url := utils.Get_Base_Url1()
	code, body := financials.ReadInvitations(base_url, admintoken, adminAccountId, "")
	logger.Logf.Infof("ReadInvitations response code  :%s", code)
	logger.Logf.Infof("ReadInvitations response body :%s", body)
	if code != 200 {
		return body, fmt.Errorf("failed to read invitations for admin cloudaccount: %s, with out put :%s", adminAccountId, body)
	}
	invitation_found := false
	result := gjson.Get(body, "invites")
	result.ForEach(func(key, value gjson.Result) bool {
		data := value.String()
		member_email := gjson.Get(data, "member_email").String()
		if member_email == memberEmail {
			if gjson.Get(data, "invitation_state").String() == "INVITE_STATE_ACCEPTED" {
				invitation_found = true
			}

			return false
		}

		return true // keep iterating
	})

	if !invitation_found {
		return "", fmt.Errorf("invitation not found for member : %s", memberEmail)
	}

	return body, nil
}

func AddMember(cloudAccId string, memberEmail string, userToken string, membertoken string, authToken string) error {
	response, err := CreateOTP(cloudAccId, memberEmail, userToken)
	logger.Logf.Infof("Create OTP response : %s", response)
	if err != nil {
		return fmt.Errorf("OTP create failed")
	}

	otpCode, errString := GetOTP(cloudAccId, memberEmail)

	if errString != "" {
		return fmt.Errorf("OTP Verification Failed")
	}

	_, err = VerifyOTP(cloudAccId, memberEmail, otpCode, userToken)
	if err != nil {
		return fmt.Errorf("OTP Verification failed")
	}

	out1, err1 := SendInvite(cloudAccId, memberEmail, userToken, membertoken, authToken)
	if err1 != nil {
		return fmt.Errorf("SendInvite failed")
	}
	logger.Logf.Infof("Response of Send Invite : %s", out1)

	inviteCode, err3 := GetInviteCode(cloudAccId, memberEmail)
	if err3 != nil {
		return fmt.Errorf("GetInviteCode failed")
	}

	_, err4 := VerifyInviteCode(cloudAccId, memberEmail, inviteCode, membertoken)
	if err4 != nil {
		return fmt.Errorf("VerifyInviteCode failed")
	}

	res, err5 := ReadInvitations(cloudAccId, memberEmail, userToken)
	if err5 != nil {
		return fmt.Errorf("VerifyInviteCode failed")
	}
	logger.Logf.Infof("Response of read invitations : %s", res)

	return nil
}

func RemoveMember(cloudAccId string, memberEmail string, userToken string, membertoken string, authToken string) (string, error) {
	base_url := utils.Get_Base_Url1()
	code, body := financials.RemoveInvitation(base_url, userToken, cloudAccId, memberEmail)
	logger.Logf.Infof("RemoveMember response code  :%s", code)
	logger.Logf.Infof("RemoveMember response body :%s", body)
	if code != 200 {
		return body, fmt.Errorf("remove invitation failed for admin account :%s, Member email :%s, error :%s", cloudAccId, memberEmail, body)
	}

	return body, nil

}

// func CreateMaasUsageRecords(payload MaasCreateUsage, authToken string, expected_response int) (string, error) {
// 	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
// 	jsonBytes, _ := json.Marshal(payload)
// 	code, body := financials.CreateMaasUsageRecords(base_url, authToken, string(jsonBytes))
// 	if expected_response != code {
// 		return body, fmt.Errorf("response code did not match while creating Maas Usage record, expected : %d, Got : %d", expected_response, code)
// 	}

// 	return body, nil

// }

func CreateMaasUsageRecords(cloudAccId string, endTime string, processingType string, quantity int, region string, startTime string, timeStamp string, transactionid string, authToken string, expected_response int) (string, error) {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products"
	payload := fmt.Sprintf(`{
		"cloudAccountId": "%s",
        "endTime": "%s",
        "properties":  {
        "serviceType": "ModelAsAService",
        "processingType": "%s"
        },
      "quantity": %d,
      "region": "%s",
      "startTime": "%s",
      "timestamp": "%s",
      "transactionId": "%s"
	    }`, cloudAccId, endTime, processingType, quantity, region, startTime, timeStamp, transactionid)
	code, body := financials.CreateMaasUsageRecords(base_url, authToken, payload)
	logger.Logf.Infof("Maas Url : ", base_url)
	logger.Logf.Infof("Maas Usage Payload : ", payload)
	if expected_response != code {
		return body, fmt.Errorf("response code did not match while creating Maas Usage record, expected : %d, Got : %d", expected_response, code)
	}

	return body, nil

}

func SearchMaasUsageRecords(payload string, authToken string, expected_response int) (string, error) {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products/search"
	//jsonBytes, _ := json.Marshal(payload)
	code, body := financials.SearchMaasUsageRecords(base_url, authToken, payload)
	if expected_response != code {
		return body, fmt.Errorf("response code did not match while searching Maas Usage record, expected : %d, Got : %d", expected_response, code)
	}

	return body, nil

}

func SearchInvalidMaasUsageRecords(payload string, authToken string, expected_response int) (string, error) {
	base_url := utils.Get_Base_Url1() + "/v1/usages/records/products/search"
	//jsonBytes, _ := json.Marshal(payload)
	code, body := financials.SearchMaasUsageRecords(base_url, authToken, payload)
	if expected_response != code {
		return body, fmt.Errorf("response code did not match while searching Maas Usage record, expected : %d, Got : %d", expected_response, code)
	}

	return body, nil

}
