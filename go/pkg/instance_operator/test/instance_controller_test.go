// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	instancecontroller "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	loadbalancer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/pkg/constants"
)

const (
	interval = time.Millisecond * 500
)

var _ = Describe("InstanceHappyPath, Create instance, happy path", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)

	const (
		testCode           = "happy"
		vnetname           = "default"
		providerVlan       = "1001"
		subnet             = "172.17.0.0/16"
		gateway            = "172.17.0.1/16"
		instanceTypeName   = "tiny-" + testCode
		namespace          = "test-project-" + testCode
		instanceName       = "test-vm-" + testCode
		availabilityZone   = "test-az-" + testCode
		region             = "test-region-" + testCode
		sshPublicKeyName1  = testCode + ".test1.example.com"
		sshPublicKeyName2  = testCode + ".test2.example.com"
		sshPublicKeyName3  = testCode + ".test3.example.com"
		sshPublicKeyValue1 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5BBBBIOtsD1/ftwKnS9yvbdzj++5ybR64IIVO5LLd1RUhZrS+ testuser1.example.com"
		sshPublicKeyValue2 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMeRFFKkDzcHybnQdilGn66jEYxteLF41rh8SNOIyvAz testuser2.example.com"
		sshPublicKeyValue3 = "ssh-ed25519 AAAAC4TmcC7PZDI7HTE5AAAAINeRGGKkDzcHntnQlelGn77jEqwteLF34rh9KNVIvyBx testuser3.example.com"
		testFinalizer      = "private.cloud.intel.com/unit-testing"
	)

	instanceTypeSpec := NewInstanceTypeSpec("tiny")
	interfaces := NewInterfaceSpecs()

	Context("Instance integration tests", func() {
		// Object references
		instanceRef := &cloudv1alpha1.Instance{}
		harvesterVmRef := &kubevirtv1.VirtualMachine{}
		sshProxyTunnelRef := &cloudv1alpha1.SshProxyTunnel{}

		// Object lookup keys
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		harvesterVmLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		harvesterVmiLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		sshProxyLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}

		harvesterVmi := &kubevirtv1.VirtualMachineInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instanceName,
				Namespace: namespace,
			},
			Spec: kubevirtv1.VirtualMachineInstanceSpec{},
		}

		// Object definitions
		nsObject := NewNamespace(namespace)
		// Creating an instance with two SshPublicKeyNames
		instance := NewInstance(namespace, instanceName, availabilityZone, region, instanceTypeSpec, NewSshPublicKeySpecs(sshPublicKeyValue1, sshPublicKeyValue2), interfaces)

		It("Should create namespace successfully", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create Instance with two sshPublicKeyNames successfully", func() {
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		})

		It("InstanceReconciler should allow multiple sshPublicKeys in spec", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Spec.SshPublicKeySpecs).Should(Equal(NewSshPublicKeySpecs(sshPublicKeyValue1, sshPublicKeyValue2)))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should create SSHProxyTunnel object with multiple sshkeys", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, sshProxyLookupKey, sshProxyTunnelRef)).Should(Succeed())
				g.Expect(sshProxyTunnelRef.Spec.SshPublicKeys).Should(Equal([]string{sshPublicKeyValue1, sshPublicKeyValue2}))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should add finalizer to Instance", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.GetFinalizers()).Should(ConsistOf(instancecontroller.InstanceFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should create Kubevirt VirtualMachine", func() {
			Eventually(func() error {
				return k8sClient.Get(ctx, harvesterVmLookupKey, harvesterVmRef)
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Accepted condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				log.Info("InstanceReconciler should set Accepted condition to true", "conditions", instanceRef.Status.Conditions)
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionAccepted)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Phase to Provisioning", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				log.Info("InstanceReconciler should set Phase to Provisioning", "instanceRef.Status.Phase", instanceRef.Status.Phase,
					"instanceRef.Status.Message", instanceRef.Status.Message, "conditions", instanceRef.Status.Conditions)
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseProvisioning))
				g.Expect(instanceRef.Status.Message).Should(MatchRegexp(".*%s.*", cloudv1alpha1.InstanceMessageProvisioningAccepted))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set IP address", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				log.Info("InstanceReconciler should set IP address", "instanceRef.Status.Interfaces", instanceRef.Status.Interfaces)
				g.Expect(len(instanceRef.Status.Interfaces)).Should(Equal(1))
				g.Expect(len(instanceRef.Status.Interfaces[0].Addresses)).Should(Equal(1))
				g.Expect(len(instanceRef.Status.Interfaces[0].Addresses[0])).ShouldNot(Equal(0))
			}, timeout, interval).Should(Succeed())
		})

		It("Simulate Kubevirt operator setting status to Ready (corresponds to Running)", func() {
			harvesterVmRef.Status.PrintableStatus = "Running"
			harvesterVmRef.Status.Conditions = append(harvesterVmRef.Status.Conditions, kubevirtv1.VirtualMachineCondition{
				Type:               kubevirtv1.VirtualMachineReady,
				Status:             v1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			})
			Expect(k8sClient.Status().Update(ctx, harvesterVmRef)).Should(Succeed())
		})

		It("Simulate Kubevirt operator creating Kubevirt VirtualMachineInstance", func() {
			Expect(k8sClient.Create(ctx, harvesterVmi)).Should(Succeed())
			Expect(k8sClient.Get(ctx, harvesterVmiLookupKey, harvesterVmi)).Should(Succeed())
		})

		It("InstanceReconciler should set Running condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				log.Info("InstanceReconciler should set Running condition to true", "annotations", instanceRef.Annotations, "message", instanceRef.Status.Message)
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionRunning)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Phase to Provisioning", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseProvisioning))
				g.Expect(instanceRef.Status.Message).Should(MatchRegexp(".*%s.*", cloudv1alpha1.InstanceMessageRunning))
			}, timeout, interval).Should(Succeed())
		})

		It("Simulate Kubevirt operator setting status to VirtualMachineInstanceAgentConnected (corresponds to StartupComplete)", func() {
			harvesterVmRef.Status.Conditions = append(harvesterVmRef.Status.Conditions, kubevirtv1.VirtualMachineCondition{
				Type:               kubevirtv1.VirtualMachineConditionType(kubevirtv1.VirtualMachineInstanceAgentConnected),
				Status:             v1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			})
			Expect(k8sClient.Status().Update(ctx, harvesterVmRef)).Should(Succeed())
		})

		It("InstanceReconciler should set StartupComplete condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionStartupComplete)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set AgentConnected condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionAgentConnected)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Phase to Ready", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseReady))
				g.Expect(instanceRef.Status.Message).Should(MatchRegexp(".*%s.*", cloudv1alpha1.InstanceMessageStartupComplete))
			}, timeout, interval).Should(Succeed())
		})

		It("Simulate SshProxyController setting SshProxyTunnel.Status.SshProxy address, port, and user", func() {
			Expect(k8sClient.Get(ctx, sshProxyLookupKey, sshProxyTunnelRef)).Should(Succeed())
			sshProxyTunnelRef.Status.ProxyUser = proxyUser
			sshProxyTunnelRef.Status.ProxyAddress = proxyAddress
			sshProxyTunnelRef.Status.ProxyPort = proxyPort
			log.Info("Simulate SshProxyController setting SshProxyTunnel.Status.SshProxy address, port, and user", "sshProxyTunnelRef", sshProxyTunnelRef)
			Expect(k8sClient.Status().Update(ctx, sshProxyTunnelRef)).Should(Succeed())
		})

		It("SshProxyTunnel.Status.SshProxy should have address, port, and user", func() {
			Expect(k8sClient.Get(ctx, sshProxyLookupKey, sshProxyTunnelRef)).Should(Succeed())
			log.Info("SshProxyTunnel.Status.SshProxy should have address, port, and user", "sshProxyRef", sshProxyTunnelRef)
			Expect(sshProxyTunnelRef.Status.ProxyUser).Should(Equal(proxyUser))
			Expect(sshProxyTunnelRef.Status.ProxyAddress).Should(Equal(proxyAddress))
			Expect(sshProxyTunnelRef.Status.ProxyPort).Should(Equal(proxyPort))
		})

		It("InstanceReconciler should set Instance.Status.SshProxy address, port, and user", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				log.Info("InstanceReconciler should set Instance.Status.SshProxy address, port, and user", "instanceRef", instanceRef)
				g.Expect(instanceRef.Status.SshProxy.ProxyUser).Should(Equal(proxyUser))
				g.Expect(instanceRef.Status.SshProxy.ProxyAddress).Should(Equal(proxyAddress))
				g.Expect(instanceRef.Status.SshProxy.ProxyPort).Should(Equal(proxyPort))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set InstanceConditionSshProxyReady condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionSshProxyReady)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should update the public key in SshProxyTunnel when instance sshPublicKeyName is changed", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				instanceRef.Spec.SshPublicKeySpecs = NewSshPublicKeySpecs(sshPublicKeyValue1, sshPublicKeyValue3)
				g.Expect(k8sClient.Update(ctx, instanceRef)).Should(Succeed())
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Spec.SshPublicKeySpecs).Should(Equal(NewSshPublicKeySpecs(sshPublicKeyValue1, sshPublicKeyValue3)))
				g.Expect(k8sClient.Get(ctx, sshProxyLookupKey, sshProxyTunnelRef)).Should(Succeed())
				g.Expect(sshProxyTunnelRef.Spec.SshPublicKeys).Should(Equal([]string{sshPublicKeyValue1, sshPublicKeyValue3}))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should update the SshProxyTunnel to have only one key when instance second sshPublicKey is removed", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				instanceRef.Spec.SshPublicKeySpecs = NewSshPublicKeySpecs(sshPublicKeyValue1)
				g.Expect(k8sClient.Update(ctx, instanceRef)).Should(Succeed())
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Spec.SshPublicKeySpecs).Should(Equal(NewSshPublicKeySpecs(sshPublicKeyValue1)))
				g.Expect(k8sClient.Get(ctx, sshProxyLookupKey, sshProxyTunnelRef)).Should(Succeed())
				g.Expect(sshProxyTunnelRef.Spec.SshPublicKeys).Should(Equal([]string{sshPublicKeyValue1}))
			}, timeout, interval).Should(Succeed())
		})

		It("Simulate Kubevirt operator setting Ready condition (corresponds to Running) to false when the virt-launcher pod cease to exists", func() {
			harvesterVmRef.Status.PrintableStatus = "Stopped"
			updateVMReadyConditonToFalse := kubevirtv1.VirtualMachineCondition{
				Type:               kubevirtv1.VirtualMachineReady,
				Status:             v1.ConditionFalse,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             "PodNotExists",
				Message:            "virt-launcher pod has not yet been scheduled",
			}
			cond := util.FindVirtualMachineStatusCondition(harvesterVmRef.Status.Conditions, kubevirtv1.VirtualMachineReady)
			if cond != nil {
				for i := range harvesterVmRef.Status.Conditions {
					if harvesterVmRef.Status.Conditions[i].Type == kubevirtv1.VirtualMachineReady {
						harvesterVmRef.Status.Conditions[i] = updateVMReadyConditonToFalse
					}
				}
			} else {
				harvesterVmRef.Status.Conditions = append(harvesterVmRef.Status.Conditions, updateVMReadyConditonToFalse)
			}
			Expect(k8sClient.Status().Update(ctx, harvesterVmRef)).Should(Succeed())
		})

		It("InstanceReconciler should set Running condition to false", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionRunning)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionFalse))
				g.Expect(cond.Message).To(ContainSubstring("virt-launcher pod has not yet been scheduled"))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Failed condition to true and Phase to Failed", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionFailed)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseFailed))
				g.Expect(instanceRef.Status.Message).To(ContainSubstring("virt-launcher pod does not exists"))
			}, timeout, interval).Should(Succeed())
		})

		It("Delete instance", func() {
			Eventually(func(g Gomega) {
				By("Adding test finalizer to avoid race condition between instance deletion and phase terminating")
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				instanceRef.SetFinalizers(append(instanceRef.GetFinalizers(), testFinalizer))
				g.Expect(k8sClient.Update(ctx, instanceRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Deleting the instance")
			Expect(k8sClient.Delete(ctx, instanceRef)).Should(Succeed())

			By("Waiting for InstanceReconcile to delete the Kubevirt VirtualMachine and Get to report NotFound error")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, harvesterVmLookupKey, harvesterVmRef)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Waiting for InstanceReconciler to delete SshProxyTunnel and Get to report NotFound error")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, sshProxyLookupKey, sshProxyTunnelRef)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Instance status phase should be Terminating")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseTerminating))
				g.Expect(instanceRef.Status.Message).Should(MatchRegexp(".*%s.*", cloudv1alpha1.InstanceMessageTerminating))
			}, timeout, "1s").Should(Succeed())

			By("Removing testFinalizer from Instance")
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(controllerutil.RemoveFinalizer(instanceRef, testFinalizer)).Should(Equal(true))
				g.Expect(k8sClient.Update(ctx, instanceRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())

			By("Instance status phase should be Terminating (not deleted yet)")
			Consistently(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseTerminating))
				g.Expect(instanceRef.Status.Message).Should(MatchRegexp(".*%s.*", cloudv1alpha1.InstanceMessageTerminating))
			}, timeout, "1s").Should(Succeed())

			By("Simulating Harvester by deleting Kubevirt VirtualMachineInstance")
			Expect(k8sClient.Get(ctx, harvesterVmiLookupKey, harvesterVmi)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, harvesterVmi)).Should(Succeed())

			By("Waiting for InstanceReconciler to delete Instance and Get to report NotFound error")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, instanceRef)
				return errors.IsNotFound(err)
			}, "30s", "1s").Should(BeTrue())
		})
	})
})

var _ = Describe("InstanceProvisioningErrors: Create instance, with provisioning errors", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	const (
		testCode          = "provisioning-errors"
		vnetname          = "default"
		providerVlan      = "1001"
		subnet            = "172.17.0.0/16"
		gateway           = "172.17.0.1/16"
		instanceTypeName  = "tiny-" + testCode
		namespace         = "test-project-" + testCode
		availabilityZone  = "test-az-" + testCode
		region            = "test-region-" + testCode
		instanceName      = "test-vm-" + testCode
		sshPublicKeyName  = testCode + ".test.example.com"
		sshPublicKeyValue = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5BBBBIOtsD1/ftwKnS9yvbdzj++5ybR64IIVO5LLd1RUhZrS+ testuser.example.com"
	)

	instanceTypeSpec := NewInstanceTypeSpec("tiny")
	interfaces := []cloudv1alpha1.InterfaceSpec{
		{
			Name:        "eth0",
			VNet:        "us-dev-1a-default",
			DnsName:     "my-virtual-machine-1.03165859732720551183.us-dev-1.cloud.intel.com",
			Nameservers: []string{"1.1.1.1"},
		},
	}

	Context("Instance integration tests", func() {
		// Object references
		instanceRef := &cloudv1alpha1.Instance{}
		harvesterVmRef := &kubevirtv1.VirtualMachine{}

		// Object lookup keys
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		harvesterVMlookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}

		// Object definitions
		nsObject := NewNamespace(namespace)
		instance := NewInstance(namespace, instanceName, availabilityZone, region, instanceTypeSpec, NewSshPublicKeySpecs(sshPublicKeyValue), interfaces)

		It("Should create namespace successfully", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create Instance successfully", func() {
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		})

		It("InstanceReconciler should add finalizer to Instance", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.GetFinalizers()).Should(ConsistOf(instancecontroller.InstanceFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should create Kubevirt VirtualMachine", func() {
			Eventually(func() error {
				return k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Accepted condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				log.Info("InstanceReconciler should set Accepted condition to true", "conditions", instanceRef.Status.Conditions)
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionAccepted)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Phase to Provisioning", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseProvisioning))
				g.Expect(instanceRef.Status.Message).Should(MatchRegexp(".*%s.*", cloudv1alpha1.InstanceMessageProvisioningAccepted))
			}, timeout, interval).Should(Succeed())
		})

		It("Should delete instance successfully", func() {
			Expect(k8sClient.Delete(ctx, instance)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, instanceRef)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})

var _ = Describe("InstanceOperatorFinalizerRemoval, Remove finalizer only after kubevirt vm has been deleted", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	const (
		testCode                 = "instance-operator-finalizer-removal"
		vnetname                 = "default"
		providerVlan             = "1001"
		subnet                   = "172.17.0.0/16"
		gateway                  = "172.17.0.1/16"
		instanceTypeName         = "tiny-" + testCode
		namespace                = "test-project-" + testCode
		availabilityZone         = "test-az-" + testCode
		region                   = "test-region-" + testCode
		instanceName             = "test-vm-" + testCode
		sshPublicKeyName         = testCode + ".test.example.com"
		sshPublicKeyValue        = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5BBBBIOtsD1/ftwKnS9yvbdzj++5ybR64IIVO5LLd1RUhZrS+ testuser.example.com"
		testKubervirtVMFinalizer = "private.cloud.intel.com/kubevirtvmfinalizer"
	)

	instanceTypeSpec := NewInstanceTypeSpec("tiny")
	interfaces := []cloudv1alpha1.InterfaceSpec{
		{
			Name:        "eth0",
			VNet:        "us-dev-1a-default",
			DnsName:     "my-virtual-machine-1.03165859732720551183.us-dev-1.cloud.intel.com",
			Nameservers: []string{"1.1.1.1"},
		},
	}

	Context("Instance integration tests", func() {
		// Object references
		instanceRef := &cloudv1alpha1.Instance{}
		harvesterVmRef := &kubevirtv1.VirtualMachine{}

		// Object lookup keys
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		harvesterVMlookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}

		// Object definitions
		nsObject := NewNamespace(namespace)
		instance := NewInstance(namespace, instanceName, availabilityZone, region, instanceTypeSpec, NewSshPublicKeySpecs(sshPublicKeyValue), interfaces)

		It("Should create namespace successfully", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create Instance successfully", func() {
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		})

		It("InstanceReconciler should add finalizer to Instance", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.GetFinalizers()).Should(ConsistOf(instancecontroller.InstanceFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should create Kubevirt VirtualMachine", func() {
			Eventually(func() error {
				return k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)
			}, timeout, interval).Should(Succeed())
		})

		It("Instance delete should fail when kubervirtVM finalizer is present", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)).Should(Succeed())
				harvesterVmRef.SetFinalizers(append(harvesterVmRef.GetFinalizers(), testKubervirtVMFinalizer))
				g.Expect(k8sClient.Update(ctx, harvesterVmRef)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, instance)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, instanceRef)
				return errors.IsNotFound(err)
			}, timeout, interval).ShouldNot(BeTrue())
		})

		It("Instance delete should succeed when kubervirtVM finalizer is removed", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)).Should(Succeed())
				harvesterVmRef.SetFinalizers(nil)
				g.Expect(k8sClient.Update(ctx, harvesterVmRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, instanceRef)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})

var _ = Describe("KubevirtVmErrors, Delete kubevirt vm after successful creation", Pending, func() {
	ctx := context.Background()

	const (
		testCode           = "kubevirtvm-errors"
		subnet             = "172.17.0.0/16"
		namespace          = "test-project-" + testCode
		instanceName       = "test-vm-" + testCode
		availabilityZone   = "test-az-" + testCode
		region             = "test-region-" + testCode
		sshPublicKeyName1  = testCode + ".test1.example.com"
		sshPublicKeyName2  = testCode + ".test2.example.com"
		sshPublicKeyValue1 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5BBBBIOtsD1/ftwKnS9yvbdzj++5ybR64IIVO5LLd1RUhZrS+ testuser1.example.com"
		sshPublicKeyValue2 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMeRFFKkDzcHybnQdilGn66jEYxteLF41rh8SNOIyvAz testuser2.example.com"
	)

	instanceTypeSpec := NewInstanceTypeSpec("tiny")
	interfaces := NewInterfaceSpecs()

	Context("Instance integration tests", func() {
		// Object references
		instanceRef := &cloudv1alpha1.Instance{}
		harvesterVmRef := &kubevirtv1.VirtualMachine{}

		// Object lookup keys
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}
		harvesterVMlookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}

		// Object definitions
		nsObject := NewNamespace(namespace)
		// Creating an instance with two SshPublicKeyNames
		instance := NewInstance(namespace, instanceName, availabilityZone, region, instanceTypeSpec, NewSshPublicKeySpecs(sshPublicKeyValue1, sshPublicKeyValue2), interfaces)

		It("Should create namespace successfully", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create Instance with two sshPublicKeyNames successfully", func() {
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		})

		It("InstanceReconciler should create Kubevirt VirtualMachine", func() {
			Eventually(func() error {
				return k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Accepted condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionAccepted)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Phase to Provisioning", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseProvisioning))
				g.Expect(instanceRef.Status.Message).Should(MatchRegexp(".*%s.*", cloudv1alpha1.InstanceMessageProvisioningAccepted))
			}, timeout, interval).Should(Succeed())
		})

		It("Simulate Kubevirt operator setting status to Ready (corresponds to Running)", func() {
			harvesterVmRef.Status.PrintableStatus = "Running"
			harvesterVmRef.Status.Conditions = append(harvesterVmRef.Status.Conditions, kubevirtv1.VirtualMachineCondition{
				Type:               kubevirtv1.VirtualMachineReady,
				Status:             v1.ConditionTrue,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "message",
			})
			Expect(k8sClient.Status().Update(ctx, harvesterVmRef)).Should(Succeed())
		})

		It("InstanceReconciler should set Running condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionRunning)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set StartupComplete condition to true", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				cond := util.FindStatusCondition(instanceRef.Status.Conditions, cloudv1alpha1.InstanceConditionStartupComplete)
				g.Expect(cond).Should(Not(BeNil()))
				g.Expect(cond.Status).Should(Equal(v1.ConditionTrue))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should set Phase to Ready", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseReady))
				g.Expect(instanceRef.Status.Message).Should(MatchRegexp(".*%s.*", cloudv1alpha1.InstanceMessageStartupComplete))
			}, timeout, interval).Should(Succeed())
		})

		It("Simulate Kubevirt vm getting terminated", func() {
			harvesterVmRef.Status.PrintableStatus = "Terminating"
			harvesterVmRef.Status.Conditions[0] = kubevirtv1.VirtualMachineCondition{
				Type:               kubevirtv1.VirtualMachineReady,
				Status:             v1.ConditionFalse,
				LastProbeTime:      metav1.Now(),
				LastTransitionTime: metav1.Now(),
				Reason:             "reason",
				Message:            "virt-launcher pod is terminating",
			}
			Expect(k8sClient.Status().Update(ctx, harvesterVmRef)).Should(Succeed())
			Expect(k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, harvesterVmRef)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})

		It("InstanceReconciler should set Phase to Failed", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.Status.Phase).Should(Equal(cloudv1alpha1.PhaseFailed))
				g.Expect(instanceRef.Status.Message).Should(MatchRegexp(".*%s.*", cloudv1alpha1.InstanceMessageFailed))
			}, timeout, interval).Should(Succeed())
		})

		It("InstanceReconciler should not create Kubevirt vm again", func() {
			Consistently(func() bool {
				err := k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)
				return errors.IsNotFound(err)
			}, "2s", "100ms").Should(BeTrue())
		})
	})
})

var _ = Describe("InstanceOperatorFinalizerRemoval, loadbalancer finalizer removal should trigger reconcile", func() {
	ctx := context.Background()

	const (
		testCode          = "instance-operator-loadbalancer-finalizer-removal"
		namespace         = "test-project-" + testCode
		availabilityZone  = "test-az-" + testCode
		region            = "test-region-" + testCode
		instanceName      = "test-vm-" + testCode
		sshPublicKeyValue = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5BBBBIOtsD1/ftwKnS9yvbdzj++5ybR64IIVO5LLd1RUhZrS+ testuser.example.com"
	)

	instanceTypeSpec := NewInstanceTypeSpec("tiny")
	interfaces := []cloudv1alpha1.InterfaceSpec{
		{
			Name:        "eth0",
			VNet:        "us-dev-1a-default",
			DnsName:     "my-virtual-machine-1.03165859732720551183.us-dev-1.cloud.intel.com",
			Nameservers: []string{"1.1.1.1"},
		},
	}

	Context("Instance integration tests", func() {
		// Object references
		instanceRef := &cloudv1alpha1.Instance{}

		// Object lookup keys
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}

		// Object definitions
		nsObject := NewNamespace(namespace)
		instance := NewInstance(namespace, instanceName, availabilityZone, region, instanceTypeSpec, NewSshPublicKeySpecs(sshPublicKeyValue), interfaces)

		It("Should create namespace successfully", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create Instance successfully", func() {
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		})

		It("InstanceReconciler should add finalizer to Instance", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.GetFinalizers()).Should(ConsistOf(instancecontroller.InstanceFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("Instance delete should fail when loadbalancer finalizer is present", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				instanceRef.SetFinalizers(append(instanceRef.GetFinalizers(), loadbalancer.LoadbalancerFinalizer))
				g.Expect(k8sClient.Update(ctx, instanceRef)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, instance)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
			// TODO Wait long enough here to ensure that Reconcile is complete. Reconcile must complete
			// so this test can confirm that loadbalancer finalizer removal will trigger another Reconcile.
			time.Sleep(5 * time.Second)
			Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		})

		It("Instance delete should succeed when loadbalancer finalizer is removed", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				instanceRef.SetFinalizers([]string{instancecontroller.InstanceFinalizer})
				g.Expect(k8sClient.Update(ctx, instanceRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, instanceRef)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})

var _ = Describe("InstanceOperatorFinalizerRemoval, delete should succeed when only the instance finalizer is remaining", func() {
	ctx := context.Background()

	const (
		testCode          = "instance-operator-finalizer-removal-inorder"
		namespace         = "test-project-" + testCode
		availabilityZone  = "test-az-" + testCode
		region            = "test-region-" + testCode
		instanceName      = "test-vm-" + testCode
		sshPublicKeyValue = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5BBBBIOtsD1/ftwKnS9yvbdzj++5ybR64IIVO5LLd1RUhZrS+ testuser.example.com"
	)

	instanceTypeSpec := NewInstanceTypeSpec("tiny")
	interfaces := []cloudv1alpha1.InterfaceSpec{
		{
			Name:        "eth0",
			VNet:        "us-dev-1a-default",
			DnsName:     "my-virtual-machine-1.03165859732720551183.us-dev-1.cloud.intel.com",
			Nameservers: []string{"1.1.1.1"},
		},
	}

	Context("Instance integration tests", func() {
		// Object references
		instanceRef := &cloudv1alpha1.Instance{}

		// Object lookup keys
		instanceLookupKey := types.NamespacedName{Namespace: namespace, Name: instanceName}

		// Object definitions
		nsObject := NewNamespace(namespace)
		instance := NewInstance(namespace, instanceName, availabilityZone, region, instanceTypeSpec, NewSshPublicKeySpecs(sshPublicKeyValue), interfaces)

		It("Should create namespace successfully", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create Instance successfully", func() {
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		})

		It("InstanceReconciler should add finalizer to Instance", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(instanceRef.GetFinalizers()).Should(ConsistOf(instancecontroller.InstanceFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("Should add loadbalancer finalizer to Instance", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				instanceRef.SetFinalizers(append(instanceRef.GetFinalizers(), loadbalancer.LoadbalancerFinalizer))
				g.Expect(k8sClient.Update(ctx, instanceRef)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
		})

		It("Instance delete should succeed when loadbalancer finalizer is removed", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				instanceRef.SetFinalizers([]string{instancecontroller.InstanceFinalizer})
				g.Expect(k8sClient.Update(ctx, instanceRef)).Should(Succeed())
				g.Expect(k8sClient.Delete(ctx, instance)).Should(Succeed())
			}, timeout, interval).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, instanceRef)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
