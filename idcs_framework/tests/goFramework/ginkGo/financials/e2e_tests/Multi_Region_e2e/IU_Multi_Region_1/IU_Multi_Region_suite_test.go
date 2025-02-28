package IU_Multi_Region_test

import (
	"flag"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"os"
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
var regionName string

func init() {

	flag.StringVar(&sshPublicKey, "sshPublicKey", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCbnYTaUpeo7bsx6pgWhf1IDymryvU7UYVOqfxv5F0Lbpg4osD8bpuopOvVvrMGEdzhxv8Kzr5VaXi8689lPaAXf4Ltx6SFYAQwx+ZdzyqLA73vWftBoTuFvqVe+/IkZ/Y7Vyw+FFZqGOJSc40dLTQ7JfoKV896x5sllP5cO995T+a6R0krX/t00f+VjAypiiK5zWIVQwbGCq4x8upgB4RPeHyQUxMeRzLZAgbEqJBTr5pnZTZjLsPKrhvDp8FRdoUhGMwr6k+pfKL4ZV9T99c0pols5xZBMnreiugPDPt6/w2zXoE/Wa3vXawYZBnt1T0iW5SFJCab85bP/8PkLPRHWtGTttZat9zKWrztVcuG/AaonNi7xtHA6AyAnNs2FnpQOx5VTaMgP2f2l95b8Gg1RN2+5NnT2yxPIN2uCIiTHHjLHxLlpg6rQUs5wUzazpjsj0vfMYTb48d8lOimBJJaMb1iz6DhNtIC9nm8mYRrMgXytMHvUSg+/pxXROaS/zMYdE1dH/PUlnWSW5P3phZ++z1RPVZc7U8k7bNOdHLqXfrAqP+hs6o+9CLfRxOGGQP0jDiYe0S+K8TAt+iiZMxGw5xOa9yItYapZj+zzSaHfLudSvzFjlC+4PY9hIkHvqU08XFzgAEtZh4fkL9HN69Ubt2NrR2Xje8YFYvb0q0d1w== sdp@internal-placeholder.com", "")
	logger.InitializeZapCustomLogger()
}

var _ = BeforeSuite(func() {
	compute_utils.LoadE2EConfig("../../../data", "multi_region.json")
	auth.Get_config_file_data("../../../data/config.json")
	financials_utils.LoadE2EConfig("../../../data", "multi_region.json")
	testsetup.Get_config_file_data("../../../data/multi_region.json")
	base_url = compute_utils.GetBaseUrl()
	compute_url = compute_utils.GetComputeUrl()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	ariaclientId = financials_utils.GetAriaClientNo()
	ariaAuth = financials_utils.GetariaAuthKey()
	testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
	regionName = financials_utils.GetRegion1Name()
	fmt.Println("ariaAuth", ariaAuth)
	fmt.Println("ariaclientId", ariaclientId)

	By("Create cloud account...")
	// Generating token wtih payload for cloud account creation with enroll API
	userName = auth.Get_UserName("Intel")
	token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	fmt.Print("USERNAME..................", userName)
	token_response, _ := auth.Get_Azure_Bearer_Token(userName)
	userToken = "Bearer " + token_response
	cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, false)
	fmt.Println("BODY...", cloudaccount_creation_body)
	Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
	cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
	cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
	Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_INTEL"`)).To(BeTrue(), "Validation failed on Enrollment of Intel user.")
	place_holder_map["cloud_account_id"] = cloud_account_created
	place_holder_map["cloud_account_type"] = cloudaccount_type
	fmt.Println("cloudAccount_id", cloud_account_created)
})

func init() {

	logger.InitializeZapCustomLogger()
}

func TestEUMultiRegion2(t *testing.T) {
	if os.Getenv("DEV") != "" {
		t.Skip("Skipping, multi region is not available in current dev environment.")
	}
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "EUMultiRegion Suite")
}
