// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/storage/vast_storage_operator/controllers/..." make test-custom

package cloud

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/mock/gomock"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	idcclientset "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/clientset/versioned"
	idcinformerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/informers/externalversions"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	//+kubebuilder:scaffold:imports
)

var (
	cfg                   *rest.Config
	k8sClient             client.Client
	testEnv               *envtest.Environment
	ctx                   context.Context
	cancel                context.CancelFunc
	timeout               time.Duration = 45 * time.Second
	poll                  time.Duration = 5 * time.Second
	vastStorageController *StorageReconciler
	stopChannel           = make(chan struct{})
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
	ctx, cancel = context.WithCancel(context.Background())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:        []string{"../../../k8s/config/crd/bases"},
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = cloudv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	configFile := "testdata/operatorconfig.yaml"
	logf.Log.Info("CtrlTest", "configFile", configFile)

	ctrlConfig := cloudv1alpha1.VastStorageOperatorConfig{}
	options := ctrl.Options{Scheme: scheme.Scheme}
	if configFile != "" {
		var err error
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
		if err != nil {
			panic(err)
		}
	}
	httpClient, err := rest.HTTPClientFor(cfg)
	httpClient.Timeout = 5 * time.Second
	if err != nil {
		panic(err)
	}
	logf.Log.Info("CtrlTest", "httpClient", httpClient)

	kubeClientSet, err := kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(kubeClientSet).NotTo(BeNil())

	idcClientSet, err := idcclientset.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(kubeClientSet).NotTo(BeNil())

	informerFactory := idcinformerfactory.NewSharedInformerFactory(idcClientSet, 10*time.Minute)

	storageControllerClient, kmsClient := NewMockVastOperatorServiceClient(true)

	vastStorageController, err = NewStorageOperator(ctx, kubeClientSet, idcClientSet, informerFactory.Private().V1alpha1().VastStorages(), informerFactory, storageControllerClient, kmsClient, k8sManager)
	Expect(err).NotTo(HaveOccurred())

	stopCh := make(chan struct{})
	informerFactory.Start(stopCh)

	err = k8sManager.Add(manager.RunnableFunc(func(context.Context) error {
		return vastStorageController.Run(ctx, stopCh)
	}))
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	close(stopChannel)
	Eventually(func() error {
		return testEnv.Stop()
	}, timeout, poll).ShouldNot(HaveOccurred())
})

// Helper function to mock KMS service client
func NewMockStorageKMSPrivateServiceClient() pb.StorageKMSPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	kms := pb.NewMockStorageKMSPrivateServiceClient(mockController)
	kmsClient := &pb.GetSecretResponse{}
	// Mock the Put call for KMSClient
	kms.EXPECT().Get(gomock.Any(), gomock.Any()).Return(kmsClient, nil).AnyTimes()
	kms.EXPECT().Put(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).AnyTimes()

	return kms
}

func NewNamespace(namespace string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func NewVastStorageInit(namespace string, storageName string, availabilityZone string, uid string, filesystemType string) *cloudv1alpha1.VastStorage {
	fsType := cloudv1alpha1.FilesystemTypeComputeGeneral
	if filesystemType == "ComputeKubernetes" {
		fsType = cloudv1alpha1.FilesystemTypeComputeKubernetes
	}

	return &privatecloudv1alpha1.VastStorage{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "private.cloud.intel.com/v1alpha1",
			Kind:       "VastStorage",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            storageName,
			Namespace:       namespace,
			UID:             "ab03e000-9a4a-48b4-b9be-f9a0ce8f9e85",
			ResourceVersion: "",
			Labels:          nil,
		},
		Spec: privatecloudv1alpha1.VastStorageSpec{
			AvailabilityZone: availabilityZone,
			FilesystemName:   "test1",
			FilesystemType:   fsType,
			CSIVolumePrefix:  "csi-volume-prefix",
			StorageRequest: privatecloudv1alpha1.VASTFilesystemStorageRequest{
				Size: "10000",
			},
			StorageClass: "GeneralPurposeStd",
			ClusterAssignment: privatecloudv1alpha1.ClusterAssignment{
				ClusterUUID:    "66efeaca-e493-4a39-b683-15978aac90d6",
				ClusterVersion: "v1.0",
				NamespaceName:  "namespace",
			},
			MountConfig: privatecloudv1alpha1.MountConfig{
				VolumePath:    "/mnt/test1",
				MountProtocol: privatecloudv1alpha1.FilesystemMountProtocolNFSV4,
			},
			Networks: privatecloudv1alpha1.Networks{
				SecurityGroups: privatecloudv1alpha1.SecurityGroups{
					IPFilters: []privatecloudv1alpha1.IPFilter{
						{Start: "192.168.1.1", End: "192.168.1.255"},
					},
				},
			},
		},
	}
}
