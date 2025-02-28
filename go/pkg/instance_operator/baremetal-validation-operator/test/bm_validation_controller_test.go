// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	bmenrollment "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	validation "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/baremetal-validation-operator/controllers/metal3.io"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	baremetalv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
)

const (
	interval = time.Millisecond * 500
)

var _ = Describe("BM Validation Controller with invalid InstanceTypes", func() {
	const testNamespace = "test-validation-1"
	const availableHostName = "available-host1"
	// Object lookup keys
	bmhLookupKey1 := types.NamespacedName{Namespace: testNamespace, Name: availableHostName}
	// Object Reference
	bmhRef := &baremetalv1alpha1.BareMetalHost{}

	var _ = Describe("creating or updating a BM Instance", func() {
		ctx := context.Background()

		It("Init", func() {
			By("creating namespaces for BMH")
			Expect(k8sClient.Create(ctx, newBmNamespace(testNamespace))).Should(Succeed())
		})

		It("create BMHost with no InstanceType", func() {
			By("creating a BMH")
			Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName, testNamespace, bmenrollment.ReadyToTestLabel))).
				Should(Succeed())
			// Ensure the operator ignores it
			Expect(k8sClient.Get(ctx, bmhLookupKey1, bmhRef)).Should(Succeed())
			Expect(bmhRef.Labels).To(HaveKeyWithValue(bmenrollment.ReadyToTestLabel, "true"))
			Expect(bmhRef.Labels).NotTo(HaveKey(bmenrollment.CheckingFailedLabel))
		})

		It("create BMHost with disabled InstanceType", func() {
			By("creating a BMH")
			Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName, testNamespace,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-spr")))).
				Should(Succeed())

			//Validation operator should accept this event and ensure the verified label is set to true
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey1, bmhRef)).Should(Succeed())
				g.Expect(bmhRef.Labels).Should(HaveKeyWithValue(bmenrollment.VerifiedLabel, "true"))
				g.Expect(bmhRef.Labels).ShouldNot(HaveKey(bmenrollment.ReadyToTestLabel))
			}, timeout, interval).Should(Succeed())
		})

		It("create BMHost with nil labels and enabled instanceType", func() {
			By("creating a BMH") // bm-icp-gaudi2 is enabled in operator config
			bmh := newAvailableBmHost(availableHostName, testNamespace,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2"))
			bmh.Labels = nil // Intentionally set this to a nil value.
			Expect(k8sClient.Create(ctx, bmh)).Should(Succeed())

			//ensure the operator ignores it.
			Expect(k8sClient.Get(ctx, bmhLookupKey1, bmhRef)).Should(Succeed())
			Expect(len(bmhRef.Labels)).To(Equal(0))
		})

		AfterEach(func() {
			opts := []client.DeleteAllOfOption{
				client.InNamespace(testNamespace),
				client.GracePeriodSeconds(5),
			}
			Expect(k8sClient.DeleteAllOf(ctx, &baremetalv1alpha1.BareMetalHost{}, opts...)).Should(Succeed())
		})
	})
})

var _ = Describe("BM Validation Controller phase test", Ordered, func() {
	const testNamespace = "test-validation-2"
	const availableHostPrefix = "available-host"
	// Object Reference
	bmhRef := &baremetalv1alpha1.BareMetalHost{}

	var _ = Describe("creating a BM Instance", func() {
		ctx := context.Background()

		It("Init", func() {
			By("creating namespaces for BMH")
			Expect(k8sClient.Create(ctx, newBmNamespace(testNamespace))).Should(Succeed())
		})

		It("Create a BMH and verify Validator.STATE_BEGIN state", func() {
			By("creating a BMH")
			availableHostName, bmhLookupKey := createBMHName(testNamespace, availableHostPrefix)
			// bm-icp-gaudi2 is enabled in operator config
			Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName, testNamespace,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2")))).
				Should(Succeed())

			//Validation operator should accept this event and start the provisioning phase
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey, bmhRef)).Should(Succeed())
				g.Expect(bmhRef.Labels).Should(HaveKeyWithValue(bmenrollment.ImagingLabel, "true"))
				g.Expect(bmhRef.Labels).Should(HaveKey(bmenrollment.ReadyToTestLabel))
			}, timeout, interval).Should(Succeed())
		})

		It("Create a BMH and verify deprovisioning stuck state", func() {
			By("creating a BMH, stuck in deprovisioning state")
			// bm-icp-gaudi2 is enabled in operator config
			availableHostName, bmhLookupKey := createBMHName(testNamespace, availableHostPrefix)
			bmh := newBMH(availableHostName, testNamespace, baremetalv1alpha1.StateDeprovisioning,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2"))
			bmh.Status.OperationHistory.Deprovision.Start.Time = time.Now().UTC().Add(-(validation.TimeoutBMHStateMinutes + 1) * time.Minute)
			Expect(k8sClient.Create(ctx, bmh)).Should(Succeed())

			//Validation operator should detect that the BMH is stuck in deprovisioning state for time greater than threshold.
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey, bmhRef)).Should(Succeed())
				g.Expect(bmhRef.Labels).Should(HaveKeyWithValue(bmenrollment.CheckingFailedLabel, "Timeout.waiting.for.deprovision.to.complete"))
				g.Expect(bmhRef.Labels).ShouldNot(HaveKey(bmenrollment.VerifiedLabel))
				g.Expect(bmhRef.Labels).ShouldNot(HaveKey(bmenrollment.ValidationIdLabel))
			}, timeout, interval).Should(Succeed())
		})

		It("Create a BMH and verify Validator.Provisioning state with ready instance state", func() {
			By("creating a BMH")
			// bm-icp-gaudi2 is enabled in operator config
			availableHostName, bmhLookupKey := createBMHName(testNamespace, availableHostPrefix)
			Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName, testNamespace,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2")))).
				Should(Succeed())

			//Create an Instance in k8s to simulate the instance marked Ready by the bm-instance operator.
			Expect(k8sClient.Create(ctx, newInstance(availableHostName+"-validation", testNamespace, cloudv1alpha1.PhaseReady))).
				Should(Succeed())

			// update BMH state to provisoned
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey, bmhRef)).Should(Succeed())
				bmhRef.Status.Provisioning.State = baremetalv1alpha1.StateProvisioned
				g.Expect(k8sClient.Update(ctx, bmhRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			//Validation operator should accept this event and start the provisioning phase
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey, bmhRef)).Should(Succeed())
				g.Expect(bmhRef.Labels).ShouldNot(HaveKeyWithValue(bmenrollment.ImagingLabel, "true"))
				g.Expect(bmhRef.Labels).Should(HaveKeyWithValue(bmenrollment.ImagingCompletedLabel, "true"))
				g.Expect(bmhRef.Labels).Should(HaveKey(bmenrollment.ReadyToTestLabel))
			}, timeout, interval).Should(Succeed())
		})

		It("Create a BMH and verify behaviour if Instance is a failed state", func() {
			By("creating a BMH")
			// bm-icp-gaudi2 is enabled in operator config
			availableHostName, bmhLookupKey := createBMHName(testNamespace, availableHostPrefix)
			bmh := newBMH(availableHostName, testNamespace, baremetalv1alpha1.StateProvisioned,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2"), bmenrollment.ImagingLabel)
			Expect(k8sClient.Create(ctx, bmh)).
				Should(Succeed())

			//Create an Instance in k8s to simulate the instance that is failed.
			Expect(k8sClient.Create(ctx, newInstance(availableHostName+"-validation", testNamespace, cloudv1alpha1.PhaseFailed))).
				Should(Succeed())

			//Validation operator should accept this event mark the bmh with Failed label.
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey, bmhRef)).Should(Succeed())
				g.Expect(bmhRef.Labels).Should(HaveKeyWithValue(bmenrollment.CheckingFailedLabel, "Instance.creation.failed"))
			}, timeout, interval).Should(Succeed())
		})

		It("Create a BMH and verify Validator.Verified state", func() {
			By("creating a BMH")
			// bm-icp-gaudi2 is enabled in operator config
			availableHostName, bmhLookupKey := createBMHName(testNamespace, availableHostPrefix)
			Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName, testNamespace,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2"),
				bmenrollment.CheckingCompletedLabel))).
				Should(Succeed())

			//Create an Instance in k8s to simulate the instance marked Ready by the bm-instance operator.
			Expect(k8sClient.Create(ctx, newInstance(availableHostName+"-validation", testNamespace, cloudv1alpha1.PhaseReady))).
				Should(Succeed())

			//Validation operator should accept this event and execute the clean up phase.
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey, bmhRef)).Should(Succeed())
				g.Expect(bmhRef.Labels).ShouldNot(HaveKeyWithValue(bmenrollment.CheckingCompletedLabel, "true"))
				g.Expect(bmhRef.Labels).Should(HaveKeyWithValue(bmenrollment.VerifiedLabel, "true"))
				g.Expect(bmhRef.Labels).ShouldNot(HaveKey(bmenrollment.ReadyToTestLabel))
			}, timeout, interval).Should(Succeed())
		})

		AfterAll(func() {
			opts := []client.DeleteAllOfOption{
				client.InNamespace(testNamespace),
				client.GracePeriodSeconds(5),
			}
			Expect(k8sClient.DeleteAllOf(ctx, &baremetalv1alpha1.BareMetalHost{}, opts...)).Should(Succeed())
			k8sClient.DeleteAllOf(ctx, &cloudv1alpha1.Instance{}, opts...)
		})
	})
})

var _ = Describe("BM Validation Controller group phase test", Ordered, func() {
	const testNamespace1 = "test-validation-3"
	const testNamespace2 = "test-validation-4"
	const availableHostPrefix = "available-host"

	// // Object Reference
	bmhRef1 := &baremetalv1alpha1.BareMetalHost{}
	bmhRef2 := &baremetalv1alpha1.BareMetalHost{}
	clusterId := "cluster-0"

	var _ = Describe("creating a BM Instance", func() {
		ctx := context.Background()

		It("Init", func() {
			By("creating namespaces for BMH")
			Expect(k8sClient.Create(ctx, newBmNamespace(testNamespace1))).Should(Succeed())
			Expect(k8sClient.Create(ctx, newBmNamespace(testNamespace2))).Should(Succeed())
		})

		It("Create a Cluster BMH and verify deprovisioning stuck state", func() {
			By("creating a BMH, stuck in deprovisioning state")
			deprovisioningHost, bmhLookupKey3 := createBMHName(testNamespace1, availableHostPrefix)
			// bm-icp-gaudi2 is enabled in operator config
			bmh := newBMH(deprovisioningHost, testNamespace1, baremetalv1alpha1.StateDeprovisioning,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2"), bmenrollment.ClusterGroupID+"=cluster-1")
			bmh.Status.OperationHistory.Deprovision.Start.Time = time.Now().UTC().Add(-(validation.TimeoutBMHStateMinutes + 1) * time.Minute)
			Expect(k8sClient.Create(ctx, bmh)).Should(Succeed())
			bmhRef := &baremetalv1alpha1.BareMetalHost{}
			//Validation operator should detect that the BMH is stuck in deprovisioning state for time greater than threshold.
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey3, bmhRef)).Should(Succeed())
				g.Expect(bmhRef.Labels).Should(HaveKeyWithValue(bmenrollment.CheckingFailedLabel, "Timeout.waiting.for.deprovision.to.complete"))
				g.Expect(bmhRef.Labels).ShouldNot(HaveKey(bmenrollment.VerifiedLabel))
				g.Expect(bmhRef.Labels).ShouldNot(HaveKey(bmenrollment.ValidationIdLabel))
			}, timeout, interval).Should(Succeed())
		})

		It("Verify Validation happens atleast two nodes in a Group", func() {
			By("creating multiple BMH")
			availableHostName1, bmhLookupKey1 := createBMHName(testNamespace1, availableHostPrefix)
			Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName1, testNamespace1,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2"),
				bmenrollment.ClusterGroupID+"="+clusterId))).
				Should(Succeed())
			availableHostName2, bmhLookupKey2 := createBMHName(testNamespace1, availableHostPrefix)
			Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName2, testNamespace1,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2"),
				bmenrollment.ClusterGroupID+"="+clusterId))).
				Should(Succeed())

			//Validation operator should accept this event and start validation of single node
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey1, bmhRef1)).Should(Succeed())
				g.Expect(bmhRef1.Labels).Should(HaveKeyWithValue(bmenrollment.ImagingLabel, "true"))
				g.Expect(bmhRef1.Labels).Should(HaveKey(bmenrollment.ReadyToTestLabel))
			}, timeout, interval).Should(Succeed())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey2, bmhRef2)).Should(Succeed())
				g.Expect(bmhRef2.Labels).Should(HaveKeyWithValue(bmenrollment.ImagingLabel, "true"))
				g.Expect(bmhRef2.Labels).Should(HaveKey(bmenrollment.ReadyToTestLabel))
			}, timeout, interval).Should(Succeed())

		})

		It("Create group BMHs to simulate a isGroupAvailable as false and then verify Validator.STATE_BEGIN state", func() {
			By("creating first BMH in deprovisioning state")
			availableHostName1, bmhLookupKey1 := createBMHName(testNamespace1, availableHostPrefix)
			Expect(k8sClient.Create(ctx, newBMH(availableHostName1, testNamespace1, baremetalv1alpha1.StateDeprovisioning,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2"),
				bmenrollment.ClusterGroupID+"="+clusterId))).
				Should(Succeed())

			By("creating second BMH in available state")
			availableHostName2, bmhLookupKey2 := createBMHName(testNamespace1, availableHostPrefix)
			Expect(k8sClient.Create(ctx, newAvailableBmHost(availableHostName2, testNamespace1,
				bmenrollment.ReadyToTestLabel, fmt.Sprintf(bmenrollment.InstanceTypeLabel, "bm-icp-gaudi2"),
				bmenrollment.ClusterGroupID+"="+clusterId))).
				Should(Succeed())

			// The grouping logic will kick in and the validation operator will try to batch both of bmh..
			// Hence will not be updated.
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey1, bmhRef1)).Should(Succeed())
				g.Expect(bmhRef1.Labels).Should(HaveKeyWithValue(bmenrollment.ImagingLabel, "true"))
				g.Expect(bmhRef1.Labels).Should(HaveKey(bmenrollment.ReadyToTestLabel))
			}, 3*time.Second, interval).ShouldNot(Succeed())

			//change the state to available.
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey1, bmhRef1)).Should(Succeed())
				bmhRef1.Status.Provisioning.State = baremetalv1alpha1.StateAvailable
				g.Expect(k8sClient.Update(ctx, bmhRef1)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			// Verify if the Validation operator has processed.
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey1, bmhRef1)).Should(Succeed())
				g.Expect(bmhRef1.Labels).Should(HaveKeyWithValue(bmenrollment.ImagingLabel, "true"))
				g.Expect(bmhRef1.Labels).Should(HaveKey(bmenrollment.ReadyToTestLabel))
			}, timeout, interval).Should(Succeed())
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, bmhLookupKey2, bmhRef1)).Should(Succeed())
				g.Expect(bmhRef1.Labels).Should(HaveKeyWithValue(bmenrollment.ImagingLabel, "true"))
				g.Expect(bmhRef1.Labels).Should(HaveKey(bmenrollment.ReadyToTestLabel))
			}, timeout, interval).Should(Succeed())
		})

		AfterAll(func() {
			opts := []client.DeleteAllOfOption{
				client.InNamespace(testNamespace1),
				client.GracePeriodSeconds(5),
			}
			Expect(k8sClient.DeleteAllOf(ctx, &baremetalv1alpha1.BareMetalHost{}, opts...)).Should(Succeed())
			k8sClient.DeleteAllOf(ctx, &cloudv1alpha1.Instance{}, opts...)
		})
	})
})

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

func newInstance(name, namespace string, phase cloudv1alpha1.InstancePhase) *cloudv1alpha1.Instance {
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
			},
			ClusterGroupId: "test-clustergroup-1a-2",
			ClusterId:      "test-clustergroup-1a-2-4",
		},
		Status: cloudv1alpha1.InstanceStatus{
			Phase: phase, // ensure it is in a ready state
		},
	}
}

func createBmHost(name, namespace string, labels ...string) *baremetalv1alpha1.BareMetalHost {
	labelMap := map[string]string{
		bmenrollment.CPUIDLabel:        "0x00001",
		bmenrollment.CPUCountLabel:     "1",
		bmenrollment.GPUModelNameLabel: "",
		bmenrollment.GPUCountLabel:     "0",
		bmenrollment.HBMModeLabel:      "",
	}
	for _, l := range labels {
		if strings.Contains(l, "=") {
			res := strings.Split(l, "=")
			labelMap[res[0]] = res[1]
		} else {
			labelMap[l] = "true"
		}
	}

	bmh := &baremetalv1alpha1.BareMetalHost{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "metal3.io/v1alpha1",
			Kind:       "BareMetalHost",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labelMap,
		},
		Spec: baremetalv1alpha1.BareMetalHostSpec{
			Online:         true,
			BootMode:       baremetalv1alpha1.Legacy,
			BootMACAddress: "11:22:33:44:55:66",
			BMC: baremetalv1alpha1.BMCDetails{
				Address:                        "redfish+http://10.11.12.13:8001/redfish/v1/Systems/1",
				CredentialsName:                "secret-1",
				DisableCertificateVerification: true,
			},
			RootDeviceHints: &baremetalv1alpha1.RootDeviceHints{
				DeviceName: "/dev/vda",
			},
			ConsumerRef: &corev1.ObjectReference{
				Name:      name + "-validation",
				Namespace: namespace,
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

func createBMHName(namespace, namePrefix string) (string, types.NamespacedName) {
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		panic(err) // rand should never fail
	}
	hostName := namePrefix + randomNumber.String()
	bmhLookupKey := types.NamespacedName{Namespace: namespace, Name: hostName}
	return hostName, bmhLookupKey
}

func newAvailableBmHost(name, namespace string, labels ...string) *baremetalv1alpha1.BareMetalHost {
	return newBMH(name, namespace, baremetalv1alpha1.StateAvailable, labels...)
}

func newBMH(name, namespace string, instanceState baremetalv1alpha1.ProvisioningState, labels ...string) *baremetalv1alpha1.BareMetalHost {
	bmh := createBmHost(name, namespace, labels...)
	bmh.Status = baremetalv1alpha1.BareMetalHostStatus{
		Provisioning: baremetalv1alpha1.ProvisionStatus{
			State: instanceState,
		},
		OperationalStatus: baremetalv1alpha1.OperationalStatusOK,
		HardwareDetails: &baremetalv1alpha1.HardwareDetails{
			NIC: []baremetalv1alpha1.NIC{
				{
					PXE:    true,
					MAC:    "11:22:33:44:55:66",
					Model:  "test-nic",
					Name:   "eno0",
					VLANID: 0,
				},
			},
		},
	}
	return bmh
}
