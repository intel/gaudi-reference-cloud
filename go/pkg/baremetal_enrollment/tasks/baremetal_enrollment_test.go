// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tasks

import (
	"context"
	goError "errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/vault/api"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	testcore "k8s.io/client-go/testing"

	bmcinterface "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mocks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BareMetal Enrollment Task Suite")
}

var Any = gomock.Any()

var _ = Describe("Enrollment Task", func() {
	const (
		newUsername      = "new-admin"
		newPassword      = "new-password"
		availabilityZone = "us-dev-1a"
	)

	var (
		ctx    context.Context
		helper *testHelper

		task       *EnrollmentTask
		deviceData *DeviceData

		mockCtrl *gomock.Controller
		netbox   *mocks.MockDCIM
		vault    *mocks.MockSecretManager
		bmc      *mocks.MockBMCInterface

		clientSet                 *fake.Clientset
		dynamicClient             *fakedynamic.FakeDynamicClient
		instanceTypeServiceClient *pb.MockInstanceTypeServiceClient
	)

	BeforeEach(func() {
		ctx = context.Background()

		deviceData = &DeviceData{
			Name:    "device-1",
			ID:      1,
			Rack:    "test-rack",
			Region:  "us-dev-1",
			Cluster: "1",
		}
		Expect(os.Setenv(dcim.DeviceNameEnvVar, deviceData.Name)).To(Succeed())
		Expect(os.Setenv(dcim.DeviceIdEnvVar, strconv.Itoa(int(deviceData.ID)))).To(Succeed())
		Expect(os.Setenv(dcim.RackNameEnvVar, deviceData.Rack)).To(Succeed())
		Expect(os.Setenv(dcim.RegionEnvVar, deviceData.Region)).To(Succeed())
		Expect(os.Setenv(dcim.ClusterNameEnvVar, deviceData.Cluster)).To(Succeed())
		Expect(os.Setenv(dcim.AvailabilityZoneEnvVar, availabilityZone)).To(Succeed())
		Expect(os.Setenv(util.ProvisioningTimeoutVar, "3600")).To(Succeed())
		Expect(os.Setenv(util.DeprovisionTimeoutVar, "3600")).To(Succeed())

		// create mock objects
		mockCtrl = gomock.NewController(GinkgoT())
		netbox = mocks.NewMockDCIM(mockCtrl)
		vault = mocks.NewMockSecretManager(mockCtrl)
		bmc = mocks.NewMockBMCInterface(mockCtrl)

		// create fake K8s clients
		clientSet = fake.NewSimpleClientset()
		scheme := runtime.NewScheme()
		baremetalv1alpha1.AddToScheme(scheme)
		dynamicClient = fakedynamic.NewSimpleDynamicClient(scheme)
		instanceTypeServiceClient = pb.NewMockInstanceTypeServiceClient(mockCtrl)

		// create test struct
		task = &EnrollmentTask{
			deviceData: deviceData,
			bmcData: &BMCData{
				URL:        "",
				Username:   "admin",
				Password:   "password",
				MACAddress: "A1:B2:C3:D4:F5",
			},
			netBox:                    netbox,
			vault:                     vault,
			bmc:                       bmc,
			clientSet:                 clientSet,
			dynamicClient:             dynamicClient,
			instanceTypeServiceClient: instanceTypeServiceClient,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("preparing for enrollment", func() {
		It("should gather device information", func() {
			Expect(task.deviceData).NotTo(BeNil())
		})
		It("should initialize dependencies", func() {
			By("initializing Vault client")
			Expect(task.vault).NotTo(BeNil())

			By("initializing NetBox client")
			Expect(task.netBox).NotTo(BeNil())

			By("initializing BMC interface")
			Expect(task.bmc).NotTo(BeNil())

			By("initializing K8s client")
			Expect(task.clientSet).NotTo(BeNil())

			By("initializing K8s dynamic client")
			Expect(task.dynamicClient).NotTo(BeNil())

			By("initializing InstanceTypeService client")
			Expect(task.instanceTypeServiceClient).NotTo(BeNil())
		})
	})

	Describe("getting device information", func() {
		It("should succeed when having all device information", func() {
			device, err := getDeviceData(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(device.Name).To(Equal(deviceData.Name))
			Expect(device.ID).To(BeEquivalentTo(deviceData.ID))
			Expect(device.Rack).To(Equal(deviceData.Rack))
			Expect(device.Region).To(Equal(deviceData.Region))
			Expect(device.Cluster).To(Equal(deviceData.Cluster))
		})

		It("should fail when missing device name", func() {
			Expect(os.Unsetenv(dcim.DeviceNameEnvVar)).To(Succeed())
			Expect(getDeviceData(ctx)).Error().To(HaveOccurred())
		})

		It("should fail when missing device ID", func() {
			Expect(os.Unsetenv(dcim.DeviceIdEnvVar)).To(Succeed())
			Expect(getDeviceData(ctx)).Error().To(HaveOccurred())
		})

		It("should fail when missing rack name", func() {
			Expect(os.Unsetenv(dcim.RackNameEnvVar)).To(Succeed())
			Expect(getDeviceData(ctx)).Error().To(HaveOccurred())
		})

		It("should fail when missing region name", func() {
			Expect(os.Unsetenv(dcim.RegionEnvVar)).To(Succeed())
			Expect(getDeviceData(ctx)).Error().To(HaveOccurred())
		})

		It("should fail when missing cluster name", func() {
			Expect(os.Unsetenv(dcim.ClusterNameEnvVar)).To(Succeed())
			Expect(getDeviceData(ctx)).Error().To(HaveOccurred())
		})

		It("should fail when missing AZ name", func() {
			Expect(os.Unsetenv(dcim.AvailabilityZoneEnvVar)).To(Succeed())
		})
	})

	Describe("getting BMC MAC Address from NetBox", func() {
		It("should return a MAC address", func() {
			netbox.EXPECT().GetBMCMACAddress(Any, Any, Any).Return("FF:FF:FF:FF:FF", nil)
			bmcMACAdress, err := task.getBMCMacAddress(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(bmcMACAdress).To(Equal("FF:FF:FF:FF:FF"))
		})
	})

	Describe("creating new BMC credentials in Vault", func() {
		It("should generate BMC credentials", func() {
			Expect(os.Setenv(envBMCEnrollUsername, defaultBMCEnrollUsername)).To(Succeed())
			username, password, err := task.generateBMCCredentials(ctx)
			Expect(username).NotTo(BeEmpty())
			Expect(password).NotTo(BeEmpty())
			Expect(err).NotTo(HaveOccurred())
		})
		It("should store BMC account credentials in Vault", func() {
			vault.EXPECT().PutBMCSecrets(ctx, Any, Any).Return(&api.KVSecret{}, nil)

			By("having a valid MAC address")
			Expect(task.storeUserBMCCredentialsInVault(ctx, newUsername, newPassword)).To(Succeed())

			By("having an invalid MAC address")
			task.bmcData.MACAddress = "not-valid"
			Expect(task.storeUserBMCCredentialsInVault(ctx, newUsername, newPassword)).Error().To(HaveOccurred())
		})
	})

	Describe("updating the admin account credentials in BMC", func() {
		It("should update the account credentials", func() {
			bmc.EXPECT().UpdateAccount(Any, Any, Any).Return(nil)
			Expect(task.updateBMCCredentialsInRedfish(ctx, newUsername, newPassword)).To(Succeed())
			Expect(task.bmcData.Username).To(Equal(newUsername))
			Expect(task.bmcData.Password).To(Equal(newPassword))
		})
		It("should only update the account if BMC is virtual", func() {
			bmc.EXPECT().IsVirtual().Return(true)
			bmc.EXPECT().UpdateAccount(Any, Any, Any).Return(bmcinterface.ErrAccountNotFound)
			Expect(task.updateBMCCredentialsInRedfish(ctx, newUsername, newPassword)).To(Succeed())
		})
		It("should create an account if not exists", func() {
			bmc.EXPECT().IsVirtual().Return(false)
			bmc.EXPECT().UpdateAccount(Any, Any, Any).Return(bmcinterface.ErrAccountNotFound)
			bmc.EXPECT().CreateAccount(Any, Any, Any).Return(nil)
			Expect(task.updateBMCCredentialsInRedfish(ctx, newUsername, newPassword)).To(Succeed())
		})
	})

	Describe("deleting BMC credentials from Vault", func() {
		It("should delete BMC credential from Vault", func() {
			vault.EXPECT().DeleteBMCSecrets(ctx, Any).Return(nil)

			By("having a valid MAC address")
			Expect(task.deleteBMCCredentialsFromVault(ctx, bmcUserSecretsPrefix)).To(Succeed())

			By("having an invalid MAC address")
			task.bmcData.MACAddress = "not-valid"
			Expect(task.deleteBMCCredentialsFromVault(ctx, bmcUserSecretsPrefix)).Error().To(HaveOccurred())
		})
	})

	Describe("getting boot MAC address from BMC", func() {
		It("should return a boot Mac Address", func() {
			bmc.EXPECT().GetHostMACAddress(ctx).Return("A1:B2:C3:D4:F5", nil)
			macAddress, err := task.getBootMacAddress(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(macAddress).NotTo(BeEmpty())
		})

		It("should fail when getting BMC returns an empty one", func() {
			bmc.EXPECT().GetHostMACAddress(ctx).Return("", nil)
			netbox.EXPECT().GetBMCMACAddress(Any, Any, Any).Return("", nil)
			macAddress, err := task.getBootMacAddress(ctx)
			Expect(err).To(HaveOccurred())
			Expect(macAddress).To(BeEmpty())
		})

		It("should succeed when getting BMC returns an empty one but netbox returns a valid MAC", func() {
			bmc.EXPECT().GetHostMACAddress(ctx).Return("", nil)
			netbox.EXPECT().GetBMCMACAddress(Any, Any, Any).Return("A1:B2:C3:D4:F5", nil)
			macAddress, err := task.getBootMacAddress(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(macAddress).NotTo(BeEmpty())
		})

		It("should return a failure if netbox returns an error", func() {
			bmc.EXPECT().GetHostMACAddress(ctx).Return("", nil)
			netbox.EXPECT().GetBMCMACAddress(Any, Any, Any).Return("", goError.New("failed to read interface"))
			_, err := task.getBootMacAddress(ctx)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("selecting a namespace for BareMetalHost", func() {
		BeforeEach(func() {
			Expect(helper.createNamespace(ctx, clientSet, "ns-1")).Error().NotTo(HaveOccurred())
			Expect(helper.createNamespace(ctx, clientSet, "ns-2")).Error().NotTo(HaveOccurred())
			Expect(helper.createBareMetalHost(ctx, dynamicClient, "bm-1", "ns-1")).Error().NotTo(HaveOccurred())
			Expect(helper.createBareMetalHost(ctx, dynamicClient, "bm-2", "ns-2")).Error().NotTo(HaveOccurred())
		})

		When("the device's enrollment namespace is set", func() {
			It("should select the device namespace first", func() {
				netbox.EXPECT().GetDeviceNamespace(ctx, deviceData.Name).Return("ns-1", nil)
				selectedNamespace, err := task.getBareMetalHostNamespace(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(selectedNamespace.Name).To(Equal("ns-1"))
				Expect(selectedNamespace.Labels).To(HaveKeyWithValue(Metal3NamespaceSelectorKey, "true"))
			})
			It("should select the device namespace first", func() {
				netbox.EXPECT().GetDeviceNamespace(ctx, deviceData.Name).Return("ns-2", nil)
				selectedNamespace, err := task.getBareMetalHostNamespace(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(selectedNamespace.Name).To(Equal("ns-2"))
				Expect(selectedNamespace.Labels).To(HaveKeyWithValue(Metal3NamespaceSelectorKey, "true"))
			})
			It("should fail if unable to get the device's custom field", func() {
				netbox.EXPECT().GetDeviceNamespace(ctx, deviceData.Name).Return("", fmt.Errorf("failed to get device's custom field"))
				selectedNamespace, err := task.getBareMetalHostNamespace(ctx)
				Expect(err).To(HaveOccurred())
				Expect(selectedNamespace).To(BeNil())
			})
		})

		When("the number of hosts are equal", func() {
			It("should select the first namespace in lexical order", func() {
				netbox.EXPECT().GetDeviceNamespace(ctx, deviceData.Name).Return("", nil)
				selectedNamespace, err := task.getBareMetalHostNamespace(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(selectedNamespace.Name).To(Equal("ns-1"))
				Expect(selectedNamespace.Labels).To(HaveKeyWithValue(Metal3NamespaceSelectorKey, "true"))
			})
		})

		When("the number of hosts differ", func() {
			It("should select the namespace with the lowest number of hosts", func() {
				netbox.EXPECT().GetDeviceNamespace(ctx, deviceData.Name).Return("", nil)
				// make ns-2 smaller than ns-1
				Expect(helper.createBareMetalHost(ctx, dynamicClient, "bm-3", "ns-1")).Error().NotTo(HaveOccurred())
				selectedNamespace, err := task.getBareMetalHostNamespace(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(selectedNamespace.Name).To(Equal("ns-2"))
				Expect(selectedNamespace.Labels).To(HaveKeyWithValue(Metal3NamespaceSelectorKey, "true"))
			})
		})

		AfterEach(func() {
			Expect(helper.deleteAllBareMetalHosts(ctx, dynamicClient)).To(Succeed())
			Expect(helper.deleteAllNamespaces(ctx, clientSet)).To(Succeed())
		})
	})

	Describe("registering a new BareMetalHost", FlakeAttempts(2), func() {
		var (
			cpuInfo         *bmcinterface.CPUInfo
			hardwareDetails *baremetalv1alpha1.HardwareDetails
			watcher         *watch.FakeWatcher
			bmh             *baremetalv1alpha1.BareMetalHost
			secretName      string
			err             error
		)

		BeforeEach(func() {
			By("getting BMC's MAC address")
			bmc.EXPECT().GetHostBMCAddress().Return("redfish+http://10.11.12.13:8001/redfish/v1/Systems/1", nil)

			By("getting host's Boot MAC address")
			bmc.EXPECT().GetHostMACAddress(ctx).Return("A1:B2:C3:D4:F5", nil)

			By("getting host's namespace")
			ns, err := helper.createNamespace(ctx, clientSet, "ns-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(ns).NotTo(BeNil())
			task.bmHostNamespace = ns

			secretName = deviceData.Name + "-bmc-secret"

			// set the hardware spec
			bmc.EXPECT().IsVirtual().AnyTimes().Return(true)

			// Get cluster size from netbox
			netbox.EXPECT().GetClusterSize(Any, Any, Any).Return(int64(2), nil).AnyTimes()

			// Get cluster network mode from netbox
			netbox.EXPECT().GetClusterNetworkMode(Any, Any, Any).Return("", nil).AnyTimes()

			cpuInfo = &bmcinterface.CPUInfo{
				Manufacturer: "Intel",
				CPUID:        "0x000000",
				Sockets:      2,
				Cores:        24,
				Threads:      1,
			}
			hardwareDetails = &baremetalv1alpha1.HardwareDetails{
				CPU: baremetalv1alpha1.CPU{
					Model: "Intel Xeon",
					Count: 48,
				},
				NIC: []baremetalv1alpha1.NIC{
					{
						MAC: "A1:B2:C3:D4:F5",
						IP:  "1.1.1.1",
					},
				},
				RAMMebibytes: 16384,
			}

			// use fake watcher to simulate BareMetalHost events
			watcher = watch.NewFake()
			DeferCleanup(watcher.Stop)
			dynamicClient.PrependWatchReactor("baremetalhosts", testcore.DefaultWatchReactor(watcher, nil))
		})

		When("BareMetalHost is available", func() {
			BeforeEach(func() {
				go func() {
					By("waiting for BareMetalHost events")
					time.Sleep(100 * time.Millisecond)

					// make the current host provisioned
					host, err := helper.getBareMetalHost(ctx, dynamicClient, deviceData.Name, task.bmHostNamespace.Name)
					Expect(err).To(BeNil())
					host.Status.Provisioning.State = baremetalv1alpha1.StateProvisioned

					// update status
					Expect(helper.updateBareMetalHostStatus(ctx, dynamicClient, host)).Error().NotTo(HaveOccurred())

					By("having BareMetalHost in provisioned state")

					// send modified event
					watcher.Modify(host)

					watcher = watch.NewFake()
					DeferCleanup(watcher.Stop)
					dynamicClient.PrependWatchReactor("baremetalhosts", testcore.DefaultWatchReactor(watcher, nil))

					host, err = helper.getBareMetalHost(ctx, dynamicClient, deviceData.Name, task.bmHostNamespace.Name)
					Expect(err).To(BeNil())

					host.Status.Provisioning.State = baremetalv1alpha1.StateAvailable

					// update host hardware information
					host.Status.HardwareDetails = hardwareDetails
					bmc.EXPECT().GetHostCPU(ctx).Return(cpuInfo, nil)

					// Expect GPUDiscover
					bmc.EXPECT().GPUDiscovery(ctx).Return(5, "gpu-model-name", nil)

					// Expect for HBMDiscovery
					bmc.EXPECT().HBMDiscovery(ctx).Return("hbm-mode-name", nil)

					// Expect for GetHwType to be called
					bmc.EXPECT().GetHwType().Return(bmcinterface.Virtual).AnyTimes()

					// update status
					Expect(helper.updateBareMetalHostStatus(ctx, dynamicClient, host)).Error().NotTo(HaveOccurred())

					// call isVirtual
					bmc.EXPECT().IsVirtual().AnyTimes().Return(true)

					instanceTypeSearchResponse := &pb.InstanceTypeSearchResponse{
						Items: []*pb.InstanceType{
							{
								Metadata: &pb.InstanceType_Metadata{
									Name: "bm-virtual-not-matching",
								},
								Spec: &pb.InstanceTypeSpec{
									Name:             "bm-virtual-not-matching",
									InstanceCategory: pb.InstanceCategory_BareMetalHost,
									Cpu: &pb.CpuSpec{
										Cores:   int32(cpuInfo.Cores),
										Id:      "0x010101",
										Sockets: int32(cpuInfo.Sockets),
										Threads: int32(cpuInfo.Threads),
									},
									Gpu: &pb.GpuSpec{
										ModelName: "gpu-model-name",
										Count:     int32(5),
									},
									HbmMode: "hbm-mode-name",
									Memory: &pb.MemorySpec{
										Size: "6Gi",
									},
								},
							},
							{
								Metadata: &pb.InstanceType_Metadata{
									Name: "bm-virtual-matching",
								},
								Spec: &pb.InstanceTypeSpec{
									Name:             "bm-virtual-matching",
									InstanceCategory: pb.InstanceCategory_BareMetalHost,
									Cpu: &pb.CpuSpec{
										Cores:   int32(cpuInfo.Cores),
										Id:      cpuInfo.CPUID,
										Sockets: int32(cpuInfo.Sockets),
										Threads: int32(cpuInfo.Threads),
									},
									Gpu: &pb.GpuSpec{
										ModelName: "gpu-model-name",
										Count:     int32(5),
									},
									HbmMode: "hbm-mode-name",
									Memory: &pb.MemorySpec{
										Size: "16Gi",
									},
								},
							},
						},
					}

					instanceTypeServiceClient.EXPECT().Search(ctx, &pb.InstanceTypeSearchRequest{}).Return(instanceTypeSearchResponse, nil)
					By("having BareMetalHost in provisioned state")

					// Get cluster size from netbox
					netbox.EXPECT().GetClusterSize(Any, Any, Any).Return(int64(2), nil).AnyTimes()

					// send modified event
					watcher.Modify(host)
				}()

				By("registering a new host")
				netbox.EXPECT().UpdateDeviceCustomFields(Any, Any, Any, Any).Return(nil).AnyTimes()
				Eventually(task.registerBareMetalHost(ctx)).Within(3 * time.Second).Should(Succeed())
			})

			It("should have BareMetalHost in available state", func() {
				bmh, err = helper.getBareMetalHost(ctx, dynamicClient, deviceData.Name, task.bmHostNamespace.Name)
				Expect(err).NotTo(HaveOccurred())
				Expect(bmh.Status.Provisioning.State).To(Equal(baremetalv1alpha1.StateAvailable))
			})

			It("should have created BareMetalHost", func() {
				bmh, err = helper.getBareMetalHost(ctx, dynamicClient, deviceData.Name, task.bmHostNamespace.Name)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should have created a Secret owned by BareMetalHost", func() {
				secret, err := clientSet.CoreV1().Secrets(bmh.Namespace).Get(ctx, bmh.Spec.BMC.CredentialsName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(secret.OwnerReferences).NotTo(BeEmpty())
				Expect(secret.OwnerReferences[0].UID).To(Equal(bmh.GetUID()))
			})

			It("should have added hardware labels and annotations to BareMetalHost", func() {
				Expect(bmh.GetLabels()).To(HaveKeyWithValue(CPUManufacturerLabel, cpuInfo.Manufacturer))
				Expect(bmh.GetLabels()).To(HaveKeyWithValue(CPUIDLabel, cpuInfo.CPUID))
				Expect(bmh.GetLabels()).To(HaveKeyWithValue(CPUSocketsLabel, strconv.Itoa(cpuInfo.Sockets)))
				Expect(bmh.GetLabels()).To(HaveKeyWithValue(CPUCoresLabel, strconv.Itoa(cpuInfo.Cores)))
				Expect(bmh.GetLabels()).To(HaveKeyWithValue(CPUThreadsLabel, strconv.Itoa(cpuInfo.Threads)))
				Expect(bmh.GetLabels()).To(HaveKeyWithValue(CPUCountLabel, strconv.Itoa(hardwareDetails.CPU.Count)))
				Expect(bmh.GetLabels()).To(HaveKeyWithValue(MemorySizeLabel, "16Gi"))
				Expect(bmh.GetLabels()).To(HaveKeyWithValue(fmt.Sprintf(InstanceTypeLabel, "bm-virtual-matching"), "true"))
				Expect(bmh.GetLabels()).To(HaveKeyWithValue(fmt.Sprintf(ComputeNodePoolLabel, NodePoolGeneral), "true"))
				Expect(bmh.GetAnnotations()).To(HaveKeyWithValue(CPUModelLabel, hardwareDetails.CPU.Model))
			})
		})

		When("BareMetalHost has error", func() {
			BeforeEach(func() {
				go func() {
					By("waiting for BareMetalHost events")
					time.Sleep(100 * time.Millisecond)

					// update host with operational error
					host, err := helper.getBareMetalHost(ctx, dynamicClient, task.deviceData.Name, task.bmHostNamespace.Name)
					Expect(err).To(BeNil())
					host.Status.OperationalStatus = baremetalv1alpha1.OperationalStatusError
					host.Status.ErrorType = baremetalv1alpha1.InspectionError
					host.Status.ErrorCount = 1

					// update status
					Expect(helper.updateBareMetalHostStatus(ctx, dynamicClient, host)).Error().NotTo(HaveOccurred())

					By("having BareMetalHost updated with error")

					// send modified event
					watcher.Modify(host)
				}()

				By("registering a new host")
				netbox.EXPECT().UpdateDeviceCustomFields(Any, Any, Any, Any).Return(nil).AnyTimes()
				Eventually(task.registerBareMetalHost(ctx)).Within(3 * time.Second).ShouldNot(Succeed())
			})

			It("should have deleted BareMetalHost", func() {
				_, err := helper.getBareMetalHost(ctx, dynamicClient, task.deviceData.Name, task.bmHostNamespace.Name)
				Expect(errors.IsNotFound(err)).To(BeTrue())
			})

			It("should have deleted BareMetalHost's Secret", func() {
				_, err := clientSet.CoreV1().Secrets(task.bmHostNamespace.Namespace).Get(ctx, secretName, metav1.GetOptions{})
				Expect(errors.IsNotFound(err)).To(BeTrue())
			})
		})

		AfterEach(func() {
			Expect(helper.deleteAllBareMetalHosts(ctx, dynamicClient)).To(Succeed())
			Expect(helper.deleteAllSecrets(ctx, clientSet, task.bmHostNamespace.Name)).To(Succeed())
			Expect(helper.deleteAllNamespaces(ctx, clientSet)).To(Succeed())
		})
	})
})

type testHelper struct{}

func (h *testHelper) createNamespace(ctx context.Context, client kubernetes.Interface, name string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				Metal3NamespaceSelectorKey: "true",
				metal3NamespaceIronicIPKey: "10.11.12.13",
			},
		},
	}
	return client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
}

func (h *testHelper) createBareMetalHost(ctx context.Context, client dynamic.Interface, name string, namespace string) (*baremetalv1alpha1.BareMetalHost, error) {
	bmh := &baremetalv1alpha1.BareMetalHost{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "metal3.io/v1alpha1",
			Kind:       "BareMetalHost",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: baremetalv1alpha1.BareMetalHostSpec{
			Online:         true,
			BootMode:       baremetalv1alpha1.Legacy,
			BootMACAddress: "a1:b2:c3:d4:f5",
			BMC: baremetalv1alpha1.BMCDetails{
				Address:                        "redfish+http://10.11.12.13:8001/redfish/v1/Systems/1",
				CredentialsName:                "secret-1",
				DisableCertificateVerification: true,
			},
			RootDeviceHints: &baremetalv1alpha1.RootDeviceHints{
				DeviceName: "/dev/vda",
			},
		},
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(bmh)
	if err != nil {
		return nil, fmt.Errorf("unable to convert %s to unstructured object: %v", bmh.Name, err)
	}

	createdObj, err := client.Resource(bmHostGVR).Namespace(namespace).Create(ctx, &unstructured.Unstructured{Object: u}, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to create BareMetalHost %q: %v", bmh.Name, err)
	}

	newHost := &baremetalv1alpha1.BareMetalHost{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(createdObj.UnstructuredContent(), bmh); err != nil {
		return nil, fmt.Errorf("unable to decode BareMetalHost object")
	}

	return newHost, nil
}

func (h *testHelper) updateBareMetalHost(ctx context.Context, client dynamic.Interface, bmh *baremetalv1alpha1.BareMetalHost) (*baremetalv1alpha1.BareMetalHost, error) {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(bmh)
	if err != nil {
		return nil, fmt.Errorf("unable to convert %s to unstructured object: %v", bmh.Name, err)
	}

	updatedObj, err := client.Resource(bmHostGVR).Namespace(bmh.Namespace).Update(ctx, &unstructured.Unstructured{Object: u}, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to update BareMetalHost %q: %v", bmh.Name, err)
	}

	updatedHost := &baremetalv1alpha1.BareMetalHost{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(updatedObj.UnstructuredContent(), bmh); err != nil {
		return nil, fmt.Errorf("unable to decode BareMetalHost object")
	}

	return updatedHost, nil
}

func (h *testHelper) updateBareMetalHostStatus(ctx context.Context, client dynamic.Interface, bmh *baremetalv1alpha1.BareMetalHost) (*baremetalv1alpha1.BareMetalHost, error) {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(bmh)
	if err != nil {
		return nil, fmt.Errorf("unable to convert %s to unstructured object: %v", bmh.Name, err)
	}

	updatedObj, err := client.Resource(bmHostGVR).Namespace(bmh.Namespace).UpdateStatus(ctx, &unstructured.Unstructured{Object: u}, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to update BareMetalHost %q: %v", bmh.Name, err)
	}

	updatedHost := &baremetalv1alpha1.BareMetalHost{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(updatedObj.UnstructuredContent(), bmh); err != nil {
		return nil, fmt.Errorf("unable to decode BareMetalHost object")
	}

	return updatedHost, nil
}

func (h *testHelper) getBareMetalHost(ctx context.Context, client dynamic.Interface, name string, namespace string) (*baremetalv1alpha1.BareMetalHost, error) {
	u, err := client.Resource(bmHostGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get BareMetalHost %q: %w", name, err)
	}

	bmh := &baremetalv1alpha1.BareMetalHost{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), bmh); err != nil {
		return nil, fmt.Errorf("unable to decode BareMetalHost object")
	}

	return bmh, nil
}

func (t *testHelper) deleteBareMetalHost(ctx context.Context, client dynamic.Interface, name string, namespace string) error {
	if err := client.Resource(bmHostGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (h *testHelper) deleteAllNamespaces(ctx context.Context, client kubernetes.Interface) error {
	list, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, ns := range list.Items {
		if err := client.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (h *testHelper) deleteAllSecrets(ctx context.Context, client kubernetes.Interface, namespace string) error {
	list, err := client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, secret := range list.Items {
		if err := client.CoreV1().Secrets(namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (h *testHelper) deleteAllBareMetalHosts(ctx context.Context, client dynamic.Interface) error {
	list, err := client.Resource(bmHostGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, bmh := range list.Items {
		err := client.Resource(bmHostGVR).Namespace(bmh.GetNamespace()).Delete(ctx, bmh.GetName(), metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func Test_oldestVersion(t *testing.T) {
	type args struct {
		v string
		w string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "v is older",
			args: args{
				v: "1.2",
				w: "1.3",
			},
			want: "1.2",
		},
		{
			name: "w is older",
			args: args{
				v: "1.2",
				w: "1.1",
			},
			want: "1.1",
		},
		{
			name: "v and w are same",
			args: args{
				v: "1.2",
				w: "1.2",
			},
			want: "1.2",
		},
		{
			name: "v and w have leading zeros",
			args: args{
				v: "01.2.0",
				w: "1.02",
			},
			want: "1.02",
		},
		{
			name: "v and w have leading zeros",
			args: args{
				v: "01.2.0",
				w: "1.01",
			},
			want: "1.01",
		},
		{
			name: "v and w have leading v",
			args: args{
				v: "v01.2",
				w: "1.02",
			},
			want: "1.02",
		},
		{
			name: "v and w have leading V",
			args: args{
				v: "V01.2",
				w: "V01.02",
			},
			want: "V01.02",
		},
		{
			name: "v and w have  Date string",
			args: args{
				v: "10/12/22 V01.2",
				w: "10/13/23 V01.02",
			},
			want: "10/13/23 V01.02",
		},
		{
			name: "v and w have  non numerical values",
			args: args{
				v: "10/12/22 V01.2",
				w: "10/13/23 V01.z2.1",
			},
			want: "10/13/23 V01.z2.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oldestVersion(tt.args.v, tt.args.w); got != tt.want {
				t.Errorf("oldestVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_meetsMiniumFWRequirements(t *testing.T) {
	type args struct {
		firmwarePresent          map[string]string
		minimumFirmwareSupported map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "all present firmware listed under supported firmware are equal or higher",
			args: args{
				firmwarePresent:          map[string]string{"fw120": "1.2.0", "fw130": "1.3.0"},
				minimumFirmwareSupported: map[string]string{"fw120": "1.2.0", "fw130": "1.3.0", "fw140": "1.4.0"},
			},

			want: true,
		},
		{
			name: "none of the present firmware are listed under supported firmware",
			args: args{
				firmwarePresent:          map[string]string{"fw150": "0.2.0", "fw160": "0.3.0"},
				minimumFirmwareSupported: map[string]string{"fw120": "1.2.0", "fw130": "1.3.0", "fw140": "1.4.0"},
			},

			want: true,
		},
		{name: "at least one of the present firmware is below the listed supported firmware versio",
			args: args{
				firmwarePresent:          map[string]string{"fw120": "1.1.0", "fw130": "1.3.0"},
				minimumFirmwareSupported: map[string]string{"fw120": "1.2.0", "fw130": "1.3.0", "fw140": "1.4.0"},
			},
			want: false,
		},
		{name: "list of present firmware is empty",
			args: args{
				firmwarePresent:          map[string]string{},
				minimumFirmwareSupported: map[string]string{"fw120": "1.2.0", "fw130": "1.3.0", "fw140": "1.4.0"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := meetsMiniumFWRequirements(context.TODO(), tt.args.firmwarePresent, tt.args.minimumFirmwareSupported)
			if got != tt.want {
				t.Errorf("meetsMiniumFWRequirements() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnrollmentTask_CurrentFirmwareVersions(t1 *testing.T) {

	// create mock objects
	mockCtrl := gomock.NewController(t1)
	defer mockCtrl.Finish()

	emptyInventory := []*redfish.SoftwareInventory{}
	populatedInvetory := []*redfish.SoftwareInventory{
		&redfish.SoftwareInventory{
			Entity: common.Entity{
				ID: "BIOS",
			},
			Version: "1.0.0",
		},
		&redfish.SoftwareInventory{
			Entity: common.Entity{
				ID: "BMC",
			},
			Version: "1.0.0",
		},
	}
	populatedInventoryMap := map[string]string{"BIOS": "1.0.0", "BMC": "1.0.0"}

	halfPopulatedInvetory := []*redfish.SoftwareInventory{
		&redfish.SoftwareInventory{
			Entity: common.Entity{
				ID: "BIOS",
			},
			Version: "1.0.0",
		},
	}
	halfPopulatedInventoryMap := map[string]string{"BIOS": "1.0.0"}

	type fields struct {
		bmc bmcinterface.Interface
	}
	tests := []struct {
		name     string
		fields   fields
		want     map[string]string
		wantErr  bool
		mockfunc func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
			s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor)
	}{
		{
			name: "Fails to get bmc client",
			want: nil,
			mockfunc: func(m *mocks.MockBMCInterface, _ *mocks.MockGoFishClientAccessor,
				_ *mocks.MockGoFishServiceAccessor, _ *mocks.MockGoFishUpdateServiceAccessor) {
				m.EXPECT().GetClient()
				m.EXPECT().IsVirtual().Return(false)
			},
			wantErr: true,
		},
		{
			name: "Fails to get bmc client service",
			want: nil,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				_ *mocks.MockGoFishServiceAccessor, _ *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService()
				m.EXPECT().IsVirtual().Return(false)
			},
			wantErr: true,
		},
		{
			name: "Fails to get updated service",
			want: nil,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, _ *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService()
			},
			wantErr: true,
		},
		{
			name: "Fails to get updated service nil, nil",
			want: nil,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, _ *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "Fails to get updated service with nil, error",
			want: nil,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(nil, fmt.Errorf("fake error"))
			},
			wantErr: true,
		},
		{
			name: "Fails to get updated service with error",
			want: nil,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(u, fmt.Errorf("fake error"))
			},
			wantErr: true,
		},
		{
			name: "Fails to get inventory",
			want: nil,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(u, nil)
				u.EXPECT().FirmwareInventories()
			},
			wantErr: true,
		},
		{
			name: "Fails to get inventory with nil, error",
			want: nil,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(u, nil)
				u.EXPECT().FirmwareInventories().Return(nil, fmt.Errorf("fake error"))
			},
			wantErr: true,
		},
		{
			name: "Fails to get inventory with error",
			want: nil,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(u, nil)
				u.EXPECT().FirmwareInventories().Return(emptyInventory, fmt.Errorf("fake error"))
			},
			wantErr: true,
		},
		{
			name: "Successfully returns inventory",
			want: populatedInventoryMap,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(u, nil)
				u.EXPECT().FirmwareInventories().Return(populatedInvetory, nil)
			},
			wantErr: false,
		},
		{
			name: "Successfully returns half inventory",
			want: halfPopulatedInventoryMap,
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor) {

				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(u, nil)
				u.EXPECT().FirmwareInventories().Return(halfPopulatedInvetory, nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			mockBmc := mocks.NewMockBMCInterface(mockCtrl)
			mockClient := mocks.NewMockGoFishClientAccessor(mockCtrl)
			mockService := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			mockUpdateService := mocks.NewMockGoFishUpdateServiceAccessor(mockCtrl)
			if tt.mockfunc != nil {
				tt.mockfunc(mockBmc, mockClient, mockService, mockUpdateService)
			}
			t := &EnrollmentTask{
				bmc: mockBmc,
			}

			got, err := t.CurrentFirmwareVersions()
			if (err != nil) != tt.wantErr {
				t1.Errorf("CurrentFirmwareVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("CurrentFirmwareVersions() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnrollmentTask_MinimumFirmwareVersionsSupported(t1 *testing.T) {
	clientSet := fake.NewSimpleClientset()
	namespaceName := "bmhostnamespace"
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespaceName,
			Labels: map[string]string{Metal3NamespaceSelectorKey: "true"},
		},
	}
	_, err := clientSet.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		t1.Errorf("Could not create namespace for testing: %v", err)
		return
	}
	wrongns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "badNamespace",
		},
	}
	_, err = clientSet.CoreV1().Namespaces().Create(context.TODO(), wrongns, metav1.CreateOptions{})
	if err != nil {
		t1.Errorf("Could not create wrong namespace for testing: %v", err)
		return
	}
	goodConfigMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      minFwConfigMapName,
			Namespace: namespaceName,
		},
		Data: map[string]string{minFwConfigMapName + ".json": "{\"hwTypeMinFirmware\": {\"Gaudi2Smc\": {\"BIOS\": \"1.0.0\",\"BMC\": \"1.1.0\" }, \"Virtual\": {\"BIOS\": \"1.0.0\",\"BMC\": \"1.0\" } }  } "},
	}
	goodFwMap := map[string]string{"BIOS": "1.0.0", "BMC": "1.1.0"}
	_, err = clientSet.CoreV1().ConfigMaps(namespaceName).Create(context.TODO(), &goodConfigMap, metav1.CreateOptions{})
	if err != nil {
		t1.Errorf("Could not create config map for testing: %v", err)
		return
	}
	type fields struct {
		bmHostNamespace *corev1.Namespace
	}
	type args struct {
		log logr.Logger
	}

	// create mock objects
	mockCtrl := gomock.NewController(t1)
	defer mockCtrl.Finish()

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     map[string]string
		wantErr  bool
		mockfunc func(m *mocks.MockBMCInterface, d *mocks.MockDCIM)
	}{
		{
			name: "Fails to get config map",
			want: nil,
			fields: fields{
				bmHostNamespace: wrongns,
			},
			wantErr: true,
			mockfunc: func(m *mocks.MockBMCInterface, d *mocks.MockDCIM) {
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Smc)
				//d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
				//	wrongns.Name, nil)
			},
		},
		{
			name: "Succefully gets config map",
			want: goodFwMap,
			fields: fields{
				bmHostNamespace: ns,
			},
			wantErr: false,
			mockfunc: func(m *mocks.MockBMCInterface, d *mocks.MockDCIM) {
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Smc)
				//d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
				//	ns.Name, nil)
			},
		},
		{
			name: "Succefully gets config map, but no match",
			want: map[string]string{},
			fields: fields{
				bmHostNamespace: ns,
			},
			wantErr: false,
			mockfunc: func(m *mocks.MockBMCInterface, d *mocks.MockDCIM) {
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Wiwynn)
				//d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
				//	ns.Name, nil)
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			mockBmc := mocks.NewMockBMCInterface(mockCtrl)
			mockDCMI := mocks.NewMockDCIM(mockCtrl)
			if tt.mockfunc != nil {
				tt.mockfunc(mockBmc, mockDCMI)
			}
			t := &EnrollmentTask{
				bmHostNamespace: tt.fields.bmHostNamespace,
				clientSet:       clientSet,
				bmc:             mockBmc,
				deviceData:      &DeviceData{Name: "mocked-dev-1"},
				netBox:          mockDCMI,
			}
			got, err := t.MinFwVersionSupported(context.TODO())
			if (err != nil) != tt.wantErr {
				t1.Errorf("MinFwVersionSupported() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("MinFwVersionSupported() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnrollmentTask_checkMinFWVersions(t1 *testing.T) {

	// create mock objects
	mockCtrl := gomock.NewController(t1)
	defer mockCtrl.Finish()
	emptyInventory := []*redfish.SoftwareInventory{}
	supportedInventory := []*redfish.SoftwareInventory{&redfish.SoftwareInventory{
		Entity: common.Entity{
			ID: "BIOS",
		},
		Version: "1.0.0",
	}}
	unsupportedInventory := []*redfish.SoftwareInventory{&redfish.SoftwareInventory{
		Entity: common.Entity{
			ID: "BIOS",
		},
		Version: "0.9.0",
	}}
	clientSet := fake.NewSimpleClientset()
	namespaceName := "bmhostnamespace"
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespaceName,
			Labels: map[string]string{Metal3NamespaceSelectorKey: "true"},
		},
	}
	_, err := clientSet.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		t1.Errorf("Could not create namespace for testing: %v", err)
		return
	}
	wrongns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "badNamespace",
		},
	}
	_, err = clientSet.CoreV1().Namespaces().Create(context.TODO(), wrongns, metav1.CreateOptions{})
	if err != nil {
		t1.Errorf("Could not create wrong namespace for testing: %v", err)
		return
	}
	supportedFwConfigMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      minFwConfigMapName,
			Namespace: namespaceName,
		},
		Data: map[string]string{minFwConfigMapName + ".json": "{\"hwTypeMinFirmware\": {\"Gaudi2Smc\": {\"BIOS\": \"1.0.0\",\"BMC\": \"1.1.0\" }, \"Virtual\": {\"BIOS\": \"1.0.0\",\"BMC\": \"1.0\" } }  } "},
	}
	_, err = clientSet.CoreV1().ConfigMaps(namespaceName).Create(context.TODO(), &supportedFwConfigMap, metav1.CreateOptions{})
	if err != nil {
		t1.Errorf("Could not create config map for testing: %v", err)
		return
	}
	type fields struct {
		bmHostNamespace *corev1.Namespace
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		want     bool
		mockfunc func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
			s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor,
			d *mocks.MockDCIM)
	}{
		{
			name: "MinFwVersionSupported returns error",
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor,
				d *mocks.MockDCIM) {

				//CurrentFirmwareVersion = rror
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Smc)
				d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
					"", fmt.Errorf("fake error"))
				// MinFwVersionSupported = error
				// use wrong namespace

			},
			wantErr: true,
			want:    false,
		},
		{
			name: "MinimumFwVersionSupported returns empty list",
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor,
				d *mocks.MockDCIM) {

				// MinFwVersionSupported returns non-empty list
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Wiwynn)

				// Called made by checkMinFWVersions
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Wiwynn)
				//d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
				//	ns.Name, nil)
			},
			fields:  fields{bmHostNamespace: ns},
			wantErr: false,
			want:    true,
		},
		{
			name: "CurrentFirmwareVersion returns error",
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor,
				d *mocks.MockDCIM) {

				// MinFwVersionSupported returns non-empty list
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Smc)
				//d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
				//	ns.Name, nil)

				//CurrentFirmwareVersion returns error
				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(nil, fmt.Errorf("fake error"))

			},
			fields:  fields{bmHostNamespace: ns},
			wantErr: true,
			want:    false,
		},

		{
			name: "CurrentFirmwareVersion returns error",
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor,
				d *mocks.MockDCIM) {

				// MinFwVersionSupported returns non-empty list
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Smc)
				//d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
				//	ns.Name, nil)

				//CurrentFirmwareVersion returns error
				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(nil, fmt.Errorf("fake error"))

			},
			fields:  fields{bmHostNamespace: ns},
			wantErr: true,
			want:    false,
		},

		{
			name: "CurrentFirmwareVersion returns empty",
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor,
				d *mocks.MockDCIM) {

				// MinFwVersionSupported returns non-empty list
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Smc)
				d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
					ns.Name, nil)

				//CurrentFirmwareVersion returns empty list
				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(u, nil)
				u.EXPECT().FirmwareInventories().Return(emptyInventory, nil)
			},
			wantErr: false,
			want:    false,
		},

		{
			name: "does not meet MiniumFWRequirements",
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor,
				d *mocks.MockDCIM) {

				// MinFwVersionSupported returns non-empty list
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Smc)
				//d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
				//	ns.Name, nil)

				//CurrentFirmwareVersion returns empty list
				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(u, nil)
				u.EXPECT().FirmwareInventories().Return(unsupportedInventory, nil)

				// MinFwVersionSupported = no error
				// use appropriate namespace

			},
			fields: fields{
				bmHostNamespace: ns,
			},
			wantErr: false,
			want:    false,
		},
		{
			name: "meets MiniumFWRequirements",
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor,
				d *mocks.MockDCIM) {

				// MinFwVersionSupported returns non-empty list
				m.EXPECT().GetHwType().Return(bmcinterface.Gaudi2Smc)
				//d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
				//	ns.Name, nil)

				//CurrentFirmwareVersion returns empty list
				m.EXPECT().IsVirtual().Return(false)
				m.EXPECT().GetClient().Return(c)
				c.EXPECT().GetService().Return(s)
				s.EXPECT().UpdateService().Return(u, nil)
				u.EXPECT().FirmwareInventories().Return(supportedInventory, nil)

				// MinFwVersionSupported = no error
				// use appropriate namespace

			},
			fields: fields{
				bmHostNamespace: ns,
			},
			wantErr: false,
			want:    true,
		},

		{
			name: "Host is virtual ",
			mockfunc: func(m *mocks.MockBMCInterface, c *mocks.MockGoFishClientAccessor,
				s *mocks.MockGoFishServiceAccessor, u *mocks.MockGoFishUpdateServiceAccessor,
				d *mocks.MockDCIM) {

				// MinFwVersionSupported returns non-empty list
				m.EXPECT().GetHwType().Return(bmcinterface.Virtual)
				//d.EXPECT().GetDeviceNamespace(gomock.Any(), gomock.Any()).Return(
				//	ns.Name, nil)
				//CurrentFirmwareVersion returns vritual
				m.EXPECT().IsVirtual().Return(true)
			},
			fields:  fields{bmHostNamespace: ns},
			wantErr: false,
			want:    true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			mockBmc := mocks.NewMockBMCInterface(mockCtrl)
			mockClient := mocks.NewMockGoFishClientAccessor(mockCtrl)
			mockService := mocks.NewMockGoFishServiceAccessor(mockCtrl)
			mockUpdateService := mocks.NewMockGoFishUpdateServiceAccessor(mockCtrl)
			mockDCMI := mocks.NewMockDCIM(mockCtrl)
			if tt.mockfunc != nil {
				tt.mockfunc(mockBmc, mockClient, mockService, mockUpdateService, mockDCMI)
			}
			t := &EnrollmentTask{
				bmHostNamespace: tt.fields.bmHostNamespace,
				bmc:             mockBmc,
				clientSet:       clientSet,
				deviceData:      &DeviceData{Name: "mocked-dev-1"},
				netBox:          mockDCMI,
			}
			if _, err := t.checkMinFWVersions(context.TODO()); (err != nil) != tt.wantErr {
				t1.Errorf("checkMinFWVersions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
