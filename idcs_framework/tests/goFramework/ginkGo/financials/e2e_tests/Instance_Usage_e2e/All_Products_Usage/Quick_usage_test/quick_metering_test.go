package Quick_usage_test_test

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/financials/financials_utils"
	"strconv"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuickMetering", func() {
	productList := []string{}
	var structResponse GetProductsResponse

	It("Validating the request of products by Standard Account", func() {
		logger.Logf.Info("Getting all products for Standard Users")
		cloudaccountId := ""
		productFilter := fmt.Sprintf(`{
			"cloudaccountId": "%s",
			"productFilter": {
				"familyId": "fabe738e-1edd-4d07-b1b4-5d9eadc9f28d"
			}
		}`, cloudaccountId)
		response_status, response_body := financials.GetProducts(base_url, token, productFilter)
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

		fmt.Println("Admin token:", token)
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
							create_payload := financials_utils.GenerateMetringPayload(serviceType, product.Metadata.InstanceType, "", instanceGroupSize, firstReadyTimeStamp)
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
})
