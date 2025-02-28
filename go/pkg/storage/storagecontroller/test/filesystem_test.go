// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"

	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filesystem", func() {
	var (
		fsMetadata   sc.FilesystemMetadata
		fsProperties sc.FilesystemProperties
		fs           sc.Filesystem

		client       *sc.StorageControllerClient
		fsResponse   *weka.Filesystem
		nsResponse   *api.Namespace
		mockCtrl     *gomock.Controller
		mockClient   *mocks.MockFilesystemServiceClient
		nsMockClient *mocks.MockNamespaceServiceClient

		clusterUUID  string
		filesystemID string
		namespaceID  string
	)

	BeforeEach(func() {
		// Initialize StorageControllerClient
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockFilesystemServiceClient(mockCtrl)
		nsMockClient = mocks.NewMockNamespaceServiceClient(mockCtrl)
		client = &sc.StorageControllerClient{
			WekaFilesystemSvcClient: mockClient,
			NamespaceSvcClient:      nsMockClient,
		}
		clusterUUID = "66efeaca-e493-4a39-b683-15978aac90d5"
		filesystemID = "fs_id"
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

		fsResponse = &weka.Filesystem{
			Id: &weka.FilesystemIdentifier{
				NamespaceId: &api.NamespaceIdentifier{
					ClusterId: &api.ClusterIdentifier{
						Uuid: clusterUUID,
					},
					Id: namespaceID,
				},
				Id: filesystemID,
			},
			Name:         "testfs",
			Status:       weka.Filesystem_STATUS_READY,
			IsEncrypted:  true,
			AuthRequired: true,
			Capacity: &weka.Filesystem_Capacity{
				TotalBytes:     20000000,
				AvailableBytes: 100000,
			},
		}
		// Set up the test input data (payload) before each test
		fsMetadata = sc.FilesystemMetadata{
			FileSystemName: "testfs",
			Encrypted:      true,
			AuthRequired:   true,
			User:           "testuser",
			Password:       "testpassword",
			NamespaceName:  "testnamespace",
			UUID:           "918b5026-d516-48c8-bfd3-5998547265b2",
		}

		fsProperties = sc.FilesystemProperties{
			FileSystemCapacity: "10",
		}

		fs = sc.Filesystem{
			Metadata:   fsMetadata,
			Properties: fsProperties,
		}
	})
	AfterEach(func() {
		mockCtrl.Finish()
	})
	Context("CreateFilesystem", func() {
		It("should create a fs with provided data", func() {
			// Set expectations on the mock
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().CreateFilesystem(gomock.Any(), gomock.Any()).Return(&weka.CreateFilesystemResponse{
				Filesystem: fsResponse,
			}, nil).Times(1)

			fs, err := client.CreateFilesystem(context.Background(), fs)
			Expect(err).To(BeNil())
			Expect(fs).NotTo(BeNil())
		})
		It("should return error if fs exists", func() {
			// Set expectations on the mock
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().CreateFilesystem(gomock.Any(), gomock.Any()).Return(&weka.CreateFilesystemResponse{
				Filesystem: fsResponse,
			}, errors.New("error")).Times(1)

			_, err := client.CreateFilesystem(context.Background(), fs)
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Get", func() {
		It("should get a fs", func() {
			//TODO: GetFilesystem func not yet implemented
			fs, err := client.GetFilesystem(context.Background(), fsMetadata)
			Expect(err).Should(BeNil())
			Expect(fs).ShouldNot(BeNil())
		})
		It("should return err if fs not found", func() {
			//TODO: GetFilesystem func not yet implemented
			metadata := sc.FilesystemMetadata{
				FileSystemName: "none",
				Encrypted:      true,
				AuthRequired:   true,
				User:           "testuser",
				Password:       "testpassword",
				NamespaceName:  "testnamespace",
			}
			fs, err := client.GetFilesystem(context.Background(), metadata)
			Expect(fs).ShouldNot(BeNil())
			Expect(err).Should(BeNil())
		})
	})
	Context("Exist", func() {
		It("should return true if the filesystem exists", func() {
			// Prepare test data and expectations
			queryParams := sc.FilesystemMetadata{
				FileSystemName: "testfs",
				Encrypted:      true,
				AuthRequired:   true,
				User:           "test-user",
				Password:       "test-password",
				NamespaceName:  "test-namespace",
				UUID:           clusterUUID,
			}
			// Set expectations on the mock
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any()).Return(&weka.ListFilesystemsResponse{
				Filesystems: []*weka.Filesystem{fsResponse},
			}, nil).Times(1)

			exists, err := client.IsFilesystemExists(context.Background(), queryParams, false)

			// Verify the result
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
		It("should return false if the filesystem does not exists", func() {
			// Prepare test data and expectations
			queryParams := sc.FilesystemMetadata{
				FileSystemName: "test-fs2",
				Encrypted:      true,
				AuthRequired:   true,
				User:           "test-user2",
				Password:       "test-password",
				NamespaceName:  "test-namespace2",
				UUID:           "918b5026-d516-48c8-bfd3-5998547265b2",
			}
			// Set expectations on the mock
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any()).Return(&weka.ListFilesystemsResponse{
				Filesystems: []*weka.Filesystem{},
			}, nil).Times(1)

			exists, err := client.IsFilesystemExists(context.Background(), queryParams, false)

			// Verify the result
			Expect(err).To(BeNil())
			Expect(exists).To(BeFalse())
		})
	})
	Context("Delete", func() {
		It("should delete a fs", func() {
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any()).Return(&weka.ListFilesystemsResponse{
				Filesystems: []*weka.Filesystem{fsResponse},
			}, nil).Times(1)
			mockClient.EXPECT().DeleteFilesystem(gomock.Any(), gomock.Any()).Return(&weka.DeleteFilesystemResponse{}, nil).Times(1)
			err := client.DeleteFilesystem(context.Background(), fsMetadata)
			Expect(err).Should(BeNil())
		})
	})
	Context("UpdateFilesystem", func() {
		It("Should succeed", func() {
			// Set expectations on the mock
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)

			mockClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any(), gomock.Any()).Return(&weka.ListFilesystemsResponse{
				Filesystems: []*weka.Filesystem{fsResponse},
			}, nil).Times(1)

			mockClient.EXPECT().UpdateFilesystem(gomock.Any(), gomock.Any()).Return(&weka.UpdateFilesystemResponse{
				Filesystem: fsResponse,
			}, nil).Times(1)

			fsRes, err := client.UpdateFilesystem(context.Background(), fs)
			Expect(err).To(BeNil())
			Expect(fsRes).NotTo(BeNil())
		})
		It("Should fail", func() {
			By("ListNamespaces returns error")
			// Set expectations on the mock
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, errors.New("error")).Times(1)

			fsRes, err := client.UpdateFilesystem(context.Background(), fs)
			Expect(err).ToNot(BeNil())
			Expect(fsRes).ToNot(BeNil())

			By("ListFilesystems returns error")
			// Set expectations on the mock
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)

			mockClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any(), gomock.Any()).Return(&weka.ListFilesystemsResponse{
				Filesystems: []*weka.Filesystem{fsResponse},
			}, errors.New("error")).Times(1)

			fsRes, err = client.UpdateFilesystem(context.Background(), fs)
			Expect(err).ToNot(BeNil())
			Expect(fsRes).ToNot(BeNil())

			By("ListFilesystems returns empty result")
			// Set expectations on the mock
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)

			mockClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any(), gomock.Any()).Return(&weka.ListFilesystemsResponse{
				Filesystems: []*weka.Filesystem{},
			}, nil).Times(1)

			fsRes, err = client.UpdateFilesystem(context.Background(), fs)
			Expect(err).To(BeNil())
			Expect(fsRes).ToNot(BeNil())

			By("WEKA UpdateFilesystem returns error")
			// Set expectations on the mock
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(1)

			mockClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any(), gomock.Any()).Return(&weka.ListFilesystemsResponse{
				Filesystems: []*weka.Filesystem{fsResponse},
			}, nil).Times(1)

			mockClient.EXPECT().UpdateFilesystem(gomock.Any(), gomock.Any()).Return(&weka.UpdateFilesystemResponse{
				Filesystem: fsResponse,
			}, errors.New("error")).Times(1)

			fsRes, err = client.UpdateFilesystem(context.Background(), fs)
			Expect(err).ToNot(BeNil())
			Expect(fsRes).ToNot(BeNil())
		})
	})
})
