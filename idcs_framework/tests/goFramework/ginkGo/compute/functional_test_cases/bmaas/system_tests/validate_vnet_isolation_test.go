//go:build validate_vnet_isolation
// +build validate_vnet_isolation

package system_tests_test

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"goFramework/framework/authentication"
	"goFramework/framework/library/bmaas/kube"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Validate vnet isolation", Ordered, Label("large"), func() {

	var base_url string
	var token string

	var user1 = make(map[string]string)
	var user2 = make(map[string]string)

	var token_url = "http://dev.oidc.cloud.intel.com.kind.local"
	var idc_global_url = "https://dev.api.cloud.intel.com.kind.local"
	var idc_regional_url = "https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local"

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../../data", "bmaas_input.json")
		base_url = idc_regional_url
		token = authentication.GetBearerTokenViaFrisby(token_url)

		user1["user_name"] = "user1"
		user2["user_name"] = "user2"
	})

	DescribeTable("Create user and reserve a system",
		func(user map[string]string) {
			By("Create cloud account")
			log.Printf("Starting the Cloud Account Creation via API...")
			username := utils.GenerateString(10) + "@intel.com"
			tid := utils.GenerateString(12)
			oid := utils.GenerateString(12)
			cloudaccount_url := idc_global_url + "/v1/cloudaccounts"
			cloudaccount_payload := fmt.Sprintf(`{"name":"%s","owner":"%s","tid":"%s","oid":"%s","type":"ACCOUNT_TYPE_INTEL"}`, user["user_name"]+"-cloudaccount-bm", username, tid, oid)
			cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_url, token, cloudaccount_payload)
			Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account")
			user["cloud_account"] = gjson.Get(cloudaccount_creation_body, "id").String()

			By("Create vnet")
			log.Printf("Starting the VNet Creation via API...")
			vnet_name := user["user_name"] + "-vnet-bm"
			payload_expected := fmt.Sprintf(`"name":"%s"`, vnet_name)
			vnet_endpoint := base_url + "/v1/cloudaccounts/" + user["cloud_account"] + "/vnets"
			vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
			log.Printf("vnet_payload: %s", vnet_payload)
			vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, token, vnet_payload)
			Expect(vnet_creation_status).To(Equal(200), "Failed to create VNet")
			Expect(strings.Contains(vnet_creation_body, payload_expected)).To(BeTrue(), "Failed to create VNet, response validation failed")
			user["vnet"] = gjson.Get(vnet_creation_body, "metadata.name").String()

			By("Create ssh public key")
			log.Printf("Starting the SSH-Public-Key Creation via API...")
			ssh_key_name := user["user_name"] + "-sshkey-bm"
			payload_expected = fmt.Sprintf(`"name":"%s"`, ssh_key_name)
			ssh_publickey_endpoint := base_url + "/v1/cloudaccounts/" + user["cloud_account"] + "/sshpublickeys"
			ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), ssh_key_name, sshPublicKey)
			log.Printf("ssh_publickey_payload: %s", ssh_publickey_payload)
			sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, token, ssh_publickey_payload)
			Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key")
			Expect(strings.Contains(sshkey_creation_body, payload_expected)).To(BeTrue(), "Failed to create SSH Public key, response validation failed")
			user["sshkey"] = gjson.Get(sshkey_creation_body, "metadata.name").String()

			By("Create BM instance")
			log.Printf("Starting the Instance Creation via API...")
			instance_name := user["user_name"] + "-instance-bm"
			payload_expected = fmt.Sprintf(`"name":"%s"`, instance_name)
			instance_endpoint := base_url + "/v1/cloudaccounts/" + user["cloud_account"] + "/instances"
			instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), instance_name, instanceType, "ubuntu-22.04-server-cloudimg-amd64-latest", user["sshkey"], user["vnet"])
			create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, token, instance_payload)
			log.Printf("Create instance status %d response %s", create_response_status, create_response_body)
			Expect(create_response_status).To(Equal(200), "Failed to create BM instance")
			Expect(strings.Contains(create_response_body, payload_expected)).To(BeTrue(), "Failed to create BM instance, response validation failed")
			user["resource_id"] = gjson.Get(create_response_body, "metadata.resourceId").String()

			By("Get BMH device by resource ID")
			log.Printf("Starting the BMH Device Retrieval via Kube...")
			time.Sleep(5 * time.Second)
			response_bmh, err := kube.GetBmhByConsumer(user["resource_id"])
			Expect(err).Error().ShouldNot(HaveOccurred())
			Expect(response_bmh.Spec.ConsumerRef.Name).To(Equal(user["resource_id"]))
			user["device_name"] = response_bmh.ObjectMeta.Name

			By("Validate BMH device is provisoned")
			log.Printf("Starting the BMH Device Validation via Kube...")
			succeded, err := kube.CheckBMHState(user["device_name"], "provisioned", 900)
			Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
				"unable to reach state within expected time")
			Expect(succeded).To(Equal(true))

			By("Validate instance is ready and get its details")
			log.Printf("Starting the Instance Validation via API...")
			instance_endpoint = base_url + "/v1/cloudaccounts/" + user["cloud_account"] + "/instances"

			var response_status int
			var response_body string
			start_time := time.Now()

			for is_instance_ready := false; !is_instance_ready; { // Until instance is ready
				time.Sleep(time.Minute) // Check every minute

				response_status, response_body = frisby.GetInstanceById(instance_endpoint, token, user["resource_id"])
				is_instance_ready = strings.Contains(response_body, `"phase":"Ready"`)
				wait_time := time.Since(start_time)

				log.Printf("Waiting for instance to be ready (from %f min ago)", wait_time.Minutes())

				// Timeout of 15min
				if wait_time > (time.Minute * 15) {
					break
				}
			}

			Expect(response_status).To(Equal(200), "Failed to retrieve BM instance")
			Expect(strings.Contains(response_body, `"phase":"Ready"`)).To(BeTrue(), "Validation failed on instance retrieval")

			user["machine_ip"] = gjson.Get(response_body, "status.interfaces.0.addresses.0").String()
			user["proxy_ip"] = gjson.Get(response_body, "status.sshProxy.proxyAddress").String()
			user["proxy_user"] = gjson.Get(response_body, "status.sshProxy.proxyUser").String()
			log.Printf("Instance IP Address is: " + user["machine_ip"])
		},
		Entry("for "+user1["user_name"], user1),
		Entry("for "+user2["user_name"], user2),
	)

	Describe("Validate system vnet isolation", func() {
		It("SSH into the BM instance", func() {
			log.Printf("Starting the Instance SSH connection via Ansible...")
			inventory_raw_data, _ := compute_utils.ConvertFileToString("../../../ansible-files", "inventory_raw.ini")
			log.Printf("Inventory raw data is: " + inventory_raw_data)
			inventory_generated := compute_utils.EnrichInventoryData(inventory_raw_data, user1["proxy_ip"], user1["proxy_user"], user1["machine_ip"], "~/.ssh/id_rsa")
			log.Printf("Inventory generated is: " + inventory_generated)
			compute_utils.WriteStringToFile("../../../ansible-files", "inventory.ini", inventory_generated)

			// Get the pod details after restart
			var output bytes.Buffer
			get_pod_cmd := exec.Command("ansible-playbook", "-i", "../../../ansible-files/inventory.ini", "--extra-vars", "another_host="+user2["machine_ip"], "../../../ansible-files/ssh-on-bm-to-ping-another-bm.yml")
			get_pod_cmd.Stdout = &output
			error := get_pod_cmd.Run()
			if error != nil {
				log.Printf("Execution of ansible playbook is not successful: %s", error)
			}

			// Log the ansible output
			ansible_output := strings.Split(output.String(), "\n")
			log.Print(ansible_output)
			Expect(strings.Contains(output.String(), "100% packet loss")).To(BeTrue(), "Failed to validate vnet isolation")
		})
	})

	DescribeTable("Delete reservation and all system data",
		func(user map[string]string) {
			By("Delete the instance created and validate deletion")
			log.Printf("Starting the Instance Deletion via API...")
			instance_endpoint := base_url + "/v1/cloudaccounts/" + user["cloud_account"] + "/instances"
			delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, token, user["resource_id"])
			Expect(delete_response_status).To(Equal(200), "Failed to delete BM instance")
			time.Sleep(5 * time.Second)

			get_response_status, _ := frisby.GetInstanceById(instance_endpoint, token, user["resource_id"])
			Expect(get_response_status).To(Equal(404), "Instance shouldn't be found")

			By("Validate instance deprovision")
			log.Printf("Starting the Instance Deprovision Validation via Kube...")
			succeded, err := kube.CheckBMHState(user["device_name"], "available", 900)
			Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
				"unable to reach state within expected time")
			Expect(succeded).To(Equal(true))

			By("Delete the SSH key created and validate deletion")
			log.Printf("Starting the SSH-Public-Key Deletion via API...")
			ssh_publickey_endpoint := base_url + "/v1/cloudaccounts/" + user["cloud_account"] + "/sshpublickeys"
			delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, user["sshkey"])
			Expect(delete_response_byname_status).To(Equal(200), "Failed to delete ssh key")
			time.Sleep(5 * time.Second)

			get_response_status, _ = frisby.GetSSHKeyByName(ssh_publickey_endpoint, token, user["sshkey"])
			Expect(get_response_status).To(Equal(404), "SSH Key shouldn't be found")

			By("Delete the vnet created and validate deletion")
			log.Printf("Starting the Vnet Deletion via API...")
			vnet_endpoint := base_url + "/v1/cloudaccounts/" + user["cloud_account"] + "/vnets"
			delete_response_byname_status, _ = frisby.DeleteVnetByName(vnet_endpoint, token, user["vnet"])
			Expect(delete_response_byname_status).To(Equal(200), "Failed to delete vnet")
			time.Sleep(5 * time.Second)

			get_response_status, _ = frisby.GetVnetByName(vnet_endpoint, token, user["vnet"])
			Expect(get_response_status).To(Equal(404), "Vnet shouldn't be found")

			By("Delete the cloud account created and validate deletion")
			log.Printf("Starting the Cloud Account Deletion via API...")
			url := idc_global_url + "/v1/cloudaccounts"
			delete_Cacc, _ := financials.DeleteCloudAccountById(url, token, user["cloud_account"])
			Expect(delete_Cacc).To(Equal(200), "Failed to delete Cloud Account")

			get_response_status, _ = financials.GetCloudAccountById(url, token, user["cloud_account"])
			Expect(get_response_status).To(Equal(404), "Cloud account shouldn't be found")
		},
		Entry("for "+user1["user_name"], user1),
		Entry("for "+user2["user_name"], user2),
	)
})
