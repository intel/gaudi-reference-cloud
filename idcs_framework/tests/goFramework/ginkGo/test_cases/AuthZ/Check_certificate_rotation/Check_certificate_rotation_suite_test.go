package Check_certificate_rotation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCheckCertificateRotation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CheckCertificateRotation Suite")
}
