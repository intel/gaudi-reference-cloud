//go:build instance
// +build instance

package bmaas

import (
	"fmt"
	"log"
	"os"

	"strings"
	"time"

	"goFramework/framework/authentication"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("BMVS Instance Flow", Ordered, Label("large"), func() {

	var base_url string
	var token string
	var idc_global_url string

	var cloud_account_created string
	var vnet_created string
	var ssh_publickey_created string
	var instance_id_created string
	var instance_name_created string

	var create_response_status int
	var create_response_body string

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../../data", "bmaas_input.json")
		var token_url = os.Getenv("TOKEN_URL_PREFIX")
		idc_global_url = os.Getenv("IDC_GLOBAL_URL_PREFIX")
		base_url = os.Getenv("IDC_REGIONAL_URL_PREFIX")

		Expect(token_url).NotTo(BeZero())
		Expect(idc_global_url).NotTo(BeZero())
		Expect(base_url).NotTo(BeZero())

		token = authentication.GetBearerTokenViaFrisby(token_url)
	})

	It("Create cloud account", func() {
		username := utils.GenerateString(10) + "@intel.com"
		tid := utils.GenerateString(12)
		oid := utils.GenerateString(12)
		cloudaccount_url := idc_global_url + "/v1/cloudaccounts"
		cloudaccount_payload := fmt.Sprintf(`{"name":"%s","owner":"%s","tid":"%s","oid":"%s","type":"ACCOUNT_TYPE_INTEL"}`, "test-cloudaccount-bm"+utils.GenerateString(10), username, tid, oid)
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_url, token, cloudaccount_payload)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "id").String()
		log.Printf("cloudAccount_id: %s", cloud_account_created)
	})

	It("Create vnet", func() {
		log.Printf("Starting the VNet Creation via API...")
		// form the endpoint and payload
		vnet_name := "test-vnet-bm"
		payload_expected := fmt.Sprintf(`"name":"%s"`, vnet_name)
		vnet_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
		vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
		log.Printf("vnet_payload: %s", vnet_payload)
		// hit the api
		vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, token, vnet_payload)
		Expect(vnet_creation_status).To(Equal(200), "Failed to create VNet")
		Expect(strings.Contains(vnet_creation_body, payload_expected)).To(BeTrue(), "Failed to create VNet, response validation failed")
		vnet_created = gjson.Get(vnet_creation_body, "metadata.name").String()
	})

	It("Create ssh public key", func() {
		log.Printf("Starting the SSH-Public-Key Creation via API...")
		// form the endpoint and payload
		ssh_key_name := "testtc-bm"
		payload_expected := fmt.Sprintf(`"name":"%s"`, ssh_key_name)
		ssh_publickey_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/sshpublickeys"
		ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), ssh_key_name, sshPublicKey)
		log.Printf("ssh_publickey_payload: %s", ssh_publickey_payload)
		// hit the api
		sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, token, ssh_publickey_payload)
		Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key")
		Expect(strings.Contains(sshkey_creation_body, payload_expected)).To(BeTrue(), "Failed to create SSH Public key, response validation failed")
		ssh_publickey_created = gjson.Get(sshkey_creation_body, "metadata.name").String()
	})

	It("Create BM instance", func() {
		log.Printf("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_name := "test-bm"
		payload_expected := fmt.Sprintf(`"name":"%s"`, instance_name)
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), instance_name, instanceType, "ubuntu-22.04-server-cloudimg-amd64-latest", ssh_publickey_created, vnet_created)
		log.Printf("instance_payload: %s", instance_payload)

		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, token, instance_payload)
		log.Printf("Create instance status %d response %s", create_response_status, create_response_body)
		Expect(create_response_status).To(Equal(200), "Failed to create BM instance")
		Expect(strings.Contains(create_response_body, payload_expected)).To(BeTrue(), "Failed to create BM instance, response validation failed")
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		instance_name_created = gjson.Get(create_response_body, "metadata.name").String()
	})

	It("Get the created BM instance and validate it", func() {
		log.Printf("Starting the Instance Creation Validation via API...")
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		// Adding a sleep because it seems to take some time to reflect the creation status
		done := make(chan struct{})
		instancePhaseValidation := compute_utils.CheckInstanceState(instance_endpoint, token, instance_id_created, "Ready", done)
		Eventually(instancePhaseValidation, 10*time.Minute, 10*time.Second).Should(BeTrue())
	})

	It("Power OFF the instance", func() {
		log.Printf("Powering off the instance...")
		// update the instance's run strategy to halted state
		updated_halt_payload := `{
			"spec": {
				"runStrategy": "Halted",
				"sshPublicKeyNames": [
				  "<<ssh-publickey-created>>"
				]
			}
		}`
		updated_halt_payload = strings.ReplaceAll(updated_halt_payload, "<<ssh-publickey-created>>", ssh_publickey_created)
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		instance_halt_status, _ := frisby.PutInstanceByName(instance_endpoint, token, instance_name_created, updated_halt_payload)
		Expect(instance_halt_status).To(Equal(200), "Failed to update BM instance")

		// validate whether the instance phase is moved to stopped state
		done := make(chan struct{})
		instancePhaseValidation := compute_utils.CheckInstanceState(instance_endpoint, token, instance_id_created, "Stopped", done)
		Eventually(instancePhaseValidation, 5*time.Minute, 10*time.Second).Should(BeTrue())
	})

	It("Power ON the instance", func() {
		log.Printf("Powering on the instance...")
		// update the instance's run strategy to Always state
		updated_start_payload := `{
			"spec": {
				"runStrategy": "Always",
				"sshPublicKeyNames": [
				  "<<ssh-publickey-created>>"
				]
			}
		}`
		updated_start_payload = strings.ReplaceAll(updated_start_payload, "<<ssh-publickey-created>>", ssh_publickey_created)
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		instance_start_status, _ := frisby.PutInstanceByName(instance_endpoint, token, instance_name_created, updated_start_payload)
		Expect(instance_start_status).To(Equal(200), "Failed to update BM instance")

		// validate whether the instance phase is moved to Ready state
		done := make(chan struct{})
		instancePhaseValidation := compute_utils.CheckInstanceState(instance_endpoint, token, instance_id_created, "Ready", done)
		Eventually(instancePhaseValidation, 5*time.Minute, 10*time.Second).Should(BeTrue())
	})

	It("Delete the instance created", func() {
		log.Printf("Starting Instance Deletion via API...")
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"

		// delete the instance created
		delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, token, instance_id_created)
		Expect(delete_response_status).To(Equal(200), "Failed to delete BM instance")
	})

	It("Validate instance deletion", func() {
		log.Printf("Starting Instance Deletion Validation via API...")
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		// Adding a sleep because it seems to take some time to reflect the deletion
		time.Sleep(15 * time.Second)
		get_response_status, _ := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
		Expect(get_response_status).To(Equal(404), "Instance shouldn't be found")
	})

})
