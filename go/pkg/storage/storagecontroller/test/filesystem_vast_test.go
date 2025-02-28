// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"

	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	storageControllerApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	storageControllerVastApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/vast"
	mocks "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/vastmocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VastFilesystem", func() {
	var (
		client       *sc.StorageControllerClient
		mockCtrl     *gomock.Controller
		mockClient   *mocks.MockFilesystemServiceClient
		clusterUUID  string
		namespaceID  string
		filesystemID string
	)

	BeforeEach(func() {
		// Initialize StorageControllerClient
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockFilesystemServiceClient(mockCtrl)
		client = &sc.StorageControllerClient{
			VastFilesystemSvcClient: mockClient,
		}
		clusterUUID = "66efeaca-e493-4a39-b683-15978aac90d5"
		namespaceID = "ns_id"
		filesystemID = "fs_id"
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("CreateFilesystem", func() {
		It("should create a fs with provided data", func() {
			params := &sc.CreateFilesystemParams{
				NamespaceID: namespaceID,
				Name:        "testfs",
				Path:        "/testfs",
				TotalBytes:  1000000,
				ClusterID:   clusterUUID,
			}
			fsResponse := &storageControllerVastApi.CreateFilesystemResponse{
				Filesystem: &storageControllerVastApi.Filesystem{
					Id: &storageControllerVastApi.FilesystemIdentifier{
						NamespaceId: &storageControllerApi.NamespaceIdentifier{
							ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: clusterUUID},
							Id:        namespaceID,
						},
						Id: filesystemID,
					},
					Name: "testfs",
				},
			}
			mockClient.EXPECT().CreateFilesystem(gomock.Any(), gomock.Any()).Return(fsResponse, nil).Times(1)

			fs, err := client.CreateVastFilesystem(context.Background(), params)
			Expect(err).To(BeNil())
			Expect(fs).NotTo(BeNil())
			Expect(fs.Name).To(Equal("testfs"))
		})

		It("should return error if creating filesystem fails", func() {
			params := &sc.CreateFilesystemParams{
				NamespaceID: namespaceID,
				Name:        "testfs",
				Path:        "/testfs",
				TotalBytes:  1000000,
				ClusterID:   clusterUUID,
			}
			mockClient.EXPECT().CreateFilesystem(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)

			fs, err := client.CreateVastFilesystem(context.Background(), params)
			Expect(err).To(HaveOccurred())
			Expect(fs).To(BeNil())
		})
	})

	Context("GetFilesystem", func() {
		It("should get a filesystem with provided data", func() {
			params := &sc.GetFilesystemParams{
				NamespaceID:  namespaceID,
				FilesystemID: filesystemID,
				ClusterID:    clusterUUID,
			}
			fsResponse := &storageControllerVastApi.GetFilesystemResponse{
				Filesystem: &storageControllerVastApi.Filesystem{
					Id: &storageControllerVastApi.FilesystemIdentifier{
						NamespaceId: &storageControllerApi.NamespaceIdentifier{
							ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: clusterUUID},
							Id:        namespaceID,
						},
						Id: filesystemID,
					},
					Name: "testfs",
				},
			}
			mockClient.EXPECT().GetFilesystem(gomock.Any(), gomock.Any()).Return(fsResponse, nil).Times(1)

			fs, err := client.GetVastFilesystem(context.Background(), params)
			Expect(err).To(BeNil())
			Expect(fs).NotTo(BeNil())
			Expect(fs.Name).To(Equal("testfs"))
		})

		It("should return error if getting filesystem fails", func() {
			params := &sc.GetFilesystemParams{
				NamespaceID:  namespaceID,
				FilesystemID: filesystemID,
				ClusterID:    clusterUUID,
			}
			mockClient.EXPECT().GetFilesystem(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)

			fs, err := client.GetVastFilesystem(context.Background(), params)
			Expect(err).To(HaveOccurred())
			Expect(fs).To(BeNil())
		})
	})

	Context("ListFilesystems", func() {
		It("should list filesystems with provided data", func() {
			params := &sc.ListFilesystemsParams{
				NamespaceID: namespaceID,
				Names:       []string{"testfs"},
				ClusterID:   clusterUUID,
			}
			fsResponse := &storageControllerVastApi.ListFilesystemsResponse{
				Filesystems: []*storageControllerVastApi.Filesystem{
					{
						Id: &storageControllerVastApi.FilesystemIdentifier{
							NamespaceId: &storageControllerApi.NamespaceIdentifier{
								ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: clusterUUID},
								Id:        namespaceID,
							},
							Id: filesystemID,
						},
						Name: "testfs",
					},
				},
			}
			mockClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any()).Return(fsResponse, nil).Times(1)

			fsList, err := client.ListVastFilesystems(context.Background(), params)
			Expect(err).To(BeNil())
			Expect(fsList).NotTo(BeNil())
			Expect(fsList).To(HaveLen(1))
			Expect(fsList[0].Name).To(Equal("testfs"))
		})

		It("should return error if listing filesystems fails", func() {
			params := &sc.ListFilesystemsParams{
				NamespaceID: namespaceID,
				Names:       []string{"testfs"},
				ClusterID:   clusterUUID,
			}
			mockClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)

			fsList, err := client.ListVastFilesystems(context.Background(), params)
			Expect(err).To(HaveOccurred())
			Expect(fsList).To(BeNil())
		})
	})

	Context("UpdateFilesystem", func() {
		It("should update a filesystem with provided data", func() {
			params := &sc.UpdateFilesystemParams{
				NamespaceID:   namespaceID,
				FilesystemID:  filesystemID,
				NewTotalBytes: 2000000,
				ClusterID:     clusterUUID,
			}
			fsResponse := &storageControllerVastApi.UpdateFilesystemResponse{
				Filesystem: &storageControllerVastApi.Filesystem{
					Id: &storageControllerVastApi.FilesystemIdentifier{
						NamespaceId: &storageControllerApi.NamespaceIdentifier{
							ClusterId: &storageControllerApi.ClusterIdentifier{Uuid: clusterUUID},
							Id:        namespaceID,
						},
						Id: filesystemID,
					},
					Name: "testfs",
				},
			}
			mockClient.EXPECT().UpdateFilesystem(gomock.Any(), gomock.Any()).Return(fsResponse, nil).Times(1)

			fs, err := client.UpdateVastFilesystem(context.Background(), params)
			Expect(err).To(BeNil())
			Expect(fs).NotTo(BeNil())
			Expect(fs.Name).To(Equal("testfs"))
		})

		It("should return error if updating filesystem fails", func() {
			params := &sc.UpdateFilesystemParams{
				NamespaceID:   namespaceID,
				FilesystemID:  filesystemID,
				NewTotalBytes: 2000000,
				ClusterID:     clusterUUID,
			}
			mockClient.EXPECT().UpdateFilesystem(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)

			fs, err := client.UpdateVastFilesystem(context.Background(), params)
			Expect(err).To(HaveOccurred())
			Expect(fs).To(BeNil())
		})
	})

	Context("DeleteFilesystem", func() {
		It("should delete a filesystem with provided data", func() {
			params := &sc.DeleteFilesystemParams{
				NamespaceID:  namespaceID,
				FilesystemID: filesystemID,
				ClusterID:    clusterUUID,
			}
			mockClient.EXPECT().DeleteFilesystem(gomock.Any(), gomock.Any()).Return(&storageControllerVastApi.DeleteFilesystemResponse{}, nil).Times(1)

			err := client.DeleteVastFilesystem(context.Background(), params)
			Expect(err).To(BeNil())
		})

		It("should return error if deleting filesystem fails", func() {
			params := &sc.DeleteFilesystemParams{
				NamespaceID:  namespaceID,
				FilesystemID: filesystemID,
				ClusterID:    clusterUUID,
			}
			mockClient.EXPECT().DeleteFilesystem(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)

			err := client.DeleteVastFilesystem(context.Background(), params)
			Expect(err).To(HaveOccurred())
		})
	})
})
