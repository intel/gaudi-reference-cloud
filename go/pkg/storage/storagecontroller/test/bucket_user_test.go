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
		user        *api.S3Principal
		usrReq      sc.ObjectUserRequest
		usrData     sc.ObjectUserData
		updateReq   sc.ObjectUserUpdateRequest
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
		policies := []*api.S3Principal_Policy{}
		policy := &api.S3Principal_Policy{
			BucketId: &api.BucketIdentifier{
				ClusterId: &api.ClusterIdentifier{
					Uuid: clusterUUID,
				},
				Id: "bucket-test1",
			},
			Prefix: "",
			Read:   true,
			Write:  true,
			Delete: true,
			//Actions: []*api.S3Principal_Policy_BucketActions{},
		}
		policies = append(policies, policy)

		usrReq = sc.ObjectUserRequest{
			ClusterId: clusterUUID,
			UserName:  "test-user",
			Password:  "password",
			Policies:  policies,
		}
		usrData = sc.ObjectUserData{
			ClusterId:   clusterUUID,
			PrincipalId: "test-user",
		}
		updateReq = sc.ObjectUserUpdateRequest{
			ClusterId:   clusterUUID,
			PrincipalId: "bucket-test1",
			UserName:    "test-user",
			Password:    "password",
			Policies:    policies,
		}
		user = &api.S3Principal{
			Id: &api.S3PrincipalIdentifier{
				ClusterId: &api.ClusterIdentifier{
					Uuid: clusterUUID,
				},
				Id: "test-user",
			},
			Name:     "test-user",
			Policies: policies,
		}

	}) //BeforEach
	AfterEach(func() {
		ctrl.Finish()
	}) //AfterEach
	Describe("Bucket User", func() {
		Context("Create Bucket User", func() {
			It("Should create user succesfully", func() {
				// set mock expectation
				mockClient.EXPECT().CreateS3Principal(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.CreateS3PrincipalResponse{S3Principal: user}, nil).Times(1)
				mockClient.EXPECT().UpdateS3PrincipalPolicies(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.UpdateS3PrincipalPoliciesResponse{S3Principal: user}, nil).Times(1)
				// make function call
				usr, err := client.CreateObjectUser(ctx, usrReq)
				Expect(err).To(BeNil())
				Expect(usr).NotTo(BeNil())
			})
			It("Should return error", func() {
				By("providing invalid input")
				req := sc.ObjectUserRequest{ClusterId: "", UserName: "", Password: ""}
				//missing clusterId
				_, err := client.CreateObjectUser(ctx, req)
				Expect(err).NotTo(BeNil())
				//missing username
				req.ClusterId = clusterUUID
				_, err = client.CreateObjectUser(ctx, req)
				Expect(err).NotTo(BeNil())
				//missing password
				req.UserName = "username"
				_, err = client.CreateObjectUser(ctx, req)
				Expect(err).NotTo(BeNil())
				//missing username
				req.Password = "password"
				_, err = client.CreateObjectUser(ctx, req)
				Expect(err).NotTo(BeNil())

				By("failing grpc function call")
				//set mock function to return error
				mockClient.EXPECT().CreateS3Principal(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
				//make function call
				_, err = client.CreateObjectUser(ctx, usrReq)
				Expect(err).NotTo(BeNil())
			})
		}) //context
		Context("Delete Bucket User", func() {
			It("Should delete user successfully", func() {
				//set mock behavior
				mockClient.EXPECT().DeleteS3Principal(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.DeleteS3PrincipalResponse{}, nil).Times(1)
				//make function call with valid input
				err := client.DeleteObjectUser(ctx, usrData)
				Expect(err).To(BeNil())
			})
			It("Should return error", func() {
				By("providing invalid input")
				req := sc.ObjectUserData{ClusterId: "", PrincipalId: ""}
				//missing clusterId
				err := client.DeleteObjectUser(ctx, req)
				Expect(err).NotTo(BeNil())
				//missing principalId
				req.ClusterId = clusterUUID
				err = client.DeleteObjectUser(ctx, req)
				Expect(err).NotTo(BeNil())
				req.PrincipalId = "test-user"

				By("failing grpc call")
				mockClient.EXPECT().DeleteS3Principal(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
				//make function call
				err = client.DeleteObjectUser(ctx, req)
				Expect(err).NotTo(BeNil())
			})
		})

		Context("Get Bucket User", func() {
			It("Should get a user successfully", func() {
				//set mock expectation
				mockClient.EXPECT().GetS3Principal(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.GetS3PrincipalResponse{S3Principal: user}, nil).Times(1)
				//make the function call
				resp, err := client.GetObjectUser(ctx, usrData)
				Expect(err).To(BeNil())
				Expect(resp).NotTo(BeNil())
			})
			It("Should return error", func() {
				By("providing invalid input")
				req := sc.ObjectUserData{ClusterId: "", PrincipalId: ""}
				//missing clusterId
				_, err := client.GetObjectUser(ctx, req)
				Expect(err).NotTo(BeNil())
				//missing principalId
				req.ClusterId = clusterUUID
				_, err = client.GetObjectUser(ctx, req)
				Expect(err).NotTo(BeNil())
				By("failing grpc call")
				mockClient.EXPECT().GetS3Principal(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
				//make the function call
				_, err = client.GetObjectUser(ctx, usrData)
				Expect(err).NotTo(BeNil())
			})
		})

		Context("Update Object User Policy", func() {
			It("Should update user policy successfully", func() {
				//set mock expectation
				mockClient.EXPECT().UpdateS3PrincipalPolicies(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.UpdateS3PrincipalPoliciesResponse{S3Principal: user}, nil).Times(1)
				//make function call
				err := client.UpdateObjectUserPolicy(ctx, updateReq)
				Expect(err).To(BeNil())
			})
			It("Should return error", func() {
				By("providing invalid input")
				req := sc.ObjectUserUpdateRequest{ClusterId: "", PrincipalId: ""}
				//missing clusterId
				err := client.UpdateObjectUserPolicy(ctx, req)
				Expect(err).NotTo(BeNil())
				//missing principalId
				req.ClusterId = clusterUUID
				err = client.UpdateObjectUserPolicy(ctx, req)
				Expect(err).NotTo(BeNil())
				//missing policies
				req.PrincipalId = "test-user"
				err = client.UpdateObjectUserPolicy(ctx, req)
				Expect(err).NotTo(BeNil())

				By("failing grpc call")
				mockClient.EXPECT().UpdateS3PrincipalPolicies(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
				//make function call
				err = client.UpdateObjectUserPolicy(ctx, updateReq)
				Expect(err).NotTo(BeNil())
			})
		})

		Context("Update Object User Password", func() {
			It("Should update user password successfully", func() {
				//set mock expectation
				mockClient.EXPECT().SetS3PrincipalCredentials(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.SetS3PrincipalCredentialsResponse{}, nil).Times(1)
				//make function call
				err := client.UpdateObjectUserPass(ctx, updateReq)
				Expect(err).To(BeNil())
			})
			It("Should return error", func() {
				By("providing invalid input")
				req := sc.ObjectUserUpdateRequest{ClusterId: "", PrincipalId: "", Password: ""}
				//missing clusterId
				err := client.UpdateObjectUserPass(ctx, req)
				Expect(err).NotTo(BeNil())
				//missing principalId
				req.ClusterId = clusterUUID
				err = client.UpdateObjectUserPass(ctx, req)
				Expect(err).NotTo(BeNil())
				//missing password
				req.PrincipalId = "test-user"
				err = client.UpdateObjectUserPass(ctx, req)
				Expect(err).NotTo(BeNil())

				By("failing grpc call")
				mockClient.EXPECT().SetS3PrincipalCredentials(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error")).Times(1)
				//make function call
				err = client.UpdateObjectUserPass(ctx, updateReq)
				Expect(err).NotTo(BeNil())
			})
		})

	}) //Describe
}) //Cluster
