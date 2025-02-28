//go:build pu_vm || All || VM
// +build pu_vm All VM

package vmaas

import (
	//"bytes"
	"encoding/json"
	"fmt"

	//"os/exec"
	"goFramework/framework/authentication"
	"strconv"
	"strings"

	//"goFramework/framework/library/financials/billing"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"

	"goFramework/utils"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var create_response_status_pu int
var create_response_body_pu string
var place_holder_map3 = make(map[string]string)
var meta_data_map3 = make(map[string]string)
var ariaclientId string

// var base_url string
// var token string

var _ = Describe("VM Instance", Ordered, Label("Compute-VMaaS-E2E"), func() {
	var base_url string
	var token string
	var cloud_account_created string
	var vnet_created string
	var ssh_publickey_name_created string
	var instance_id_created string
	var ariaAuth string

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../data", "vmaas_input.json")
		financials_utils.LoadE2EConfig("../../data", "billing.json")
		base_url = compute_utils.GetBaseUrl()
		token = authentication.GetBearerTokenViaFrisby(base_url)
		ariaclientId = financials_utils.GetAriaClientNo()
		ariaAuth = financials_utils.GetariaAuthKey()
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("ariaclientId", ariaclientId)
	})

	It("Create cloud account", func() {
		groups := "DevCloud%20Console%20Standard"
		tid := utils.GenerateString(12)
		username := utils.GenerateString(10) + "@example.com"
		enterpriseId := utils.GenerateString(12)
		idp := "intelcorpintb2c.onmicrosoft.com"
		cloudaccount_enroll_url := base_url + "/v1/cloudaccounts/enroll"
		// Generating token wtih payload for cloud account creation with enroll API
		enroll_token_endpoint := fmt.Sprintf("%s/%s?idp=%s&tid=%s&email=%s&groups=%s&enterpriseId=%s", base_url, "token", idp, tid, username, groups, enterpriseId)
		token_response, _ := financials.GetEnrollToken(enroll_token_endpoint)
		fmt.Println("enroll_token_endpoint", enroll_token_endpoint)
		enroll_payload := `{"premium":true}`
		// cloud account creation
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_enroll_url, "Bearer "+token_response, enroll_payload)
		//fmt.Println("cloudaccount_creation_body", cloudaccount_creation_body)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of standard user")
		place_holder_map3["cloud_account_id"] = cloud_account_created
		place_holder_map3["cloud_account_type"] = cloudaccount_type
		fmt.Println("cloudAccount_id", cloud_account_created)
		// cloud_account_created = "730302673781"
		// place_holder_map3["cloud_account_id"] = cloud_account_created
		//place_holder_map3["cloud_account_type"] = cloudaccount_type
	})

	It("Create vnet with name", func() {
		fmt.Println("Starting the VNet Creation via API...")
		// form the endpoint and payload
		vnet_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
		vnet_name := "Autovnet-" + utils.GenerateSSHKeyName(4)
		vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)

		// hit the api
		vnet_creation_status, vnet_creation_body := frisby.CreateSSHKey(vnet_endpoint, token, vnet_payload)
		vnet_created = gjson.Get(vnet_creation_body, "metadata.name").String()
		Expect(vnet_creation_status).To(Equal(200), "Failed to create VNet")
		Expect(vnet_name).To(Equal(vnet_created), "Failed to create Vnet, response validation failed")
		//Expect(strings.Contains(vnet_creation_body, `"name":"premium-vnet"`)).To(BeTrue(), "Failed to create Vnet, response validation failed")
	})

	It("Create ssh public key with name", func() {
		fmt.Println("Starting the SSH-Public-Key Creation via API...")
		// form the endpoint and payload
		ssh_publickey_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
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
		fmt.Println("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		vm_name_pu := "autovm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name_pu, instanceType, "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
		place_holder_map3["instance_type"] = instanceType

		// hit the api
		create_response_status_pu, create_response_body_pu = frisby.CreateInstance(instance_endpoint, token, instance_payload)
		VMName_pu := gjson.Get(create_response_body_pu, "metadata.name").String()
		Expect(create_response_status_pu).To(Equal(200), "Failed to create VM instance")
		Expect(vm_name_pu).To(Equal(VMName_pu), "Failed to create VM instance, resposne validation failed")
		//Expect(strings.Contains(create_response_body_pu, `"name":"e2e-puvm-instance2"`)).To(BeTrue(), "Failed to create VM instance, resposne validation failed")
		place_holder_map3["instance_creation_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	It("Get the created instance and validate", func() {
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		// Adding a sleep because it seems to take some time to reflect the creation status
		time.Sleep(180 * time.Second)
		instance_id_created = gjson.Get(create_response_body_pu, "metadata.resourceId").String()
		response_status, response_body := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)

		Expect(response_status).To(Equal(200), "Failed to retrieve VM instance")
		Expect(strings.Contains(response_body, `"phase":"Ready"`)).To(BeTrue(), "Validation failed on instance retrieval")
		place_holder_map3["resource_id"] = instance_id_created
		place_holder_map3["machine_ip"] = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()
		fmt.Println("IP Address is :" + place_holder_map3["machine_ip"])
	})

	// It("SSH into the instance", func() {
	// 	// SSH to the instance goes here
	// 	inventory_raw_data, _ := compute_utils.ConvertFileToString("../../ansible-files", "inventory_raw.ini")
	// 	fmt.Println("Inventory Raw Data is :" + inventory_raw_data)
	// 	inventory_generated := compute_utils.EnrichInventoryData(inventory_raw_data, proxyIp, "guest", place_holder_map3["machine_ip"], "~/.ssh/id_rsa")
	// 	fmt.Println("Inventory generated is :" + inventory_generated)
	// 	compute_utils.WriteStringToFile("../../ansible-files", "inventory.ini", inventory_generated)

	// 	// Get the pod details after restart
	// 	var output bytes.Buffer
	// 	get_pod_cmd := exec.Command("ansible-playbook", "-i", "../../ansible-files/inventory.ini", "../../ansible-files/ssh-and-apt-get-on-vm.yml")
	// 	get_pod_cmd.Stdout = &output
	// 	error := get_pod_cmd.Run()
	// 	if error != nil {
	// 		fmt.Println("Execution of ansible playbook is not successful: ", error)
	// 	}

	// 	// Log the ansible output
	// 	ansible_output := strings.Split(output.String(), "\n")
	// 	fmt.Println(ansible_output)

	// 	// keeping the sleep time for billing purpose
	// 	time.Sleep(300 * time.Second)
	// })

	It("Delete the created instance", func() {
		instance_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
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
		place_holder_map3["instance_deletion_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	It("Delete the SSH key created", func() {
		fmt.Println("Delete the SSH-Public-Key Created above...")
		ssh_publickey_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"

		delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_name_created)
		Expect(delete_response_byname_status).To(Equal(200), "assert ssh-public-key deletion response code")
	})

	// Deletion of vnet is not working - Bug is open for the same
	/*It("Delete the Vnet created", func() {
		fmt.Println("Delete the Vnet Created above...")
		vnet_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
		// Deletion of vnet via name
		delete_response_byname_status, _ := frisby.DeleteVnetByName(vnet_endpoint, token, vnet_created)
		Expect(delete_response_byname_status).To(Equal(200), "assert vnet deletion response code")
	})*/

	// product catalogue enrichment
	// product catalogue enrichment
	It("Get the created instance details from poduct catalog and validate", func() {
		// product_id := place_holder_map["resource_id"]
		// product_id := "3bc52387-da79-4947-a562-ab7a88c38e1d"
		product_filter := `{"name":"vm-spr-tny"}`
		response_status, response_body := financials.GetProducts(base_url, token, product_filter)
		var structResponse GetProductsResponse
		json.Unmarshal([]byte(response_body), &structResponse)
		fmt.Println("structResponse", structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		Expect(strings.Contains(response_body, `"name":"vm-spr-tny"`)).To(BeTrue(), "Validation failed on instance retrieval")
		meta_data_map3["Highlight"] = structResponse.Products[0].Metadata.Highlight
		meta_data_map3["disks.size"] = structResponse.Products[0].Metadata.Disks
		meta_data_map3["instanceType"] = structResponse.Products[0].Metadata.InstanceType
		meta_data_map3["memory.size"] = structResponse.Products[0].Metadata.Memory
		meta_data_map3["region"] = structResponse.Products[0].Metadata.Region
	})

	// metering
	It("Get Metering data related to product and validate", func() {
		search_payload := financials_utils.EnrichMmeteringSearchPayload(compute_utils.GetMeteringSearchPayload(),
			place_holder_map3["resource_id"], place_holder_map3["cloud_account_id"])
		fmt.Println("search_payload", search_payload)
		metering_api_base_url := base_url + "/v1/meteringrecords"
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
		Expect(place_holder_map3["cloud_account_id"]).To(Equal(metering_record_cloudAccountId), "Validation failed on cloud account id retrieval")
		Expect(place_holder_map3["resource_id"]).To(Equal(metering_record_resourceId), "Validation failed on resource id retrieval")
		Expect(meta_data_map3["instanceType"]).To(Equal(metering_record_instanceType), "Validation failed on instance retrieval")
		place_holder_map3["transactionId"] = metering_record_transactionId
		place_holder_map3["reported"] = metering_record_reported
	})

	// billing
	// Make sure billing account is not created for Premium user
	It("Validate billing account is Created for Premium User", func() {
		client_acct_id := "idc." + cloud_account_created
		response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(place_holder_map3["cloud_account_id"], ariaclientId, ariaAuth)
		client_acc_id1 := gjson.Get(responseBody, "client_acct_id").String()
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		Expect(client_acct_id).To(Equal(client_acc_id1), "Validation failed fetching billing account details from aria")
		Expect(strings.Contains(responseBody, `"master_plan_count" : 1`)).To(BeTrue(), "Validation failed fetching billing account details from aria")
		Expect(strings.Contains(responseBody, `"error_msg" : "OK"`)).To(BeTrue(), "Validation failed fetching billing account details from aria")
	})

	It("Delete cloud account", func() {
		fmt.Println("Delete cloud account")
		url := base_url + "/v1/cloudaccounts"
		delete_Cacc, _ := financials.DeleteCloudAccountById(url, token, place_holder_map3["cloud_account_id"])
		Expect(delete_Cacc).To(Equal(200), "Failed to delete Cloud Account")
	})

	It("Validating deletion of cloud account by retrieving cloud account details", func() {
		fmt.Println("Validating deletion of cloud account by retrieving cloud account details")
		url := base_url + "/v1/cloudaccounts"
		resp_body, _ := financials.GetCloudAccountById(url, token, place_holder_map3["cloud_account_id"])
		fmt.Println("resp_body", resp_body)
	})

})
