package bm_gaudi3_test

import (
	"flag"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

// endpoints
var (
	instanceEndpoint string
	sshEndpoint      string
	vnetEndpoint     string
)

// Across test suite level variables
var (
	sshPublicKey   string
	cloudAccount   string
	vnet           string
	sshkeyName     string
	sshkeyId       string
	token          string
	kubeConfigPath string
	logInstance    *logger.CustomLogger
)

// flag
var (
	skipBMCreation    string
	region            string
	cloudAccountId    string
	vnetName          string
	availabilityZone  string
	userEmail         string
	userToken         string
	creationSucceeded bool
)

// suite level variables
var (
	computeUrl          string
	cloudaccountUrl     string
	testEnv             string
	machineImageMapping map[string]string
)

func init() {
	flag.StringVar(&sshPublicKey, "sshPublicKey", "", "")
	flag.StringVar(&region, "region", "us-dev-1", "")
	flag.StringVar(&cloudAccountId, "cloudAccountId", "", "")
	flag.StringVar(&vnetName, "vnetName", "", "")
	flag.StringVar(&testEnv, "testEnv", "", "")
	flag.StringVar(&userEmail, "userEmail", "", "")
	flag.StringVar(&userToken, "userToken", "", "")
	flag.StringVar(&skipBMCreation, "skipBMCreation", "true", "")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IDC - BMaaS Gaudi3 Business usecases")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	logger.InitializeLogger(5)
	logInstance = logger.GetLogger()

	utils.SuitelevelLogSetup(logInstance)

	logInstance.Println("Running setup on Node 5")
	logInstance.Println("Starting BMaaS Gaudi3 test suite")

	if skipBMCreation == "true" {
		Skip("Skipping the Entire BM flow due to the flag")
	}

	kubeConfigPath = "../../resources/kubeconfig/env-kubeconfig.yaml"

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../../resources", "bmaas_business_input.json")

	// Test Environment set up
	computeUrl, cloudaccountUrl, _, token = utils.TestEnvSetup(testEnv, region, userEmail, "../../resources/auth_config.json", userToken)

	// Suite setup
	cloudAccount, instanceEndpoint, sshEndpoint, vnetEndpoint, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl,
		computeUrl, "ACCOUNT_TYPE_INTEL", cloudAccountId)

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpoint, true, sshEndpoint, sshPublicKey, vnetName)

	// Load machineImages list
	machineImagesList := utils.GetJsonArray("machineImagesMapping")
	machineImageMapping = utils.CreateMIAndITInputMap(machineImagesList)

	// populate availability zone
	availabilityZone = utils.GetAvailabilityZone(region, true)

	creationSucceeded = true
	return []byte(token)
}, func() {
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	if creationSucceeded {
		// Delete ssh key, cloudaccount and vnet if needed
		utils.SuiteCleanup(testEnv, cloudaccountUrl, sshkeyId, token, sshEndpoint, vnetEndpoint, vnet, cloudAccount)
	}
	logInstance.Println("Test teardown completed")
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	utils.ReportGeneration(report)
})
