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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	plfeature "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/feature"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/internal/cache"
	st "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/testing"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestMostAllocatedScoringStrategy(t *testing.T) {
	defaultResources := []config.ResourceSpec{
		{Name: string(v1.ResourceCPU), Weight: 1},
		{Name: string(v1.ResourceMemory), Weight: 1},
	}
	extendedRes := "abc.com/xyz"
	extendedResourceLeastAllocatedSet := []config.ResourceSpec{
		{Name: string(v1.ResourceCPU), Weight: 1},
		{Name: string(v1.ResourceMemory), Weight: 1},
		{Name: extendedRes, Weight: 1},
	}

	tests := []struct {
		name           string
		requestedPod   *v1.Pod
		nodes          []*v1.Node
		existingPods   []*v1.Pod
		expectedScores framework.NodeScoreList
		resources      []config.ResourceSpec
		wantErrs       field.ErrorList
	}{
		{
			// Node1 scores (used resources) on 0-MaxNodeScore scale
			// CPU Score: (0 * MaxNodeScore)  / 4000 = 0
			// Memory Score: (0 * MaxNodeScore) / 10000 = 0
			// Node1 Score: (0 + 0) / 2 = 0
			// Node2 scores (used resources) on 0-MaxNodeScore scale
			// CPU Score: (0 * MaxNodeScore) / 4000 = 0
			// Memory Score: (0 * MaxNodeScore) / 10000 = 0
			// Node2 Score: (0 + 0) / 2 = 0
			name:         "nothing scheduled, nothing requested",
			requestedPod: st.MakePod().Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "4000", "memory": "10000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "4000", "memory": "10000"}).Obj(),
			},
			existingPods:   nil,
			expectedScores: []framework.NodeScore{{Name: "node1", Score: framework.MinNodeScore}, {Name: "node2", Score: framework.MinNodeScore}},
			resources:      defaultResources,
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Score: (3000 * MaxNodeScore) / 4000 = 75
			// Memory Score: (5000 * MaxNodeScore) / 10000 = 50
			// Node1 Score: (75 + 50) / 2 = 6
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Score: (3000 * MaxNodeScore) / 6000 = 50
			// Memory Score: (5000 * MaxNodeScore) / 10000 = 50
			// Node2 Score: (50 + 50) / 2 = 50
			name: "nothing scheduled, resources requested, differently sized machines",
			requestedPod: st.MakePod().
				Req(map[v1.ResourceName]string{"cpu": "1000", "memory": "2000"}).
				Req(map[v1.ResourceName]string{"cpu": "2000", "memory": "3000"}).
				Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "4000", "memory": "10000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "6000", "memory": "10000"}).Obj(),
			},
			existingPods:   nil,
			expectedScores: []framework.NodeScore{{Name: "node1", Score: 62}, {Name: "node2", Score: 50}},
			resources:      defaultResources,
		},
		{
			name: "Resources not set, nothing scheduled, resources requested, differently sized machines",
			requestedPod: st.MakePod().
				Req(map[v1.ResourceName]string{"cpu": "1000", "memory": "2000"}).
				Req(map[v1.ResourceName]string{"cpu": "2000", "memory": "3000"}).
				Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "4000", "memory": "10000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "6000", "memory": "10000"}).Obj(),
			},
			existingPods:   nil,
			expectedScores: []framework.NodeScore{{Name: "node1", Score: framework.MinNodeScore}, {Name: "node2", Score: framework.MinNodeScore}},
			resources:      nil,
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Score: (6000 * MaxNodeScore) / 10000 = 60
			// Memory Score: (0 * MaxNodeScore) / 20000 = 0
			// Node1 Score: (60 + 0) / 2 = 30
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Score: (6000 * MaxNodeScore) / 10000 = 60
			// Memory Score: (5000 * MaxNodeScore) / 20000 = 25
			// Node2 Score: (60 + 25) / 2 = 42
			name:         "no resources requested, pods scheduled with resources",
			requestedPod: st.MakePod().Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "10000", "memory": "20000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "10000", "memory": "20000"}).Obj(),
			},
			existingPods: []*v1.Pod{
				st.MakePod().Node("node1").Req(map[v1.ResourceName]string{"cpu": "3000", "memory": "0"}).Obj(),
				st.MakePod().Node("node1").Req(map[v1.ResourceName]string{"cpu": "3000", "memory": "0"}).Obj(),
				st.MakePod().Node("node2").Req(map[v1.ResourceName]string{"cpu": "3000", "memory": "0"}).Obj(),
				st.MakePod().Node("node2").Req(map[v1.ResourceName]string{"cpu": "3000", "memory": "5000"}).Obj(),
			},
			expectedScores: []framework.NodeScore{{Name: "node1", Score: 30}, {Name: "node2", Score: 42}},
			resources:      defaultResources,
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Score: (6000 * MaxNodeScore) / 10000 = 60
			// Memory Score: (5000 * MaxNodeScore) / 20000 = 25
			// Node1 Score: (60 + 25) / 2 = 42
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Score: (6000 * MaxNodeScore) / 10000 = 60
			// Memory Score: (10000 * MaxNodeScore) / 20000 = 50
			// Node2 Score: (60 + 50) / 2 = 55
			name: "resources requested, pods scheduled with resources",
			requestedPod: st.MakePod().
				Req(map[v1.ResourceName]string{"cpu": "1000", "memory": "2000"}).
				Req(map[v1.ResourceName]string{"cpu": "2000", "memory": "3000"}).
				Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "10000", "memory": "20000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "10000", "memory": "20000"}).Obj(),
			},
			existingPods: []*v1.Pod{
				st.MakePod().Node("node1").Req(map[v1.ResourceName]string{"cpu": "3000", "memory": "0"}).Obj(),
				st.MakePod().Node("node2").Req(map[v1.ResourceName]string{"cpu": "3000", "memory": "5000"}).Obj(),
			},
			expectedScores: []framework.NodeScore{{Name: "node1", Score: 42}, {Name: "node2", Score: 55}},
			resources:      defaultResources,
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Score: 5000 * MaxNodeScore / 5000 return 100
			// Memory Score: (9000 * MaxNodeScore) / 10000 = 90
			// Node1 Score: (100 + 90) / 2 = 95
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Score: (5000 * MaxNodeScore) / 10000 = 50
			// Memory Score: 8000 *MaxNodeScore / 8000 return 100
			// Node2 Score: (50 + 100) / 2 = 75
			name: "resources requested equal node capacity",
			requestedPod: st.MakePod().
				Req(map[v1.ResourceName]string{"cpu": "2000", "memory": "4000"}).
				Req(map[v1.ResourceName]string{"cpu": "3000", "memory": "5000"}).
				Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "5000", "memory": "10000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "10000", "memory": "9000"}).Obj(),
			},
			existingPods:   nil,
			expectedScores: []framework.NodeScore{{Name: "node1", Score: 95}, {Name: "node2", Score: 75}},
			resources:      defaultResources,
		},
		{
			// CPU Score: (3000 *100) / 4000 = 75
			// Memory Score: (5000 *100) / 10000 = 50
			// Node1 Score: (75 * 1 + 50 * 2) / (1 + 2) = 58
			// CPU Score: (3000 *100) / 6000 = 50
			// Memory Score: (5000 *100) / 10000 = 50
			// Node2 Score: (50 * 1 + 50 * 2) / (1 + 2) = 50
			name: "nothing scheduled, resources requested, differently sized machines",
			requestedPod: st.MakePod().
				Req(map[v1.ResourceName]string{"cpu": "1000", "memory": "2000"}).
				Req(map[v1.ResourceName]string{"cpu": "2000", "memory": "3000"}).
				Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "4000", "memory": "10000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "6000", "memory": "10000"}).Obj(),
			},
			existingPods:   nil,
			expectedScores: []framework.NodeScore{{Name: "node1", Score: 58}, {Name: "node2", Score: 50}},
			resources: []config.ResourceSpec{
				{Name: "memory", Weight: 2},
				{Name: "cpu", Weight: 1},
			},
		},
		{
			// Node1 scores on 0-MaxNodeScore scale
			// CPU Fraction: 300 / 250 = 100%
			// Memory Fraction: 600 / 1000 = 60%
			// Node1 Score: (100 + 60) / 2 = 80
			// Node2 scores on 0-MaxNodeScore scale
			// CPU Fraction: 100 / 250 = 40%
			// Memory Fraction: 200 / 1000 = 20%
			// Node2 Score: (20 + 40) / 2 = 30
			name:         "no resources requested, pods scheduled, nonzero request for resource",
			requestedPod: st.MakePod().Container("container").Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "250m", "memory": "1000Mi"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "250m", "memory": "1000Mi"}).Obj(),
			},
			existingPods: []*v1.Pod{
				st.MakePod().Node("node1").Container("container").Obj(),
				st.MakePod().Node("node1").Container("container").Obj(),
			},
			expectedScores: []framework.NodeScore{{Name: "node1", Score: 80}, {Name: "node2", Score: 30}},
			resources:      defaultResources,
		},
		{
			// Bypass extended resource if the pod does not request.
			// For both nodes: cpuScore and memScore are 50
			// Given that extended resource score are intentionally bypassed,
			// the final scores are:
			// - node1: (50 + 50) / 2 = 50
			// - node2: (50 + 50) / 2 = 50
			name: "bypass extended resource if the pod does not request",
			requestedPod: st.MakePod().
				Req(map[v1.ResourceName]string{"cpu": "1000", "memory": "2000"}).
				Req(map[v1.ResourceName]string{"cpu": "2000", "memory": "3000"}).
				Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "6000", "memory": "10000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "6000", "memory": "10000", v1.ResourceName(extendedRes): "4"}).Obj(),
			},
			resources:      extendedResourceLeastAllocatedSet,
			existingPods:   nil,
			expectedScores: []framework.NodeScore{{Name: "node1", Score: 50}, {Name: "node2", Score: 50}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := framework.NewCycleState()
			snapshot := cache.NewSnapshot(test.existingPods, test.nodes)
			fh, _ := runtime.NewFramework(nil, nil, runtime.WithSnapshotSharedLister(snapshot))

			p, err := NewFit(
				&config.NodeResourcesFitArgs{
					ScoringStrategy: &config.ScoringStrategy{
						Type:      config.MostAllocated,
						Resources: test.resources,
					},
				}, fh, plfeature.Features{})

			if diff := cmp.Diff(test.wantErrs.ToAggregate(), err, ignoreBadValueDetail); diff != "" {
				t.Fatalf("got err (-want,+got):\n%s", diff)
			}
			if err != nil {
				return
			}

			var gotScores framework.NodeScoreList
			for _, n := range test.nodes {
				score, status := p.(framework.ScorePlugin).Score(context.Background(), state, test.requestedPod, n.Name)
				if !status.IsSuccess() {
					t.Errorf("unexpected error: %v", status)
				}
				gotScores = append(gotScores, framework.NodeScore{Name: n.Name, Score: score})
			}

			if diff := cmp.Diff(test.expectedScores, gotScores); diff != "" {
				t.Errorf("Unexpected scores (-want,+got):\n%s", diff)
			}
		})
	}
}
