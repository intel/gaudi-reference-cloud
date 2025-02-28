package gRPC_reflect_comms_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGRPCReflectComms(t *testing.T) {
	if os.Getenv("MULTI_RUNNER") != "" {
		t.Skip("Skipping not suitable for multi runner container")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "GRPCReflectComms Suite")
}
