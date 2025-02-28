//go:build bm_dev3 || All || BM
// +build bm_dev3 All BM

package bmaas

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"goFramework/framework/library/auth"
	"goFramework/framework/library/bmaas/kube"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/ginkGo/compute/compute_utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("BMaaS E2E flow", Ordered, Label("large"), func() {

	var base_url string
	var token string

	var cloud_account_created string
	var vnet_created string
	var ssh_publickey_created string
	var instance_id_created string

	var create_response_status int
	var create_response_body string
	var place_holder_map = make(map[string]string)

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../data", "bmaas_input_dev3.json")
		base_url = compute_utils.GetBaseUrl()

		auth.Get_config_file_data("../../data/auth_data.json")
		token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
	})

	It("Create cloud account", func() {
		log.Printf("Starting the Cloud Account Creation via API...")
		cloud_account_created = "247959607216"
	})

	It("Create vnet", func() {
		log.Printf("Starting the VNet Creation via API...")
		vnet_created = "us-dev3-1a-default"
	})

	It("Create ssh public key", func() {
		log.Printf("Starting the SSH-Public-Key Creation via API...")
		// form the endpoint and payload
		ssh_key_name := "dev3-bm-test-sshkey"
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
		instance_name := "dev3-bm-test-instance"
		payload_expected := fmt.Sprintf(`"name":"%s"`, instance_name)
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), instance_name, instanceType, "ubuntu-22.04-server-cloudimg-amd64-latest", ssh_publickey_created, vnet_created)
		log.Printf("instance_payload: %s", instance_payload)
		place_holder_map["instance_type"] = instanceType
		log.Printf("instance_type: %s", instanceType)

		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, token, instance_payload)
		log.Printf("Create instance status %d response %s", create_response_status, create_response_body)
		Expect(create_response_status).To(Equal(200), "Failed to create BM instance")
		Expect(strings.Contains(create_response_body, payload_expected)).To(BeTrue(), "Failed to create BM instance, response validation failed")
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		place_holder_map["resource_id"] = instance_id_created
		place_holder_map["instance_creation_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	It("Get BMH device by resource ID", func() {
		log.Printf("Starting the BMH Device Retrieval via Kube...")
		time.Sleep(5 * time.Second)
		response_bmh, err := kube.GetBmhByConsumer(instance_id_created)
		Expect(err).Error().ShouldNot(HaveOccurred())
		Expect(response_bmh.Spec.ConsumerRef.Name).To(Equal(instance_id_created))
		place_holder_map["device_name"] = response_bmh.ObjectMeta.Name
	})

	It("Validate BMH device is provisoned", func() {
		log.Printf("Starting the BMH Device Validation via Kube...")
		succeded, err := kube.CheckBMHState(place_holder_map["device_name"], "provisioned", 900)
		Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(true))
	})

	It("Validate instance is ready and get its details", func() {
		log.Printf("Starting the Instance Validation via API...")
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"

		var response_status int
		var response_body string
		start_time := time.Now()

		for is_instance_ready := false; !is_instance_ready; { // Until instance is ready
			time.Sleep(time.Minute) // Check every minute

			response_status, response_body = frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
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

		place_holder_map["machine_ip"] = gjson.Get(response_body, "status.interfaces.0.addresses.0").String()
		place_holder_map["proxy_ip"] = gjson.Get(response_body, "status.sshProxy.proxyAddress").String()
		place_holder_map["proxy_user"] = gjson.Get(response_body, "status.sshProxy.proxyUser").String()
		log.Printf("Instance IP Address is: " + place_holder_map["machine_ip"])
	})

	It("SSH into the BM instance", func() {
		log.Printf("Starting the Instance SSH connection via Ansible...")
		inventory_raw_data, _ := compute_utils.ConvertFileToString("../../ansible-files", "inventory_raw.ini")
		log.Printf("Inventory raw data is: " + inventory_raw_data)
		inventory_generated := compute_utils.EnrichInventoryData(inventory_raw_data, place_holder_map["proxy_ip"], place_holder_map["proxy_user"], place_holder_map["machine_ip"], "~/.ssh/id_rsa")
		log.Printf("Inventory generated is: " + inventory_generated)
		compute_utils.WriteStringToFile("../../ansible-files", "inventory.ini", inventory_generated)

		// Get the pod details after restart
		var output bytes.Buffer
		get_pod_cmd := exec.Command("ansible-playbook", "-i", "../../ansible-files/inventory.ini", "../../ansible-files/ssh-and-apt-get-on-bm.yml")
		get_pod_cmd.Stdout = &output
		error := get_pod_cmd.Run()
		if error != nil {
			log.Printf("Execution of ansible playbook is not successful: %s", error)
		}

		// Log the ansible output
		ansible_output := strings.Split(output.String(), "\n")
		log.Print(ansible_output)
	})

	It("Delete the instance created and validate deletion", func() {
		log.Printf("Starting the Instance Deletion via API...")
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"

		// delete the instance created
		delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, token, instance_id_created)
		Expect(delete_response_status).To(Equal(200), "Failed to delete BM instance")

		var get_response_status int
		start_time := time.Now()

		for is_instance_deleted := false; !is_instance_deleted; { // Until instance is deleted
			time.Sleep(5 * time.Second) // Check every 5 seconds

			get_response_status, _ = frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
			is_instance_deleted = (get_response_status == 404)

			wait_time := time.Since(start_time)

			log.Printf("Waiting for instance to be deleted (from %f s ago)", wait_time.Seconds())

			// Timeout of 30s
			if wait_time > (time.Second * 30) {
				break
			}
		}

		// validate the deletion
		Expect(get_response_status).To(Equal(404), "Instance shouldn't be found")
	})

	It("Validate instance deprovision", func() {
		log.Printf("Starting the Instance Deprovision Validation via Kube...")
		succeded, err := kube.CheckBMHState(place_holder_map["device_name"], "available", 900)
		Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(true))
	})

	It("Delete the SSH key created and validate deletion", func() {
		log.Printf("Starting the SSH-Public-Key Deletion via API...")
		ssh_publickey_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/sshpublickeys"

		// delete the ssh key created
		delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_created)
		Expect(delete_response_byname_status).To(Equal(200), "Failed to delete ssh key")
		time.Sleep(5 * time.Second)

		// validate the deletion
		get_response_status, _ := frisby.GetSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_created)
		Expect(get_response_status).To(Equal(404), "SSH Key shouldn't be found")
	})

})
