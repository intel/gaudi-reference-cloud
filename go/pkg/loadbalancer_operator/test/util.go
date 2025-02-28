// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
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

func NewLoadbalancer(loadbalancerName, namespace string) *loadbalancerv1alpha1.Loadbalancer {

	return &loadbalancerv1alpha1.Loadbalancer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "private.cloud.intel.com/v1alpha1",
			APIVersion: "Loadbalancer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      loadbalancerName,
			Namespace: namespace,
		},
		Spec: loadbalancerv1alpha1.LoadbalancerSpec{
			Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
				Pool: loadbalancerv1alpha1.VPool{
					Port:              8080,
					LoadBalancingMode: "",
					MinActiveMembers:  0,
					Monitor:           "",
					InstanceSelectors: map[string]string{
						"external-lb": "true",
					},
				},
				VIP: loadbalancerv1alpha1.VServer{
					Port:       9090,
					IPType:     "TCP",
					Persist:    "",
					IPProtocol: "",
				},
				Owner: "",
			}},
			Security: loadbalancerv1alpha1.LoadbalancerSecurity{
				Sourceips: []string{
					"134.134.137.84",
				},
			},
			Labels: nil,
		},
	}
}
