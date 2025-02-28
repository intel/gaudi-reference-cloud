package gRPC_external_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check vault roles and create a certificate for us-dev instance scheduler", func() {
	It("Should issue Certificates for every role in Vault PKI with no errors", func() {
		roles := checkRoles()
		Expect(roles).ToNot(BeNil())
		role := "us-dev-1-grpc-proxy-external"
		common_name := "us-dev-1-grpc-proxy-external.idcs-system.svc.cluster.local"
		Expect(GenerateCertificate(role, common_name)).To(BeTrue())
	})
})
