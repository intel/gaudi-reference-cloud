// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	str "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_operator/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"
	mocks "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func NewStorage(namespace string, storageName string, filesystemType string) *privatecloudv1alpha1.Storage {
	return NewStorageInit(
		namespace,
		storageName,
		"az1",
		"ab03e000-9a4a-48b4-b9be-f9a0ce8f9e84",
		filesystemType,
	)
}

var (
	fsMetadata     sc.FilesystemMetadata
	fsProperties   sc.FilesystemProperties
	fs             sc.Filesystem
	userMockClient *mocks.MockUserServiceClient
	strclient      *sc.StorageControllerClient
	fsResponse     *weka.Filesystem
	nsResponse     *api.Namespace
	mockCtrl       *gomock.Controller
	mockFsClient   *mocks.MockFilesystemServiceClient
	nsMockClient   *mocks.MockNamespaceServiceClient
	user           sc.User
	metadata       sc.UserMetadata
	ctx            context.Context
	userResponse   *api.User
	clusterUUID    string
	filesystemID   string
	namespaceID    string
	namespace      sc.Namespace
	nsmetadata     sc.NamespaceMetadata
	mockKmsClient  pb.StorageKMSPrivateServiceClient
)

func NewMockOperatorServiceClient(kmsSuccess bool) (*storagecontroller.StorageControllerClient, pb.StorageKMSPrivateServiceClient) {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	mockKmsClient := NewMockStorageKMSPrivateServiceClient()
	if kmsSuccess {
		mockKmsClient = NewFakeMockStorageKMSPrivateServiceClient()
	}

	mockCtrl = gomock.NewController(GinkgoT())
	mockFSClient := mocks.NewMockFilesystemServiceClient(mockCtrl)
	nsMockClient = mocks.NewMockNamespaceServiceClient(mockCtrl)
	clusterUUID = "66efeaca-e493-4a39-b683-15978aac90d5"
	filesystemID = "fs_id"
	userID := "user_id"
	namespaceID = "ns_id"
	mockUserClient := mocks.NewMockUserServiceClient(mockCtrl)

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
	Expect(existsrequest).NotTo(BeNil())

	mockFSClient.EXPECT().ListFilesystems(gomock.Any(), gomock.Any(), gomock.Any()).Return(&weka.ListFilesystemsResponse{
		Filesystems: []*weka.Filesystem{fsResponse},
	}, nil).AnyTimes()
	nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
		Namespaces: []*api.Namespace{nsResponse},
	}, nil).AnyTimes()

	mockFSClient.EXPECT().DeleteFilesystem(gomock.Any(), gomock.Any(), gomock.Any()).Return(&weka.DeleteFilesystemResponse{}, nil).AnyTimes()
	nsMockClient.EXPECT().DeleteNamespace(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.DeleteNamespaceResponse{}, nil).AnyTimes()

	mockUserClient.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Return(&api.ListUsersResponse{
		Users: []*api.User{userResponse},
	}, nil).AnyTimes()
	nsMockClient.EXPECT().CreateNamespace(gomock.Any(), gomock.Any()).Return(&api.CreateNamespaceResponse{
		Namespace: nsResponse,
	}, nil).AnyTimes()
	mockFSClient.EXPECT().CreateFilesystem(gomock.Any(), gomock.Any()).Return(&weka.CreateFilesystemResponse{
		Filesystem: fsResponse,
	}, nil).AnyTimes()

	nsMockClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.UpdateNamespaceResponse{}, nil).AnyTimes()

	mockUserClient.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(&api.CreateUserResponse{
		User: userResponse,
	}, nil).AnyTimes()

	return strclient, mockKmsClient
}

var _ = Describe("Storage Operator", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log
	var (
		managerStoppable *stoppable.Stoppable
	)

	BeforeEach(func() {
		By("Creating manager")
		k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
			Scheme:  scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
		})
		Expect(err).ToNot(HaveOccurred())

		client, kms_client := NewMockOperatorServiceClient(false)
		Expect(client).NotTo(BeNil())
		Expect(kms_client).NotTo(BeNil())

		_, err = str.NewStorageOperator(ctx, k8sManager, client, kms_client)

		By("Starting manager")
		managerStoppable = stoppable.New(k8sManager.Start)
		managerStoppable.Start(ctx)
	})

	AfterEach(func() {
		By("Stopping manager")
		Expect(managerStoppable.Stop(ctx)).Should(Succeed())
		By("Manager stopped")
	})
	Context("Operator tests", func() {
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		filesystemType := "ComputeGeneral"
		storage := NewStorage(namespace, filesystemName, filesystemType)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: filesystemName}
		storageRef := &privatecloudv1alpha1.Storage{}
		nsObject := NewNamespace(namespace)
		Expect(storage).NotTo(BeNil())
		It("Should create namespace successfully", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create storage successfully", func() {
			Expect(k8sClient.Create(ctx, storage)).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
		})

		It("When a Storage is created and in Running Condition", func() {
			namespace := uuid.NewString()
			filesystemName := uuid.NewString()
			filesystemType := "ComputeGeneral"
			storage := NewStorage(namespace, filesystemName, filesystemType)

			Expect(storage).NotTo(BeNil())
		})

		It("Reconciler should add finalizer to storages CRD", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
				g.Expect(storageRef.GetFinalizers()).Should(ConsistOf(str.StorageFinalizer, str.StorageMeteringMonitorFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("Storage Resource should have Accepted condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
				log.Info("Storage Reconciler should set Accepted condition to true", logkeys.StatusConditions, storageRef.Status.Conditions)
				cond := util.FindStatusCondition(storageRef.Status.Conditions, privatecloudv1alpha1.StorageConditionAccepted)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("Storage Reconciler should set Phase to Provisioning", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
				log.Info("Storage Reconciler should set Phase to Provisioning", logkeys.StatusPhase, storageRef.Status.Phase,
					logkeys.StatusMessage, storageRef.Status.Message, logkeys.StatusConditions, storageRef.Status.Conditions)
				g.Expect(storageRef.Status.Phase).Should(Equal(privatecloudv1alpha1.FilesystemPhaseReady))
				g.Expect(storageRef.Status.Message).Should(MatchRegexp(".*%s.*", privatecloudv1alpha1.StorageMessageRunning))
			}, timeout, interval).Should(Succeed())
		})

		It("Waiting for storage to be deleted from K8s", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Delete(ctx, storageRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

		})

		namespace2 := ""
		filesystemName2 := ""
		filesystemType = "ComputeGeneral"
		storage2 := NewStorage(namespace2, filesystemName, filesystemType)
		storageLookupKey2 := types.NamespacedName{Namespace: namespace2, Name: filesystemName2}
		storageRef2 := &privatecloudv1alpha1.Storage{}
		nsObject2 := NewNamespace(namespace)
		Expect(storage2).NotTo(BeNil())
		It("Should give error while creating namespace", func() {
			Expect(k8sClient.Create(ctx, nsObject2)).ShouldNot(Succeed())
		})

		It("Should give error while creating storage", func() {
			Expect(k8sClient.Create(ctx, nsObject2)).ShouldNot(Succeed())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey2, storageRef2)).ShouldNot(Succeed())
			}, timeout, interval).Should(Succeed())
		})

	})

	It("When we delete a storage the object will be deleted and phase and status will be set accordingly and object will be removed from weka", func() {
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		filesystemType := "ComputeGeneral"
		storage := NewStorage(namespace, filesystemName, filesystemType)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: filesystemName}
		storageRef := &privatecloudv1alpha1.Storage{}
		nsObject := NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("Creating Storage successfully")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Create(ctx, storage)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for storage to be created in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Updating storage status in K8s (simulate Storage Operator)")
		runningCondition := privatecloudv1alpha1.StorageCondition{
			Type:   privatecloudv1alpha1.StorageConditionRunning,
			Status: v1.ConditionTrue,
		}
		storageRef.Status.Conditions = append(storageRef.Status.Conditions, runningCondition)
		Expect(k8sClient.Status().Update(ctx, storageRef)).Should(Succeed())

		By("Waiting for storage to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Delete(ctx, storageRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for storage to be not found so as to confirm its deletion")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, storageLookupKey, storageRef))).Should(BeTrue())
		}, timeout, interval).Should(Succeed())

	})

	Context("Operator tests for k8s flow", func() {
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		filesystemType := "ComputeKubernetes"
		storage := NewStorage(namespace, filesystemName, filesystemType)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: filesystemName}
		storageRef := &privatecloudv1alpha1.Storage{}
		nsObject := NewNamespace(namespace)
		Expect(storage).NotTo(BeNil())
		It("Should create namespace", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create storage successfully for K8s object", func() {
			Expect(k8sClient.Create(ctx, storage)).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
		})

		It("When a Storage is created for k8s flow", func() {
			namespace := uuid.NewString()
			filesystemName := uuid.NewString()
			filesystemType := "ComputeKubernetes"
			storage := NewStorage(namespace, filesystemName, filesystemType)

			Expect(storage).NotTo(BeNil())
		})

		It("Reconciler should add finalizer to storages CRD", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
				g.Expect(storageRef.GetFinalizers()).Should(ConsistOf(str.StorageFinalizer, str.StorageMeteringMonitorFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("Storage Resource should have Accepted condition to true ", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
				log.Info("Storage Reconciler should set Accepted condition to true", "conditions", storageRef.Status.Conditions)
				cond := util.FindStatusCondition(storageRef.Status.Conditions, privatecloudv1alpha1.StorageConditionAccepted)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("Storage Reconciler should set Phase to Ready", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
				log.Info("Storage Reconciler should set Phase to Provisioning", "storageRef.Status.Phase", storageRef.Status.Phase,
					"storageRef.Status.Message", storageRef.Status.Message, "conditions", storageRef.Status.Conditions)
				g.Expect(storageRef.Status.Phase).Should(Equal(privatecloudv1alpha1.FilesystemPhaseReady))
				g.Expect(storageRef.Status.Message).Should(MatchRegexp(".*%s.*", privatecloudv1alpha1.StorageMessageRunning))
			}, timeout, interval).Should(Succeed())
		})

		It("Waiting for storage to be deleted from K8s", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Delete(ctx, storageRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

		})

		namespace2 := ""
		filesystemName2 := ""
		filesystemType = "ComputeKubernetes"
		storage2 := NewStorage(namespace2, filesystemName, filesystemType)
		storageLookupKey2 := types.NamespacedName{Namespace: namespace2, Name: filesystemName2}
		storageRef2 := &privatecloudv1alpha1.Storage{}
		nsObject2 := NewNamespace(namespace)
		Expect(storage2).NotTo(BeNil())
		It("Should give error while creating namespace", func() {
			Expect(k8sClient.Create(ctx, nsObject2)).ShouldNot(Succeed())
		})

		It("Should give error while creating storage", func() {
			Expect(k8sClient.Create(ctx, nsObject2)).ShouldNot(Succeed())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, storageLookupKey2, storageRef2)).ShouldNot(Succeed())
			}, timeout, interval).Should(Succeed())
		})
	})

	It("When we delete a storage the object will be deleted and phase and status will be set accordingly and object will be removed from weka", func() {
		namespace := uuid.NewString()
		filesystemName := uuid.NewString()
		filesystemType := "ComputeKubernetes"
		storage := NewStorage(namespace, filesystemName, filesystemType)
		storageLookupKey := types.NamespacedName{Namespace: namespace, Name: filesystemName}
		storageRef := &privatecloudv1alpha1.Storage{}
		nsObject := NewNamespace(namespace)

		By("Creating namespace successfully")
		Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())

		By("Creating Storage successfully")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Create(ctx, storage)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for storage to be created in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, storageLookupKey, storageRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Updating storage status in K8s (simulate Storage Operator)")
		runningCondition := privatecloudv1alpha1.StorageCondition{
			Type:   privatecloudv1alpha1.StorageConditionRunning,
			Status: v1.ConditionTrue,
		}
		storageRef.Status.Conditions = append(storageRef.Status.Conditions, runningCondition)
		Expect(k8sClient.Status().Update(ctx, storageRef)).Should(Succeed())

		By("Waiting for storage to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Delete(ctx, storageRef)).Should(Succeed())
		}, timeout, interval).Should(Succeed())

		By("Waiting for storage to be not found so as to confirm its deletion")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, storageLookupKey, storageRef))).Should(BeTrue())
		}, timeout, interval).Should(Succeed())

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

func NewFakeMockStorageKMSPrivateServiceClient() pb.StorageKMSPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	kms := pb.NewMockStorageKMSPrivateServiceClient(mockController)
	kms.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("new error")).AnyTimes()
	kms.EXPECT().Put(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).AnyTimes()

	return kms
}
