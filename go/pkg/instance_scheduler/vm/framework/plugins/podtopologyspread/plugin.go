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

package podtopologyspread

import (
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/parallelize"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/feature"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/names"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// ErrReasonConstraintsNotMatch is used for PodTopologySpread filter error.
	ErrReasonConstraintsNotMatch = "node(s) didn't match pod topology spread constraints"
	// ErrReasonNodeLabelNotMatch is used when the node doesn't hold the required label.
	ErrReasonNodeLabelNotMatch = ErrReasonConstraintsNotMatch + " (missing required label)"
)

var systemDefaultConstraints = []v1.TopologySpreadConstraint{
	{
		TopologyKey:       v1.LabelHostname,
		WhenUnsatisfiable: v1.ScheduleAnyway,
		MaxSkew:           3,
	},
	{
		TopologyKey:       v1.LabelTopologyZone,
		WhenUnsatisfiable: v1.ScheduleAnyway,
		MaxSkew:           5,
	},
}

// PodTopologySpread is a plugin that ensures pod's topologySpreadConstraints is satisfied.
type PodTopologySpread struct {
	systemDefaulted                     bool
	parallelizer                        parallelize.Parallelizer
	defaultConstraints                  []v1.TopologySpreadConstraint
	sharedLister                        framework.SharedLister
	enableMinDomainsInPodTopologySpread bool
}

var _ framework.PreFilterPlugin = &PodTopologySpread{}
var _ framework.FilterPlugin = &PodTopologySpread{}
var _ framework.PreScorePlugin = &PodTopologySpread{}
var _ framework.ScorePlugin = &PodTopologySpread{}
var _ framework.EnqueueExtensions = &PodTopologySpread{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = names.PodTopologySpread
)

// Name returns name of the plugin. It is used in logs, etc.
func (pl *PodTopologySpread) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(plArgs runtime.Object, h framework.Handle, fts feature.Features) (framework.Plugin, error) {
	if h.SnapshotSharedLister() == nil {
		return nil, fmt.Errorf("SnapshotSharedlister is nil")
	}
	args, err := getArgs(plArgs)
	if err != nil {
		return nil, err
	}
	pl := &PodTopologySpread{
		parallelizer:                        h.Parallelizer(),
		sharedLister:                        h.SnapshotSharedLister(),
		defaultConstraints:                  args.DefaultConstraints,
		enableMinDomainsInPodTopologySpread: true,
	}
	return pl, nil
}

func getArgs(obj runtime.Object) (config.PodTopologySpreadArgs, error) {
	ptr, ok := obj.(*config.PodTopologySpreadArgs)
	if !ok {
		return config.PodTopologySpreadArgs{}, fmt.Errorf("want args to be of type PodTopologySpreadArgs, got %T", obj)
	}
	return *ptr, nil
}

// EventsToRegister returns the possible events that may make a Pod
// failed by this plugin schedulable.
func (pl *PodTopologySpread) EventsToRegister() []framework.ClusterEvent {
	return []framework.ClusterEvent{
		// All ActionType includes the following events:
		// - Add. An unschedulable Pod may fail due to violating topology spread constraints,
		// adding an assigned Pod may make it schedulable.
		// - Update. Updating on an existing Pod's labels (e.g., removal) may make
		// an unschedulable Pod schedulable.
		// - Delete. An unschedulable Pod may fail due to violating an existing Pod's topology spread constraints,
		// deleting an existing Pod may make it schedulable.
		{Resource: framework.Pod, ActionType: framework.All},
		// Node add|delete|updateLabel maybe lead an topology key changed,
		// and make these pod in scheduling schedulable or unschedulable.
		{Resource: framework.Node, ActionType: framework.Add | framework.Delete | framework.UpdateNodeLabel},
	}
}
