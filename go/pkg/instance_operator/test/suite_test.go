// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This test suite utilizes the following components:
//
//   - Kubernetes API Server
//   - etcd (for Kubernetes)
//   - VM Instance Operator
//
// Run interactively with:
// BAZEL_EXTRA_OPTS="--test_output=streamed --test_arg=-test.v --test_arg=-ginkgo.vv --test_env=ZAP_LOG_LEVEL=-127 //go/pkg/instance_operator/..." make test-custom
package test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	vminstancecontroller "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/vm/controllers"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sRestConfig    *rest.Config
	k8sClient        client.Client
	testEnv          *envtest.Environment
	scheme           *runtime.Scheme
	timeout          time.Duration = 30 * time.Second
	poll             time.Duration = 10 * time.Millisecond
	proxyUser        string        = "guest"
	proxyAddress     string        = "ssh.us-dev-1.cloud.intel.com"
	proxyPort        int           = 22
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
			"../../k8s/config/crd/bases",
			"../testdata/crd",
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
	Expect(cloudv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
	Expect(kubevirtv1.AddToScheme(scheme)).NotTo(HaveOccurred())

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

	By("Loading test configuration")
	vmOperatorConfig := NewTestVmInstanceOperatorConfig("../testdata", scheme)

	By("Configuring HTTP client")
	httpClient, err := rest.HTTPClientFor(k8sRestConfig)
	Expect(err).ToNot(HaveOccurred())
	httpClient.Timeout = 5 * time.Second

	By("Creating Mock VNet Private Service Client")
	vNetPrivateClient := NewMockVNetPrivateServiceClient()

	By("Creating Mock VNet Service Client")
	vNetClient := NewMockVNetServiceClient()

	By("Creating VM instance controller")
	_, err = vminstancecontroller.NewVmInstanceReconciler(ctx, k8sManager, vNetPrivateClient, vNetClient, vmOperatorConfig)
	Expect(err).Should(Succeed())

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

func NewMockVNetPrivateServiceClient() pb.VNetPrivateServiceClient {
	mockController := gomock.NewController(GinkgoT())
	vNetClient := pb.NewMockVNetPrivateServiceClient(mockController)
	vNetPrivate := &pb.VNetPrivate{
		Metadata: &pb.VNetPrivate_Metadata{
			CloudAccountId: "6448250923448510",
			Name:           "us-dev-1a-default",
			ResourceId:     "6787226a-2a55-4d6f-bae9-fa2a2ca2450a",
		},
		Spec: &pb.VNetSpecPrivate{
			Region:           "us-dev-1",
			AvailabilityZone: "us-dev-1a",
			Subnet:           "176.16.23.0",
			PrefixLength:     24,
			Gateway:          "172.16.23.1",
			VlanId:           1023,
		},
	}
	vNetReserveAddressResponse := &pb.VNetReserveAddressResponse{
		Address: "172.16.23.3",
	}
	vNetClient.EXPECT().ReserveSubnet(gomock.Any(), gomock.Any()).Return(vNetPrivate, nil).AnyTimes()
	vNetClient.EXPECT().ReleaseSubnet(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).AnyTimes()
	vNetClient.EXPECT().ReserveAddress(gomock.Any(), gomock.Any()).Return(vNetReserveAddressResponse, nil).AnyTimes()
	vNetClient.EXPECT().ReleaseAddress(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).AnyTimes()
	return vNetClient
}

func NewMockVNetServiceClient() pb.VNetServiceClient {
	mockController := gomock.NewController(GinkgoT())
	vNetClient := pb.NewMockVNetServiceClient(mockController)
	vNet := &pb.VNet{
		Metadata: &pb.VNet_Metadata{
			CloudAccountId: "6448250923448510",
			Name:           "us-dev-1a-default",
			ResourceId:     "6787226a-2a55-4d6f-bae9-fa2a2ca2450a",
		},
		Spec: &pb.VNetSpec{
			Region:           "us-dev-1",
			AvailabilityZone: "us-dev-1a",
			PrefixLength:     24,
		},
	}
	vNetClient.EXPECT().Put(gomock.Any(), gomock.Any()).Return(vNet, nil).AnyTimes()
	vNetClient.EXPECT().Get(gomock.Any(), gomock.Any()).Return(vNet, nil).AnyTimes()
	vNetClient.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(&emptypb.Empty{}, nil).AnyTimes()
	return vNetClient
}
