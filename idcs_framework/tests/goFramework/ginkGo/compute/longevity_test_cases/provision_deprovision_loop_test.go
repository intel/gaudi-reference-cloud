//go:build provision_deprovision_loop || All || BM
// +build provision_deprovision_loop All BM

package longevity_test_cases_test

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

var _ = Describe("Provision and deprovision loop", Ordered, Label("large"), func() {

	var base_url string
	var token string

	var reservation_data = make(map[string]string)
	reservation_data["user_name"] = "longevity-user"

	var token_url = "http://dev.oidc.cloud.intel.com.kind.local"
	var idc_global_url = "https://dev.api.cloud.intel.com.kind.local"
	var idc_regional_url = "https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local"

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../data", "bmaas_input.json")
		base_url = idc_regional_url
		token = authentication.GetBearerTokenViaFrisby(token_url)
	})

	Describe("Create user data", func() {
		It("Create cloud account, vnet and sshkey", func() {
			By("Create cloud account")
			log.Printf("Starting the Cloud Account Creation via API...")
			username := utils.GenerateString(10) + "@intel.com"
			tid := utils.GenerateString(12)
			oid := utils.GenerateString(12)
			cloudaccount_url := idc_global_url + "/v1/cloudaccounts"
			cloudaccount_payload := fmt.Sprintf(`{"name":"%s","owner":"%s","tid":"%s","oid":"%s","type":"ACCOUNT_TYPE_INTEL"}`, reservation_data["user_name"] + "-cloudaccount-bm", username, tid, oid)
			cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_url, token, cloudaccount_payload)
			Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account")
			reservation_data["cloud_account_id"] = gjson.Get(cloudaccount_creation_body, "id").String()

			By("Create vnet")
			log.Printf("Starting the VNet Creation via API...")
			vnet_name := reservation_data["user_name"] + "-vnet-bm"
			log.Printf(vnet_name)
			log.Printf(token)
			payload_expected := fmt.Sprintf(`"name":"%s"`, vnet_name)
			vnet_endpoint := base_url + "/v1/cloudaccounts/" + reservation_data["cloud_account_id"] + "/vnets"
			vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
			log.Printf("vnet_payload: %s", vnet_payload)
			vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, token, vnet_payload)
			Expect(vnet_creation_status).To(Equal(200), "Failed to create VNet")
			Expect(strings.Contains(vnet_creation_body, payload_expected)).To(BeTrue(), "Failed to create VNet, response validation failed")
			reservation_data["vnet_name"] = gjson.Get(vnet_creation_body, "metadata.name").String()

			By("Create ssh public key")
			log.Printf("Starting the SSH-Public-Key Creation via API...")
			ssh_key_name := reservation_data["user_name"] + "-sshkey-bm"
			payload_expected = fmt.Sprintf(`"name":"%s"`, ssh_key_name)
			ssh_publickey_endpoint := base_url + "/v1/cloudaccounts/" + reservation_data["cloud_account_id"] + "/sshpublickeys"
			ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), ssh_key_name, sshPublicKey)
			log.Printf("ssh_publickey_payload: %s", ssh_publickey_payload)
			sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, token, ssh_publickey_payload)
			Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key")
			Expect(strings.Contains(sshkey_creation_body, payload_expected)).To(BeTrue(), "Failed to create SSH Public key, response validation failed")
			reservation_data["sshkey_name"] = gjson.Get(sshkey_creation_body, "metadata.name").String()
		})
	})

	Describe("Provision and deprovision loop", func() {
		for i := 1; i <= 48; i++ {
			It("Provision instance", func() {
				By("Create BM instance")
				log.Printf("Instance #%d", i)
				log.Printf("Starting the Instance Creation via API...")
				instance_name := reservation_data["user_name"] + "-instance-bm"
				payload_expected := fmt.Sprintf(`"name":"%s"`, instance_name)
				instance_endpoint := base_url + "/v1/cloudaccounts/" + reservation_data["cloud_account_id"] + "/instances"
				instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), instance_name, instanceType, "ubuntu-22.04-server-cloudimg-amd64-latest", reservation_data["sshkey_name"], reservation_data["vnet_name"])
				create_response_status, create_response_body := frisby.CreateInstance(instance_endpoint, token, instance_payload)
				log.Printf("Create instance status %d response %s", create_response_status, create_response_body)
				Expect(create_response_status).To(Equal(200), "Failed to create BM instance")
				Expect(strings.Contains(create_response_body, payload_expected)).To(BeTrue(), "Failed to create BM instance, response validation failed")
				reservation_data["resource_id"] = gjson.Get(create_response_body, "metadata.resourceId").String()

				By("Get BMH device by resource ID")
				log.Printf("Starting the BMH Device Retrieval via Kube...")
				time.Sleep(5 * time.Second)
				response_bmh, err := kube.GetBmhByConsumer(reservation_data["resource_id"])
				Expect(err).Error().ShouldNot(HaveOccurred())
				Expect(response_bmh.Spec.ConsumerRef.Name).To(Equal(reservation_data["resource_id"]))
				reservation_data["device_name"] = response_bmh.ObjectMeta.Name

				By("Validate BMH device is provisoned")
				log.Printf("Starting the BMH Device Validation via Kube...")
				succeded, err := kube.CheckBMHState(reservation_data["device_name"], "provisioned", 900)
				Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
					"unable to reach state within expected time")
				Expect(succeded).To(Equal(true))

				By("Validate instance is ready and get its details")
				log.Printf("Starting the Instance Validation via API...")
				instance_endpoint = base_url + "/v1/cloudaccounts/" + reservation_data["cloud_account_id"] + "/instances"

				var response_status int
				var response_body string
				start_time := time.Now()

				for is_instance_ready := false; !is_instance_ready; { // Until instance is ready
					time.Sleep(time.Minute) // Check every minute

					response_status, response_body = frisby.GetInstanceById(instance_endpoint, token, reservation_data["resource_id"])
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

				reservation_data["machine_ip"] = gjson.Get(response_body, "status.interfaces.0.addresses.0").String()
				reservation_data["proxy_ip"] = gjson.Get(response_body, "status.sshProxy.proxyAddress").String()
				reservation_data["proxy_user"] = gjson.Get(response_body, "status.sshProxy.proxyUser").String()
				log.Printf("Instance IP Address is: " + reservation_data["machine_ip"])
			})

			It("SSH into the BM instance", func() {
				log.Printf("Starting the Instance SSH connection via Ansible...")
				inventory_raw_data, _ := compute_utils.ConvertFileToString("../ansible-files", "inventory_raw.ini")
				log.Printf("Inventory raw data is: " + inventory_raw_data)
				inventory_generated := compute_utils.EnrichInventoryData(inventory_raw_data, reservation_data["proxy_ip"], reservation_data["proxy_user"], reservation_data["machine_ip"], "~/.ssh/id_rsa")
				log.Printf("Inventory generated is: " + inventory_generated)
				compute_utils.WriteStringToFile("../ansible-files", "inventory.ini", inventory_generated)
	
				// Get the pod details after restart
				var output bytes.Buffer
				get_pod_cmd := exec.Command("ansible-playbook", "-i", "../ansible-files/inventory.ini", "../ansible-files/ssh-and-apt-get-on-bm.yml")
				get_pod_cmd.Stdout = &output
				error := get_pod_cmd.Run()
				if error != nil {
					log.Printf("Execution of ansible playbook is not successful: %s", error)
				}
	
				// Log the ansible output
				ansible_output := strings.Split(output.String(), "\n")
				log.Print(ansible_output)
			})

			It("Deprovision instance", func() {
				By("Delete the instance created and validate deletion")
				log.Printf("Starting the Instance Deletion via API...")
				instance_endpoint := base_url + "/v1/cloudaccounts/" + reservation_data["cloud_account_id"] + "/instances"
				delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, token, reservation_data["resource_id"])
				Expect(delete_response_status).To(Equal(200), "Failed to delete BM instance")
				time.Sleep(5 * time.Second)
	
				get_response_status, _ := frisby.GetInstanceById(instance_endpoint, token, reservation_data["resource_id"])
				Expect(get_response_status).To(Equal(404), "Instance shouldn't be found")
	
				By("Validate instance deprovision")
				log.Printf("Starting the Instance Deprovision Validation via Kube...")
				succeded, err := kube.CheckBMHState(reservation_data["device_name"], "available", 900)
				Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
					"unable to reach state within expected time")
				Expect(succeded).To(Equal(true))
			})
		}
	})

	Describe("Delete all user data", func() {
		It("Delete cloud account, vnet and sshkey", func() {
			By("Delete the SSH key created and validate deletion")
			log.Printf("Starting the SSH-Public-Key Deletion via API...")
			ssh_publickey_endpoint := base_url + "/v1/cloudaccounts/" + reservation_data["cloud_account_id"] + "/sshpublickeys"
			delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, reservation_data["sshkey_name"])
			Expect(delete_response_byname_status).To(Equal(200), "Failed to delete ssh key")
			time.Sleep(5 * time.Second)

			get_response_status, _ := frisby.GetSSHKeyByName(ssh_publickey_endpoint, token, reservation_data["sshkey_name"])
			Expect(get_response_status).To(Equal(404), "SSH Key shouldn't be found")

			By("Delete the vnet created and validate deletion")
			log.Printf("Starting the Vnet Deletion via API...")
			vnet_endpoint := base_url + "/v1/cloudaccounts/" + reservation_data["cloud_account_id"] + "/vnets"
			delete_response_byname_status, _ = frisby.DeleteVnetByName(vnet_endpoint, token, reservation_data["vnet_name"])
			Expect(delete_response_byname_status).To(Equal(200), "Failed to delete vnet")
			time.Sleep(5 * time.Second)

			get_response_status, _ = frisby.GetVnetByName(vnet_endpoint, token, reservation_data["vnet_name"])
			Expect(get_response_status).To(Equal(404), "Vnet shouldn't be found")

			By("Delete the cloud account created and validate deletion")
			log.Printf("Starting the Cloud Account Deletion via API...")
			url := idc_global_url + "/v1/cloudaccounts"
			delete_Cacc, _ := financials.DeleteCloudAccountById(url, token, reservation_data["cloud_account_id"])
			Expect(delete_Cacc).To(Equal(200), "Failed to delete Cloud Account")

			get_response_status, _ = financials.GetCloudAccountById(url, token, reservation_data["cloud_account_id"])
			Expect(get_response_status).To(Equal(404), "Cloud account shouldn't be found")
		})
	})
})
