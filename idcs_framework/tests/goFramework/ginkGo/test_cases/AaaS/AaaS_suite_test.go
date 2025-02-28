package AaaS_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var base_url string
var token string
var userName string
var userNameSU string
var cloudAccountType string
var cloudAccIdid string
var sshPublicKey string
var compute_url string
var ariaAuth string
var ariaclientId string
var userToken string
var userTokenSU string
var cloud_account_created string
var place_holder_map = make(map[string]string)
var place_holder_map_su = make(map[string]string)
var pg_pass string
var expirationDate string
var invitationMessage string
var apiKey string
var inboxIdStandard string
var inboxIdPremium string
var vnet_created string
var ssh_publickey_name_created string
var create_response_status int
var create_response_body string
var instance_id_created string
var user_role_id string
var user_role_id_su string
var permission_role_id string

//var meta_data_map = make(map[string]string)
//var resourceInfo testsetup.ResourcesInfo

func init() {
	logger.InitializeZapCustomLogger()
}

var _ = BeforeSuite(func() {
	compute_utils.LoadE2EConfig("../../financials/data", "vmaas_input.json")
	auth.Get_config_file_data("../../financials/data/config.json")
	financials_utils.LoadE2EConfig("../../financials/data", "billing.json")
	userName = auth.Get_UserName("Premium")
	fmt.Println(userName)
	base_url = compute_utils.GetBaseUrl()
	compute_url = compute_utils.GetComputeUrl()
	token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	fmt.Print(token)
	ariaclientId = financials_utils.GetAriaClientNo()
	ariaAuth = financials_utils.GetariaAuthKey()
	pg_pass = financials_utils.GetCloudAccDBPAssword()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	expirationDate = financials_utils.GetExpirationDate()
	invitationMessage = financials_utils.GetInvitationMessage()
	apiKey = financials_utils.GetMailSlurpKey()
	inboxIdPremium = financials_utils.GetInboxIdPremium()
	inboxIdStandard = financials_utils.GetInboxIdEnterprise()
	fmt.Println("ariaAuth", ariaAuth)
	fmt.Println("ariaclientId", ariaclientId)
	By("Create cloud account premium...")
	// Generating token wtih payload for cloud account creation with enroll API
	token_response, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + token_response
	cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, true)
	fmt.Println("BODY...", cloudaccount_creation_body)
	Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
	cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
	cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
	Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of Premium user")
	place_holder_map["cloud_account_id"] = cloud_account_created
	place_holder_map["cloud_account_type"] = cloudaccount_type
	fmt.Println("cloudAccount_id- for premium user", cloud_account_created)

	By("Create cloud account standard...")
	// Generating token wtih payload for cloud account creation with enroll API
	userNameSU = auth.Get_UserName("Enterprise")
	token_response, _ = auth.Get_Azure_Bearer_Token(userNameSU)
	userTokenSU = "Bearer " + token_response
	cloudaccount_creation_status, cloudaccount_creation_body = financials.CreateCloudAccountE2E(base_url, token, userTokenSU, userNameSU, false)
	fmt.Println("BODY...", cloudaccount_creation_body)
	Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
	cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
	cloudaccount_type = gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
	Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_STANDARD"`)).To(BeTrue(), "Validation failed on Enrollment of Standard user")
	place_holder_map_su["cloud_account_id"] = cloud_account_created
	place_holder_map_su["cloud_account_type"] = cloudaccount_type
	fmt.Println("cloudAccount_id for standard user", cloud_account_created)
})

func TestAaaS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AaaS Suite")
}

var _ = AfterSuite(func() {
	// DELETE SSH KEYS
	By("Delete the SSH key created")
	logger.Logf.Info("Delete SSH keys...")
	fmt.Println("Delete the SSH-Public-Key Created...")
	ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/" + "sshpublickeys"
	delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, userToken, ssh_publickey_name_created)
	Expect(delete_response_byname_status).To(Or(Equal(200), Equal(503)), "assert ssh-public-key deletion response code")
	// DELETE FileStorage
	By("Delete the FS created")
	logger.Logf.Info("Delete FileStorage...")
	fmt.Println("Delete the FileStorage Created...")
	fs_delete_response_code, fs_delete_response_body := financials.DeleteFileSystemByResourceId(compute_url, userToken, place_holder_map["cloud_account_id"], place_holder_map["resource_id_fileSystem"])
	logger.Log.Info("FS delete response: " + fs_delete_response_body)
	Expect(fs_delete_response_code).To(Or(Equal(200), Equal(404), Equal(503)), "Failed removing created FileStorage...")
	// DELETE ObjectStorage
	By("Delete the OS created")
	logger.Logf.Info("Delete ObjectStorage...")
	fmt.Println("Delete the ObjectStorage Created...")
	os_delete_response_code, os_delete_response_body := financials.DeleteObjectStorageByResourceId(compute_url, userToken, place_holder_map["cloud_account_id"], place_holder_map["resource_id_objectStorage"])
	logger.Log.Info("OS delete response: " + os_delete_response_body)
	Expect(os_delete_response_code).To(Or(Equal(200), Equal(404), Equal(503)), "Failed removing created ObjectStorage...")
	//Remove user from Multi User flow
	By("Admin Revokes invitation...")
	fmt.Println("base_url: ", base_url)
	code, body := financials.RemoveInvitation(base_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
	Expect(code).To(Or(Equal(200), Equal(404)), "Failed to reject invite."+body)
})
