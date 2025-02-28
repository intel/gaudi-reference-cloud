package Regional_cloudaccount_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRegionalCloudAccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RegionalCloudAccount Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
