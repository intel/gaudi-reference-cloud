package STaaS_Partial_Credits_IU_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/financials/financials_utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Intel User STaaS Usage", Ordered, Label("Usages-e2e"), func() {

	// ADD CREDITS TO ACCOUNT
	It("Create Cloud credits for intel user by redeeming coupons", func() {
		logger.Logf.Info("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/cloudcredits/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 1, "idc_billing@intel.com", 1)
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
		fmt.Println("STaaS File system Vol created with resourceId: ", instance_id_created)
		Expect(create_response_code).NotTo(Equal(403), "User should not be able to create a volume with partial credits.")
	})
})
