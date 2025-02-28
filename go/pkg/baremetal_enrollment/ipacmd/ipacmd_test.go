// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Copyright © 2023 Intel Corporation
package ipacmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mocks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
)

func TestIpacmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ipacmd Suite")
}

var Any = gomock.Any()

var _ = Describe("IPA Command", func() {
	const (
		testRegion   = "us-dev-1"
		bmhIpAddress = "127.0.0.1"
	)

	var (
		ctx context.Context

		mockCtrl      *gomock.Controller
		vault         *mocks.MockSecretManager
		sshManager    *mocks.MockSSHManagerAccessor
		sshPrivateKey string
		err           error
	)

	BeforeEach(func() {
		ctx = context.Background()

		// create mock objects
		mockCtrl = gomock.NewController(GinkgoT())
		vault = mocks.NewMockSecretManager(mockCtrl)
		sshManager = mocks.NewMockSSHManagerAccessor(mockCtrl)
		sshPrivateKey, err = util.GenerateSSHPrivateKey()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("getting New IPA Command Helper", func() {
		It("should initialize dependencies", func() {
			By("initializing Vault client")
			Expect(vault).NotTo(BeNil())
		})

		It("should return command helper", func() {
			vault.EXPECT().GetIPAImageSSHPrivateKey(ctx, Any).Return(sshPrivateKey, nil)
			helper, err := NewIpaCmdHelper(ctx, vault, sshManager, bmhIpAddress, testRegion)
			Expect(helper).ToNot(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("could not return private key", func() {
			vault.EXPECT().GetIPAImageSSHPrivateKey(ctx, Any).Return("", fmt.Errorf("privateKey is empty"))
			helper, err := NewIpaCmdHelper(ctx, vault, sshManager, bmhIpAddress, testRegion)
			Expect(helper).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("private key is not parsale", func() {
			vault.EXPECT().GetIPAImageSSHPrivateKey(ctx, Any).Return("", nil)
			helper, err := NewIpaCmdHelper(ctx, vault, sshManager, bmhIpAddress, testRegion)
			Expect(helper).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})
})
