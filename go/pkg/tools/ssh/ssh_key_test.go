// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test_tools

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSH Key Tests", func() {
	It("CreateSshRsaKeyPair", func() {
		comment := "user@example.com"
		privateKey, publicKey, err := CreateSshRsaKeyPair(4096, comment)
		Expect(err).Should(Succeed())
		Expect(privateKey).ShouldNot(BeEmpty())
		Expect(publicKey).ShouldNot(BeEmpty())
	})
})
