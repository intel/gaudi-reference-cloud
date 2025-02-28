// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloud

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	"google.golang.org/protobuf/types/known/emptypb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	testEnv          *envtest.Environment
	k8sRestConfig    *rest.Config
	scheme           *runtime.Scheme
	k8sClient        client.Client
	managerStoppable *stoppable.Stoppable
	timeout          time.Duration = 10 * time.Second
)

const (
	interval                    = time.Millisecond * 500
	maxRequeueTimeMillliseconds = time.Millisecond * 500
)

func NewNamespace(namespace string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func NewStorageInit(namespace string, storageName string, availabilityZone string, uid string) *cloudv1alpha1.Storage {
	return &privatecloudv1alpha1.Storage{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "private.cloud.intel.com/v1alpha1",
			Kind:       "Storage",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            storageName,
			Namespace:       namespace,
			UID:             "ab03e000-9a4a-48b4-b9be-f9a0ce8f9e84",
			ResourceVersion: "",
			Labels:          nil,
		},
		Spec: privatecloudv1alpha1.StorageSpec{
			AvailabilityZone: availabilityZone,
			StorageRequest: privatecloudv1alpha1.FilesystemStorageRequest{
				Size: "1000000",
			},

			ProviderSchedule: privatecloudv1alpha1.FilesystemSchedule{
				FilesystemName: "testfs",
				Cluster: privatecloudv1alpha1.AssignedCluster{
					UUID: "66efeaca-e493-4a39-b683-15978aac90d5",
				},
			},

			StorageClass:  "DefaultFS",
			AccessModes:   "ReadWrite",
			MountProtocol: "Weka",
			Encrypted:     false,
		},
	}
}

func TestStorageReplicator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StorageOperator Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("BeforeSuite")
	log.Info("BEGIN")
	defer log.Info("END")

	By("Configuring scheme")
	scheme = runtime.NewScheme()
	Expect(privatecloudv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

})

var (
	fsMetadata           sc.FilesystemMetadata
	fsProperties         sc.FilesystemProperties
	fs                   sc.Filesystem
	userMockClient       *mocks.MockUserServiceClient
	strclient            *sc.StorageControllerClient
	fsResponse           *weka.Filesystem
	nsResponse           *api.Namespace
	mockCtrl             *gomock.Controller
	nsMockClient         *mocks.MockNamespaceServiceClient
	user                 sc.User
	metadata             sc.UserMetadata
	ctx                  context.Context
	userResponse         *api.User
	clusterUUID          string
	filesystemID         string
	namespaceID          string
	namespace            sc.Namespace
	namespace1           *api.GetNamespaceResponse
	nsmetadata           sc.NamespaceMetadata
	mockFSClient         *mocks.MockFilesystemServiceClient
	mockUserClient       *mocks.MockUserServiceClient
	mockControllerSecret *gomock.Controller
	mockKmsClient        pb.StorageKMSPrivateServiceClient
	request              *weka.ListFilesystemsRequest
)

func NewStorage(namespace string, storageName string) *privatecloudv1alpha1.Storage {
	return NewStorageInit(
		namespace,
		storageName,
		"az1",
		"ab03e000-9a4a-48b4-b9be-f9a0ce8f9e84",
	)
}

func NewMockOperatorServiceClient() (pb.StorageKMSPrivateServiceClient, *api.Namespace, *weka.ListFilesystemsRequest, *weka.Filesystem, *sc.StorageControllerClient) {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	clusterUUID = "66efeaca-e493-4a39-b683-15978aac90d5"
	filesystemID = "fs_id"
	userID := "user_id"
	namespaceID = "ns_id"
	mockCtrl = gomock.NewController(GinkgoT())
	mockControllerSecret = gomock.NewController(GinkgoT())
	Expect(mockControllerSecret).NotTo(BeNil())
	mockKmsClient := NewMockStorageKMSPrivateServiceClient()
	mockFSClient = mocks.NewMockFilesystemServiceClient(mockCtrl)
	Expect(mockFSClient).NotTo(BeNil())
	mockUserClient = mocks.NewMockUserServiceClient(mockCtrl)
	Expect(mockUserClient).NotTo(BeNil())
	nsMockClient = mocks.NewMockNamespaceServiceClient(mockCtrl)

	strclient = &sc.StorageControllerClient{
		WekaFilesystemSvcClient: mockFSClient,
		NamespaceSvcClient:      nsMockClient,
		UserSvcClient:           mockUserClient,
	}

	nsmetadata := sc.NamespaceMetadata{
		ClusterId: "your-cluster-id",
		Name:      "test-namespace",
		User:      "test-user",
		Password:  "test-password",
		UUID:      clusterUUID,
	}
	namespace = sc.Namespace{
		Metadata:   nsmetadata,
		Properties: sc.NamespaceProperties{Quota: "2"},
	}

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
		Name:         "testfs-2",
		Status:       weka.Filesystem_STATUS_READY,
		IsEncrypted:  true,
		AuthRequired: true,
		Capacity: &weka.Filesystem_Capacity{
			TotalBytes:     20000000,
			AvailableBytes: 100000,
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
	existsrequest := &weka.ListFilesystemsRequest{
		NamespaceId: &api.NamespaceIdentifier{
			ClusterId: &api.ClusterIdentifier{
				Uuid: clusterUUID,
			},
			Id: namespaceID,
		},
		Filter: &weka.ListFilesystemsRequest_Filter{
			Names: []string{"testfs"},
		},
		AuthCtx: &api.AuthenticationContext{
			Scheme: &api.AuthenticationContext_Basic_{
				Basic: &api.AuthenticationContext_Basic{
					Principal:   "test",
					Credentials: "test",
				},
			},
		},
	}
	//set mock expectations
	nsResponse := &api.Namespace{
		Id: &api.NamespaceIdentifier{
			ClusterId: &api.ClusterIdentifier{
				Uuid: "918b5026-d516-48c8-bfd3-5998547265b2",
			},
			Id: "test",
		},
		Name: "ns123456789013",
		Quota: &api.Namespace_Quota{
			TotalBytes: 100000000,
		},
	}
	Expect(fsResponse).NotTo(BeNil())
	nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
		Namespaces: []*api.Namespace{nsResponse},
	}, nil).AnyTimes()
	nsMockClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(&api.UpdateNamespaceResponse{
		Namespace: nsResponse,
	}, nil).AnyTimes()
	Expect(mockFSClient).NotTo(BeNil())
	mockFSClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any()).Return(&weka.ListFilesystemsResponse{
		Filesystems: []*weka.Filesystem{fsResponse},
	}, nil).AnyTimes()
	mockFSClient.EXPECT().UpdateFilesystem(gomock.Any(), gomock.Any()).Return(&weka.UpdateFilesystemResponse{
		Filesystem: fsResponse,
	}, nil).AnyTimes()
	mockFSClient.EXPECT().DeleteFilesystem(gomock.Any(), gomock.Any()).Return(&weka.DeleteFilesystemResponse{}, nil).AnyTimes()
	Expect(existsrequest).NotTo(BeNil())
	mockFSClient.EXPECT().CreateFilesystem(gomock.Any(), gomock.Any()).Return(&weka.CreateFilesystemResponse{Filesystem: fsResponse}, nil).AnyTimes()
	return mockKmsClient, nsResponse, existsrequest, fsResponse, strclient
}

var _ = Describe("Storage Operator", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	BeforeEach(func() {
	})
	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("Operator tests", func() {
		It("Should create the Storage Namespaces and FileSystems", func() {
			ctx = context.Background()
			storage := NewStorage("123456789012", "file1")
			Expect(storage).NotTo(BeNil())
			mockKmsClient, nsResponse, request, fsResponse, strclient = NewMockOperatorServiceClient()
			r := &StorageReconciler{
				StorageControllerClient: strclient,
				kmsClient:               mockKmsClient,
			}
			Expect(r).NotTo(BeNil())
			Expect(mockKmsClient).NotTo(BeNil())
			err := r.createFileSystem(ctx, storage)
			Expect(err).To(BeNil())
		})
		It("Should update Storage Namespace and FileSystem", func() {
			ctx := context.Background()
			storage := NewStorage("123456789013", "test2")
			storage.Status = privatecloudv1alpha1.StorageStatus{
				Size: "2000",
			}
			mockKmsClient, nsResponse, request, fsResponse, strclient = NewMockOperatorServiceClient()
			r := &StorageReconciler{
				StorageControllerClient: strclient,
				kmsClient:               mockKmsClient,
			}
			err := r.updateFileSystem(ctx, storage)
			Expect(err).To(BeNil())
		})
		It("Should update Org", func() {
			ctx := context.Background()
			storage := NewStorage("123456789013", "test3")
			storage.Status = privatecloudv1alpha1.StorageStatus{
				Size: "500000",
			}
			mockKmsClient, nsResponse, request, fsResponse, strclient = NewMockOperatorServiceClient()
			r := &StorageReconciler{
				StorageControllerClient: strclient,
				kmsClient:               mockKmsClient,
			}
			err := r.updateOrg(ctx, storage)
			Expect(err).To(BeNil())
		})
	})
	It("Should delete the Storage Namespaces and FileSystems", func() {
		ctx = context.Background()
		storage := NewStorage("123456789012", "file1")
		Expect(storage).NotTo(BeNil())
		mockKmsClient, nsResponse, request, fsResponse, strclient = NewMockOperatorServiceClient()
		r := &StorageReconciler{
			StorageControllerClient: strclient,
			kmsClient:               mockKmsClient,
		}
		Expect(r).NotTo(BeNil())
		Expect(mockKmsClient).NotTo(BeNil())
		err := r.deleteStorage(ctx, storage)
		Expect(err).To(BeNil())
		Expect(r.deleteStorage(ctx, storage)).Should(Succeed())
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
