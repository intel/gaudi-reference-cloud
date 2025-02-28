package gRPC_external_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGRPCExternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GRPCExternal Suite")
}
