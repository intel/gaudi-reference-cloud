package STaaS_5GB_PU_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

type GetProductsResponse struct {
	Products []struct {
		Name        string    `json:"name"`
		ID          string    `json:"id"`
		Created     time.Time `json:"created"`
		VendorID    string    `json:"vendorId"`
		FamilyID    string    `json:"familyId"`
		Description string    `json:"description"`
		Metadata    struct {
			Category     string `json:"category"`
			Disks        string `json:"disks.size"`
			DisplayName  string `json:"displayName"`
			Desc         string `json:"family.displayDescription"`
			DispName     string `json:"family.displayName"`
			Highlight    string `json:"highlight"`
			Information  string `json:"information"`
			InstanceType string `json:"instanceType"`
			Memory       string `json:"memory.size"`
			Processor    string `json:"processor"`
			Region       string `json:"region"`
			Service      string `json:"service"`
		} `json:"metadata"`
		Eccn      string `json:"eccn"`
		Pcq       string `json:"pcq"`
		MatchExpr string `json:"matchExpr"`
		Rates     []struct {
			AccountType string `json:"accountType"`
			Rate        string `json:"rate"`
			Unit        string `json:"unit"`
			UsageExpr   string `json:"usageExpr"`
		} `json:"rates"`
	} `json:"products"`
}

var base_url string
var token string
var userName string
var cloudAccountType string
var cloudAccIdid string
var sshPublicKey string
var compute_url string
var ariaAuth string
var ariaclientId string
var userToken string
var cloud_account_created string
var place_holder_map = make(map[string]string)
var vnet_created string
var ssh_publickey_name_created string
var create_response_status int
var create_response_body string
var instance_id_created string
var meta_data_map = make(map[string]string)
var resourceInfo testsetup.ResourcesInfo
var staas_payload string

func init() {

	logger.InitializeZapCustomLogger()
}

var _ = BeforeSuite(func() {
	compute_utils.LoadE2EConfig("../../../data", "staas_input.json")
	auth.Get_config_file_data("../../../data/config.json")
	financials_utils.LoadE2EConfig("../../../data", "billing.json")
	base_url = compute_utils.GetBaseUrl()
	compute_url = compute_utils.GetComputeUrl()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	ariaclientId = financials_utils.GetAriaClientNo()
	ariaAuth = financials_utils.GetariaAuthKey()
	staas_payload = compute_utils.GetStaaSPayload("5GB", "")
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	fmt.Println("ariaAuth", ariaAuth)
	fmt.Println("ariaclientId", ariaclientId)
	fmt.Println("STaaS payload: ", staas_payload)

	By("Create cloud account...")
	// Generating token wtih payload for cloud account creation with enroll API
	userName = auth.Get_UserName("Premium")
	token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	fmt.Print("USERNAME..................", userName)
	token_response, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + token_response
	cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, true)
	fmt.Println("BODY...", cloudaccount_creation_body)
	Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
	cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
	cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
	Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Account type is not Premium, Validation Failed.")
	place_holder_map["cloud_account_id"] = cloud_account_created
	place_holder_map["cloud_account_type"] = cloudaccount_type
	fmt.Println("cloudAccount_id", cloud_account_created)
})

func TestSTaaS5GBPU(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "STaaS5GBPU Suite")
}

// Delete volume created....
var _ = AfterSuite(func() {
	//DELETE THE INSTANCE CREATED
	By("Delete FileSystem Created...")
	if place_holder_map["resource_id"] != "" {
		code, body := financials.DeleteFileSystemByResourceId(compute_url, token, place_holder_map["cloud_account_id"], place_holder_map["resource_id"])
		fmt.Println("DeleteFileSystemByResourceId Response: ", code, body)
		Expect(code).To(Equal(200), "Failed to delete FileSystem")
	}
})
