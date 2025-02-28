package Check_vault_roles_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check vault roles", func() {
	It("Should get roles from Vault PKI successfully", func() {
		roles := checkRoles()
		Expect(roles).ToNot(BeNil())
	})
})
