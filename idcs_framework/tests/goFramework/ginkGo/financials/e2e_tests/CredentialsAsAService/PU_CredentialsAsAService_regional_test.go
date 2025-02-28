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

var _ = Describe("CredentialsAsAService", Ordered, Label("Regional"), func() {
	var base_url string
	var token string
	var userName string
	var userToken string
	var cloud_account_created string

	var ariaAuth string
	var place_holder_map = make(map[string]string)

	var ariaclientId_su string
	var clientId string

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../data", "vmaas_input.json")
		auth.Get_config_file_data("../../data/config.json")
		financials_utils.LoadE2EConfig("../../data", "billing.json")
		userName = auth.Get_UserName("Premium")
		base_url = compute_utils.GetBaseUrl()
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

	It("Create user credential test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/user/credentials/" + place_holder_map["cloud_account_id"] + "/create"
		payload := `{appClientName:"test"}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Or(Equal(200), Equal(404)), "Failed to create credentials. ResponseBody "+responseBody)
	})

	It("Get user credentials", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/user/credentials/" + place_holder_map["cloud_account_id"] + "/list"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		clientId = gjson.Get(responseBody, "appClients[0].clientId").String()
		Expect(response_status).To(Or(Equal(200), Equal(404)), "Failed to get 200 All credentials. ResponseBody "+responseBody)
	})

	It("Delete user credential test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/user/credentials/" + place_holder_map["cloud_account_id"] + "/delete?clientId=" + clientId
		response_status, responseBody := frisby.DeleteInstance2(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Or(Equal(200), Equal(404)), "Failed to delete credentials. ResponseBody "+responseBody)
	})

	It("Revoke member user credential test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/user/credentials/" + place_holder_map["cloud_account_id"] + "/member/revoke"
		payload := `{cloudaccountId :"` + place_holder_map["cloud_account_id"] + `", revoked: "yes", memberEmail: "test@intel.com"}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Or(Equal(200), Equal(404)), "Failed to revoke credentials. ResponseBody "+responseBody)
	})

	It("Revoke user credential test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/user/credentials/" + place_holder_map["cloud_account_id"] + "/revoke"
		payload := `{cloudaccountId :"` + place_holder_map["cloud_account_id"] + `", revoked: "yes"}`
		response_status, responseBody := frisby.CreateInstance(baseUrl, userToken, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Or(Equal(200), Equal(404)), "Failed to revoke credentials. ResponseBody "+responseBody)
	})

	It("Get All IK8 Cluster test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/iks/clusters"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Or(Equal(200), Equal(404)), "Failed to get 200 All IK8 Cluster. ResponseBody "+responseBody)
	})

	It("Get machine images test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/machineimages/"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(200), "Failed to get 200 machine images. ResponseBody "+responseBody)
	})

	It("Get vnets test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/vnets"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(200), "Failed to get 200 vnets. ResponseBody "+responseBody)
	})

	It("Get all filesystems test", func() {
		fmt.Print(base_url)
		baseUrl := base_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/filesystems"
		response_status, responseBody := frisby.GetAllInstance(baseUrl, userToken)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).To(Equal(200), "Failed to get 200 filesystems. ResponseBody "+responseBody)
	})
})
