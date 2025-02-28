package STaaS_Partial_Credits_SU_test

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

var (
	base_url                   string
	token                      string
	userName                   string
	cloudAccountType           string
	cloudAccIdid               string
	sshPublicKey               string
	compute_url                string
	ariaAuth                   string
	ariaclientId               string
	userToken                  string
	cloud_account_created      string
	place_holder_map           = make(map[string]string)
	vnet_created               string
	ssh_publickey_name_created string
	create_response_status     int
	instance_id_created        string
	create_response_body       string
	meta_data_map              = make(map[string]string)
	resourceInfo               testsetup.ResourcesInfo
	staas_payload              string
)

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
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	staas_payload = compute_utils.GetStaaSPayload("2000GB", "")
	fmt.Println("ariaAuth", ariaAuth)
	fmt.Println("ariaclientId", ariaclientId)
	fmt.Println("STaaS payload: ", staas_payload)

	By("Create cloud account...")
	// Generating token wtih payload for cloud account creation with enroll API
	userName = auth.Get_UserName("Standard")
	token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	fmt.Print("USERNAME..................", userName)
	token_response, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + token_response
	cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, false)
	fmt.Println("BODY...", cloudaccount_creation_body)
	Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
	cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
	cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
	Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_STANDARD"`)).To(BeTrue(), "Account type is not Standard, Validation Failed.")
	place_holder_map["cloud_account_id"] = cloud_account_created
	place_holder_map["cloud_account_type"] = cloudaccount_type
	fmt.Println("cloudAccount_id", cloud_account_created)
})

func TestSTaaSPartialCreditsSU(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "STaaSPartialCreditsSU Suite")
}

var _ = AfterSuite(func() {
	//DELETE THE INSTANCE CREATED
	By("Delete FileSystem Created...")
	if place_holder_map["resource_id"] != "" {
		code, body := financials.DeleteFileSystemByResourceId(compute_url, token, place_holder_map["cloud_account_id"], place_holder_map["resource_id"])
		fmt.Println("DeleteFileSystemByResourceId Response: ", code, body)
		Expect(code).To(Equal(200), "Failed to delete FileSystem")
	}
})
