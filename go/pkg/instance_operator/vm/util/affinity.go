// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	k8sv1 "k8s.io/api/core/v1"
)

// Returns the desired pod affinity for the instance.
// This is used for creating Kubevirt VirtualMachine resources.
func AffinityForInstance(instance *cloudv1alpha1.Instance) (*k8sv1.Affinity, error) {
	affinity := &k8sv1.Affinity{}
	if err := util.UpdateAffinityForInstanceType(instance, affinity); err != nil {
		return nil, err
	}
	if err := util.UpdateAffinityForPartition(instance, affinity); err != nil {
		return nil, err
	}
	if err := util.UpdateAffinityForNodeId(instance, affinity); err != nil {
		return nil, err
	}
	if err := util.UpdateAffinityForNodePools(instance, affinity); err != nil {
		return nil, err
	}
	return affinity, nil
}

// Returns the desired topology spread constraint for the instance.
// Harvester does not currently support TopologySpreadConstraints.
// It requires Kubevirt 0.57.0+. See https://kubevirt.io/2022/changelog-v0.57.0.html.
// This is used for creating Kubevirt VirtualMachine resources.
func TopologySpreadConstraintsForInstance(instance *cloudv1alpha1.Instance) ([]k8sv1.TopologySpreadConstraint, error) {
	return nil, nil
}
