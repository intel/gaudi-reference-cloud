package Portal_UI_global_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPortalUIGlobal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PortalUIGlobal Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
