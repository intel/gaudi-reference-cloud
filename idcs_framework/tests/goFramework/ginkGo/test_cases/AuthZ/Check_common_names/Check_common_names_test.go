package Check_common_names_test

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const ALLOWED_DOMAINS_PATH = "../../../test_config/authz_resources/authz_config.json"

var _ = Describe("Check and validate common names in vault roles", func() {
	It("Should validate common names in PKI with expected roles JSON", func() {
		roles := checkRoles()
		Expect(roles).ToNot(BeNil())
		values := iterateRolesAndCheckValue(roles, "allowed_domains")
		Expect(values).ToNot(BeNil())
		expected_values := retrieveValuesFromJson("domains", ALLOWED_DOMAINS_PATH)
		Expect(values).ToNot(BeNil())

		for _, key := range values {
			result := checkForValue(key.(string), expected_values)
			fmt.Println("Role " + key.(string) + " found: " + strconv.FormatBool(result))
		}

		Expect(values).ToNot(BeNil())
	})
})
