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
	"context"
	"reflect"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/feature"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/internal/cache"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeResourcesBalancedAllocation(t *testing.T) {
	labels1 := map[string]string{
		"foo": "bar",
		"baz": "blah",
	}
	labels2 := map[string]string{
		"bar": "foo",
		"baz": "blah",
	}
	machine1Spec := v1.PodSpec{
		NodeName: "machine1",
	}
	machine2Spec := v1.PodSpec{
		NodeName: "machine2",
	}
	noResources := v1.PodSpec{
		Containers: []v1.Container{},
	}
	cpuOnly := v1.PodSpec{
		NodeName: "machine1",
		Containers: []v1.Container{
			{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("1000m"),
						v1.ResourceMemory: resource.MustParse("0"),
					},
				},
			},
			{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2000m"),
						v1.ResourceMemory: resource.MustParse("0"),
					},
				},
			},
		},
	}
	cpuOnly2 := cpuOnly
	cpuOnly2.NodeName = "machine2"
	cpuAndMemory := v1.PodSpec{
		NodeName: "machine2",
		Containers: []v1.Container{
			{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("1000m"),
						v1.ResourceMemory: resource.MustParse("2000"),
					},
				},
			},
			{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2000m"),
						v1.ResourceMemory: resource.MustParse("3000"),
					},
				},
			},
		},
	}
	nonZeroContainer := v1.PodSpec{
		Containers: []v1.Container{{}},
	}
	nonZeroContainer1 := v1.PodSpec{
		NodeName:   "machine1",
		Containers: []v1.Container{{}},
	}

	defaultResourceBalancedAllocationSet := []config.ResourceSpec{
		{Name: string(v1.ResourceCPU), Weight: 1},
		{Name: string(v1.ResourceMemory), Weight: 1},
	}
	scalarResource := map[string]int64{
		"nvidia.com/gpu": 8,
	}

	tests := []struct {
		pod          *v1.Pod
		pods         []*v1.Pod
		nodes        []*v1.Node
		expectedList framework.NodeScoreList
		name         string
		args         config.NodeResourcesBalancedAllocationArgs
	}{
		{
			// Node1 scores (remaining resources) on 0-MaxNodeScore scale
			// CPU Fraction: 0 / 4000 = 0%
			// Memory Fraction: 0 / 10000 = 0%
			// Node1 Score: (1-0) * MaxNodeScore = MaxNodeScore
			// Node2 scores (remaining resources) on 0-MaxNodeScore scale
			// CPU Fraction: 0 / 4000 = 0 %
			// Memory Fraction: 0 / 10000 = 0%
			// Node2 Score: (1-0) * MaxNodeScore = MaxNodeScore
			pod:          &v1.Pod{Spec: noResources},
			nodes:        []*v1.Node{makeNode("machine1", 4000, 10000, nil), makeNode("machine2", 4000, 10000, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: framework.MaxNodeScore}, {Name: "machine2", Score: framework.MaxNodeScore}},
			name:         "nothing scheduled, nothing requested",
			args:         config.NodeResourcesBalancedAllocationArgs{Resources: defaultResourceBalancedAllocationSet},
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Fraction: 3000 / 4000= 75%
			// Memory Fraction: 5000 / 10000 = 50%
			// Node1 std: (0.75 - 0.5) / 2 = 0.125
			// Node1 Score: (1 - 0.125)*MaxNodeScore = 87
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Fraction: 3000 / 6000= 50%
			// Memory Fraction: 5000/10000 = 50%
			// Node2 std: 0
			// Node2 Score: (1-0) * MaxNodeScore = MaxNodeScore
			pod:          &v1.Pod{Spec: cpuAndMemory},
			nodes:        []*v1.Node{makeNode("machine1", 4000, 10000, nil), makeNode("machine2", 6000, 10000, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: 87}, {Name: "machine2", Score: framework.MaxNodeScore}},
			name:         "nothing scheduled, resources requested, differently sized machines",
			args:         config.NodeResourcesBalancedAllocationArgs{Resources: defaultResourceBalancedAllocationSet},
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Fraction: 0 / 4000= 0%
			// Memory Fraction: 0 / 10000 = 0%
			// Node1 std: 0
			// Node1 Score: (1-0) * MaxNodeScore = MaxNodeScore
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Fraction: 0 / 4000= 0%
			// Memory Fraction: 0 / 10000 = 0%
			// Node2 std: 0
			// Node2 Score: (1-0) * MaxNodeScore = MaxNodeScore
			pod:          &v1.Pod{Spec: noResources},
			nodes:        []*v1.Node{makeNode("machine1", 4000, 10000, nil), makeNode("machine2", 4000, 10000, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: framework.MaxNodeScore}, {Name: "machine2", Score: framework.MaxNodeScore}},
			name:         "no resources requested, pods without container scheduled",
			pods: []*v1.Pod{
				{Spec: machine1Spec, ObjectMeta: metav1.ObjectMeta{Labels: labels2}},
				{Spec: machine1Spec, ObjectMeta: metav1.ObjectMeta{Labels: labels1}},
				{Spec: machine2Spec, ObjectMeta: metav1.ObjectMeta{Labels: labels1}},
				{Spec: machine2Spec, ObjectMeta: metav1.ObjectMeta{Labels: labels1}},
			},
			args: config.NodeResourcesBalancedAllocationArgs{Resources: defaultResourceBalancedAllocationSet},
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Fraction: 0 / 250 = 0%
			// Memory Fraction: 0 / 1000 = 0%
			// Node1 std: (0 - 0) / 2 = 0
			// Node1 Score: (1 - 0)*MaxNodeScore = 100
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Fraction: 0 / 250 = 0%
			// Memory Fraction: 0 / 1000 = 0%
			// Node2 std: (0 - 0) / 2 = 0
			// Node2 Score: (1 - 0)*MaxNodeScore = 100
			pod:          &v1.Pod{Spec: nonZeroContainer},
			nodes:        []*v1.Node{makeNode("machine1", 250, 1000*1024*1024, nil), makeNode("machine2", 250, 1000*1024*1024, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: 100}, {Name: "machine2", Score: 100}},
			name:         "no resources requested, pods with container scheduled",
			pods: []*v1.Pod{
				{Spec: nonZeroContainer1},
				{Spec: nonZeroContainer1},
			},
			args: config.NodeResourcesBalancedAllocationArgs{Resources: defaultResourceBalancedAllocationSet},
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Fraction: 6000 / 10000 = 60%
			// Memory Fraction: 0 / 20000 = 0%
			// Node1 std: (0.6 - 0) / 2 = 0.3
			// Node1 Score: (1 - 0.3)*MaxNodeScore = 70
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Fraction: 6000 / 10000 = 60%
			// Memory Fraction: 5000 / 20000 = 25%
			// Node2 std: (0.6 - 0.25) / 2 = 0.175
			// Node2 Score: (1 - 0.175)*MaxNodeScore = 82
			pod:          &v1.Pod{Spec: noResources},
			nodes:        []*v1.Node{makeNode("machine1", 10000, 20000, nil), makeNode("machine2", 10000, 20000, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: 70}, {Name: "machine2", Score: 82}},
			name:         "no resources requested, pods scheduled with resources",
			pods: []*v1.Pod{
				{Spec: cpuOnly, ObjectMeta: metav1.ObjectMeta{Labels: labels2}},
				{Spec: cpuOnly, ObjectMeta: metav1.ObjectMeta{Labels: labels1}},
				{Spec: cpuOnly2, ObjectMeta: metav1.ObjectMeta{Labels: labels1}},
				{Spec: cpuAndMemory, ObjectMeta: metav1.ObjectMeta{Labels: labels1}},
			},
			args: config.NodeResourcesBalancedAllocationArgs{Resources: defaultResourceBalancedAllocationSet},
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Fraction: 6000 / 10000 = 60%
			// Memory Fraction: 5000 / 20000 = 25%
			// Node1 std: (0.6 - 0.25) / 2 = 0.175
			// Node1 Score: (1 - 0.175)*MaxNodeScore = 82
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Fraction: 6000 / 10000 = 60%
			// Memory Fraction: 10000 / 20000 = 50%
			// Node2 std: (0.6 - 0.5) / 2 = 0.05
			// Node2 Score: (1 - 0.05)*MaxNodeScore = 95
			pod:          &v1.Pod{Spec: cpuAndMemory},
			nodes:        []*v1.Node{makeNode("machine1", 10000, 20000, nil), makeNode("machine2", 10000, 20000, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: 82}, {Name: "machine2", Score: 95}},
			name:         "resources requested, pods scheduled with resources",
			pods: []*v1.Pod{
				{Spec: cpuOnly},
				{Spec: cpuAndMemory},
			},
			args: config.NodeResourcesBalancedAllocationArgs{Resources: defaultResourceBalancedAllocationSet},
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Fraction: 6000 / 10000 = 60%
			// Memory Fraction: 5000 / 20000 = 25%
			// Node1 std: (0.6 - 0.25) / 2 = 0.175
			// Node1 Score: (1 - 0.175)*MaxNodeScore = 82
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Fraction: 6000 / 10000 = 60%
			// Memory Fraction: 10000 / 50000 = 20%
			// Node2 std: (0.6 - 0.2) / 2 = 0.2
			// Node2 Score: (1 - 0.2)*MaxNodeScore = 80
			pod:          &v1.Pod{Spec: cpuAndMemory},
			nodes:        []*v1.Node{makeNode("machine1", 10000, 20000, nil), makeNode("machine2", 10000, 50000, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: 82}, {Name: "machine2", Score: 80}},
			name:         "resources requested, pods scheduled with resources, differently sized machines",
			pods: []*v1.Pod{
				{Spec: cpuOnly},
				{Spec: cpuAndMemory},
			},
			args: config.NodeResourcesBalancedAllocationArgs{Resources: defaultResourceBalancedAllocationSet},
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Fraction: 6000 / 6000 = 1
			// Memory Fraction: 0 / 10000 = 0
			// Node1 std: (1 - 0) / 2 = 0.5
			// Node1 Score: (1 - 0.5)*MaxNodeScore = 50
			// Node1 Score: MaxNodeScore - (1 - 0) * MaxNodeScore = 0
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Fraction: 6000 / 6000 = 1
			// Memory Fraction 5000 / 10000 = 50%
			// Node2 std: (1 - 0.5) / 2 = 0.25
			// Node2 Score: (1 - 0.25)*MaxNodeScore = 75
			pod:          &v1.Pod{Spec: cpuOnly},
			nodes:        []*v1.Node{makeNode("machine1", 6000, 10000, nil), makeNode("machine2", 6000, 10000, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: 50}, {Name: "machine2", Score: 75}},
			name:         "requested resources at node capacity",
			pods: []*v1.Pod{
				{Spec: cpuOnly},
				{Spec: cpuAndMemory},
			},
			args: config.NodeResourcesBalancedAllocationArgs{Resources: defaultResourceBalancedAllocationSet},
		},
		{
			pod:          &v1.Pod{Spec: noResources},
			nodes:        []*v1.Node{makeNode("machine1", 0, 0, nil), makeNode("machine2", 0, 0, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: 100}, {Name: "machine2", Score: 100}},
			name:         "zero node resources, pods scheduled with resources",
			pods: []*v1.Pod{
				{Spec: cpuOnly},
				{Spec: cpuAndMemory},
			},
			args: config.NodeResourcesBalancedAllocationArgs{Resources: defaultResourceBalancedAllocationSet},
		},
		// Only one node (machine1) has the scalar resource, pod doesn't request the scalar resource and the scalar resource should be skipped for consideration.
		// Node1: std = 0, score = 100
		// Node2: std = 0, score = 100
		{
			pod:          &v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{}}}},
			nodes:        []*v1.Node{makeNode("machine1", 3500, 40000, scalarResource), makeNode("machine2", 3500, 40000, nil)},
			expectedList: []framework.NodeScore{{Name: "machine1", Score: 100}, {Name: "machine2", Score: 100}},
			name:         "node without the scalar resource results to a higher score",
			pods: []*v1.Pod{
				{Spec: cpuOnly},
				{Spec: cpuOnly2},
			},
			args: config.NodeResourcesBalancedAllocationArgs{Resources: []config.ResourceSpec{
				{Name: string(v1.ResourceCPU), Weight: 1},
				{Name: "nvidia.com/gpu", Weight: 1},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := cache.NewSnapshot(test.pods, test.nodes)
			fh, _ := runtime.NewFramework(nil, nil, runtime.WithSnapshotSharedLister(snapshot))
			p, _ := NewBalancedAllocation(&test.args, fh, feature.Features{})
			for i := range test.nodes {
				hostResult, err := p.(framework.ScorePlugin).Score(context.Background(), nil, test.pod, test.nodes[i].Name)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(test.expectedList[i].Score, hostResult) {
					t.Errorf("got score %v for host %v, expected %v", hostResult, test.nodes[i].Name, test.expectedList[i].Score)
				}
			}
		})
	}
}
