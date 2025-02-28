package STaaS_5GB_EU_test

import (
	"fmt"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/testsetup"
	"strings"
	"time"

	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check Enterprise Object Storage", Ordered, Label("Object Storage"), func() {

	// CREATE NEW INSTANCE
	It("Create STaaS Instance", func() {
		fmt.Println("Starting the Instance Creation via STaaS API...")
		fmt.Print(base_url)
		create_response_code, create_response_body := financials.CreateFileSystem(compute_url, token, staas_payload, cloud_account_created)
		Expect(create_response_code).NotTo(Equal(200), "Failed. ObjectStorage was created")
	})
})