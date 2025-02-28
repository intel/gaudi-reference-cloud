package Portal_UI_regional_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPortalUIRegion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PortalUIRegion Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
