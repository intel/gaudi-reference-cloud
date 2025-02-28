package Check_common_names_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCheckCommonNames(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CheckCommonNames Suite")
}
