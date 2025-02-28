// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.

/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package noderesources

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
)

// mostResourceScorer favors nodes with most requested resources.
// It calculates the percentage of memory and CPU requested by pods scheduled on the node, and prioritizes
// based on the maximum of the average of the fraction of requested to capacity.
//
// Details:
// (cpu(MaxNodeScore * sum(requested) / capacity) + memory(MaxNodeScore * sum(requested) / capacity)) / weightSum
func mostResourceScorer(resToWeightMap resourceToWeightMap) func(requested, allocable resourceToValueMap) int64 {
	return func(requested, allocable resourceToValueMap) int64 {
		var nodeScore, weightSum int64
		for resource := range requested {
			weight := resToWeightMap[resource]
			resourceScore := mostRequestedScore(requested[resource], allocable[resource])
			nodeScore += resourceScore * weight
			weightSum += weight
		}
		if weightSum == 0 {
			return 0
		}
		return nodeScore / weightSum
	}
}

// The used capacity is calculated on a scale of 0-MaxNodeScore (MaxNodeScore is
// constant with value set to 100).
// 0 being the lowest priority and 100 being the highest.
// The more resources are used the higher the score is. This function
// is almost a reversed version of noderesources.leastRequestedScore.
func mostRequestedScore(requested, capacity int64) int64 {
	if capacity == 0 {
		return 0
	}
	if requested > capacity {
		// `requested` might be greater than `capacity` because pods with no
		// requests get minimum values.
		requested = capacity
	}

	return (requested * framework.MaxNodeScore) / capacity
}
