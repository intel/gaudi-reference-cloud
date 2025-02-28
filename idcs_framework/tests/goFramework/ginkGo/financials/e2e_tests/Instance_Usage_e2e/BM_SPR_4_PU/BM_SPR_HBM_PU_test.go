package BM_SPR_HBM_PU_test

import (
	"encoding/json"
	"fmt"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"strconv"
	"strings"
	"time"

	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Premium User Product Usage - VM Medium", Ordered, Label("Usages-e2e"), func() {
	instanceType := "bm-spr"
	instanceOS := "ubuntu-22.04-spr-metal-cloudimg-amd64-v20240115"
	It("Delete cloud account", func() {
		fmt.Println("Delete cloud account")
		url := base_url + "/v1/cloudaccounts"
		cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, token)
		if err == nil {
			financials.DeleteCloudAccountById(url, token, cloudAccId)
		}
	})

	// CREATE CLOUDACCOUNT
	It("Create cloud account", func() {
		cloudaccount_enroll_url := base_url + "/v1/cloudaccounts/enroll"
		// Generating token wtih payload for cloud account creation with enroll API
		token_response, _ := auth.Get_Azure_Bearer_Token(userName)
		userToken = "Bearer " + token_response
		enroll_payload := `{"premium":true}`
		// cloud account creation
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccount(cloudaccount_enroll_url, userToken, enroll_payload)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of Intel user")
		place_holder_map["cloud_account_id"] = cloud_account_created
		place_holder_map["cloud_account_type"] = cloudaccount_type
		fmt.Println("cloudAccount_id", cloud_account_created)
	})

	It("Validate Premium cloudAccount", func() {
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).NotTo(BeNil(), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
	})

	//CREATE VNET
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
		sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, token, ssh_publickey_payload)
		Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key")
		ssh_publickey_name_created = gjson.Get(sshkey_creation_body, "metadata.name").String()
		Expect(sshkey_name).To(Equal(ssh_publickey_name_created), "Failed to create SSH Public key, response validation failed")
	})

	// LAUNCH INSTANCE WITHOUT CREDITS
	It("Create paid BM test instance", func() {
		logger.Logf.Info("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		bm_name_iu := "autobm-test-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, instanceType, "ubuntu-22.04-server-cloudimg-amd64-latest", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = instanceType
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
		Expect(create_response_status).To(Equal(403), "Expected response code on paid instance for iu user should be 403 when cloud credits are zero")
		Expect(strings.Contains(create_response_body, `"message":"paid service not allowed"`)).To(BeTrue(), "Failed to validate ")
	})

	// ADD CREDITS TO ACCOUNT
	It("Create Cloud credits for intel user by redeeming coupons", func() {
		logger.Logf.Info("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/cloudcredits/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 100, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		logger.Logf.Info("Redeem credits to current user...")
		redeem_coupon_endpoint := base_url + "/v1/cloudcredits/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), cloud_account_created)
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
		fmt.Println("Payload", redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Failed to redeem coupon")
	})

	// CREATE NEW INSTANCE
	It("Create paid test instance", func() {
		fmt.Println("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		bm_name_iu := "autobm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, instanceType, instanceOS, ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = instanceType
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
		BMName_iu := gjson.Get(create_response_body, "metadata.name").String()
		Expect(create_response_status).To(Equal(200), "Failed to create BM instance")
		Expect(bm_name_iu).To(Equal(BMName_iu), "Failed to create BM instance, resposne validation failed")
		place_holder_map["instance_creation_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	// GET THE INSTANCE
	It("Get the created instance and validate", func() {
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		// Adding a sleep because it seems to take some time to reflect the creation status
		time.Sleep(180 * time.Second)
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		response_status, response_body := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
		Expect(response_status).To(Equal(200), "Failed to retrieve BM instance")
		Eventually(func() bool {
			response_status, response_body := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
			Expect(response_status).To(Equal(200), "Failed to retrieve VM instance")
			if strings.Contains(response_body, `"phase":"Ready"`) {
				return true
			}
			fmt.Println("Waiting for instance to be ready...")
			return false
		}, 1*time.Hour, 3*time.Minute).Should(BeTrue(), "Validation failed on instance retrieval, Image is not in ready phase")
		place_holder_map["resource_id"] = instance_id_created
		place_holder_map["machine_ip"] = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()
		fmt.Println("IP Address is :" + place_holder_map["machine_ip"])
	})

	// product catalogue enrichment
	It("Get the created instance details from poduct catalog and validate", func() {
		product_filter := fmt.Sprintf(`{
			"cloudaccountId": "%s",
			"productFilter": { "accountType": "%s", "name": "%s"}
		}`, place_holder_map["cloud_account_id"], place_holder_map["cloud_account_type"], instanceType)
		response_status, response_body := financials.GetProducts(base_url, token, product_filter)
		var structResponse GetProductsResponse
		json.Unmarshal([]byte(response_body), &structResponse)
		fmt.Println("structResponse", structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		Expect(strings.Contains(response_body, `"name":"bm-spr"`)).To(BeTrue(), "Validation failed on instance retrieval")
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

	It("Validate billing account is created for Premium User", func() {
		response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(place_holder_map["cloud_account_id"], ariaclientId, ariaAuth)
		fmt.Println("ariaclientId", ariaclientId)
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("responseBody", responseBody)
		fmt.Println(place_holder_map["cloud_account_id"])
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		//Expect(strings.Contains(responseBody, `"error_msg" : "account does not exist"`)).To(BeTrue(), "Validation failed fetching billing account details from aria")
	})

	It("Wait for the usages to record", func() {
		time.Sleep(financials_utils.GetUsageTime() * time.Minute)
	})

	It("Validate Usage calculation for the paid product and validate credits depletion", func() {
		var err error
		var usage float64
		//wait for some time to fetch some usage
		zeroamt := 0
		resourceInfo, err = testsetup.GetInstanceDetails(userName, base_url, token, compute_url)
		Expect(err).To(BeNil(), "Failed: Failed to get instance details")

		Eventually(func() bool {
			fmt.Println("RESOURCE...", resourceInfo)
			usage1, _ := testsetup.GetUsageAndValidateTotalUsage(userName, resourceInfo, base_url, token, compute_url)
			fmt.Println("USAGE...", usage1)
			if usage1 > float64(zeroamt) {
				return true
			}
			fmt.Println("trying again...")
			return false
		}, 6*time.Hour, 5*time.Minute).Should(BeTrue())

		// Need to add scheduler here
		baseUrl := base_url + "/v1/cloudcredits/credit"
		response_status, responseBody := financials.GetCredits(baseUrl, token, place_holder_map["cloud_account_id"])
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		fmt.Println("Response credits....", response_status, responseBody)
		usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
		usedAmount = testsetup.RoundFloat(usedAmount, 2)
		fmt.Println("USED...", usedAmount)
		Expect(usage).Should(BeNumerically(">", 0), "Failed: Validating Used Credits failed failed should be greater than zero")
		Expect(usage).To(Equal(usedAmount), "Estimated usage from Credit page is not equal to Credit from CLoud credits")
		Expect(usage).Should(BeNumerically(">", float64(zeroamt)), "Failed: Validating Usage failed failed should be greater than zero")
	})
})
