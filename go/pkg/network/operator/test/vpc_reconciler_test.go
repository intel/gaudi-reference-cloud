// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	timeout time.Duration = 10 * time.Second
	maxId   int64         = 1_000_000_000_000
)

var _ = Describe("Create vpc, happy path", func() {
	ctx := context.Background()
	defer GinkgoRecover()

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	Context("VPC integration tests", func() {
		It("Reconciles a new vpc", func() {

			cloudaccountId, err := NewCloudAcctId()
			Expect(err).Should(Succeed())
			vpcName := uuid.NewString()
			resp, err := vpcServiceClient.Create(ctx, &pb.VPCCreateRequest{
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
						NameOrId:       &pb.VPCMetadataReference_ResourceId{ResourceId: resp.Metadata.ResourceId},
					},
				})
				g.Expect(err).Should(Succeed())
				g.Expect(got.Spec.CidrBlock).Should(Equal("10.0.0.0/16"))
				g.Expect(got.Status.Phase).Should(Equal(pb.VPCPhase_VPCPhase_Ready))
				g.Expect(got.Status.Message).Should(Equal("VPC ready"))
			}, timeout).Should(Succeed())
		})
	})
})
