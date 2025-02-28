package STaaS_500GB_PU_test

import (
	"fmt"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"strings"
	"time"

	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Premium User Product Usage VM Small", Ordered, Label("Usages-e2e"), func() {

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
	It("Create STaaS Instance", func() {
		fmt.Println("Starting the Instance Creation via STaaS API...")
		fmt.Print(base_url)
		create_response_code, create_response_body := financials.CreateFileSystem(compute_url, token, staas_payload, cloud_account_created)
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		fmt.Println("resourceId: ", instance_id_created)
		Expect(create_response_code).To(Equal(200), "Failed to create FileSystem")
	})

	// GET THE INSTANCE
	It("Get the created instance and validate", func() {
		fmt.Println("Cloud Account id: ", cloud_account_created)
		fmt.Println("Created instance id: ", instance_id_created)
		fmt.Println("Starting the Instance retrieval via STaaS API...")
		response_status, response_body := financials.GetFileSystemStatusByResourceId(compute_url, token, cloud_account_created, instance_id_created)
		fmt.Print("Response body: ", response_body)
		Expect(response_status).To(Equal(200), "Failed to retrieve VM instance")
		Eventually(func() bool {
			response_status, response_body := financials.GetFileSystemStatusByResourceId(compute_url, token, cloud_account_created, instance_id_created)
			Expect(response_status).To(Equal(200), "Failed to retrieve FileSystem")
			fmt.Print("Response body: ", response_body)
			if strings.Contains(response_body, `"phase":"FSReady"`) {
				return true
			}
			fmt.Println("Waiting for instance to be ready...")
			return false
		}, 1*time.Hour, 3*time.Minute).Should(BeTrue(), "Validation failed on instance retrieval, fileSystem is not in ready phase")
		place_holder_map["resource_id"] = instance_id_created
		place_holder_map["file_size"] = gjson.Get(response_body, "spec.request.storage").String()
		Expect(place_holder_map["file_size"]).To(Equal("500GB"), "Validation failed on file size retrieval")
	})

	// GET METERING RECORDS
	It("Get Metering data related to product and validate", func() {
		search_payload := fmt.Sprintf(`{
				"cloudAccountId": "%s"
		}`, place_holder_map["cloud_account_id"])
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
			metering_record_instanceType = v.Get("properties.serviceType").String()
			break
		}
		fmt.Println("metering_record_cloudAccountId", metering_record_cloudAccountId)
		fmt.Println("metering_record_resourceId", metering_record_resourceId)
		fmt.Println("metering_record_reported", metering_record_reported)
		fmt.Println("metering_record_transactionId", metering_record_transactionId)
		fmt.Println("metering_record_instanceType", metering_record_instanceType)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		Expect(place_holder_map["cloud_account_id"]).To(Equal(metering_record_cloudAccountId), "Validation failed on cloud account id retrieval")
		//Expect(place_holder_map["resource_id"]).To(Equal(metering_record_resourceId), "Validation failed on resource id retrieval")
		//Expect(meta_data_map["instanceType"]).To(Equal(metering_record_instanceType), "Validation failed on instance retrieval")
		place_holder_map["transactionId"] = metering_record_transactionId
		place_holder_map["reported"] = metering_record_reported
	})

	// GET BILLING ACCOUNT
	It("Validate billing account is created for Premium User", func() {
		response_status, responseBody := financials.GetAriaAccountDetailsAllForClientId(place_holder_map["cloud_account_id"], ariaclientId, ariaAuth)
		fmt.Println("ariaclientId", ariaclientId)
		fmt.Println("ariaAuth", ariaAuth)
		fmt.Println("responseBody", responseBody)
		fmt.Println(place_holder_map["cloud_account_id"])
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		//Expect(strings.Contains(responseBody, `"error_msg" : "account does not exist"`)).To(BeTrue(), "Validation failed fetching billing account details from aria")
	})

	// GET CREDITS AND USAGE
	It("Validate Usage calculation for the paid product and validate credits depletion", func() {
		var err error
		var usage float64
		//wait for some time to fetch some usage
		zeroamt := 0
		resourceInfo, err = testsetup.GetInstanceDetails(userName, base_url, token, compute_url)
		Expect(err).To(BeNil(), "Failed: Failed to get instance details")

		Eventually(func() bool {
			fmt.Println("RESOURCE...", resourceInfo)
			token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
			fmt.Println("TOKEN...", token)
			usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + place_holder_map["cloud_account_id"]
			usage_response_status, usage_response_body := financials.GetUsage(usage_url, token)
			Expect(usage_response_status).To(Equal(200), "Failed to retrieve Billing Account Usages")
			usage = gjson.Get(usage_response_body, "totalAmount").Float()
			fmt.Println("USAGE...", usage)

			logger.Log.Info("Usage URL" + usage_url)

			if usage > float64(zeroamt) {
				return true
			}
			fmt.Println("Waiting 40 more minutes to get usages...")
			return false
		}, 6*time.Hour, 40*time.Minute).Should(BeTrue())

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

// Get ResourceId from metering records... And create custom method to validate usages by staas resource
