// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This test suite utilizes the following components:
//
//   - Kubernetes API Server
//   - etcd (for Kubernetes)
//   - Mock of Compute API Server (GRPC)
//   - Loadbalancer Replicator (Compute API Server to K8s Loadbalancer)
//
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/loadbalancer_replicator/..." make test-custom
package test

import (
	"context"
	"testing"

	lbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_replicator/convert"
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

var (
	testEnv               *envtest.Environment
	k8sRestConfig         *rest.Config
	k8sClient             client.Client
	scheme                *runtime.Scheme
	loadbalancerConverter *convert.LoadBalancerConverter
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scheduler Suite")
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
			"../../loadbalancer_operator/config/crd/bases", // LoadBalancer APIs
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
	Expect(lbv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

	By("Creating Kubernetes client")
	k8sClient, err = client.New(k8sRestConfig, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	loadbalancerConverter, err = convert.NewLoadBalancerConverter("us-dev-1", "us-dev-1a")
	Expect(err).NotTo(HaveOccurred())
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
