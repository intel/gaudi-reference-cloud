// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/internal/controller"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

const (
	interval = time.Millisecond * 500
)

var _ = Describe("Create firewall rule", func() {
	ctx := context.Background()

	defer GinkgoRecover()

	const (
		testCode         = "lbs"
		loadbalancerName = "testlb-" + testCode
		firewallName     = "testfw-" + testCode
	)

	Context("FirewallRule integration tests", func() {
		// Object definitions
		nsObject := NewNamespace(customerId1)

		// Object lookup keys
		firewallLookupKey := types.NamespacedName{Namespace: customerId1, Name: firewallName}
		firewallRef := &v1alpha1.FirewallRule{}

		// Create a firewallrule
		firewall := NewFirewallRule(firewallName, customerId1, vip1)

		It("Should create namespace successfully", func() {
			Expect(k8sClient.Create(ctx, nsObject)).Should(Succeed())
		})

		It("Should create Firewallrule successfully", func() {
			Expect(k8sClient.Create(ctx, firewall)).Should(Succeed())
		})

		It("FirewallReconciler should add finalizer to Loadbalancer", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, firewallLookupKey, firewallRef)).Should(Succeed())
				g.Expect(firewallRef.GetFinalizers()).Should(ConsistOf(controller.FirewallFinalizer))
			}, timeout, interval).Should(Succeed())
		})

		It("FirewallReconciler should set the state to Reconciling", func() {
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, firewallLookupKey, firewallRef)).Should(Succeed())
				g.Expect(firewallRef.Status.State).Should(Equal(v1alpha1.READY))
			}, timeout, interval).Should(Succeed())
		})
	})
})
