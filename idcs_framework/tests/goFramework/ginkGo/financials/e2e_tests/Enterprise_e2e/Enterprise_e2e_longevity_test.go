package Enterprise_e2e_longevity_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"math"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidwall/gjson"
)

var _ = Describe("Check Enterprise User Rates", Ordered, Label("Enterprise-E2E-Intel"), func() {

	var vnet_created string
	var ssh_publickey_name_created string
	var create_response_status int
	var create_response_body string
	var instance_id_created string

	It("Validate Enterprise cloudAccount", func() {
		auth.Get_config_file_data("../../data/config.json")
		token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		code, body = financials.GetCloudAccountById(url, token, place_holder_map["cloud_account_id"])
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).NotTo(BeNil(), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
		//code, body = financials.GetCredits(base_url+"/v1/cloudcredits/credit", token, cloudAccIdid)
		//Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Credits details")
		//cloudAccountCredits := gjson.Get(body, "totalRemainingAmount").Int()
		//Expect(cloudAccountCredits).To(Equal(int64(0)), "Failed to validate cloud account credits")
	})

	//CREATE VNET
	It("Create vnet with name", func() {
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
		time.Sleep(180 * time.Second)
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		var response_body string
		var response_status int

		Eventually(func() {
			response_status, response_body = frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
			fmt.Print("response: ", response_body)
			Expect(response_body).Should(ContainSubstring(`"phase":"Ready"`))
			Expect(response_status).Should(Equal(200))
		}, 10*time.Minute, 30*time.Second)
		place_holder_map["resource_id"] = instance_id_created
		place_holder_map["machine_ip"] = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()
	})

	It("Wait for the usages to record", func() {
		time.Sleep(30 * time.Minute)
	})

	It("Validate billing account is not Created for Enterprise User", func() {
		response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(place_holder_map["cloud_account_id"], ariaclientId, ariaAuth)
		fmt.Println("ariaclientId", ariaclientId)
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("responseBody", responseBody)
		fmt.Println(place_holder_map["cloud_account_id"])
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
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
		}, 7*time.Hour, 30*time.Minute).Should(BeTrue(), "Failed: Failed to get usage details")

		Eventually(func() bool {
			fmt.Println("RESOURCE...", resourceInfo)
			token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
			baseUrl := base_url + "/v1/cloudcredits/credit"
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
		}, 7*time.Hour, 30*time.Minute).Should(BeTrue(), "Failed: Failed to get usage details")

		usedAmount = math.Round(usedAmount*100) / 100
		usage = math.Round(usage*100) / 100
		Expect(usage).Should(BeNumerically(">", 0), "Failed: Validating Used Credits failed failed should be greater than zero")
		Expect(usage).To(Equal(usedAmount), "Estimated usage from Credit page is not equal to Credit from CLoud credits")
		Expect(usage).Should(BeNumerically(">", float64(zeroamt)), "Failed: Validating Usage failed failed should be greater than zero")
	})
})
