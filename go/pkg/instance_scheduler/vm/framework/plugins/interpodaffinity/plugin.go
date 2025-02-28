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

package interpodaffinity

import (
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/parallelize"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/names"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = names.InterPodAffinity
)

var _ framework.PreFilterPlugin = &InterPodAffinity{}
var _ framework.FilterPlugin = &InterPodAffinity{}
var _ framework.PreScorePlugin = &InterPodAffinity{}
var _ framework.ScorePlugin = &InterPodAffinity{}
var _ framework.EnqueueExtensions = &InterPodAffinity{}

// InterPodAffinity is a plugin that checks inter pod affinity
type InterPodAffinity struct {
	parallelizer parallelize.Parallelizer
	args         config.InterPodAffinityArgs
	sharedLister framework.SharedLister
}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *InterPodAffinity) Name() string {
	return Name
}

// EventsToRegister returns the possible events that may make a failed Pod
// schedulable
func (pl *InterPodAffinity) EventsToRegister() []framework.ClusterEvent {
	return []framework.ClusterEvent{
		// All ActionType includes the following events:
		// - Delete. An unschedulable Pod may fail due to violating an existing Pod's anti-affinity constraints,
		// deleting an existing Pod may make it schedulable.
		// - Update. Updating on an existing Pod's labels (e.g., removal) may make
		// an unschedulable Pod schedulable.
		// - Add. An unschedulable Pod may fail due to violating pod-affinity constraints,
		// adding an assigned Pod may make it schedulable.
		{Resource: framework.Pod, ActionType: framework.All},
		{Resource: framework.Node, ActionType: framework.Add | framework.UpdateNodeLabel},
	}
}

// New initializes a new plugin and returns it.
func New(plArgs runtime.Object, h framework.Handle) (framework.Plugin, error) {
	if h.SnapshotSharedLister() == nil {
		return nil, fmt.Errorf("SnapshotSharedlister is nil")
	}
	args, err := getArgs(plArgs)
	if err != nil {
		return nil, err
	}
	pl := &InterPodAffinity{
		parallelizer: h.Parallelizer(),
		args:         args,
		sharedLister: h.SnapshotSharedLister(),
	}

	return pl, nil
}

func getArgs(obj runtime.Object) (config.InterPodAffinityArgs, error) {
	ptr, ok := obj.(*config.InterPodAffinityArgs)
	if !ok {
		return config.InterPodAffinityArgs{}, fmt.Errorf("want args to be of type InterPodAffinityArgs, got %T", obj)
	}
	return *ptr, nil
}
