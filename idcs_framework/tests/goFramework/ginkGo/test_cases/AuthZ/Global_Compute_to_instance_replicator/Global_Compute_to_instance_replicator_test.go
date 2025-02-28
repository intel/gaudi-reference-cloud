package Global_Compute_to_instance_replicator_test

import (
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/ginkGo/test_cases/testutils"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type CreateCloudAccountEnrollStruct struct {
	Premium bool `json:"premium"`
}

var BEARER = ""

var _ = Describe("Should communicate without giving 500 errors", func() {
	It("Check compute communications", func() {
		logger.Log.Info("Retrieve the cloud account Azure token")
		time.Sleep(40 * time.Second)
		error := false
		auth.Get_config_file_data("../../../test_config/authz_resources/authz_config.json")
		authToken, err := auth.Get_Azure_Bearer_Token("premiumjulytestacc@proton.me")

		if err != nil {
			Fail(err.Error())
		}

		bearer := "Bearer " + authToken

		BEARER = bearer

		url_passed := os.Getenv("REGIONAL_URL") + "/v1/machineimages"

		method := "GET"

		testutils.ResourceRequest(BEARER, url_passed, method)

		Expect(error).To(BeFalse())
	})

	It("Check compute communications 2", func() {
		error := false
		url_passed := os.Getenv("REGIONAL_URL") + "proto.VNetService/Ping"
		method := "POST"

		testutils.ResourceRequest(BEARER, url_passed, method)

		Expect(error).To(BeFalse())
	})
})
