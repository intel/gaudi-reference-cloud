package All_Products_EU_test

import (
	"encoding/json"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
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
var structResponse GetProductsResponse
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

var _ = BeforeSuite(func() {
	compute_utils.LoadE2EConfig("../../../data", "product_catalog_e2e.json")
	auth.Get_config_file_data("../../../data/config.json")
	financials_utils.LoadE2EConfig("../../../data", "billing.json")
	testsetup.Get_config_file_data("../../../data/product_catalog_e2e.json")
	userName = auth.Get_UserName("Enterprise")
	base_url = compute_utils.GetBaseUrl()
	compute_url = compute_utils.GetComputeUrl()
	token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	ariaclientId = financials_utils.GetAriaClientNo()
	ariaAuth = financials_utils.GetariaAuthKey()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)

	By("Delete cloud account...")
	url := base_url + "/v1/cloudaccounts"
	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, token)
	if err == nil {
		financials.DeleteCloudAccountById(url, token, cloudAccId)
	}

	By("Create cloud account...")
	// Generating token wtih payload for cloud account creation with enroll API
	userToken = testsetup.Get_User_Token(userName, "ACCOUNT_TYPE_ENTERPRISE", "300007423870")
	//userToken, _ = auth.Get_Azure_Bearer_Token(userName)
	_, oid, owner, parentid, tid := cloudAccounts.CAcc_RandomPayload_gen()
	get_CAcc_id, create_payload, respCode := cloudAccounts.CreateCloudAccountWithOIDC(userName, oid, owner, parentid, "13564452", tid, true, false, false, false, false,
		true, true, "ACCOUNT_TYPE_ENTERPRISE", 200, token, base_url)
	out, err := json.Marshal(create_payload)
	if err != nil {
		panic(err)
	}
	Expect(respCode).To(Equal(200), "Failed to create Cloud Account")
	cloud_account_created = get_CAcc_id
	cloudaccount_type := gjson.Get(string(out), "type").String()

	By("Validate cloud account...")
	Expect(strings.Contains(string(out), `"type":"ACCOUNT_TYPE_ENTERPRISE"`)).To(BeTrue(), "Account type is not Enterprise, Validation Failed.")
	Expect(strings.Contains(string(out), `"paidServicesAllowed":true`)).To(BeTrue(), "Allowed paid services is not true, Validation Failed.")
	place_holder_map["cloud_account_id"] = cloud_account_created
	place_holder_map["cloud_account_type"] = cloudaccount_type
	billing.CreateBillingAccountWithSpecificCloudAccountIdOIDC(base_url, get_CAcc_id, 200, token)
})

func init() {

	logger.InitializeZapCustomLogger()
}

func TestAllProductsEU(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "AllProductsEU Suite")
}
