package PU_GTS_CountryCode_token_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("BMaaS Instance", Ordered, Label("Compute-VMaaS-E2E"), func() {
	var base_url string
	var compute_url string
	var token string
	var userName string
	var userToken string
	var cloud_account_created string
	var vnet_created string

	var ariaAuth string

	var create_response_status int
	var create_response_body string
	var place_holder_map = make(map[string]string)

	var ariaclientId_su string
	instanceOS := "ubuntu-22.04-spr-metal-cloudimg-amd64-v20240115"

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../../data", "vmaas_input.json")
		auth.Get_config_file_data("../../../data/config.json")
		financials_utils.LoadE2EConfig("../../../data", "billing.json")
		userName = auth.Get_UserName("Premium")
		base_url = compute_utils.GetBaseUrl()
		compute_url = compute_utils.GetComputeUrl()
		token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
		ariaclientId_su = financials_utils.GetAriaClientNo()
		ariaAuth = financials_utils.GetariaAuthKey()
		testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("ariaclientId_su", ariaclientId_su)
	})

	It("Delete cloud account", func() {
		fmt.Println("Delete cloud account")
		url := base_url + "/v1/cloudaccounts"
		cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, token)
		if err == nil {
			financials.DeleteCloudAccountById(url, token, cloudAccId)
		}
	})

	It("Create cloud account", func() {
		// Generating token wtih payload for cloud account creation with enroll API
		// cloud account creation
		token_response, _ := auth.Get_Azure_Bearer_Token(userName)
		userToken = "Bearer " + token_response
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, true)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll, body"+cloudaccount_creation_body)
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of Premium user, Body: "+cloudaccount_creation_body)
		place_holder_map["cloud_account_id"] = cloud_account_created
		place_holder_map["cloud_account_type"] = cloudaccount_type
		fmt.Println("cloudAccount_id", cloud_account_created)
	})

	// ADD CREDITS TO ACCOUNT
	It("Create Cloud credits for premium user by redeeming coupons", func() {
		logger.Logf.Info("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/cloudcredits/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 100, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon, body"+coupon_creation_body)

		logger.Logf.Info("Redeem credits to current user...")
		redeem_coupon_endpoint := base_url + "/v1/cloudcredits/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), cloud_account_created)
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, userToken, redeem_payload)
		fmt.Println("Payload", redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Failed to redeem coupon")
	})

	// CREATE SSH KEY
	It("Create ssh public key with name", func() {
		logger.Logf.Info("Starting the SSH-Public-Key Creation via API...")
		// form the endpoint and payload
		ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
		sshPublicKey = utils.GetSSHKey()
		fmt.Println("SSH key is" + sshPublicKey)
		sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
		fmt.Println("SSH  end point ", ssh_publickey_endpoint)
		ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
		// hit the api
		sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)
		Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key, body:"+sshkey_creation_body)
		ssh_publickey_name_created = gjson.Get(sshkey_creation_body, "metadata.name").String()
		Expect(sshkey_name).To(Equal(ssh_publickey_name_created), "Failed to create SSH Public key, response validation failed")
	})

	// CREATE NEW INSTANCE
	It("Create BMaaS Instance", func() {
		fmt.Println("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		bm_name_iu := "autobm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, instanceType, instanceOS, ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = instanceType
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
		fmt.Println("Instance: ", create_response_body)
		Expect(create_response_status).NotTo(Equal(200), "Failed, BMaaS instance was created with wrong GTS, body: "+create_response_body)
		fmt.Println("PASS_GTS US")
	})

})
