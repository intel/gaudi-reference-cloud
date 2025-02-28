package Check_ttl_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCheckTtl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CheckTtl Suite")
}
