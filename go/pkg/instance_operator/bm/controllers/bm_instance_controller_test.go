// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package privatecloud

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gomegatypes "github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	bmenrollment "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	instancetest "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/test"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mocks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mygofish"
	"github.com/stmcginnis/gofish/redfish"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BM Instance Controller Suite")
}

var Any = gomock.Any()

var _ = Describe("BM Instance Controller", func() {

	const (
		testNamespace1      = "test-ns-1"
		testNamespace2      = "test-ns-2"
		availableHostName   = "available-host"
		unavailableHostName = "unavailable-host"
	)

	var (
		ctx                context.Context
		backend            *BmInstanceBackend
		k8sClient          client.WithWatch
		instance           *cloudv1alpha1.Instance
		gofishManager      *mocks.MockGoFishManagerAccessor
		mockCtrl           *gomock.Controller
		comSys             *mocks.MockGoFishComputerSystemAccessor
		machineImageServer *httptest.Server
	)

	BeforeEach(func() {
		ctx = context.Background()
		machineImageServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("\n")) // empty checksum file since the checksum is not populated in the tests.
			w.WriteHeader(http.StatusOK)
		}))
		scheme := runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(scheme)).NotTo(HaveOccurred())
		Expect(cloudv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
		Expect(baremetalv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

		k8sClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()

		bmOperatorConfig := instancetest.NewTestBmInstanceOperatorConfig("../testdata", scheme)
		bmOperatorConfig.OsHttpServerUrl = machineImageServer.URL

		bmcSecrets := BMCSecrets{
			Username: "user",
			Password: "password",
			URL:      "https://url.com",
		}
		mockCtrl = gomock.NewController(GinkgoT())
		gofishManager = mocks.NewMockGoFishManagerAccessor(mockCtrl)
		client := mocks.NewMockGoFishClientAccessor(mockCtrl)
		service := mocks.NewMockGoFishServiceAccessor(mockCtrl)
		client.EXPECT().GetService().AnyTimes().Return(service)
		gofishManager.EXPECT().Connect(Any).AnyTimes().Return(client, nil)

		systems := make([]mygofish.GoFishComputerSystemAccessor, 1)
		comSys = mocks.NewMockGoFishComputerSystemAccessor(mockCtrl)
		systems[0] = comSys
		service.EXPECT().Systems().AnyTimes().Return(systems, nil)
		comSys.EXPECT().ODataID().AnyTimes().Return("SushyODataID")
		comSys.EXPECT().Manufacturer().AnyTimes().Return(bmc.SushyEmulator)
		comSys.EXPECT().Model().AnyTimes().Return("sushy")
		comSys.EXPECT().PowerState().AnyTimes().Return(redfish.OffPowerState)

		bmc, err := bmc.New(
			gofishManager,
			&bmc.Config{
				URL:      bmcSecrets.URL,
				Username: bmcSecrets.Username,
				Password: bmcSecrets.Password,
			})
		Expect(bmc).ToNot(BeNil())
		Expect(err).NotTo(HaveOccurred())
		By("creating BM instance controller")

		backend = &BmInstanceBackend{
			Client:       k8sClient,
			Cfg:          bmOperatorConfig,
			bmcSecrets:   bmcSecrets,
			bmcInterface: bmc,
		}

		By("creating namespaces for BareMetalHosts")
		Expect(k8sClient.Create(ctx, newBmNamespace(testNamespace1))).Should(Succeed())
		Expect(k8sClient.Create(ctx, newBmNamespace(testNamespace2))).Should(Succeed())

		By("creating available BareMetalHosts")
		Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName, testNamespace1))).Should(Succeed())
		Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName, testNamespace2))).Should(Succeed())

		By("creating unavailable BareMetalHosts")
		Expect(k8sClient.Create(ctx, newUnavailableBmHost(unavailableHostName, testNamespace1))).Should(Succeed())
		Expect(k8sClient.Create(ctx, newUnavailableBmHost(unavailableHostName, testNamespace2))).Should(Succeed())

		By("creating a BM Instance")
		instance = newBmInstance("bm-instance", "bm-instance")
		Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
	})

	AfterEach(func() {
		defer machineImageServer.Close()
	})

	Describe("creating or updating a BM Instance", func() {
		var chosenHost *baremetalv1alpha1.BareMetalHost

		BeforeEach(func() {
			Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())
			chosenHost = getConsumedHost(ctx, k8sClient)
		})

		When("host is chosen", func() {
			It("should add a consumer reference to the host", func() {
				Expect(chosenHost.Spec.ConsumerRef).ToNot(BeNil())
				Expect(chosenHost.Spec.ConsumerRef.Name).To(Equal(instance.Name))
				Expect(chosenHost.Spec.ConsumerRef.Namespace).To(Equal(instance.Namespace))
			})

			It("should create a Secret that contains user data", func() {
				userdataSecret := &corev1.Secret{}
				objKey := types.NamespacedName{Namespace: chosenHost.Namespace, Name: "usrbm-instance-secret"}
				Expect(k8sClient.Get(ctx, objKey, userdataSecret)).To(Succeed())
				Expect(userdataSecret.Type).To(Equal(metal3SecretType))
				Expect(userdataSecret.Data).NotTo(BeEmpty())
				Expect(userdataSecret.Data["userData"]).NotTo(BeEmpty())
			})

			It("should create a Secret that contains network data", func() {
				userdataSecret := &corev1.Secret{}
				objKey := types.NamespacedName{Namespace: chosenHost.Namespace, Name: "network-bm-instance-secret"}
				Expect(k8sClient.Get(ctx, objKey, userdataSecret)).To(Succeed())
				Expect(userdataSecret.Type).To(Equal(metal3SecretType))
				Expect(userdataSecret.Data).NotTo(BeEmpty())
				Expect(userdataSecret.Data["networkData"]).NotTo(BeEmpty())
			})

			It("should annotate the Instance with the chosen host key", func() {
				hostKey := fmt.Sprintf("%s/%s", chosenHost.Namespace, chosenHost.Name)
				Expect(instance.GetAnnotations()).To(HaveKeyWithValue(HostAnnotation, hostKey))
			})

			It("should update the Instance with 'Accepted' state", func() {
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionAccepted, corev1.ConditionTrue))
			})
		})

		When("host is provisioning", func() {
			It("should update the Instance with 'Running' state", func() {
				chosenHost.Status.Provisioning.State = baremetalv1alpha1.StateProvisioning
				Expect(k8sClient.Update(ctx, chosenHost)).To(Succeed())
				Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())
				chosenHost = getConsumedHost(ctx, k8sClient)
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionRunning, corev1.ConditionTrue))
			})
		})

		When("host is provisioned", func() {
			BeforeEach(func() {
				chosenHost.Status.Provisioning.State = baremetalv1alpha1.StateProvisioned
				Expect(k8sClient.Update(ctx, chosenHost)).To(Succeed())
				Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())
				chosenHost = getConsumedHost(ctx, k8sClient)
			})

			It("should get change HCI to disabled", func() {
				Expect(chosenHost.Status.HardwareDetails.NIC).NotTo(BeEmpty())
			})

			It("should update the Instance Condition with HCI Enabled state to False", func() {
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionHCIEnabled, corev1.ConditionFalse))
			})
		})

		When("host is provisioned", func() {
			BeforeEach(func() {
				chosenHost.Status.Provisioning.State = baremetalv1alpha1.StateProvisioned
				By("HCI should be disabled")
				condition := cloudv1alpha1.InstanceCondition{
					Type:               cloudv1alpha1.InstanceConditionHCIEnabled,
					Status:             corev1.ConditionFalse,
					LastProbeTime:      metav1.Now(),
					LastTransitionTime: metav1.Now(),
				}
				util.SetStatusCondition(&instance.Status.Conditions, condition)
				chosenHost.Status.Provisioning.State = baremetalv1alpha1.StateProvisioned
				Expect(k8sClient.Update(ctx, chosenHost)).To(Succeed())
				Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())
				chosenHost = getConsumedHost(ctx, k8sClient)
			})

			It("should get change KCS to disabled", func() {
				Expect(chosenHost.Status.HardwareDetails.NIC).NotTo(BeEmpty())
			})

			It("should update the Instance Condition with KCS Enabled state to False", func() {
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionKcsEnabled, corev1.ConditionFalse))
			})
		})

		When("host is provisioned", func() {
			BeforeEach(func() {
				chosenHost.Status.Provisioning.State = baremetalv1alpha1.StateProvisioned
				By("HCI should be disabled")
				condition := cloudv1alpha1.InstanceCondition{
					Type:               cloudv1alpha1.InstanceConditionHCIEnabled,
					Status:             corev1.ConditionFalse,
					LastProbeTime:      metav1.Now(),
					LastTransitionTime: metav1.Now(),
				}
				util.SetStatusCondition(&instance.Status.Conditions, condition)
				By("KCS should be disabled")
				condition = cloudv1alpha1.InstanceCondition{
					Type:               cloudv1alpha1.InstanceConditionKcsEnabled,
					Status:             corev1.ConditionFalse,
					LastProbeTime:      metav1.Now(),
					LastTransitionTime: metav1.Now(),
				}
				util.SetStatusCondition(&instance.Status.Conditions, condition)
				Expect(k8sClient.Update(ctx, chosenHost)).To(Succeed())
				Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())
				chosenHost = getConsumedHost(ctx, k8sClient)
			})

			It("should get network switch data from lldp", func() {
				Expect(chosenHost.Status.HardwareDetails.NIC).NotTo(BeEmpty())
			})

			It("should update the Instance state with 'StartupComplete' state", func() {
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionAgentConnected, corev1.ConditionTrue))
			})
		})
	})

	Describe("Powercycling node based on RunStrategy", func() {
		var chosenHost *baremetalv1alpha1.BareMetalHost

		BeforeEach(func() {
			By("Consuming an available host")
			Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())

			By("host should be in 'Provisioned' state")
			chosenHost = getConsumedHost(ctx, k8sClient)
			chosenHost.Status.Provisioning.State = baremetalv1alpha1.StateProvisioned
			Expect(k8sClient.Update(ctx, chosenHost)).To(Succeed())
		})

		When("host is provisioned", func() {
			BeforeEach(func() {
				By("Updating host to be Online during 'Started' Condition")
				Expect(backend.updateHostOnlineStatus(ctx, chosenHost, true)).Error().NotTo(HaveOccurred())
				chosenHost = getConsumedHost(ctx, k8sClient)

				By("Changing instance RunStrategy to 'Halted' in 'Started' Condition")
				instance.Spec.RunStrategy = cloudv1alpha1.RunStrategyHalted
				backend.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStarted)
			})

			It("Ensure that Instance is in 'Started' Condition to apply 'Halted' RunStrategy", func() {
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStarted, corev1.ConditionTrue))
			})

			It("Instance Condition should update to 'Stopping' Condition while powering off", func() {
				By("Ensure that host should be Online")
				Expect(getHostOnlineStatus(ctx, k8sClient, chosenHost)).To(BeTrue())
				By("Ensure that instance should not be in 'Stopping' Condition")
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStopping, corev1.ConditionFalse))

				By("associate")
				Expect(util.IsInstanceStarted(instance)).To(BeTrue())
				Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())

				By("host should not be Online")
				Expect(getHostOnlineStatus(ctx, k8sClient, chosenHost)).To(BeFalse())
				By("instance should be in 'Stopping' Condition")
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStopping, corev1.ConditionTrue))
			})
		})

		When("host is provisioned", func() {
			BeforeEach(func() {
				chosenHost = getConsumedHost(ctx, k8sClient)

				By("Changing instance RunStrategy to 'Halted' in 'Stopping' Condition")
				instance.Spec.RunStrategy = cloudv1alpha1.RunStrategyHalted
				backend.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStopping)
			})

			It("Ensure that Instance is in 'Stopping' Condition", func() {
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStopping, corev1.ConditionTrue))
			})

			It("Instance Condition should update to 'Stopped' Condition", func() {
				By("Ensure that instance should not be in 'Stopped' Condition")
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStopped, corev1.ConditionFalse))

				By("associate")
				Expect(util.IsInstanceStopping(instance)).To(BeTrue())
				Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())

				By("instance should be in 'Stopped' Condition")
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStopped, corev1.ConditionTrue))
			})
		})

		When("host is provisioned", func() {
			BeforeEach(func() {
				By("Updating host not to be Online during 'Stopped' Condition")
				Expect(backend.updateHostOnlineStatus(ctx, chosenHost, false)).Error().NotTo(HaveOccurred())
				chosenHost = getConsumedHost(ctx, k8sClient)

				By("Changing Instance RunStrategy to 'Always' in 'Stopped' Condition")
				instance.Spec.RunStrategy = cloudv1alpha1.RunStrategyAlways
				backend.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStopped)
			})

			It("Ensure that Instance is in 'Stopped' Condition to apply 'Always' RunStrategy", func() {
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStopped, corev1.ConditionTrue))
			})

			It("Instance Condition should update to 'Starting' while powering on", func() {
				By("Ensure that host should not be Online")
				Expect(getHostOnlineStatus(ctx, k8sClient, chosenHost)).To(BeFalse())
				By("Ensure that instance should not be in 'Starting' Condition")
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStarting, corev1.ConditionFalse))

				By("associate")
				Expect(util.IsInstanceStoppedCompleted(instance)).To(BeTrue())
				Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())

				By("host should be Online")
				Expect(getHostOnlineStatus(ctx, k8sClient, chosenHost)).To(BeTrue())
				By("instance should be in 'Starting' Condition")
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStarting, corev1.ConditionTrue))
			})
		})

		When("host is provisioned", func() {
			BeforeEach(func() {
				By("Updating host to NOT be Online during 'Stopped' Condition")
				Expect(backend.updateHostOnlineStatus(ctx, chosenHost, false)).Error().NotTo(HaveOccurred())
				chosenHost = getConsumedHost(ctx, k8sClient)

				By("Changing Instance RunStrategy to 'RerunOnFailure' in 'Stopped' Condition")
				instance.Spec.RunStrategy = cloudv1alpha1.RunStrategyRerunOnFailure
				backend.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStopped)
			})

			It("Instance Condition should update to 'Stopped' to apply 'RerunOnFailure' RunStrategy", func() {
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStopped, corev1.ConditionTrue))
			})

			It("Instance Condition should update to 'Starting' while powering on", func() {
				By("Ensure that host should not be Online")
				Expect(getHostOnlineStatus(ctx, k8sClient, chosenHost)).To(BeFalse())
				By("Ensure that instance should not be in 'Starting' Condition")
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStarting, corev1.ConditionFalse))

				By("associate")
				Expect(util.IsInstanceStoppedCompleted(instance)).To(BeTrue())
				Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())

				By("host should be Online")
				Expect(getHostOnlineStatus(ctx, k8sClient, chosenHost)).To(BeTrue())
				By("instance should be in 'Starting' Condition")
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionStarting, corev1.ConditionTrue))
			})
		})

		When("host is provisioned", func() {
			BeforeEach(func() {
				By("HCI should be disabled")
				condition := cloudv1alpha1.InstanceCondition{
					Type:               cloudv1alpha1.InstanceConditionHCIEnabled,
					Status:             corev1.ConditionFalse,
					LastProbeTime:      metav1.Now(),
					LastTransitionTime: metav1.Now(),
				}
				util.SetStatusCondition(&instance.Status.Conditions, condition)
				By("KCS should be disabled")
				condition = cloudv1alpha1.InstanceCondition{
					Type:               cloudv1alpha1.InstanceConditionKcsEnabled,
					Status:             corev1.ConditionFalse,
					LastProbeTime:      metav1.Now(),
					LastTransitionTime: metav1.Now(),
				}
				util.SetStatusCondition(&instance.Status.Conditions, condition)
				Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())

				By("Changing Instance Condition to 'Starting'")
				backend.updateInstanceConditions(ctx, instance, cloudv1alpha1.InstanceConditionStarting)
			})

			It("instance should be in 'Started' Condition for 'Always' and 'RerunOnFailure' RunStrategy", func() {
				Expect(util.IsInstanceAgentConnected(instance)).To(BeTrue())
				Expect(instance.Status.Conditions).To(HaveCondition(cloudv1alpha1.InstanceConditionAgentConnected, corev1.ConditionTrue))
			})
		})
	})

	Describe("deleting a BM instance", func() {
		var consumedHost, availableHost *baremetalv1alpha1.BareMetalHost
		BeforeEach(func() {
			By("consuming an available host")
			Expect(backend.CreateOrUpdateInstance(ctx, instance)).Error().NotTo(HaveOccurred())
			By("making host available again")
			consumedHost = getConsumedHost(ctx, k8sClient)
			condition := cloudv1alpha1.InstanceCondition{
				Type:               cloudv1alpha1.InstanceConditionKcsEnabled,
				Status:             corev1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
			}
			util.SetStatusCondition(&instance.Status.Conditions, condition)
			Expect(backend.DeleteResources(ctx, instance)).Error().NotTo(HaveOccurred())
			By("getting the host that became available")
			availableHost = &baremetalv1alpha1.BareMetalHost{}
			objKey := types.NamespacedName{Namespace: consumedHost.Namespace, Name: consumedHost.Name}
			Expect(k8sClient.Get(ctx, objKey, availableHost)).To(Succeed())
		})

		It("should remove consumer data from host spec", func() {
			Expect(availableHost.Spec.Online).To(BeTrue())
			Expect(availableHost.Spec.Image).To(BeNil())
			Expect(availableHost.Spec.UserData).To(BeNil())
			Expect(availableHost.Spec.MetaData).To(BeNil())
			Expect(availableHost.Spec.NetworkData).To(BeNil())
			Expect(availableHost.Spec.ConsumerRef).To(BeNil())
		})

		It("should remove the associated secrets from K8s", func() {
			objKey := types.NamespacedName{Namespace: consumedHost.Spec.UserData.Namespace, Name: consumedHost.Spec.UserData.Name}
			Expect(k8sClient.Get(ctx, objKey, &corev1.Secret{})).To(MatchK8sError(errors.IsNotFound))
		})
	})

	Describe("getting a specific BareMetalHost based on Instance's annotation", func() {
		When("the matching host is found", func() {
			It("should return the host with no error", func() {
				By("setting valid host key in Instance's annotations")
				hostKey := fmt.Sprintf("%s/%s", testNamespace1, availableHostName)
				instance.SetAnnotations(map[string]string{HostAnnotation: hostKey})
				Expect(k8sClient.Update(ctx, instance)).To(Succeed())

				host, patch, err := backend.getHost(ctx, instance)
				Expect(err).To(BeNil())
				Expect(patch).NotTo(BeNil())
				Expect(host).NotTo(BeNil())
				Expect(host.GetNamespace()).To(Equal(testNamespace1))
				Expect(host.GetName()).To(Equal(availableHostName))
			})
		})

		When("the matching host is not found", func() {
			It("cannot get non-nil annotations from Instance", func() {
				By("setting Instance's annotation to nil")
				instance.SetAnnotations(nil)

				host, patch, err := backend.getHost(ctx, instance)
				Expect(err).To(BeNil())
				Expect(patch).To(BeNil())
				Expect(host).To(BeNil())
			})

			It("cannot find the host key from Instance's annotations", func() {
				By("setting annotation to contain no host key")
				instance.SetAnnotations(map[string]string{})

				host, patch, err := backend.getHost(ctx, instance)
				Expect(err).To(BeNil())
				Expect(patch).To(BeNil())
				Expect(host).To(BeNil())
			})

			It("cannot get BareMetalHost from K8s based on the given host annotation", func() {
				By("setting non-matching host key")
				instance.SetAnnotations(map[string]string{HostAnnotation: "test-namespace-1/no-matching-host"})

				host, patch, err := backend.getHost(ctx, instance)
				Expect(err).To(BeNil())
				Expect(patch).To(BeNil())
				Expect(host).To(BeNil())
			})
		})
	})

	Describe("choosing an available BareMetalHost to be consumed by the Instance", func() {
		When("host is chosen", func() {
			It("should choose a host in consumable condition", func() {
				chosenHost, helper, err := backend.findHost(ctx, instance)
				Expect(err).To(BeNil())
				Expect(helper).ToNot(BeNil())
				Expect(chosenHost).ToNot(BeNil())

				Expect(chosenHost.DeletionTimestamp).To(BeNil())
				Expect(chosenHost.Annotations).ToNot(HaveKey(baremetalv1alpha1.PausedAnnotation))
				Expect(chosenHost.Status.Provisioning.State).To(Equal(baremetalv1alpha1.StateAvailable))
				Expect(chosenHost.Status.OperationalStatus).To(Equal(baremetalv1alpha1.OperationalStatusOK))
				Expect(chosenHost.Status.ErrorCount).To(Equal(0))
				Expect(chosenHost.Status.ErrorMessage).To(BeEmpty())
			})
		})
	})

	Describe("checking if host's consumer reference matches an Instance", func() {
		var host *baremetalv1alpha1.BareMetalHost

		BeforeEach(func() {
			host = newAvailableBmHost(availableHostName, testNamespace1)
			Expect(backend.setHostConsumerRef(ctx, host, instance)).To(Succeed())
		})

		It("should return false when API Version does not match", func() {
			host.Spec.ConsumerRef.APIVersion = "unknown"
			Expect(consumerRefMatches(host.Spec.ConsumerRef, instance)).To(BeFalse())
		})

		It("should return false when Kind does not match", func() {
			host.Spec.ConsumerRef.Kind = "unknown"
			Expect(consumerRefMatches(host.Spec.ConsumerRef, instance)).To(BeFalse())
		})

		It("should return false when Name does not match", func() {
			host.Spec.ConsumerRef.Name = "unknown"
			Expect(consumerRefMatches(host.Spec.ConsumerRef, instance)).To(BeFalse())
		})

		It("should return false when Namespace does not match", func() {
			host.Spec.ConsumerRef.Namespace = "unknown"
			Expect(consumerRefMatches(host.Spec.ConsumerRef, instance)).To(BeFalse())
		})

		It("should return true when the consumer reference matches the instance", func() {
			Expect(consumerRefMatches(host.Spec.ConsumerRef, instance)).To(BeTrue())
		})
	})

	Describe("adding data from Instance to host spec", func() {
		var host *baremetalv1alpha1.BareMetalHost

		BeforeEach(func() {
			host = newAvailableBmHost(availableHostName, testNamespace1)
			Expect(instance.Spec.Interfaces).NotTo(BeEmpty())
			Expect(instance.Status.Interfaces).NotTo(BeEmpty())
		})

		It("should gather the user data", func() {
			var disks map[string][]baremetalv1alpha1.Storage
			userData, err := backend.setUserData(ctx, host, instance, disks)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance.Spec.SshPublicKeySpecs).To(HaveLen(2))
			Expect(userData).To(SatisfyAll(
				ContainSubstring("#cloud-config"),
				ContainSubstring("hostname:"),
				ContainSubstring(instance.Spec.Interfaces[0].DnsName),
				ContainSubstring(instance.Spec.SshPublicKeySpecs[0].SshPublicKey),
				ContainSubstring(instance.Spec.SshPublicKeySpecs[1].SshPublicKey)))
		})

		It("should gather the network data", func() {
			Expect(instance.Spec.Interfaces).To(HaveLen(1))
			Expect(instance.Spec.Interfaces[0].Nameservers).NotTo(BeEmpty())
			Expect(instance.Status.Interfaces[0].Addresses).NotTo(BeEmpty())

			networkData, err := backend.setNetworkData(ctx, host, instance)
			Expect(err).NotTo(HaveOccurred())
			Expect(networkData).To(SatisfyAll(
				ContainSubstring("links:"),
				ContainSubstring("networks:"),
				ContainSubstring(host.Spec.BootMACAddress),
				ContainSubstring(instance.Spec.Interfaces[0].Nameservers[0])))
		})

		It("should add the gathered data to host spec", func() {
			Expect(backend.setHostSpec(
				host,
				"userdata-secret-name",
				"networkdata-secret-name",
				instance,
			)).Should(Succeed())

			Expect(host.Spec.Online).To(BeTrue())
			Expect(host.Spec.Image).NotTo(BeNil())
			Expect(host.Spec.Image.URL).NotTo(BeEmpty())
			Expect(host.Spec.Image.Checksum).NotTo(BeEmpty())
			Expect(host.Spec.Image.DiskFormat).NotTo(BeNil())

			Expect(host.Spec.UserData).NotTo(BeNil())
			Expect(host.Spec.UserData.Name).NotTo(BeEmpty())
			Expect(host.Spec.UserData.Namespace).NotTo(BeEmpty())

			Expect(host.Spec.NetworkData).NotTo(BeNil())
			Expect(host.Spec.NetworkData.Name).NotTo(BeEmpty())
			Expect(host.Spec.NetworkData.Namespace).NotTo(BeEmpty())
		})
	})

	DescribeTable("calculating CPU count",
		func(sockets, cores, threads, expectedCount int) {
			instance.Spec.InstanceTypeSpec.Cpu.Sockets = uint32(sockets)
			instance.Spec.InstanceTypeSpec.Cpu.Cores = uint32(cores)
			instance.Spec.InstanceTypeSpec.Cpu.Threads = uint32(threads)
			Expect(calculateCPUCount(instance)).To(Equal(expectedCount))
		},
		Entry("single cpu", 1, 1, 1, 1),
		Entry("multi-socket", 2, 1, 1, 2),
		Entry("multi-core", 1, 2, 1, 2),
		Entry("multi-thread", 1, 1, 2, 2),
	)

	AfterEach(func() {
		Expect(k8sClient.DeleteAllOf(ctx, &cloudv1alpha1.Instance{})).Should(Succeed())
		Expect(k8sClient.DeleteAllOf(ctx, &baremetalv1alpha1.BareMetalHost{})).Should(Succeed())
		Expect(k8sClient.DeleteAllOf(ctx, &corev1.Namespace{})).Should(Succeed())
	})
})

func newBmInstance(name, namespace string) *cloudv1alpha1.Instance {
	networkInterface := cloudv1alpha1.InterfaceSpec{
		Name:        "eth0",
		VNet:        "us-dev-1a-default",
		DnsName:     "my-baremetal-machine-test-1.03165859732720551183.us-dev-1.cloud.intel.com",
		Nameservers: []string{"1.1.1.1"},
	}

	return &cloudv1alpha1.Instance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "private.cloud.intel.com/v1alpha1",
			Kind:       "Instance",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"cluster-id": "test-ns-2",
				"node-id":    "available-host",
			},
		},
		Spec: cloudv1alpha1.InstanceSpec{
			AvailabilityZone: "test-az",
			Region:           "test-region",
			RunStrategy:      "RerunOnFailure",
			Interfaces: []cloudv1alpha1.InterfaceSpec{
				networkInterface,
			},
			InstanceTypeSpec: cloudv1alpha1.InstanceTypeSpec{
				Name:             "test-baremetal",
				DisplayName:      "baremetal node for testing",
				Description:      "Intel Test Server",
				InstanceCategory: cloudv1alpha1.InstanceCategoryBareMetalHost,
				Disks: []cloudv1alpha1.DiskSpec{
					{Size: "128Gi"},
				},
				Cpu: cloudv1alpha1.CpuSpec{
					Cores:     1,
					Sockets:   2,
					Threads:   1,
					ModelName: "Intel Test Server",
					Id:        "0x00001",
				},
				Gpu: cloudv1alpha1.GpuSpec{
					Count: 0,
				},
				Memory: cloudv1alpha1.MemorySpec{
					Size:      "8Gi",
					DimmSize:  "8Gi",
					DimmCount: 1,
					Speed:     4800,
				},
			},
			MachineImageSpec: cloudv1alpha1.MachineImageSpec{
				Name: "ubuntu-22.04",
			},
			SshPublicKeyNames: []string{"SshPublicKeyName1", "SshPublicKeyName2"},
			SshPublicKeySpecs: []cloudv1alpha1.SshPublicKeySpec{
				{
					SshPublicKey: "ssh-ed25519 AAA testuser1.example.com",
				},
				{
					SshPublicKey: "ssh-ed25519 BBB testuser2.example.com",
				},
			},
			ClusterGroupId: "test-clustergroup-1a-2",
			ClusterId:      "test-clustergroup-1a-2-4",
		},
		Status: cloudv1alpha1.InstanceStatus{
			Interfaces: []cloudv1alpha1.InstanceInterfaceStatus{
				{
					Name:         networkInterface.Name,
					VNet:         networkInterface.VNet,
					DnsName:      networkInterface.DnsName,
					Addresses:    []string{"1.2.3.4"},
					PrefixLength: 24,
					Subnet:       "Subnet1",
					Gateway:      "Gateway1",
					VlanId:       1001,
				},
			},
		},
	}
}

func newBmNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind: "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				bmenrollment.Metal3NamespaceSelectorKey: "true",
			},
		},
	}
}

func newBmHost(name, namespace string) *baremetalv1alpha1.BareMetalHost {
	bmh := &baremetalv1alpha1.BareMetalHost{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "metal3.io/v1alpha1",
			Kind:       "BareMetalHost",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				bmenrollment.CPUIDLabel:        "0x00001",
				bmenrollment.CPUCountLabel:     "2",
				bmenrollment.GPUModelNameLabel: "",
				bmenrollment.GPUCountLabel:     "0",
				bmenrollment.HBMModeLabel:      "",
			},
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
		Status: baremetalv1alpha1.BareMetalHostStatus{
			Provisioning: baremetalv1alpha1.ProvisionStatus{
				State: baremetalv1alpha1.StateAvailable,
			},
		},
	}
	return bmh
}

func newAvailableBmHost(name, namespace string) *baremetalv1alpha1.BareMetalHost {
	bmh := newBmHost(name, namespace)
	bmh.Status = baremetalv1alpha1.BareMetalHostStatus{
		Provisioning: baremetalv1alpha1.ProvisionStatus{
			State: baremetalv1alpha1.StateAvailable,
		},
		OperationalStatus: baremetalv1alpha1.OperationalStatusOK,
		HardwareDetails: &baremetalv1alpha1.HardwareDetails{
			NIC: []baremetalv1alpha1.NIC{
				{
					PXE:    true,
					MAC:    "a1:b2:c3:d4:f5",
					Model:  "test-nic",
					Name:   "eno0",
					VLANID: 0,
					LLDP: baremetalv1alpha1.LLDP{
						SwitchPortId:     "0",
						SwitchSystemName: "test-switch.app.intel.com",
					},
				},
			},
		},
	}
	return bmh
}

func newUnavailableBmHost(name, namespace string) *baremetalv1alpha1.BareMetalHost {
	bmh := newBmHost(name, namespace)
	bmh.Status = baremetalv1alpha1.BareMetalHostStatus{
		Provisioning: baremetalv1alpha1.ProvisionStatus{
			State: baremetalv1alpha1.StateNone,
		},
	}
	return bmh
}

func getConsumedHost(ctx context.Context, client client.Client) *baremetalv1alpha1.BareMetalHost {
	hostList := &baremetalv1alpha1.BareMetalHostList{}
	if err := client.List(ctx, hostList); err != nil {
		return nil
	}
	for i, host := range hostList.Items {
		if host.Spec.ConsumerRef != nil {
			return &hostList.Items[i]
		}
	}
	return nil
}

func getHostOnlineStatus(ctx context.Context, client client.Client, host *baremetalv1alpha1.BareMetalHost) (bool, error) {
	updatedHost := &baremetalv1alpha1.BareMetalHost{}
	err := client.Get(ctx, types.NamespacedName{Name: host.Name, Namespace: host.Namespace}, updatedHost)
	if err != nil {
		return false, fmt.Errorf("failed to get host Online status: %w", err)
	}

	return updatedHost.Spec.Online, nil
}

func HaveCondition(conditionType cloudv1alpha1.InstanceConditionType, conditionStatus corev1.ConditionStatus) gomegatypes.GomegaMatcher {
	return WithTransform(func(conditions []cloudv1alpha1.InstanceCondition) bool {
		for _, condition := range conditions {
			if condition.Type == conditionType &&
				condition.Status == conditionStatus {
				return true
			}
		}
		return false
	}, BeTrue())
}

func MatchK8sError(predicate func(err error) bool) gomegatypes.GomegaMatcher {
	return WithTransform(func(errorToCheck error) bool {
		return predicate(errorToCheck)
	}, BeTrue())
}
