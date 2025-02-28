package jupyter_test

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
var userToken string
var premiumCloudAccount = make(map[string]string)

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
	userToken = "Bearer " + tokenResponse
	enrollPayload := `{"premium":true}`
	cloudAccountCreationStatus, cloudAccountCreationBody := financials.CreateCloudAccount(cloudAccountUrl, userToken, enrollPayload)
	Expect(cloudAccountCreationStatus).To(Equal(200), "Failed to create Cloud Account using enroll")
	premiumCloudAccount["id"] = gjson.Get(cloudAccountCreationBody, "cloudAccountId").String()
	premiumCloudAccount["type"] = gjson.Get(cloudAccountCreationBody, "cloudAccountType").String()
	Expect(strings.Contains(cloudAccountCreationBody, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of Premium user")
	logger.Logf.Info("Created premium cloud account:", premiumCloudAccount["id"])
})

func TestJupyter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jupyter Suite")
}

var _ = AfterSuite(func() {
	url := baseUrl + "/v1/cloudaccounts"
	logger.Logf.Info("Deleting cloud account after suite")
	financials.DeleteCloudAccountById(url, adminToken, premiumCloudAccount["id"])
})
