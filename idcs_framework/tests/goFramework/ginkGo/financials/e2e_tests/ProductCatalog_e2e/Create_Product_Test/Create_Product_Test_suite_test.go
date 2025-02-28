package Create_Product_Test_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// EXEC DELETE PRODUCT COMMAND
func ExecDeleteProductCommand(dirname string) error {
	productDeleteCmd := "kubectl delete -f " + dirname + "/" + "-n idcs-system"
	err := financials_utils.ExecShellCommand(productDeleteCmd)
	return err
}

// EXEC CREATE PRODUCT COMMAND
func ExecCreateProductCommand(dirname string) error {
	productApplyCmd := "kubectl apply -f " + dirname + "/" + " -n idcs-system"
	//productDeleteCmd := "kubectl delete -f " + dirname + "/" + "-n idcs-system"
	err := financials_utils.ExecShellCommand(productApplyCmd)
	return err
}

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
var cloud_account_created string
var place_holder_map = make(map[string]string)
var userToken string
var vnet_created string
var ssh_publickey_name_created string
var create_response_status int
var create_response_body string
var instance_id_created string
var meta_data_map = make(map[string]string)
var resourceInfo testsetup.ResourcesInfo

func init() {

	logger.InitializeZapCustomLogger()
}

var _ = BeforeSuite(func() {
	time.Sleep(20 * time.Second)
	compute_utils.LoadE2EConfig("../../../data", "vmaas_input.json")
	auth.Get_config_file_data("../../../data/config.json")
	financials_utils.LoadE2EConfig("../../../data", "billing.json")
	base_url = compute_utils.GetBaseUrl()
	compute_url = compute_utils.GetComputeUrl()
	userName = auth.Get_UserName("Intel")
	token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	fmt.Print(token)
	ariaclientId = financials_utils.GetAriaClientNo()
	ariaAuth = financials_utils.GetariaAuthKey()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	fmt.Println("ariaAuth", ariaAuth)
	fmt.Println("ariaclientId", ariaclientId)

	By("Starting product create process...")
	error := ExecCreateProductCommand("product")
	fmt.Print(error)
	//Expect(error.Error()).To(BeNil(), "Failed to execute command, error: "+error.Error())
	logger.Logf.Info("Product Created in current context.")
})

var _ = AfterEach(func() {
	token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
})

func TestCreateProductTest(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "CreateProductTest Suite")
}
