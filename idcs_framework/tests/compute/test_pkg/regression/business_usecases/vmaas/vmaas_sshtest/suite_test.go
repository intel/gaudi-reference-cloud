package sshtest

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
	sshPublicKey     string
	cloudAccount     string
	vnet             string
	sshkeyName       string
	sshkeyId         string
	token            string
	availabilityZone string
	logInstance      *logger.CustomLogger
)

// flag
var (
	skipVMCreation    string
	userEmail         string
	userToken         string
	region            string
	vnetName          string
	cloudAccountId    string
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
	flag.StringVar(&region, "region", "us-staging-1", "")
	flag.StringVar(&cloudAccountId, "cloudAccountId", "", "")
	flag.StringVar(&vnetName, "vnetName", "", "")
	flag.StringVar(&testEnv, "testEnv", "", "")
	flag.StringVar(&userEmail, "userEmail", "", "")
	flag.StringVar(&userToken, "userToken", "", "")
	flag.StringVar(&skipVMCreation, "skipVMCreation", "false", "")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IDC - VMaaS SSH Business usecases")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	logger.InitializeLogger(10)
	logInstance = logger.GetLogger()

	utils.SuitelevelLogSetup(logInstance)

	logInstance.Println("Running setup on Node 10")
	logInstance.Println("Starting VMaaS SSH test suite")

	if skipVMCreation == "true" {
		Skip("Skipping the Entire VM flow due to the flag")
	}

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../../resources", "vmaas_business_input.json")

	// Test Environmen set up
	computeUrl, cloudaccountUrl, _, token = utils.TestEnvSetup(testEnv, region, userEmail, "../../resources/auth_config.json", userToken)

	// Suite setup
	cloudAccount, instanceEndpoint, sshEndpoint, vnetEndpoint, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl, computeUrl, "ACCOUNT_TYPE_INTEL", cloudAccountId)

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpoint, false, sshEndpoint, sshPublicKey, vnetName)

	// Load machineImages list
	machineImagesList := utils.GetJsonArray("machineImagesMapping")
	machineImageMapping = utils.CreateMIAndITInputMap(machineImagesList)

	// populate availability zone
	availabilityZone = utils.GetAvailabilityZone(region, false)

	creationSucceeded = true
	return []byte(token)
}, func(data []byte) {

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
