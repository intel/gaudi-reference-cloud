// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This test suite utilizes the following components:
//
//   - Kubernetes API Server
//   - etcd (for Kubernetes)
//   - Mock of Metering Server (GRPC)
//   - Storage Metering Monitor (Metering Server to Metering DB )
//
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/compute_metering_monitor/..." make test-custom
package metering_monitor

import (
	"context"
	"testing"
	"time"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	interval                    = time.Second * 10
	maxRequeueTimeMillliseconds = time.Millisecond * 500
)

var (
	k8sRestConfig *rest.Config
	k8sClient     client.Client
	testEnv       *envtest.Environment
	scheme        *runtime.Scheme
	timeout       time.Duration = 90 * time.Second
	monitor       *MeteringMonitor
)

func NewNamespace(namespace string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func NewStorageInit(namespace string, storageName string, availabilityZone string, uid string) *privatecloudv1alpha1.Storage {
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
			StorageClass:  "DefaultFS",
			AccessModes:   "ReadWrite",
			MountProtocol: "Weka",
			Encrypted:     false,
		},

		Status: privatecloudv1alpha1.StorageStatus{
			Conditions: []privatecloudv1alpha1.StorageCondition{
				{
					Type:               privatecloudv1alpha1.StorageConditionFailed,
					LastTransitionTime: metav1.Time{Time: time.Now()},
				},
			},
		},
		//StorageClass
		//ProviderSchedule

	}
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metering Monitor Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()
	log.SetDefaultLogger()
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
	log.Info("BEGIN")
	defer log.Info("END")
	By("Stopping Kubernetes API Server")
	Eventually(func() error {
		return testEnv.Stop()
	}).ShouldNot(HaveOccurred())
})
