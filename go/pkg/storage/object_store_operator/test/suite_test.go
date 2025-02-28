// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"testing"

	"time"

	"github.com/golang/mock/gomock"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	v1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	testEnv       *envtest.Environment
	k8sRestConfig *rest.Config
	scheme        *runtime.Scheme
	k8sClient     client.Client
	timeout       time.Duration = 10 * time.Second
)

const (
	interval                    = time.Millisecond * 500
	maxRequeueTimeMillliseconds = time.Millisecond * 500
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
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
			"../../../k8s/config/crd/bases",
		},
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: true,
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

	//+kubebuilder:scaffold:scheme

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
	By("Stopping Kubernetes API Server")
	Eventually(func() error {
		return testEnv.Stop()
	}).ShouldNot(HaveOccurred())
})

func NewMockStorageControllerClient() *sc.StorageControllerClient {
	mockCtrl := gomock.NewController(GinkgoT())
	mockS3Client := mocks.NewMockS3ServiceClient(mockCtrl)
	//Set mock expectations
	mockS3Client.EXPECT().CreateBucket(gomock.Any(), gomock.Any()).Return(&stcnt_api.CreateBucketResponse{
		Bucket: &stcnt_api.Bucket{
			Id: &stcnt_api.BucketIdentifier{
				ClusterId: &stcnt_api.ClusterIdentifier{
					Uuid: "8623ccaa-704e-4839-bc72-9a89daa20111",
				},
				Id: "test-bucket",
			},
			Name: "test-bucket",
			Capacity: &stcnt_api.Bucket_Capacity{
				TotalBytes:     10000000000,
				AvailableBytes: 10000000000,
			},
			EndpointUrl: "https://pdx05-minio-dev-2.us-staging-1.cloud.intel.com:9000",
		},
	}, nil).AnyTimes()
	mockS3Client.EXPECT().DeleteBucket(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	// initialize storagecontrollerclient
	strclient := &sc.StorageControllerClient{
		S3ServiceClient: mockS3Client,
	}

	return strclient
}

func NewBucket(namespace string, bucketName string) *v1alpha1.ObjectStore {
	return NewObjectStoreInit(
		namespace,
		bucketName,
		"az1",
		"",
	)
}

func NewNamespace(namespace string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func NewObjectStoreInit(namespace string, bucketName string, availabilityZone string, uid string) *v1alpha1.ObjectStore {
	return &v1alpha1.ObjectStore{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "private.cloud.intel.com/v1alpha1",
			Kind:       "ObjectStore",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bucketName,
			Namespace: namespace,
			UID:       "ab03e000-9a4a-48b4-b9be-f9a0ce8f9e84",
			Labels:    nil,
			//CreationTimestamp: ,
			//DeletionTimestamp: ,

		},
		Spec: v1alpha1.ObjectStoreSpec{
			AvailabilityZone:   availabilityZone,
			Versioned:          false,
			Quota:              "10000000000",
			BucketAccessPolicy: v1alpha1.BucketAccessPolicyReadWrite,
			ObjectStoreBucketSchedule: v1alpha1.ObjectStoreBucketSchedule{
				ObjectStoreCluster: v1alpha1.ObjectStoreAssignedCluster{
					Name: "minio",
					UUID: "8623ccaa-704e-4839-bc72-9a89daa20111",
					Addr: "https://pdx05-minio-dev-2.us-staging-1.cloud.intel.com:9000",
				},
			},
		},
		Status: v1alpha1.ObjectStoreStatus{
			Phase: v1alpha1.ObjectStorePhasePhaseReady,
		},
	}
}

func objectCondition() v1alpha1.ObjectStoreCondition {
	return v1alpha1.ObjectStoreCondition{
		Type:               v1alpha1.ObjectStoreConditionAccepted,
		Status:             v1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             v1alpha1.ObjectStoreConditionReasonAccepted,
		Message:            "bucket is ready",
	}
}
