package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	auth_admin "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/auth"
	oidc_token "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/auth/oidc_auth"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	storage "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/service_apis"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

// Hardcoded list of general purpose clusters per region
// NOTE: The map key has to match the value in variable test_env.
var GPClusterMap = map[string][]string{
	"us-dev3-1":    {"pdx05-dev-4"},
	"us-staging-1": {"pdx05-dev-4"},
	"us-staging-3": {"pdx11-2"},
	"us-staging-4": {"phx06-1"},
	"us-region-1":  {"pdx07-4"},
	"us-region-2":  {"phx04-5"},
	"us-region-3":  {"phx02-3"},
	"us-region-4":  {"cmh01-2"},
	"us-qa1-1":     {"pdx09-1", "pdx09-2"},
}

func Get_Azure_Bearer_Token_Retry(userEmail string) string {
	var err error
	token := ""
	for i := 0; i < 5; i++ {
		token, err = auth_admin.Get_Azure_Bearer_Token(userEmail)
		Expect(err).NotTo(HaveOccurred())
		if token != "Bearer " {
			break
		}

		logger.Logf.Infof("Try #%d Empty token was returned, sleeping for 5 seconds before retrying token retrieval...", i+1)
		time.Sleep(5 * time.Second)
	}

	Expect(token).ToNot(Equal("Bearer "), "token is empty after multiple retries")
	return token
}

func StTestEnvSetup(test_env, region, authConfigPath, userEmail, userToken string) (string, string, string, string) {
	var storage_url string = "https://internal-placeholder.com"
	var cloudaccount_url string = "https://internal-placeholder.com"

	if userToken != "" {
		logger.Log.Info("Using token provided via command line")
	} else {
		logger.Log.Info("User has not provided a token.  User token will be retrieved via UI.")
	}

	token := userToken
	region = strings.TrimSpace(region)

	var oidc_url string = "https://internal-placeholder.com/token?email=admin@intel.com&groups=IDC.Admin"
	if test_env == "kind" {
		storage_url = "https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local"
		cloudaccount_url = "https://dev.api.cloud.intel.com.kind.local"
		oidc_url = "http://dev.oidc.cloud.intel.com.kind.local/token?email=admin@intel.com&groups=IDC.Admin"

		if token == "" {
			token = oidc_token.GetBearerTokenViaResty(oidc_url)
		}

		return storage_url, cloudaccount_url, oidc_url, token
	} else if test_env == "staging" {
		os.Setenv("proxy_required", "true")
		os.Setenv("https_proxy", "http://internal-placeholder.com:912")
		auth_admin.Get_config_file_data(authConfigPath)

		regionNumber := region[len(region)-1:]
		storage_url = "https://staging-idc-us-" + regionNumber + ".eglb.intel.com"
		cloudaccount_url = "https://staging.api.idcservice.net"

		if token == "" {
			// User didn't specify token so retrieve a user token via the API
			token = Get_Azure_Bearer_Token_Retry(userEmail)
		}

		logger.Logf.Infof("Using storage_url: %s", storage_url)
		logger.Logf.Infof("Using cloudaccount_url: %s", cloudaccount_url)
		return storage_url, cloudaccount_url, "", token

	} else if test_env == "qa1" {
		os.Setenv("proxy_required", "true")
		os.Setenv("https_proxy", "http://internal-placeholder.com:912")
		auth_admin.Get_config_file_data(authConfigPath)

		regionNumber := region[len(region)-1:]
		storage_url = "https://qa1-idc-us-" + regionNumber + ".eglb.intel.com"
		cloudaccount_url = "https://qa1.api.idcservice.net"

		if token == "" {
			// User didn't specify token so retrieve a user token via the API
			token = Get_Azure_Bearer_Token_Retry(userEmail)
		}

		logger.Logf.Infof("Using storage_url: %s", storage_url)
		logger.Logf.Infof("Using cloudaccount_url: %s", cloudaccount_url)
		return storage_url, cloudaccount_url, "", token

	} else if test_env == "production" {
		os.Setenv("proxy_required", "true")
		os.Setenv("https_proxy", "http://internal-placeholder.com:912")
		auth_admin.Get_config_file_data(authConfigPath)
		storage_url = "https://compute-" + region + "-api.cloud.intel.com"
		cloudaccount_url = "https://compute.api.idcservice.net"

		if token == "" {
			// User didn't specify token so retrieve a user token via the API
			token = Get_Azure_Bearer_Token_Retry(userEmail)
		}

		logger.Logf.Infof("Using storage_url: %s", storage_url)
		logger.Logf.Infof("Using cloudaccount_url: %s", cloudaccount_url)
		return storage_url, cloudaccount_url, "", token
	} else if test_env == "dev3" {
		os.Setenv("proxy_required", "true")
		os.Setenv("https_proxy", "http://internal-placeholder.com:912")
		auth_admin.Get_config_file_data(authConfigPath)
		storage_url = "https://internal-placeholder.com"
		cloudaccount_url = "https://dev.api.idcservice.net"

		if token == "" {
			// User didn't specify token so retrieve a user token via the API
			token = Get_Azure_Bearer_Token_Retry(userEmail)
		}

		return storage_url, cloudaccount_url, "", token
	} else {
		storage_url = strings.Replace(storage_url, "dev", test_env, 1)
		cloudaccount_url = strings.Replace(cloudaccount_url, "dev", test_env, 1)
		oidc_url = strings.Replace(oidc_url, "dev", test_env, 1)

		if token == "" {
			token = oidc_token.GetBearerTokenViaResty(oidc_url)
		}

		return storage_url, cloudaccount_url, oidc_url, token
	}
}

func StSuiteSetup(test_env string, token string, cloudaccount_url string, storage_url string,
	account_type string) (string, string, string, string, string, string, string, string, string, string) {
	var cloudAccount string

	if test_env != "kind" {
		cloudAccount = GetStCloudAccount()
	} else {
		username := GetRandomStringWithLimit(10)
		switch account_type {
		case "ACCOUNT_TYPE_INTEL":
			username = username + "@intel.com"

		case "ACCOUNT_TYPE_STANDARD":
			username = username + "@standard.com"

		case "ACCOUNT_TYPE_PREMIUM":
			username = username + "@premium.com"
		default:
			fmt.Println("Invalid account type...")
		}
		cloudAccount = CreateCloudAccount(cloudaccount_url+"/v1/cloudaccounts", token, username, account_type)
	}

	instance_endpoint := storage_url + "/v1/cloudaccounts/" + cloudAccount + "/" + "instances"
	instance_group_endpoint := storage_url + "/v1/cloudaccounts/" + cloudAccount + "/" + "instancegroups"
	storage_endpoint := storage_url + "/v1/cloudaccounts/" + cloudAccount + "/" + "filesystems"
	ssh_endpoint := storage_url + "/v1/cloudaccounts/" + cloudAccount + "/" + "sshpublickeys"
	vnet_endpoint := storage_url + "/v1/cloudaccounts/" + cloudAccount + "/vnets"
	bucket_endpoint := storage_url + "/v1/cloudaccounts/" + cloudAccount + "/" + "objects/buckets"
	principal_endpoint := storage_url + "/v1/cloudaccounts/" + cloudAccount + "/" + "objects/users"
	rule_endpoint := storage_url + "/v1/cloudaccounts/" + cloudAccount + "/" + "objects/buckets/id/"
	user_endpoint := storage_url + "/v1/cloudaccounts/" + cloudAccount + "/" + "filesystems/name/"
	return cloudAccount, instance_endpoint, ssh_endpoint, vnet_endpoint, bucket_endpoint, principal_endpoint, rule_endpoint, user_endpoint, storage_endpoint, instance_group_endpoint
}

func StSuiteDependenciesSetup(test_env string, token string, vnet_url string, bmaasEnabled bool, ssh_url string, ssh_key_value string) (string, string, string) {
	var vnet, ssh_key_name, ssh_key_Id, vnet_payload string
	if test_env == "staging" || test_env == "production" || test_env == "dev3" {
		vnet = GetStVnetName()
	} else {
		vnet_name := "automation-vnet-" + GetRandomString()
		if bmaasEnabled {
			vnet_payload = fmt.Sprintf(`{"metadata": {"name": "%s"},"spec": {"region": "us-dev-1","availabilityZone": "us-dev-1b","prefixLength": 22}}`, vnet_name)
		} else {
			vnet_payload = fmt.Sprintf(`{"metadata": {"name": "%s"},"spec": {"region": "us-dev-1","availabilityZone": "us-dev-1a","prefixLength": 22}}`, vnet_name)
		}
		vnet_response_status, vnet_response_body := storage.CreateVnet(vnet_url, token, vnet_payload)
		Expect(vnet_response_status).To(Equal(200), "Failed to create VNet")
		vnet = gjson.Get(vnet_response_body, "metadata.name").String()
	}

	ssh_key_name = "automation-sshkey-" + GetRandomString() + "@intel.com"
	sshkey_payload := fmt.Sprintf(`{"metadata": {"name": "%s"},"spec": {"sshPublicKey": "%s"}}`, "<<ssh-key-name>>", "<<ssh-user-public-key>>")
	ssh_response_status, ssh_response_body := storage.CreateSSHKey(ssh_url, token, sshkey_payload, ssh_key_name, ssh_key_value)
	Expect(ssh_response_status).To(Equal(200), "assertion failed on response code")
	Expect(strings.Contains(ssh_response_body, `"name":"`+ssh_key_name+`"`)).To(BeTrue(), "assertion failed on response body")
	ssh_key_Id = gjson.Get(ssh_response_body, "metadata.resourceId").String()
	return vnet, ssh_key_name, ssh_key_Id
}

func StSuiteCleanup(test_env string, cloudaccount_url string, sshkeyId string, token string, ssh_endpoint string,
	vnet_endpoint string, vnet string, cloudAccount string) error {
	// ssh key deletion
	sshkey_deletion_status, sshkey_deletion_body := storage.DeleteSSHKeyById(ssh_endpoint, token, sshkeyId)
	Expect(sshkey_deletion_status).To(Equal(200), sshkey_deletion_body)

	// vnet and cloud account deletion
	if test_env == "kind" {
		delete_response_byid_status, delete_response_byid_body := storage.DeleteVnetByName(vnet_endpoint, token, vnet)
		Expect(delete_response_byid_status).To(Equal(200), delete_response_byid_body)

		DeleteCloudAccount(cloudaccount_url+"/v1/cloudaccounts", cloudAccount, token)
	}
	return nil
}

func DeleteCloudAccountForStorage(test_env string, token string, cloudaccount_url string, cloudAccount string) error {
	// cloud account deletion
	if test_env == "kind" {
		DeleteCloudAccount(cloudaccount_url+"/v1/cloudaccounts", cloudAccount, token)
	}
	return nil
}

func IsGPCluster(key, value string) bool {
	clusters, ok := GPClusterMap[key]
	if !ok {
		return false // Key does not exist
	}
	for _, cluster := range clusters {
		if cluster == value {
			return true // Cluster the friend
		}
	}
	return false // Cluster not found
}

func GetRandomStringWithLimit(n int) string {
	var charset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321")
	rand.Seed(time.Now().UnixNano())
	str := make([]rune, n)
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}
	return string(str)
}

// cloud account creation util without service api
func CreateCloudAccount(cloudaccount_url string, token string, username string, account_type string) string {
	tid := GetRandomStringWithLimit(12)
	oid := GetRandomStringWithLimit(12)
	name := "compute-cloudaccount-" + GetRandomString()
	cloudaccount_payload := fmt.Sprintf(`{"name":"%s","owner":"%s","tid":"%s","oid":"%s","type":"%s"}`, name, username, tid, oid, account_type)
	var ca_payload_map map[string]interface{}
	json.Unmarshal([]byte(cloudaccount_payload), &ca_payload_map)
	ca_resty_response := client.Post(cloudaccount_url, token, ca_payload_map)
	cloudaccount_creation_status, cloudaccount_creation_body := client.LogRestyInfo(ca_resty_response, "POST API")
	Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account")
	cloudAccount := gjson.Get(cloudaccount_creation_body, "id").String()
	return cloudAccount
}

func DeleteCloudAccount(cloudaccount_url string, cloudAccountId string, token string) {
	url := fmt.Sprintf("%s/id/%s", cloudaccount_url, cloudAccountId)
	logger.Log.Info("Deleting the cloud account with id: " + cloudAccountId)
	ca_resty_response := client.Delete(url, token)
	cloudaccount_creation_status, _ := client.LogRestyInfo(ca_resty_response, "DELETE API")
	Expect(cloudaccount_creation_status).To(Equal(200), "Failed to delete Cloud Account")
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
		tc_name := strings.Join(specReport.ContainerHierarchyTexts, "-")
		tc_status := specReport.State.String()
		var failure_reason string

		if tc_status == "failed" {
			tmpl = `<tr style="color: red"><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`
			failure_reason = specReport.FailureMessage()
		} else if tc_status == "skipped" {
			tmpl = `<tr style="color: orange"><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`
			failure_reason = "NA"
		} else {
			failure_reason = "NA"
		}

		if tc_name != "" {
			output := fmt.Sprintf(tmpl, strconv.Itoa(index), tc_name, tc_status, failure_reason)
			fmt.Fprintf(f, output)
		}

	}
	fmt.Fprintf(f, `</table></body></html>`)
	f.Close()
}
