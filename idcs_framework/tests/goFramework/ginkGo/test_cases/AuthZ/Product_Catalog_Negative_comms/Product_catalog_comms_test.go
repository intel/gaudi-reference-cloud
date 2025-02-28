package product_catalog_service_negative_test

import (
	"context"
	"goFramework/framework/common/logger"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("test Product catalog communications", func() {
	It("Should panic", func() {
		logger.InitializeZapCustomLogger()
		logger.Log.Info("Starting Console communications test suite")
		grpcutil.UseTLS = true
		ctx := context.Background()
		EmbedService(ctx)
		Expect(func() {
			grpcutil.StartTestServices(ctx)
		}).To(Panic())
		logger.Log.Info("Stoping services")
	})
})
