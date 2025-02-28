//go:build su_vm || All || pu_su_iu
// +build su_vm All pu_su_iu

package vmaas

import (
	"fmt"
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

var _ = Describe("VM Instance", Ordered, Label("Compute-VMaaS-E2E"), func() {
	var base_url string
	var compute_url string
	var userName string
	var token string
	var userToken string
	var cloud_account_created string
	var vnet_created string
	var ssh_publickey_name_created string
	var ariaAuth string

	var create_response_status int
	var create_response_body string
	var place_holder_map = make(map[string]string)
	var ariaclientId string

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../../financials/data", "vmaas_input.json")
		financials_utils.LoadE2EConfig("../../../financials/data", "billing.json")
		auth.Get_config_file_data("../../../financials/data/config.json")
		userName := auth.Get_UserName("Standard")
		base_url = compute_utils.GetBaseUrl()
		compute_url = compute_utils.GetComputeUrl()
		token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
		userToken, _ = auth.Get_Azure_Bearer_Token(userName)
		userToken = "Bearer " + userToken
		ariaclientId = financials_utils.GetAriaClientNo()
		ariaAuth = financials_utils.GetariaAuthKey()
		testsetup.ProductUsageData = make(map[string]testsetup.UsageData)
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("ariaclientId", ariaclientId)
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
		cloudaccount_enroll_url := base_url + "/v1/cloudaccounts/enroll"
		// Generating token wtih payload for cloud account creation with enroll API
		token_response, _ := auth.Get_Azure_Bearer_Token(userName)
		enroll_payload := `{"premium":false}`
		// cloud account creation
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_enroll_url, "Bearer "+token_response, enroll_payload)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_STANDARD"`)).To(BeTrue(), "Validation failed on Enrollment of Standard user")
		place_holder_map["cloud_account_id"] = cloud_account_created
		place_holder_map["cloud_account_type"] = cloudaccount_type
		fmt.Println("cloudAccount_id", cloud_account_created)
	})

	It("Create vnet with name", func() {
		// fmt.Println("Starting the VNet Creation via API...")
		// // form the endpoint and payload
		vnet_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
		vnet_name := compute_utils.GetVnetName()
		vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
		// hit the api
		fmt.Println("Vnet end point ", vnet_endpoint)
		vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
		vnet_created = gjson.Get(vnet_creation_body, "metadata.name").String()
		Expect(vnet_creation_status).To(Equal(200), "Failed to create VNet")
		Expect(vnet_name).To(Equal(vnet_created), "Failed to create Vnet, response validation failed")
	})

	It("Create ssh public key with name", func() {
		fmt.Println("Starting the SSH-Public-Key Creation via API...")
		// form the endpoint and payload
		ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
		fmt.Println("SSH key is" + sshPublicKey)
		sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
		fmt.Println("SSH  end point ", ssh_publickey_endpoint)
		ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
		// hit the api
		sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, token, ssh_publickey_payload)
		Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key")
		ssh_publickey_name_created = gjson.Get(sshkey_creation_body, "metadata.name").String()
		Expect(sshkey_name).To(Equal(ssh_publickey_name_created), "Failed to create SSH Public key, response validation failed")

	})

	It("Create paid vm instance when Standard user have zero cloud credits", func() {
		fmt.Println("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		vm_name_iu := "autovm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name_iu, "vm-spr-med", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = "vm-spr-med"
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
		Expect(create_response_status).To(Equal(403), "Expected response code on paid instance for Standard user should be 403 when cloud credits are zero")
		Expect(strings.Contains(create_response_body, `"message":"paid service not allowed"`)).To(BeTrue(), "Failed to validate ")

	})

	It("Create Cloud credits for Standard user by redeeming coupons", func() {
		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/billing/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayloadStandard(), 0.000001, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		// Redeem coupon
		redeem_coupon_endpoint := base_url + "/v1/billing/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), place_holder_map["cloud_account_id"])
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Failed to redeem coupon")
	})

	It("Create paid vm instance", func() {
		fmt.Println("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		vm_name_iu := "autovm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name_iu, "vm-spr-med", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = "vm-spr-med"
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
		Expect(create_response_status).To(Equal(403), "Failed to create VM instance")
	})

	It("Delete the SSH key created", func() {
		fmt.Println("Delete the SSH-Public-Key Created above...")
		ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
		delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_name_created)
		Expect(delete_response_byname_status).To(Equal(200), "assert ssh-public-key deletion response code")
	})

	It("Delete cloud account", func() {
		fmt.Println("Delete cloud account")
		url := base_url + "/v1/cloudaccounts"
		delete_Cacc, _ := financials.DeleteCloudAccountById(url, token, place_holder_map["cloud_account_id"])
		Expect(delete_Cacc).To(Equal(200), "Failed to delete Cloud Account")
	})

	It("Validating deletion of cloud account by retrieving cloud account details", func() {
		fmt.Println("Validating deletion of cloud account by retrieving cloud account details")
		url := base_url + "/v1/cloudaccounts"
		resp_body, _ := financials.GetCloudAccountById(url, token, place_holder_map["cloud_account_id"])
		fmt.Println("resp_body", resp_body)
	})

})
