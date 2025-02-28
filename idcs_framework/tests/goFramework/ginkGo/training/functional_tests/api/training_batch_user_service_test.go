package api_test

import (
	"goFramework/framework/service_api/training"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("TrainingBatchUserService", Ordered, Label("TrainingBatchUserService"), func() {
	It("should fail to list expiry for unregistered cloud account", func() {
		status, response := training.GetExpiryTimeById(baseApiUrl, standardCloudAccount["token"], standardCloudAccount["id"])
		Expect(status).To(Equal(500), "GetExpiryTimeById didn't fail when it should have")
		Expect(strings.Contains(response, "Our services are currently experiencing an interruption.")).To(BeTrue(), "Recieved something other than expected error")
	})

	It("should list a real exipry date for a registered cloud account", func() {
		status, response := training.GetExpiryTimeById(baseApiUrl, premiumCloudAccount["token"], premiumCloudAccount["id"])
		Expect(status).To(Equal(200), "GetExpiryTimeById failed")
		expiry := gjson.Get(response, "expiryDate").String()
		Expect(expiry).ToNot(BeNil())
	})
})
