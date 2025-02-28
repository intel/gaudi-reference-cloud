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

var _ = Describe("Lifecyclerule", func() {
	var (
		client      *sc.StorageControllerClient
		ctrl        *gomock.Controller
		mockClient  *mocks.MockS3ServiceClient
		ctx         context.Context
		lcr         *api.LifecycleRule
		lc          sc.LifecycleRule
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

		lc = sc.LifecycleRule{
			ClusterId: clusterUUID,
			BucketId:  "bucket-test1",
		}
		lcr = &api.LifecycleRule{
			Id: &api.LifecycleRuleIdentifier{
				Id: "test-rule",
			},
			Prefix:               "",
			ExpireDays:           5,
			NoncurrentExpireDays: 2,
			DeleteMarker:         true,
		}

	}) //BeforEach
	AfterEach(func() {
		ctrl.Finish()
	}) //AfterEach

	Context("Create Bucket LifecycleRule", func() {
		It("Should create bucket lifecycle rule succesfully", func() {
			res := []*api.LifecycleRule{}
			res = append(res, lcr)
			// set mock expectation
			mockClient.EXPECT().CreateLifecycleRules(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.CreateLifecycleRulesResponse{LifecycleRules: res}, nil).Times(1)
			//make request with valid payload
			rule, err := client.CreateBucketLifecycleRules(ctx, lc)
			//Should not have error
			Expect(err).To(BeNil())
			Expect(rule).NotTo(BeNil())
		})
		It("Should return error", func() {
			By("providing invalid input")
			req := sc.LifecycleRule{ClusterId: "", BucketId: ""}
			//missing clusterId
			_, err := client.CreateBucketLifecycleRules(ctx, req)
			Expect(err).NotTo(BeNil())
			//missing bucketId
			req.ClusterId = clusterUUID
			_, err = client.CreateBucketLifecycleRules(ctx, req)
			Expect(err).NotTo(BeNil())

			By("failing grpc call")
			//set mock behavior to return error
			mockClient.EXPECT().CreateLifecycleRules(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
			//make request with valid payload
			_, err = client.CreateBucketLifecycleRules(ctx, lc)
			//Should have error
			Expect(err).NotTo(BeNil())
		})
	}) //Context

}) //Describe
