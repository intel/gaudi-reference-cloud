package Regional_compute_api_server

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

var _ = Describe("Check regional compute call functionality", func() {
	It("Api call should return a response code different of 500", func() {
		logger.Log.Info("Retrieve the cloud account Azure token")
		error := false
		time.Sleep(200 * time.Second)
		auth.Get_config_file_data("../../../test_config/authz_resources/authz_config_p2.json")
		authToken, err := auth.Get_Azure_Bearer_Token("premiumjulytestacc@proton.me")

		if err != nil {
			fmt.Print(err)
		}
		bearer := "Bearer " + authToken

		BEARER = bearer

		url_passed := os.Getenv("REGIONAL_URL") + "v1/ping"

		method := "GET"

		testutils.ResourceRequest(BEARER, url_passed, method)

		Expect(error).To(BeFalse())
	})

	It("Api call should return a response code different of 500", func() {
		url_passed := os.Getenv("REGIONAL_URL") + "v1/machineimages"
		method := "GET"
		error := false

		testutils.ResourceRequest(BEARER, url_passed, method)

		Expect(error).To(BeFalse())
	})
})
