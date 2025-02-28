package STaaS_No_Credits_IU_test

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"goFramework/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Intel User Usages", Ordered, Label("Product-Catalog-E2E-Intel"), func() {
	productList := []string{}
	var found bool
	var structResponse GetProductsResponse

	It("Validate Intel cloudAccount", func() {
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

	It("Validating the request of products by Intel Account", func() {
		logger.Logf.Info("Getting all products for Intel Users")
		product_filter := fmt.Sprintf(`{
            "cloudaccountId": "%s",
            "productFilter": { "accountType": "%s"}
        }`, cloudAccIdid, cloudAccountType)
		response_status, response_body := financials.GetProducts(base_url, userToken, product_filter)
		json.Unmarshal([]byte(response_body), &structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		logger.Log.Info("Products Successfully retrieved")
	})

	It("Create Cloud credits for intel user by redeeming coupons", func() {
		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/cloudcredits/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 100, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		// Redeem coupon
		redeem_coupon_endpoint := base_url + "/v1/cloudcredits/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), place_holder_map["cloud_account_id"])
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Failed to redeem coupon")
	})

	It("Push Metering data for all open products", func() {
		for _, product := range structResponse.Products {
			prodName := product.Name
			instanceName := utils.GenerateString(10)
			productList = append(productList, prodName)
			fmt.Println("Products", productList)
			now := time.Now().UTC()
			previousDate := now.Add(3 * time.Hour).Format("2006-01-02T15:04:05.999999Z")
			fmt.Println("Metering Date", previousDate)
			create_payload := financials_utils.EnrichMeteringCreatePayload(compute_utils.GetMeteringCreatePayload(),
				uuid.NewString(), uuid.NewString(), place_holder_map["cloud_account_id"], previousDate, prodName, instanceName, "10000")
			fmt.Println("create_payload", create_payload)
			metering_api_base_url := base_url + "/v1/meteringrecords"
			fmt.Println("metering create paylaod", metering_api_base_url)
			response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, token, create_payload)
			Expect(response_status).To(Equal(200), "Failed to create Metering Records")
		}
	})

	It("Wait for usages to show up", func() {
		time.Sleep(20 * time.Minute)

	})

	It("Validate Usages showing up for all products", func() {
		usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + place_holder_map["cloud_account_id"]
		usage_response_status, usage_response_body := financials.GetUsage(usage_url, token)
		Expect(usage_response_status).To(Equal(200), "Failed to validate usage_response_status")
		logger.Logf.Info("usage_response_body: %s ", usage_response_body)
		result := gjson.Parse(usage_response_body)
		arr := gjson.Get(result.String(), "usages")
		for _, s := range productList {
			found = false
			arr.ForEach(func(key, value gjson.Result) bool {
				data := value.String()
				logger.Logf.Infof("Usage Data : %s", data)
				if gjson.Get(data, "productType").String() == s {
					found = true
					Amount := gjson.Get(data, "amount").String()
					actualAMount, _ := strconv.ParseFloat(Amount, 64)
					Expect(actualAMount).Should(BeNumerically(">", float64(0)), "Failed to get positive usage")
				}
				return true // keep iterating
			})

			Expect(found).To(Equal(true), "Usage not found for the product", s)

		}

	})

	It("Validate Credits", func() {
		baseUrl := base_url + "/v1/cloudcredits/credit"
		response_status, responseBody := financials.GetCredits(baseUrl, token, place_holder_map["cloud_account_id"])
		Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Cloud Credits")
		usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
		usedAmount = testsetup.RoundFloat(usedAmount, 0)
		Expect(usedAmount).Should(BeNumerically(">", float64(0)), "Failed to validate used credits")

	})
})
