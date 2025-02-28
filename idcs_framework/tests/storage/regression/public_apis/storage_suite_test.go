package public_apis

import (
	"flag"
	"fmt"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"
	storage_utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

// Endpoints
var storage_endpoint string
var user_endpoint string
var bucket_endpoint string
var principal_endpoint string
var rule_endpoint string

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
var userEmail string
var userToken string
var testEnv string
var scaleTestRun bool
var bucket_payload string
var rule_ids []string
var scale_first_bucket_id string
var scale_first_bucket_name string
var scale_bucket_count int
var scale_volume_count int
var scale_principal_count int
var scale_rule_count int
var concurr_account string
var concurr_region string
var concurrTestRun bool
var vastEnabled bool

const (
	scale_bucket_name_prefix = "scale-bucket-"
)

func init() {
	flag.StringVar(&cloudAccount, "cloudAccount", "", "Target cloud account")
	flag.StringVar(&testEnv, "testEnv", "staging", "Target environment: staging or production")
	flag.StringVar(&region, "region", "us-staging-1", "Target region")
	flag.StringVar(&userToken, "userToken", "", "User specified token")
	flag.StringVar(&userEmail, "userEmail", "", "User's email")
	flag.BoolVar(&scaleTestRun, "scaleTestRun", false, "Run scale tests")
	flag.IntVar(&scale_bucket_count, "scaleBucketCount", 50, "Number of buckets to create for scale test")
	flag.IntVar(&scale_volume_count, "scaleVolumeCount", 50, "Number of volumes to create for scale test")
	flag.IntVar(&scale_principal_count, "scalePrincipalCount", 50, "Number of principals to create for scale test")
	flag.IntVar(&scale_rule_count, "scaleRuleCount", 50, "Number of rules to create for scale test")
	flag.StringVar(&concurr_account, "concurrAccount", "810990250449", "Valid account for concurrency runs")
	flag.StringVar(&concurr_region, "concurrRegion", "us-staging-1", "Valid region for concurrency runs")
	flag.BoolVar(&concurrTestRun, "concurrTestRun", false, "Run concurrency tests")
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
	flag.Parse()
	logger.Logf.Infof("Starting storage public APIs Regression test suite using email: %s, account: %s, region: %s", userEmail, cloudAccount, region)

	if vastEnabled {
		logger.Log.Info("Test run will create VAST volumes")
	} else {
		logger.Log.Info("Test run will create WEKA volumes")
	}

	// load the test data and populate the test suite data into objects
	storage_utils.LoadStorageTestConfig("./resources", "staas_input.json", cloudAccount, testEnv, region, "", "")
	test_env = storage_utils.GetStTestEnv()

	// Test Environment set up
	storage_url, cloudaccount_url, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_config.json", userEmail, userToken)

	// Suite setup
	storage_cloud_account, _, _, _, bucket_endpoint, principal_endpoint, rule_endpoint, user_endpoint, storage_endpoint, _ = storage_utils.StSuiteSetup(test_env, token,
		cloudaccount_url, storage_url, "ACCOUNT_TYPE_INTEL")

	if scaleTestRun {
		// Create buckets
		scale_bucket_payload := storage_utils.GetBucketPayload()
		storage_utils.CreateMultipleBuckets(bucket_endpoint, token, scale_bucket_payload, cloudAccount, scale_bucket_name_prefix, scale_bucket_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_config.json", userEmail, userToken)

		// Create multiple principals on the first test bucket
		scale_principal_payload := storage_utils.GetPrincipalPayload()
		scale_first_bucket_name = fmt.Sprintf("%s-%s0000", cloudAccount, scale_bucket_name_prefix)
		storage_utils.CreateMultiplePrincipals(principal_endpoint, token, scale_principal_payload, scale_first_bucket_name, scale_principal_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_config.json", userEmail, userToken)

		// Get bucket ID of first bucket created for scale tests
		status, body := storage.GetBucketByName(bucket_endpoint, token, scale_first_bucket_name)
		Expect(status).To(Equal(200), body)
		scale_first_bucket_id = gjson.Get(body, "metadata.resourceId").String()

		// Create lifecycle rules on first bucket created for scale test
		scale_rule_payload := storage_utils.GetRulePayload()
		var err error
		rule_ids, err = storage_utils.CreateMultipleRules(rule_endpoint, token, scale_rule_payload, scale_first_bucket_id, scale_rule_count)
		Expect(err).NotTo(HaveOccurred())

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_config.json", userEmail, userToken)

		// Create multple volumes for scale test
		scale_storage_payload := storage_utils.GetStoragePayload()
		storage_utils.CreateMultipleFilesystems(storage_endpoint, token, scale_storage_payload, "1TB", scale_volume_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_config.json", userEmail, userToken)
	}
})

var _ = AfterSuite(func() {
	if scaleTestRun {
		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_config.json", userEmail, userToken)

		// Delete principals created for scale tests
		storage_utils.DeleteMultiplePrincipals(principal_endpoint, token, scale_principal_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_config.json", userEmail, userToken)

		// Delete buckets created for scale tests
		storage_utils.DeleteMultipleBuckets(bucket_endpoint, token, cloudAccount, scale_bucket_name_prefix, scale_bucket_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_config.json", userEmail, userToken)

		// Delete volumes created for scale tests
		storage_utils.DeleteMultipleFilesystems(storage_endpoint, token, scale_volume_count)
	}

	// Delete cloudAccount is needed
	storage_utils.DeleteCloudAccountForStorage(test_env, token, cloudaccount_url, storage_cloud_account)
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	storage_utils.ReportGeneration(report)
})
