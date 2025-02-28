package All_Products_PU_per_product_test

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/financials/financials_utils"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var (
	tokenRefreshInterval = 1 * time.Hour
	usageCheckInterval   = 15 * time.Minute
	usageCheckTimeout    = 8 * time.Hour
	creditCheckTimeout   = 8 * time.Hour
	creditCheckInterval  = 15 * time.Minute
	sleepBetweenMetering = 1 * time.Second
	sleepBetweenBatches  = 5 * time.Second
	batchSize            = 50
	initialUsageWait     = 20 * time.Minute
	additionalUsageWait  = 40 * time.Minute
	additionalCreditWait = 40 * time.Minute
	structResponse       GetProductsResponse
)

var _ = Describe("Check Premium User Usages", Ordered, Label("Vacuum-tests-Premium"), func() {
	It("Validate Premium cloudAccount", func() {
		logger.Logf.Info("Checking user cloudAccount")
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).NotTo(BeNil(), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
		logger.Log.Info("CloudAccount Validated")
	})

	It("Validating the request of products by Premium Account", func() {
		logger.Logf.Info("Getting all products for Premium Users")
		product_filter := fmt.Sprintf(`{
        }`)
		response_status, response_body := financials.GetProductsAdmin(base_url, token, product_filter)
		json.Unmarshal([]byte(response_body), &structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		logger.Log.Info("Products Successfully retrieved")
	})

	It("Create Cloud credits for Premium user by redeeming coupons", func() {
		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/cloudcredits/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 100, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		// Redeem coupon
		redeem_coupon_endpoint := base_url + "/v1/cloudcredits/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), place_holder_map["cloud_account_id"])
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, userToken, redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Failed to redeem coupon")
	})

	It("Validate Usage and credit depletion per product", func() {
		fmt.Println("PRODS...", structResponse.Products)
		var meteringAPIBaseURL = base_url + "/v1/meteringrecords"
		for _, product := range structResponse.Products {
			prodName := product.Name
			serviceTypes := financials_utils.GetServiceType(product.MatchExpr)
			instanceGroupSize := financials_utils.GetInstanceGroupSize(product.MatchExpr)
			groupSize, err := strconv.Atoi(instanceGroupSize)
			Expect(err).NotTo(HaveOccurred(), "Error transforming groupSize: %s", err)

			fmt.Printf("Product: %s\n", prodName)
			product := product // capture range variable

			fmt.Printf("Push Metering data for product: %s", prodName)
			if len(serviceTypes) >= 1 {
				for _, serviceType := range serviceTypes {
					time.Sleep(sleepBetweenMetering)
					currentTime := time.Now().Add(-1 * time.Hour)
					firstReadyTimeStamp := currentTime.Format(time.RFC3339Nano)

					for i := 1; i <= groupSize; i++ {
						createPayload := financials_utils.GenerateMetringPayload(serviceType, product.Metadata.InstanceType, place_holder_map["cloud_account_id"], instanceGroupSize, firstReadyTimeStamp)
						payloadStr := financials_utils.ConvertStructToJson(createPayload)
						responseStatus, body := financials.CreateMeteringRecords(meteringAPIBaseURL, token, payloadStr)
						if responseStatus == 403 || responseStatus == 401 {
							token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
							responseStatus, body = financials.CreateMeteringRecords(meteringAPIBaseURL, token, payloadStr)
						}
						Expect(responseStatus).To(Equal(200), "Failed to create Metering Records: %s", body)

						if i%batchSize == 0 {
							time.Sleep(sleepBetweenBatches)
						}
					}
				}
			}

			fmt.Printf("Validate Usages showing up for product: %s", prodName)
			var usageURLTemplate = base_url + "/v1/billing/usages?cloudAccountId=%s"
			usageURL := fmt.Sprintf(usageURLTemplate, place_holder_map["cloud_account_id"])
			var usageResponseStatus int
			var usageResponseBody string
			var arr gjson.Result

			Eventually(func() bool {
				token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
				usageResponseStatus, usageResponseBody = financials.GetUsage(usageURL, token)
				Expect(usageResponseStatus).To(Equal(200), "Failed to validate usage_response_status")
				arr = gjson.Get(gjson.Parse(usageResponseBody).String(), "usages")
				return len(arr.Array()) > 0
			}, usageCheckTimeout, usageCheckInterval).Should(BeTrue())

			found := false
			arr.ForEach(func(key, value gjson.Result) bool {
				data := value.String()
				fmt.Println("Data: ", data)
				productType := gjson.Get(data, "productType").String()
				if strings.Contains(productType, "tr-") || strings.Contains(productType, "pre-") {
					found = true
				} else {
					found = true
					sAmount := gjson.Get(data, "amount").String()
					actualAmount, _ := strconv.ParseFloat(sAmount, 64)
					Expect(actualAmount).Should(BeNumerically(">", float64(0)), "Failed to get positive usage")
				}
				return true
			})
			Expect(found).To(BeTrue(), "Usage not found for the product: %s", prodName)

			fmt.Printf("Validate Credits for product: %s", prodName)
			Eventually(func() bool {
				token_response, _ := auth.Get_Azure_Bearer_Token(userName)
				userToken = "Bearer " + token_response
				fmt.Println("TOKEN...", userToken)
				var creditURL = base_url + "/v1/cloudcredits/credit"
				responseStatus, responseBody := financials.GetCredits(creditURL, userToken, place_holder_map["cloud_account_id"])
				Expect(responseStatus).To(Equal(200), "Failed to retrieve Billing Account Cloud Credits")
				usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
				return usedAmount > 0
			}, creditCheckTimeout, creditCheckInterval).Should(BeTrue())
		}
	})
})
