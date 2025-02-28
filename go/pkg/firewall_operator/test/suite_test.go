// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This test suite utilizes the following components:
//
//   - Kubernetes API Server
//   - etcd (for Kubernetes)
//   - Loadbalancer Operator
//
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/firewall_operator/..." make test-custom
package test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	firewallv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/internal/controller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/internal/provider"
	mock_provider "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/internal/provider/mock"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
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
			"../config/crd/bases",        // Firewall APIs
			"../../k8s/config/crd/bases", // Instance APIs
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
	fwProvider := NewMockIDCFWClient()

	fwReconciler := &controller.FirewallRuleReconciler{
		Client:                     k8sManager.GetClient(),
		Scheme:                     k8sManager.GetScheme(),
		FirewallRuleProviderConfig: &controller.FirewallRuleProviderConfig{},
		Provider:                   fwProvider,
	}

	// Setup the Firewall CRD Reconciler
	err = ctrl.NewControllerManagedBy(k8sManager).
		For(&firewallv1alpha1.FirewallRule{}).Complete(fwReconciler)
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

const vip1 = "1.2.3.4"
const customerId1 = "123456789012"

var vip1ExistingCustomerAccess = []provider.Rule{{
	CustomerId:  customerId1,
	Environment: "unittest",
	DestIp:      vip1,
	Port:        "80",
	Region:      "us-unittest-1",
	SourceIp:    "any",
}}

func NewMockIDCFWClient() provider.FirewallProvider {

	mockController := gomock.NewController(GinkgoT())
	mockProvider := mock_provider.NewMockFirewallProvider(mockController)

	mockProvider.EXPECT().SyncFirewallRules(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockProvider.EXPECT().GetExistingCustomerAccess(gomock.Any(), customerId1, vip1).Return(vip1ExistingCustomerAccess, nil).AnyTimes()
	mockProvider.EXPECT().GetExistingCustomerAccess(gomock.Any(), gomock.Any(), gomock.Any()).Return([]provider.Rule{}, nil).AnyTimes()
	return mockProvider
}
