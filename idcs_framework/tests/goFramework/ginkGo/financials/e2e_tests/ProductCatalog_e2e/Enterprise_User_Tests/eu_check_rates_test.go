package ProductCatalog_e2e_test

import (
	"encoding/json"
	"fmt"

	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Enterprise User Rates", Ordered, Label("Product-Catalog-E2E"), func() {
	var structResponse GetProductsResponse

	It("Validate Enterprise cloudAccount", func() {
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		code, body = financials.GetCloudAccountById(url, token, place_holder_map["cloud_account_id"])
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).NotTo(BeNil(), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
	})

	It("Validating the request of products by Enterprise Account", func() {
		product_filter := fmt.Sprintf(`{
			"cloudaccountId": "%s",
			"productFilter": { "accountType": "%s"}
		}`, cloudAccIdid, cloudAccountType)
		response_status, response_body := financials.GetProducts(base_url, token, product_filter)
		json.Unmarshal([]byte(response_body), &structResponse)
		fmt.Print(structResponse.Products)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
	})

	It("Validating Product Rates", func() {
		for _, product := range structResponse.Products {
			if len(product.Rates) != 0 {
				Expect(product.Rates[0].AccountType).To(Equal(cloudAccountType))
			} else {
				fmt.Println(product.Name + " has no rates associated.")
			}
		}
	})
})
