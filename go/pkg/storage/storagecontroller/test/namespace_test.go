// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"

	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespace", func() {
	var (
		namespace sc.Namespace
		metadata  sc.NamespaceMetadata

		nsResponse *api.Namespace
		client     *sc.StorageControllerClient
		ctrl       *gomock.Controller
		mockClient *mocks.MockNamespaceServiceClient
		ctx        context.Context

		clusterUUID string
		namespaceID string
	)

	BeforeEach(func() {
		// Initialize StorageControllerClient
		ctrl = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockNamespaceServiceClient(ctrl)
		client = &sc.StorageControllerClient{
			NamespaceSvcClient: mockClient,
		}
		clusterUUID = "66efeaca-e493-4a39-b683-15978aac90d5"
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
		// Set up the test input data (payload) before each test
		ctx = context.Background()
		metadata = sc.NamespaceMetadata{
			ClusterId: "your-cluster-id",
			Name:      "test-namespace",
			User:      "test-user",
			Password:  "test-password",
			UUID:      clusterUUID,
		}
		namespace = sc.Namespace{
			Metadata:   metadata,
			Properties: sc.NamespaceProperties{Quota: "2"},
		}

	})

	AfterEach(func() {
		ctrl.Finish()
	})
	Describe("CreateNamespace", func() {
		It("should create a namespace without error", func() {

			mockClient.EXPECT().CreateNamespace(gomock.Any(), gomock.Any()).Return(&api.CreateNamespaceResponse{
				Namespace: nsResponse,
			}, nil).Times(1)

			err := client.CreateNamespace(ctx, namespace)
			Expect(err).To(BeNil())
		})
	})
	Describe("IsNamespaceExists", func() {
		Context("when the namespace exists", func() {
			It("should return true without error", func() {
				request := &api.ListNamespacesRequest{
					ClusterId: &api.ClusterIdentifier{
						Uuid: clusterUUID,
					},
					Filter: &api.ListNamespacesRequest_Filter{
						Names: []string{"test-namespace"},
					},
				}

				// Set up expectations
				mockClient.EXPECT().ListNamespaces(gomock.Any(), request, gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{nsResponse},
				}, nil).Times(1)

				exists, err := client.IsNamespaceExists(ctx, metadata)
				Expect(err).To(BeNil())
				Expect(exists).To(BeTrue())

			})
		})

		Context("when the namespace does not exist", func() {
			It("should return false without error", func() {
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{},
				}, nil).Times(1)

				exists, err := client.IsNamespaceExists(ctx, metadata)
				Expect(err).To(BeNil())
				Expect(exists).To(BeFalse())
			})
		})

		Context("when there is an error", func() {
			It("should return an false", func() {
				// Set up expectations
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)

				exists, err := client.IsNamespaceExists(ctx, metadata)
				Expect(err).NotTo(BeNil())
				Expect(exists).To(BeFalse())
			})
		})
	})
	Describe("GetNamespace", func() {
		Context("When namespace exists", func() {
			It("should return ns without errors", func() {
				request := &api.ListNamespacesRequest{
					ClusterId: &api.ClusterIdentifier{
						Uuid: clusterUUID,
					},
					Filter: &api.ListNamespacesRequest_Filter{
						Names: []string{"test-namespace"},
					},
				}
				mockClient.EXPECT().ListNamespaces(gomock.Any(), request, gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{nsResponse},
				}, nil).Times(1)
				ns, err := client.GetNamespace(ctx, metadata)
				Expect(err).To(BeNil())
				Expect(ns).NotTo(BeNil())

				By("Having innerfunction function return error")
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)

				ns2, err2 := client.GetNamespace(ctx, metadata)
				Expect(err2).NotTo(BeNil())
				Expect(ns2).NotTo(BeNil())
			})
		})
	})

	Describe("ModifyNameSpace", func() {
		Context("When givin valid input for existing ns", func() {
			It("Should modify ns without error", func() {
				request := &api.ListNamespacesRequest{
					ClusterId: &api.ClusterIdentifier{
						Uuid: clusterUUID,
					},
					Filter: &api.ListNamespacesRequest_Filter{
						Names: []string{"test-namespace"},
					},
				}
				mockClient.EXPECT().ListNamespaces(gomock.Any(), request, gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{nsResponse},
				}, nil).Times(1)
				mockClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.UpdateNamespaceResponse{
					Namespace: nsResponse,
				}, nil).Times(1)
				err := client.ModifyNamespace(ctx, namespace, true, "100")
				Expect(err).To(BeNil())
			})
		})
		Context("When error is thrown within function", func() {
			It("Should return error", func() {
				request := &api.ListNamespacesRequest{
					ClusterId: &api.ClusterIdentifier{
						Uuid: clusterUUID,
					},
					Filter: &api.ListNamespacesRequest_Filter{
						Names: []string{"test-namespace"},
					},
				}

				By("Failing in update step")
				mockClient.EXPECT().ListNamespaces(gomock.Any(), request, gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{nsResponse},
				}, nil).Times(1)
				mockClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)

				err := client.ModifyNamespace(ctx, namespace, true, "100")
				Expect(err).NotTo(BeNil())

			})
		})
	})

	Describe("DeleteNamespace", func() {
		It("should delete a namespace without error", func() {
			mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().DeleteNamespace(gomock.Any(), gomock.Any()).Return(&api.DeleteNamespaceResponse{}, nil).Times(1)

			err := client.DeleteNamespace(ctx, metadata)
			Expect(err).To(BeNil())
		})
	})

	Describe("GetAllFileSystemOrgs", func() {
		Context("GetAllFileSystemOrgs", func() {
			It("should succeed", func() {
				By("With namespaces list populated")
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{nsResponse},
				}, nil).Times(1)

				ns, found, err := client.GetAllFileSystemOrgs(ctx, "66efeaca-e493-4a39-b683-15978aac90d5")
				Expect(err).To(BeNil())
				Expect(found).To(BeTrue())
				Expect(ns).NotTo(BeNil())

				By("With empty namespaces list")
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{},
				}, nil).Times(1)

				ns, found, err = client.GetAllFileSystemOrgs(ctx, "66efeaca-e493-4a39-b683-15978aac90d5")
				Expect(err).To(BeNil())
				Expect(found).To(BeFalse())
				Expect(ns).To(BeNil())

			})
			It("should fail", func() {
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
				ns, _, err := client.GetAllFileSystemOrgs(ctx, "66efeaca-e493-4a39-b683-15978aac90d5")
				Expect(err).ToNot(BeNil())
				Expect(ns).To(BeNil())
			})
		})
	})

	Describe("UpdateVastIPFilyers", func() {
		Context("When namespace exists", func() {
			It("should update IP filters without error", func() {
				// Mock ListNamespaces to return the namespace
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{nsResponse},
				}, nil).Times(1)

				// Mock UpdateNamespace to succeed
				mockClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(&api.UpdateNamespaceResponse{
					Namespace: nsResponse,
				}, nil).Times(1)

				err := client.UpdateVastIPFilyers(ctx, namespace)
				Expect(err).To(BeNil())
			})
		})

		Context("When namespace does not exist", func() {
			It("should return nil without error", func() {
				// Mock ListNamespaces to return not found
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{},
				}, nil).Times(1)

				err := client.UpdateVastIPFilyers(ctx, namespace)
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("ModifyVastNamespace", func() {
		Context("When given valid input for increasing size", func() {
			It("should modify namespace without error", func() {
				// Mock UpdateNamespace to succeed
				mockClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(&api.UpdateNamespaceResponse{
					Namespace: nsResponse,
				}, nil).Times(1)

				err := client.ModifyVastNamespace(ctx, namespace, true, "100")
				Expect(err).To(BeNil())
			})
		})

		Context("When given valid input for decreasing size", func() {
			It("should modify namespace without error", func() {
				// Mock UpdateNamespace to succeed
				mockClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(&api.UpdateNamespaceResponse{
					Namespace: nsResponse,
				}, nil).Times(1)

				err := client.ModifyVastNamespace(ctx, namespace, false, "100")
				Expect(err).To(BeNil())
			})
		})

		Context("When there is an error converting existing size to integer", func() {
			It("should return an error", func() {
				// Set an invalid existing size
				namespace.Properties.Quota = "invalid-size"

				err := client.ModifyVastNamespace(ctx, namespace, true, "100")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("When there is an error updating the namespace", func() {
			It("should return an error", func() {
				// Mock UpdateNamespace to return an error
				mockClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)

				err := client.ModifyVastNamespace(ctx, namespace, true, "100")
				Expect(err).NotTo(BeNil())
			})
		})

	})

	Describe("GetVastNamespace", func() {
		Context("When namespace exists", func() {
			It("should return namespace object without error", func() {
				// Mock getNamespaceByName to return the namespace
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{nsResponse},
				}, nil).Times(1)

				ns, err := client.GetVastNamespace(ctx, metadata)
				Expect(err).To(BeNil())
				Expect(ns).NotTo(BeNil())
			})
		})

		Context("When namespace does not exist", func() {
			It("should return empty namespace object without error", func() {
				// Mock getNamespaceByName to return not found
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
					Namespaces: []*api.Namespace{},
				}, nil).Times(1)

				ns, err := client.GetVastNamespace(ctx, metadata)
				Expect(err).To(BeNil())
				Expect(ns).To(Equal(sc.Namespace{}))
			})
		})

		Context("When there is an error finding the namespace", func() {
			It("should return an error", func() {
				// Mock getNamespaceByName to return an error
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)

				ns, err := client.GetVastNamespace(ctx, metadata)
				Expect(err).NotTo(BeNil())
				Expect(ns).To(Equal(sc.Namespace{}))
			})
		})

		Context("When there is an error finding the namespace", func() {
			It("should return an error", func() {
				// Mock ListNamespaces to return an error
				mockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)

				err := client.UpdateVastIPFilyers(ctx, namespace)
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
