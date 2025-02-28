package Check_ttl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check tll value is set", func() {
	It("Should Check ttl value is set for all services", func() {
		roles := checkRoles()
		Expect(roles).ToNot(BeNil())
		check := iterateRolesToGetValue(roles)
		Expect(check).ToNot(BeFalse())
	})
})
