package Check_create_certificate_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCheckCreateCertificate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CheckCreateCertificate Suite")
}
