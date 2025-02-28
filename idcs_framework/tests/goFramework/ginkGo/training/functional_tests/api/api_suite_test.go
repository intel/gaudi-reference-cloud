package api_test

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/training/training_utils"
	"goFramework/testsetup"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var baseUrl string
var baseApiUrl string
var adminToken string
var premiumCloudAccount = make(map[string]string)
var standardCloudAccount = make(map[string]string)
var slurmaasManagementCloudAccountId string

var _ = BeforeSuite(func() {
	logger.InitializeZapCustomLogger()

	training_utils.LoadConfig("../../../training/data", "training_staging.json")
	baseUrl = training_utils.GetBaseUrl()
	baseApiUrl = training_utils.GetBaseApiUrl()
	auth.Get_config_file_data("../../../training/data/training_staging.json")
	adminToken = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	username := auth.Get_UserName("Premium")

	By("Delete premium cloud account if exists...")
	url := baseUrl + "/v1/cloudaccounts"
	cloudAccId, err := testsetup.GetCloudAccountId(username, baseUrl, adminToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, adminToken, cloudAccId)
	} else {
		logger.Log.Info("Existing cloud account not found")
	}

	By("Create premium cloud account...")
	cloudAccountUrl := baseUrl + "/v1/cloudaccounts/enroll"
	tokenResponse, _ := auth.Get_Azure_Bearer_Token(username)
	premiumCloudAccount["token"] = "Bearer " + tokenResponse
	enrollPayload := `{"premium":true}`
	cloudAccountCreationStatus, cloudAccountCreationBody := financials.CreateCloudAccount(cloudAccountUrl, premiumCloudAccount["token"], enrollPayload)
	Expect(cloudAccountCreationStatus).To(Equal(200), "Failed to create Cloud Account using enroll")
	premiumCloudAccount["id"] = gjson.Get(cloudAccountCreationBody, "cloudAccountId").String()
	premiumCloudAccount["type"] = gjson.Get(cloudAccountCreationBody, "cloudAccountType").String()
	Expect(strings.Contains(cloudAccountCreationBody, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of Premium user")
	logger.Logf.Info("Created premium cloud account:", premiumCloudAccount["id"])

	By("Delete standard cloud account if exists...")
	username = auth.Get_UserName("Standard")
	url = baseUrl + "/v1/cloudaccounts"
	cloudAccId, err = testsetup.GetCloudAccountId(username, baseUrl, adminToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, adminToken, cloudAccId)
	} else {
		logger.Log.Info("Existing cloud account not found")
	}

	By("Create standard cloud account...")
	cloudAccountUrl = baseUrl + "/v1/cloudaccounts/enroll"
	tokenResponse, _ = auth.Get_Azure_Bearer_Token(username)
	standardCloudAccount["token"] = "Bearer " + tokenResponse
	enrollPayload = `{"premium":false}`
	cloudAccountCreationStatus, cloudAccountCreationBody = financials.CreateCloudAccount(cloudAccountUrl, standardCloudAccount["token"], enrollPayload)
	Expect(cloudAccountCreationStatus).To(Equal(200), "Failed to create Cloud Account using enroll")
	standardCloudAccount["id"] = gjson.Get(cloudAccountCreationBody, "cloudAccountId").String()
	standardCloudAccount["type"] = gjson.Get(cloudAccountCreationBody, "cloudAccountType").String()
	Expect(strings.Contains(cloudAccountCreationBody, `"cloudAccountType":"ACCOUNT_TYPE_STANDARD"`)).To(BeTrue(), "Validation failed on Enrollment of Premium user")
	logger.Logf.Info("Created premium cloud account:", standardCloudAccount["id"])

	By("Info for SlurmaaS Management account...")
	username = training_utils.GetSlurmaasManagementUsername()
	slurmaasManagementCloudAccountId, err = testsetup.GetCloudAccountId(username, baseUrl, adminToken)
	Expect(err).To(BeNil())
})

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}

var _ = AfterSuite(func() {
	url := baseUrl + "/v1/cloudaccounts"
	logger.Logf.Info("Deleting cloud accounts after suite")
	financials.DeleteCloudAccountById(url, adminToken, premiumCloudAccount["id"])
	financials.DeleteCloudAccountById(url, adminToken, standardCloudAccount["id"])
})
