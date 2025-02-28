package bmaas_multitenancy

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
	instanceEndpointCA1, instanceEndpointCA2 string
	sshEndpointCA1, sshEndpointCA2           string
	vnetEndpointCA1, vnetEndpointCA2         string
)

// Across test suite level variables
var (
	sshPublicKey                     string
	cloudAccount, secondCloudAccount string
	vnet, secondVnet                 string
	sshkeyName, secondSSHKeyName     string
	sshkeyId, secondSSHKeyId         string
	token                            string
	availabilityZone                 string
	logInstance                      *logger.CustomLogger
)

// flag
var (
	skipBMCreation                       string
	userEmail                            string
	userToken                            string
	region                               string
	vnetName                             string
	cloudAccountId, secondCloudAccountId string
	creationSucceeded                    bool
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
	flag.StringVar(&region, "region", "", "")
	flag.StringVar(&cloudAccountId, "cloudAccountId", "", "")
	flag.StringVar(&secondCloudAccountId, "secondCloudAccountId", "", "")
	flag.StringVar(&vnetName, "vnetName", "", "")
	flag.StringVar(&testEnv, "testEnv", "", "")
	flag.StringVar(&userEmail, "userEmail", "", "")
	flag.StringVar(&userToken, "userToken", "", "")
	flag.StringVar(&skipBMCreation, "skipBMCreation", "true", "")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IDC - BMaaS Multitenancy Business usecases")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	logger.InitializeLogger(7)
	logInstance = logger.GetLogger()

	utils.SuitelevelLogSetup(logInstance)

	logInstance.Println("Running setup on Node 7")
	logInstance.Println("Starting BMaaS Multitenancy test suite")

	if skipBMCreation == "true" {
		Skip("Skipping the Entire BM flow due to the flag")
	}

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../../resources", "bmaas_business_input.json")

	// Test Environmen set up
	computeUrl, cloudaccountUrl, _, token = utils.TestEnvSetup(testEnv, region, userEmail, "../../resources/auth_config.json", userToken)

	// Suite setup
	cloudAccount, instanceEndpointCA1, sshEndpointCA1, vnetEndpointCA1, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl, computeUrl, "ACCOUNT_TYPE_INTEL", cloudAccountId)

	// second cloudaccount
	secondCloudAccount, instanceEndpointCA2, sshEndpointCA2, vnetEndpointCA2, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl, computeUrl, "ACCOUNT_TYPE_INTEL", secondCloudAccountId)

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpointCA1, false, sshEndpointCA1, sshPublicKey, vnetName)

	// Second user key and vnet
	secondVnet, secondSSHKeyName, secondSSHKeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpointCA2, false, sshEndpointCA2, sshPublicKey, vnetName)

	// Load machineImages list
	machineImagesList := utils.GetJsonArray("machineImagesMapping")
	machineImageMapping = utils.CreateMIAndITInputMap(machineImagesList)

	// populate availability zone
	availabilityZone = utils.GetAvailabilityZone(region, false)

	creationSucceeded = true
	return []byte(token)
}, func() {

})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	if creationSucceeded {
		// Delete ssh key, cloudaccount and vnet if needed
		utils.SuiteCleanup(testEnv, cloudaccountUrl, sshkeyId, token, sshEndpointCA1, vnetEndpointCA1, vnet, cloudAccount)

		// Delete second ssh key, cloudaccount and vnet if needed
		utils.SuiteCleanup(testEnv, cloudaccountUrl, secondSSHKeyId, token, sshEndpointCA2, vnetEndpointCA2, secondVnet, secondCloudAccount)
	}
	logInstance.Println("Test teardown completed")
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	utils.ReportGeneration(report)
})
