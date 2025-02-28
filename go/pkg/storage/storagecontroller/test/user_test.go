// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"github.com/golang/mock/gomock"

	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User", func() {
	var (
		user     sc.User
		metadata sc.UserMetadata

		userResponse *api.User
		nsResponse   *api.Namespace

		client       *sc.StorageControllerClient
		ctrl         *gomock.Controller
		mockClient   *mocks.MockUserServiceClient
		nsMockClient *mocks.MockNamespaceServiceClient
		ctx          context.Context

		clusterUUID string
		userID      string
		namespaceID string
	)
	BeforeEach(func() {
		// Initialize StorageControllerClient
		ctrl = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockUserServiceClient(ctrl)
		nsMockClient = mocks.NewMockNamespaceServiceClient(ctrl)

		client = &sc.StorageControllerClient{
			UserSvcClient:      mockClient,
			NamespaceSvcClient: nsMockClient,
		}

		clusterUUID = "918b5026-d516-48c8-bfd3-5998547265b2"
		userID = "user_id"
		namespaceID = "ns_id"

		nsResponse = &api.Namespace{
			Id: &api.NamespaceIdentifier{
				ClusterId: &api.ClusterIdentifier{
					Uuid: clusterUUID,
				},
				Id: namespaceID,
			},
			Name: "test-namespace",
			Quota: &api.Namespace_Quota{
				TotalBytes: 50000000,
			},
		}

		userResponse = &api.User{
			Id: &api.UserIdentifier{
				NamespaceId: &api.NamespaceIdentifier{
					ClusterId: &api.ClusterIdentifier{
						Uuid: clusterUUID,
					},
					Id: namespaceID,
				},
				Id: userID,
			},
			Name: "test-user",
			Role: api.User_ROLE_ADMIN,
		}

		// Set up the test input data (payload) before each test
		ctx = context.Background()
		metadata = sc.UserMetadata{
			Role:              "default",
			NamespaceUser:     "test-user",
			NamespacePassword: "test-pass",
			NamespaceName:     "test",
			UUID:              clusterUUID,
		}
		user = sc.User{
			Metadata: metadata,
			Properties: sc.UserProperties{
				NewUser:         "user",
				NewUserPassword: "pass",
			},
		}

	}) //BeforEach
	AfterEach(func() {
		ctrl.Finish()
	}) //AfterEach
	Describe("CreateUser", func() {
		It("should create a user without error", func() {
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(&api.CreateUserResponse{
				User: userResponse,
			}, nil).Times(1)

			err := client.CreateUser(ctx, user)
			Expect(err).To(BeNil())
		})
	})
	Describe("IsUserExists", func() {
		Context("when the user exists", func() {
			It("should return true without error", func() {
				// Set up expectations
				nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{nsResponse},
				}, nil).Times(1)
				mockClient.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Return(&api.ListUsersResponse{
					Users: []*api.User{userResponse},
				}, nil).Times(1)

				exists, err := client.IsUserExists(ctx, user)
				Expect(err).To(BeNil())
				Expect(exists).To(BeTrue())

			})
		})
		Context("when the user does not exist", func() {
			It("should return false without error", func() {
				var metadata2 = sc.UserMetadata{
					Role:              "default",
					NamespaceUser:     "test-user2",
					NamespacePassword: "test-pass2",
					NamespaceName:     "test2",
				}
				var user2 = sc.User{
					Metadata: metadata2,
					Properties: sc.UserProperties{
						NewUser:         "user2",
						NewUserPassword: "pass2",
					},
				}
				// Set up expectations
				nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{nsResponse},
				}, nil).Times(1)
				mockClient.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Return(&api.ListUsersResponse{
					Users: []*api.User{},
				}, nil).Times(1)

				exists, err := client.IsUserExists(ctx, user2)
				Expect(err).To(BeNil())
				Expect(exists).To(BeFalse())
			})
		})

	}) //IsExist
	Describe("DeleteUser", func() {
		It("should delete a User without error", func() {
			request := &api.DeleteUserRequest{
				UserId: &api.UserIdentifier{
					NamespaceId: &api.NamespaceIdentifier{
						ClusterId: &api.ClusterIdentifier{
							Uuid: clusterUUID,
						},
						Id: namespaceID,
					},
					Id: userID,
				},
				AuthCtx: &api.AuthenticationContext{
					Scheme: &api.AuthenticationContext_Basic_{
						Basic: &api.AuthenticationContext_Basic{
							Principal:   "test-user",
							Credentials: "test-pass",
						},
					},
				},
			}
			// Set up expectations
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().ListUsers(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListUsersResponse{
				Users: []*api.User{userResponse},
			}, nil).Times(1)
			mockClient.EXPECT().DeleteUser(gomock.Any(), request, gomock.Any()).Return(&api.DeleteUserResponse{}, nil).Times(1)
			var queryParams = sc.DeletUserData{
				NamespaceUser:     metadata.NamespaceUser,
				NamespacePassword: metadata.NamespacePassword,
				NamespaceName:     metadata.NamespaceName,
				UsertoBeDeleted:   metadata.NamespaceUser,
				UUID:              metadata.UUID,
			}
			err := client.DeleteUser(ctx, queryParams)
			Expect(err).To(BeNil())
		})
	})

	Describe("Update User Password", func() {
		It("Should Update User Password successfully", func() {
			req := &api.UpdateUserPasswordRequest{
				UserId: &api.UserIdentifier{
					NamespaceId: &api.NamespaceIdentifier{
						ClusterId: &api.ClusterIdentifier{
							Uuid: clusterUUID,
						},
						Id: namespaceID,
					},
					Id: userID,
				},
				NewPassword: "abc123",
				AuthCtx: &api.AuthenticationContext{
					Scheme: &api.AuthenticationContext_Basic_{
						Basic: &api.AuthenticationContext_Basic{
							Principal:   "test-user",
							Credentials: "test-pass",
						},
					},
				},
			}
			Expect(req).ToNot(BeNil())

			var queryParams = sc.UpdateUserpassword{
				NamespaceUser:     metadata.NamespaceUser,
				NamespacePassword: metadata.NamespacePassword,
				NamespaceName:     metadata.NamespaceName,
				UsertoBeUpdated:   metadata.NamespaceUser,
				NewPassword:       "abc123",
				UUID:              metadata.UUID,
			}
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().ListUsers(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListUsersResponse{
				Users: []*api.User{userResponse},
			}, nil).Times(1)
			mockClient.EXPECT().UpdateUserPassword(gomock.Any(), req, gomock.Any()).Return(&api.UpdateUserPasswordResponse{}, nil).Times(1)

			err := client.UpdateUserPassword(ctx, queryParams)
			Expect(err).To(BeNil())

		})
	})

	Describe("CreateVastUser", func() {
		Context("When user creation is successful", func() {
			It("should create a user without error", func() {
				// Mock CreateUser to succeed
				mockClient.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

				_, err := client.CreateVastUser(ctx, user)
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("DeleteVastUser", func() {
		Context("When user deletion is successful", func() {
			It("should delete the user without error", func() {
				// Mock DeleteUser to succeed
				mockClient.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

				queryParams := sc.DeleteUserData{
					ClusterUUID: clusterUUID,
					NamespaceID: namespaceID,
					UserID:      userID,
				}

				err := client.DeleteVastUser(ctx, queryParams)
				Expect(err).To(BeNil())
			})
		})

	})

}) //User
