package gRPC_internal_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGRPCInternal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GRPCInternal Suite")
}
