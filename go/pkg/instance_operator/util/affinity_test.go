// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	k8sv1 "k8s.io/api/core/v1"
)

var _ = Describe("Affinity Utils", func() {
	// Helper function to create initial testAffinity
	generateTestAffinity := func(numTerms int) *k8sv1.Affinity {
		terms := make([]k8sv1.NodeSelectorTerm, numTerms)
		for i := 0; i < numTerms; i++ {
			terms[i] = k8sv1.NodeSelectorTerm{
				MatchExpressions: []k8sv1.NodeSelectorRequirement{
					{
						Key:      fmt.Sprintf("instance-type.cloud.intel.com/preexisting%d", i),
						Operator: k8sv1.NodeSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			}
		}
		return &k8sv1.Affinity{
			NodeAffinity: &k8sv1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &k8sv1.NodeSelector{
					NodeSelectorTerms: terms,
				},
			},
		}
	}

	Context("when updating affinity for node pools", func() {
		It("should handle 2 initial affinity terms with 2 compute node pools", func() {
			instance := &cloudv1alpha1.Instance{
				Spec: cloudv1alpha1.InstanceSpec{
					ComputeNodePools: []string{"pool1", "pool2"},
				},
			}
			affinity := generateTestAffinity(2)
			err := UpdateAffinityForNodePools(instance, affinity)
			Expect(err).NotTo(HaveOccurred())
			Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(ConsistOf(
				k8sv1.NodeSelectorTerm{
					MatchExpressions: []k8sv1.NodeSelectorRequirement{
						{
							Key:      "instance-type.cloud.intel.com/preexisting0",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
						{
							Key:      "pool.cloud.intel.com/pool1",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
					MatchFields: nil,
				},
				k8sv1.NodeSelectorTerm{
					MatchExpressions: []k8sv1.NodeSelectorRequirement{
						{
							Key:      "instance-type.cloud.intel.com/preexisting0",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
						{
							Key:      "pool.cloud.intel.com/pool2",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
					MatchFields: nil,
				},
				k8sv1.NodeSelectorTerm{
					MatchExpressions: []k8sv1.NodeSelectorRequirement{
						{
							Key:      "instance-type.cloud.intel.com/preexisting1",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
						{
							Key:      "pool.cloud.intel.com/pool1",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
					MatchFields: nil,
				},
				k8sv1.NodeSelectorTerm{
					MatchExpressions: []k8sv1.NodeSelectorRequirement{
						{
							Key:      "instance-type.cloud.intel.com/preexisting1",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
						{
							Key:      "pool.cloud.intel.com/pool2",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
					MatchFields: nil,
				},
			))
		})

		It("should handle empty compute node pools", func() {
			instance := &cloudv1alpha1.Instance{
				Spec: cloudv1alpha1.InstanceSpec{
					ComputeNodePools: []string{},
				},
			}
			affinity := generateTestAffinity(1)
			err := UpdateAffinityForNodePools(instance, affinity)
			Expect(err).NotTo(HaveOccurred())
			Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(Equal([]k8sv1.NodeSelectorTerm{
				{
					MatchExpressions: []k8sv1.NodeSelectorRequirement{
						{
							Key:      "instance-type.cloud.intel.com/preexisting0",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			}))
		})

		It("should handle empty affinity", func() {
			affinity := generateTestAffinity(0)
			instance := &cloudv1alpha1.Instance{
				Spec: cloudv1alpha1.InstanceSpec{
					ComputeNodePools: []string{"pool1", "pool2"},
				},
			}
			err := UpdateAffinityForNodePools(instance, affinity)
			Expect(err).NotTo(HaveOccurred())
			Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(ConsistOf(
				k8sv1.NodeSelectorTerm{
					MatchExpressions: []k8sv1.NodeSelectorRequirement{
						{
							Key:      "pool.cloud.intel.com/pool1",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
					MatchFields: nil,
				},
				k8sv1.NodeSelectorTerm{
					MatchExpressions: []k8sv1.NodeSelectorRequirement{
						{
							Key:      "pool.cloud.intel.com/pool2",
							Operator: k8sv1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
					MatchFields: nil,
				},
			))
		})
	})
})
