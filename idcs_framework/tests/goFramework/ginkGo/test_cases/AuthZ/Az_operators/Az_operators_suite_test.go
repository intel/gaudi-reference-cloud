package AZ_operators_compute_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAzOperators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AzOperators Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
