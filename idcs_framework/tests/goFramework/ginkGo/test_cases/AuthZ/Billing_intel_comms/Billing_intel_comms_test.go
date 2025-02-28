package billing_intel_comms_test

import (
	"context"
	"goFramework/framework/common/logger"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("test billing Intel communications", func() {
	It("Should communicate successfully", func() {
		logger.InitializeZapCustomLogger()
		logger.Log.Info("Starting Billing intel communications test")
		expectedResult := true
		grpcutil.UseTLS = true
		ctx := context.Background()
		EmbedService(ctx)
		grpcutil.StartTestServices(ctx)
		logger.Log.Info("Stoping services...")
		defer grpcutil.StopTestServices()
		Expect(expectedResult).To(BeTrue())
	})
})
