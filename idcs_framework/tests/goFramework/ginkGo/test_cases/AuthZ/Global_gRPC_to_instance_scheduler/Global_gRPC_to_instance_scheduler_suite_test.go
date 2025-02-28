package Global_gRPC_to_instance_scheduler_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGlobalGRPCToInstanceScheduler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GlobalGRPCToInstanceScheduler Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
