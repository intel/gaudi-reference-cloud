package Global_product_catalog_api_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGlobalProductCatalogApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GlobalProductCatalogApi Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
