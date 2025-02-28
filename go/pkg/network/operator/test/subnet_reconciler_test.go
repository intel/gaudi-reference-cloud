// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Create subnet, happy path", func() {
	ctx := context.Background()
	defer GinkgoRecover()

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	Context("Subnet integration tests", func() {
		It("Reconciles a new subnet", func() {

			cloudaccountId, err := NewCloudAcctId()
			Expect(err).Should(Succeed())

			// Create VPC
			vpcName := uuid.NewString()
			vpc, err := vpcServiceClient.Create(ctx, &pb.VPCCreateRequest{
				Metadata: &pb.VPCMetadataCreate{
					CloudAccountId: cloudaccountId,
					Name:           vpcName,
				},
				Spec: &pb.VPCSpec{
					CidrBlock: "10.0.0.0/16",
				},
			})
			Expect(err).Should(Succeed())

			By("Waiting for vpc to be created in SDN")
			Eventually(func(g Gomega) {
				got, err := vpcPrivateServiceClient.GetPrivate(ctx, &pb.VPCGetPrivateRequest{
					Metadata: &pb.VPCMetadataReference{
						CloudAccountId: cloudaccountId,
						NameOrId:       &pb.VPCMetadataReference_ResourceId{ResourceId: vpc.Metadata.ResourceId},
					},
				})
				g.Expect(err).Should(Succeed())
				g.Expect(got.Spec.CidrBlock).Should(Equal("10.0.0.0/16"))
				g.Expect(got.Status.Phase).Should(Equal(pb.VPCPhase_VPCPhase_Ready))
				g.Expect(got.Status.Message).Should(Equal("VPC ready"))
			}, timeout).Should(Succeed())

			// Create Subnet
			subnetName := uuid.NewString()
			subnet, err := subnetServiceClient.Create(ctx, &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudaccountId,
					Name:           subnetName,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.0/16",
					AvailabilityZone: "us-dev-1a",
					VpcId:            vpc.Metadata.ResourceId,
				},
			})
			Expect(err).Should(Succeed())

			By("Waiting for subnet to be created in SDN")
			Eventually(func(g Gomega) {
				got, err := subnetPrivateServiceClient.GetPrivate(ctx, &pb.SubnetGetPrivateRequest{
					Metadata: &pb.SubnetMetadataReference{
						CloudAccountId: cloudaccountId,
						NameOrId: &pb.SubnetMetadataReference_ResourceId{
							ResourceId: subnet.Metadata.ResourceId,
						},
					},
				})

				g.Expect(err).Should(Succeed())
				g.Expect(got.Spec.CidrBlock).Should(Equal("10.0.0.0/16"))
				g.Expect(got.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Ready))
				g.Expect(got.Status.Message).Should(Equal("Subnet ready"))
			}, timeout).Should(Succeed())
		})
		It("Reconciles a subnet delete", func() {
			cloudaccountId, err := NewCloudAcctId()
			Expect(err).Should(Succeed())

			// Create VPC
			vpcName := uuid.NewString()
			vpc, err := vpcServiceClient.Create(ctx, &pb.VPCCreateRequest{
				Metadata: &pb.VPCMetadataCreate{
					CloudAccountId: cloudaccountId,
					Name:           vpcName,
				},
				Spec: &pb.VPCSpec{
					CidrBlock: "10.0.0.0/16",
				},
			})
			Expect(err).Should(Succeed())

			By("Waiting for vpc to be created in SDN")
			Eventually(func(g Gomega) {
				got, err := vpcPrivateServiceClient.GetPrivate(ctx, &pb.VPCGetPrivateRequest{
					Metadata: &pb.VPCMetadataReference{
						CloudAccountId: cloudaccountId,
						NameOrId:       &pb.VPCMetadataReference_ResourceId{ResourceId: vpc.Metadata.ResourceId},
					},
				})
				g.Expect(err).Should(Succeed())
				g.Expect(got.Spec.CidrBlock).Should(Equal("10.0.0.0/16"))
				g.Expect(got.Status.Phase).Should(Equal(pb.VPCPhase_VPCPhase_Ready))
				g.Expect(got.Status.Message).Should(Equal("VPC ready"))
			}, timeout).Should(Succeed())

			// Create Subnet
			subnetName := uuid.NewString()
			subnet, err := subnetServiceClient.Create(ctx, &pb.SubnetCreateRequest{
				Metadata: &pb.SubnetMetadataCreate{
					CloudAccountId: cloudaccountId,
					Name:           subnetName,
				},
				Spec: &pb.SubnetSpec{
					CidrBlock:        "10.0.0.0/16",
					AvailabilityZone: "us-dev-1a",
					VpcId:            vpc.Metadata.ResourceId,
				},
			})
			Expect(err).Should(Succeed())

			By("Waiting for subnet to be created in SDN")
			Eventually(func(g Gomega) {
				got, err := subnetPrivateServiceClient.GetPrivate(ctx, &pb.SubnetGetPrivateRequest{
					Metadata: &pb.SubnetMetadataReference{
						CloudAccountId: cloudaccountId,
						NameOrId: &pb.SubnetMetadataReference_ResourceId{
							ResourceId: subnet.Metadata.ResourceId,
						},
					},
				})
				g.Expect(err).Should(Succeed())
				g.Expect(got.Spec.CidrBlock).Should(Equal("10.0.0.0/16"))
				g.Expect(got.Status.Phase).Should(Equal(pb.SubnetPhase_SubnetPhase_Ready))
				g.Expect(got.Status.Message).Should(Equal("Subnet ready"))
			}, timeout).Should(Succeed())

			By("Deleting subnet")
			deleteSubnetReq := &pb.SubnetDeleteRequest{
				Metadata: &pb.SubnetMetadataReference{
					CloudAccountId: cloudaccountId,
					NameOrId:       &pb.SubnetMetadataReference_ResourceId{ResourceId: subnet.Metadata.ResourceId},
				},
				Spec: &pb.SubnetSpec{
					VpcId: vpc.Metadata.ResourceId,
				},
			}

			_, err = subnetServiceClient.Delete(ctx, deleteSubnetReq)
			Expect(err).Should(Succeed())

			By("Waiting for subnet to be deleted in SDN")
			Eventually(func(g Gomega) {
				_, err := subnetPrivateServiceClient.GetPrivate(ctx, &pb.SubnetGetPrivateRequest{
					Metadata: &pb.SubnetMetadataReference{
						CloudAccountId: cloudaccountId,
						NameOrId: &pb.SubnetMetadataReference_ResourceId{
							ResourceId: subnet.Metadata.ResourceId,
						},
					},
				})
				g.Expect(err).ShouldNot(Succeed())
				g.Expect(status.Code(err)).Should(Equal(codes.NotFound)) //
			})
		})
	})
})
