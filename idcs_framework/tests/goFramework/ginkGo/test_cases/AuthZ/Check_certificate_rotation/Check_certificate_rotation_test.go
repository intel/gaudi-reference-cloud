package Check_certificate_rotation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check certificate rotate functionality", func() {
	It("Should rotate Certificates successfully", func() {
		result := RotateCertificates()
		Expect(result).ToNot(BeFalse())
	})
})
