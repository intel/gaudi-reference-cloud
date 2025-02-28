package Check_Updated_Product_Test_test

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

func ValidateProductInJson(productName string) bool {
	for _, product := range structResponseJSON.Products {
		if product.Name == productName {
			return true
		}
	}
	return false
}

var _ = Describe("Check Intel User Rates", Ordered, Label("Product-Catalog-E2E-Intel"), func() {
	var cloudAccountType string
	var cloudAccIdid string
	var cloud_account_created string
	var place_holder_map = make(map[string]string)
	var userToken string

	It("Delete cloud account", func() {
		fmt.Println("Delete cloud account")
		url := base_url + "/v1/cloudaccounts"
		cloudAccId, err := testsetup.GetCloudAccountId(userName, base_url, token)
		if err == nil {
			financials.DeleteCloudAccountById(url, token, cloudAccId)
		}
	})

	// CREATE CLOUDACCOUNT
	It("Create cloud account", func() {
		// Generating token wtih payload for cloud account creation with enroll API
		token_response, _ := auth.Get_Azure_Bearer_Token(userName)
		fmt.Print("TOKEN...", token_response)
		userToken = "Bearer " + token_response
		cloudaccount_creation_status, cloudaccount_creation_body := financials.CreateCloudAccountE2E(base_url, token, userToken, userName, false)
		fmt.Println("BODY...", cloudaccount_creation_body)
		Expect(cloudaccount_creation_status).To(Equal(200), "Failed to create Cloud Account using enroll")
		cloud_account_created = gjson.Get(cloudaccount_creation_body, "cloudAccountId").String()
		cloudaccount_type := gjson.Get(cloudaccount_creation_body, "cloudAccountType").String()
		Expect(strings.Contains(cloudaccount_creation_body, `"cloudAccountType":"ACCOUNT_TYPE_INTEL"`)).To(BeTrue(), "Validation failed on Enrollment of Intel user")
		place_holder_map["cloud_account_id"] = cloud_account_created
		place_holder_map["cloud_account_type"] = cloudaccount_type
		fmt.Println("cloudAccount_id", cloud_account_created)
	})

	It("Validate Intel cloudAccount", func() {
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).NotTo(BeNil(), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
	})

	It("Validating the request of products by Intel Account", func() {
		logger.Logf.Info("Getting all products for Intel Users")
		product_filter := fmt.Sprintf(`{
			"cloudaccountId": "%s",
			"productFilter": { "accountType": "%s", "name": "bm-spr-test"}
		}`, cloudAccIdid, cloudAccountType)
		response_status, response_body := financials.GetProducts(base_url, token, product_filter)
		json.Unmarshal([]byte(response_body), &structResponse)
		Expect(response_status).To(Equal(200), "Failed to retrieve Product Details")
		logger.Log.Info("Products Successfully retrieved")
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

	It("Get Products from JSON", func() {
		response := financials_utils.GetProductsFromJson()
		err := json.Unmarshal(response, &structResponseJSON)
		Expect(err).To(BeNil(), "Failed to unmarshall products")
		fmt.Print("Struct from JSON", structResponseJSON)
	})

	It("Validating Products", func() {
		for _, product := range structResponse.Products {
			if product.Name != "bm-spr-test" {
				Expect(ValidateProductInJson(product.Name)).To(BeTrue(), "Product: "+product.Name+" Not found in JSON results.")
			}
		}
	})
})
