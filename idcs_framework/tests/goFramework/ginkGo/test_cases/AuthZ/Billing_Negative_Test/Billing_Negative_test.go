package billing_negative_test

import (
	"context"
	"goFramework/framework/common/logger"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("test billing communications", func() {
	It("Should Panic", func() {
		logger.InitializeZapCustomLogger()
		logger.Log.Info("Starting Billing communications test")
		grpcutil.UseTLS = true
		ctx := context.Background()
		EmbedService(ctx)
		Expect(func() {
			grpcutil.StartTestServices(ctx)
		}).To(Panic())
		logger.Log.Info("Stoping services...")
	})
})
