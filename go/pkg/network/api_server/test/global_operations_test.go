// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	defaultCidrForVPC    = "172.31.0.0/16"
	defaultCidrForSubnet = "172.31.0.0/20"
)

var _ = Describe("Create Default", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	Context("Default API", func() {
		It("Create Default VPC should succeed", func() {
			cloudAccountId := cloudaccount.MustNewId()
			defaultReq := &pb.CreateDefaultRequest{
				Metadata: &pb.VPCMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
			}
			defaultVPC, err := globalOperationsServiceClient.CreateDefault(ctx, defaultReq)
			Expect(err).Should(Succeed())
			Expect(defaultVPC.Metadata.Name).Should(Equal(defaultVPC.Metadata.ResourceId))
			Expect(defaultVPC.Spec.CidrBlock).Should(Equal(defaultCidrForVPC))

			// get the corresponding Subnet.
			subnetsForVPC, err := subnetServiceClient.Search(ctx, &pb.SubnetSearchRequest{
				Metadata: &pb.SubnetMetadataSearch{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.SubnetSpec{
					VpcId: defaultVPC.Metadata.ResourceId,
				},
			})
			Expect(err).Should(Succeed())

			// check that each subnet contains default and properly connected to VPC.
			for _, subnet := range subnetsForVPC.Items {
				Expect(subnet.Spec.VpcId).Should(Equal(defaultVPC.Metadata.ResourceId))
				Expect(subnet.Spec.CidrBlock).Should(Equal(defaultCidrForSubnet))
			}

		})
		It("Create 2 Default Networks should succeed", func() {
			// Create first default network
			cloudAccountId := cloudaccount.MustNewId()
			defaultReq := &pb.CreateDefaultRequest{
				Metadata: &pb.VPCMetadataCreate{
					CloudAccountId: cloudAccountId,
				},
			}
			defaultVPC, err := globalOperationsServiceClient.CreateDefault(ctx, defaultReq)
			Expect(err).Should(Succeed())
			Expect(defaultVPC.Metadata.Name).Should(Equal(defaultVPC.Metadata.ResourceId))
			Expect(defaultVPC.Spec.CidrBlock).Should(Equal(defaultCidrForVPC))

			// Create another default network
			anotherDefaultVPC, err := globalOperationsServiceClient.CreateDefault(ctx, defaultReq)
			Expect(err).Should(Succeed())
			Expect(anotherDefaultVPC.Metadata.Name).Should(Equal(anotherDefaultVPC.Metadata.ResourceId))
			Expect(anotherDefaultVPC.Spec.CidrBlock).Should(Equal(defaultCidrForVPC))
			// Make sure the first and second are diffrent
			Expect(anotherDefaultVPC.Metadata.Name).Should(Not(Equal(defaultVPC.Metadata.Name)))

		})
	})
})
