// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/instance_operator/..." make test-custom
package privatecloud

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVmController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VM Instance Controller Suite")
}

var _ = Describe("CreateVmErrors, Validating IPv4 Addresses in error messages", func() {
	Context("Valid IP Addresses Found", func() {
		It("should correctly match valid IPv4 addresses", func() {
			Expect(ipPattern.MatchString("failed to connect to 192.168.1.10 due to timeout")).To(BeTrue())
			Expect(ipPattern.MatchString("failed to connect: 255.255.255.255")).To(BeTrue())
			Expect(ipPattern.MatchString("failed to connect to 1.2.3.4: timeout.")).To(BeTrue())
			Expect(ipPattern.MatchString("failed to connect to 001.002.003.040: timeout.")).To(BeTrue())
		})
	})

	Context("Invalid IP Addresses Found", func() {
		It("should not match invalid IPv4 addresses", func() {
			Expect(ipPattern.MatchString("Invalid IP: 999.999.999.999")).To(BeFalse())
			Expect(ipPattern.MatchString("Another Invalid IP: 256.100.50.25")).To(BeFalse())
			Expect(ipPattern.MatchString("some random error message without an IP")).To(BeFalse())
			Expect(ipPattern.MatchString("")).To(BeFalse())
		})
	})
})
