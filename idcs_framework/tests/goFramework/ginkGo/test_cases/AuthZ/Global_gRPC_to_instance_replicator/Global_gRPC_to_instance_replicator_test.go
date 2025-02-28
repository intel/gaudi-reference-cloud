package Global_gRPC_to_instance_replicator_test

import (
	"fmt"
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

var _ = Describe("Check vault roles and create a certificate for us-dev instance scheduler", func() {
	It("Check compute communications", func() {
		logger.Log.Info("Retrieve the cloud account Azure token")
		error := false
		time.Sleep(320 * time.Second)
		auth.Get_config_file_data("../../../test_config/authz_resources/authz_config.json")
		authToken, err := auth.Get_Azure_Bearer_Token("premiumjulytestacc@proton.me")

		if err != nil {
			fmt.Print(err)
		}
		bearer := "Bearer " + authToken

		BEARER = bearer

		url_passed := os.Getenv("REGIONAL_URL") + "proto.InstanceService/Ping"
		method := "POST"

		testutils.ResourceRequest(BEARER, url_passed, method)

		Expect(error).To(BeFalse())
	})

	It("Check compute communications 2", func() {
		error := false
		url_passed := os.Getenv("REGIONAL_URL") + "v1/cloudaccounts/{metadata.cloudAccountId}/instances"
		method := "GET"

		testutils.ResourceRequest(BEARER, url_passed, method)

		Expect(error).To(BeFalse())
	})
})
