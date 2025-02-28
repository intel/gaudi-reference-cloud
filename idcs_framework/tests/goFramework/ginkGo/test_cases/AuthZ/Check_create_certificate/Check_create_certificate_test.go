package Check_create_certificate_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check vault roles and create a certificate", func() {
	It("Should issue Certificates for every role in Vault PKI with no errors", func() {
		roles := checkRoles()
		Expect(roles).ToNot(BeNil())
		domains := iterateRolesAndCheckValue(roles, "allowed_domains")
		Expect(domains).ToNot(BeNil())
		Expect(iterateRolesToIssueCertificates(roles)).To(BeTrue())
	})
})
