//go:build pu_bill_01 || All || VM || pu_su_iu || bhawna
// +build pu_bill_01 All VM pu_su_iu bhawna

package vmaas

import (
	//"bytes"
	"fmt"
	//"os/exec"
	// "encoding/json"

	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"

	// "goFramework/testsetup"
	"strconv"
	"strings"

	"goFramework/utils"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

// var base_url string
// var token string

var _ = Describe("VM Instance", Ordered, Label("Financial-Invoice-E2E"), func() {
	var base_url string
	var compute_url string
	var userName string
	//var token_url string
	var token string
	var userToken string
	var cloud_account_created string
	var vnet_created string
	var ssh_publickey_name_created string
	var instance_id_created string
	var ariaAuth string
	var create_response_status int
	var create_response_body string
	var place_holder_map = make(map[string]string)
	var meta_data_map = make(map[string]string)
	var ariaclientId string

	BeforeAll(func() {
		compute_utils.LoadE2EConfig("../../../financials/data", "vmaas_input.json")
		financials_utils.LoadE2EConfig("../../../financials/data", "billing.json")
		auth.Get_config_file_data("../../../financials/data/config.json")
		base_url = compute_utils.GetBaseUrl()
		compute_url = compute_utils.GetComputeUrl()
		userName := auth.Get_UserName("Premium")
		token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
		userToken, _ = auth.Get_Azure_Bearer_Token(userName)
		userToken = "Bearer " + userToken
		ariaclientId = financials_utils.GetAriaClientNo()
		ariaAuth = financials_utils.GetariaAuthKey()
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("ariaclientId", ariaclientId)
	})

	// It("Delete cloud account", func() {
	// 	fmt.Println("Delete cloud account")
	// 	url := base_url + "/v1/cloudaccounts"
	// 	cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, token)
	// 	if err != nil {
	// 		financials.DeleteCloudAccountById(url, token, cloudAccId)
	// 	}

	// })

	It("Create cloud account", func() {
		fmt.Println("create cloudaccount")
		token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
		userToken, _ = auth.Get_Azure_Bearer_Token(userName)
		fmt.Println("usertoken", userToken)
		userToken = "Bearer " + userToken
		cloudaccount_enroll_url := base_url + "/v1/cloudaccounts/enroll"
		enroll_payload := `{"premium":true}`
		// cloud account creation
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, true)
		fmt.Println("cloudaccount_creation_body", cloudaccount_creation_body)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of premium user")
		Expect(strings.Contains(cloudaccount_creation_body, `"action":"ENROLL_ACTION_COUPON_OR_CREDIT_CARD"`)).To(BeTrue(), "Validation failed on Enrollment of premium user")

		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/billing/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 100, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		// Redeem coupon
		redeem_coupon_endpoint := base_url + "/v1/billing/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), cloud_account_created)
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Failed to redeem coupon")

		cloudaccount_creation_status, cloudaccount_creation_body = financials.CreateCloudAccount(cloudaccount_enroll_url, userToken, enroll_payload)
		fmt.Println("cloudaccount_creation_body", cloudaccount_creation_body)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type = gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_PREMIUM"`)).To(BeTrue(), "Validation failed on Enrollment of premium user")
		// Expect(strings.Contains(cloudaccount_creation_body, `"action":"ENROLL_ACTION_NONE"`)).To(BeTrue(), "Validation failed on Enrollment of premium user")
		place_holder_map["cloud_account_id"] = cloud_account_created
		place_holder_map["cloud_account_type"] = cloudaccount_type
		place_holder_map["cloud_account_id"] = cloud_account_created
		fmt.Println("cloudAccount_id", cloud_account_created)
	})

	It("Create vnet with name", func() {
		fmt.Println("Starting the VNet Creation via API...")
		// // form the endpoint and payload
		vnet_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
		vnet_name := "us-dev-1a-default"
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

	It("Create vm instance", func() {
		fmt.Println("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		vm_name_iu := "autovm-" + utils.GenerateSSHKeyName(4)
		vnet_created := "us-dev-1a-default"
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), vm_name_iu, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = "vm-spr-sml"
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, token, instance_payload)
		VMName_iu := gjson.Get(create_response_body, "metadata.name").String()
		Expect(create_response_status).To(Equal(200), "Failed to create VM instance")
		Expect(vm_name_iu).To(Equal(VMName_iu), "Failed to create VM instance, resposne validation failed")
		place_holder_map["instance_creation_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
		place_holder_map["paid_instance_name"] = vm_name_iu
	})

	It("Get the created instance and validate", func() {
		token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
		userToken, _ = auth.Get_Azure_Bearer_Token(userName)
		userToken = "Bearer " + userToken
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		// Adding a sleep because it seems to take some time to reflect the creation status
		time.Sleep(5 * time.Minute)
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		response_status, response_body := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
		Expect(response_status).To(Equal(200), "Failed to retrieve VM instance")
		Expect(strings.Contains(response_body, `"phase":"Ready"`)).To(BeTrue(), "Validation failed on instance retrieval")
		place_holder_map["resource_id"] = instance_id_created
		place_holder_map["machine_ip"] = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()

		fmt.Println("IP Address is :" + place_holder_map["machine_ip"])
	})

	// post metering usages for previous month
	It("Create metering usages for two days ago and validate", func() {
		now := time.Now().UTC()
		previousDate := now.AddDate(0, 0, -35).Format("2006-01-02T15:04:05.999999Z")
		fmt.Println("previousDate", previousDate)
		create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
			uuid.NewString(), place_holder_map["resource_id"], place_holder_map["cloud_account_id"], previousDate,
			place_holder_map["instance_type"], place_holder_map["paid_instance_name"], "1000")
		fmt.Println("create_payload", create_payload)
		metering_api_base_url := base_url + "/v1/meteringrecords"
		response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, token, create_payload)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
	})
	// Add scheduler call, or break test so thath usages can be reported.

	It("Validate Usage calculation for the paid product and validate credits depletion", func() {
		resourceInfo, _ := testsetup.GetInstanceDetails(userName, base_url, token, compute_url)
		//Expect(err).To(Equal(err), "Failed: Failed to get instance details")
		//wait for some time to fetch some usage
		time.Sleep(6 * time.Minute)
		usage1, _ := testsetup.GetUsageAndValidateTotalUsage(userName, resourceInfo, base_url, token, compute_url)
		fmt.Println("Usage of free instance ", usage1)
		// Need to add scheduler here
		baseUrl := base_url + "/v1/billing/credit"
		response_status, responseBody := financials.GetCredits(baseUrl, token, place_holder_map["cloud_account_id"])
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
		Expect(usage1).To(Equal(usedAmount), "Billing Reported Non Zero Credit for Free Product Usage")

	})

	// get pending invoice number and regenrate invoice
	It("Get pending invoice number and discard pending invoice", func() {
		response_status, response_body := financials.GetAriaPendingInvoiceNumberForClientId(place_holder_map["cloud_account_id"], ariaclientId, ariaAuth)
		Expect(response_status).To(Equal(200), "Failed to retrieve pending invoice number")
		json := gjson.Parse(response_body)
		pendingInvoice := json.Get("pending_invoice")
		var directive int64 = 2
		pendingInvoice.ForEach(func(_, value gjson.Result) bool {
			invoiceNo := value.Get("invoice_no").String()
			fmt.Println("Discarding pending Invoice No:", invoiceNo)
			response_status, _ = financials.ManageAriaPendingInvoiceForClientId(place_holder_map["cloud_account_id"], invoiceNo, ariaclientId, ariaAuth, directive)
			Expect(response_status).To(Equal(200), "Failed to discard pending invoice number")
			return true
		})
	})

	It("Generate Invoice at Account level", func() {
		// client_acct_id := "idc." + cloud_account_created
		response_status, response_body := financials.GenerateAriaInvoiceForClientId(place_holder_map["cloud_account_id"], ariaclientId, ariaAuth)
		Expect(response_status).To(Equal(200), "Failed to Generate Invoice")
		json := gjson.Parse(response_body)
		pendingInvoice := json.Get("out_invoices")
		var directive int64 = 1
		pendingInvoice.ForEach(func(_, value gjson.Result) bool {
			invoiceNo := value.Get("invoice_no").String()
			fmt.Println("Approving pending Invoice No:", invoiceNo)
			response_status, _ = financials.ManageAriaPendingInvoiceForClientId(place_holder_map["cloud_account_id"], invoiceNo, ariaclientId, ariaAuth, directive)
			Expect(response_status).To(Equal(200), "Failed to Approving pending Invoice")
			return true
		})
	})

	// get billing invoice
	It("Get billing invoice for clientId", func() {
		fmt.Println("Get billing invoice for clientId")
		url := base_url + "/v1/billing/invoices"
		respCode, invoices := financials.GetInvoice(url, token, place_holder_map["cloud_account_id"])
		Expect(respCode).To(Equal(200), "Failed to get billing invoice for clientId")
		fmt.Println(invoices)

		jsonInvoices := gjson.Parse(invoices).Get("invoices")

		jsonInvoices.ForEach(func(_, value gjson.Result) bool {
			invoiceNo := value.Get("id").String()
			fmt.Println("invoiceNo", invoiceNo)
			// Bug is open for download link
			// downloadLink := value.Get("downloadLink").String()
			//Expect(downloadLink).NotTo(BeNil(), "Invoice download link unavailable nil.")

			//invoice details
			url := base_url + "/v1/billing/invoices/detail"
			//TOdo invoiceNo
			respCode, detail := financials.GetInvoicewithInvoiceId(url, token, place_holder_map["cloud_account_id"], invoiceNo)
			Expect(respCode).To(Equal(200), "Failed to get billing invoice details for clientId")
			fmt.Println(detail) // Empty Response

			// invoices statement
			url = base_url + "/v1/billing/invoices/statement"
			//TOdo invoiceNo
			respCode, statement := financials.GetInvoicewithInvoiceId(url, token, place_holder_map["cloud_account_id"], invoiceNo)
			Expect(respCode).To(Equal(200), "Failed to get billing invoice statement for clientId")
			fmt.Println(statement)
			return true
		})

		//invoices unbilled
		url = base_url + "/v1/billing/invoices/unbilled"
		respCode, _ = financials.GetInvoice(url, token, place_holder_map["cloud_account_id"])
		Expect(respCode).To(Equal(200), "Failed to get billing invoice for clientId")
	})

	// It("Delete the SSH key created", func() {
	// 	fmt.Println("Delete the SSH-Public-Key Created above...")
	// 	ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"

	// 	delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_name_created)
	// 	Expect(delete_response_byname_status).To(Equal(200), "assert ssh-public-key deletion response code")
	// })

	// Deletion of vnet is not working - Bug is open for the same
	/*It("Delete the Vnet created", func() {
		fmt.Println("Delete the Vnet Created above...")
		vnet_endpoint := base_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
		// Deletion of vnet via name
		delete_response_byname_status, _ := frisby.DeleteVnetByName(vnet_endpoint, token, vnet_created)
		Expect(delete_response_byname_status).To(Equal(200), "assert vnet deletion response code")
	})*/

	// It("Delete cloud account", func() {
	// 	fmt.Println("Delete cloud account")
	// 	url := base_url + "/v1/cloudaccounts"
	// 	delete_Cacc, _ := financials.DeleteCloudAccountById(url, token, place_holder_map["cloud_account_id"])
	// 	Expect(delete_Cacc).To(Equal(200), "Failed to delete Cloud Account")
	// })

	// It("Validating deletion of cloud account by retrieving cloud account details", func() {
	// 	fmt.Println("Validating deletion of cloud account by retrieving cloud account details")
	// 	url := base_url + "/v1/cloudaccounts"
	// 	resp_body, _ := financials.GetCloudAccountById(url, token, place_holder_map["cloud_account_id"])
	// 	fmt.Println("resp_body", resp_body)
	// })

})
