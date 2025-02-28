// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This test suite utilizes the following components:
//
//   - Kubernetes API Server
//   - etcd (for Kubernetes)
//   - Loadbalancer Operator
//
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/loadbalancer_operator/..." make test-custom
package test

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	firewallv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/controller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/processor"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/provider"
	mock_provider "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/provider/mock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8scontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sRestConfig    *rest.Config
	k8sClient        client.Client
	testEnv          *envtest.Environment
	scheme           *runtime.Scheme
	timeout          time.Duration = 10 * time.Second
	poll             time.Duration = 10 * time.Millisecond
	managerStoppable *stoppable.Stoppable
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	log.SetDefaultLogger()
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	By("Starting Kubernetes API Server")
	testEnv = &envtest.Environment{
		// When adding CRDS, be sure to add them to the data list in BUILD.bazel.
		CRDDirectoryPaths: []string{
			"../config/crd/bases",                      // Loadbalancer APIs
			"../../firewall_operator/config/crd/bases", // Firewall APIs
			"../../k8s/config/crd/bases",               // Instance APIs
			"../test-data/crd",
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
	Expect(v1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
	Expect(firewallv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
	Expect(privatecloudv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	By("Creating Kubernetes client")
	k8sClient, err = client.New(k8sRestConfig, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("Creating manager")
	k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
		Scheme:  scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	By("Configuring HTTP client")
	httpClient, err := rest.HTTPClientFor(k8sRestConfig)
	Expect(err).ToNot(HaveOccurred())
	httpClient.Timeout = 5 * time.Second

	By("Creating Mock Loadbalancer Provider")
	lbProvider := NewMockClient()

	By("Loading test configuration")
	appProcessor := processor.NewProcessor(k8sManager.GetClient(), lbProvider, k8sManager.GetScheme(), "us-dev-1", "us-dev-1a")

	lbr := &controller.LoadbalancerReconciler{
		Client:     k8sManager.GetClient(),
		Scheme:     k8sManager.GetScheme(),
		LBProvider: lbProvider,
		Processor:  appProcessor,
	}

	// Setup the Loadbalancer CRD Reconciler
	err = ctrl.NewControllerManagedBy(k8sManager).
		For(&loadbalancerv1alpha1.Loadbalancer{},
			builder.WithPredicates(
				// Reconcile Loadbalancer if Loadbalancer spec changes or annotation changes.
				predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{}),
			),
		).
		Owns(&firewallv1alpha1.FirewallRule{}).
		Watches(
			&privatecloudv1alpha1.Instance{},
			handler.EnqueueRequestsFromMapFunc(lbr.MapInstanceToLoadbalancer)).
		WithOptions(k8scontroller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(lbr)
	Expect(err).ToNot(HaveOccurred())

	By("Starting manager")
	managerStoppable = stoppable.New(k8sManager.Start)
	managerStoppable.Start(ctx)
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	By("Stopping manager")
	Expect(managerStoppable.Stop(ctx)).Should(Succeed())
	By("Manager stopped")
	By("Stopping Kubernetes API Server")
	Eventually(func() error {
		return testEnv.Stop()
	}, timeout, poll).ShouldNot(HaveOccurred())
})

func NewMockClient() provider.Provider {
	mockProvider := mock_provider.NewMockProvider()
	return mockProvider
}
