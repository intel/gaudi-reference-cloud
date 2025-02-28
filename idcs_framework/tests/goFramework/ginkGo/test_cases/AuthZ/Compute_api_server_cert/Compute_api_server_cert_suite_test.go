package Compute_api_server_cert_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestComputeApiServerCert(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ComputeApiServerCert Suite")
}
