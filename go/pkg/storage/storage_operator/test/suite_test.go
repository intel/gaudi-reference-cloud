// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"testing"

	"time"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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

func NewStorageInit(namespace string, storageName string, availabilityZone string, uid string, filesystemType string) *cloudv1alpha1.Storage {
	fsType := cloudv1alpha1.FilesystemTypeComputeGeneral
	if filesystemType == "ComputeKubernetes" {
		fsType = cloudv1alpha1.FilesystemTypeComputeKubernetes
	}

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
				Size: "1000",
			},

			ProviderSchedule: privatecloudv1alpha1.FilesystemSchedule{
				FilesystemName: "testfs",
				Cluster: privatecloudv1alpha1.AssignedCluster{
					UUID: "66efeaca-e493-4a39-b683-15978aac90d5",
				},
			},

			StorageClass:   "DefaultFS",
			FilesystemType: fsType,
			AccessModes:    "ReadWrite",
			MountProtocol:  "Weka",
			Encrypted:      false,
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
	Eventually(func() error {
		return testEnv.Stop()
	}).ShouldNot(HaveOccurred())
})
