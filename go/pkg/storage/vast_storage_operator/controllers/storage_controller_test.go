package cloud

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"

	storageControllerVastApi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/vast"

	"errors"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/vast"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	vastmocks "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/vastmocks"
	"k8s.io/client-go/rest"
)

var (
	nsResponse *api.Namespace

	strClient            *sc.StorageControllerClient
	nsMockClient         *mocks.MockNamespaceServiceClient
	userMockClient       *mocks.MockUserServiceClient
	mockFSClient         *vastmocks.MockFilesystemServiceClient
	mockUserClient       *mocks.MockUserServiceClient
	mockControllerSecret *gomock.Controller
	mockKmsClient        pb.StorageKMSPrivateServiceClient
	k8sRestConfig        *rest.Config
	mockCtrl             *gomock.Controller
	clusterUUID          string
	filesystemID         string
	namespaceID          string
	namespace            sc.Namespace
	nsmetadata           sc.NamespaceMetadata
	strclient            *sc.StorageControllerClient
	fsMetadata           sc.FilesystemMetadata
	fsProperties         sc.FilesystemProperties
	fs                   sc.Filesystem
	fsResponse           *vast.Filesystem
	metadata             sc.UserMetadata
)

func NewVastStorage(namespace string, storageName string, filesystemType string) *privatecloudv1alpha1.VastStorage {
	return NewVastStorageInit(
		namespace,
		storageName,
		"az1",
		"ab03e000-9a4a-48b4-b9be-f9a0ce8f9e84",
		filesystemType,
	)
}

var _ = Describe("StorageReconciler", func() {
	ctx := context.Background()

	It("Test Add Event", func() {
		// Create a fake VastStorage resource
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		filesystemType := "ComputeGeneralStd"
		storage := NewVastStorage(namespace, filesystemName, filesystemType)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: filesystemName}
		storageRef := &privatecloudv1alpha1.VastStorage{}
		nsObject := NewNamespace(namespace)

		Expect(storage).NotTo(BeNil())
		Expect(k8sClient).NotTo(BeNil())

		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		Expect(k8sClient.Create(ctx, storage)).Should(Succeed())
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
			g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
			g.Expect(storageRef.Spec.FilesystemName).Should(Equal("test1"))
			print("this is storage ref")
			fmt.Printf("this is storage ref: %+v\n", storageRef.Spec)
			g.Expect(storageRef.GetFinalizers()).Should(ConsistOf(VastStorageFinalizer, VastStorageMeteringMonitorFinalizer))

		}, timeout, interval).Should(Succeed())

	})
	It("When a Storage is created and in Running Condition", func() {
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		filesystemType := "ComputeGeneralStd"
		storage := NewVastStorage(namespace, filesystemName, filesystemType)
		Expect(storage.Status).NotTo(BeNil())
		Expect(storage.Status.Phase).NotTo(BeNil())
	})

})

func NewMockVastOperatorServiceClient(kmsSuccess bool) (*storagecontroller.StorageControllerClient, pb.StorageKMSPrivateServiceClient) {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	mockKmsClient := NewMockStorageKMSPrivateServiceClient()
	if kmsSuccess {
		mockKmsClient = NewFakeMockStorageKMSPrivateServiceClient()
	}

	mockCtrl = gomock.NewController(GinkgoT())
	mockFSClient := vastmocks.NewMockFilesystemServiceClient(mockCtrl)
	nsMockClient = mocks.NewMockNamespaceServiceClient(mockCtrl)
	clusterUUID = "66efeaca-e493-4a39-b683-15978aac90d6"
	filesystemID = "1"
	namespaceID = "2"
	mockUserClient := mocks.NewMockUserServiceClient(mockCtrl)

	strclient = &sc.StorageControllerClient{
		VastFilesystemSvcClient: mockFSClient,
		NamespaceSvcClient:      nsMockClient,
		UserSvcClient:           mockUserClient,
	}

	nsmetadata := sc.NamespaceMetadata{
		ClusterId: "your-cluster-id",
		Name:      "test-namespace",
		UUID:      clusterUUID,
	}
	namespace = sc.Namespace{
		Metadata:   nsmetadata,
		Properties: sc.NamespaceProperties{Quota: "2"},
	}

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

	fsMetadata = sc.FilesystemMetadata{
		FileSystemName: "testfs",
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

	fsResponse = &vast.Filesystem{
		Id: &vast.FilesystemIdentifier{
			NamespaceId: &api.NamespaceIdentifier{
				ClusterId: &api.ClusterIdentifier{
					Uuid: clusterUUID,
				},
				Id: namespaceID,
			},
			Id: filesystemID,
		},
		Name: "testfs-2",
		Capacity: &vast.Filesystem_Capacity{
			TotalBytes:     20000000,
			AvailableBytes: 100000,
		},
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

	existsrequest := &vast.ListFilesystemsRequest{
		NamespaceId: &api.NamespaceIdentifier{
			ClusterId: &api.ClusterIdentifier{
				Uuid: clusterUUID,
			},
			Id: namespaceID,
		},
		Filter: &vast.ListFilesystemsRequest_Filter{
			Names: []string{"testfs"},
		},
	}
	Expect(existsrequest).NotTo(BeNil())

	nsMockClient.EXPECT().CreateNamespace(gomock.Any(), gomock.Any()).Return(&api.CreateNamespaceResponse{
		Namespace: nsResponse,
	}, nil).AnyTimes()

	// ListVastFilesystems
	mockFSClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any()).Return(&storageControllerVastApi.ListFilesystemsResponse{
		Filesystems: []*storageControllerVastApi.Filesystem{fsResponse},
	}, nil).AnyTimes()

	// GetVastFilesystem
	mockFSClient.EXPECT().GetFilesystem(gomock.Any(), gomock.Any()).Return(&storageControllerVastApi.GetFilesystemResponse{
		Filesystem: fsResponse,
	}, nil).AnyTimes()

	// CreateVastFilesystem
	mockFSClient.EXPECT().CreateFilesystem(gomock.Any(), gomock.Any()).Return(&storageControllerVastApi.CreateFilesystemResponse{
		Filesystem: fsResponse,
	}, nil).AnyTimes()

	// UpdateVastFilesystem
	mockFSClient.EXPECT().UpdateFilesystem(gomock.Any(), gomock.Any()).Return(&storageControllerVastApi.UpdateFilesystemResponse{
		Filesystem: fsResponse,
	}, nil).AnyTimes()

	// DeleteVastFilesystem
	mockFSClient.EXPECT().DeleteFilesystem(gomock.Any(), gomock.Any()).Return(&storageControllerVastApi.DeleteFilesystemResponse{}, nil).AnyTimes()

	nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
		Namespaces: []*api.Namespace{nsResponse},
	}, nil).AnyTimes()

	nsMockClient.EXPECT().CreateNamespace(gomock.Any(), gomock.Any()).Return(&api.CreateNamespaceResponse{
		Namespace: nsResponse,
	}, nil).AnyTimes()

	nsMockClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.UpdateNamespaceResponse{}, nil).AnyTimes()

	return strclient, mockKmsClient
}

func NewFakeMockStorageKMSPrivateServiceClient() pb.StorageKMSPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	kms := pb.NewMockStorageKMSPrivateServiceClient(mockController)
	kms.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("new error")).AnyTimes()
	kms.EXPECT().Put(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).AnyTimes()

	return kms
}
