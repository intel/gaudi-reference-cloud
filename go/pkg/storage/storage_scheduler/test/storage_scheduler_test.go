// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"

	gomock "github.com/golang/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_scheduler/pkg/server"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	weka "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	scheduler     *server.StorageSchedulerServiceServer
	mockClient    *mocks.MockClusterServiceClient
	mockNsClient  *mocks.MockNamespaceServiceClient
	mockFsClient  *mocks.MockFilesystemServiceClient
	mockKmsClient pb.StorageKMSPrivateServiceClient
	client1       *sc.StorageControllerClient
	req           *pb.FilesystemScheduleRequest
	req2          *pb.FilesystemScheduleRequest
	bkReq         *pb.BucketScheduleRequest
	resp          *api.ListClustersResponse
	nsResp        *api.ListNamespacesResponse
	fsResp        *weka.ListFilesystemsResponse
)

var _ = Describe("FilesystemServiceServer", func() {
	BeforeEach(func() {
		var err error
		// Initialize calling object
		scheduler, err = NewMockStorageSchedulerServer()
		Expect(err).To(BeNil())
		client1 = &sc.StorageControllerClient{
			ClusterSvcClient:        mockClient,
			NamespaceSvcClient:      mockNsClient,
			WekaFilesystemSvcClient: mockFsClient,
		}
		request := &pb.FilesystemCapacity{
			Storage: "5GB",
		}
		spec := &pb.FilesystemSpecPrivate{
			AvailabilityZone: "az1",
			Request:          request,
			StorageClass:     pb.FilesystemStorageClass_GeneralPurpose,
		}

		req = &pb.FilesystemScheduleRequest{
			CloudaccountId: "123456789021",
			//Assignments: resSched,
			RequestSpec: spec,
		}
		req2 = &pb.FilesystemScheduleRequest{
			CloudaccountId: "123456789021",
			Assignments: []*pb.ResourceSchedule{
				{
					ClusterName:    "Cluster1",
					ClusterAddr:    "Location1",
					ClusterUUID:    "BackendUUID1",
					ClusterVersion: "1",
				},
			},
			RequestSpec: spec,
		}
		bkReq = &pb.BucketScheduleRequest{
			RequestSpec: &pb.ObjectBucketSpecPrivate{
				AvailabilityZone: "az1",
				Versioned:        false,
				AccessPolicy:     *pb.BucketAccessPolicy_READ_WRITE.Enum(),
				InstanceType:     "storage-object",
				Request: &pb.StorageCapacityRequest{
					Size: "5GB",
				},
			},
		}

		nsResp = &api.ListNamespacesResponse{
			Namespaces: []*api.Namespace{
				{
					Id: &api.NamespaceIdentifier{
						ClusterId: &api.ClusterIdentifier{
							Uuid: "7773ccaa-704e-4839-bc72-9a89daa20111",
						},
						Id: "123456789012",
					},
					Name: "cluster1",
					Quota: &api.Namespace_Quota{
						TotalBytes: 10000000000,
					},
				},
			},
		}

		fsResp = &weka.ListFilesystemsResponse{
			Filesystems: []*weka.Filesystem{
				{
					Id: &weka.FilesystemIdentifier{
						NamespaceId: &api.NamespaceIdentifier{
							ClusterId: &api.ClusterIdentifier{
								Uuid: "7773ccaa-704e-4839-bc72-9a89daa20111",
							},
							Id: "123456789012",
						},
					},
					Name: "file-1",
					Capacity: &weka.Filesystem_Capacity{
						TotalBytes: 1000000000000,
					},
				},
			},
		}

		resp = &api.ListClustersResponse{
			Clusters: []*api.Cluster{
				{
					Id: &api.ClusterIdentifier{
						Uuid: "8623ccaa-704e-4839-bc72-9a89daa20111",
					},
					Labels: map[string]string{
						"category": "development",
					},
					Name:     "Cluster1",
					Location: "Location1",
					Type:     api.Cluster_TYPE_WEKA,
					Health: &api.Cluster_Health{
						Status: 1,
					},
					Capacity: &api.Cluster_Capacity{
						Storage: &api.Cluster_Capacity_Storage{
							AvailableBytes: 10000000000,
							TotalBytes:     10000000000,
						},
						Namespaces: &api.Cluster_Capacity_Namespaces{
							TotalCount:     256,
							AvailableCount: 50,
						},
					},
				},
				{
					Id: &api.ClusterIdentifier{
						Uuid: "7773ccaa-704e-4839-bc72-9a89daa20111",
					},
					Labels: map[string]string{
						"category": "development",
					},
					Name:     "Cluster2",
					Location: "Location2",
					Type:     api.Cluster_TYPE_MINIO,
					Health: &api.Cluster_Health{
						Status: 1,
					},
					Capacity: &api.Cluster_Capacity{
						Storage: &api.Cluster_Capacity_Storage{
							AvailableBytes: 10000000000,
							TotalBytes:     10000000000,
						},
						Namespaces: &api.Cluster_Capacity_Namespaces{
							TotalCount:     256,
							AvailableCount: 50,
						},
					},
				},
			},
		}

	})

	Context("NewStorageSchedularService when, supplied valid argument", func() {
		It("Should return storageSchedularService", func() {
			//Test Constructor
			sClient := &sc.StorageControllerClient{}
			mockKmsClient := NewMockStorageKMSPrivateServiceClient()

			client, err := server.NewStorageSchedulerService(sClient, mockKmsClient, false)
			Expect(err).To(BeNil())
			Expect(client).NotTo(BeNil())
		})
	})
	Context("NewStorageSchedularService when, invalid argument", func() {
		It("Should return error", func() {
			//Test Constructor
			client2, err2 := server.NewStorageSchedulerService(nil, nil, false)
			Expect(err2).NotTo(BeNil())
			Expect(client2).To(BeNil())
		})
	})
	Context("Schedule", func() {
		It("Should schedule a filesystem successfully", func() {
			By("Providing valid input and recieving no errors from internal function calls")
			mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			out, err := scheduler.ScheduleFile(context.Background(), req)
			Expect(err).To(BeNil())
			Expect(out).NotTo(BeNil())
			mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()
			out4, err4 := scheduler.ScheduleFile(context.Background(), req2)

			Expect(err4).To(BeNil())
			Expect(out4).NotTo(BeNil())

		})
		It("Should fail to schedule a filesystem", func() {
			By("Having an error occur in an internal function call")
			//If - case
			mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			out2, err2 := scheduler.ScheduleFile(context.Background(), req)
			Expect(err2).NotTo(BeNil())
			Expect(out2).To(BeNil())

			//Else - case
			mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()
			out3, err3 := scheduler.ScheduleFile(context.Background(), req2)
			Expect(err3).To(BeNil())
			Expect(out3).NotTo(BeNil())

		})
		It("Should schedule bucket successfully", func() {
			By("Providing valid input and recieving no errors from internal function calls")
			mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			out, err := scheduler.ScheduleBucket(context.Background(), bkReq)
			Expect(err).To(BeNil())
			Expect(out).NotTo(BeNil())
		})
		It("Should fail to schedule a bucket", func() {
			By("Having an error occur in an internal function call")
			//If - case
			mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			out, err := scheduler.ScheduleBucket(context.Background(), bkReq)
			Expect(err).NotTo(BeNil())
			Expect(out).To(BeNil())

			//Else - case
			bkReq.RequestSpec.Request.Size = "10000000"
			out, err = scheduler.ScheduleBucket(context.Background(), bkReq)
			Expect(err).NotTo(BeNil())
			Expect(out).To(BeNil())
		})
	})
	Context("ListFilesystemOrgs", func() {
		It("Should return ns info", func() {
			By("Providing valid input and recieving no errors from internal function calls")
			mockNsClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(nsResp, nil).Times(1)
			client, err := scheduler.ListFilesystemOrgs(context.Background(), &pb.FilesystemOrgsGetRequestPrivate{
				ClusterId: "7773ccaa-704e-4839-bc72-9a89daa20111",
				Prefix:    "",
			})
			Expect(err).To(BeNil())
			Expect(client).NotTo(BeNil())
		})
		It("Should fail to return ns info", func() {
			By("Having an error occur in an internal function call")
			mockNsClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			client, err := scheduler.ListFilesystemOrgs(context.Background(), &pb.FilesystemOrgsGetRequestPrivate{
				ClusterId: "7773ccaa-704e-4839-bc72-9a89daa20111",
				Prefix:    "",
			})
			Expect(err).NotTo(BeNil())
			Expect(client).To(BeNil())
		})
	})
	Context("ListFilesystemInOrgs", func() {
		It("Should return fs info", func() {
			By("Providing valid input and recieving no errors from internal function calls")
			mockNsClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(nsResp, nil).Times(1)
			mockFsClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any()).Return(fsResp, nil).Times(1)
			client, err := scheduler.ListFilesystemInOrgs(context.Background(), &pb.FilesystemInOrgGetRequestPrivate{
				CloudAccountId:     "123456789012",
				Name:               "tester",
				ClusterId:          "7773ccaa-704e-4839-bc72-9a89daa20111",
				NamespaceCredsPath: "/path/to/secret",
			})
			Expect(err).To(BeNil())
			Expect(client).NotTo(BeNil())
		})
		It("Should fail to return ns info", func() {
			By("Having an error occur in an internal function call")
			mockNsClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			client, err := scheduler.ListFilesystemInOrgs(context.Background(), &pb.FilesystemInOrgGetRequestPrivate{
				CloudAccountId:     "123456789012",
				Name:               "tester",
				ClusterId:          "7773ccaa-704e-4839-bc72-9a89daa20111",
				NamespaceCredsPath: "/path/to/secret",
			})
			Expect(err).NotTo(BeNil())
			Expect(client).To(BeNil())
		})
	})
	Context("ListClusters", func() {
		It("Should return cluster info", func() {
			By("Providing valid input and recieving no errors from internal function calls")
			rs := pb.NewMockFilesystemSchedulerPrivateService_ListClustersServer(gomock.NewController(GinkgoT()))
			rs.EXPECT().Context().Return(context.Background()).AnyTimes()
			rs.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
			mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			err := scheduler.ListClusters(&pb.ListClusterRequest{
				Filters: []string{"hello"},
			}, rs)
			Expect(err).To(BeNil())
		})
		It("Should fail to return cluster info", func() {
			By("Having an error occur in an internal function call")
			rs := pb.NewMockFilesystemSchedulerPrivateService_ListClustersServer(gomock.NewController(GinkgoT()))
			rs.EXPECT().Context().Return(context.Background()).AnyTimes()
			rs.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
			mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			err := scheduler.ListClusters(&pb.ListClusterRequest{
				Filters: []string{"hello"},
			}, rs)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("PingPrivate", func() {
		It("Should return no errors", func() {

			client, err := scheduler.PingPrivate(context.Background(), &emptypb.Empty{})
			Expect(err).To(BeNil())
			Expect(client).NotTo(BeNil())
		})
	})
})

func NewMockStorageKMSPrivateServiceClient() pb.StorageKMSPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	kms := pb.NewMockStorageKMSPrivateServiceClient(mockController)
	kmsClient := &pb.GetSecretResponse{
		Secrets: make(map[string]string),
	}
	kmsClient.Secrets["username"] = "user"
	kmsClient.Secrets["password"] = "pass"
	// Mock the Put call for KMSClient
	kms.EXPECT().Get(gomock.Any(), gomock.Any()).Return(kmsClient, nil).AnyTimes()
	kms.EXPECT().Put(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).AnyTimes()

	return kms
}
