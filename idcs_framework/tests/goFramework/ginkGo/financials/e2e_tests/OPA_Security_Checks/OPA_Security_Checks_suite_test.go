package OPA_Security_Checks_test

import (
	"flag"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
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

func init() {
	flag.StringVar(&sshPublicKey, "sshPublicKey", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCbnYTaUpeo7bsx6pgWhf1IDymryvU7UYVOqfxv5F0Lbpg4osD8bpuopOvVvrMGEdzhxv8Kzr5VaXi8689lPaAXf4Ltx6SFYAQwx+ZdzyqLA73vWftBoTuFvqVe+/IkZ/Y7Vyw+FFZqGOJSc40dLTQ7JfoKV896x5sllP5cO995T+a6R0krX/t00f+VjAypiiK5zWIVQwbGCq4x8upgB4RPeHyQUxMeRzLZAgbEqJBTr5pnZTZjLsPKrhvDp8FRdoUhGMwr6k+pfKL4ZV9T99c0pols5xZBMnreiugPDPt6/w2zXoE/Wa3vXawYZBnt1T0iW5SFJCab85bP/8PkLPRHWtGTttZat9zKWrztVcuG/AaonNi7xtHA6AyAnNs2FnpQOx5VTaMgP2f2l95b8Gg1RN2+5NnT2yxPIN2uCIiTHHjLHxLlpg6rQUs5wUzazpjsj0vfMYTb48d8lOimBJJaMb1iz6DhNtIC9nm8mYRrMgXytMHvUSg+/pxXROaS/zMYdE1dH/PUlnWSW5P3phZ++z1RPVZc7U8k7bNOdHLqXfrAqP+hs6o+9CLfRxOGGQP0jDiYe0S+K8TAt+iiZMxGw5xOa9yItYapZj+zzSaHfLudSvzFjlC+4PY9hIkHvqU08XFzgAEtZh4fkL9HN69Ubt2NrR2Xje8YFYvb0q0d1w== sdp@internal-placeholder.com", "")
	logger.InitializeZapCustomLogger()
}

var _ = BeforeSuite(func() {
	compute_utils.LoadE2EConfig("../../data", "vmaas_input.json")
	auth.Get_config_file_data("../../data/config.json")
	financials_utils.LoadE2EConfig("../../data", "billing.json")
	userName = auth.Get_UserName("Premium")
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
	Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
	cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
	cloudaccount_type = gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
	Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_STANDARD"`)).To(BeTrue(), "Validation failed on Enrollment of Standard user")
	place_holder_map_su["cloud_account_id"] = cloud_account_created
	place_holder_map_su["cloud_account_type"] = cloudaccount_type
	fmt.Println("cloudAccount_id for standard user", cloud_account_created)
})

func TestOPASecurityChecks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OPASecurityChecks Suite")
}

var _ = AfterSuite(func() {
	By("Admin Revokes invitation...")
	fmt.Println("base_url: ", base_url)
	code, body := financials.RemoveInvitation(base_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
	Expect(code).To(Or(Equal(200), Equal(404)), "Failed to reject invite."+body)
})
