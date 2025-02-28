package Global_gRPC_to_instance_scheduler_test

import (

	//"goFramework/utils"

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
	It("Should issue Certificates for every role in Vault PKI with no errors", func() {
		logger.Log.Info("Retrieve the cloud account Azure token")
		error := false
		time.Sleep(420 * time.Second)
		auth.Get_config_file_data("../../../test_config/authz_resources/authz_config.json")
		authToken, err := auth.Get_Azure_Bearer_Token("premiumjulytestacc@proton.me")

		if err != nil {
			fmt.Print(err)
		}
		bearer := "Bearer " + authToken

		BEARER = bearer

		url_passed := os.Getenv("REGIONAL_URL") + "v1/instancetypes"

		method := "GET"

		testutils.ResourceRequest(BEARER, url_passed, method)

		Expect(error).To(BeFalse())
	})

	It("Should issue Certificates for every role in Vault PKI with no errors", func() {
		url_passed := os.Getenv("REGIONAL_URL") + "proto.InstanceTypeService/SearchStream"
		error := false
		method := "POST"

		testutils.ResourceRequest(BEARER, url_passed, method)

		Expect(error).To(BeFalse())
	})
})
