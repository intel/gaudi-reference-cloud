package e2e

import (
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"
	storage_utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

// endpoints
var instance_endpoint string
var bucket_endpoint string
var principal_endpoint string
var rule_endpoint string
var sshkey_endpoint string
var vnet_endpoint string
var storage_endpoint string
var user_endpoint string
var instance_group_endpoint string

// flags
var skipBMCreation string
var skipBMClusterCreation string

// Across test suite level variables
var sshPublicKey string
var sshPrivateKey string
var cloudAccount string
var vnet string
var sshkeyName string
var sshkeyId string
var token string
var region string
var availabilityZone string
var superComputeRun bool
var isGaudiInstance bool
var userEmail string
var userToken string
var testEnv string
var instanceType string
var machineImage string
var scaleTestRun bool
var scale_bucket_count int
var scale_volume_count int
var scale_principal_count int
var scale_rule_count int
var vastEnabled bool

// suite level variables
var compute_url string
var storage_url string
var cloudaccount_url string
var oidc_url string
var test_env string
var scale_first_bucket_id string
var rule_ids []string

const (
	scale_bucket_name_prefix           = "scale-bucket-"
	max_vm_creation_timeout_in_min int = 10
	max_vm_deletion_timeout_in_min int = 10
)

func init() {
	flag.StringVar(&sshPublicKey, "sshPublicKey", "", "")
	flag.StringVar(&sshPrivateKey, "sshPrivateKey", "", "")
	flag.StringVar(&cloudAccount, "cloudAccount", "", "Target cloud account")
	flag.StringVar(&testEnv, "testEnv", "staging", "Target environment: staging or production")
	flag.StringVar(&region, "region", "us-staging-1", "Target region")
	flag.StringVar(&instanceType, "instanceType", "vm-spr-sml", "Instance type for VM or BM creation")
	flag.StringVar(&machineImage, "machineImage", "ubuntu-2204-jammy-v20250123", "Machine image for VM or BM creation")
	flag.BoolVar(&superComputeRun, "superComputeRun", false, "Set to true for Supercompute runs which will use Gaudi instance groups")
	flag.StringVar(&userToken, "userToken", "", "Optional user specified token")
	flag.StringVar(&userEmail, "userEmail", "", "User's email for login")
	flag.BoolVar(&scaleTestRun, "scaleTestRun", false, "Set to true for scale test runs")
	flag.IntVar(&scale_bucket_count, "scaleBucketCount", 50, "Number of buckets to create for scale test")
	flag.IntVar(&scale_volume_count, "scaleVolumeCount", 50, "Number of volumes to create for scale test")
	flag.IntVar(&scale_principal_count, "scalePrincipalCount", 50, "Number of principals to create for scale test")
	flag.IntVar(&scale_rule_count, "scaleRuleCount", 50, "Number of rules to create for scale test")
	flag.BoolVar(&vastEnabled, "vastEnabled", false, "Set to true when VAST volumes are enabled")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	logger.Init()
	RunSpecs(t, "IDC - Storage E2E Suite")
}

var _ = BeforeSuite(func() {
	logger.Init()
	flag.Parse()
	logger.Logf.Infof("Starting storage public APIs E2E test suite using email: %s, account: %s, region: %s", userEmail, cloudAccount, region)

	if superComputeRun {
		logger.Log.Info("Supercompute test run using Gaudi instance group")
	} else {
		logger.Log.Info("Test run using individual VMs or BMs")
	}

	if vastEnabled {
		logger.Log.Info("Test run will create VAST volumes")
	} else {
		logger.Log.Info("Test run will create WEKA volumes")
	}

	isGaudiInstance = false
	if strings.Contains(strings.ToLower(instanceType), "gaudi") {
		isGaudiInstance = true
	}

	// load the test data and populate the test suite data into objects
	storage_utils.LoadStorageTestConfig("./resources", "staas_e2e_input.json", cloudAccount, testEnv, region, instanceType, machineImage)

	test_env = storage_utils.GetStTestEnv()

	// Test Environment set up
	compute_url, cloudaccount_url, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

	// Suite setup
	cloudAccount, instance_endpoint, sshkey_endpoint, vnet_endpoint, bucket_endpoint, principal_endpoint, rule_endpoint, user_endpoint, storage_endpoint, instance_group_endpoint = storage_utils.StSuiteSetup(test_env, token,
		cloudaccount_url, compute_url, "ACCOUNT_TYPE_INTEL")

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = storage_utils.StSuiteDependenciesSetup(test_env, token, vnet_endpoint, false, sshkey_endpoint, sshPublicKey)

	// populate availability zone
	availabilityZone = storage_utils.GetAvailabilityZone(region, false)

	if scaleTestRun {
		scale_bucket_payload := storage_utils.GetBucketPayload()
		storage_utils.CreateMultipleBuckets(bucket_endpoint, token, scale_bucket_payload, cloudAccount, scale_bucket_name_prefix, scale_bucket_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

		// Create multiple principals on the first test bucket
		scale_principal_payload := storage_utils.GetPrincipalPayload()
		scale_bucket_name := fmt.Sprintf("%s-%s0000", cloudAccount, scale_bucket_name_prefix)
		storage_utils.CreateMultiplePrincipals(principal_endpoint, token, scale_principal_payload, scale_bucket_name, scale_principal_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

		// Get bucket ID of first bucket created for scale tests
		first_bucket_name := fmt.Sprintf("%s-%s0000", cloudAccount, scale_bucket_name_prefix)
		status, body := storage.GetBucketByName(bucket_endpoint, token, first_bucket_name)
		Expect(status).To(Equal(200), body)
		scale_first_bucket_id = gjson.Get(body, "metadata.resourceId").String()

		// Create lifecycle rules on first bucket created for scale test
		scale_rule_payload := storage_utils.GetRulePayload()
		var err error
		rule_ids, err = storage_utils.CreateMultipleRules(rule_endpoint, token, scale_rule_payload, scale_first_bucket_id, scale_rule_count)
		Expect(err).NotTo(HaveOccurred())

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

		// Create volumes for scale test
		scale_storage_payload := storage_utils.GetStoragePayload()
		storage_utils.CreateMultipleFilesystems(storage_endpoint, token, scale_storage_payload, "1TB", scale_volume_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)
	}
})

var _ = AfterSuite(func() {
	time.Sleep(10 * time.Second)

	if scaleTestRun {
		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

		// Delete principals created for scale tests
		storage_utils.DeleteMultiplePrincipals(principal_endpoint, token, scale_principal_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

		// Delete buckets created for scale tests
		storage_utils.DeleteMultipleBuckets(bucket_endpoint, token, cloudAccount, scale_bucket_name_prefix, scale_bucket_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)

		// Delete volumes created for scale tests
		storage_utils.DeleteMultipleFilesystems(storage_endpoint, token, scale_volume_count)

		// Refresh token
		_, _, _, token = storage_utils.StTestEnvSetup(test_env, region, "./resources/auth_e2e_config.json", userEmail, userToken)
	}

	// clean up
	storage_utils.StSuiteCleanup(test_env, cloudaccount_url, sshkeyId, token, sshkey_endpoint, vnet_endpoint, vnet, cloudAccount)
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	storage_utils.ReportGeneration(report)
})
