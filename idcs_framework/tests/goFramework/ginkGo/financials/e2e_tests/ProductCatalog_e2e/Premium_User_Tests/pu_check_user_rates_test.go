package ProductCatalog_e2e_test

import (
	"fmt"
	"time"

	"encoding/json"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

type GetProductsResponse1 struct {
	Products []struct {
		Name        string    `json:"name"`
		ID          string    `json:"id"`
		Created     time.Time `json:"created"`
		VendorID    string    `json:"vendorId"`
		FamilyID    string    `json:"familyId"`
		Description string    `json:"description"`
		Metadata    struct {
			Category     string `json:"category"`
			Disks        string `json:"disks.size"`
			DisplayName  string `json:"displayName"`
			Desc         string `json:"family.displayDescription"`
			DispName     string `json:"family.displayName"`
			Highlight    string `json:"highlight"`
			Information  string `json:"information"`
			InstanceType string `json:"instanceType"`
			Memory       string `json:"memory.size"`
			Processor    string `json:"processor"`
			Region       string `json:"region"`
			Service      string `json:"service"`
		} `json:"metadata"`
		Eccn      string `json:"eccn"`
		Pcq       string `json:"pcq"`
		MatchExpr string `json:"matchExpr"`
		Rates     []struct {
			AccountType string `json:"accountType"`
			Rate        string `json:"rate"`
			Unit        string `json:"unit"`
			UsageExpr   string `json:"usageExpr"`
		} `json:"rates"`
	} `json:"products"`
}

var _ = Describe("Check Premium User Rates", Ordered, Label("Product-Catalog-E2E"), func() {
	It("Validate Premium cloudAccount", func() {
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).NotTo(BeNil(), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
	})

	It("Validating the request of products by Intel Account", func() {
		product_filter := fmt.Sprintf(`{
			"cloudaccountId": "%s",
			"productFilter": { "accountType": "%s"}
		}`, cloudAccIdid, cloudAccountType)
		response_status, response_body := financials.GetProducts(base_url, token, product_filter)
		json.Unmarshal([]byte(response_body), &structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
	})

	It("Validating Product Rates", func() {
		for _, product := range structResponse.Products {
			if len(product.Rates) != 0 {
				Expect(product.Rates[0].AccountType).To(Equal(cloudAccountType))
			} else {
				Fail(product.Name + " has no rates associated.")
			}
		}
	})
})
