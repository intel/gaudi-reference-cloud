package Regional_vm_instance_scheduler_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRegionalVmInstanceScheduler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RegionalVmInstanceScheduler Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
