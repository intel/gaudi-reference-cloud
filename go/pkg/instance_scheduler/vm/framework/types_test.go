// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.
/*
Copyright 2018 The Kubernetes Authors.

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

package framework

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestNewResource(t *testing.T) {
	tests := []struct {
		name         string
		resourceList v1.ResourceList
		expected     *Resource
	}{
		{
			name:         "empty resource",
			resourceList: map[v1.ResourceName]resource.Quantity{},
			expected:     &Resource{},
		},
		{
			name: "complex resource",
			resourceList: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:                      *resource.NewScaledQuantity(4, -3),
				v1.ResourceMemory:                   *resource.NewQuantity(2000, resource.BinarySI),
				v1.ResourcePods:                     *resource.NewQuantity(80, resource.BinarySI),
				v1.ResourceEphemeralStorage:         *resource.NewQuantity(5000, resource.BinarySI),
				"scalar.test/" + "scalar1":          *resource.NewQuantity(1, resource.DecimalSI),
				v1.ResourceHugePagesPrefix + "test": *resource.NewQuantity(2, resource.BinarySI),
			},
			expected: &Resource{
				MilliCPU:         4,
				Memory:           2000,
				EphemeralStorage: 5000,
				AllowedPodNumber: 80,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewResource(test.resourceList)
			if !reflect.DeepEqual(test.expected, r) {
				t.Errorf("expected: %#v, got: %#v", test.expected, r)
			}
		})
	}
}

func TestResourceClone(t *testing.T) {
	tests := []struct {
		resource *Resource
		expected *Resource
	}{
		{
			resource: &Resource{},
			expected: &Resource{},
		},
		{
			resource: &Resource{
				MilliCPU:         4,
				Memory:           2000,
				EphemeralStorage: 5000,
				AllowedPodNumber: 80,
				ScalarResources:  map[v1.ResourceName]int64{"scalar.test/scalar1": 1, "hugepages-test": 2},
			},
			expected: &Resource{
				MilliCPU:         4,
				Memory:           2000,
				EphemeralStorage: 5000,
				AllowedPodNumber: 80,
				ScalarResources:  map[v1.ResourceName]int64{"scalar.test/scalar1": 1, "hugepages-test": 2},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			r := test.resource.Clone()
			// Modify the field to check if the result is a clone of the origin one.
			test.resource.MilliCPU += 1000
			if !reflect.DeepEqual(test.expected, r) {
				t.Errorf("expected: %#v, got: %#v", test.expected, r)
			}
		})
	}
}

func TestResourceAddScalar(t *testing.T) {
	tests := []struct {
		resource       *Resource
		scalarName     v1.ResourceName
		scalarQuantity int64
		expected       *Resource
	}{
		{
			resource:       &Resource{},
			scalarName:     "scalar1",
			scalarQuantity: 100,
			expected: &Resource{
				ScalarResources: map[v1.ResourceName]int64{"scalar1": 100},
			},
		},
		{
			resource: &Resource{
				MilliCPU:         4,
				Memory:           2000,
				EphemeralStorage: 5000,
				AllowedPodNumber: 80,
				ScalarResources:  map[v1.ResourceName]int64{"hugepages-test": 2},
			},
			scalarName:     "scalar2",
			scalarQuantity: 200,
			expected: &Resource{
				MilliCPU:         4,
				Memory:           2000,
				EphemeralStorage: 5000,
				AllowedPodNumber: 80,
				ScalarResources:  map[v1.ResourceName]int64{"hugepages-test": 2, "scalar2": 200},
			},
		},
	}

	for _, test := range tests {
		t.Run(string(test.scalarName), func(t *testing.T) {
			test.resource.AddScalar(test.scalarName, test.scalarQuantity)
			if !reflect.DeepEqual(test.expected, test.resource) {
				t.Errorf("expected: %#v, got: %#v", test.expected, test.resource)
			}
		})
	}
}

func TestSetMaxResource(t *testing.T) {
	tests := []struct {
		resource     *Resource
		resourceList v1.ResourceList
		expected     *Resource
	}{
		{
			resource: &Resource{},
			resourceList: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:              *resource.NewScaledQuantity(4, -3),
				v1.ResourceMemory:           *resource.NewQuantity(2000, resource.BinarySI),
				v1.ResourceEphemeralStorage: *resource.NewQuantity(5000, resource.BinarySI),
			},
			expected: &Resource{
				MilliCPU:         4,
				Memory:           2000,
				EphemeralStorage: 5000,
			},
		},
		{
			resource: &Resource{
				MilliCPU:         4,
				Memory:           4000,
				EphemeralStorage: 5000,
			},
			resourceList: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:                      *resource.NewScaledQuantity(4, -3),
				v1.ResourceMemory:                   *resource.NewQuantity(2000, resource.BinarySI),
				v1.ResourceEphemeralStorage:         *resource.NewQuantity(7000, resource.BinarySI),
				"scalar.test/scalar1":               *resource.NewQuantity(4, resource.DecimalSI),
				v1.ResourceHugePagesPrefix + "test": *resource.NewQuantity(5, resource.BinarySI),
			},
			expected: &Resource{
				MilliCPU:         4,
				Memory:           4000,
				EphemeralStorage: 7000,
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			test.resource.SetMaxResource(test.resourceList)
			if !reflect.DeepEqual(test.expected, test.resource) {
				t.Errorf("expected: %#v, got: %#v", test.expected, test.resource)
			}
		})
	}
}

type testingMode interface {
	Fatalf(format string, args ...interface{})
}

func makeBasePod(t testingMode, nodeName, objName, cpu, mem, extended string, ports []v1.ContainerPort, volumes []v1.Volume) *v1.Pod {
	req := v1.ResourceList{}
	if cpu != "" {
		req = v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpu),
			v1.ResourceMemory: resource.MustParse(mem),
		}
		if extended != "" {
			parts := strings.Split(extended, ":")
			if len(parts) != 2 {
				t.Fatalf("Invalid extended resource string: \"%s\"", extended)
			}
			req[v1.ResourceName(parts[0])] = resource.MustParse(parts[1])
		}
	}
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID(objName),
			Namespace: "node_info_cache_test",
			Name:      objName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Resources: v1.ResourceRequirements{
					Requests: req,
				},
				Ports: ports,
			}},
			NodeName: nodeName,
			Volumes:  volumes,
		},
	}
}

func TestNewNodeInfo(t *testing.T) {
	nodeName := "test-node"
	pods := []*v1.Pod{
		makeBasePod(t, nodeName, "test-1", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}, nil),
		makeBasePod(t, nodeName, "test-2", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}, nil),
	}

	expected := &NodeInfo{
		Requested: &Resource{
			MilliCPU:         300,
			Memory:           1524,
			EphemeralStorage: 0,
			AllowedPodNumber: 0,
			ScalarResources:  map[v1.ResourceName]int64(nil),
		},
		NonZeroRequested: &Resource{
			MilliCPU:         300,
			Memory:           1524,
			EphemeralStorage: 0,
			AllowedPodNumber: 0,
			ScalarResources:  map[v1.ResourceName]int64(nil),
		},
		Allocatable:  &Resource{},
		Generation:   2,
		ImageStates:  map[string]*ImageStateSummary{},
		PVCRefCounts: map[string]int{},
		Pods: []*PodInfo{
			{
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "node_info_cache_test",
						Name:      "test-1",
						UID:       types.UID("test-1"),
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceCPU:    resource.MustParse("100m"),
										v1.ResourceMemory: resource.MustParse("500"),
									},
								},
								Ports: []v1.ContainerPort{
									{
										HostIP:   "127.0.0.1",
										HostPort: 80,
										Protocol: "TCP",
									},
								},
							},
						},
						NodeName: nodeName,
					},
				},
			},
			{
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "node_info_cache_test",
						Name:      "test-2",
						UID:       types.UID("test-2"),
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceCPU:    resource.MustParse("200m"),
										v1.ResourceMemory: resource.MustParse("1Ki"),
									},
								},
								Ports: []v1.ContainerPort{
									{
										HostIP:   "127.0.0.1",
										HostPort: 8080,
										Protocol: "TCP",
									},
								},
							},
						},
						NodeName: nodeName,
					},
				},
			},
		},
	}

	gen := generation
	ni := NewNodeInfo(pods...)
	if ni.Generation <= gen {
		t.Errorf("Generation is not incremented. previous: %v, current: %v", gen, ni.Generation)
	}
	expected.Generation = ni.Generation
	if !reflect.DeepEqual(expected, ni) {
		t.Errorf("expected: %#v, got: %#v", expected, ni)
	}
}

func TestNodeInfoClone(t *testing.T) {
	nodeName := "test-node"
	tests := []struct {
		nodeInfo *NodeInfo
		expected *NodeInfo
	}{
		{
			nodeInfo: &NodeInfo{
				Requested:        &Resource{},
				NonZeroRequested: &Resource{},
				Allocatable:      &Resource{},
				Generation:       2,
				ImageStates:      map[string]*ImageStateSummary{},
				PVCRefCounts:     map[string]int{},
				Pods: []*PodInfo{
					{
						Pod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "node_info_cache_test",
								Name:      "test-1",
								UID:       types.UID("test-1"),
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Resources: v1.ResourceRequirements{
											Requests: v1.ResourceList{
												v1.ResourceCPU:    resource.MustParse("100m"),
												v1.ResourceMemory: resource.MustParse("500"),
											},
										},
										Ports: []v1.ContainerPort{
											{
												HostIP:   "127.0.0.1",
												HostPort: 80,
												Protocol: "TCP",
											},
										},
									},
								},
								NodeName: nodeName,
							},
						},
					},
					{
						Pod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "node_info_cache_test",
								Name:      "test-2",
								UID:       types.UID("test-2"),
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Resources: v1.ResourceRequirements{
											Requests: v1.ResourceList{
												v1.ResourceCPU:    resource.MustParse("200m"),
												v1.ResourceMemory: resource.MustParse("1Ki"),
											},
										},
										Ports: []v1.ContainerPort{
											{
												HostIP:   "127.0.0.1",
												HostPort: 8080,
												Protocol: "TCP",
											},
										},
									},
								},
								NodeName: nodeName,
							},
						},
					},
				},
			},
			expected: &NodeInfo{
				Requested:        &Resource{},
				NonZeroRequested: &Resource{},
				Allocatable:      &Resource{},
				Generation:       2,
				ImageStates:      map[string]*ImageStateSummary{},
				PVCRefCounts:     map[string]int{},
				Pods: []*PodInfo{
					{
						Pod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "node_info_cache_test",
								Name:      "test-1",
								UID:       types.UID("test-1"),
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Resources: v1.ResourceRequirements{
											Requests: v1.ResourceList{
												v1.ResourceCPU:    resource.MustParse("100m"),
												v1.ResourceMemory: resource.MustParse("500"),
											},
										},
										Ports: []v1.ContainerPort{
											{
												HostIP:   "127.0.0.1",
												HostPort: 80,
												Protocol: "TCP",
											},
										},
									},
								},
								NodeName: nodeName,
							},
						},
					},
					{
						Pod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "node_info_cache_test",
								Name:      "test-2",
								UID:       types.UID("test-2"),
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Resources: v1.ResourceRequirements{
											Requests: v1.ResourceList{
												v1.ResourceCPU:    resource.MustParse("200m"),
												v1.ResourceMemory: resource.MustParse("1Ki"),
											},
										},
										Ports: []v1.ContainerPort{
											{
												HostIP:   "127.0.0.1",
												HostPort: 8080,
												Protocol: "TCP",
											},
										},
									},
								},
								NodeName: nodeName,
							},
						},
					},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			ni := test.nodeInfo.Clone()
			// Modify the field to check if the result is a clone of the origin one.
			test.nodeInfo.Generation += 10
			if !reflect.DeepEqual(test.expected, ni) {
				t.Errorf("expected: %#v, got: %#v", test.expected, ni)
			}
		})
	}
}

func TestNodeInfoAddPod(t *testing.T) {
	nodeName := "test-node"
	pods := []*v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "node_info_cache_test",
				Name:      "test-1",
				UID:       types.UID("test-1"),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("100m"),
								v1.ResourceMemory: resource.MustParse("500"),
							},
						},
						Ports: []v1.ContainerPort{
							{
								HostIP:   "127.0.0.1",
								HostPort: 80,
								Protocol: "TCP",
							},
						},
					},
				},
				NodeName: nodeName,
				Overhead: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("500m"),
				},
				Volumes: []v1.Volume{
					{
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: "pvc-1",
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "node_info_cache_test",
				Name:      "test-2",
				UID:       types.UID("test-2"),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("200m"),
							},
						},
						Ports: []v1.ContainerPort{
							{
								HostIP:   "127.0.0.1",
								HostPort: 8080,
								Protocol: "TCP",
							},
						},
					},
				},
				NodeName: nodeName,
				Overhead: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("500m"),
					v1.ResourceMemory: resource.MustParse("500"),
				},
				Volumes: []v1.Volume{
					{
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: "pvc-1",
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "node_info_cache_test",
				Name:      "test-3",
				UID:       types.UID("test-3"),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU: resource.MustParse("200m"),
							},
						},
						Ports: []v1.ContainerPort{
							{
								HostIP:   "127.0.0.1",
								HostPort: 8080,
								Protocol: "TCP",
							},
						},
					},
				},
				InitContainers: []v1.Container{
					{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("500m"),
								v1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
					},
				},
				NodeName: nodeName,
				Overhead: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("500m"),
					v1.ResourceMemory: resource.MustParse("500"),
				},
				Volumes: []v1.Volume{
					{
						VolumeSource: v1.VolumeSource{
							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
								ClaimName: "pvc-2",
							},
						},
					},
				},
			},
		},
	}
	expected := &NodeInfo{
		node: &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		},
		Requested: &Resource{
			MilliCPU:         2300,
			Memory:           209716700, //1500 + 200MB in initContainers
			EphemeralStorage: 0,
			AllowedPodNumber: 0,
			ScalarResources:  map[v1.ResourceName]int64(nil),
		},
		NonZeroRequested: &Resource{
			MilliCPU:         2300,
			Memory:           419431900, //200MB(initContainers) + 200MB(default memory value) + 1500 specified in requests/overhead
			EphemeralStorage: 0,
			AllowedPodNumber: 0,
			ScalarResources:  map[v1.ResourceName]int64(nil),
		},
		Allocatable:  &Resource{},
		Generation:   2,
		ImageStates:  map[string]*ImageStateSummary{},
		PVCRefCounts: map[string]int{"node_info_cache_test/pvc-1": 2, "node_info_cache_test/pvc-2": 1},
		Pods: []*PodInfo{
			{
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "node_info_cache_test",
						Name:      "test-1",
						UID:       types.UID("test-1"),
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceCPU:    resource.MustParse("100m"),
										v1.ResourceMemory: resource.MustParse("500"),
									},
								},
								Ports: []v1.ContainerPort{
									{
										HostIP:   "127.0.0.1",
										HostPort: 80,
										Protocol: "TCP",
									},
								},
							},
						},
						NodeName: nodeName,
						Overhead: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("500m"),
						},
						Volumes: []v1.Volume{
							{
								VolumeSource: v1.VolumeSource{
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
										ClaimName: "pvc-1",
									},
								},
							},
						},
					},
				},
			},
			{
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "node_info_cache_test",
						Name:      "test-2",
						UID:       types.UID("test-2"),
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceCPU: resource.MustParse("200m"),
									},
								},
								Ports: []v1.ContainerPort{
									{
										HostIP:   "127.0.0.1",
										HostPort: 8080,
										Protocol: "TCP",
									},
								},
							},
						},
						NodeName: nodeName,
						Overhead: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("500m"),
							v1.ResourceMemory: resource.MustParse("500"),
						},
						Volumes: []v1.Volume{
							{
								VolumeSource: v1.VolumeSource{
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
										ClaimName: "pvc-1",
									},
								},
							},
						},
					},
				},
			},
			{
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "node_info_cache_test",
						Name:      "test-3",
						UID:       types.UID("test-3"),
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceCPU: resource.MustParse("200m"),
									},
								},
								Ports: []v1.ContainerPort{
									{
										HostIP:   "127.0.0.1",
										HostPort: 8080,
										Protocol: "TCP",
									},
								},
							},
						},
						InitContainers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceCPU:    resource.MustParse("500m"),
										v1.ResourceMemory: resource.MustParse("200Mi"),
									},
								},
							},
						},
						NodeName: nodeName,
						Overhead: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("500m"),
							v1.ResourceMemory: resource.MustParse("500"),
						},
						Volumes: []v1.Volume{
							{
								VolumeSource: v1.VolumeSource{
									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
										ClaimName: "pvc-2",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ni := fakeNodeInfo()
	gen := ni.Generation
	for _, pod := range pods {
		ni.AddPod(pod)
		if ni.Generation <= gen {
			t.Errorf("Generation is not incremented. Prev: %v, current: %v", gen, ni.Generation)
		}
		gen = ni.Generation
	}

	expected.Generation = ni.Generation
	if !reflect.DeepEqual(expected, ni) {
		t.Errorf("expected: %#v, got: %#v", expected, ni)
	}
}

func TestNodeInfoRemovePod(t *testing.T) {
	nodeName := "test-node"
	pods := []*v1.Pod{
		makeBasePod(t, nodeName, "test-1", "100m", "500", "",
			[]v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}},
			[]v1.Volume{{VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: "pvc-1"}}}}),
		makeBasePod(t, nodeName, "test-2", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}, nil),
	}

	// add pod Overhead
	for _, pod := range pods {
		pod.Spec.Overhead = v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("500"),
		}
	}

	tests := []struct {
		pod              *v1.Pod
		errExpected      bool
		expectedNodeInfo *NodeInfo
	}{
		{
			pod:         makeBasePod(t, nodeName, "non-exist", "0", "0", "", []v1.ContainerPort{{}}, []v1.Volume{}),
			errExpected: true,
			expectedNodeInfo: &NodeInfo{
				node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node",
					},
				},
				Requested: &Resource{
					MilliCPU:         1300,
					Memory:           2524,
					EphemeralStorage: 0,
					AllowedPodNumber: 0,
					ScalarResources:  map[v1.ResourceName]int64(nil),
				},
				NonZeroRequested: &Resource{
					MilliCPU:         1300,
					Memory:           2524,
					EphemeralStorage: 0,
					AllowedPodNumber: 0,
					ScalarResources:  map[v1.ResourceName]int64(nil),
				},
				Allocatable:  &Resource{},
				Generation:   2,
				ImageStates:  map[string]*ImageStateSummary{},
				PVCRefCounts: map[string]int{"node_info_cache_test/pvc-1": 1},
				Pods: []*PodInfo{
					{
						Pod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "node_info_cache_test",
								Name:      "test-1",
								UID:       types.UID("test-1"),
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Resources: v1.ResourceRequirements{
											Requests: v1.ResourceList{
												v1.ResourceCPU:    resource.MustParse("100m"),
												v1.ResourceMemory: resource.MustParse("500"),
											},
										},
										Ports: []v1.ContainerPort{
											{
												HostIP:   "127.0.0.1",
												HostPort: 80,
												Protocol: "TCP",
											},
										},
									},
								},
								NodeName: nodeName,
								Overhead: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("500"),
								},
								Volumes: []v1.Volume{
									{
										VolumeSource: v1.VolumeSource{
											PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
												ClaimName: "pvc-1",
											},
										},
									},
								},
							},
						},
					},
					{
						Pod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "node_info_cache_test",
								Name:      "test-2",
								UID:       types.UID("test-2"),
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Resources: v1.ResourceRequirements{
											Requests: v1.ResourceList{
												v1.ResourceCPU:    resource.MustParse("200m"),
												v1.ResourceMemory: resource.MustParse("1Ki"),
											},
										},
										Ports: []v1.ContainerPort{
											{
												HostIP:   "127.0.0.1",
												HostPort: 8080,
												Protocol: "TCP",
											},
										},
									},
								},
								NodeName: nodeName,
								Overhead: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("500"),
								},
							},
						},
					},
				},
			},
		},
		{
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "node_info_cache_test",
					Name:      "test-1",
					UID:       types.UID("test-1"),
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("500"),
								},
							},
							Ports: []v1.ContainerPort{
								{
									HostIP:   "127.0.0.1",
									HostPort: 80,
									Protocol: "TCP",
								},
							},
						},
					},
					NodeName: nodeName,
					Overhead: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("500m"),
						v1.ResourceMemory: resource.MustParse("500"),
					},
					Volumes: []v1.Volume{
						{
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "pvc-1",
								},
							},
						},
					},
				},
			},
			errExpected: false,
			expectedNodeInfo: &NodeInfo{
				node: &v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-node",
					},
				},
				Requested: &Resource{
					MilliCPU:         700,
					Memory:           1524,
					EphemeralStorage: 0,
					AllowedPodNumber: 0,
					ScalarResources:  map[v1.ResourceName]int64(nil),
				},
				NonZeroRequested: &Resource{
					MilliCPU:         700,
					Memory:           1524,
					EphemeralStorage: 0,
					AllowedPodNumber: 0,
					ScalarResources:  map[v1.ResourceName]int64(nil),
				},
				Allocatable:  &Resource{},
				Generation:   3,
				ImageStates:  map[string]*ImageStateSummary{},
				PVCRefCounts: map[string]int{},
				Pods: []*PodInfo{
					{
						Pod: &v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "node_info_cache_test",
								Name:      "test-2",
								UID:       types.UID("test-2"),
							},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Resources: v1.ResourceRequirements{
											Requests: v1.ResourceList{
												v1.ResourceCPU:    resource.MustParse("200m"),
												v1.ResourceMemory: resource.MustParse("1Ki"),
											},
										},
										Ports: []v1.ContainerPort{
											{
												HostIP:   "127.0.0.1",
												HostPort: 8080,
												Protocol: "TCP",
											},
										},
									},
								},
								NodeName: nodeName,
								Overhead: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("500"),
								},
							},
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			ni := fakeNodeInfo(pods...)

			gen := ni.Generation
			err := ni.RemovePod(ctx, test.pod)
			if err != nil {
				if test.errExpected {
					expectedErrorMsg := fmt.Errorf("no corresponding pod %s in pods of node %s", test.pod.Name, ni.Node().Name)
					if expectedErrorMsg == err {
						t.Errorf("expected error: %v, got: %v", expectedErrorMsg, err)
					}
				} else {
					t.Errorf("expected no error, got: %v", err)
				}
			} else {
				if ni.Generation <= gen {
					t.Errorf("Generation is not incremented. Prev: %v, current: %v", gen, ni.Generation)
				}
			}

			test.expectedNodeInfo.Generation = ni.Generation
			if !reflect.DeepEqual(test.expectedNodeInfo, ni) {
				t.Errorf("expected: %#v, got: %#v", test.expectedNodeInfo, ni)
			}
		})
	}
}

func fakeNodeInfo(pods ...*v1.Pod) *NodeInfo {
	ni := NewNodeInfo(pods...)
	ni.SetNode(&v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	})
	return ni
}

func TestGetNamespacesFromPodAffinityTerm(t *testing.T) {
	tests := []struct {
		name string
		term *v1.PodAffinityTerm
		want sets.String
	}{
		{
			name: "podAffinityTerm_namespace_empty",
			term: &v1.PodAffinityTerm{},
			want: sets.String{metav1.NamespaceDefault: sets.Empty{}},
		},
		{
			name: "podAffinityTerm_namespace_not_empty",
			term: &v1.PodAffinityTerm{
				Namespaces: []string{metav1.NamespacePublic, metav1.NamespaceSystem},
			},
			want: sets.NewString(metav1.NamespacePublic, metav1.NamespaceSystem),
		},
		{
			name: "podAffinityTerm_namespace_selector_not_nil",
			term: &v1.PodAffinityTerm{
				NamespaceSelector: &metav1.LabelSelector{},
			},
			want: sets.String{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := getNamespacesFromPodAffinityTerm(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "topologies_pod",
					Namespace: metav1.NamespaceDefault,
				},
			}, test.term)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
