package Global_gRPC_to_compute_api_comm_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGlobalGRPCToComputeApiComm(t *testing.T) {
	if os.Getenv("MULTI_RUNNER") != "" {
		t.Skip("Skipping not suitable for multi runner container")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "GlobalGRPCToComputeApiComm Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
