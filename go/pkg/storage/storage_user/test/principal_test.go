package server

import (
	"errors"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	clusterid      = "8623ccaa-704e-4839-bc72-9a89daa20111"
	cloudaccountId = "123456789012"
)

var _ = Describe("FilesystemServiceServer", func() {
	Context("CreatePrincipal", func() {
		It("Should create principal successfully", func() {
			//set mock expectations
			mockS3ServiceClient.EXPECT().CreateS3Principal(gomock.Any(), gomock.Any()).Return(&sc.CreateS3PrincipalResponse{
				S3Principal: &sc.S3Principal{
					Name: "tester",
					Id: &sc.S3PrincipalIdentifier{
						ClusterId: &sc.ClusterIdentifier{
							Uuid: clusterid,
						},
						Id: "principal-id",
					},
				},
			}, nil).Times(1)
			mockS3ServiceClient.EXPECT().UpdateS3PrincipalPolicies(gomock.Any(), gomock.Any()).Return(&sc.UpdateS3PrincipalPoliciesResponse{}, nil).Times(1)
			res, err := fsUser.CreateBucketUser(ctx, createPrincipalReq)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail to create principal", func() {
			By("Throwing error from sds")
			mockS3ServiceClient.EXPECT().CreateS3Principal(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
			_, err := fsUser.CreateBucketUser(ctx, createPrincipalReq)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Update Principal policy", func() {
		It("Should update principal policy successfully", func() {
			//set mock expectations
			mockS3ServiceClient.EXPECT().UpdateS3PrincipalPolicies(gomock.Any(), gomock.Any()).Return(&sc.UpdateS3PrincipalPoliciesResponse{}, nil).Times(1)
			res, err := fsUser.UpdateBucketUserPolicy(ctx, updatePolicyReq)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())

		})
		It("Should fail update policy", func() {
			By("Throwing error from sds")
			mockS3ServiceClient.EXPECT().UpdateS3PrincipalPolicies(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
			_, err := fsUser.UpdateBucketUserPolicy(ctx, updatePolicyReq)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Update Principal credentials", func() {
		It("Should update principal credentials successfully", func() {
			//set mock expectations
			mockS3ServiceClient.EXPECT().SetS3PrincipalCredentials(gomock.Any(), gomock.Any()).Return(&sc.SetS3PrincipalCredentialsResponse{}, nil).Times(1)
			res, err := fsUser.UpdateBucketUserCredentials(ctx, updateCredReq)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail update credentials", func() {
			By("Throwing error from sds")
			mockS3ServiceClient.EXPECT().SetS3PrincipalCredentials(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
			_, err := fsUser.UpdateBucketUserCredentials(ctx, updateCredReq)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Delete Principal", func() {
		It("Should delete principal successfully", func() {
			//set mock expectations
			mockS3ServiceClient.EXPECT().DeleteS3Principal(gomock.Any(), gomock.Any()).Return(&sc.DeleteS3PrincipalResponse{}, nil).Times(1)
			res, err := fsUser.DeleteBucketUser(ctx, &pb.DeleteBucketUserParams{
				CloudAccountId: cloudaccountId,
				ClusterId:      clusterid,
				PrincipalId:    "test-id",
			})
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail delete principal", func() {
			By("Throwing error from sds")
			mockS3ServiceClient.EXPECT().DeleteS3Principal(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
			_, err := fsUser.DeleteBucketUser(ctx, &pb.DeleteBucketUserParams{
				CloudAccountId: cloudaccountId,
				ClusterId:      clusterid,
				PrincipalId:    "test-id",
			})
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Get Bucket Capacity", func() {
		It("Should retrieve bucket capacity successfully", func() {
			//set mock expectations
			mockS3ServiceClient.EXPECT().ListBuckets(gomock.Any(), gomock.Any()).Return(&sc.ListBucketsResponse{
				Buckets: []*sc.Bucket{bucket},
			}, nil).Times(1)
			res, err := fsUser.GetBucketCapacity(ctx, &pb.BucketFilter{ClusterId: clusterid, BucketId: "test-bucket"})
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail get capacity", func() {
			By("Throwing error from sds")
			mockS3ServiceClient.EXPECT().ListBuckets(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
			_, err := fsUser.GetBucketCapacity(ctx, &pb.BucketFilter{ClusterId: clusterid, BucketId: "test-bucket"})
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Ping", func() {
		It("Should be successful", func() {
			_, err := fsUser.PingBucketUserPrivate(ctx, nil)
			Expect(err).To(BeNil())

		})
	})
})
