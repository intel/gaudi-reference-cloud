// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/clientset/versioned/scheme"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog_operator/apis/private.cloud/v1alpha1"
	productcontroller "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog_operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	k8sRestConfig     *rest.Config
	restClient        *rest.RESTClient
	k8sClient         client.Client
	testEnv           *envtest.Environment
	productReconciler *productcontroller.ProductReconciler
	billingSyncClient pb.BillingProductCatalogSyncServiceClient
	clientConn        *grpc.ClientConn
	timeout           time.Duration = 10 * time.Second
	poll              time.Duration = 10 * time.Millisecond
	managerStoppable  *stoppable.Stoppable
	listener          *bufconn.Listener
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Product catalog gRPC service test")
}

func getBufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return listener.Dial()
	}
}

type mockClient struct {
	response *http.Response
	err      error
}

func (c *mockClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.response, c.err
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
			"../../productcatalog_operator/config/crd/bases",
		},
		ErrorIfCRDPathMissing:    true,
		AttachControlPlaneOutput: true,
	}
	var err error
	k8sRestConfig, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sRestConfig).NotTo(BeNil())

	By("Configuring scheme")
	Expect(clientgoscheme.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(cloudv1alpha1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	k8sRestConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: cloudv1alpha1.GroupName, Version: "v1alpha1"}
	k8sRestConfig.APIPath = "/apis"
	k8sRestConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	k8sRestConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	//+kubebuilder:scaffold:scheme
	k8sClient, err = client.New(k8sRestConfig, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	//+kubebuilder:scaffold:scheme

	By("Creating Kubernetes client")
	restClient, err = rest.UnversionedRESTClientFor(k8sRestConfig)
	Expect(err).NotTo(HaveOccurred())
	Expect(restClient).NotTo(BeNil())

	By("Creating manager")
	k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	By("Creating product controller")
	listener = bufconn.Listen(1024 * 1024)
	clientConn, err = grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(getBufDialer(listener)), grpc.WithInsecure())
	billingSyncClient = pb.NewBillingProductCatalogSyncServiceClient(clientConn)
	productReconciler = &productcontroller.ProductReconciler{Client: k8sManager.GetClient(), Scheme: k8sManager.GetScheme(), BillingSyncClient: billingSyncClient}
	productReconciler.SetupWithManager(k8sManager)
	Expect(err).Should(Succeed())

	By("Starting Manager")
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
