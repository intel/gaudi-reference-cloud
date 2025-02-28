// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package scheduler

import (
	instanceoperatorutil "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Returns the desired pod affinity for the instance.
// This is used for scheduling instances.
// For creating Kubevirt VirtualMachine resources, see go/pkg/instance_operator/vm/util/affinity.go.
func AffinityForInstanceScheduling(instance *cloudv1alpha1.Instance, allowInGroup bool, net instanceoperatorutil.InstanceNetworkInfo) (*k8sv1.Affinity, error) {
	affinity := &k8sv1.Affinity{}
	switch instance.Spec.InstanceTypeSpec.InstanceCategory {
	case cloudv1alpha1.InstanceCategoryBareMetalHost:
		if err := instanceoperatorutil.UpdateAffinityForBMInstanceType(instance, affinity, allowInGroup, net); err != nil {
			return nil, err
		}
	default:
		if err := instanceoperatorutil.UpdateAffinityForInstanceType(instance, affinity); err != nil {
			return nil, err
		}
	}

	// Ensure the instance runs on nodes that possess given pool-labels
	if err := instanceoperatorutil.UpdateAffinityForNodePools(instance, affinity); err != nil {
		return nil, err
	}
	return affinity, nil
}

// Returns the desired topology spread constraint for the instance.
// This is used for scheduling instances.
func TopologySpreadConstraintsForInstanceScheduling(instance *cloudv1alpha1.Instance) ([]k8sv1.TopologySpreadConstraint, error) {
	constraints := []k8sv1.TopologySpreadConstraint{}
	for _, instanceConstraint := range instance.Spec.TopologySpreadConstraints {
		if len(instanceConstraint.LabelSelector.MatchLabels) > 0 {
			constraint := k8sv1.TopologySpreadConstraint{
				MaxSkew:           1,
				TopologyKey:       instanceoperatorutil.TopologySpreadTopologyKey,
				WhenUnsatisfiable: k8sv1.DoNotSchedule,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: instanceConstraint.LabelSelector.MatchLabels,
				},
			}
			constraints = append(constraints, constraint)
		}
	}
	return constraints, nil
}
