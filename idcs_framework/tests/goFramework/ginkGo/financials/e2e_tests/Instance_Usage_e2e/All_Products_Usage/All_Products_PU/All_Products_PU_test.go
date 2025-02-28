package All_Products_PU_test

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slices"
)

var _ = Describe("Check Premium User Usages", Label("Vacuum-tests-Premium"), func() {
	productList := []string{}
	var structResponse GetProductsResponse

	It("Validate Premium cloudAccount using getByName", func() {
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

	It("Validate Premium cloudAccount using getById", func() {
		logger.Logf.Info("Checking user cloudAccount")
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountById(url, token, place_holder_map["cloud_account_id"])
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).NotTo(BeNil(), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
		logger.Log.Info("CloudAccount Validated")
	})

	It("Whitelist MaaS products", func() {
		logger.Logf.Info("Checking user cloudAccount")
		admin_name := "idc_billing@intel.com"
		code, _ := financials.Whitelist_Cloud_Account_MaaS(base_url, token, cloudAccIdid, admin_name)
		Expect(code).To(Equal(200), "Failed to Whitelist MaaS Products")
		logger.Log.Info("MaaS Products Whitelisted succesfully.")
	})

	It("Whitelist IKS products", func() {
		logger.Logf.Info("Checking user cloudAccount")
		admin_name := "idc_billing@intel.com"
		code, _ := financials.Whitelist_Cloud_Account_IKS(base_url, token, cloudAccIdid, admin_name)
		Expect(code).To(Equal(200), "Failed to Whitelist IKS products.")
		logger.Log.Info("IKS Products Whitelisted succesfully.")
	})

	It("Whitelist BM products", func() {
		logger.Logf.Info("Checking user cloudAccount")
		admin_name := "idc_billing@intel.com"
		code, _ := financials.Whitelist_Cloud_Account_Gaudi3(base_url, token, cloudAccIdid, admin_name)
		Expect(code).To(Equal(200), "Failed to Whitelist Gaudi3 GNR products.")
		code, _ = financials.Whitelist_Cloud_Account_Gaudi3_8node(base_url, token, cloudAccIdid, admin_name)
		Expect(code).To(Equal(200), "Failed to Whitelist Gaudi3 8 Node products.")
		code, _ = financials.Whitelist_Cloud_Account_Gaudi2_32_node(base_url, token, cloudAccIdid, admin_name)
		Expect(code).To(Equal(200), "Failed to Whitelist Gaudi2 32 Node products.")
		code, _ = financials.Whitelist_Cloud_Account_Gaudi2_single_node(base_url, token, cloudAccIdid, admin_name)
		Expect(code).To(Equal(200), "Failed to Whitelist Gaudi2 Single Node products.")
		logger.Log.Info("Gaudi3 Products Whitelisted succesfully.")
	})

	It("Validating the request of products by Premium Account", func() {
		logger.Logf.Info("Getting all products for Premium Users")
		product_filter := fmt.Sprintf(`{
            "cloudaccountId": "%s",
			"productFilter": {
			}
        }`, cloudAccIdid)
		response_status, response_body := financials.GetProducts(base_url, token, product_filter)
		err := json.Unmarshal([]byte(response_body), &structResponse)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		logger.Log.Info("Products Successfully retrieved")
	})

	It("Validate products with billingEnable", func() {
		for _, product := range structResponse.Products {
			fmt.Println("Product Name:", product.Name)
			fmt.Println("BillingEnable:", product.Metadata.BillingEnable)
			if product.Metadata.BillingEnable == "false" {
				fmt.Println("Skipping product due to BillingEnable being false")
				continue
			}
			fmt.Println("Processing product:", product.Name)
		}
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

	It("Push Metering data for all open products", func() {
		var wg sync.WaitGroup
		maxGoroutines := 10 // Set the maximum number of concurrent goroutines
		guard := make(chan struct{}, maxGoroutines)

		for _, product := range structResponse.Products {
			prodName := product.Name
			if product.Metadata.BillingEnable == "false" {
				fmt.Println("Product: ", product.Name)
				fmt.Println("BillingEnable: ", product.Metadata.BillingEnable)
				fmt.Println("Skipping product due to BillingEnable being false")
				continue
			}
			productList = append(productList, prodName)
			fmt.Println("Products", productList)
			serviceTypes := financials_utils.GetServiceType(product.MatchExpr)
			instanceGroupSize := financials_utils.GetInstanceGroupSize(product.MatchExpr)

			// Set default groupSize to 1 if instanceGroupSize is empty
			var groupSize int
			if instanceGroupSize == "" {
				groupSize = 1
			} else {
				var err error
				groupSize, err = strconv.Atoi(instanceGroupSize)
				if err != nil {
					fmt.Println("Invalid groupSize, defaulting to 1")
					groupSize = 1
				}
			}

			if len(serviceTypes) >= 1 {
				for _, serviceType := range serviceTypes {
					wg.Add(1)
					guard <- struct{}{} // Acquire a slot
					go func(product Product, serviceType string, groupSize int) {
						defer GinkgoRecover()
						defer wg.Done()
						defer func() { <-guard }() // Release the slot

						fmt.Println("Service Type: ", serviceType)
						metering_api_base_url := base_url + "/v1/meteringrecords"
						current_time := time.Now().Add(-1 * time.Hour)
						firstReadyTimeStamp := current_time.Format(time.RFC3339Nano)
						fmt.Println("Creation time to be set: ", firstReadyTimeStamp)
						fmt.Println("groupSize", groupSize)
						for i := 1; i <= groupSize; i++ {
							fmt.Println("groupSize", groupSize)
							fmt.Println("iteration", i)
							create_payload := financials_utils.GenerateMetringPayload(serviceType, product.Metadata.InstanceType, cloudAccIdid, instanceGroupSize, firstReadyTimeStamp)
							str := financials_utils.ConvertStructToJson(create_payload)
							fmt.Println("metering create payload", str)
							if serviceType == "ModelAsAService" {
								usage_api_base_url := base_url + "/v1/usages/records/products"
								response_status, _ := financials.CreateUsageRecords(usage_api_base_url, token, str)
								if response_status == 401 {
									token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
									response_status, _ = financials.CreateUsageRecords(usage_api_base_url, token, str)
								}
								if response_status != 200 {
									fmt.Printf("Failed to create Usage Records for product %s, service type %s\n", product.Name, serviceType)
								}
							} else {
								response_status, _ := financials.CreateMeteringRecords(metering_api_base_url, token, str)
								if response_status == 401 {
									token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
									response_status, _ = financials.CreateMeteringRecords(metering_api_base_url, token, str)
								}
								if response_status != 200 {
									fmt.Printf("Failed to create Metering Records for product %s, service type %s\n", product.Name, serviceType)
								}
							}
						}
					}(product, serviceType, groupSize)
				}
			}
		}

		wg.Wait()
	})
	It("Wait for usages to show up", func() {
		time.Sleep(40 * time.Minute)
	})

	It("Validate Usages showing up for all products", func() {
		usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + place_holder_map["cloud_account_id"]
		var usage_response_status int
		var usage_response_body string
		arr := gjson.Result{}
		var notFoundProducts []string

		Eventually(func() bool {
			token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
			fmt.Println("TOKEN...", token)
			usage_response_status, usage_response_body = financials.GetUsage(usage_url, token)
			fmt.Println("usage_response_body", usage_response_body)
			fmt.Println("usage_response_status", usage_response_status)
			Expect(usage_response_status).To(Equal(200), "Failed to validate usage_response_status")
			logger.Logf.Info("usage_response_body: %s ", usage_response_body)
			result := gjson.Parse(usage_response_body)
			arr = gjson.Get(result.String(), "usages")
			fmt.Println("products usages", arr)
			fmt.Println("PROD ARRR", arr.Array())
			if len(arr.Array()) > 0 {
				return true
			}
			fmt.Println("Waiting 40 more minutes to get products usages...")
			return false
		}, 8*time.Hour, 15*time.Minute).Should(BeTrue())

		for _, s := range productList {
			// Ignore products starting with "tr-" or "pre-"
			if strings.HasPrefix(s, "tr-") || strings.HasPrefix(s, "pre-") {
				continue
			}

			found := false
			fmt.Println("Product Name", s)
			arr.ForEach(func(key, value gjson.Result) bool {
				data := value.String()
				logger.Logf.Infof("Usage Data : %s", data)
				product := gjson.Get(data, "productType").String()
				logger.Logf.Infof("Product Data : %s", product)
				if strings.Contains(product, s) {
					found = true
					sAmount := gjson.Get(data, "amount").String()
					sRate := gjson.Get(data, "rate").String()
					sMinUsed := gjson.Get(data, "minsUsed").String()
					actualAmount, _ := strconv.ParseFloat(sAmount, 64)
					rate, _ := strconv.ParseFloat(sRate, 64)
					minUsed, _ := strconv.ParseFloat(sMinUsed, 64)
					expectedAmount := minUsed * rate

					// Define a tolerance for floating-point comparison
					tolerance := 0.0001

					Expect(actualAmount).Should(BeNumerically(">", float64(0)), "Failed to get positive usage")
					Expect(math.Abs(actualAmount-expectedAmount)).Should(BeNumerically("<=", tolerance), "Actual amount and expected amount are not within the tolerance range")
				}
				return true // keep iterating
			})
			if !found {
				notFoundProducts = append(notFoundProducts, s)
			}
		}
		if len(notFoundProducts) > 0 {
			fmt.Printf("Usages not found for the following products for %s: \n", place_holder_map["cloud_account_type"])
			for _, product := range notFoundProducts {
				fmt.Println(product)
			}
			fmt.Printf("Total products with no usage found for %s: %d\n", place_holder_map["cloud_account_type"], len(notFoundProducts))
		}
		Expect(len(notFoundProducts)).To(Equal(0), "Some products did not have usage data")
	})

	It("Validate Credits", func() {
		Eventually(func() bool {
			fmt.Println("RESOURCE...", resourceInfo)
			token_response, _ := auth.Get_Azure_Bearer_Token(userName)
			userToken = "Bearer " + token_response
			fmt.Println("TOKEN...", userToken)
			baseUrl := base_url + "/v1/cloudcredits/credit"
			response_status, responseBody := financials.GetCredits(baseUrl, userToken, place_holder_map["cloud_account_id"])
			Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Cloud Credits")
			remainingAmount := gjson.Get(responseBody, "totalRemainingAmount").Float()
			remainingAmount = testsetup.RoundFloat(remainingAmount, 0)
			Expect(remainingAmount).NotTo(BeZero(), "Remaining amount cannot be zero from start.")
			usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
			usedAmount = testsetup.RoundFloat(usedAmount, 0)
			if usedAmount > float64(float64(0)) {
				return true
			}
			fmt.Println("Waiting 15 more minutes to get credit depletion...")
			return false
		}, 2*time.Hour, 15*time.Minute).Should(BeTrue())
	})
})

var _ = Describe("Check Product Catalog Validations", Ordered, Label("Vacuum-tests-Premium"), func() {
	var structResponse GetProductsResponse
	It("Validating the request of products Admin", func() {
		logger.Logf.Info("Getting all products for Standard Users")
		product_filter := fmt.Sprintf(`{
        }`)
		response_status, response_body := financials.GetProductsAdmin(base_url, token, product_filter)
		json.Unmarshal([]byte(response_body), &structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		logger.Log.Info("Products Successfully retrieved")
	})

	for _, product := range structResponse.Products {
		serviceTypes := financials_utils.GetServiceType(product.MatchExpr)
		instanceGroupSize := financials_utils.GetInstanceGroupSize(product.MatchExpr)
		groupSize, _ := strconv.Atoi(instanceGroupSize)
		fails := InterceptGomegaFailures(func() {
			if slices.Contains(serviceTypes, "ComputeAsAService") || slices.Contains(serviceTypes, "KubernetesAsAServices") {
				Expect(groupSize).To(BeNumerically(">", 0), "Product: "+product.Name+" has Instance Group Size: "+instanceGroupSize)
			}
			Expect(len(product.Description)).To(BeNumerically("<=", 100), "Product: "+product.Name+" description exceeds 100 characters")
			Expect(len(product.Name)).To(BeNumerically("<=", 100), "Product "+product.Name+" name exceeds 100 characters")
		})
		failures = append(failures, fails...)
	}

	It("Check Product Catalog Validations failures", func() {
		if len(failures) > 0 {
			for _, failure := range failures {
				fmt.Println(failure)
			}
			gomega.Expect(len(failures)).To(gomega.BeZero(), "There were test failures")
		}
	})
})
