// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	firewallv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

const (
	interval         = time.Millisecond * 500
	testCode         = "lbs"
	namespace        = "test-project-" + testCode
	loadbalancerName = "testlb-" + testCode
	firewallName     = "testfw-" + testCode
)

var lbVIPProvisioned = &v1alpha1.Loadbalancer{
	ObjectMeta: metav1.ObjectMeta{
		Name:      loadbalancerName,
		Namespace: namespace,
	},
	Spec: v1alpha1.LoadbalancerSpec{
		Listeners: []v1alpha1.LoadbalancerListener{{
			Pool: v1alpha1.VPool{
				Port:              8080,
				LoadBalancingMode: "",
				MinActiveMembers:  0,
				Monitor:           "",
				InstanceSelectors: map[string]string{
					"external-lb": "true",
				},
			},
			VIP: v1alpha1.VServer{
				Port:       9090,
				IPType:     "TCP",
				Persist:    "",
				IPProtocol: "",
			},
			Owner: "",
		}, {
			Pool: v1alpha1.VPool{
				Port:              8080,
				LoadBalancingMode: "",
				MinActiveMembers:  0,
				Monitor:           "",
				InstanceSelectors: map[string]string{
					"external-lb": "true",
				},
			},
			VIP: v1alpha1.VServer{
				Port:       9091,
				IPType:     "TCP",
				Persist:    "",
				IPProtocol: "",
			},
			Owner: "",
		}},
		Security: v1alpha1.LoadbalancerSecurity{
			Sourceips: []string{"134.134.137.84"},
		},
	},
}

var _ = Describe("InstanceHappyPath, Create instance, happy path", func() {
	ctx := context.Background()

	defer GinkgoRecover()

	Context("Instance integration tests", func() {
		// Object definitions
		nsObject := NewNamespace(namespace)
		loadbalancerRef := &v1alpha1.Loadbalancer{}
		firewallRuleRef1 := &firewallv1alpha1.FirewallRule{}
		firewallRuleRef2 := &firewallv1alpha1.FirewallRule{}

		// Object lookup keys
		loadbalancerLookupKey := types.NamespacedName{Namespace: namespace, Name: loadbalancerName}
		firewallRuleLookupKey1 := types.NamespacedName{Namespace: namespace, Name: loadbalancerName + "-9090"}
		firewallRuleLookupKey2 := types.NamespacedName{Namespace: namespace, Name: loadbalancerName + "-9091"}

		It("Should create namespace successfully", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create Loadbalancer successfully", func() {
			Expect(k8sClient.Create(ctx, lbVIPProvisioned)).Should(Succeed())
		})

		It("LoadbalancerReconciler should add finalizer to Loadbalancer", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, loadbalancerLookupKey, loadbalancerRef)).Should(Succeed())
				g.Expect(loadbalancerRef.GetFinalizers()).Should(ConsistOf(constants.LoadbalancerFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("LoadbalancerReconciler should have a VIP", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, loadbalancerLookupKey, loadbalancerRef)).Should(Succeed())
				g.Expect(loadbalancerRef.Status.Vip).Should(Equal("1.2.3.4"))
			}, timeout, interval).Should(Succeed())
		})

		It("LoadbalancerReconciler should list FirewallRule object", func() {
			Eventually(func(g Gomega) {
				fwList := &firewallv1alpha1.FirewallRuleList{}
				g.Expect(k8sClient.List(ctx, fwList)).Should(Succeed())
				g.Expect(len(fwList.Items)).Should(Equal(2))
			}, timeout, interval).Should(Succeed())
		})

		It("LoadbalancerReconciler should create FirewallRule object", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, firewallRuleLookupKey1, firewallRuleRef1)).Should(Succeed())
				g.Expect(firewallRuleRef1.GetFinalizers()).Should(ConsistOf(constants.LoadbalancerFinalizer))
				g.Expect(firewallRuleRef1.Status.State).Should(Equal(firewallv1alpha1.RECONCILING))

				g.Expect(k8sClient.Get(ctx, firewallRuleLookupKey2, firewallRuleRef2)).Should(Succeed())
				g.Expect(firewallRuleRef2.GetFinalizers()).Should(ConsistOf(constants.LoadbalancerFinalizer))
				g.Expect(firewallRuleRef2.Status.State).Should(Equal(firewallv1alpha1.RECONCILING))
			}, timeout, interval).Should(Succeed())
		})
	})
})
