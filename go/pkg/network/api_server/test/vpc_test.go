// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("VPC API Integration Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	Context("Create VPC successfully", func() {
		It("Create VPC should succeed", func() {
			cidr := "10.0.0.0/16"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			got, err := vpcServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())

			Expect(got.Metadata.Name).Should(Equal(name))
			Expect(got.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(got.Metadata.ResourceId).ShouldNot(Equal(name))
			Expect(got.Metadata.CloudAccountId).Should(Equal(cloudAccountId))
			Expect(got.Metadata.Labels).Should(Equal(defaultLabels))
			Expect(got.Spec.CidrBlock).Should(Equal(cidr))
		})
		It("Create VPC with name of a deleted vpc should succeed", func() {
			// Create VPC
			cidr := "10.0.0.0/16"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			got, err := vpcServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())

			Expect(got.Metadata.Name).Should(Equal(name))

			// Delete vpc.
			_, err = vpcServiceClient.Delete(ctx, &pb.VPCDeleteRequest{
				Metadata: &pb.VPCMetadataReference{
					CloudAccountId: cloudAccountId,
					NameOrId:       &pb.VPCMetadataReference_ResourceId{ResourceId: got.Metadata.ResourceId},
				},
			})
			Expect(err).Should(Succeed())

			// actually delete from db. (mock the operator success)
			query := fmt.Sprintf("UPDATE vpc SET deleted_timestamp = NOW() WHERE resource_id = '%s'", got.Metadata.ResourceId)
			err = runNetworkDBQuery(ctx, query)
			Expect(err).Should(Succeed())

			// create another vpc with same name.
			createReq = NewCreateVPCRequest(cloudAccountId, name, cidr)
			got, err = vpcServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(got.Metadata.Name).Should(Equal(name))
		})
	})

	Context("Create VPC should fail", func() {
		It("Create VPC with existing name should fail", func() {
			// Create VPC
			cidr := "10.0.0.0/16"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			got, err := vpcServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())

			Expect(got.Metadata.Name).Should(Equal(name))

			// create another vpc with same name.
			got, err = vpcServiceClient.Create(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.AlreadyExists))
		})
	})

	Context("Create VPC with invalid CIDR should fail", func() {
		It("Should fail when netmask prefix is less then 16", func() {
			cidr := "10.0.0.0/15"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			_, err := vpcServiceClient.Create(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
			Expect(err.Error()).To(ContainSubstring("Invalid CIDR"))
		})
		It("Should fail when netmask prefix is more then 28", func() {
			cidr := "10.0.0.0/30"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			_, err := vpcServiceClient.Create(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
			Expect(err.Error()).To(ContainSubstring("Invalid CIDR"))
		})
		It("Should fail when netmask prefix is invalid", func() {
			cidrs := []string{
				"10.0.0.0/invalid",           // invalid netmask
				"192.168.0.0/255.255.255.85", // invalid canonical: mask has non-contiguous 1s and 0s
			}

			for _, cidr := range cidrs {
				name := "default"

				cloudAccountId := cloudaccount.MustNewId()
				createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
				_, err := vpcServiceClient.Create(ctx, createReq)
				Expect(err).ShouldNot(Succeed())
				Expect(err.Error()).To(ContainSubstring("Invalid CIDR"))
			}
		})

		It("Should fail when netmask address is not IPv4", func() {
			cidrs := []string{
				"10.20.30.41.50/24",
				"260.254.5.1/24",
				"2001:db8::/32",
			}

			for _, cidr := range cidrs {
				name := "default"

				cloudAccountId := cloudaccount.MustNewId()
				createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
				_, err := vpcServiceClient.Create(ctx, createReq)
				Expect(err).ShouldNot(Succeed())
				Expect(err.Error()).To(ContainSubstring("Invalid CIDR"))

			}
		})

		It("Should fail when netmask overlap with locallink block (169.254.0.0/16)", func() {
			cidrs := []string{
				"169.252.0.1/14", // (169.252.0.1 - 169.255.255.254) => Start before 169.254.0.0 and end after 169.254.0.0/16
				"169.254.5.1/24", // (169.254.0.1 - 169.254.255.254) => Inside locallink
			}

			for _, cidr := range cidrs {
				name := "default"

				cloudAccountId := cloudaccount.MustNewId()
				createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
				_, err := vpcServiceClient.Create(ctx, createReq)
				Expect(err).ShouldNot(Succeed())
				Expect(err.Error()).To(ContainSubstring("Invalid CIDR"))
			}
		})

	})

	Context("Update API", func() {
		It("Update labels by VPC name should succeed", func() {
			cidr := "10.0.0.0/16"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			createResp, err := vpcServiceClient.Create(ctx, createReq)
			resourceId := createResp.Metadata.ResourceId
			Expect(err).Should(Succeed())

			updateReq := NewUpdateVPCByNameRequest(cloudAccountId, name)
			_, updateErr := vpcServiceClient.Update(ctx, updateReq)
			Expect(updateErr).Should(Succeed())

			getReq := NewGetVPCByIdRequest(cloudAccountId, resourceId)
			getResp, err := vpcServiceClient.Get(ctx, getReq)
			Expect(err).Should(Succeed())

			Expect(getResp.Metadata.Name).Should(Equal(name))
			Expect(getResp.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId))
			Expect(getResp.Metadata.Labels).Should(Equal(updatedLabels))
		})

		It("Update labels by vpc id should succeed", func() {
			cidr := "10.0.0.0/16"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			createResp, err := vpcServiceClient.Create(ctx, createReq)
			resourceId := createResp.Metadata.ResourceId
			Expect(err).Should(Succeed())

			updateReq := NewUpdateVPCByIdRequest(cloudAccountId, resourceId)
			_, updateRespErr := vpcServiceClient.Update(ctx, updateReq)
			Expect(updateRespErr).Should(Succeed())

			getReq := NewGetVPCByIdRequest(cloudAccountId, resourceId)
			getResp, err := vpcServiceClient.Get(ctx, getReq)
			Expect(err).Should(Succeed())

			Expect(getResp.Metadata.Name).Should(Equal(name))
			Expect(getResp.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId))
			Expect(getResp.Metadata.Labels).Should(Equal(updatedLabels))
		})

		It("Update req with missing metadata should not succeed", func() {
			cidr := "10.0.0.0/16"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			createResp, err := vpcServiceClient.Create(ctx, createReq)
			resourceId := createResp.Metadata.ResourceId
			Expect(err).Should(Succeed())

			updateReq := NewUpdateVPCByIdRequest_MissingMetadata(cloudAccountId, resourceId)
			_, updateRespErr := vpcServiceClient.Update(ctx, updateReq)
			Expect(updateRespErr).ShouldNot(Succeed())
			Expect(updateRespErr.Error()).To(ContainSubstring("missing metadata"))
		})

	})

	Context("Get API", func() {
		It("Get by VPC Id should succeed", func() {
			cidr := "10.0.0.0/16"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			createResp, err := vpcServiceClient.Create(ctx, createReq)
			resourceId := createResp.Metadata.ResourceId
			Expect(err).Should(Succeed())

			getReq := NewGetVPCByIdRequest(cloudAccountId, resourceId)
			getResp, err := vpcServiceClient.Get(ctx, getReq)
			Expect(err).Should(Succeed())

			Expect(getResp.Metadata.Name).Should(Equal(name))
			Expect(getResp.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId))
			Expect(getResp.Metadata.Labels).Should(Equal(defaultLabels))
		})

		It("Get by VPC Name should succeed", func() {
			cidr := "10.0.0.0/16"
			name := "default"

			cloudAccountId := cloudaccount.MustNewId()
			createReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
			_, err := vpcServiceClient.Create(ctx, createReq)
			Expect(err).Should(Succeed())

			getReq := NewGetVPCByNameRequest(cloudAccountId, name)
			getResp, err := vpcServiceClient.Get(ctx, getReq)
			Expect(err).Should(Succeed())

			Expect(getResp.Metadata.Name).Should(Equal(name))
			Expect(getResp.Metadata.ResourceId).Should(Not(BeEmpty()))
			Expect(getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId))
			Expect(getResp.Metadata.Labels).Should(Equal(defaultLabels))
		})
	})
})
