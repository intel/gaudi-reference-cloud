package compute_usage

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func init() {
}

var _ = func() bool { testing.Init(); return true }()

func TestRegressionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	logger.Init()
	RunSpecs(t, "IDC - compute util testing")
}

var _ = BeforeSuite(func() {
	logger.Init()
	os.Setenv("proxy_required", "true")
})

var _ = AfterSuite(func() {

})
