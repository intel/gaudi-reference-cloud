// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dcim "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	mocks "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mocks"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	baremetalv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	validationcontroller "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/baremetal-validation-operator/controllers/metal3.io"
	instancetest "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/test"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient        client.Client
	testEnv          *envtest.Environment
	scheme           *runtime.Scheme
	k8sRestConfig    *rest.Config
	managerStoppable *stoppable.Stoppable
	timeout          time.Duration = 60 * time.Second
	poll             time.Duration = 10 * time.Millisecond
	mockController   *gomock.Controller
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			"../../k8s/config/crd/bases",
			"../testdata/crd",
		},
		ErrorIfCRDPathMissing:    false,
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
	Expect(baremetalv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(k8sRestConfig, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("Loading test configuration")
	bmOperatorConfig := instancetest.NewTestBmInstanceOperatorConfig("../testdata", scheme)

	By("Creating manager")
	gracefulShutdownTimeout := time.Duration(0)
	k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
		Scheme:                  scheme,
		Metrics:                 metricsserver.Options{BindAddress: "0"},
		GracefulShutdownTimeout: &gracefulShutdownTimeout,
	})
	Expect(err).ToNot(HaveOccurred())
	mockController = gomock.NewController(GinkgoT())
	By("Creating Validation controller")
	err = (&validationcontroller.BaremetalhostsReconciler{
		Client:               k8sManager.GetClient(),
		Scheme:               k8sManager.GetScheme(),
		Cfg:                  bmOperatorConfig,
		ComputePrivateClient: NewMockInstanceClient(ctx, mockController),
		ImageFinder:          validationcontroller.NewImageFinder(NewMockImageClient(ctx, mockController), k8sManager.GetClient()),
		EventRecorder:        record.NewFakeRecorder(10),
		NetBoxClient:         NewMockNetboxClient(ctx, mockController),
	}).SetupWithManager(ctx, k8sManager)
	Expect(err).ToNot(HaveOccurred())

	By("Starting manager")
	managerStoppable = stoppable.New(k8sManager.Start)
	managerStoppable.Start(ctx)
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	mockController.Finish()
	By("Stopping manager")
	Expect(managerStoppable.Stop(ctx)).Should(Succeed())
	By("Manager stopped")
	By("Stopping Kubernetes API Server")
	Eventually(func() error {
		return testEnv.Stop()
	}, timeout, poll).ShouldNot(HaveOccurred())
})

func NewMockInstanceClient(ctx context.Context, mockController *gomock.Controller) pb.InstancePrivateServiceClient {
	instancePvtClient := pb.NewMockInstancePrivateServiceClient(mockController)
	resp := &pb.InstanceCreateMultiplePrivateResponse{
		Instances: []*pb.InstancePrivate{
			{
				Metadata: &pb.InstanceMetadataPrivate{
					CloudAccountId: "00000000001",
					Name:           "mockInstance",
				},
				Spec: &pb.InstanceSpecPrivate{
					MachineImage: "ubuntu-22.04-server-cloudimg-amd64-latest",
				},
			},
		},
	}
	instancePvtClient.EXPECT().CreateMultiplePrivate(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()
	return instancePvtClient
}

func NewMockImageClient(ctx context.Context, mockController *gomock.Controller) pb.MachineImageServiceClient {
	machineImageClient := pb.NewMockMachineImageServiceClient(mockController)
	resp := &pb.MachineImageSearchResponse{
		Items: []*pb.MachineImage{
			{
				Metadata: &pb.MachineImage_Metadata{
					Name: "ubuntu-22.04-server-cloudimg-amd64-latest",
				},
			},
		},
	}
	machineImageClient.EXPECT().Search(gomock.Any(), gomock.Any()).Return(resp, nil).AnyTimes()
	machineImage := &pb.MachineImage{
		Metadata: &pb.MachineImage_Metadata{
			Name: "ubuntu-22.04-server-cloudimg-amd64-latest",
		},
		Spec: &pb.MachineImageSpec{
			DisplayName: "Ubuntu 22.04 LTS (Jammy Jellyfish) v20221204",
			UserName:    "ubuntu",
			Components: []*pb.MachineImageComponent{
				{
					Name:        "Ubuntu 22.04 LTS",
					Type:        "OS",
					Version:     "22.04",
					Description: "Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.",
					InfoUrl:     "https://releases.ubuntu.com/jammy",
					ImageUrl:    "https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png",
				},
				{
					Name:        "SynapseAI SW",
					Type:        "Firmware Kit",
					Version:     "1.17.0-495",
					Description: "Designed to facilitate high-performance DL training on Habana Gaudi accelerators.",
					InfoUrl:     "https://docs.habana.ai/en/latest/SW_Stack_Packages_Installation/Synapse_SW_Stack_Installation.html#sw-stack-packages-installation",
				},
			},
		},
	}
	machineImageClient.EXPECT().Get(gomock.Any(), gomock.Any()).Return(machineImage, nil).AnyTimes()
	return machineImageClient
}

func NewMockNetboxClient(ctx context.Context, mockController *gomock.Controller) dcim.DCIM {
	netbox := mocks.NewMockDCIM(mockController)
	netbox.EXPECT().GetDeviceId(gomock.Any(), gomock.Any()).Return(int64(1), nil).AnyTimes()
	netbox.EXPECT().UpdateDeviceCustomFields(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	netbox.EXPECT().UpdateBMValidationStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	return netbox

}
