package bm_e2e

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
	instanceEndpoint      string
	instanceGroupEndpoint string
	sshEndpoint           string
	vnetEndpoint          string
	meteringEndpoint      string
	machineImageEndpoint  string
	productsEndpoint      string
	usagesEndpoint        string
)

// Across test suite level variables
var (
	sshPublicKey          string
	cloudAccount          string
	vnet                  string
	sshkeyName            string
	sshkeyId              string
	token                 string
	availabilityZone      string
	instanceGroupSize     string
	bmMachineImage        string
	bmClusterMachineImage string
	logInstance           *logger.CustomLogger
)

// flag
var (
	skipBMCreation        string
	skipBMClusterCreation string
	region                string
	cloudAccountId        string
	vnetName              string
	creationSucceeded     bool
	adminToken            string
	userEmail             string
	userToken             string
	bmInstanceType        string
	bmClusterInstanceType string
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
	flag.StringVar(&testEnv, "testEnv", "", "")
	flag.StringVar(&userEmail, "userEmail", "", "")
	flag.StringVar(&userToken, "userToken", "", "")
	flag.StringVar(&bmInstanceType, "bmInstanceType", "", "")
	flag.StringVar(&bmClusterInstanceType, "bmClusterInstanceType", "", "")
	flag.StringVar(&bmMachineImage, "bmMachineImage", "", "")
	flag.StringVar(&bmClusterMachineImage, "bmClusterMachineImage", "", "")
	flag.StringVar(&instanceGroupSize, "instanceGroupSize", "2", "")
	flag.StringVar(&skipBMCreation, "skipBMCreation", "false", "")
	flag.StringVar(&skipBMClusterCreation, "skipBMClusterCreation", "false", "")
	flag.StringVar(&adminToken, "adminToken", "", "")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IDC - BMaaS E2E")
}

var _ = BeforeSuite(func() {
	logger.InitializeLogger(0)
	logInstance = logger.GetLogger()
	utils.SuitelevelLogSetup(logInstance)

	logInstance.Println("Running setup on Node 0")
	logInstance.Println("Starting BMaaS E2E test suite")

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../resources", "compute_e2e_input.json")

	// Test Environmen set up
	computeUrl, cloudaccountUrl, _, token = utils.TestEnvSetup(testEnv, region, userEmail, "../resources/auth_e2e_config.json", userToken)

	// Fetch admin token
	logInstance.Println("Fetching admin token...")
	if adminToken != "" {
		adminToken = "Bearer " + adminToken
	} else {
		adminToken = utils.FetchAdminToken(testEnv)
	}

	// Suite setup
	cloudAccount, instanceEndpoint, sshEndpoint, vnetEndpoint, _, machineImageEndpoint = utils.SuiteSetup(testEnv, token, cloudaccountUrl,
		computeUrl, "ACCOUNT_TYPE_INTEL", cloudAccountId)

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpoint, true, sshEndpoint, sshPublicKey, vnetName)
	meteringEndpoint = cloudaccountUrl + "/v1/" + "meteringrecords"
	productsEndpoint = cloudaccountUrl + "/v1/" + "products"
	usagesEndpoint = cloudaccountUrl + "/v1/" + "billing/usages"
	instanceGroupEndpoint = computeUrl + "/v1/cloudaccounts/" + cloudAccount + "/instancegroups"

	// populate availability zone
	availabilityZone = utils.GetAvailabilityZone(region, false)
	creationSucceeded = true

	// Instance types and machine images list
	imageAndProductsList = utils.GetImagesAndProductsList(machineImageEndpoint, token)
	logInstance.Println("images and products List: ", imageAndProductsList)
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
