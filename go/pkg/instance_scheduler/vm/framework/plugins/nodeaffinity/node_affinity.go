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

package nodeaffinity

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/helper"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/names"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
)

// NodeAffinity is a plugin that checks if a pod node selector matches the node label.
type NodeAffinity struct {
	handle              framework.Handle
	addedNodeSelector   *nodeaffinity.NodeSelector
	addedPrefSchedTerms *nodeaffinity.PreferredSchedulingTerms
}

var _ framework.PreFilterPlugin = &NodeAffinity{}
var _ framework.FilterPlugin = &NodeAffinity{}
var _ framework.PreScorePlugin = &NodeAffinity{}
var _ framework.ScorePlugin = &NodeAffinity{}
var _ framework.EnqueueExtensions = &NodeAffinity{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = names.NodeAffinity

	// preScoreStateKey is the key in CycleState to NodeAffinity pre-computed data for Scoring.
	preScoreStateKey = "PreScore" + Name

	// preFilterStateKey is the key in CycleState to NodeAffinity pre-compute data for Filtering.
	preFilterStateKey = "PreFilter" + Name

	// ErrReasonPod is the reason for Pod's node affinity/selector not matching.
	ErrReasonPod = "node(s) didn't match Pod's node affinity/selector"

	// errReasonEnforced is the reason for added node affinity not matching.
	errReasonEnforced = "node(s) didn't match scheduler-enforced node affinity"

	// errReasonConflict is the reason for pod's conflicting affinity rules.
	errReasonConflict = "pod affinity terms conflict"
)

// Name returns name of the plugin. It is used in logs, etc.
func (pl *NodeAffinity) Name() string {
	return Name
}

type preFilterState struct {
	requiredNodeSelectorAndAffinity nodeaffinity.RequiredNodeAffinity
}

// Clone just returns the same state because it is not affected by pod additions or deletions.
func (s *preFilterState) Clone() framework.StateData {
	return s
}

// EventsToRegister returns the possible events that may make a Pod
// failed by this plugin schedulable.
func (pl *NodeAffinity) EventsToRegister() []framework.ClusterEvent {
	return []framework.ClusterEvent{
		{Resource: framework.Node, ActionType: framework.Add | framework.Update},
	}
}

// PreFilter builds and writes cycle state used by Filter.
func (pl *NodeAffinity) PreFilter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	state := &preFilterState{requiredNodeSelectorAndAffinity: nodeaffinity.GetRequiredNodeAffinity(pod)}
	cycleState.Write(preFilterStateKey, state)
	affinity := pod.Spec.Affinity
	if affinity == nil ||
		affinity.NodeAffinity == nil ||
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil ||
		len(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		return nil, nil
	}

	// Check if there is affinity to a specific node and return it.
	terms := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	var nodeNames sets.String
	for _, t := range terms {
		var termNodeNames sets.String
		for _, r := range t.MatchFields {
			if r.Key == metav1.ObjectNameField && r.Operator == v1.NodeSelectorOpIn {
				// The requirements represent ANDed constraints, and so we need to
				// find the intersection of nodes.
				s := sets.NewString(r.Values...)
				if termNodeNames == nil {
					termNodeNames = s
				} else {
					termNodeNames = termNodeNames.Intersection(s)
				}
			}
		}
		if termNodeNames == nil {
			// If this term has no node.Name field affinity,
			// then all nodes are eligible because the terms are ORed.
			return nil, nil
		}
		// If the set is empty, it means the terms had affinity to different
		// sets of nodes, and since they are ANDed, then the pod will not match any node.
		if len(termNodeNames) == 0 {
			return nil, framework.NewStatus(framework.UnschedulableAndUnresolvable, errReasonConflict)
		}
		nodeNames = nodeNames.Union(termNodeNames)
	}
	if nodeNames != nil {
		return &framework.PreFilterResult{NodeNames: nodeNames}, nil
	}
	return nil, nil

}

// PreFilterExtensions not necessary for this plugin as state doesn't depend on pod additions or deletions.
func (pl *NodeAffinity) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

// Filter checks if the Node matches the Pod .spec.affinity.nodeAffinity and
// the plugin's added affinity.
func (pl *NodeAffinity) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	node := nodeInfo.Node()
	if node == nil {
		return framework.NewStatus(framework.Error, "node not found")
	}
	if pl.addedNodeSelector != nil && !pl.addedNodeSelector.Match(node) {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, errReasonEnforced)
	}

	s, err := getPreFilterState(state)
	if err != nil {
		// Fallback to calculate requiredNodeSelector and requiredNodeAffinity
		// here when PreFilter is disabled.
		s = &preFilterState{requiredNodeSelectorAndAffinity: nodeaffinity.GetRequiredNodeAffinity(pod)}
	}

	// Ignore parsing errors for backwards compatibility.
	match, err := s.requiredNodeSelectorAndAffinity.Match(node)
	if err != nil {
		return framework.AsStatus(err)
	}
	if !match {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReasonPod)
	}

	return nil
}

// preScoreState computed at PreScore and used at Score.
type preScoreState struct {
	preferredNodeAffinity *nodeaffinity.PreferredSchedulingTerms
}

// Clone implements the mandatory Clone interface. We don't really copy the data since
// there is no need for that.
func (s *preScoreState) Clone() framework.StateData {
	return s
}

// PreScore builds and writes cycle state used by Score and NormalizeScore.
func (pl *NodeAffinity) PreScore(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	if len(nodes) == 0 {
		return nil
	}
	preferredNodeAffinity, err := getPodPreferredNodeAffinity(pod)
	if err != nil {
		return framework.AsStatus(err)
	}
	state := &preScoreState{
		preferredNodeAffinity: preferredNodeAffinity,
	}
	cycleState.Write(preScoreStateKey, state)
	return nil
}

// Score returns the sum of the weights of the terms that match the Node.
// Terms came from the Pod .spec.affinity.nodeAffinity and from the plugin's
// default affinity.
func (pl *NodeAffinity) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := pl.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.AsStatus(fmt.Errorf("getting node %q from Snapshot: %w", nodeName, err))
	}

	node := nodeInfo.Node()

	var count int64
	if pl.addedPrefSchedTerms != nil {
		count += pl.addedPrefSchedTerms.Score(node)
	}

	s, err := getPreScoreState(state)
	if err != nil {
		// Fallback to calculate preferredNodeAffinity here when PreScore is disabled.
		preferredNodeAffinity, err := getPodPreferredNodeAffinity(pod)
		if err != nil {
			return 0, framework.AsStatus(err)
		}
		s = &preScoreState{
			preferredNodeAffinity: preferredNodeAffinity,
		}
	}

	if s.preferredNodeAffinity != nil {
		count += s.preferredNodeAffinity.Score(node)
	}

	return count, nil
}

// NormalizeScore invoked after scoring all nodes.
func (pl *NodeAffinity) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	return helper.DefaultNormalizeScore(framework.MaxNodeScore, false, scores)
}

// ScoreExtensions of the Score plugin.
func (pl *NodeAffinity) ScoreExtensions() framework.ScoreExtensions {
	return pl
}

// New initializes a new plugin and returns it.
func New(plArgs runtime.Object, h framework.Handle) (framework.Plugin, error) {
	args, err := getArgs(plArgs)
	if err != nil {
		return nil, err
	}
	pl := &NodeAffinity{
		handle: h,
	}
	if args.AddedAffinity != nil {
		if ns := args.AddedAffinity.RequiredDuringSchedulingIgnoredDuringExecution; ns != nil {
			pl.addedNodeSelector, err = nodeaffinity.NewNodeSelector(ns)
			if err != nil {
				return nil, fmt.Errorf("parsing addedAffinity.requiredDuringSchedulingIgnoredDuringExecution: %w", err)
			}
		}
		// TODO: parse requiredDuringSchedulingRequiredDuringExecution when it gets added to the API.
		if terms := args.AddedAffinity.PreferredDuringSchedulingIgnoredDuringExecution; len(terms) != 0 {
			pl.addedPrefSchedTerms, err = nodeaffinity.NewPreferredSchedulingTerms(terms)
			if err != nil {
				return nil, fmt.Errorf("parsing addedAffinity.preferredDuringSchedulingIgnoredDuringExecution: %w", err)
			}
		}
	}
	return pl, nil
}

func getArgs(obj runtime.Object) (config.NodeAffinityArgs, error) {
	ptr, ok := obj.(*config.NodeAffinityArgs)
	if !ok {
		return config.NodeAffinityArgs{}, fmt.Errorf("args are not of type NodeAffinityArgs, got %T", obj)
	}
	return *ptr, nil
}

func getPodPreferredNodeAffinity(pod *v1.Pod) (*nodeaffinity.PreferredSchedulingTerms, error) {
	affinity := pod.Spec.Affinity
	if affinity != nil && affinity.NodeAffinity != nil && affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		return nodeaffinity.NewPreferredSchedulingTerms(affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
	}
	return nil, nil
}

func getPreScoreState(cycleState *framework.CycleState) (*preScoreState, error) {
	c, err := cycleState.Read(preScoreStateKey)
	if err != nil {
		return nil, fmt.Errorf("reading %q from cycleState: %w", preScoreStateKey, err)
	}

	s, ok := c.(*preScoreState)
	if !ok {
		return nil, fmt.Errorf("invalid PreScore state, got type %T", c)
	}
	return s, nil
}

func getPreFilterState(cycleState *framework.CycleState) (*preFilterState, error) {
	c, err := cycleState.Read(preFilterStateKey)
	if err != nil {
		return nil, fmt.Errorf("reading %q from cycleState: %v", preFilterStateKey, err)
	}

	s, ok := c.(*preFilterState)
	if !ok {
		return nil, fmt.Errorf("invalid PreFilter state, got type %T", c)
	}
	return s, nil
}
