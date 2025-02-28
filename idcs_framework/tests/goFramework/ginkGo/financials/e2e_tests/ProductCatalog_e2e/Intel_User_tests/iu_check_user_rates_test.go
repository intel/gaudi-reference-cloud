package ProductCatalog_e2e_test_Intel

import (
	"fmt"
	"goFramework/framework/common/logger"

	"encoding/json"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Intel User Rates", Ordered, Label("Product-Catalog-E2E-Intel"), func() {
	var cloudAccountType string
	var cloudAccIdid string

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
		response_status, response_body := financials.GetProducts(base_url, token, product_filter)
		json.Unmarshal([]byte(response_body), &structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		logger.Log.Info("Products Successfully retrieved")
	})

	It("Validating Product Rates", func() {
		for _, product := range structResponse.Products {
			if len(product.Rates) != 0 {
				Expect(product.Rates[0].AccountType).To(Equal(cloudAccountType))
				logger.Log.Info("Product: " + product.Name + " has Intel rates: " + product.Rates[0].Rate)
			} else {
				Fail(product.Name + " has no rates associated.")
			}
		}
	})
})
