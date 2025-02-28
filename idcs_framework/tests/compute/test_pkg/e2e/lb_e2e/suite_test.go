package lb_e2e

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
	instanceEndpoint     string
	sshEndpoint          string
	vnetEndpoint         string
	meteringEndpoint     string
	machineImageEndpoint string
	productsEndpoint     string
	usagesEndpoint       string
	lbInstanceEndpoint   string
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
	region            string
	cloudAccountId    string
	vnetName          string
	creationSucceeded bool
	adminToken        string
	userEmail         string
	userToken         string
)

// suite level variables
var (
	testEnv                 string
	computeUrl              string
	cloudaccountUrl         string
	imageAndProductsList    map[string][]string
	predefinedMachineImages map[string][]string
)

func init() {
	flag.StringVar(&sshPublicKey, "sshPublicKey", "", "")
	flag.StringVar(&region, "region", "us-dev-1", "")
	flag.StringVar(&cloudAccountId, "cloudAccountId", "", "")
	flag.StringVar(&vnetName, "vnetName", "", "")
	flag.StringVar(&testEnv, "testEnv", "kind", "")
	flag.StringVar(&userEmail, "userEmail", "", "")
	flag.StringVar(&adminToken, "adminToken", "", "")
	flag.StringVar(&userToken, "userToken", "", "")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IDC - LBaaS E2E")
}

var _ = BeforeSuite(func() {
	logger.InitializeLogger(0)
	logInstance = logger.GetLogger()
	utils.SuitelevelLogSetup(logInstance)

	logInstance.Println("Running setup on Node 0")
	logInstance.Println("Starting LBaaS E2E test suite")

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../resources", "compute_e2e_input.json")

	// Test Environmen set up
	computeUrl, cloudaccountUrl, _, token = utils.TestEnvSetup(testEnv, region, userEmail, "../resources/auth_e2e_config.json", userToken)

	// Suite setup
	cloudAccount, instanceEndpoint, sshEndpoint, vnetEndpoint, _, machineImageEndpoint = utils.SuiteSetup(testEnv, token, cloudaccountUrl,
		computeUrl, "ACCOUNT_TYPE_INTEL", cloudAccountId)

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpoint, false, sshEndpoint, sshPublicKey, vnetName)
	meteringEndpoint = cloudaccountUrl + "/v1/" + "meteringrecords"
	productsEndpoint = cloudaccountUrl + "/v1/" + "products"
	usagesEndpoint = cloudaccountUrl + "/v1/" + "billing/usages"
	lbInstanceEndpoint = computeUrl + "/v1/cloudaccounts/" + cloudAccount + "/loadbalancers"

	// populate availability zone
	availabilityZone = utils.GetAvailabilityZone(region, false)
	creationSucceeded = true

	// Instance types and machine images list
	imageAndProductsList = utils.GetImagesAndProductsList(machineImageEndpoint, token)
	logger.Logf.Info("images and products List: ", imageAndProductsList)

	// pre defined machine images for specific instance types
	getpredefinedMachineImagesList := utils.GetJsonObject("preDefinedMachineImagesList")
	predefinedMachineImages = utils.JsonToMap(getpredefinedMachineImagesList)
})

var _ = AfterSuite(func() {

	if creationSucceeded {
		// Delete ssh key, cloudaccount and vnet if needed
		utils.SuiteCleanup(testEnv, cloudaccountUrl, sshkeyId, token, sshEndpoint, vnetEndpoint, vnet, cloudAccount)
	}

})

var _ = ReportAfterSuite("custom report", func(report Report) {
	utils.ReportGeneration(report)
})
