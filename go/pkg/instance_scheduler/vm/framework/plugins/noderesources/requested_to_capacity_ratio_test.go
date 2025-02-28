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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/apis/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	plfeature "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/feature"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/plugins/helper"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/internal/cache"
	st "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/testing"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestRequestedToCapacityRatioScoringStrategy(t *testing.T) {
	defaultResources := []config.ResourceSpec{
		{Name: string(v1.ResourceCPU), Weight: 1},
		{Name: string(v1.ResourceMemory), Weight: 1},
	}

	shape := []config.UtilizationShapePoint{
		{Utilization: 0, Score: 10},
		{Utilization: 100, Score: 0},
	}

	tests := []struct {
		name           string
		requestedPod   *v1.Pod
		nodes          []*v1.Node
		existingPods   []*v1.Pod
		expectedScores framework.NodeScoreList
		resources      []config.ResourceSpec
		shape          []config.UtilizationShapePoint
		wantErrs       field.ErrorList
	}{
		{
			name:         "nothing scheduled, nothing requested (default - least requested nodes have priority)",
			requestedPod: st.MakePod().Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "4000", "memory": "10000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "4000", "memory": "10000"}).Obj(),
			},
			existingPods: []*v1.Pod{
				st.MakePod().Node("node1").Obj(),
				st.MakePod().Node("node1").Obj(),
			},
			expectedScores: []framework.NodeScore{{Name: "node1", Score: framework.MaxNodeScore}, {Name: "node2", Score: framework.MaxNodeScore}},
			resources:      defaultResources,
			shape:          shape,
		},
		{
			name: "nothing scheduled, resources requested, differently sized machines (default - least requested nodes have priority)",
			requestedPod: st.MakePod().
				Req(map[v1.ResourceName]string{"cpu": "1000", "memory": "2000"}).
				Req(map[v1.ResourceName]string{"cpu": "2000", "memory": "3000"}).
				Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "4000", "memory": "10000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "6000", "memory": "10000"}).Obj(),
			},
			existingPods: []*v1.Pod{
				st.MakePod().Node("node1").Obj(),
				st.MakePod().Node("node1").Obj(),
			},
			expectedScores: []framework.NodeScore{{Name: "node1", Score: 38}, {Name: "node2", Score: 50}},
			resources:      defaultResources,
			shape:          shape,
		},
		{
			name:         "no resources requested, pods scheduled with resources (default - least requested nodes have priority)",
			requestedPod: st.MakePod().Obj(),
			nodes: []*v1.Node{
				st.MakeNode().Name("node1").Capacity(map[v1.ResourceName]string{"cpu": "4000", "memory": "10000"}).Obj(),
				st.MakeNode().Name("node2").Capacity(map[v1.ResourceName]string{"cpu": "6000", "memory": "10000"}).Obj(),
			},
			existingPods: []*v1.Pod{
				st.MakePod().Node("node1").Req(map[v1.ResourceName]string{"cpu": "3000", "memory": "5000"}).Obj(),
				st.MakePod().Node("node2").Req(map[v1.ResourceName]string{"cpu": "3000", "memory": "5000"}).Obj(),
			},
			expectedScores: []framework.NodeScore{{Name: "node1", Score: 38}, {Name: "node2", Score: 50}},
			resources:      defaultResources,
			shape:          shape,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := framework.NewCycleState()
			snapshot := cache.NewSnapshot(test.existingPods, test.nodes)
			fh, _ := runtime.NewFramework(nil, nil, runtime.WithSnapshotSharedLister(snapshot))

			p, err := NewFit(&config.NodeResourcesFitArgs{
				ScoringStrategy: &config.ScoringStrategy{
					Type:      config.RequestedToCapacityRatio,
					Resources: test.resources,
					RequestedToCapacityRatio: &config.RequestedToCapacityRatioParam{
						Shape: shape,
					},
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
				t.Errorf("Unexpected nodes (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestBrokenLinearFunction(t *testing.T) {
	type Assertion struct {
		p        int64
		expected int64
	}
	type Test struct {
		points     []helper.FunctionShapePoint
		assertions []Assertion
	}

	tests := []Test{
		{
			points: []helper.FunctionShapePoint{{Utilization: 10, Score: 1}, {Utilization: 90, Score: 9}},
			assertions: []Assertion{
				{p: -10, expected: 1},
				{p: 0, expected: 1},
				{p: 9, expected: 1},
				{p: 10, expected: 1},
				{p: 15, expected: 1},
				{p: 19, expected: 1},
				{p: 20, expected: 2},
				{p: 89, expected: 8},
				{p: 90, expected: 9},
				{p: 99, expected: 9},
				{p: 100, expected: 9},
				{p: 110, expected: 9},
			},
		},
		{
			points: []helper.FunctionShapePoint{{Utilization: 0, Score: 2}, {Utilization: 40, Score: 10}, {Utilization: 100, Score: 0}},
			assertions: []Assertion{
				{p: -10, expected: 2},
				{p: 0, expected: 2},
				{p: 20, expected: 6},
				{p: 30, expected: 8},
				{p: 40, expected: 10},
				{p: 70, expected: 5},
				{p: 100, expected: 0},
				{p: 110, expected: 0},
			},
		},
		{
			points: []helper.FunctionShapePoint{{Utilization: 0, Score: 2}, {Utilization: 40, Score: 2}, {Utilization: 100, Score: 2}},
			assertions: []Assertion{
				{p: -10, expected: 2},
				{p: 0, expected: 2},
				{p: 20, expected: 2},
				{p: 30, expected: 2},
				{p: 40, expected: 2},
				{p: 70, expected: 2},
				{p: 100, expected: 2},
				{p: 110, expected: 2},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			function := helper.BuildBrokenLinearFunction(test.points)
			for _, assertion := range test.assertions {
				assert.InDelta(t, assertion.expected, function(assertion.p), 0.1, "points=%v, p=%f", test.points, assertion.p)
			}
		})
	}
}

func TestResourceBinPackingSingleExtended(t *testing.T) {
	extendedResource := "intel.com/foo"
	extendedResource1 := map[string]int64{
		"intel.com/foo": 4,
	}
	extendedResource2 := map[string]int64{
		"intel.com/foo": 8,
	}

	extendedResourcePod1 := v1.PodSpec{
		Containers: []v1.Container{
			{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(extendedResource): resource.MustParse("2"),
					},
				},
			},
		},
	}
	machine2Pod := extendedResourcePod1
	machine2Pod.NodeName = "machine2"
	tests := []struct {
		pod            *v1.Pod
		pods           []*v1.Pod
		nodes          []*v1.Node
		expectedScores framework.NodeScoreList
		name           string
	}{
		{
			//  Node1 Score = Node2 Score = 0 as the incoming Pod doesn't request extended resource.
			pod:            st.MakePod().Obj(),
			nodes:          []*v1.Node{makeNode("machine1", 4000, 10000*1024*1024, extendedResource2), makeNode("machine2", 4000, 10000*1024*1024, extendedResource1)},
			expectedScores: []framework.NodeScore{{Name: "machine1", Score: 0}, {Name: "machine2", Score: 0}},
			name:           "nothing scheduled, nothing requested",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := framework.NewCycleState()
			snapshot := cache.NewSnapshot(test.pods, test.nodes)
			fh, _ := runtime.NewFramework(nil, nil, runtime.WithSnapshotSharedLister(snapshot))
			args := config.NodeResourcesFitArgs{
				ScoringStrategy: &config.ScoringStrategy{
					Type: config.RequestedToCapacityRatio,
					Resources: []config.ResourceSpec{
						{Name: "intel.com/foo", Weight: 1},
					},
					RequestedToCapacityRatio: &config.RequestedToCapacityRatioParam{
						Shape: []config.UtilizationShapePoint{
							{Utilization: 0, Score: 0},
							{Utilization: 100, Score: 1},
						},
					},
				},
			}
			p, err := NewFit(&args, fh, plfeature.Features{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var gotList framework.NodeScoreList
			for _, n := range test.nodes {
				score, status := p.(framework.ScorePlugin).Score(context.Background(), state, test.pod, n.Name)
				if !status.IsSuccess() {
					t.Errorf("unexpected error: %v", status)
				}
				gotList = append(gotList, framework.NodeScore{Name: n.Name, Score: score})
			}

			if diff := cmp.Diff(test.expectedScores, gotList); diff != "" {
				t.Errorf("Unexpected nodescore list (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestResourceBinPackingMultipleExtended(t *testing.T) {
	extendedResource1 := "intel.com/foo"
	extendedResource2 := "intel.com/bar"
	extendedResources1 := map[string]int64{
		"intel.com/foo": 4,
		"intel.com/bar": 8,
	}

	extendedResources2 := map[string]int64{
		"intel.com/foo": 8,
		"intel.com/bar": 4,
	}

	extnededResourcePod1 := v1.PodSpec{
		Containers: []v1.Container{
			{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(extendedResource1): resource.MustParse("2"),
						v1.ResourceName(extendedResource2): resource.MustParse("2"),
					},
				},
			},
		},
	}
	machine2Pod := extnededResourcePod1
	machine2Pod.NodeName = "machine2"
	tests := []struct {
		pod            *v1.Pod
		pods           []*v1.Pod
		nodes          []*v1.Node
		expectedScores framework.NodeScoreList
		name           string
	}{
		{

			// resources["intel.com/foo"] = 3
			// resources["intel.com/bar"] = 5
			// Node1 scores (used resources) on 0-10 scale
			// Node1 Score:
			// intel.com/foo:
			// rawScoringFunction(used + requested / available)
			// resourceScoringFunction((0+0),8)
			//  = 0/8 * 100 = 0 = rawScoringFunction(0)
			// intel.com/bar:
			// rawScoringFunction(used + requested / available)
			// resourceScoringFunction((0+0),4)
			//  = 0/4 * 100 = 0 = rawScoringFunction(0)
			// Node1 Score: (0 * 3) + (0 * 5) / 8 = 0

			// Node2 scores (used resources) on 0-10 scale
			// rawScoringFunction(used + requested / available)
			// intel.com/foo:
			// rawScoringFunction(used + requested / available)
			// resourceScoringFunction((0+0),4)
			//  = 0/4 * 100 = 0 = rawScoringFunction(0)
			// intel.com/bar:
			// rawScoringFunction(used + requested / available)
			// resourceScoringFunction((0+0),8)
			//  = 0/8 * 100 = 0 = rawScoringFunction(0)
			// Node2 Score: (0 * 3) + (0 * 5) / 8 = 0

			pod:            st.MakePod().Obj(),
			nodes:          []*v1.Node{makeNode("machine1", 4000, 10000*1024*1024, extendedResources2), makeNode("machine2", 4000, 10000*1024*1024, extendedResources1)},
			expectedScores: []framework.NodeScore{{Name: "machine1", Score: 0}, {Name: "machine2", Score: 0}},
			name:           "nothing scheduled, nothing requested",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := framework.NewCycleState()
			snapshot := cache.NewSnapshot(test.pods, test.nodes)
			fh, _ := runtime.NewFramework(nil, nil, runtime.WithSnapshotSharedLister(snapshot))

			args := config.NodeResourcesFitArgs{
				ScoringStrategy: &config.ScoringStrategy{
					Type: config.RequestedToCapacityRatio,
					Resources: []config.ResourceSpec{
						{Name: "intel.com/foo", Weight: 3},
						{Name: "intel.com/bar", Weight: 5},
					},
					RequestedToCapacityRatio: &config.RequestedToCapacityRatioParam{
						Shape: []config.UtilizationShapePoint{
							{Utilization: 0, Score: 0},
							{Utilization: 100, Score: 1},
						},
					},
				},
			}

			p, err := NewFit(&args, fh, plfeature.Features{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var gotScores framework.NodeScoreList
			for _, n := range test.nodes {
				score, status := p.(framework.ScorePlugin).Score(context.Background(), state, test.pod, n.Name)
				if !status.IsSuccess() {
					t.Errorf("unexpected error: %v", status)
				}
				gotScores = append(gotScores, framework.NodeScore{Name: n.Name, Score: score})
			}

			if diff := cmp.Diff(test.expectedScores, gotScores); diff != "" {
				t.Errorf("Unexpected nodescore list (-want,+got):\n%s", diff)
			}
		})
	}
}
