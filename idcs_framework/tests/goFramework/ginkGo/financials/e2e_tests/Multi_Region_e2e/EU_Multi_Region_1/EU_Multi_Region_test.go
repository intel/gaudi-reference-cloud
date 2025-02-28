package EU_Multi_Region_test

import (
	"fmt"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidwall/gjson"
)

var _ = Describe("Check Enterprise User Rates Multi Region", Ordered, Label("Multi-Region-E2E"), func() {

	It("Validate Enterprise cloudAccount", func() {
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		code, body = financials.GetCloudAccountById(url, token, place_holder_map["cloud_account_id"])
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		fmt.Println("Account type.", cloudAccountType)
		Expect(cloudAccountType).To(Equal("ACCOUNT_TYPE_ENTERPRISE"), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
		code, body = financials.GetCredits(base_url+"/v1/billing/credit", token, cloudAccIdid)
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Credits details")
		cloudAccountCredits := gjson.Get(body, "totalRemainingAmount").Int()
		Expect(cloudAccountCredits).To(Equal(int64(0)), "Failed to validate cloud account credits")
	})

	It("Validate Usages region 1", func() {
		url := base_url + "/v1/billing/usages"
		response_code, response_body := financials.GetCreditsByRegion(url, token, cloud_account_created, regionName)
		fmt.Println(response_body)
		Expect(response_code).To(Equal(200), "Failed to retrieve CloudAccount usage details")
	})
})
