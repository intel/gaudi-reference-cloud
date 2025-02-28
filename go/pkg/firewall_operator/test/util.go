// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	firewallv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewNamespace(namespace string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func NewFirewallRule(fwRuleName, namespace, vip string) *firewallv1alpha1.FirewallRule {

	return &firewallv1alpha1.FirewallRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "private.cloud.intel.com/v1alpha1",
			APIVersion: "FirewallRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fwRuleName,
			Namespace: namespace,
		},
		Spec: firewallv1alpha1.FirewallRuleSpec{
			DestinationIP: vip,
			Port:          "80",
			Protocol:      "tcp",
			SourceIPs:     []string{"any"},
		},
	}
}
