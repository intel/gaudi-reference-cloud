// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StorageControllerClient", func() {
	var (
		client *sc.StorageControllerClient
	)

	BeforeEach(func() {
		client = &sc.StorageControllerClient{}
	})

	Context("Init", func() {
		Context("With a valid server address", func() {
			var (
				mockServerAddr = "localhost:12345"
				mockMtlsFlag   = false
			)

			It("should initialize the client without errors", func() {
				err := client.Init(context.Background(), mockServerAddr, mockMtlsFlag)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.ClusterSvcClient).NotTo(BeNil())
				Expect(client.NamespaceSvcClient).NotTo(BeNil())
				Expect(client.WekaFilesystemSvcClient).NotTo(BeNil())
				Expect(client.UserSvcClient).NotTo(BeNil())
			})
		})

		Context("With an empty server address", func() {
			It("should return an error", func() {
				err := client.Init(context.Background(), "", false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("storage controller server address missing"))
			})
		})

	})
})
