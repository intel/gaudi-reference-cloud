package Global_grpc_rest_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGlobalGrpcRest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GlobalGrpcRest Suite")
}
