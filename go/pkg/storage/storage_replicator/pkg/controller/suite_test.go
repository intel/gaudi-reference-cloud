// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/timestamp"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_replicator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	interval = time.Millisecond * 500
)

var (
	testEnv       *envtest.Environment
	k8sRestConfig *rest.Config
	scheme        *runtime.Scheme
	k8sClient     client.Client
	timeout       time.Duration = 10 * time.Second
)

func TestStorageReplicator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StorageReplicator Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("BeforeSuite")
	log.Info("BEGIN")
	defer log.Info("END")

	By("Starting Kubernetes API Server")
	testEnv = &envtest.Environment{
		// When adding CRDS, be sure to add them to the data list in BUILD.bazel.
		CRDDirectoryPaths: []string{
			"../../../../k8s/config/crd/bases",
		},
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: true,
		//BinaryAssetsDirectory:    "/usr/bin/etcd",
	}
	Expect(testEnv).NotTo(BeNil())
	var err error
	k8sRestConfig, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sRestConfig).NotTo(BeNil())

	By("Configuring scheme")
	scheme = runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).NotTo(HaveOccurred())
	Expect(privatecloudv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
	By("Creating Kubernetes client")

	k8sClient, err = client.New(k8sRestConfig, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("AfterSuite")
	log.Info("Begin")
	defer log.Info("End")
	Eventually(func() error {
		return testEnv.Stop()
	}).ShouldNot(HaveOccurred())
})

// Mock gRPC clientConn
func NewGRPCClient() *pb.MockFilesystemPrivateServiceClient {
	client := pb.NewMockFilesystemPrivateServiceClient(gomock.NewController(GinkgoT()))
	return client
}

// Override New Replicator service constructor to use mock gRPC client
func NewMockReplicatorService() *StorageReplicatorService {
	clusterClient, err := dynamic.NewForConfig(k8sRestConfig)
	Expect(err).To(BeNil())

	cfg := config.NewDefaultConfig()
	resource := schema.GroupVersionResource{Group: "private.cloud.intel.com", Version: "v1alpha1", Resource: "storages"}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, corev1.NamespaceAll, nil)
	informer := factory.ForResource(resource).Informer()
	mockClient := NewGRPCClient()
	rService := &StorageReplicatorService{
		syncTicker:       time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
		Cfg:              cfg,
		storageAPIClient: mockClient,
		k8sclient:        clusterClient,
		informer:         informer,
	}
	// cfg.IDCServiceConfig.StorageAPIGrpcEndpoint = "localhost"
	// replicator, err:= controller.NewStorageReplicatorService(context.Background(),cfg,k8sRestConfig)

	return rService
}

func NewStorage(namespace string, storageName string) *privatecloudv1alpha1.Storage {
	return NewStorageInit(
		namespace,
		storageName,
		"az1",
		storageName,
	)
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
			//CreationTimestamp: ,
			//DeletionTimestamp: ,

		},
		Spec: privatecloudv1alpha1.StorageSpec{
			AvailabilityZone: availabilityZone,
			StorageRequest: privatecloudv1alpha1.FilesystemStorageRequest{
				Size: "1000",
			},
			// ProviderSchedule: privatecloudv1alpha1.FilesystemSchedule{
			// 	FilesystemName: "storage_v1",
			// },
			StorageClass:  "ComputeGeneral",
			AccessModes:   "ReadWrite",
			MountProtocol: "Weka",
			Encrypted:     false,
		},

		//StorageClass
		//ProviderSchedule

	}
}

func NewRequest(delete bool) *pb.FilesystemRequestResponse {
	var version = new(string)
	*version = "v2"
	var dTime *timestamppb.Timestamp
	if delete {
		dTime = timestamppb.Now()
	}
	fsResp := &pb.FilesystemRequestResponse{
		Filesystem: &pb.FilesystemPrivate{
			Metadata: &pb.FilesystemMetadataPrivate{
				CloudAccountId:    "123456789012",
				Name:              "test",
				ResourceId:        "6787226a-2a55-4d6f-bae9-fa2a2ca2450a",
				ResourceVersion:   "1",
				Description:       "Sample Filesystem",
				Labels:            map[string]string{"key": "value"},
				CreationTimestamp: &timestamp.Timestamp{Seconds: 1637077200, Nanos: 0},
				DeletionTimestamp: dTime,
			}, //Meta
			Spec: &pb.FilesystemSpecPrivate{
				AvailabilityZone: "az1",
				Request: &pb.FilesystemCapacity{
					Storage: "2000000000",
				},
				MountProtocol: 0,
				Encrypted:     false,
				Scheduler: &pb.FilesystemSchedule{
					FilesystemName: "test",
					Cluster: &pb.AssignedCluster{
						ClusterName:    "1",
						ClusterAddr:    "1",
						ClusterUUID:    "1",
						ClusterVersion: version,
					},
					Namespace: &pb.AssignedNamespace{
						Name:            "123456789012",
						CredentialsPath: "/path/to/secret",
					},
				},
			}, //Spec
			Status: &pb.FilesystemStatusPrivate{
				Phase:   pb.FilesystemPhase_FSProvisioning,
				Message: "Filesystem is being provisioned",
				// Add relevant fields from your status
			}, // Status
		}, //FSPrivate
	}
	return fsResp
}
