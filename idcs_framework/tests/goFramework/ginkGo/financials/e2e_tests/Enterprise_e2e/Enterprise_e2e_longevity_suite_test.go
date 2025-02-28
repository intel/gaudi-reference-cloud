package Enterprise_e2e_longevity_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/library/financials/cloudAccounts"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"strconv"
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

func init() {

	flag.StringVar(&sshPublicKey, "sshPublicKey", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCbnYTaUpeo7bsx6pgWhf1IDymryvU7UYVOqfxv5F0Lbpg4osD8bpuopOvVvrMGEdzhxv8Kzr5VaXi8689lPaAXf4Ltx6SFYAQwx+ZdzyqLA73vWftBoTuFvqVe+/IkZ/Y7Vyw+FFZqGOJSc40dLTQ7JfoKV896x5sllP5cO995T+a6R0krX/t00f+VjAypiiK5zWIVQwbGCq4x8upgB4RPeHyQUxMeRzLZAgbEqJBTr5pnZTZjLsPKrhvDp8FRdoUhGMwr6k+pfKL4ZV9T99c0pols5xZBMnreiugPDPt6/w2zXoE/Wa3vXawYZBnt1T0iW5SFJCab85bP/8PkLPRHWtGTttZat9zKWrztVcuG/AaonNi7xtHA6AyAnNs2FnpQOx5VTaMgP2f2l95b8Gg1RN2+5NnT2yxPIN2uCIiTHHjLHxLlpg6rQUs5wUzazpjsj0vfMYTb48d8lOimBJJaMb1iz6DhNtIC9nm8mYRrMgXytMHvUSg+/pxXROaS/zMYdE1dH/PUlnWSW5P3phZ++z1RPVZc7U8k7bNOdHLqXfrAqP+hs6o+9CLfRxOGGQP0jDiYe0S+K8TAt+iiZMxGw5xOa9yItYapZj+zzSaHfLudSvzFjlC+4PY9hIkHvqU08XFzgAEtZh4fkL9HN69Ubt2NrR2Xje8YFYvb0q0d1w== sdp@internal-placeholder.com", "")
	logger.InitializeZapCustomLogger()
}

var _ = BeforeSuite(func() {
	compute_utils.LoadE2EConfig("../../data", "product_catalog_e2e.json")
	auth.Get_config_file_data("../../data/config.json")
	financials_utils.LoadE2EConfig("../../data", "product_catalog_e2e.json")
	testsetup.Get_config_file_data("../../data/product_catalog_e2e.json")
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
	Expect(strings.Contains(string(out), `"type":"ACCOUNT_TYPE_ENTERPRISE"`)).To(BeTrue(), "Account type is not Enterprise, Validation Failed.")
	Expect(strings.Contains(string(out), `"paidServicesAllowed":true`)).To(BeTrue(), "Allowed paid services is not true, Validation Failed.")
	place_holder_map["cloud_account_id"] = cloud_account_created
	place_holder_map["cloud_account_type"] = cloudaccount_type
	success, message := billing.CreateBillingAccountWithSpecificCloudAccountIdOIDC(base_url, get_CAcc_id, 200, token)
	fmt.Println("success", success)
	fmt.Println("MESSAGE1", message)
})

func init() {
	logger.InitializeZapCustomLogger()
}

func TestEnterpriseE2eLongevity(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "EnterpriseE2eLongevity Suite")
}

var _ = AfterEach(func() {
	token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
})

var _ = AfterSuite(func() {
	//DELETE THE INSTANCE CREATED
	By("Deleted instance created...")
	instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
	time.Sleep(10 * time.Second)
	// delete the instance created
	delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, token, instance_id_created)
	Expect(delete_response_status).To(Equal(200), "Failed to delete VM instance")
	time.Sleep(5 * time.Second)
	// validate the deletion
	// Adding a sleep because it seems to take some time to reflect the deletion status
	time.Sleep(1 * time.Minute)
	get_response_status, _ := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
	Expect(get_response_status).To(Equal(404), "Resource shouldn't be found")
	place_holder_map["instance_deletion_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	// DELETE SSH KEYS
	By("Delete SSH keys...")
	logger.Logf.Info("Delete SSH keys...")
	fmt.Println("Delete the SSH-Public-Key Created above...")
	ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
	delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_name_created)
	Expect(delete_response_byname_status).To(Equal(200), "assert ssh-public-key deletion response code")
})
