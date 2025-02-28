//go:build iu_vm_80 || All || pu_su_iu
// +build iu_vm_80 All pu_su_iu

package vmaas

import (
	"encoding/json"
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
	var create_response_status int
	var create_response_body string
	var place_holder_map = make(map[string]string)
	var meta_data_map = make(map[string]string)
	var ariaclientId string
	var resourceInfo testsetup.ResourcesInfo

	var base_url string
	var compute_url string
	var userName string
	var token string
	var userToken string
	var cloud_account_created string
	var vnet_created string
	var ssh_publickey_name_created string
	var instance_id_created string
	var ariaAuth string

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../../financials/data", "vmaas_input.json")
		financials_utils.LoadE2EConfig("../../../financials/data", "billing.json")
		auth.Get_config_file_data("../../../financials/data/config.json")
		userName := auth.Get_UserName("Intel")
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

	It("Start Credit Expire scheduler", func() {
		// err := testsetup.StartSchedulers(base_url, token, "START_CREDIT_USAGE_SCHEDULER")
		// if err != nil {
		//     logger.Logf.Errorf("Starting scheduler for credit usage failed")
		// }
		// err = testsetup.StartSchedulers(base_url, token, "START_CREDIT_EXPIRY_SCHEDULER")
		// if err != nil {
		//     logger.Logf.Errorf("Starting scheduler for credit expiry failed")
		// }
		time.Sleep(10 * time.Second)
	})

	It("Create cloud account", func() {
		cloudaccount_enroll_url := base_url + "/v1/cloudaccounts/enroll"
		// Generating token wtih payload for cloud account creation with enroll API
		// cloud account creation
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, false)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_INTEL"`)).To(BeTrue(), "Validation failed on Enrollment of Intel user")
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

	// It("Create paid vm instance when iu user have zero cloud credits", func() {
	// 	fmt.Println("Starting the Instance Creation via API...")
	// 	// form the endpoint and payload
	// 	instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
	// 	vm_name_iu := "autovm-" + utils.GenerateSSHKeyName(4)
	// 	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name_iu, "vm-spr-med", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
	// 	fmt.Println("instance_payload", instance_payload)
	// 	place_holder_map["instance_type"] = "vm-spr-med"
	// 	// hit the api
	// 	create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
	// 	Expect(create_response_status).To(Equal(403), "Expected response code on paid instance for iu user should be 403 when cloud credits are zero")
	// 	Expect(strings.Contains(create_response_body, `"message":"paid service not allowed"`)).To(BeTrue(), "Failed to validate ")

	// })

	It("Create Cloud credits for intel user by redeeming coupons", func() {
		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/billing/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 0.1, "idc_billing@intel.com", 1)
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
		VMName_iu := gjson.Get(create_response_body, "metadata.name").String()
		Expect(create_response_status).To(Equal(200), "Failed to create VM instance")
		Expect(vm_name_iu).To(Equal(VMName_iu), "Failed to create VM instance, resposne validation failed")
		place_holder_map["instance_creation_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	It("Get the created instance and validate", func() {
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		// Adding a sleep because it seems to take some time to reflect the creation status
		time.Sleep(180 * time.Second)
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		response_status, response_body := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
		fmt.Println("response_status", response_status)
		fmt.Println("response_body", response_body)
		Expect(response_status).To(Equal(200), "Failed to retrieve VM instance")
		Expect(strings.Contains(response_body, `"phase":"Ready"`)).To(BeTrue(), "Validation failed on instance retrieval")
		place_holder_map["resource_id"] = instance_id_created
		place_holder_map["machine_ip"] = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()
		fmt.Println("IP Address is :" + place_holder_map["machine_ip"])
	})
	// It("SSH into the instance", func() {
	// 	// SSH to the instance goes here
	// 	inventory_raw_data, _ := compute_utils.ConvertFileToString("../../ansible-files", "inventory_raw.ini")
	// 	fmt.Println("Inventory Raw Data is :" + inventory_raw_data)
	// 	inventory_generated := compute_utils.EnrichInventoryData(inventory_raw_data, proxyIp, place_holder_map["machine_ip"], "~/.ssh/id_rsa")
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

	// Deletion of vnet is not working - Bug is open for the same
	// It("Delete the Vnet created", func() {
	// 	fmt.Println("Delete the Vnet Created above...")
	// 	vnet_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
	// 	// Deletion of vnet via name
	// 	delete_response_byname_status, _ := frisby.DeleteVnetByName(vnet_endpoint, token, vnet_created)
	// 	Expect(delete_response_byname_status).To(Equal(200), "assert vnet deletion response code")
	// })

	// product catalogue enrichment
	It("Get the created instance details from poduct catalog and validate", func() {
		product_filter := fmt.Sprintf(`{
			"cloudaccountId": "%s",
			"productFilter": { "name":"vm-spr-med"}
		}`, place_holder_map["cloud_account_id"])
		response_status, response_body := financials.GetProducts(base_url, token, product_filter)
		var structResponse GetProductsResponse
		json.Unmarshal([]byte(response_body), &structResponse)
		fmt.Println("structResponse", structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		Expect(strings.Contains(response_body, `"name":"vm-spr-med"`)).To(BeTrue(), "Validation failed on instance retrieval")
		meta_data_map["Highlight"] = structResponse.Products[0].Metadata.Highlight
		meta_data_map["disks.size"] = structResponse.Products[0].Metadata.Disks
		meta_data_map["instanceType"] = structResponse.Products[0].Metadata.InstanceType
		meta_data_map["memory.size"] = structResponse.Products[0].Metadata.Memory
		meta_data_map["region"] = structResponse.Products[0].Metadata.Region
		fmt.Println("meta_data_map", meta_data_map)
	})

	// metering
	It("Get Metering data related to product and validate", func() {
		search_payload := financials_utils.EnrichMmeteringSearchPayload(compute_utils.GetMeteringSearchPayload(),
			place_holder_map["resource_id"], place_holder_map["cloud_account_id"])
		fmt.Println("search_payload", search_payload)
		metering_api_base_url := base_url + "/v1/meteringrecords"
		response_status, response_body := financials.SearchAllMeteringRecords(metering_api_base_url, token, search_payload)
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
		Expect(place_holder_map["cloud_account_id"]).To(Equal(metering_record_cloudAccountId), "Validation failed on cloud account id retrieval")
		Expect(place_holder_map["resource_id"]).To(Equal(metering_record_resourceId), "Validation failed on resource id retrieval")
		Expect(meta_data_map["instanceType"]).To(Equal(metering_record_instanceType), "Validation failed on instance retrieval")
		place_holder_map["transactionId"] = metering_record_transactionId
		place_holder_map["reported"] = metering_record_reported
	})

	It("Validate billing account is not Created for Intel User", func() {
		response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(place_holder_map["cloud_account_id"], ariaclientId, ariaAuth)
		fmt.Println("ariaclientId", ariaclientId)
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("responseBody", responseBody)
		fmt.Println(place_holder_map["cloud_account_id"])
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		Expect(strings.Contains(responseBody, `"error_msg" : "account does not exist"`)).To(BeTrue(), "Validation failed fetching billing account details from aria")
	})

	It("Wait for some time for the instance to run", func() {
		resourceInfo, _ = testsetup.GetInstanceDetails(userName, base_url, token, compute_url)
		time.Sleep(10 * time.Minute)

	})

	// It("Delete the SSH key created", func() {
	// 	fmt.Println("Delete the SSH-Public-Key Created above...")
	// 	ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
	// 	delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_name_created)
	// 	Expect(delete_response_byname_status).To(Equal(200), "assert ssh-public-key deletion response code")
	// })

	// It("Delete the created instance", func() {
	// 	instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
	// 	time.Sleep(10 * time.Second)
	// 	// delete the instance created
	// 	delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, token, instance_id_created)
	// 	Expect(delete_response_status).To(Equal(200), "Failed to delete VM instance")
	// 	time.Sleep(5 * time.Second)
	// 	// validate the deletion
	// 	// Adding a sleep because it seems to take some time to reflect the deletion status
	// 	time.Sleep(1 * time.Minute)
	// 	get_response_status, _ := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
	// 	Expect(get_response_status).To(Equal(404), "Resource shouldn't be found")
	// 	place_holder_map["instance_deletion_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	// })

	It("Start Credit Expire scheduler", func() {
		// err := testsetup.StartSchedulers(base_url, token, "START_POST_USAGE_SCHEDULER")
		// if err != nil {
		//     logger.Logf.Errorf("Starting scheduler for credit usage failed")
		// }
		// //time.Sleep(3 * time.Minute)
		// err = testsetup.StartSchedulers(base_url, token, "START_CREDIT_USAGE_SCHEDULER")
		// if err != nil {
		//     logger.Logf.Errorf("Starting scheduler for credit usage failed")
		// }
		// //time.Sleep(3 * time.Minute)
		// err = testsetup.StartSchedulers(base_url, token, "START_CREDIT_EXPIRY_SCHEDULER")
		// if err != nil {
		//     logger.Logf.Errorf("Starting scheduler for credit expiry failed")
		// }
		time.Sleep(10 * time.Second)
	})

	It("Validate Credit Depletion", func() {
		//time.Sleep(6 * time.Minute)
		usageAmount, _ := testsetup.GetUsageAndValidateTotalUsage(userName, resourceInfo, base_url, token, compute_url)
		fmt.Println("Usage of free instance ", usageAmount)
		baseUrl := base_url + "/v1/billing/credit"
		response_status, responseBody := financials.GetCredits(baseUrl, token, place_holder_map["cloud_account_id"])
		//Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
		remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
		fmt.Println("Get Credits response  ", response_status)
		fmt.Println("Used amount from credits ", usedAmount)
		fmt.Println("remainingAmount  from credits ", remainingAmount)
		fmt.Println("usageAmount  from usage ", usageAmount)
		Expect(usageAmount).To(Equal(usedAmount), "Amount calculated from usage didnt match with used amount from credit")
		Expect(remainingAmount).To(Equal(0), "remainingAmount from credit is not equal to zero")
		Expect(usedAmount).To(Equal(0.1), "Used amount from credit is not equal to the coupon amount")
	})

	It("Validate Notifications sent to user for 80% and 100% usage", func() {
		response_status, responseBody := financials.GetNotificationsShortPoll(base_url, token, place_holder_map["cloud_account_id"])
		fmt.Println("Response Status to get notifications", response_status)
		fmt.Println("Response Body to get notifications", responseBody)
		numberofNotifications := gjson.Get(responseBody, "numberOfNotifications").Int()
		Expect(numberofNotifications).Should(BeNumerically(">=", 1), "Validation failed on numberofNotifications in notifications")
		result := gjson.Parse(responseBody)
		notificationArray := gjson.Get(result.String(), "notifications")
		notificationArray.ForEach(func(key, value gjson.Result) bool {
			data := value.String()
			cloudAccId := gjson.Get(data, "cloudAccountId").String()
			notificationType := gjson.Get(data, "notificationType").String()
			serviceName := gjson.Get(data, "serviceName").String()
			status := gjson.Get(data, "status").String()
			severity := gjson.Get(data, "severity").String()
			message := gjson.Get(data, "message").String()

			// Assertions

			Expect(place_holder_map["cloud_account_id"]).To(Equal(cloudAccId), "Validation failed on cloud account id in notifications")
			Expect(notificationType).To(Equal("CREDIT_THRESHOLD_REACHED"), "Validation failed on notificationType in notifications")
			Expect(serviceName).To(Equal("billing"), "Validation failed on serviceName id in notifications")
			Expect(status).To(Equal("active"), "Validation failed on status id in notifications")
			Expect(severity).To(Equal("low"), "Validation failed on severity id in notifications")
			Expect(message).To(Equal("cloud credit threshold reached"), "Validation failed on message id in notifications")
			return true // keep iterating
		})

	})

	It("Delete the SSH key created", func() {
		fmt.Println("Delete the SSH-Public-Key Created above...")
		ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
		delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_name_created)
		Expect(delete_response_byname_status).To(Equal(200), "assert ssh-public-key deletion response code")
	})

	It("Delete the created instance", func() {
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
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
		place_holder_map["instance_deletion_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
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
