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

package testing

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// PodWrapper wraps a Pod inside.
type PodWrapper struct{ v1.Pod }

// MakePod creates a Pod wrapper.
func MakePod() *PodWrapper {
	return &PodWrapper{v1.Pod{}}
}

// Obj returns the inner Pod.
func (p *PodWrapper) Obj() *v1.Pod {
	return &p.Pod
}

// Container appends a container into PodSpec of the inner pod.
func (p *PodWrapper) Container(s string) *PodWrapper {
	p.Spec.Containers = append(p.Spec.Containers, v1.Container{
		Name:  fmt.Sprintf("con%d", len(p.Spec.Containers)),
		Image: s,
	})
	return p
}

// Node sets `s` as the nodeName of the inner pod.
func (p *PodWrapper) Node(s string) *PodWrapper {
	p.Spec.NodeName = s
	return p
}

// Req adds a new container to the inner pod with given resource map.
func (p *PodWrapper) Req(resMap map[v1.ResourceName]string) *PodWrapper {
	if len(resMap) == 0 {
		return p
	}

	res := v1.ResourceList{}
	for k, v := range resMap {
		res[k] = resource.MustParse(v)
	}
	p.Spec.Containers = append(p.Spec.Containers, v1.Container{
		Name:  fmt.Sprintf("con%d", len(p.Spec.Containers)),
		Image: "registry.k8s.io/pause:3.9",
		Resources: v1.ResourceRequirements{
			Requests: res,
			Limits:   res,
		},
	})
	return p
}

// NodeWrapper wraps a Node inside.
type NodeWrapper struct{ v1.Node }

// MakeNode creates a Node wrapper.
func MakeNode() *NodeWrapper {
	w := &NodeWrapper{v1.Node{}}
	return w.Capacity(nil)
}

// Obj returns the inner Node.
func (n *NodeWrapper) Obj() *v1.Node {
	return &n.Node
}

// Name sets `s` as the name of the inner pod.
func (n *NodeWrapper) Name(s string) *NodeWrapper {
	n.SetName(s)
	return n
}

// Capacity sets the capacity and the allocatable resources of the inner node.
// Each entry in `resources` corresponds to a resource name and its quantity.
// By default, the capacity and allocatable number of pods are set to 32.
func (n *NodeWrapper) Capacity(resources map[v1.ResourceName]string) *NodeWrapper {
	res := v1.ResourceList{
		v1.ResourcePods: resource.MustParse("32"),
	}
	for name, value := range resources {
		res[name] = resource.MustParse(value)
	}
	n.Status.Capacity, n.Status.Allocatable = res, res
	return n
}
