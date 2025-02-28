package PU_CredentialsAsAService_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("CredentialsAsAService", Ordered, Label("Admin endpoints"), func() {
	var base_url string
	var token string
	var userName string
	var userToken string
	var cloud_account_created string

	var ariaAuth string
	var place_holder_map = make(map[string]string)

	var ariaclientId_su string

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../data", "vmaas_input.json")
		auth.Get_config_file_data("../../data/config.json")
		financials_utils.LoadE2EConfig("../../data", "billing.json")
		userName = auth.Get_UserName("Premium")
		base_url = compute_utils.GetBaseUrl()
		compute_url = compute_utils.GetComputeUrl()
		token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
		ariaclientId_su = financials_utils.GetAriaClientNo()
		ariaAuth = financials_utils.GetariaAuthKey()
		testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("ariaclientId_su", ariaclientId_su)
	})

	It("Delete cloud account", func() {
		fmt.Println("Delete cloud account")
		url := base_url + "/v1/cloudaccounts"
		cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, token)
		if err == nil {
			financials.DeleteCloudAccountById(url, token, cloudAccId)
		}
	})

	It("Create cloud account", func() {
		// Generating token wtih payload for cloud account creation with enroll API
		// cloud account creation
		token_response, _ := auth.Get_Azure_Bearer_Token(userName)
		userToken = "Bearer " + token_response
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, true)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll, body"+cloudaccount_creation_body)
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of Premium user, Body: "+cloudaccount_creation_body)
		place_holder_map["cloud_account_id"] = cloud_account_created
		place_holder_map["cloud_account_type"] = cloudaccount_type
		fmt.Println("cloudAccount_id", cloud_account_created)
	})

	It("Get usage test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/usages?" + place_holder_map["cloud_account_id"]
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 usages. ResponseBody "+responseBody)
	})

	It("Get coupon test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/coupons"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 coupons. ResponseBody "+responseBody)
	})

	It("Post coupon test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/coupons"
		payload := `{}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to Post 403 coupons. ResponseBody "+responseBody)
	})

	It("Post coupon disable test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/coupons/disable"
		payload := `{}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to Post 403 disable coupons. ResponseBody "+responseBody)
	})

	It("Get billing options test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/options?Id=" + place_holder_map["cloud_account_id"]
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 billing options. ResponseBody "+responseBody)
	})

	It("Get billing credit test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/credit"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 billing credit. ResponseBody "+responseBody)
	})

	It("Get billing credit with cloudAccountId test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/credit?cloudAccountId=" + place_holder_map["cloud_account_id"]
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Or(Equal(403), Equal(200)), "Failed to get 403 billing credit with CloudAccountId. ResponseBody "+responseBody) // These tests need to be fixed as they use normal user token not credentials as a service token
	})

	It("Post billing credit test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/credit"
		payload := `{"amountUsed": 0, "cloudAccountId": "string", "couponCode": "string", "created": "2024-10-08T21:46:00.744Z"}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 billing credit. ResponseBody "+responseBody)
	})

	It("Get billing unaplied test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/credit/unapplied"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 billing unapplied. ResponseBody "+responseBody)
	})

	It("Get billing instances deactivate test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/billing/instances/deactivate"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 billing deactivate. ResponseBody "+responseBody)
	})
	//Cloudaccount admin endpoints
	It("Get cloudaccount name test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudaccounts/name/test@intel.com"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 cloudaccounts name. ResponseBody "+responseBody)
	})

	It("Get cloudaccount test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudaccounts"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 cloudaccounts. ResponseBody "+responseBody)
	})

	It("Delete cloudaccount Id test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudaccounts/Id/" + place_holder_map["cloud_account_id"]
		response_status, responseBody := frisby.DeleteInstance2(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Or(Equal(403), Equal(404)), "Failed to get 403 cloudaccounts Id. ResponseBody "+responseBody)
	})

	It("Get cloudaccount Id test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudaccounts/Id/" + place_holder_map["cloud_account_id"]
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 cloudaccounts Id. ResponseBody "+responseBody)
	})

	It("Patch cloudaccount Id test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudaccounts/Id/" + place_holder_map["cloud_account_id"]
		payload := `{}`
		response_status, responseBody := frisby.PatchInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 cloudaccounts Id. ResponseBody "+responseBody)
	})

	It("Post cloudaccount test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudaccounts/"
		payload := `{}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 cloudaccounts. ResponseBody "+responseBody)
	})
	//Cloudcredits admin endpoints
	It("Get cloud credits state test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudcredits/state?cloudAccountId=" + place_holder_map["cloud_account_id"]
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 state. ResponseBody "+responseBody)
	})

	It("Post cloud credits test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudcredits/credit"
		payload := `{}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 credits. ResponseBody "+responseBody)
	})

	It("Post cloud credits coupons disable test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudcredits/coupons/disable"
		payload := `{}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 coupon disable. ResponseBody "+responseBody)
	})

	It("Post cloud credits creditmigrate test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudcredits/credit/creditmigrate"
		payload := `{}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 coupon creditmigrate. ResponseBody "+responseBody)
	})
	//Metering admin endpoints
	It("Patch metering records test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/meteringrecords"
		payload := `{}`
		response_status, responseBody := frisby.PatchInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 metering records. ResponseBody "+responseBody)
	})

	It("Post metering records search empty test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/meteringrecords/search"
		payload := `{}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 metering records empty search. ResponseBody "+responseBody)
	})

	It("Post metering records search test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/meteringrecords/search"
		payload := `{"cloudAccountId": "450780064418", "resourceId": "123abc", "timestamp": "2024-10-08T21:46:00.744Z", "transactionId": "101"}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 metering records search. ResponseBody "+responseBody)
	})

	It("Get metering records previous test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/meteringrecords/previous?id=" + place_holder_map["cloud_account_id"]
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 metering previous. ResponseBody "+responseBody)
	})

	It("Post metering records invalid search", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/meteringrecords/invalid/search"
		payload := `{}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(403), "Failed to get 403 metering records invalid search. ResponseBody "+responseBody)
	})
})
