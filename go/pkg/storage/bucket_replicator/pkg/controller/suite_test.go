package controller

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/timestamp"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/bucket_replicator/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
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

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bucket Replicator Server Suite")
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
func NewGRPCClient() *pb.MockObjectStorageServicePrivateClient {
	client := pb.NewMockObjectStorageServicePrivateClient(gomock.NewController(GinkgoT()))
	return client
}

func NewRequest(delete bool) *pb.ObjectBucketSearchPrivateResponse {
	var version = new(string)
	*version = "v2"
	var dTime *timestamppb.Timestamp
	if delete {
		dTime = timestamppb.Now()
	}
	bkResp := &pb.ObjectBucketSearchPrivateResponse{
		Bucket: &pb.ObjectBucketPrivate{
			Metadata: &pb.ObjectBucketMetadataPrivate{
				CloudAccountId:    "123456789012",
				Name:              "test",
				ResourceId:        "6787226a-2a55-4d6f-bae9-fa2a2ca2450a",
				ResourceVersion:   "1",
				Description:       "Sample Bucket",
				Labels:            map[string]string{"key": "value"},
				CreationTimestamp: &timestamp.Timestamp{Seconds: 1637077200, Nanos: 0},
				DeletionTimestamp: dTime,
			}, //Meta
			Spec: &pb.ObjectBucketSpecPrivate{
				AvailabilityZone: "az1",
				Request: &pb.StorageCapacityRequest{
					Size: "2000000000",
				},
				Versioned: true,
				Schedule: &pb.BucketSchedule{
					Cluster: &pb.AssignedCluster{
						ClusterName:    "1",
						ClusterAddr:    "1",
						ClusterUUID:    "1",
						ClusterVersion: version,
					},
				},
			}, //Spec
			Status: &pb.ObjectBucketStatus{
				Phase:   pb.BucketPhase_BucketProvisioning,
				Message: "Bucket is being provisioned",
				// Add relevant fields from your status
			}, // Status
		},
	}
	return bkResp
}

func NewMockReplicatorService() *BucketReplicatorService {
	clusterClient, err := dynamic.NewForConfig(k8sRestConfig)
	Expect(err).To(BeNil())

	cfg := config.NewDefaultConfig()
	resource := schema.GroupVersionResource{Group: "private.cloud.intel.com", Version: "v1alpha1", Resource: "objectstores"}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(clusterClient, time.Minute, corev1.NamespaceAll, nil)
	informer := factory.ForResource(resource).Informer()
	mockClient := NewGRPCClient()
	brService := &BucketReplicatorService{
		syncTicker:      time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
		Cfg:             cfg,
		bucketAPIClient: mockClient,
		k8sclient:       clusterClient,
		informer:        informer,
	}

	return brService
}
