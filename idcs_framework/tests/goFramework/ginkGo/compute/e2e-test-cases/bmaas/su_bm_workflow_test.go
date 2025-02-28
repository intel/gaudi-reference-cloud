//go:build su_bm || All || BM
// +build su_bm All BM

package bmaas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"goFramework/framework/authentication"
	"goFramework/framework/library/bmaas/kube"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("BMaaS E2E flow for Standard user", Ordered, Label("large"), func() {

	var base_url string
	var token string

	var cloud_account_created string
	var vnet_created string
	var ssh_publickey_created string
	var instance_id_created string

	var create_response_status int
	var create_response_body string
	var place_holder_map6 = make(map[string]string)
	var meta_data_map6 = make(map[string]string)
	var ariaclientIdStandardBM string
	var ariaAuthStandardBM string

	var token_url = "http://dev.oidc.cloud.intel.com.kind.local"
	var idc_global_url = "https://dev.api.cloud.intel.com.kind.local"
	var idc_regional_url = "https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local"

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../data", "bmaas_input.json")
		financials_utils.LoadE2EConfig("../../data", "billing.json")
		//base_url = compute_utils.GetBaseUrl()
		base_url = idc_regional_url
		token = authentication.GetBearerTokenViaFrisby(token_url)
		ariaclientIdStandardBM = financials_utils.GetAriaClientNo()
		ariaAuthStandardBM = financials_utils.GetariaAuthKey()
		fmt.Println("ariaAuthStandardBM", ariaAuthStandardBM)
		fmt.Println("ariaclientIdStandardBM", ariaclientIdStandardBM)
	})

	XIt("Create cloud account enroll", func() {
		log.Printf("Starting the Cloud Account Creation via API...")
		groups := "DevCloud%20Console%20Standard"
		tid := utils.GenerateString(12)
		username := utils.GenerateString(10) + "@example.com"
		enterpriseId := utils.GenerateString(12)
		idp := "intelcorpintb2c.onmicrosoft.com"
		cloudaccount_enroll_url := idc_global_url + "/v1/cloudaccounts/enroll"
		enroll_token_endpoint := fmt.Sprintf("%s/%s?idp=%s&tid=%s&email=%s&groups=%s&enterpriseId=%s", token_url, "token", idp, tid, username, groups, enterpriseId)
		token_response, _ := financials.GetEnrollToken(enroll_token_endpoint)
		fmt.Println("enroll_token_endpoint", enroll_token_endpoint)
		enroll_payload := `{"premium":false}`
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_enroll_url, "Bearer "+token_response, enroll_payload)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_STANDARD"`)).To(BeTrue(), "Validation failed on Enrollment of standard user")
		place_holder_map6["cloud_account_id"] = cloud_account_created
		place_holder_map6["cloud_account_type"] = cloudaccount_type
		fmt.Println("cloudAccount_id", cloud_account_created)
	})

	It("Create cloud account", func() {
		username := utils.GenerateString(10) + "@intel.com"
		tid := utils.GenerateString(12)
		oid := utils.GenerateString(12)
		cloudaccount_url := idc_global_url + "/v1/cloudaccounts"
		cloudaccount_payload := fmt.Sprintf(`{"name":"%s","owner":"%s","tid":"%s","oid":"%s","type":"ACCOUNT_TYPE_INTEL"}`, "standard-cloudaccount-bm", username, tid, oid)
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_url, token, cloudaccount_payload)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "id").String()
		place_holder_map6["cloud_account_id"] = cloud_account_created
		log.Printf("cloudAccount_id: %s", cloud_account_created)
	})

	It("Create vnet", func() {
		log.Printf("Starting the VNet Creation via API...")
		// form the endpoint and payload
		vnet_name := "standard-vnet-bm"
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
		ssh_key_name := "e2e-standardtc-bm"
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
		instance_name := "e2e-su-bm"
		payload_expected := fmt.Sprintf(`"name":"%s"`, instance_name)
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), instance_name, instanceType, "ubuntu-22.04-server-cloudimg-amd64-latest", ssh_publickey_created, vnet_created)
		log.Printf("instance_payload: %s", instance_payload)
		place_holder_map6["instance_type"] = instanceType
		log.Printf("instance_type: %s", instanceType)

		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, token, instance_payload)
		log.Printf("Create instance status %d response %s", create_response_status, create_response_body)
		Expect(create_response_status).To(Equal(200), "Failed to create BM instance")
		Expect(strings.Contains(create_response_body, payload_expected)).To(BeTrue(), "Failed to create BM instance, response validation failed")
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		place_holder_map6["resource_id"] = instance_id_created
		place_holder_map6["instance_creation_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	It("Get BMH device by resource ID", func() {
		log.Printf("Starting the BMH Device Retrieval via Kube...")
		time.Sleep(5 * time.Second)
		response_bmh, err := kube.GetBmhByConsumer(instance_id_created)
		Expect(err).Error().ShouldNot(HaveOccurred())
		Expect(response_bmh.Spec.ConsumerRef.Name).To(Equal(instance_id_created))
		place_holder_map6["device_name"] = response_bmh.ObjectMeta.Name
	})

	It("Validate BMH device is provisoned", func() {
		log.Printf("Starting the BMH Device Validation via Kube...")
		succeded, err := kube.CheckBMHState(place_holder_map6["device_name"], "provisioned", 900)
		Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+
			"unable to reach state within expected time")
		Expect(succeded).To(Equal(true))
	})

	It("Validate instance is ready and get its details", func() {
		log.Printf("Starting the Instance Validation via API...")
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"

		var response_status int
		var response_body string
		start_time := time.Now();

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

		place_holder_map6["machine_ip"] = gjson.Get(response_body, "status.interfaces.0.addresses.0").String()
		place_holder_map6["proxy_ip"] = gjson.Get(response_body, "status.sshProxy.proxyAddress").String()
		place_holder_map6["proxy_user"] = gjson.Get(response_body, "status.sshProxy.proxyUser").String()
		log.Printf("Instance IP Address is: " + place_holder_map6["machine_ip"])
	})

	It("SSH into the BM instance", func() {
		log.Printf("Starting the Instance SSH connection via Ansible...")
		inventory_raw_data, _ := compute_utils.ConvertFileToString("../../ansible-files", "inventory_raw.ini")
		log.Printf("Inventory raw data is: " + inventory_raw_data)
		inventory_generated := compute_utils.EnrichInventoryData(inventory_raw_data, place_holder_map6["proxy_ip"], place_holder_map6["proxy_user"], place_holder_map6["machine_ip"], "~/.ssh/id_rsa")
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
		start_time := time.Now();
		
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
		succeded, err := kube.CheckBMHState(place_holder_map6["device_name"], "available", 900)
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

	It("Delete the vnet created and validate deletion", func() {
		log.Printf("Starting the Vnet Deletion via API...")
		vnet_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"

		// delete the vnet created
		delete_response_byname_status, _ := frisby.DeleteVnetByName(vnet_endpoint, token, vnet_created)
		Expect(delete_response_byname_status).To(Equal(200), "Failed to delete vnet")
		time.Sleep(5 * time.Second)

		// validate the deletion
		get_response_status, _ := frisby.GetVnetByName(vnet_endpoint, token, vnet_created)
		Expect(get_response_status).To(Equal(404), "Vnet shouldn't be found")
	})

	// product catalogue enrichment
	It("Get the created instance details from product catalog and validate", func() {
		log.Printf("Starting the Instance Retrieval from Product Catalog via API...")
		// product_id := place_holder_map6["resource_id"]
		payload_expected := fmt.Sprintf(`"name":"%s"`, instanceType)
		product_filter := fmt.Sprintf(`{%s}`, payload_expected)
		response_status, response_body := financials.GetProducts(idc_global_url, token, product_filter)
		var structResponse GetProductsResponse
		json.Unmarshal([]byte(response_body), &structResponse)
		fmt.Println("structResponse", structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		Expect(strings.Contains(response_body, payload_expected)).To(BeTrue(), "Validation failed on instance retrieval")
		meta_data_map6["Highlight"] = structResponse.Products[0].Metadata.Highlight
		meta_data_map6["disks.size"] = structResponse.Products[0].Metadata.Disks
		meta_data_map6["instanceType"] = structResponse.Products[0].Metadata.InstanceType
		meta_data_map6["memory.size"] = structResponse.Products[0].Metadata.Memory
		meta_data_map6["region"] = structResponse.Products[0].Metadata.Region
		fmt.Println("meta_data_map6", meta_data_map6)
	})

	// metering
	It("Get Metering data related to product and validate", func() {
		log.Printf("Starting the Metering Data Retrieval via API...")
		search_payload := financials_utils.EnrichMmeteringSearchPayload(compute_utils.GetMeteringSearchPayload(),
			place_holder_map6["resource_id"], place_holder_map6["cloud_account_id"])
		fmt.Println("search_payload", search_payload)
		metering_api_base_url := idc_global_url + "/v1/meteringrecords"
		response_status, response_body := financials.SearchAllMeteringRecords(metering_api_base_url, token, search_payload)
		// metering_record_cloudAccountId := gjson.Get(response_body,"result.cloudAccountId").String()
		//metering_record_resourceId := gjson.Get(response_body,"result.resourceId").String()
		//metering_record_transactionId := gjson.Get(response_body,"result.transactionId").String()
		// metering_record_reported := gjson.Get(response_body,"reported").String()
		//metering_record_reported := gjson.Get(response_body,"result.reported").String()
		//metering_record_instanceType := gjson.Get(response_body,"result.instanceType").String()
		var metering_record_cloudAccountId string
		var metering_record_resourceId string
		var metering_record_reported string
		var metering_record_transactionId string
		var metering_record_instanceType string
		result := gjson.Parse(response_body)
		arr := gjson.Get(result.String(), "..#.result")
		for _, v := range arr.Array() {
			metering_record_cloudAccountId = v.Get("cloudAccountId").String()
			metering_record_resourceId = v.Get("resourceId").String()
			metering_record_reported = v.Get("reported").String()
			metering_record_transactionId = v.Get("transactionId").String()
			metering_record_instanceType = v.Get("properties.instanceType").String()
			break
		}
		fmt.Println("metering_record_cloudAccountId", metering_record_cloudAccountId)
		fmt.Println("metering_record_resourceId", metering_record_resourceId)
		fmt.Println("metering_record_reported", metering_record_reported)
		fmt.Println("metering_record_transactionId", metering_record_transactionId)
		fmt.Println("metering_record_instanceType", metering_record_instanceType)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		Expect(place_holder_map6["cloud_account_id"]).To(Equal(metering_record_cloudAccountId), "Validation failed on cloud account id retrieval")
		Expect(place_holder_map6["resource_id"]).To(Equal(metering_record_resourceId), "Validation failed on resource id retrieval")
		Expect(meta_data_map6["instanceType"]).To(Equal(metering_record_instanceType), "Validation failed on instance retrieval")
		place_holder_map6["transactionId"] = metering_record_transactionId
		place_holder_map6["reported"] = metering_record_reported
	})

	XIt("Validate billing account is Created for Standard User", func() {
		log.Printf("Starting the Billing Account Validation via API...")
		response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(place_holder_map6["cloud_account_id"], ariaclientIdStandardBM, ariaAuthStandardBM)
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		fmt.Println("billing_account_response", responseBody)
		Expect(strings.Contains(responseBody, `"error_msg" : "account does not exist"`)).To(BeTrue(), "Validation failed fetching billing account details from aria")
	})

	It("Delete cloud account", func() {
		log.Printf("Starting the Cloud Account Deletion via API...")
		url := idc_global_url + "/v1/cloudaccounts"

		// delete the cloud account created
		delete_Cacc, _ := financials.DeleteCloudAccountById(url, token, place_holder_map6["cloud_account_id"])
		Expect(delete_Cacc).To(Equal(200), "Failed to delete Cloud Account")

		// validate the deletion
		get_response_status, _ := financials.GetCloudAccountById(url, token, place_holder_map6["cloud_account_id"])
		Expect(get_response_status).To(Equal(404), "Cloud account shouldn't be found")
	})
})
