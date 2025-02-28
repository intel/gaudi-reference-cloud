package quota_management

import (
	"flag"
	"testing"
	// "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	intelInstanceEndpoint, intelVnetEndpoint, intelSSHEndpoint, intelCloudAccount, intelVnet, intelSSHKeyName, intelUsername                      string
	standardInstanceEndpoint, standardVnetEndpoint, standardSSHEndpoint, standardCloudAccount, standardVnet, standardSSHKeyName, standardUsername string
	premiumInstanceEndpoint, premiumVnetEndpoint, premiumSSHEndpoint, premiumCloudAccount, premiumVnet, premiumSSHKeyName, premiumUsername        string
)

// flag
var (
	skipVMCreation                     string
	region, cloudAccountId, vnetName   string
	creationSucceeded                  bool
	sshPublicKey, userEmail, userToken string
)

// suite level variables
var (
	computeUrl       string
	cloudaccountUrl  string
	oidcUrl          string
	testEnv          string
	token            string
	availabilityZone string
	logInstance      *logger.CustomLogger
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

func TestBusinessSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IDC - VMaaS Quota Enforcement Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	logger.InitializeLogger(9)
	logInstance = logger.GetLogger()

	utils.SuitelevelLogSetup(logInstance)

	logInstance.Println("Running setup on Node 9")
	logInstance.Println("Starting VMaaS Quota enforcement Business Usecase test suite")

	if skipVMCreation == "true" {
		Skip("Skipping the Entire VM flow due to the flag")
	}

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../../resources", "vmaas_business_input.json")

	// update the configmap and restart pod (kubeconfig file for env should be provided - if it is not kind cluster)
	/*
		client.SetUpKubeClient("<<kubeconfig>>", testEnv)
		Expect(client.UpdateQuotaConfigMap("us-dev-1-compute-api-server", "config.yaml")).To(Equal(true))
		Expect(client.RestartPod("us-dev-1-compute-api-server-")).To(Equal(true))
	*/

	// Get dependencies
	computeUrl, cloudaccountUrl, oidcUrl, token = utils.TestEnvSetup(testEnv, region, userEmail, "../../resources/auth_config.json", userToken)

	if testEnv == "staging" || testEnv == "dev3" {
		intelCloudAccount = utils.GetJsonValue("intelcloudAccount")
		standardCloudAccount = utils.GetJsonValue("standardCloudAccount")
		premiumCloudAccount = utils.GetJsonValue("premiumCloudAccount")
		intelVnet = utils.GetJsonValue("vnetName")
		standardVnet = utils.GetJsonValue("vnetName")
		premiumVnet = utils.GetJsonValue("vnetName")
	} else {
		// intel cloud-account creation
		intelCloudAccount, intelInstanceEndpoint, intelVnetEndpoint, intelSSHEndpoint, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl,
			computeUrl, "ACCOUNT_TYPE_INTEL", intelCloudAccount)

		// standard cloud-account creation
		standardCloudAccount, standardInstanceEndpoint, standardVnetEndpoint, standardSSHEndpoint, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl,
			computeUrl, "ACCOUNT_TYPE_STANDARD", standardCloudAccount)

		// premium cloud-account creation
		premiumCloudAccount, premiumInstanceEndpoint, premiumVnetEndpoint, premiumSSHEndpoint, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl,
			computeUrl, "ACCOUNT_TYPE_PREMIUM", premiumCloudAccount)
	}

	// ssh key creation
	intelVnet, intelSSHKeyName, _ = utils.SuiteDependenciesSetup(testEnv, token, intelVnetEndpoint, false, intelSSHEndpoint, sshPublicKey, intelVnet)
	standardVnet, standardSSHKeyName, _ = utils.SuiteDependenciesSetup(testEnv, token, standardVnetEndpoint, false, standardSSHEndpoint, sshPublicKey, standardVnet)
	premiumVnet, premiumSSHKeyName, _ = utils.SuiteDependenciesSetup(testEnv, token, premiumVnetEndpoint, false, premiumSSHEndpoint, sshPublicKey, premiumVnet)

	// populate availability zone
	availabilityZone = utils.GetAvailabilityZone(region, false)

	creationSucceeded = true
	return []byte(token)
}, func() {

})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	if creationSucceeded {
		// revert the configmap and restart pod
		/*
			client.SetUpKubeClient("<<kubeconfig>>", testEnv)
			Expect(client.RevertQuotaConfigMap("us-dev-1-compute-api-server", "config.yaml")).To(Equal(true))
			Expect(client.RestartPod("us-dev-1-compute-api-server-")).To(Equal(true))
		*/

		// ssh key deletion
		account_ids := []string{intelCloudAccount, standardCloudAccount, premiumCloudAccount}
		ssh_names := []string{intelSSHKeyName, standardSSHKeyName, premiumSSHKeyName}
		utils.DeleteMultiSSHKeys(computeUrl+"/v1/cloudaccounts/", token, account_ids, ssh_names)

		// vnet and cloud account deletion
		if testEnv != "staging" && testEnv != "dev3" {
			// vnet deletion
			vnets_names := []string{intelVnet, standardVnet, premiumVnet}
			utils.DeleteMultiVNets(computeUrl+"/v1/cloudaccounts/", token, account_ids, vnets_names)

			// cloud account deletion
			utils.DeleteCloudAccount(cloudaccountUrl+"/v1/cloudaccounts", intelCloudAccount, token)
			utils.DeleteCloudAccount(cloudaccountUrl+"/v1/cloudaccounts", standardCloudAccount, token)
			utils.DeleteCloudAccount(cloudaccountUrl+"/v1/cloudaccounts", premiumCloudAccount, token)
		}
	}
	logInstance.Println("Test teardown completed")
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	utils.ReportGeneration(report)
})
