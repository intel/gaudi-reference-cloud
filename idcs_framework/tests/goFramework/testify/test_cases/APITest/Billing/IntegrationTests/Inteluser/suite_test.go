package BillingAPITest

import (
	"bytes"
	"time"

	"goFramework/framework/common/logger"
	"goFramework/framework/common/report_portal"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"io"
	"os"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

var rescueStdout *os.File
var r *os.File
var w *os.File
var Logger *zap.Logger

type BillingAPITestSuite struct {
	suite.Suite
	buf       bytes.Buffer
	testdata  report_portal.TestData
	suitedata report_portal.SuiteData
	stepdata  report_portal.TestStep
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestBillingAPITestSuite(t *testing.T) {
	suite.Run(t, new(BillingAPITestSuite))
}

func (suite *BillingAPITestSuite) SetupSuite() {
	os.Setenv("Test_Suite", "testify")
	os.Setenv("fetch_admin_role_token", "True")
	auth.Get_config_file_data("../../../../../test_config/config.json")
	compute_utils.LoadE2EConfig("../../../../../../ginkGo/compute/data", "vmaas_input.json")
	financials_utils.LoadE2EConfig("../../../../../../ginkGo/financials/data", "billing.json")
	logger.InitializeZapCustomLogger()
	rescueStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w
	logger.Log.Info("Starting Billing REST   test suite")
	logger.Log.Info("Starting Report Portal Setup")
	logger.Log.Info("-- Start Launch --")
	//report_portal.StartLaunch(report_portal.Launchdata)
	logger.Log.Info("-- Start Suite --")
	suite.suitedata.Name = "IDC Billing REST  Test Suite"
	suite.suitedata.Description = "Suite - IDC Billing REST  End Point Tests"
	//report_portal.StartSuite(report_portal.Launchdata, &suite.suitedata, &suite.testdata)
	suite.testdata.Name = suite.T().Name()
	logger.Log.Info("-- Start Test --")
	//log.SetOutput(&suite.buf)
	suite.testdata.Description = "Test - IDC Billing REST  Test"
	utils.Get_config_file_data()
	utils.Get_Billing_config_file_data()
	utils.Get_Metering_config_file_data()
	utils.Get_config_file_data()
	utils.Get_CA_config_file_data()
	utils.LoadCAConfig("../../../test_config/cloud_accounts.json")
	//report_portal.StartChildContainer(report_portal.Launchdata, &suite.testdata)

}

func (suite *BillingAPITestSuite) TearDownSuite() {
	logger.Log.Info("Finishing Sample test suite")
	logger.Log.Info("-- Finish Test --")
	//report_portal.FinishChildContainer(&suite.testdata)
	logger.Log.Info("-- Finish Suite --")
	//report_portal.FinishSuite(report_portal.Launchdata, &suite.testdata)
	logger.Log.Info("-- Finish Launch --")
	//report_portal.FinishLaunch(report_portal.Launchdata)
}

func (suite *BillingAPITestSuite) SetupTest() {
	logger.Log.Info("Starting IDC Billing REST  Functional test cases")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("################################################################################################################")
	logger.Logf.Infof(" Executing Test ------------ %s ", suite.T().Name())
	logger.Log.Info("################################################################################################################")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	suite.stepdata.Name = suite.T().Name()
	//report_portal.StartChildStep(report_portal.Launchdata, &suite.stepdata, &suite.testdata)
	//Enroll Standard User before test

	logger.Log.Info("Creating a intel User Cloud Account using Enroll API")
	os.Setenv("intel_user_test", "True")
	authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	tid := cloudAccounts.Rand_token_payload_gen()
	enterpriseId := "testeid-01"
	//Create Cloud Account
	userName := utils.Get_UserName("Intel")
	userToken, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + userToken
	base_url := utils.Get_Base_Url1()
	url := base_url + "/v1/cloudaccounts"
	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, authToken)
	if err == nil {
		financials.DeleteCloudAccountById(url, authToken, cloudAccId)
	}
	time.Sleep(1 * time.Minute)

	get_CAcc_id, acc_type, _ := cloudAccounts.CreateCAccwithEnroll("intel", tid, userName, enterpriseId, false, 200)
	assert.NotEqual(suite.T(), get_CAcc_id, "False", "Test Failed while creating an enterprise user cloud account using Enroll API")
	assert.Equal(suite.T(), acc_type, "ACCOUNT_TYPE_INTEL", "Test Failed while validating type of cloud account")
	ret_value1, responsePayload := cloudAccounts.GetCAccById(get_CAcc_id, 200)
	// paidSerAllowed := gjson.Get(responsePayload, "paidServicesAllowed").String()
	// lowCred := gjson.Get(responsePayload, "lowCredits").String()
	CountryCode := gjson.Get(responsePayload, "countryCode").String()
	// assert.Equal(suite.T(), "false", paidSerAllowed, "Failed: Validation Failed , Get on cloud account for paidServices")
	assert.Equal(suite.T(), "", CountryCode, "Failed: Validation of country code")
	// assert.Equal(suite.T(), "true", lowCred, "Failed: Validation Failed , Get on cloud account for lowCredits")
	assert.NotEqual(suite.T(), ret_value1, "False", "Test Failed to get cloudaccount by id")

}

// func (suite *BillingAPITestSuite) BeforeTest(suiteName, testName string) {

// }

// func (suite *BillingAPITestSuite) AfterTest(suiteName, testName string) {

// }

func (suite *BillingAPITestSuite) TearDownTest() {
	logger.Log.Info("In teardown IDC Billing REST  Functional test cases")
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	//out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	suite.testdata.TestLog = buf.String()
	logger.Log.Info("-- Finish Test Step --")
	if suite.T().Failed() == true {
		suite.stepdata.StepStatus = "FAILED"
	} else {
		suite.stepdata.StepStatus = "PASSED"
	}
	//report_portal.FinishChildStep(report_portal.Launchdata, &suite.stepdata, &suite.testdata)
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("################################################################################################################")
	logger.Logf.Infof(" Finished Test ------------ %s ", suite.T().Name())
	logger.Log.Info("################################################################################################################")
	logger.Log.Info("                                                                                                                ")
	logger.Log.Info("                                                                                                                ")

	// logger.Log.Info("Deleting Standard User")
	// base_url := utils.Get_Base_Url1()
	// userName := utils.Get_UserName("Standard")
	// authToken := "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	// logger.Log.Info("userName" + userName)

}
