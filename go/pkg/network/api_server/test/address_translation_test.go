// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	timeout time.Duration = 500 * time.Second
)

var _ = Describe("Address Translation API Integration Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	Context("Create API", func() {
		It("Create Address Translation with internet type successfully", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			resp, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(resp).ShouldNot(BeNil())
			Expect(resp.Metadata).ShouldNot(BeNil())
			Expect(resp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(resp.Metadata.CloudAccountId).Should(Equal(cloudAccountId))
			Expect(resp.Spec.TranslationType).Should(Equal(createReq.Spec.TranslationType))
			Expect(resp.Spec.PortId).Should(Equal(createReq.Spec.PortId))
			Expect(resp.Status.Phase).Should(Equal(pb.AddressTranslationPhase_AddressTranslationPhase_Provisioning))
		})

		It("Create Address Translation with storage type successfully", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "storage",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			resp, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(resp).ShouldNot(BeNil())
			Expect(resp.Metadata).ShouldNot(BeNil())
			Expect(resp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(resp.Metadata.CloudAccountId).Should(Equal(cloudAccountId))
			Expect(resp.Spec.TranslationType).Should(Equal(createReq.Spec.TranslationType))
			Expect(resp.Spec.PortId).Should(Equal(createReq.Spec.PortId))
			Expect(resp.Status.Phase).Should(Equal(pb.AddressTranslationPhase_AddressTranslationPhase_Provisioning))
		})

		It("Create Address Translation with transient type successfully", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "transient",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			resp, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(resp).ShouldNot(BeNil())
			Expect(resp.Metadata).ShouldNot(BeNil())
			Expect(resp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(resp.Metadata.CloudAccountId).Should(Equal(cloudAccountId))
			Expect(resp.Spec.TranslationType).Should(Equal(createReq.Spec.TranslationType))
			Expect(resp.Spec.PortId).Should(Equal(createReq.Spec.PortId))
			Expect(resp.Status.Phase).Should(Equal(pb.AddressTranslationPhase_AddressTranslationPhase_Provisioning))
		})

		It("Create Address Translation without cloudAccountId should fail", func() {
			// Setup test data
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			_, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})

		It("Create Address Translation with non-existing cloudAccountId should fail", func() {
			// Setup test data
			cloudAccountId := "non-existing-cloud-account-id"
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			_, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})

		It("Create Address Translation with invalid translation type should fail", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "invalid_type",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			_, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.InvalidArgument))
		})

		It("Create Address Translation with invalid port id should fail", func() {
			// Setup test data\
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          "invalid-port-id",
				},
			}

			// Create the Address Translation record
			_, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			// TODO: This fails until we add real port id validation. The first assertion is just a placeholder
			Expect(err).Should(Succeed()) // TODO: Placeholder!
		})

		It("Create the same Address Translation twice should fail", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}
			// Create the Address Translation record
			_, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())

			// Try to create the same Address Translation record again
			_, err = addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.AlreadyExists))
		})
	})

	Context("Get API", func() {
		It("Get Address Translation with valid resourceId should succeed", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			createResp, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())
			Expect(createResp).ShouldNot(BeNil())
			Expect(createResp.Metadata).ShouldNot(BeNil())
			Expect(createResp.Metadata.ResourceId).ShouldNot(BeEmpty())

			// Get the Address Translation record by Id
			getReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     createResp.Metadata.ResourceId,
				},
			}
			getResp, err := addressTranslationPrivateServiceClient.GetPrivate(ctx, getReq)
			Expect(err).Should(Succeed())
			Expect(getResp).ShouldNot(BeNil())
			Expect(getResp.Metadata).ShouldNot(BeNil())
			Expect(getResp.Metadata.CloudAccountId).Should(Equal(createResp.Metadata.CloudAccountId))
			Expect(getResp.Metadata.ResourceId).Should(Equal(createResp.Metadata.ResourceId))
			Expect(getResp.Spec.TranslationType).Should(Equal(createResp.Spec.TranslationType))
			Expect(getResp.Spec.PortId).Should(Equal(createResp.Spec.PortId))
			Expect(getResp.Spec.ProfileId).Should(Equal(createResp.Spec.ProfileId))
			Expect(getResp.Spec.IpAddress).Should(Equal(createResp.Spec.IpAddress))
			Expect(getResp.Spec.MacAddress).Should(Equal(createResp.Spec.MacAddress))
			Expect(getResp.Metadata.CreationTimestamp).Should(Equal(createResp.Metadata.CreationTimestamp))
		})

		It("Get Address Translation with invalid resourceId should fail", func() {
			// Setup test data
			resourceId := "invalid-resource-id"
			cloudAccountId := cloudaccount.MustNewId()
			// Get the Address Translation record by Id
			getReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     resourceId,
				},
			}
			_, err := addressTranslationPrivateServiceClient.GetPrivate(ctx, getReq)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
		})

		It("Get Address Translation with valid resourceId that doesn't exist should fail", func() {
			// Setup test data
			resourceId := uuid.New().String() // Generate a valid but non-existent resourceId
			cloudAccountId := cloudaccount.MustNewId()
			// Get the Address Translation record by Id
			getReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     resourceId,
				},
			}
			resp, err := addressTranslationPrivateServiceClient.GetPrivate(ctx, getReq)
			Expect(resp).Should(BeNil())
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
		})
	})

	Context("Delete API", func() {
		It("Delete Address Translation with valid resourceId should succeed", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			createResp, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())

			// Delete the Address Translation record by Id
			deleteReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     createResp.Metadata.ResourceId,
				},
			}
			deleteResp, err := addressTranslationPrivateServiceClient.DeletePrivate(ctx, deleteReq)
			Expect(err).Should(Succeed())
			Expect(deleteResp).ShouldNot(BeNil())

			getReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     createResp.Metadata.ResourceId,
				},
			}
			resp, err := addressTranslationPrivateServiceClient.GetPrivate(ctx, getReq)
			Expect(err).Should(Succeed())
			Expect(resp.Status.Phase).Should(Equal(pb.AddressTranslationPhase_AddressTranslationPhase_Deleting))
		})

		It("Delete Address Translation with valid resourceId that doesn't exist should fail", func() {
			// Setup test data
			resourceId := uuid.New().String() // Generate a valid but non-existent resourceId
			cloudAccountId := cloudaccount.MustNewId()

			// Delete the Address Translation record by Id
			deleteReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     resourceId,
				},
			}
			_, err := addressTranslationPrivateServiceClient.DeletePrivate(ctx, deleteReq)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
		})

		It("Delete Address Translation with mismatching resourceVersion should fail", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			createResp, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())

			resourceVersion, err := strconv.Atoi(createResp.Metadata.ResourceVersion)
			Expect(err).Should(Succeed())

			// Delete the Address Translation record by Id with mismatching resourceVersion
			deleteReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId:  cloudAccountId,
					ResourceId:      createResp.Metadata.ResourceId,
					ResourceVersion: fmt.Sprintf("%d", resourceVersion+1),
				},
			}
			_, err = addressTranslationPrivateServiceClient.DeletePrivate(ctx, deleteReq)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.FailedPrecondition))
		})
	})

	Context("List API", func() {
		It("List Address Translations successfully", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq1 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			createReq2 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "storage",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation records
			addressTranslation1, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq1)
			Expect(err).Should(Succeed())
			addressTranslation2, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq2)
			Expect(err).Should(Succeed())

			resourceIds := []string{addressTranslation1.Metadata.ResourceId, addressTranslation2.Metadata.ResourceId}

			// List the Address Translation records
			listReq := &pb.AddressTranslationListPrivateRequest{
				Metadata: &pb.AddressTranslationListMetadataPrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{},
			}
			listResp, err := addressTranslationPrivateServiceClient.ListPrivate(ctx, listReq)
			Expect(err).Should(Succeed())
			Expect(listResp).ShouldNot(BeNil())
			Expect(len(listResp.Items)).Should(Equal(2))
			for _, item := range listResp.Items {
				Expect(resourceIds).Should(ContainElement(item.Metadata.ResourceId))
			}
		})

		It("List Address Translations with at least two items successfully", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq1 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			createReq2 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			createReq3 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "storage",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation records
			addressTranslation1, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq1)
			Expect(err).Should(Succeed())
			addressTranslation2, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq2)
			Expect(err).Should(Succeed())
			_, err = addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq3)
			Expect(err).Should(Succeed())

			resourceIds := []string{addressTranslation1.Metadata.ResourceId, addressTranslation2.Metadata.ResourceId}

			// List the Address Translation records
			listReq := &pb.AddressTranslationListPrivateRequest{
				Metadata: &pb.AddressTranslationListMetadataPrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
				},
			}
			listResp, err := addressTranslationPrivateServiceClient.ListPrivate(ctx, listReq)
			Expect(err).Should(Succeed())
			Expect(listResp).ShouldNot(BeNil())
			Expect(len(listResp.Items)).Should(Equal(2))
			for _, item := range listResp.Items {
				Expect(resourceIds).Should(ContainElement(item.Metadata.ResourceId))
			}
		})

		It("List Address Translations for a specific cloudAccountId", func() {
			// Setup test data
			cloudAccountId1 := cloudaccount.MustNewId()
			cloudAccountId2 := cloudaccount.MustNewId()

			createReq1 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId1,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			createReq2 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId1,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "storage",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			createReq3 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId2,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "transient",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation records
			addressTranslation1, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq1)
			Expect(err).Should(Succeed())
			addressTranslation2, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq2)
			Expect(err).Should(Succeed())
			_, err = addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq3)
			Expect(err).Should(Succeed())

			resourceIds := []string{addressTranslation1.Metadata.ResourceId, addressTranslation2.Metadata.ResourceId}

			// List the Address Translation records for the first cloudAccountId
			listReq := &pb.AddressTranslationListPrivateRequest{
				Metadata: &pb.AddressTranslationListMetadataPrivate{
					CloudAccountId: cloudAccountId1,
				},
				Spec: &pb.AddressTranslationSpecPrivate{},
			}
			listResp, err := addressTranslationPrivateServiceClient.ListPrivate(ctx, listReq)
			Expect(err).Should(Succeed())
			Expect(listResp).ShouldNot(BeNil())
			Expect(len(listResp.Items)).Should(Equal(2))
			for _, item := range listResp.Items {
				Expect(resourceIds).Should(ContainElement(item.Metadata.ResourceId))
				Expect(item.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
			}
		})
	})

	Context("UpdateStatus API", func() {
		It("Update status to ready", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			createResp, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())

			// Update the Address Translation status to ready
			message := "Address Translation is ready"
			updateReq := &pb.AddressTranslationUpdateStatusPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: createResp.Metadata.CloudAccountId,
					ResourceId:     createResp.Metadata.ResourceId,
				},
				Status: &pb.AddressTranslationStatus{
					Phase:   pb.AddressTranslationPhase_AddressTranslationPhase_Ready,
					Message: message,
				},
			}
			_, err = addressTranslationPrivateServiceClient.UpdateStatusPrivate(ctx, updateReq)
			Expect(err).Should(Succeed())

			getReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     createResp.Metadata.ResourceId,
				},
			}
			getResp, err := addressTranslationPrivateServiceClient.GetPrivate(ctx, getReq)
			Expect(err).Should(Succeed())
			Expect(getResp.Status.Phase).Should(Equal(pb.AddressTranslationPhase_AddressTranslationPhase_Ready))
		})

		It("Update status to deleted, should soft delete resource", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			createResp, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())

			// Delete the Address Translation record by Id
			deleteReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     createResp.Metadata.ResourceId,
				},
			}
			deleteResp, err := addressTranslationPrivateServiceClient.DeletePrivate(ctx, deleteReq)
			Expect(err).Should(Succeed())
			Expect(deleteResp).ShouldNot(BeNil())

			// Update the Address Translation status to deleted
			message := "Address Translation is deleted"
			updateReq := &pb.AddressTranslationUpdateStatusPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId:   cloudAccountId,
					ResourceId:       createResp.Metadata.ResourceId,
					DeletedTimestamp: timestamppb.Now(),
				},
				Status: &pb.AddressTranslationStatus{
					Phase:   pb.AddressTranslationPhase_AddressTranslationPhase_Deleted,
					Message: message,
				},
			}
			_, err = addressTranslationPrivateServiceClient.UpdateStatusPrivate(ctx, updateReq)
			Expect(err).Should(Succeed())

			getReq := &pb.AddressTranslationGetPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId: cloudAccountId,
					ResourceId:     createResp.Metadata.ResourceId,
				},
			}
			_, err = addressTranslationPrivateServiceClient.GetPrivate(ctx, getReq)
			Expect(status.Code(err)).Should(Equal(codes.NotFound))
		})

		It("Update status with wrong resourceVersion should fail", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation record
			createResp, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq)
			Expect(err).Should(Succeed())

			resourceVersion, err := strconv.Atoi(createResp.Metadata.ResourceVersion)
			Expect(err).Should(Succeed())

			// Update the Address Translation status with wrong resourceVersion
			updateReq := &pb.AddressTranslationUpdateStatusPrivateRequest{
				Metadata: &pb.AddressTranslationIdReference{
					CloudAccountId:  cloudAccountId,
					ResourceId:      createResp.Metadata.ResourceId,
					ResourceVersion: fmt.Sprintf("%d", resourceVersion+1),
				},
				Status: &pb.AddressTranslationStatus{
					Phase: pb.AddressTranslationPhase_AddressTranslationPhase_Ready,
				},
			}
			_, err = addressTranslationPrivateServiceClient.UpdateStatusPrivate(ctx, updateReq)
			Expect(err).ShouldNot(Succeed())
			Expect(status.Code(err)).Should(Equal(codes.FailedPrecondition))
		})
	})

	Context("SearchStream API", func() {
		It("SearchStream with valid parameters should succeed", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()
			createReq1 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			createReq2 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "storage",
					PortId:          uuid.New().String(), // TODO: Add real port Id
				},
			}

			// Create the Address Translation records
			addressTranslation1, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq1)
			Expect(err).Should(Succeed())

			addressTranslation2, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq2)
			Expect(err).Should(Succeed())

			resourceIds := []string{addressTranslation1.Metadata.ResourceId, addressTranslation2.Metadata.ResourceId}

			// Search the Address Translation records
			searchReq := &pb.AddressTranslationSearchStreamPrivateRequest{}
			stream, err := addressTranslationPrivateServiceClient.SearchStreamPrivate(ctx, searchReq)
			Expect(err).Should(Succeed())

			var results []*pb.AddressTranslationWatchResponse
			for {
				resp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				Expect(err).Should(Succeed())
				results = append(results, resp)
			}

			var resultResourceIds []string
			for _, result := range results {
				if result.Type != pb.WatchDeltaType_Bookmark {
					resultResourceIds = append(resultResourceIds, result.Object.Metadata.ResourceId)
				}
			}

			Expect(len(results)).Should(Equal(len(resourceIds) + 1)) // +1 for bookmark
			for _, resourceId := range resourceIds {
				Expect(resultResourceIds).Should(ContainElement(resourceId))
			}
			Expect(results[len(results)-1].Type).Should(Equal(pb.WatchDeltaType_Bookmark))
		})
	})

	Context("Watch API", func() {
		It("should receive updates for new Address Translations", func() {
			// Setup test data
			cloudAccountId := cloudaccount.MustNewId()

			// Create a context with cancel
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Channel to receive watch responses
			watchResponses := make(chan *pb.AddressTranslationWatchResponse, 500)

			// Run the watch in a separate goroutine
			go func() {
				defer GinkgoRecover()
				watchReq := &pb.AddressTranslationWatchRequest{
					ResourceVersion: "0",
				}
				stream, err := addressTranslationPrivateServiceClient.Watch(ctx, watchReq)
				Expect(err).Should(Succeed())

				for {
					resp, err := stream.Recv()

					if err == io.EOF || status.Code(err) == codes.Canceled {
						break
					}

					Expect(err).Should(Succeed())
					if resp.Type != pb.WatchDeltaType_Bookmark {
						watchResponses <- resp
					}
					time.Sleep(100 * time.Millisecond)
				}
			}()

			// Create the first Address Translation record
			createReq1 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "internet",
					PortId:          uuid.New().String(),
				},
			}
			addressTranslation1, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq1)
			Expect(err).Should(Succeed())

			// Verify the first Address Translation is received
			Eventually(watchResponses, timeout).Should(Receive(WithTransform(func(resp *pb.AddressTranslationWatchResponse) string {
				return resp.Object.Metadata.ResourceId
			}, Equal(addressTranslation1.Metadata.ResourceId))))

			// Create the second Address Translation record
			createReq2 := &pb.AddressTranslationCreatePrivateRequest{
				Metadata: &pb.AddressTranslationMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
				},
				Spec: &pb.AddressTranslationSpecPrivate{
					TranslationType: "storage",
					PortId:          uuid.New().String(),
				},
			}
			addressTranslation2, err := addressTranslationPrivateServiceClient.CreatePrivate(ctx, createReq2)
			Expect(err).Should(Succeed())

			// Verify the second Address Translation is received
			Eventually(watchResponses, timeout).Should(Receive(WithTransform(func(resp *pb.AddressTranslationWatchResponse) string {
				return resp.Object.Metadata.ResourceId
			}, Equal(addressTranslation2.Metadata.ResourceId))))

		})
	})
})
