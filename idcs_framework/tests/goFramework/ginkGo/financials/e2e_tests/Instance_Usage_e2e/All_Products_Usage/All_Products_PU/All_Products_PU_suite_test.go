package All_Products_PU_test

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

type Product struct {
	Name        string    `json:"name"`
	ID          string    `json:"id"`
	Created     time.Time `json:"created"`
	VendorID    string    `json:"vendorId"`
	FamilyID    string    `json:"familyId"`
	Description string    `json:"description"`
	Metadata    struct {
		BillingEnable string `json:"billingEnable"`
		Category      string `json:"category"`
		Disks         string `json:"disks.size"`
		DisplayName   string `json:"displayName"`
		Desc          string `json:"family.displayDescription"`
		DispName      string `json:"family.displayName"`
		Highlight     string `json:"highlight"`
		Information   string `json:"information"`
		InstanceType  string `json:"instanceType"`
		Memory        string `json:"memory.size"`
		Processor     string `json:"processor"`
		Region        string `json:"region"`
		Service       string `json:"service"`
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
}

type GetProductsResponse struct {
	Products []Product `json:"products"`
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
var failures []string

func init() {

	logger.InitializeZapCustomLogger()
}

var _ = BeforeSuite(func() {
	compute_utils.LoadE2EConfig("../../../../data", "vmaas_input.json")
	auth.Get_config_file_data("../../../../data/config.json")
	financials_utils.LoadE2EConfig("../../../../data", "billing.json")
	base_url = compute_utils.GetBaseUrl()
	compute_url = compute_utils.GetComputeUrl()
	userName = auth.Get_UserName("PremiumV1")
	token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	fmt.Print(token)
	ariaclientId = financials_utils.GetAriaClientNo()
	ariaAuth = financials_utils.GetariaAuthKey()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	fmt.Println("ariaAuth", ariaAuth)
	fmt.Println("ariaclientId", ariaclientId)

	By("Create cloud account...")
	// Generating token wtih payload for cloud account creation with enroll API
	token_response, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + token_response
	cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, true)
	fmt.Println("BODY...", cloudaccount_creation_body)
	Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
	cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
	cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
	Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of Intel user")
	place_holder_map["cloud_account_id"] = cloud_account_created
	place_holder_map["cloud_account_type"] = cloudaccount_type
	fmt.Println("cloudAccount_id", cloud_account_created)
})

func TestAllProductsPU(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "AllProductsPU Suite")
}
