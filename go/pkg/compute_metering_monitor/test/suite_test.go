// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This test suite utilizes the following components:
//
//   - Kubernetes API Server
//   - etcd (for Kubernetes)
//   - Mock of Metering Server (GRPC)
//   - Compute Metering Monitor (Metering Server to Metering DB )
//
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/compute_metering_monitor/..." make test-custom
package test

import (
	"context"
	"testing"
	"time"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	interval                    = time.Millisecond * 500
	maxRequeueTimeMillliseconds = time.Millisecond * 500
)

var (
	k8sRestConfig *rest.Config
	k8sClient     client.Client
	testEnv       *envtest.Environment
	scheme        *runtime.Scheme
	timeout       time.Duration = 10 * time.Second
)

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
			"../../k8s/config/crd/bases",
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
