// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestYourFunction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Test Suite")
}

var (
	mockUserService                   *UserService
	mockBucketAPIClient               pb.ObjectStorageServicePrivateClient
	strclient                         sc.StorageControllerClient
	mockS3ServiceClient               *mocks.MockS3ServiceClient
	mockCtrl                          *gomock.Controller
	objectStorageServicePrivateClient *pb.MockObjectStorageServicePrivateClient
)

func NewMockObjectStorageServicePrivateClient() pb.ObjectStorageServicePrivateClient {
	mockController := gomock.NewController(GinkgoT())
	objectStorageServicePrivateClient = pb.NewMockObjectStorageServicePrivateClient(mockController)

	objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	objectStorageServicePrivateClient.EXPECT().RemoveBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	return objectStorageServicePrivateClient
}

var _ = Describe("Principle Update Scheduler", func() {
	Context("getSuccessStatus", func() {
		It("Should succeed", func() {
			rv := getSuccessStatus(pb.BucketSubnetEventStatus_E_ADDING)
			Expect(rv).To(Equal(pb.BucketSubnetEventStatus_E_ADDED))

			rv = getSuccessStatus(pb.BucketSubnetEventStatus_E_DELETING)
			Expect(rv).To(Equal(pb.BucketSubnetEventStatus_E_DELETED))

			rv = getSuccessStatus(pb.BucketSubnetEventStatus_E_UNSPECIFIED)
			Expect(rv).To(Equal(pb.BucketSubnetEventStatus_E_UNSPECIFIED))
		})
	})

	Context("getUpdatedPolicies", func() {
		It("Should succeed", func() {
			Spec := []*pb.ObjectUserPermissionSpec{
				{
					BucketId:   "example-bucket-id",
					Prefix:     "example-prefix",
					Permission: []pb.BucketPermission{pb.BucketPermission_ReadBucket, pb.BucketPermission_WriteBucket},                     // Assuming BucketPermissionRead and BucketPermissionWrite are valid enum values.
					Actions:    []pb.ObjectBucketActions{pb.ObjectBucketActions_GetBucketLocation, pb.ObjectBucketActions_GetBucketPolicy}, // Assuming ObjectBucketActionsCreate and ObjectBucketActionsDelete are valid enum values.
				},
				// Add more instances as needed.
			}

			updatedVnet := &pb.VNetPrivate{
				Metadata: &pb.VNetPrivate_Metadata{
					CloudAccountId: "example-cloud-account-id",
					Name:           "example-vnet-name",
					ResourceId:     "example-resource-id",
				},
				Spec: &pb.VNetSpecPrivate{
					Region:           "example-region",
					AvailabilityZone: "example-availability-zone",
					Subnet:           "example-subnet",
					PrefixLength:     24, // Example prefix length.
					Gateway:          "example-gateway",
					VlanId:           123, // Example VLAN ID.
					VlanDomain:       "example-vlan-domain",
					AddressSpace:     "example-address-space",
				},
			}

			rv := getUpdatedPolicies(Spec, updatedVnet, pb.BucketSubnetEventStatus_E_ADDING, "clusterid")
			Expect(rv).NotTo(BeNil())

			rv = getUpdatedPolicies(Spec, updatedVnet, pb.BucketSubnetEventStatus_E_DELETING, "clusterid")
			Expect(rv).NotTo(BeNil())
		})
	})

	Context("HandlePrincipalSecurityGroupUpdate", func() {
		It("Should pass", func() {
			mockBucketAPIClient = NewMockObjectStorageServicePrivateClient()

			mockCtrl = gomock.NewController(GinkgoT())
			mockS3ServiceClient = mocks.NewMockS3ServiceClient(mockCtrl)

			strclient = sc.StorageControllerClient{
				S3ServiceClient: mockS3ServiceClient,
			}

			mockUserService = &UserService{
				strCntClient:    strclient,
				bucketAPIClient: mockBucketAPIClient,
			}

			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			mockStream := pb.NewMockObjectStorageServicePrivate_GetBucketSubnetEventClient(mockCtrl)
			mockStream.EXPECT().Context().Return(context.Background()).AnyTimes()

			vnet := pb.VNetPrivate{
				Metadata: &pb.VNetPrivate_Metadata{
					CloudAccountId: "example-cloud-account-id",
					Name:           "example-vnet-name",
					ResourceId:     "example-resource-id",
				},
				Spec: &pb.VNetSpecPrivate{
					Region:           "example-region",
					AvailabilityZone: "example-availability-zone",
					Subnet:           "example-subnet",
					PrefixLength:     24, // Example prefix length.
					Gateway:          "example-gateway",
					VlanId:           123, // Example VLAN ID.
					VlanDomain:       "example-vlan-domain",
					AddressSpace:     "example-address-space",
				},
			}

			bpSpec := []*pb.ObjectUserPermissionSpec{
				{
					BucketId:   "example-bucket-id",
					Prefix:     "example-prefix",
					Permission: []pb.BucketPermission{pb.BucketPermission_ReadBucket, pb.BucketPermission_WriteBucket},                     // Assuming BucketPermissionRead and BucketPermissionWrite are valid enum values.
					Actions:    []pb.ObjectBucketActions{pb.ObjectBucketActions_GetBucketLocation, pb.ObjectBucketActions_GetBucketPolicy}, // Assuming ObjectBucketActionsCreate and ObjectBucketActionsDelete are valid enum values.
				},
				// Add more instances as needed.
			}

			bucketprincipals := []*pb.BucketPrincipal{
				{
					ClusterId:      "example-cluster-id",
					PrincipalId:    "example-principal-id",
					AccessEndpoint: "example-access-endpoint",
					ClusterName:    "example-cluster-name",
					Spec:           bpSpec,
				},
			}

			eventType := pb.BucketSubnetEventStatus_E_ADDING

			bucketsubnetupdateevent := pb.BucketSubnetUpdateEvent{
				EventType:  eventType,
				Vnet:       &vnet,
				Principals: bucketprincipals,
			}

			gomock.InOrder(
				// First value returned by Recv
				mockStream.EXPECT().Recv().Return(&bucketsubnetupdateevent, errors.New("some error")),

				// // Second value returned by Recv
				// mockStream.EXPECT().Recv().Return(nil, io.EOF), // Indicate the end of the stream
			)
			mockS3ServiceClient.EXPECT().UpdateS3PrincipalPolicies(gomock.Any(), gomock.Any()).Return(&api.UpdateS3PrincipalPoliciesResponse{
				S3Principal: &api.S3Principal{
					Id: &api.S3PrincipalIdentifier{
						ClusterId: &api.ClusterIdentifier{
							Uuid: "8623ccaa-704e-4839-bc72-9a89daa20111",
						},
						Id: "8623ccaa-704e-4839-bc72-9a89daa20111",
					},
					Name: "user",
				},
			}, nil).AnyTimes()
			objectStorageServicePrivateClient.EXPECT().GetBucketSubnetEvent(gomock.Any(), gomock.Any()).Return(mockStream, nil).AnyTimes()
			objectStorageServicePrivateClient.EXPECT().UpdateBucketSubnetStatus(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
			err2 := mockUserService.HandlePrincipalSecurityGroupUpdate(context.Background())
			Expect(err2).To(BeNil())
		})

		It("Should fail", func() {
			By("GetBucketSubnetEvent returns error")
			mockBucketAPIClient = NewMockObjectStorageServicePrivateClient()
			mockCtrl = gomock.NewController(GinkgoT())
			mockS3ServiceClient = mocks.NewMockS3ServiceClient(mockCtrl)

			strclient = sc.StorageControllerClient{
				S3ServiceClient: mockS3ServiceClient,
			}

			mockUserService = &UserService{
				strCntClient:    strclient,
				bucketAPIClient: mockBucketAPIClient,
			}

			objectStorageServicePrivateClient.EXPECT().GetBucketSubnetEvent(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).AnyTimes()
			err2 := mockUserService.HandlePrincipalSecurityGroupUpdate(context.Background())
			Expect(err2).ToNot(BeNil())
		})
	})
})
