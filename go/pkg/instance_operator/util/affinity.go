// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"fmt"
	"strings"

	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bmenrollment "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
)

const TopologySpreadTopologyKey string = "cloud.intel.com/partition"
const ComputeNodePoolLabelPrefix string = "pool.cloud.intel.com/"
const InstanceTypeLabelPrefix string = "instance-type.cloud.intel.com/"
const NetworkModeKey string = "cloud.intel.com/network-mode"

// Add required affinity for instance type.
func UpdateAffinityForInstanceType(instance *cloudv1alpha1.Instance, affinity *k8sv1.Affinity) error {
	ensureNonEmptyRequiredAffinity(affinity)
	// Add MatchExpression to each NodeSelectorTerm.
	for i := range affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		expr := k8sv1.NodeSelectorRequirement{
			Key:      LabelKeyForInstanceType(instance.Spec.InstanceType),
			Operator: k8sv1.NodeSelectorOpIn,
			Values:   []string{"true"},
		}
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions =
			append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions, expr)
	}
	return nil
}

// If the scheduling request has an empty ComputeNodePools list,
// then assume this is being called by a legacy Compute API Server and that compute node pools should not be considered for scheduling.
// Otherwise, only schedule the instance on a node with at least one matching compute node pool ID.
// This function will add the applicable affinity rule.
func UpdateAffinityForNodePools(instance *cloudv1alpha1.Instance, affinity *k8sv1.Affinity) error {
	if len(instance.Spec.ComputeNodePools) == 0 {
		return nil
	}

	ensureNonEmptyRequiredAffinity(affinity)
	existingTerms := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms

	// Pre-allocate newTerms with estimated necessary capacity to reduce slice expansion operations
	numPools := len(instance.Spec.ComputeNodePools)
	newTerms := make([]k8sv1.NodeSelectorTerm, 0, len(existingTerms)*numPools)

	// Combine existing terms with new pool terms
	for _, term := range existingTerms {
		if len(term.MatchFields) > 0 {
			return fmt.Errorf("MatchFields affinity rule is unsupported and must be empty")
		}
		for _, pool := range instance.Spec.ComputeNodePools {
			// Construct the NodeSelectorRequirement for the current pool
			poolExpr := k8sv1.NodeSelectorRequirement{
				Key:      LabelKeyForComputeNodePools(pool),
				Operator: k8sv1.NodeSelectorOpIn,
				Values:   []string{"true"},
			}

			// Create a combined NodeSelectorTerm including the existing term's expressions and the current pool's expression
			combinedExpressions := append([]k8sv1.NodeSelectorRequirement{}, term.MatchExpressions...) // Make a copy to avoid modifying the original term's slice
			combinedExpressions = append(combinedExpressions, poolExpr)
			newTerms = append(newTerms, k8sv1.NodeSelectorTerm{
				MatchExpressions: combinedExpressions,
			})
		}
	}

	// Update the original node selector terms with the newly combined terms
	affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = newTerms

	return nil
}

// Add required affinity for partition determined by scheduler.
func UpdateAffinityForPartition(instance *cloudv1alpha1.Instance, affinity *k8sv1.Affinity) error {
	ensureNonEmptyRequiredAffinity(affinity)
	if instance.Spec.Partition != "" {
		// Add MatchExpression to each NodeSelectorTerm.
		for i := range affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
			expr := k8sv1.NodeSelectorRequirement{
				Key:      TopologySpreadTopologyKey,
				Operator: k8sv1.NodeSelectorOpIn,
				Values:   []string{instance.Spec.Partition},
			}
			affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions =
				append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions, expr)
		}
	}
	return nil
}

// Add preferred affinity for NodeId determined by scheduler.
func UpdateAffinityForNodeId(instance *cloudv1alpha1.Instance, affinity *k8sv1.Affinity) error {
	// Add preferred affinity for NodeId.
	term := k8sv1.PreferredSchedulingTerm{
		Weight: 50, // range is 1-100
		Preference: k8sv1.NodeSelectorTerm{
			MatchFields: []k8sv1.NodeSelectorRequirement{
				{
					Key:      "metadata.name",
					Operator: k8sv1.NodeSelectorOpIn,
					Values:   []string{instance.Spec.NodeId},
				},
			},
		},
	}
	affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
		affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, term)
	return nil
}

func LabelKeyForInstanceType(instanceType string) string {
	return fmt.Sprintf("instance-type.cloud.intel.com/%s", instanceType)
}

func LabelKeyForComputeNodePools(nodePool string) string {
	return fmt.Sprintf(ComputeNodePoolLabelPrefix+"%s", nodePool)
}

// Ensure that affinity has at least one NodeSelectorTerm.
func ensureNonEmptyRequiredAffinity(affinity *k8sv1.Affinity) {
	if affinity.NodeAffinity == nil {
		affinity.NodeAffinity = &k8sv1.NodeAffinity{}
	}
	if affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &k8sv1.NodeSelector{}
	}
	if len(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms =
			append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, k8sv1.NodeSelectorTerm{})
	}
}

// Add required affinity for BM instance type.
func UpdateAffinityForBMInstanceType(instance *cloudv1alpha1.Instance, affinity *k8sv1.Affinity, allowInGroup bool, net InstanceNetworkInfo) error {
	if affinity.NodeAffinity == nil {
		affinity.NodeAffinity = &k8sv1.NodeAffinity{}
	}
	if affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &k8sv1.NodeSelector{}
	}

	affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms =
		append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, k8sv1.NodeSelectorTerm{})

	unschedulable := k8sv1.NodeSelectorRequirement{
		Key:      bmenrollment.UnschedulableLabel,
		Operator: k8sv1.NodeSelectorOpNotIn,
		Values:   []string{"true"},
	}
	affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
		append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, unschedulable)

	// InstanceType Label
	instanceTypeExp := k8sv1.NodeSelectorRequirement{
		Key:      fmt.Sprintf(bmenrollment.InstanceTypeLabel, strings.ToLower(instance.Spec.InstanceTypeSpec.Name)),
		Operator: k8sv1.NodeSelectorOpIn,
		Values:   []string{"true"},
	}
	// Add matching expression
	affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
		append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, instanceTypeExp)

	// Check if system is been verified
	instanceVerifiedExp := k8sv1.NodeSelectorRequirement{
		Key:      bmenrollment.VerifiedLabel,
		Operator: k8sv1.NodeSelectorOpIn,
		Values:   []string{"true"},
	}
	// Add matching expression
	affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
		append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, instanceVerifiedExp)

	// match instance to the desired groups
	if len(net.SuperComputeGroupIDs) > 0 {
		groupMemberExp := k8sv1.NodeSelectorRequirement{
			Key:      bmenrollment.SuperComputeGroupID,
			Operator: k8sv1.NodeSelectorOpIn,
			Values:   net.SuperComputeGroupIDs,
		}
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
			append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, groupMemberExp)
	}

	if len(net.ClusterGroupIDs) > 0 {
		groupMemberExp := k8sv1.NodeSelectorRequirement{
			Key:      bmenrollment.ClusterGroupID,
			Operator: k8sv1.NodeSelectorOpIn,
			Values:   net.ClusterGroupIDs,
		}
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
			append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, groupMemberExp)
	}

	// match instance to the desired network pool
	allowedNetworkModes := []string{net.NetworkMode}
	if net.NetworkMode != "" {
		if net.NetworkMode == bmenrollment.NetworkModeVVV {
			// allow new single instances or instance group to also use VVX or "" network mode
			allowedNetworkModes = append(allowedNetworkModes, bmenrollment.NetworkModeVVXStandalone, "")
		}
	} else {
		// use VVX by default if network mode is not specified
		allowedNetworkModes = []string{bmenrollment.NetworkModeVVXStandalone, ""}
	}
	networkPoolExp := k8sv1.NodeSelectorRequirement{
		Key:      bmenrollment.NetworkModeLabel,
		Operator: k8sv1.NodeSelectorOpIn,
		Values:   allowedNetworkModes,
	}
	affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
		append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, networkPoolExp)

	// match instance to the desired firmware version if specified
	if fwVersions, ok := instance.Spec.Labels[bmenrollment.FWVersionLabel]; ok && fwVersions != "" {
		versions := strings.Split(fwVersions, "_")
		fwVersionExp := k8sv1.NodeSelectorRequirement{
			Key:      bmenrollment.FWVersionLabel,
			Operator: k8sv1.NodeSelectorOpIn,
			Values:   versions,
		}
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
			append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, fwVersionExp)
	}

	// pod affinity
	if affinity.PodAffinity == nil {
		affinity.PodAffinity = &k8sv1.PodAffinity{}
	}
	if affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = []k8sv1.PodAffinityTerm{}
	}

	// we need to check if this node needs to be in a cluster
	if instance.Spec.InstanceGroup != "" {
		if net.NetworkMode != bmenrollment.NetworkModeXBX {
			// this is skipped for BGP network mode since it can span accross multiple cluster groups
			// ensure that the instance group runs nodes from the same cluster group.
			affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution =
				append(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, k8sv1.PodAffinityTerm{
					TopologyKey: bmenrollment.ClusterGroupID,
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      bmenrollment.ClusterGroup,
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{instance.Spec.InstanceGroup},
							},
						},
					},
				})
		}

		if len(net.SuperComputeGroupIDs) > 0 {
			// ensure that the instance group runs on the nodes from the same SC group
			affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution =
				append(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, k8sv1.PodAffinityTerm{
					TopologyKey: bmenrollment.SuperComputeGroupID,
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      bmenrollment.ClusterGroup,
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{instance.Spec.InstanceGroup},
							},
						},
					},
				})
		}

		// ensure that the instance group runs on the nodes that use the same network mode
		affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution =
			append(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, k8sv1.PodAffinityTerm{
				TopologyKey: bmenrollment.NetworkModeLabel,
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      bmenrollment.ClusterGroup,
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{instance.Spec.InstanceGroup},
						},
					},
				},
			})
	} else {
		if !allowInGroup {
			// filter out nodes that are in a group
			nonefromGroup := k8sv1.NodeSelectorRequirement{
				Key:      bmenrollment.ClusterGroupID,
				Operator: k8sv1.NodeSelectorOpDoesNotExist,
			}
			affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
				append(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, nonefromGroup)
		}
	}

	return nil
}
