package STaaS_Low_Credits_EU_Check_test

import (
	"fmt"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"

	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Enterprise User STaaS Usage", Ordered, Label("Usages-e2e"), func() {

	// ADD CREDITS TO ACCOUNT
	It("Create Cloud credits for Standard user by redeeming coupons", func() {
		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/cloudcredits/coupons"
		//small VM = $0.45 per hour
		//$31.95 = 71 hours
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayloadStandard(), 31.95, "eu_user", 1)
		fmt.Println("Payload", coupon_payload)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		// Redeem coupon
		redeem_coupon_endpoint := base_url + "/v1/cloudcredits/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), place_holder_map["cloud_account_id"])
		fmt.Println("Payload", redeem_payload)
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Standard users should  redeem coupons")
	})

	// CREATE NEW INSTANCE
	It("Create STaaS Instance", func() {
		fmt.Println("Starting the Instance Creation via STaaS API...")
		fmt.Print(base_url)
		create_response_code, create_response_body := financials.CreateFileSystem(compute_url, token, staas_payload, cloud_account_created)
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		fmt.Println("resourceId: ", instance_id_created)
		Expect(create_response_code).NotTo(Equal(200), "User should  be able to create a volume with credits.")
	})

	It("Validate Usage calculation for the paid product and validate credits depletion", func() {
		var err error
		zeroamt := 0
		resourceInfo, err = testsetup.GetInstanceDetails(userName, base_url, token, compute_url)
		Expect(err).To(BeNil(), "Failed: Failed to get instance details")

		fmt.Println("RESOURCE...", resourceInfo)
		usage, _ := testsetup.GetUsageAndValidateTotalUsage(userName, resourceInfo, base_url, token, compute_url)
		fmt.Println("USAGE...", usage)

		token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
		baseUrl := base_url + "/v1/cloudcredits/credit"
		response_status, responseBody := financials.GetCredits(baseUrl, token, place_holder_map["cloud_account_id"])
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Details from Aria")
		fmt.Println("Response credits....", response_status, responseBody)
		usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
		usedAmount = testsetup.RoundFloat(usedAmount, 2)
		fmt.Println("USED...", usedAmount)
		Expect(usage).Should(BeNumerically("=", 0), "Failed: Validating Used Credits failed failed should be equals to zero")
		Expect(usage).To(Equal(usedAmount), "Estimated usage from Credit page is not equal to Credit from Cloud credits")
		Expect(usage).Should(BeNumerically("=", float64(zeroamt)), "Failed: Validating Usage failed failed should be equals to zero")
	})

})
