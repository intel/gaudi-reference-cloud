package Global_gRPC_to_instance_replicator_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGlobalGRPCToInstanceReplicator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GlobalGRPCToInstanceReplicator Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
