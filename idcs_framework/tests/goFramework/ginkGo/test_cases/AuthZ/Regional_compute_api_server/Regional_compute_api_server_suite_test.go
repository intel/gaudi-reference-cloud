package Regional_compute_api_server_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRegionalComputeApiServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RegionalComputeApiServer Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
