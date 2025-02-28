package diffMachineImage

import (
	"flag"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

// endpoints
var (
	instanceEndpoint string
	sshEndpoint      string
	vnetEndpoint     string
)

// Across test suite level variables
var (
	availabilityZone string
	cloudAccount     string
	vnet             string
	sshkeyName       string
	sshkeyId         string
	token            string
	allMachineImages []string
	resourceIds      []string
	logInstance      *logger.CustomLogger
)

// flag
var (
	skipVMCreation    string
	region            string
	cloudAccountId    string
	sshPublicKey      string
	userEmail         string
	userToken         string
	vnetName          string
	creationSucceeded bool
)

// suite level variables
var (
	computUrl       string
	cloudaccountUrl string
	testEnv         string
)

func init() {
	flag.StringVar(&sshPublicKey, "sshPublicKey", "", "")
	flag.StringVar(&region, "region", "us-dev-1", "")
	flag.StringVar(&cloudAccountId, "cloudAccountId", "", "")
	flag.StringVar(&vnetName, "vnetName", "", "")
	flag.StringVar(&testEnv, "testEnv", "", "")
	flag.StringVar(&userEmail, "userEmail", "", "")
	flag.StringVar(&userToken, "userToken", "", "")
	flag.StringVar(&skipVMCreation, "skipVMCreation", "true", "")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IDC - VMaaS Creation with all machine images Business usecases")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	logger.InitializeLogger(11)
	logInstance = logger.GetLogger()

	utils.SuitelevelLogSetup(logInstance)

	logInstance.Println("Running setup on Node 11")
	logInstance.Println("Starting VMaaS Creation with all machine images test suite")

	if skipVMCreation == "true" {
		Skip("Skipping the Entire VM flow due to the flag")
	}

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../../resources", "vmaas_business_input.json")

	// Test Environment set up
	computUrl, cloudaccountUrl, _, token = utils.TestEnvSetup(testEnv, region, userEmail, "../../resources/auth_config.json", userToken)

	// Suite setup
	cloudAccount, instanceEndpoint, sshEndpoint, vnetEndpoint, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl, computUrl, "ACCOUNT_TYPE_INTEL", cloudAccountId)

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpoint, true, sshEndpoint, sshPublicKey, vnetName)

	// Load AllMachineImages list
	var err error
	allMachineImages, err = utils.GetAllImagesForInstanceType(utils.GetJsonValue("instanceTypeToBeCreated"))
	logInstance.Println("allMachineImages: ", allMachineImages)
	Expect(err).ShouldNot(HaveOccurred(), "Error fetching machineImages")
	Expect(allMachineImages).ToNot(BeNil())

	// populate availability zone
	availabilityZone = utils.GetAvailabilityZone(region, true)

	creationSucceeded = true
	return []byte(token)
}, func() {

})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	// Clean up if instances are not deleted
	for _, resourceId := range resourceIds {
		getStatusCode, _ := service_apis.GetInstanceById(instanceEndpoint, token, resourceId)
		if getStatusCode != 404 {
			// Clean up
			logInstance.Println("Remove the instance via DELETE api using resource id")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, resourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Deletion using Id")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, resourceId)
			Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
		} else {
			logInstance.Println("This instance is already deleted")
		}
	}

	if creationSucceeded {
		// Delete ssh key, cloudaccount and vnet if needed
		utils.SuiteCleanup(testEnv, cloudaccountUrl, sshkeyId, token, sshEndpoint, vnetEndpoint, vnet, cloudAccount)
	}
	logInstance.Println("Test teardown completed")
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	utils.ReportGeneration(report)
})
