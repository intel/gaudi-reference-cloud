package Compute_api_server_cert_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check vault roles and create a certificate for us-dev instance scheduler", func() {
	It("Should issue Certificates for every role in Vault PKI with no errors", func() {
		roles := checkRoles()
		Expect(roles).ToNot(BeNil())
		role := "us-dev-1-compute-api-server"
		common_name := "us-dev-1-compute-api-server.idcs-system.svc.cluster.local"
		Expect(GenerateCertificate(role, common_name)).To(BeTrue())
	})
})
