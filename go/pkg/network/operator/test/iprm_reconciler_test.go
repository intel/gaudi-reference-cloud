// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Port happy path", func() {
	ctx := context.Background()
	defer GinkgoRecover()

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	Context("iprm integration tests", func() {
		It("Reconcile ReservePort when GetPort return port does not exists in sdnController", func() {
			sdnServiceClient.EXPECT().GetPort(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, status.Error(codes.NotFound, "not found")).AnyTimes()
			sdnServiceClient.EXPECT().CreatePort(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

			cloudAccountId, err := NewCloudAcctId()
			Expect(err).Should(Succeed())

			vpcName := uuid.NewString()
			_, subnet, err := createVPCAndSubnet(ctx, cloudAccountId, vpcName)

			// Reserve port for the subnet
			portReserved, err := iprmPrivateServiceClient.ReservePort(ctx, &pb.ReservePortRequest{
				Metadata: &pb.PortMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.PortSpecPrivate{
					SubnetId: subnet.Metadata.ResourceId,
				},
			})

			Expect(err).Should(Succeed())
			By("Waiting for port to be created in SDN3")
			Eventually(func(g Gomega) {
				got, err := iprmPrivateServiceClient.GetPortPrivate(ctx, &pb.GetPortPrivateRequest{
					Metadata: &pb.PortMetadataReference{
						CloudAccountId: cloudAccountId,
						ResourceId:     portReserved.Metadata.ResourceId,
					},
				})

				g.Expect(err).Should(Succeed())
				g.Expect(got.Status.Phase).Should(Equal(pb.PortPhase_PortPhase_Ready))
				g.Expect(got.Status.Message).Should(Equal("Port ready"))
			}, timeout).Should(Succeed())

		})

		It("Reconcile ReservePort when GetPort return port exists", func() {
			// TODO: GetPort return port exists
			sdnServiceClient.EXPECT().GetPort(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

			cloudAccountId, err := NewCloudAcctId()
			Expect(err).Should(Succeed())

			vpcName := uuid.NewString()
			_, subnet, err := createVPCAndSubnet(ctx, cloudAccountId, vpcName)

			// Reseave port for the subnet
			portReserved, err := iprmPrivateServiceClient.ReservePort(ctx, &pb.ReservePortRequest{
				Metadata: &pb.PortMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.PortSpecPrivate{
					SubnetId: subnet.Metadata.ResourceId,
				},
			})

			Expect(err).Should(Succeed())
			By("Waiting for port to be created in SDN")
			Eventually(func(g Gomega) {
				got, err := iprmPrivateServiceClient.GetPortPrivate(ctx, &pb.GetPortPrivateRequest{
					Metadata: &pb.PortMetadataReference{
						CloudAccountId: cloudAccountId,
						ResourceId:     portReserved.Metadata.ResourceId,
					},
				})

				g.Expect(err).Should(Succeed())
				g.Expect(got.Status.Phase).Should(Equal(pb.PortPhase_PortPhase_Ready))
				g.Expect(got.Status.Message).Should(Equal("Port ready"))
			}, timeout).Should(Succeed())

		})

		It("Reconcile ReleasePort", func() {
			sdnServiceClient.EXPECT().GetPort(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
			sdnServiceClient.EXPECT().DeletePort(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

			cloudAccountId, err := NewCloudAcctId()
			Expect(err).Should(Succeed())

			vpcName := uuid.NewString()
			_, subnet, err := createVPCAndSubnet(ctx, cloudAccountId, vpcName)

			// Reseave port for the subnet
			portReserved, err := iprmPrivateServiceClient.ReservePort(ctx, &pb.ReservePortRequest{
				Metadata: &pb.PortMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.PortSpecPrivate{
					SubnetId: subnet.Metadata.ResourceId,
				},
			})

			Expect(err).Should(Succeed())
			By("Waiting for port to be created in SDN")
			Eventually(func(g Gomega) {
				got, err := iprmPrivateServiceClient.GetPortPrivate(ctx, &pb.GetPortPrivateRequest{
					Metadata: &pb.PortMetadataReference{
						CloudAccountId: cloudAccountId,
						ResourceId:     portReserved.Metadata.ResourceId,
					},
				})

				g.Expect(err).Should(Succeed())
				g.Expect(got.Status.Phase).Should(Equal(pb.PortPhase_PortPhase_Ready))
				g.Expect(got.Status.Message).Should(Equal("Port ready"))
			}, timeout).Should(Succeed())

			// release port
			_, err = iprmPrivateServiceClient.ReleasePort(ctx, &pb.ReleasePortRequest{
				Metadata: &pb.PortMetadataReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     portReserved.Metadata.ResourceId,
				},
			})
			Expect(err).Should(Succeed())

			By("Waiting for port to be released/removed from SDN")
			Eventually(func(g Gomega) {
				_, err := iprmPrivateServiceClient.GetPortPrivate(ctx, &pb.GetPortPrivateRequest{
					Metadata: &pb.PortMetadataReference{
						CloudAccountId: cloudAccountId,
						ResourceId:     portReserved.Metadata.ResourceId,
					},
				})

				g.Expect(err).ShouldNot(Succeed())
				g.Expect(status.Code(err)).Should(Equal(codes.NotFound)) //
			})

		})
	})
})
