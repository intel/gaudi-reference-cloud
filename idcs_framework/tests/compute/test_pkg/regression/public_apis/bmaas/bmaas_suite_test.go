package bmaas

import (
	"flag"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

// endpoints
var (
	instanceEndpoint      string
	sshEndpoint           string
	vnetEndpoint          string
	instanceTypeEndpoint  string
	machineImageEndpoint  string
	instanceGroupEndpoint string
)

// Across test suite level variables
var (
	cloudAccount     string
	vnet             string
	sshkeyName       string
	sshkeyId         string
	token            string
	availabilityZone string
	logInstance      *logger.CustomLogger
)

// flags
var (
	vnetName              string
	cloudAccountId        string
	userEmail             string
	userToken             string
	sshPublicKey          string
	skipBMCreation        string
	skipBMClusterCreation string
	region                string
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
	flag.StringVar(&skipBMCreation, "skipBMCreation", "false", "")
	flag.StringVar(&skipBMClusterCreation, "skipBMClusterCreation", "true", "")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IDC - Compute Regression Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	logger.InitializeLogger(0)
	logInstance = logger.GetLogger()

	utils.SuitelevelLogSetup(logInstance)

	logInstance.Println("Running setup on Node 0")
	logInstance.Println("Starting Public APIs test suite")

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../resources", "bmaas_input.json")

	// Test Environment set up
	computeUrl, cloudaccountUrl, _, token = utils.TestEnvSetup(testEnv, region, userEmail, "../resources/auth_config.json", userToken)

	// Suite setup
	cloudAccount, instanceEndpoint, sshEndpoint, vnetEndpoint, instanceTypeEndpoint, machineImageEndpoint = utils.SuiteSetup(testEnv, token, cloudaccountUrl, computeUrl, "ACCOUNT_TYPE_INTEL", cloudAccountId)

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpoint, true, sshEndpoint, sshPublicKey, vnetName)

	instanceGroupEndpoint = computeUrl + "/v1/cloudaccounts/" + cloudAccount + "/instancegroups"

	// Load machineImages list
	machineImagesList := utils.GetJsonArray("machineImagesMapping")
	machineImageMapping = utils.CreateMIAndITInputMap(machineImagesList)

	// populate availability zone
	availabilityZone = utils.GetAvailabilityZone(region, true)
	return []byte(token)
}, func() {
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	time.Sleep(10 * time.Second)
	// Delete ssh key, cloudaccount and vnet if needed
	utils.SuiteCleanup(testEnv, cloudaccountUrl, sshkeyId, token, sshEndpoint, vnetEndpoint, vnet, cloudAccount)
	logInstance.Println("Test teardown completed")
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	utils.ReportGeneration(report)
})
