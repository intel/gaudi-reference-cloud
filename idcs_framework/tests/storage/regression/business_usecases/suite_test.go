package storage

import (
	"flag"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage_utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Endpoints
var storage_endpoint string
var bucket_endpoint string
var principal_endpoint string
var rule_endpoint string
var user_endpoint string

// Across test suite level variables
var token string
var storage_cloud_account string

// suite level variables
var storage_url string
var cloudaccount_url string
var oidc_url string
var test_env string

// User specified parameters
var cloudAccount string
var region string
var userToken string
var userEmail string
var testEnv string
var vastEnabled bool

func init() {
	flag.StringVar(&cloudAccount, "cloudAccount", "", "Target cloud account")
	flag.StringVar(&testEnv, "testEnv", "staging", "Target environment: staging or production")
	flag.StringVar(&region, "region", "us-staging-1", "Target region")
	flag.StringVar(&userToken, "userToken", "", "User specified token")
	flag.StringVar(&userEmail, "userEmail", "", "User's email for login")
	flag.BoolVar(&vastEnabled, "vastEnabled", false, "Set to true when VAST volumes are enabled")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	logger.Init()
	RunSpecs(t, "IDC - Storage services deployed environment Suite")
}

var _ = BeforeSuite(func() {
	logger.Init()
	logger.Log.Info("Starting Storage Public APIs test suite")

	// load the test data and populate the test suite data into objects
	storage_utils.LoadStorageTestConfig("./resources", "staas_fs_size.json", cloudAccount, testEnv, region, "", "")
	test_env = storage_utils.GetStTestEnv()

	// Test Environment set up
	storage_url, cloudaccount_url, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_config.json", userEmail, userToken)

	// Suite setup
	storage_cloud_account, _, _, _, bucket_endpoint, principal_endpoint, rule_endpoint, user_endpoint, storage_endpoint, _ = storage_utils.StSuiteSetup(test_env, token,
		cloudaccount_url, storage_url, "ACCOUNT_TYPE_INTEL")
})

var _ = AfterSuite(func() {
	// Delete cloudaccount if needed
	storage_utils.DeleteCloudAccountForStorage(test_env, token, cloudaccount_url, storage_cloud_account)
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	storage_utils.ReportGeneration(report)
})
