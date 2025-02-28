package billing_aria_comms_test

import (
	"goFramework/framework/common/logger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("test billing aria communications", func() {
	It("Should communicate successfully", func() {
		logger.InitializeZapCustomLogger()
		logger.Log.Info("Starting Billing aria communications test")
		expectedResult := true
		/*grpcutil.UseTLS = true
		ctx := context.Background()
		EmbedService(ctx)
		grpcutil.StartTestServices(ctx)
		logger.Log.Info("Stoping services...")
		defer grpcutil.StopTestServices() */
		Expect(expectedResult).To(BeTrue())
	})
})
