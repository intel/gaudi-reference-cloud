package test

import (
	"context"
	"database/sql"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	server "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/api_server/pkg/server"
	adminServer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_admin_api_server/pkg/server"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"

	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	ctx             context.Context
	testDb          *manageddb.TestDb
	managedDb       *manageddb.ManagedDb
	sqlDb           *sql.DB
	storageAdminSvc *adminServer.StorageAdminServiceClient
)

func TestStorageAdminService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Admin Service Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()

	grpcutil.UseTLS = false
	log.SetDefaultLogger()

	By("Starting database")
	testDb = &manageddb.TestDb{}
	var err error
	managedDb, err = testDb.Start(ctx)
	// Check Db setup succeeds
	Expect(err).Should(Succeed())
	Expect(managedDb).ShouldNot(BeNil())
	Expect(managedDb.Migrate(ctx, db.Fsys, "migrations")).Should(Succeed())
	sqlDb, err = managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	Expect(sqlDb).ShouldNot(BeNil())

	// Set up mock cloudaccount
	cloudAccountClient := NewMockCloudAccountServiceClient()

	// Set up quota
	cloudaccounts := make(map[string]server.LaunchQuota)
	storageQuota := map[string]int{
		"filesystems":     1,
		"totalSizeGB":     1000,
		"vm-icp-gaudi2-1": 8,
		"buckets":         1,
	}
	cloudaccounts["STANDARD"] = server.LaunchQuota{
		StorageQuota: storageQuota,
	}
	quota := server.CloudAccountQuota{
		CloudAccounts: cloudaccounts,
	}

	qmsClient := NewMockQuotaManagementPrivateServiceClient()
	qService := &server.QuotaService{}
	err = qService.Init(ctx, sqlDb, quota, qmsClient)
	Expect(err).Should(Succeed())

	// Set up mock filesystem service client
	filesystemServiceClient := NewMockFilesystemServiceClient()

	// Set up mock storage controller service client
	strCntClient := NewMockStorageControllerClient()
	maxCapacity := int64(200) // Replace with the appropriate value
	selectedRegion := "us-dev-1"
	maxVolumesAllowed := int64(10)
	maxBucketsAllowed := int64(10)

	// Initialize StorageAdminServiceClient
	storageAdminSvc, err = adminServer.NewStorageAdminServiceClient(ctx, sqlDb, cloudAccountClient, qService, filesystemServiceClient, strCntClient, maxCapacity, selectedRegion, maxVolumesAllowed, maxBucketsAllowed)
	Expect(err).Should(Succeed())
	Expect(storageAdminSvc).ShouldNot(BeNil())
})

// Helper functions to mock productClient
func NewMockProductCatalogServiceClient() pb.ProductCatalogServiceClient {
	mockController := gomock.NewController(GinkgoT())
	product := pb.NewMockProductCatalogServiceClient(mockController)

	mapMetadata := make(map[string]string)
	mapMetadata["volume.size.min"] = "1"
	mapMetadata["volume.size.max"] = "1"

	// define product
	p1 := &pb.Product{
		Name:     "storage-file",
		Metadata: mapMetadata,
	}

	// Set mocking behavior
	product.EXPECT().AdminRead(gomock.Any(), gomock.Any()).Return(&pb.ProductResponse{
		Products: []*pb.Product{
			p1,
		},
	}, nil).AnyTimes()
	return product
}

func NewMockCloudAccountServiceClient() pb.CloudAccountServiceClient {
	mockController := gomock.NewController(GinkgoT())
	cloudAccountClient := pb.NewMockCloudAccountServiceClient(mockController)

	// create mock cloudaccount
	cloudAccount := &pb.CloudAccount{
		Type: pb.AccountType_ACCOUNT_TYPE_STANDARD,
		Id:   "123456789012",
		Name: "123456789012",
	}

	cloudAccountClient.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(cloudAccount, nil).AnyTimes()
	return cloudAccountClient
}

func NewMockStorageControllerClient() *sc.StorageControllerClient {
	mockCtrl := gomock.NewController(GinkgoT())
	mockS3Client := mocks.NewMockS3ServiceClient(mockCtrl)
	mockClusterSvcClient := mocks.NewMockClusterServiceClient(mockCtrl)

	//Set mock expectations
	listBucketResponse := &stcnt_api.ListBucketsResponse{
		Buckets: []*stcnt_api.Bucket{
			{
				Id: &stcnt_api.BucketIdentifier{
					ClusterId: &stcnt_api.ClusterIdentifier{
						Uuid: "8623ccaa-704e-4839-bc72-9a89daa20111",
					},
					Id: "123456789012-1",
				},
				Name: "123456789012-1",
				Capacity: &stcnt_api.Bucket_Capacity{
					TotalBytes:     50000000000,
					AvailableBytes: 40000000000,
				},
				EndpointUrl: "https://pdx05-minio-dev-2.us-staging-1.cloud.intel.com:9000",
			},
			{

				Id: &stcnt_api.BucketIdentifier{
					ClusterId: &stcnt_api.ClusterIdentifier{
						Uuid: "8623ccaa-704e-4839-bc72-9a89daa20111",
					},
					Id: "123456789012-2",
				},
				Name: "test-bucket-2",
				Capacity: &stcnt_api.Bucket_Capacity{
					TotalBytes:     50000000000,
					AvailableBytes: 40000000000,
				},
				EndpointUrl: "https://pdx05-minio-dev-2.us-staging-1.cloud.intel.com:9000",
			},
			{

				Id: &stcnt_api.BucketIdentifier{
					ClusterId: &stcnt_api.ClusterIdentifier{
						Uuid: "8623ccaa-704e-4839-bc72-9a89daa20111",
					},
					Id: "123456789012-3",
				},
				Name: "test-bucket-3",
				Capacity: &stcnt_api.Bucket_Capacity{
					TotalBytes:     20000000000,
					AvailableBytes: 10000000000,
				},
				EndpointUrl: "https://pdx05-minio-dev-2.us-staging-1.cloud.intel.com:9000",
			},
		},
	}
	mockS3Client.EXPECT().ListBuckets(gomock.Any(), gomock.Any()).Return(listBucketResponse, nil).AnyTimes()

	// Create mock for list clusters
	clusterList := &stcnt_api.ListClustersResponse{
		Clusters: []*stcnt_api.Cluster{
			{
				Id: &api.ClusterIdentifier{
					Uuid: "8623ccaa-704e-4839-bc72-9a89daa20111",
				},
				Labels: map[string]string{
					"category": "development",
				},
				Name:     "pdx05-dev-4",
				Location: "Location1",
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
			{
				Id: &api.ClusterIdentifier{
					Uuid: "7773ccaa-704e-4839-bc72-9a89daa20111",
				},
				Labels: map[string]string{
					"category": "development",
				},
				Name:     "Cluster2",
				Location: "Location2",
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
		},
	}
	mockClusterSvcClient.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(clusterList, nil).AnyTimes()
	// initialize storagecontrollerclient
	strclient := &sc.StorageControllerClient{
		S3ServiceClient:  mockS3Client,
		ClusterSvcClient: mockClusterSvcClient,
	}
	return strclient
}

func NewMockFilesystemServiceClient() pb.FilesystemPrivateServiceClient {
	mockCtrl := gomock.NewController(GinkgoT())
	client := pb.NewMockFilesystemPrivateServiceClient(mockCtrl)
	resp := pb.NewMockFilesystemPrivateService_SearchFilesystemRequestsClient(mockCtrl)
	resp.EXPECT().Recv().Return(&pb.FilesystemRequestResponse{

		Filesystem: &pb.FilesystemPrivate{
			Metadata: &pb.FilesystemMetadataPrivate{
				CloudAccountId: "123456789012",
				Name:           "test-volume",
			},
			Spec: &pb.FilesystemSpecPrivate{
				AvailabilityZone: "az1",
				Request:          &pb.FilesystemCapacity{Storage: "7GB"},
				Scheduler: &pb.FilesystemSchedule{
					Namespace: &pb.AssignedNamespace{Name: "ns123456789012"},
					Cluster: &pb.AssignedCluster{
						ClusterAddr: "pdx05-dev-4.us-staging-1.cloud.intel.com",
					},
				},
			},
			Status: &pb.FilesystemStatusPrivate{
				Mount: &pb.FilesystemMountStatusPrivate{
					ClusterAddr: "pdx-test-1",
				},
			},
		},
	}, nil).Times(2)
	resp.EXPECT().Recv().Return(nil, io.EOF).Times(1)
	client.EXPECT().SearchFilesystemRequests(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()
	client.EXPECT().DeletePrivate(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	return client
}

func NewMockQuotaManagementPrivateServiceClient() pb.QuotaManagementPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	client := pb.NewMockQuotaManagementPrivateServiceClient(mockController)
	ServiceResources := []*pb.ServiceQuotaResource{
		{
			QuotaConfig: &pb.QuotaConfig{
				Limits: 50,
			},
		},
	}
	client.EXPECT().GetResourceQuotaPrivate(gomock.Any(), gomock.Any()).Return(&pb.ServiceQuotasPrivate{
		DefaultQuota: &pb.ServiceQuotaPrivate{
			ServiceName:      "",
			ServiceResources: ServiceResources,
		},
		CustomQuota: &pb.ServiceQuotaPrivate{
			ServiceName:      "",
			ServiceResources: ServiceResources,
		},
	}, nil).AnyTimes()

	return client
}

var _ = AfterSuite(func() {
	// Clean up any resources here
})
