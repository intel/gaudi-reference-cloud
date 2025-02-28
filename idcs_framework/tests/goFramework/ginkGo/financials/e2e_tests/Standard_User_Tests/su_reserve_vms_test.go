//go:build su_vm_load
// +build su_vm_load

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
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("VM Instance", Ordered, Label("Compute-VMaaS-E2E"), func() {
	var base_url string
	var compute_url string
	var token string
	var userName string
	var userToken string
	var cloud_account_created string
	var vnet_created string
	var ssh_publickey_name_created string

	var ariaAuth string

	var create_response_status int
	var create_response_body string
	var place_holder_map = make(map[string]string)

	var ariaclientId_su string

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

		cloudaccount_enroll_url := base_url + "/v1/cloudaccounts/enroll"
		// Generating token wtih payload for cloud account creation with enroll API
		token_response, _ := auth.Get_Azure_Bearer_Token(userName)
		enroll_payload := `{"premium":false}`
		// cloud account creation
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_enroll_url, "Bearer "+token_response, enroll_payload)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_STANDARD"`)).To(BeTrue(), "Validation failed on Enrollment of standard user")
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
		ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)

		// hit the api
		sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, token, ssh_publickey_payload)
		Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key")
		ssh_publickey_name_created = gjson.Get(sshkey_creation_body, "metadata.name").String()
		Expect(sshkey_name).To(Equal(ssh_publickey_name_created), "Failed to create SSH Public key, response validation failed")
		//ssh_publickey_name_created = gjson.Get(sshkey_creation_body, "metadata.name").String()
	})

	It("Create vm instance", func() {

		for i := 1; i <= 50; i++ {
			fmt.Println("Starting the Instance Creation via API...", i)
			instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
			vm_name := "autovm-" + utils.GenerateSSHKeyName(4)
			instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name, instanceType, "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
			fmt.Println("instance_payload", instance_payload)

			// hit the api
			create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, token, instance_payload)
			VMName_iu := gjson.Get(create_response_body, "metadata.name").String()
			Expect(create_response_status).To(Equal(200), "Failed to create VM instance")
			Expect(vm_name).To(Equal(VMName_iu), "Failed to create VM instance, resposne validation failed")
			time.Sleep(15 * time.Second)
		}
		// form the endpoint and payload

		//Expect(strings.Contains(create_response_body, `"name":"e2e-suvm1"`)).To(BeTrue(), "Failed to create VM instance, resposne validation failed")
		place_holder_map["instance_creation_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

})
