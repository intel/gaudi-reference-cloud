// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	var (
		client      *sc.StorageControllerClient
		ctrl        *gomock.Controller
		mockClient  *mocks.MockS3ServiceClient
		bucket      *api.Bucket
		bktMeta     sc.BucketMetadata
		bktSpec     sc.BucketSpec
		bktPolicy   sc.BucketPolicy
		ctx         context.Context
		clusterUUID string
	)
	BeforeEach(func() {
		// Initialize StorageControllerClient
		ctrl = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockS3ServiceClient(ctrl)
		client = &sc.StorageControllerClient{
			S3ServiceClient: mockClient,
		}
		// Set up the test input data (payload) before each test
		ctx = context.Background()
		clusterUUID = "66efeaca-e493-4a39-b683-15978aac90d5"
		bucket = &api.Bucket{
			Id: &api.BucketIdentifier{
				ClusterId: &api.ClusterIdentifier{
					Uuid: clusterUUID,
				},
			},
			Name:      "bucket-test1",
			Versioned: true,
			Capacity: &api.Bucket_Capacity{
				TotalBytes:     10000000000,
				AvailableBytes: 10000000000,
			},
			EndpointUrl: "https://test-endpoint-url",
		}
		bktMeta = sc.BucketMetadata{
			Name:      "bucket-test1",
			BucketId:  "s3kd34n4e239ds0",
			ClusterId: clusterUUID,
		}
		bktSpec = sc.BucketSpec{
			AccessPolicy:   sc.BucketAccessPolicyReadWrite,
			Versioned:      true,
			Totalbytes:     10000000000,
			AvailableBytes: 10000000000,
		}
		bktPolicy = sc.BucketPolicy{
			ClusterId: clusterUUID,
			BucketId:  "bucket-test1",
			Policy:    sc.BucketAccessPolicyReadOnly,
		}

	}) //BeforEach
	AfterEach(func() {
		ctrl.Finish()
	}) //AfterEach

	Context("Create Bucket", func() {
		It("Should create bucket succesfully", func() {
			// set mock expectation
			mockClient.EXPECT().CreateBucket(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.CreateBucketResponse{
				Bucket: bucket,
			}, nil).Times(1)
			//make request with valid payload
			bkt, err := client.CreateBucket(ctx, sc.Bucket{
				Metadata: bktMeta,
				Spec:     bktSpec,
			})
			//Should not have error
			Expect(err).To(BeNil())
			Expect(bkt).NotTo(BeNil())
		})
		It("Should return error", func() {
			By("providing invalid input")
			//missing bucket name in metadata
			_, err := client.CreateBucket(ctx, sc.Bucket{
				Metadata: sc.BucketMetadata{
					BucketId:  "test",
					ClusterId: "99928378372",
				},
			})
			Expect(err).NotTo(BeNil())
			//clusterId is empty
			_, err = client.CreateBucket(ctx, sc.Bucket{
				Metadata: sc.BucketMetadata{
					BucketId:  "test",
					Name:      "bucket-test1",
					ClusterId: "",
				},
			})
			Expect(err).NotTo(BeNil())
			By("failing grpc call")
			//set mock behavior to return error
			mockClient.EXPECT().CreateBucket(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
			//make request with valid payload
			bkt, err := client.CreateBucket(ctx, sc.Bucket{
				Metadata: bktMeta,
				Spec:     bktSpec,
			})
			//Should have error
			Expect(err).NotTo(BeNil())
			Expect(bkt).To(BeNil())
		})
	}) //Context
	Context("Delete Bucket", func() {
		It("Should delete bucket successfully", func() {
			//set mock behavior
			mockClient.EXPECT().DeleteBucket(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.DeleteBucketResponse{}, nil).Times(1)
			//make request with valid payload
			err := client.DeleteBucket(ctx, bktMeta, true)
			Expect(err).To(BeNil())
		})
		It("Should return error", func() {
			By("providing invalid input")
			//missing bucketId
			req := sc.BucketMetadata{BucketId: "", ClusterId: ""}
			err := client.DeleteBucket(ctx, req, true)
			Expect(err).NotTo(BeNil())
			req.BucketId = "bucket-test1"
			//missing ClusterId
			err = client.DeleteBucket(ctx, req, true)
			Expect(err).NotTo(BeNil())

			By("failing grpc call")
			//set mock behavior to return error
			mockClient.EXPECT().DeleteBucket(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
			//make request with valid inputs
			err = client.DeleteBucket(ctx, bktMeta, true)
			Expect(err).NotTo(BeNil())
		})
	}) //context

	Context("GetBucketPolicy", func() {
		It("Should get bucket policy successfully", func() {
			//set mock behavior
			mockClient.EXPECT().GetBucketPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.GetBucketPolicyResponse{
				BucketId: &api.BucketIdentifier{
					ClusterId: &api.ClusterIdentifier{
						Uuid: "9919293993312",
					},
					Id: "bucket-test1",
				},
				Policy: api.BucketAccessPolicy_BUCKET_ACCESS_POLICY_READ_WRITE,
			}, nil).Times(1)

			//make function call
			policy, err := client.GetBucketPolicy(ctx, bktMeta)
			Expect(err).To(BeNil())
			Expect(policy).NotTo(BeNil())
		})
		It("Should return error", func() {
			By("supplying invalid input")
			req := sc.BucketMetadata{BucketId: "", ClusterId: ""}
			//missing bucketId
			_, err := client.GetBucketPolicy(ctx, req)
			Expect(err).NotTo(BeNil())
			//missing clusterId
			req.BucketId = "bucket-test1"
			_, err = client.GetBucketPolicy(ctx, req)
			Expect(err).NotTo(BeNil())

			By("failing grpc call")
			//set mock function to return error
			mockClient.EXPECT().GetBucketPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
			//make request with valid input
			_, err = client.GetBucketPolicy(ctx, bktMeta)
			Expect(err).NotTo(BeNil())
		})
	}) // context
	Context("UpdateBucketPolicy", func() {
		It("Should update bucket policy successfully", func() {
			//set mock behavior
			mockClient.EXPECT().UpdateBucketPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.UpdateBucketPolicyResponse{}, nil)
			//make function call
			err := client.UpdateBucketPolicy(ctx, bktPolicy)
			Expect(err).To(BeNil())
		})
		It("Should return error", func() {
			By("supplying invalid input")
			req := sc.BucketPolicy{BucketId: "", ClusterId: ""}
			//missing bucketId
			err := client.UpdateBucketPolicy(ctx, req)
			Expect(err).NotTo(BeNil())
			req.BucketId = "bucket-test1"
			//missing clusterId
			err = client.UpdateBucketPolicy(ctx, req)
			Expect(err).NotTo(BeNil())

			By("failing grpc call")
			//set mock function to return error
			mockClient.EXPECT().UpdateBucketPolicy(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
			//make function call
			err = client.UpdateBucketPolicy(ctx, bktPolicy)
			Expect(err).NotTo(BeNil())
		})
	})

}) //Describe
