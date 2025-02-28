package financials

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/frisby_client"
	"goFramework/framework/library/auth"
	"goFramework/ginkGo/financials/financials_utils"
	"strings"

	"github.com/tidwall/gjson"
)

const PREMIUM_CLOUDACCOUNT_TYPE = "ACCOUNT_TYPE_PREMIUM"

func getCloudAccount(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//CloudAccount response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createCloudAccount(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//CloudAccount response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteCloudAccount(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	//CloudAccount response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func CreateCloudAccount(CloudAccount_api_base_url string, token string, CloudAccount_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(CloudAccount_api_payload), &jsonMap)
	response_status, response_body := createCloudAccount(CloudAccount_api_base_url, token, jsonMap)
	return response_status, response_body
}

// Private Methods CREATE Premium / Standard
func createCloudAccountPremium(cloudaccount_url string, userToken string) (int, string) {
	enroll_payload := `{
		"premium":false, 
  		"termsStatus": true
	}`
	cloudaccount_enroll_url := cloudaccount_url + "/enroll"
	fmt.Println("CloudAccountEnroll", cloudaccount_enroll_url)
	cloudaccount_creation_status, cloudaccount_creation_body := CreateCloudAccount(cloudaccount_enroll_url, userToken, enroll_payload)

	return cloudaccount_creation_status, cloudaccount_creation_body
}

func createClouadAccountStandard(cloudaccount_url string, userToken string) (int, string) {
	enroll_payload := `{
		"premium":false,
		"termsStatus": true
	}`
	cloudaccount_enroll_url := cloudaccount_url + "/enroll"
	cloudaccount_creation_status, cloudaccount_creation_body := CreateCloudAccount(cloudaccount_enroll_url, userToken, enroll_payload)
	return cloudaccount_creation_status, cloudaccount_creation_body
}

func upgradeCloudAccount(url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	frisby_response := frisby_client.Post(url, token, jsonMap)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//CloudAccount response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func UpgradeCloudaccount(CloudAccount_api_base_url string, cloudaccountId string, adminToken string, accountType string, couponCode string) (int, string) {
	var upgrade_cloudaccount_url = CloudAccount_api_base_url + "/v1/cloudaccounts/upgrade"
	fmt.Println("upgrade_cloudaccount_url", upgrade_cloudaccount_url)
	upgrade_payload := fmt.Sprintf(`{
		"cloudAccountId":%s,
		"cloudAccountUpgradeToType": %s,
		"code": %s
	}`, cloudaccountId, accountType, couponCode)
	get_response_byid_status, get_response_byid_body := upgradeCloudAccount(upgrade_cloudaccount_url, adminToken, upgrade_payload)
	return get_response_byid_status, get_response_byid_body
}

func GetCloudAccountById(CloudAccount_api_base_url string, token string, CloudAccount_id string) (int, string) {
	var get_byCloudAccount_id_url = CloudAccount_api_base_url + "/id/" + CloudAccount_id
	fmt.Println("get_byCloudAccount_id_url", get_byCloudAccount_id_url)
	get_response_byid_status, get_response_byid_body := getCloudAccount(get_byCloudAccount_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetCloudAccountByName(CloudAccount_api_base_url string, token string, CloudAccount_name string) (int, string) {
	var get_byCloudAccount_name_url = CloudAccount_api_base_url + "/name/" + CloudAccount_name
	fmt.Println("GET CA URL: ", get_byCloudAccount_name_url)
	get_response_byname_status, get_response_byname_body := getCloudAccount(get_byCloudAccount_name_url, token)
	return get_response_byname_status, get_response_byname_body
}

func GetAllCloudAccount(CloudAccount_api_base_url string, token string) (int, string) {
	get_allCloudAccounts_response_status, get_allCloudAccounts_response_body := getCloudAccount(CloudAccount_api_base_url, token)
	return get_allCloudAccounts_response_status, get_allCloudAccounts_response_body
}

func DeleteCloudAccountById(CloudAccount_api_base_url string, token string, CloudAccount_id string) (int, string) {
	var delete_byCloudAccount_id_url = CloudAccount_api_base_url + "/id/" + CloudAccount_id
	fmt.Println("delete_byCloudAccount_id_url", delete_byCloudAccount_id_url)
	delete_response_byid_body, delete_response_byid_status := deleteCloudAccount(delete_byCloudAccount_id_url, token)
	return delete_response_byid_body, delete_response_byid_status
}

func DeleteCloudAccountByName(CloudAccount_api_base_url string, token string, CloudAccount_name string) (int, string) {
	var delete_byCloudAccount_name_url = CloudAccount_api_base_url + "/name/" + CloudAccount_name
	delete_response_byname_body, delete_response_byname_status := deleteCloudAccount(delete_byCloudAccount_name_url, token)
	return delete_response_byname_body, delete_response_byname_status
}

func GetEnrollToken(url string) (string, int) {
	frisby_response := frisby_client.Get(url, "")
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//CloudAccount response schema and common validation goes here - yet to be implemented
	return responseBody, responseCode
}

// CREATE CA

func CreateCloudAccountE2E(url string, adminToken string, usertoken string, userName string, premium bool) (int, string) {
	fmt.Println("Initial URL: ", url)
	cloudaccount_url := url + "/v1/cloudaccounts"
	code, body := GetCloudAccountByName(cloudaccount_url, adminToken, userName)
	create_coupon_endpoint := url + "/v1/cloudcredits/coupons"
	upgrade_url := url + "/v1/cloudaccounts/upgrade"
	cloudAccId := gjson.Get(body, "id").String()
	/*if strings.Contains(url, "staging") && cloudAccId != "" {
		return 200, body
	}*/
	if code == 200 || code == 404 {
		fmt.Println("ID: ", cloudAccId)
		if code == 404 {
			fmt.Println("Checking cloudaccount with lowercase...")
			code, body = GetCloudAccountByName(cloudaccount_url, adminToken, strings.ToLower(userName))
			cloudAccId = gjson.Get(body, "id").String()
			fmt.Println("Code: ", code)
			fmt.Println("body: ", body)
		}
		if cloudAccId != "" {
			fmt.Println("Deleting existing CloudAccount")
			code, body := DeleteCloudAccountById(cloudaccount_url, adminToken, cloudAccId)
			if code != 200 {
				panic("Error Deleting CloudAccount " + body)
			}
		}
		if premium {
			fmt.Println("Premium Flow.....", cloudAccId)
			coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 2, "idc_billing@intel.com", 1)
			fmt.Println("Token", adminToken)
			fmt.Println("Coupon Payload", coupon_payload)
			coupon_creation_status, coupon_creation_body := CreateCoupon(create_coupon_endpoint, adminToken, coupon_payload)
			fmt.Println("Coupon creation status.....", coupon_creation_status)
			fmt.Println("Coupon creation body.....", coupon_creation_body)
			couponId := gjson.Get(coupon_creation_body, "code").String()

			create_account_status, create_account_body := createCloudAccountPremium(cloudaccount_url, usertoken)

			fmt.Println("Account creation status.....", create_account_status)
			fmt.Println("Account creation body.....", create_account_body)
			fmt.Println("Coupon Id.....", couponId)
			cloud_account_created := gjson.Get(create_account_body, "cloudAccountId").String()

			upgrade_payload := fmt.Sprintf(`{
				"cloudAccountId": "%s",
				"cloudAccountUpgradeToType": "%s",
				"code": "%s"
			}`, cloud_account_created, PREMIUM_CLOUDACCOUNT_TYPE, couponId)
			fmt.Print("Upgrade Payload...", upgrade_payload)
			token_response, _ := auth.Get_Azure_Bearer_Token(userName)
			userTokenPU := "Bearer " + token_response
			return UpgradeWithCoupon(upgrade_url, userTokenPU, upgrade_payload)
		}
		fmt.Println("Creating Standard Flow.....")
		return createClouadAccountStandard(cloudaccount_url, usertoken)
	}
	return 500, body
}
