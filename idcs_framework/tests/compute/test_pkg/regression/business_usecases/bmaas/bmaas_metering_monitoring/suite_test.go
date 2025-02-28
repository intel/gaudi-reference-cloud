package metering_monitor

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
	sshEndpoint           string
	vnetEndpoint          string
	meteringEndpoint      string
	instanceGroupEndpoint string
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
	region                string
	vnetName              string
	cloudAccountId        string
	availabilityZone      string
	skipBMCreation        string
	userEmail             string
	userToken             string
	creationSucceeded     bool
	skipBMClusterCreation string
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
	RunSpecs(t, "IDC - BMaaS Metering Monitoring Business usecases")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	logger.InitializeLogger(5)
	logInstance = logger.GetLogger()

	utils.SuitelevelLogSetup(logInstance)

	logInstance.Println("Running setup on Node 5")
	logInstance.Println("Starting BMaaS Metering Monitoring Business Usecase test suite")

	skipBMCreation = utils.SkipTests(testEnv)

	if skipBMCreation == "true" {
		Skip("Skipping the Entire BM flow due to the flag")
	}

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../../resources", "bmaas_business_input.json")

	// update the configmap and restart pod (kubeconfig file for env should be provided - if it is not kind cluster)
	//client.SetUpKubeClient("<<kubeconfig>>", testEnv)
	//Expect(client.UpdateMeteringConfigMap("us-dev-1a-compute-metering-monitor-manager-config", "controller_manager_config.yaml")).To(Equal(true))
	//Expect(client.RestartPod("metering-monitor")).To(Equal(true))

	// Test Environmen set up
	computeUrl, cloudaccountUrl, _, token = utils.TestEnvSetup(testEnv, region, userEmail, "../../resources/auth_config.json", userToken)

	// Suite setup
	cloudAccount, instanceEndpoint, sshEndpoint, vnetEndpoint, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl, computeUrl, "ACCOUNT_TYPE_INTEL", cloudAccountId)

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpoint, false, sshEndpoint, sshPublicKey, vnetName)
	meteringEndpoint = cloudaccountUrl + "/v1/" + "meteringrecords"
	instanceGroupEndpoint = computeUrl + "/v1/cloudaccounts/" + cloudAccount + "/instancegroups"

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
		utils.SuiteCleanup(testEnv, cloudaccountUrl, sshkeyId, token, sshEndpoint, vnetEndpoint, vnet, cloudAccount)
	}
	logInstance.Println("Test teardown completed")
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	utils.ReportGeneration(report)
})
