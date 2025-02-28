// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.

/*
Copyright 2017 The Kubernetes Authors.

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
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	schedutil "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"

	v1 "k8s.io/api/core/v1"
)

// resourceToWeightMap contains resource name and weight.
type resourceToWeightMap map[v1.ResourceName]int64

// scorer is decorator for resourceAllocationScorer
type scorer func(args *config.NodeResourcesFitArgs) *resourceAllocationScorer

// resourceAllocationScorer contains information to calculate resource allocation score.
type resourceAllocationScorer struct {
	Name string
	// used to decide whether to use Requested or NonZeroRequested for
	// cpu and memory.
	useRequested        bool
	scorer              func(requested, allocable resourceToValueMap) int64
	resourceToWeightMap resourceToWeightMap
}

// resourceToValueMap is keyed with resource name and valued with quantity.
type resourceToValueMap map[v1.ResourceName]int64

// score will use `scorer` function to calculate the score.
func (r *resourceAllocationScorer) score(
	ctx context.Context,
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo) (int64, *framework.Status) {
	log := log.FromContext(ctx).WithName("resourceAllocationScorer.score")
	log.Info("Resource to Weight Map", logkeys.ResourceToWeightMap, r.resourceToWeightMap)
	node := nodeInfo.Node()
	if node == nil {
		return 0, framework.NewStatus(framework.Error, "node not found")
	}
	if r.resourceToWeightMap == nil {
		return 0, framework.NewStatus(framework.Error, "resources not found")
	}

	requested := make(resourceToValueMap)
	allocatable := make(resourceToValueMap)
	for resource := range r.resourceToWeightMap {
		alloc, req := r.calculateResourceAllocatableRequest(ctx, nodeInfo, pod, resource)
		if alloc != 0 {
			// Only fill the extended resource entry when it's non-zero.
			allocatable[resource], requested[resource] = alloc, req
		}
	}

	score := r.scorer(requested, allocatable)

	log.V(3).Info("Listing internal info for allocatable resources, requested resources and score", logkeys.Pod,
		pod, logkeys.Node, node, logkeys.ResourceAllocationScorer, r.Name,
		logkeys.AllocatableResource, allocatable, logkeys.RequestedResource, requested, logkeys.ResourceScore, score,
	)

	return score, nil
}

// calculateResourceAllocatableRequest returns 2 parameters:
// - 1st param: quantity of allocatable resource on the node.
// - 2nd param: aggregated quantity of requested resource on the node.
// Note: if it's an extended resource, and the pod doesn't request it, (0, 0) is returned.
func (r *resourceAllocationScorer) calculateResourceAllocatableRequest(ctx context.Context, nodeInfo *framework.NodeInfo, pod *v1.Pod, resource v1.ResourceName) (int64, int64) {
	log := log.FromContext(ctx).WithName("resourceAllocationScorer.calculateResourceAllocatableRequest")
	requested := nodeInfo.NonZeroRequested
	if r.useRequested {
		requested = nodeInfo.Requested
	}

	podRequest := r.calculatePodResourceRequest(pod, resource)
	// If it's an extended resource, and the pod doesn't request it. We return (0, 0)
	// as an implication to bypass scoring on this resource.
	if podRequest == 0 && schedutil.IsScalarResourceName(resource) {
		return 0, 0
	}
	switch resource {
	case v1.ResourceCPU:
		return nodeInfo.Allocatable.MilliCPU, (requested.MilliCPU + podRequest)
	case v1.ResourceMemory:
		return nodeInfo.Allocatable.Memory, (requested.Memory + podRequest)
	case v1.ResourceEphemeralStorage:
		return nodeInfo.Allocatable.EphemeralStorage, (nodeInfo.Requested.EphemeralStorage + podRequest)
	default:
		if _, exists := nodeInfo.Allocatable.ScalarResources[resource]; exists {
			return nodeInfo.Allocatable.ScalarResources[resource], (nodeInfo.Requested.ScalarResources[resource] + podRequest)
		}
	}
	log.V(3).Info("Requested resource is omitted for node score calculation", logkeys.ResourceName, resource)
	return 0, 0
}

// calculatePodResourceRequest returns the total non-zero requests. If Overhead is defined for the pod and the
// PodOverhead feature is enabled, the Overhead is added to the result.
// podResourceRequest = max(sum(podSpec.Containers), podSpec.InitContainers) + overHead
func (r *resourceAllocationScorer) calculatePodResourceRequest(pod *v1.Pod, resource v1.ResourceName) int64 {
	var podRequest int64
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		value := schedutil.GetRequestForResource(resource, &container.Resources.Requests, !r.useRequested)
		podRequest += value
	}

	for i := range pod.Spec.InitContainers {
		initContainer := &pod.Spec.InitContainers[i]
		value := schedutil.GetRequestForResource(resource, &initContainer.Resources.Requests, !r.useRequested)
		if podRequest < value {
			podRequest = value
		}
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil {
		if quantity, found := pod.Spec.Overhead[resource]; found {
			podRequest += quantity.Value()
		}
	}

	return podRequest
}

// resourcesToWeightMap make weightmap from resources spec
func resourcesToWeightMap(resources []config.ResourceSpec) resourceToWeightMap {
	resourceToWeightMap := make(resourceToWeightMap)
	for _, resource := range resources {
		resourceToWeightMap[v1.ResourceName(resource.Name)] = resource.Weight
	}
	return resourceToWeightMap
}
