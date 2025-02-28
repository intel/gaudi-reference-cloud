package PU_Multi_Region_test

import (
	"fmt"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Premium User Rates Multi Region", Ordered, Label("Multi-Region-E2E"), func() {

	It("Validate Premium cloudAccount", func() {
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).To(Equal("ACCOUNT_TYPE_PREMIUM"), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
	})

	It("Validate Usages region 1", func() {
		url := base_url + "/v1/billing/usages"
		response_code, response_body := financials.GetCreditsByRegion(url, token, cloud_account_created, regionName)
		fmt.Println(response_body)
		Expect(response_code).To(Equal(200), "Failed to retrieve CloudAccount usage details")
	})
})
