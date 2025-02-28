package loadbalancer

import (
	"flag"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"testing"
	"time"
)

// endpoints
var (
	instanceEndpoint   string
	lbInstanceEndpoint string
	sshEndpoint        string
	vnetEndpoint       string
)

// Across test suite level variables
var (
	cloudAccount     string
	vnet             string
	sshkeyName       string
	sshkeyId         string
	token            string
	availabilityZone string
	monitorTypes     []gjson.Result
)

// flags
var (
	vnetName       string
	cloudAccountId string
	userEmail      string
	userToken      string
	region         string
	sshPublicKey   string
)

// suite level variables
var (
	computeUrl      string
	cloudaccountUrl string
	testEnv         string
)

func init() {
	flag.StringVar(&sshPublicKey, "sshPublicKey", "", "")
	flag.StringVar(&cloudAccountId, "cloudAccountId", "", "")
	flag.StringVar(&vnetName, "vnetName", "", "")
	flag.StringVar(&testEnv, "testEnv", "", "")
	flag.StringVar(&userEmail, "userEmail", "", "")
	flag.StringVar(&userToken, "userToken", "", "")
	flag.StringVar(&region, "region", "us-dev-1", "")
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	logger.Init()
	RunSpecs(t, "IDC - Compute Regression Suite")
}

var _ = BeforeSuite(func() {
	logger.Init()
	logger.Log.Info("Starting Business use case test suite for LB")

	// load the test data and populate the test suite data into objects
	utils.LoadConfig("../../resources", "lb_business_input.json")

	// Test Environment set up
	computeUrl, cloudaccountUrl, _, token = utils.TestEnvSetup(testEnv, region, userEmail, "../../resources/auth_config.json", userToken)

	// Suite setup
	cloudAccount, instanceEndpoint, sshEndpoint, vnetEndpoint, _, _ = utils.SuiteSetup(testEnv, token, cloudaccountUrl, computeUrl, "ACCOUNT_TYPE_INTEL", cloudAccountId)

	// Create VNET and SSH public key - Suite Dependencies
	vnet, sshkeyName, sshkeyId = utils.SuiteDependenciesSetup(testEnv, token, vnetEndpoint, false, sshEndpoint, sshPublicKey, vnetName)

	// populate availability zone
	availabilityZone = utils.GetAvailabilityZone(region, false)
	lbInstanceEndpoint = computeUrl + "/v1/cloudaccounts/" + cloudAccount + "/loadbalancers"
	monitorTypes = utils.GetJsonArray("lbMonitorType")
})

var _ = AfterSuite(func() {
	time.Sleep(10 * time.Second)
	// Delete ssh key, cloudaccount and vnet if needed
	utils.SuiteCleanup(testEnv, cloudaccountUrl, sshkeyId, token, sshEndpoint, vnetEndpoint, vnet, cloudAccount)

})

var _ = ReportAfterSuite("custom report", func(report Report) {
	utils.ReportGeneration(report)
})
