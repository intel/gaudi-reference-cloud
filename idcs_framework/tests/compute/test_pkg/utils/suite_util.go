package utils

import (
	"fmt"
	old_auth_token "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/auth"
	oidc_token "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/auth/oidc_auth"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"os"
	"strconv"
	"strings"
)

func getComputeBaseUrl(testEnv string) string {
	switch testEnv {
	case "dev3":
		return "https://dev3-compute-us-dev3-#-api-cloud.eglb.intel.com"
	case "staging":
		return "https://internal-placeholder.com"
	case "qa1":
		return "https://internal-placeholder.com"
	case "prod":
		return "https://compute-us-region-#-api.cloud.intel.com"
	default:
		return "https://dev.compute.us-dev-#.api.cloud.intel.com.kind.local"
	}
}

func replaceRegion(url string, region string) string {
	if strings.Contains(url, "#") {
		parts := strings.Split(region, "-")
		// Extract the last part as the region number
		regionNumber := parts[len(parts)-1]
		return strings.Replace(url, "#", regionNumber, 1)
	}
	return url
}

func TestEnvSetup(testEnv, region, userEmail, authConfigPath, userTokenFlag string) (string, string, string, string) {
	var (
		userToken       string
		computeUrl      string
		err             error
		baseComputeUrl  string = getComputeBaseUrl(testEnv)
		cloudaccountUrl string = "https://dev-api.idcs-dev.intel.com"
		oidcUrl         string = "https://dev-oidc.idcs-dev.intel.com/token?email=admin@intel.com&groups=IDC.Admin"
	)
	if testEnv == "staging" || testEnv == "qa1" || testEnv == "dev3" || testEnv == "prod" {
		err = os.Setenv("proxy_required", "true")
		if err != nil {
			logInstance.Println(err.Error())
		}
		err = os.Setenv("https_proxy", "http://internal-placeholder.com:912")
		if err != nil {
			logInstance.Println(err.Error())
		}
	}
	if testEnv == "kind" {
		computeUrl = replaceRegion(baseComputeUrl, region)
		cloudaccountUrl = "https://dev.api.cloud.intel.com.kind.local"
		oidcUrl = "http://dev.oidc.cloud.intel.com.kind.local/token?email=admin@intel.com&groups=IDC.Admin"
		userToken = oidc_token.GetBearerTokenViaResty(oidcUrl)
	} else if testEnv == "staging" || testEnv == "qa1" || testEnv == "dev3" {
		old_auth_token.Get_config_file_data(authConfigPath)
		computeUrl = replaceRegion(baseComputeUrl, region)
		if testEnv == "staging" {
			cloudaccountUrl = "https://staging.api.idcservice.net"
		} else if testEnv == "qa1" {
			cloudaccountUrl = "https://qa1.api.idcservice.net"
		} else if testEnv == "dev3" {
			cloudaccountUrl = "https://dev.api.idcservice.net"
		}
	} else if testEnv == "prod" {
		computeUrl = replaceRegion(baseComputeUrl, region)
		cloudaccountUrl = "https://api.idcservice.net"
	} else {
		devBaseUrl := "https://dev-compute-us-dev-1-api.idcs-dev.intel.com"
		computeUrl = strings.Replace(devBaseUrl, "dev", testEnv, 1)
		cloudaccountUrl = strings.Replace(cloudaccountUrl, "dev", testEnv, 1)
		oidcUrl = strings.Replace(oidcUrl, "dev", testEnv, 1)
		userToken = oidc_token.GetBearerTokenViaResty(oidcUrl)
	}

	if testEnv == "staging" || testEnv == "dev3" || testEnv == "qa1" || testEnv == "prod" {
		if userTokenFlag != "" {
			userToken = "Bearer " + strings.TrimSpace(userTokenFlag)
		} else {
			userToken, err = old_auth_token.Get_Azure_Bearer_Token(userEmail)
			Expect(err).ShouldNot(HaveOccurred(), err)
		}
	}
	return computeUrl, cloudaccountUrl, oidcUrl, userToken
}

func FetchAdminToken(testEnv string) string {
	var (
		adminToken string
		oidcUrl    string = "https://dev-oidc.idcs-dev.intel.com/token?email=admin@intel.com&groups=IDC.Admin"
	)
	if testEnv == "kind" {
		oidcUrl = "http://dev.oidc.cloud.intel.com.kind.local/token?email=admin@intel.com&groups=IDC.Admin"
		adminToken = oidc_token.GetBearerTokenViaResty(oidcUrl)
		return adminToken
	} else if testEnv == "staging" || testEnv == "dev3" || testEnv == "qa1" {
		adminToken = old_auth_token.Get_Azure_Admin_Bearer_Token(testEnv)
		return adminToken
	} else if testEnv == "prod" {
		adminToken = ""
		return adminToken
	} else {
		oidcUrl = strings.Replace(oidcUrl, "dev", testEnv, 1)
		adminToken = oidc_token.GetBearerTokenViaResty(oidcUrl)
		return adminToken
	}
}

func SuiteSetup(testEnv string, token string, cloudaccountUrl string,
	computeUrl string, accountType string, cloudAccountId string) (string, string, string, string, string, string) {
	var cloudAccount string
	if testEnv == "staging" || testEnv == "dev3" || testEnv == "prod" || testEnv == "qa1" {
		logInstance.Println("Fetching cloudAccountId from the command..")
		cloudAccount = cloudAccountId
	} else {
		username := GetRandomStringWithLimit(10)
		switch accountType {
		case "ACCOUNT_TYPE_INTEL":
			username = username + "@intel.com"

		case "ACCOUNT_TYPE_STANDARD":
			username = username + "@standard.com"

		case "ACCOUNT_TYPE_PREMIUM":
			username = username + "@premium.com"
		default:
			logInstance.Println("Invalid account type...")
		}
		cloudAccount = CreateCloudAccount(cloudaccountUrl+"/v1/cloudaccounts", token, username, accountType)
	}
	logInstance.Println("Cloud account is : ", cloudAccount)

	instanceEndpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccount + "/" + "instances"
	sshEndpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccount + "/" + "sshpublickeys"
	vnetEndpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccount + "/vnets"
	instanceTypeEndpoint := computeUrl + "/v1/instancetypes"
	machineImageEndpoint := computeUrl + "/v1/machineimages"
	return cloudAccount, instanceEndpoint, sshEndpoint, vnetEndpoint, instanceTypeEndpoint, machineImageEndpoint
}

func SuiteDependenciesSetup(testEnv string, token string, vnetUrl string, bmaasEnabled bool,
	sshUrl string, sshKeyPath string, vnetName string) (string, string, string) {
	var vnet, sshKeyName, sshKeyId, vnetPayload string
	if testEnv == "staging" || testEnv == "dev3" || testEnv == "prod" || testEnv == "qa1" {
		logInstance.Println("Fetching vnetName from the command..")
		vnet = vnetName
	} else {
		vnetName := "automation-vnet-" + GetRandomString()
		if bmaasEnabled {
			vnetPayload = fmt.Sprintf(`{"metadata": {"name": "%s"},"spec": {"region": "us-dev-1","availabilityZone": "us-dev-1b","prefixLength": 22}}`, vnetName)
		} else {
			vnetPayload = fmt.Sprintf(`{"metadata": {"name": "%s"},"spec": {"region": "us-dev-1","availabilityZone": "us-dev-1a","prefixLength": 22}}`, vnetName)
		}
		statusCode, responseBody := service_apis.CreateVnetWithCustomizedPayload(vnetUrl, token, vnetPayload)
		Expect(statusCode).To(Equal(200), "Failed to create VNet: %v", responseBody)
		vnet = gjson.Get(responseBody, "metadata.name").String()
	}

	// SSH Key creation
	sshKeyName = "automation-sshkey-" + GetRandomString() + "@intel.com"
	logInstance.Println("Fetching SSH key from the given path...")
	sshKeyValue, err := ReadPublicKey(sshKeyPath)
	Expect(err).Should(Succeed(), "Couldn't read the SSH public key from the specified path."+sshKeyPath)
	sshPayload := fmt.Sprintf(`{"metadata": {"name": "%s"},"spec": {"sshPublicKey": "%s"}}`, "<<ssh-key-name>>", "<<ssh-user-public-key>>")
	statusCode, responseBody := service_apis.CreateSSHKey(sshUrl, token, sshPayload, sshKeyName, sshKeyValue)
	Expect(statusCode).To(Equal(200), responseBody)
	Expect(strings.Contains(responseBody, `"name":"`+sshKeyName+`"`)).To(BeTrue(), "Failed to create SSH Key: %v", responseBody)
	sshKeyId = gjson.Get(responseBody, "metadata.resourceId").String()
	return vnet, sshKeyName, sshKeyId
}

// MI - machine image & IT - Instance type
func CreateMIAndITInputMap(resultArray []gjson.Result) map[string]string {
	mapping := make(map[string]string)

	// result.ForEach(func(_, value gjson.Result) bool {
	for _, value := range resultArray {
		creationType := value.Get("instanceType").String()
		image := value.Get("machineImage").String()
		mapping[creationType] = image
		// return true
	}

	return mapping
}

func SuitelevelLogSetup(logInstance *logger.CustomLogger) {
	SetLogger(logInstance)
	client.SetLogger(logInstance)
	old_auth_token.SetLogger(logInstance)
	service_apis.SetLogger(logInstance)
}

func SuiteCleanup(testEnv string, cloudaccountUrl string, sshkeyId string, token string, sshEndpoint string,
	vnetEndpoint string, vnet string, cloudAccount string) error {
	// ssh key deletion
	statusCode, responseBody := service_apis.DeleteSSHKeyById(sshEndpoint, token, sshkeyId)
	Expect(statusCode).To(Equal(200), responseBody)

	// vnet and cloud account deletion
	if testEnv != "staging" && testEnv != "dev3" && testEnv != "prod" && testEnv != "qa1" {
		deleteStatusCode, responseBody := service_apis.DeleteVnetByName(vnetEndpoint, token, vnet)
		Expect(deleteStatusCode).To(Or(Equal(404), Equal(200), Equal(400)), responseBody)

		DeleteCloudAccount(cloudaccountUrl+"/v1/cloudaccounts", cloudAccount, token)
	}
	return nil
}

func ReportGeneration(report Report) {
	f, _ := os.Create("report.html")
	for index, specReport := range report.SpecReports {
		var tmpl = `<tr style="color: green"><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`

		if index == 0 {
			fmt.Fprintf(f, `<html><head><style>table, th, td {border: 1px solid black;}</style></head><body>
			<u><b><p>Test suite Details   : `+strings.Join(strings.Split(report.SuiteDescription, ""), "")+`</p></b></u>
			<p>Execution-Start-Time : `+report.StartTime.UTC().String()+`</p>
			<p>Execution-End-Time   : `+report.EndTime.UTC().String()+`</p>
			<p>Total-Execution-Time : `+report.RunTime.String()+`</p>
			<table><tr><th>Sl.no</th><th>Testcase-Name</th><th>Execution-Status</th><th>Failure-Reason</th></tr>`)
		}
		tcName := strings.Join(specReport.ContainerHierarchyTexts, "-")
		tcStatus := specReport.State.String()
		var failureReason string

		if tcStatus == "failed" || tcStatus == "panicked" || tcStatus == "timedout" {
			tmpl = `<tr style="color: red"><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`
			failureReason = specReport.FailureMessage()
		} else if tcStatus == "skipped" {
			tmpl = `<tr style="color: orange"><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`
			failureReason = "NA"
		} else if tcStatus == "pending" {
			tmpl = `<tr style="color: blue"><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`
			failureReason = "NA"
		} else {
			failureReason = "NA"
		}

		if tcName != "" {
			output := fmt.Sprintf(tmpl, strconv.Itoa(index), tcName, tcStatus, failureReason)
			fmt.Fprintf(f, output)
		}

	}
	fmt.Fprintf(f, `</table></body></html>`)
	f.Close()
}

func GetAllImagesForInstanceType(instanceType string) ([]string, error) {
	imageMapping, err := GetBUAllImagesMapping()
	if err != nil {
		return nil, err
	}

	// Retrieve the images for the specified account type
	images, ok := imageMapping[instanceType]
	if !ok {
		logInstance.Println("instance type not found in the mapping")
		return nil, nil
	}
	return images, nil
}

func SkipTests(environment string) string {
	// Determine the behavior of `skipBMCreation` based on the environment
	if environment == "qa1" {
		return "false"
	} else {
		return "true"
	}
}
