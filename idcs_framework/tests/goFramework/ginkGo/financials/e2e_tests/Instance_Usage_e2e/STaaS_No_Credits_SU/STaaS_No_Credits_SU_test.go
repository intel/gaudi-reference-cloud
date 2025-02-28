package STaaS_No_Credits_SU_test

package STaaS_5GB_PU_test

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

var _ = Describe("Check Premium User Product Usage of STaaS without credits.", Ordered, Label("Usages-e2e"), func() {
	// CREATE NEW INSTANCE
	It("Create STaaS Instance without credits", func() {
		fmt.Println("Starting the Instance Creation via STaaS API...")
		fmt.Print(base_url)
		create_response_code, create_response_body := financials.CreateFileSystem(compute_url, token, staas_payload, cloud_account_created)
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		fmt.Println("resourceId: ", instance_id_created)
		Expect(create_response_code).NotTo(Equal(200), "User should not be able to create fileSystems without credits...")
	})
})