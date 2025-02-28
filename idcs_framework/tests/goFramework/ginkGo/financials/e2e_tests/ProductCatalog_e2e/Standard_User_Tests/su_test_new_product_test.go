package ProductCatalog_e2e_test

import (
	"encoding/json"
	"fmt"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"math"
	"strconv"
	"strings"
	"time"

	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/library/financials/billing"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Standard User Rates", Ordered, Label("Product-Catalog-E2E-Intel"), func() {

	It("Sync PRODUCT", func() {
		ret_value, _ := billing.SyncProductPCe2e("Test01 Master Plan", "validPayload", 200, base_url)
		fmt.Print("SYNC....", ret_value)
	})

	It("Validate Standard cloudAccount", func() {
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
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, "bm-spr-test", "ubuntu-22.04-server-cloudimg-amd64-latest", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = "bm-spr-test"
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
		Expect(create_response_status).To(Equal(403), "Expected response code on paid instance for iu user should be 403 when cloud credits are zero")
		Expect(strings.Contains(create_response_body, `"message":"paid service not allowed"`)).To(BeTrue(), "Failed to validate ")
	})
	// ADD CREDITS TO ACCOUNT
	It("Create Cloud credits for Standard user by redeeming coupons", func() {
		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/billing/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayloadStandard(), 100, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		// Redeem coupon
		redeem_coupon_endpoint := base_url + "/v1/billing/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), place_holder_map["cloud_account_id"])
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Standard users should  redeem coupons")
	})

	// CREATE NEW INSTANCE
	It("Create paid test instance", func() {
		fmt.Println("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		bm_name_iu := "autobm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = "vm-spr-sml"
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
		BMName_iu := gjson.Get(create_response_body, "metadata.name").String()
		Expect(create_response_status).To(Equal(200), "Failed to create VM instance")
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
		Expect(response_status).To(Equal(200), "Failed to retrieve VM instance")
		Expect(strings.Contains(response_body, `"phase":"Ready"`)).To(BeTrue(), "Validation failed on instance retrieval")
		place_holder_map["resource_id"] = instance_id_created
		place_holder_map["machine_ip"] = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()
		fmt.Println("IP Address is :" + place_holder_map["machine_ip"])
	})

	It("Validating the request of products by Standard Account", func() {
		logger.Logf.Info("Getting all products for Intel Users")
		product_filter := fmt.Sprintf(`{
			"cloudaccountId": "%s",
			"productFilter": { "accountType": "%s", "name": "vm-spr-sml"}
		}`, cloudAccIdid, cloudAccountType)
		response_status, response_body := financials.GetProducts(base_url, token, product_filter)
		json.Unmarshal([]byte(response_body), &structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		meta_data_map["Highlight"] = structResponse.Products[0].Metadata.Highlight
		meta_data_map["disks.size"] = structResponse.Products[0].Metadata.Disks
		meta_data_map["instanceType"] = structResponse.Products[0].Metadata.InstanceType
		meta_data_map["memory.size"] = structResponse.Products[0].Metadata.Memory
		meta_data_map["region"] = structResponse.Products[0].Metadata.Region
		logger.Log.Info("Products Successfully retrieved")
	})

	It("Validate billing account is not Created for Intel User", func() {
		response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(place_holder_map["cloud_account_id"], ariaclientId, ariaAuth)
		fmt.Println("ariaclientId", ariaclientId)
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("responseBody", responseBody)
		fmt.Println(place_holder_map["cloud_account_id"])
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		//Expect(strings.Contains(responseBody, `"error_msg" : "account does not exist"`)).To(BeTrue(), "Validation failed fetching billing account details from aria")
	})

	It("Wait for the usages to record", func() {
		time.Sleep(30 * time.Minute)
	})

	It("Validate Usage calculation for the paid product and validate credits depletion", func() {
		var err error
		var usage float64
		var usedAmount float64
		//wait for some time to fetch some usage
		zeroamt := 0
		resourceInfo, err = testsetup.GetInstanceDetails(userName, base_url, token, compute_url)
		Expect(err).To(BeNil(), "Failed: Failed to get instance details")

		Eventually(func() bool {
			fmt.Println("RESOURCE...", resourceInfo)
			token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
			usage, _ = testsetup.GetUsageAndValidateTotalUsage(userName, resourceInfo, base_url, token, compute_url)
			fmt.Println("USAGE...", usage)
			if usage > float64(zeroamt) {
				return true
			}
			fmt.Println("Waiting 40 more minutes to get usages...")
			return false
		}, 7*time.Hour, 15*time.Minute).Should(BeTrue(), "Failed: Failed to get usage details")

		Eventually(func() bool {
			fmt.Println("RESOURCE...", resourceInfo)
			token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
			baseUrl := base_url + "/v1/billing/credit"
			response_status, responseBody := financials.GetCredits(baseUrl, token, place_holder_map["cloud_account_id"])
			Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
			fmt.Println("Response credits....", response_status, responseBody)
			usedAmount = gjson.Get(responseBody, "totalUsedAmount").Float()
			usedAmount = testsetup.RoundFloat(usedAmount, 2)
			fmt.Println("USED...", usedAmount)
			if usedAmount > float64(zeroamt) {
				return true
			}
			fmt.Println("Waiting 40 more minutes to get used amount...")
			return false
		}, 7*time.Hour, 15*time.Minute).Should(BeTrue(), "Failed: Failed to get usage details")

		usedAmount = math.Round(usedAmount*100) / 100
		usage = math.Round(usage*100) / 100
		Expect(usage).Should(BeNumerically(">", 0), "Failed: Validating Used Credits failed failed should be greater than zero")
		Expect(usage).To(Equal(usedAmount), "Estimated usage from Credit page is not equal to Credit from CLoud credits")
		Expect(usage).Should(BeNumerically(">", float64(zeroamt)), "Failed: Validating Usage failed failed should be greater than zero")
	})
})
