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

package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/internal/cache"

	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	testSchedulerName = "test-scheduler"
	clusterID         = "cluster-1"
)

func TestUpdatePodInCache(t *testing.T) {
	ttl := 10 * time.Second
	nodeName := "node"

	tests := []struct {
		name   string
		oldObj interface{}
		newObj interface{}
	}{
		{
			name:   "pod updated with the same UID",
			oldObj: withPodName(podWithPort("oldUID", nodeName, 80), "pod"),
			newObj: withPodName(podWithPort("oldUID", nodeName, 8080), "pod"),
		},
		{
			name:   "pod updated with different UIDs",
			oldObj: withPodName(podWithPort("oldUID", nodeName, 80), "pod"),
			newObj: withPodName(podWithPort("newUID", nodeName, 8080), "pod"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sched := &Scheduler{
				Cache: cache.New(ttl, ctx.Done()),
			}
			sched.addPodToCache(ctx, clusterID, tt.oldObj)
			sched.updatePodInCache(ctx, clusterID, tt.oldObj, tt.newObj)

			if tt.oldObj.(*v1.Pod).UID != tt.newObj.(*v1.Pod).UID {
				if pod, err := sched.Cache.GetPod(tt.oldObj.(*v1.Pod)); err == nil {
					t.Errorf("Get pod UID %v from cache but it should not happen", pod.UID)
				}
			}
			pod, err := sched.Cache.GetPod(tt.newObj.(*v1.Pod))
			if err != nil {
				t.Errorf("Failed to get pod from scheduler: %v", err)
			}
			if pod.UID != tt.newObj.(*v1.Pod).UID {
				t.Errorf("Want pod UID %v, got %v", tt.newObj.(*v1.Pod).UID, pod.UID)
			}
		})
	}
}

func withPodName(pod *v1.Pod, name string) *v1.Pod {
	pod.Name = name
	return pod
}

func TestAddNodeToCache(t *testing.T) {
	ttl := 10 * time.Second
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Spec: v1.NodeSpec{
			ProviderID: "idc/node1",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sched := &Scheduler{
		Cache: cache.New(ttl, ctx.Done()),
	}
	sched.addNodeToCache(ctx, clusterID, node)
	assert.Equal(t, sched.Cache.NodeCount(), 1)
	sched.Cache.NodeCount()
}

func TestUpdateNodeInCache(t *testing.T) {
	ttl := 10 * time.Second
	oldNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			Labels: map[string]string{
				"region": "us-east-1",
				"zone":   "us-east-1a",
				"env":    "test",
			},
		},
		Spec: v1.NodeSpec{
			ProviderID: "idc/node1",
		},
	}

	newNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			Labels: map[string]string{
				"region": "us-west-1",
				"zone":   "us-west-1a",
				"env":    "test",
			},
		},
		Spec: v1.NodeSpec{
			ProviderID: "idc/nodenew",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sched := &Scheduler{
		Cache: cache.New(ttl, ctx.Done()),
	}
	sched.addNodeToCache(ctx, clusterID, oldNode)

	assert.Equal(t, sched.Cache.Dump().Nodes["cluster-1/node1"].Node().Labels["region"], "us-east-1")
	sched.updateNodeInCache(ctx, clusterID, oldNode, newNode)
	assert.Equal(t, sched.Cache.Dump().Nodes["cluster-1/node1"].Node().Labels["region"], "us-west-1")
}

func TestDeleteNodeFromCache(t *testing.T) {
	ttl := 10 * time.Second
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			Labels: map[string]string{
				"region": "us-east-1",
				"zone":   "us-east-1a",
				"env":    "test",
			},
		},
		Spec: v1.NodeSpec{
			ProviderID: "idc/node1",
		},
	}

	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node2",
			Labels: map[string]string{
				"region": "us-west-1",
				"zone":   "us-west-1a",
				"env":    "test",
			},
		},
		Spec: v1.NodeSpec{
			ProviderID: "idc/node2",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sched := &Scheduler{
		Cache: cache.New(ttl, ctx.Done()),
	}
	sched.addNodeToCache(ctx, clusterID, node1)
	sched.addNodeToCache(ctx, clusterID, node2)
	assert.Equal(t, sched.Cache.NodeCount(), 2)

	sched.deleteNodeFromCache(ctx, clusterID, node1)
	assert.Equal(t, sched.Cache.NodeCount(), 1)

	sched.deleteNodeFromCache(ctx, clusterID, node2)
	assert.Equal(t, sched.Cache.NodeCount(), 0)
}

func TestAddPodToCache(t *testing.T) {
	ttl := 10 * time.Second
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			UID:  types.UID(uuid.NewString()),
			Labels: map[string]string{
				"region": "us-east-1",
				"zone":   "us-east-1a",
				"env":    "test",
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sched := &Scheduler{
		Cache: cache.New(ttl, ctx.Done()),
	}
	sched.addPodToCache(ctx, clusterID, pod1)
	assert.Equal(t, sched.Cache.NodeCount(), 1)
	sched.Cache.NodeCount()
}

func podWithPort(id, desiredHost string, port int) *v1.Pod {
	pod := podWithID(id, desiredHost)
	pod.Spec.Containers = []v1.Container{
		{Name: "ctr", Ports: []v1.ContainerPort{{HostPort: int32(port)}}},
	}
	return pod
}

func podWithID(id, desiredHost string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
			UID:  types.UID(id),
		},
		Spec: v1.PodSpec{
			NodeName:      desiredHost,
			SchedulerName: testSchedulerName,
		},
	}
}

func TestDeletePodFromCache(t *testing.T) {
	ttl := 10 * time.Second
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			UID:  types.UID(uuid.NewString()),
			Labels: map[string]string{
				"region": "us-east-1",
				"zone":   "us-east-1a",
				"env":    "test",
			},
		},
	}

	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node2",
			UID:  types.UID(uuid.NewString()),
			Labels: map[string]string{
				"region": "us-west-1",
				"zone":   "us-west-1a",
				"env":    "test",
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sched := &Scheduler{
		Cache: cache.New(ttl, ctx.Done()),
	}
	sched.addPodToCache(ctx, clusterID, pod1)
	sched.addPodToCache(ctx, clusterID, pod2)
	podCount, err := sched.Cache.PodCount()
	if err != nil {
		t.Errorf("error encountered while reading pod count from cache: %v", err)
	}
	assert.Equal(t, podCount, 2)

	sched.deletePodFromCache(ctx, clusterID, pod1)
	podCount, err = sched.Cache.PodCount()
	if err != nil {
		t.Errorf("error encountered while reading pod count from cache: %v", err)
	}
	assert.Equal(t, podCount, 1)

	sched.deletePodFromCache(ctx, clusterID, pod2)
	podCount, err = sched.Cache.PodCount()
	if err != nil {
		t.Errorf("error encountered while reading pod count from cache: %v", err)
	}
	assert.Equal(t, podCount, 0)
}
