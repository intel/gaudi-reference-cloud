package Metering_monitoring_service_test

import (
	"goFramework/framework/common/logger"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMeteringMonitoringApiService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MeteringMonitoringApiService Suite")
}

var _ = BeforeSuite(func() {
	os.Setenv("Test_Suite", "ginkGo")
	logger.InitializeZapCustomLogger()
})
