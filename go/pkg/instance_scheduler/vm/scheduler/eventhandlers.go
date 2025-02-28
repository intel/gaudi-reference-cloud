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
	"fmt"
	"strconv"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	metal3Informerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/informers/externalversions"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	v1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	bmenrollment "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
)

const (
	availableBMMemoryDefault = "10000Gi"
	allowedReservations      = "1"
)

func (sched *Scheduler) addNodeToCache(ctx context.Context, clusterId string, obj interface{}) {
	log := log.FromContext(ctx).WithName("Scheduler.addNodeToCache")
	node, ok := obj.(*v1.Node)
	if !ok {
		log.Error(nil, "Cannot convert to *v1.Node", logkeys.Object, obj)
		return
	}
	addClusterIdToNode(clusterId, node)
	sched.Cache.AddNode(ctx, node)
	log.V(3).Info("Add event for node", logkeys.Node, node)
}

func (sched *Scheduler) addBMNewNode(ctx context.Context, bmh *baremetalv1alpha1.BareMetalHost, node *v1.Node, clusterId string) error {
	log := log.FromContext(ctx).WithName("Scheduler.addBMNewNode")
	pausedNode := false

	//convert bmh into a node
	node.Status.Phase = v1.NodeRunning
	node.ObjectMeta = metav1.ObjectMeta{
		Labels: bmh.Labels,
	}
	// check annotations
	annotations := bmh.GetAnnotations()
	if annotations != nil {
		if _, ok := annotations[baremetalv1alpha1.PausedAnnotation]; ok {
			pausedNode = true
		}
	}
	// setup name
	node.Name = bmh.Namespace + "/" + bmh.Name
	addClusterIdToNode(clusterId, node)
	// Check if system has label populated; this is an indication that the system has finished enrollment
	if len(bmh.Labels) > 0 {
		//update data
		gpu, err := strconv.ParseInt(bmh.Labels[bmenrollment.GPUCountLabel], 10, 64)
		if err != nil {
			log.Error(err, "Scheduler to parse GPU value", logkeys.Node, node)
			return err
		}
		gpuQty := *resource.NewMilliQuantity(int64(gpu), resource.DecimalSI)
		cpu, err := strconv.ParseInt(bmh.Labels[bmenrollment.CPUCountLabel], 10, 64)
		if err != nil {
			log.Error(err, "Scheduler failed to convert CPU value", logkeys.Node, node)
			return err
		}
		cpuQty := *resource.NewMilliQuantity(int64(cpu)*1000, resource.DecimalSI)
		memoryQty, err := resource.ParseQuantity(availableBMMemoryDefault)
		if err != nil {
			log.Error(err, "Scheduler failed to convert memoryQty value to resource Quantity", logkeys.Node, node)
			return err
		}
		allowedPods, err := resource.ParseQuantity(allowedReservations)
		if err != nil {
			log.Error(err, "Scheduler failed to convert allowedReservations value to resource Quantity", logkeys.Node, node)
			return err
		}
		node.Status = v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    cpuQty,
				v1.ResourceMemory: memoryQty,
				v1.ResourcePods:   allowedPods,
				"gpu":             gpuQty,
			},
		}
		//check if there is a pod reserved for this resource
		resourceAssos := map[string]string{}
		if _, ok := bmh.Labels[bmenrollment.LastAssociatedInstance]; ok {
			resourceAssos[framework.ResourceIdPodLabel] = bmh.Labels[bmenrollment.LastAssociatedInstance]
		}
		if _, ok := bmh.Labels[bmenrollment.LastClusterGroup]; ok {
			resourceAssos[bmenrollment.ClusterGroup] = bmh.Labels[bmenrollment.LastClusterGroup]
		}
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: resourceAssos,
			},
			Spec: v1.PodSpec{
				NodeName:      node.Name,
				SchedulerName: DefaultSchedulerName,
			},
		}
		// if system is available
		if bmh.Status.Provisioning.State == baremetalv1alpha1.StateAvailable &&
			bmh.Spec.ConsumerRef == nil &&
			bmh.Status.ErrorCount == 0 &&
			bmh.GetDeletionTimestamp() == nil &&
			!pausedNode {
			// delete the pod
			if err := sched.Cache.RemovePod(ctx, pod); err != nil {
				log.Error(err, "Scheduler cache RemovePod failed", logkeys.Pod, pod)
			}
		} else {
			// check if pod has assigned consumer id
			if bmh.Spec.ConsumerRef != nil {
				resourceAssos[framework.ResourceIdPodLabel] = bmh.Spec.ConsumerRef.Name
				if _, ok := bmh.Labels[bmenrollment.LastClusterGroup]; ok {
					resourceAssos[bmenrollment.ClusterGroup] = bmh.Labels[bmenrollment.LastClusterGroup]
				}
				pod.ObjectMeta.Labels = resourceAssos
			}
			if err := sched.Cache.AddPod(ctx, pod); err != nil {
				log.Error(err, "Scheduler cache AddPod failed", logkeys.Pod, pod)
			}
		}
	}
	return nil
}

func (sched *Scheduler) addBMNodeToCache(ctx context.Context, clusterId string, obj interface{}) {
	log := log.FromContext(ctx).WithName("Scheduler.addBMNodeToCache")

	node := v1.Node{}
	bmh := obj.(*baremetalv1alpha1.BareMetalHost)
	if err := sched.addBMNewNode(ctx, bmh, &node, clusterId); err != nil {
		return
	}
	// Update node name and add it to nodes
	sched.Cache.AddNode(ctx, &node)
	log.V(3).Info("Add event for node", logkeys.Node, &node)
}

func (sched *Scheduler) updateNodeInCache(ctx context.Context, clusterId string, oldObj, newObj interface{}) {
	log := log.FromContext(ctx).WithName("Scheduler.updateNodeInCache")
	oldNode, ok := oldObj.(*v1.Node)
	if !ok {
		log.Error(nil, "Cannot convert oldObj to *v1.Node", logkeys.OldObject, oldObj)
		return
	}
	newNode, ok := newObj.(*v1.Node)
	if !ok {
		log.Error(nil, "Cannot convert newObj to *v1.Node", logkeys.NewObject, newObj)
		return
	}

	addClusterIdToNode(clusterId, oldNode)
	addClusterIdToNode(clusterId, newNode)
	sched.Cache.UpdateNode(ctx, oldNode, newNode)
}

func (sched *Scheduler) updateBMNodeInCache(ctx context.Context, clusterId string, oldObj, newObj interface{}) {
	sched.deleteBMNodeToCache(ctx, clusterId, oldObj)
	sched.addBMNodeToCache(ctx, clusterId, newObj)
}

func (sched *Scheduler) deleteNodeFromCache(ctx context.Context, clusterId string, obj interface{}) {
	log := log.FromContext(ctx).WithName("Scheduler.deleteNodeFromCache")

	var node *v1.Node
	switch t := obj.(type) {
	case *v1.Node:
		node = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		node, ok = t.Obj.(*v1.Node)
		if !ok {
			log.Error(nil, "Cannot convert to *v1.Node", logkeys.Object, t.Obj)
			return
		}
	default:
		log.Error(nil, "Cannot convert to *v1.Node", logkeys.Object, t)
		return
	}
	if node == nil {
		panic("node is nil.")
	}
	addClusterIdToNode(clusterId, node)
	log.V(3).Info("Delete event for node", logkeys.Node, node)
	if err := sched.Cache.RemoveNode(ctx, node); err != nil {
		log.Error(err, "Scheduler cache RemoveNode failed")
	}
}

func (sched *Scheduler) deleteBMNodeToCache(ctx context.Context, clusterId string, obj interface{}) {
	log := log.FromContext(ctx).WithName("Scheduler.deleteBMNodeToCache")
	node := v1.Node{}
	bmh := obj.(*baremetalv1alpha1.BareMetalHost)
	// convert bmh into a node
	node.Name = bmh.Namespace + "/" + bmh.Name
	addClusterIdToNode(clusterId, &node)
	// remove pod in case it exists
	//check if there is a pod reserved for this resource
	resourceAssos := map[string]string{}
	if _, ok := bmh.Labels[bmenrollment.LastAssociatedInstance]; ok {
		resourceAssos[framework.ResourceIdPodLabel] = bmh.Labels[bmenrollment.LastAssociatedInstance]
	}
	if _, ok := bmh.Labels[bmenrollment.LastClusterGroup]; ok {
		resourceAssos[bmenrollment.ClusterGroup] = bmh.Labels[bmenrollment.LastClusterGroup]
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: resourceAssos,
		},
		Spec: v1.PodSpec{
			NodeName:      node.Name,
			SchedulerName: DefaultSchedulerName,
		},
	}
	// delete pod
	log.V(3).Info("Delete event for scheduled pod", logkeys.Pod, pod)
	if err := sched.Cache.RemovePod(ctx, pod); err != nil {
		log.Error(err, "Scheduler cache RemovePod failed", logkeys.Pod, pod)
	}
	//delete node
	log.V(3).Info("Delete event for node", logkeys.Node, &node)
	if err := sched.Cache.RemoveNode(ctx, &node); err != nil {
		log.Error(err, "Scheduler cache RemoveNode failed")
	}
}

func (sched *Scheduler) addPodToCache(ctx context.Context, clusterId string, obj interface{}) {
	log := log.FromContext(ctx).WithName("Scheduler.addPodToCache")
	pod, ok := obj.(*v1.Pod)
	if !ok {
		log.Error(nil, "Cannot convert to *v1.Pod", logkeys.Object, obj)
		return
	}
	addClusterIdToPod(clusterId, pod)
	log.V(3).Info("Add event for scheduled pod", logkeys.Pod, pod)

	if err := sched.Cache.AddPod(ctx, pod); err != nil {
		log.Error(err, "Scheduler cache AddPod failed", logkeys.Pod, pod)
	}
}

func (sched *Scheduler) updatePodInCache(ctx context.Context, clusterId string, oldObj, newObj interface{}) {
	log := log.FromContext(ctx).WithName("Scheduler.updatePodInCache")
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		log.Error(nil, "Cannot convert oldObj to *v1.Pod", logkeys.OldObject, oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		log.Error(nil, "Cannot convert newObj to *v1.Pod", logkeys.NewObject, newObj)
		return
	}
	addClusterIdToPod(clusterId, oldPod)
	addClusterIdToPod(clusterId, newPod)
	log.V(3).Info("Update event for scheduled pod", logkeys.Pod, oldPod)

	if err := sched.Cache.UpdatePod(ctx, oldPod, newPod); err != nil {
		log.Error(err, "Scheduler cache UpdatePod failed", logkeys.Pod, oldPod)
	}
}

func (sched *Scheduler) deletePodFromCache(ctx context.Context, clusterId string, obj interface{}) {
	log := log.FromContext(ctx).WithName("Scheduler.deletePodFromCache")
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			log.Error(nil, "Cannot convert to *v1.Pod", logkeys.Object, t.Obj)
			return
		}
	default:
		log.Error(nil, "Cannot convert to *v1.Pod", logkeys.Pod, t)
		return
	}
	if pod == nil {
		panic("pod is nil.")
	}
	addClusterIdToPod(clusterId, pod)
	log.V(3).Info("Delete event for scheduled pod", logkeys.Pod, pod)
	if err := sched.Cache.RemovePod(ctx, pod); err != nil {
		log.Error(err, "Scheduler cache RemovePod failed", logkeys.Pod, pod)
	}
}

// assignedPod selects pods that are assigned (scheduled and running).
func assignedPod(pod *v1.Pod) bool {
	return len(pod.Spec.NodeName) != 0
}

// addAllEventHandlers is a helper function used in tests and in Scheduler
// to add event handlers for various informers.
func addAllEventHandlers(
	sched *Scheduler,
	informerFactory interface{},
	clusterId string,
) {
	ctx := context.Background()
	if clusterId == BmaasLocalCluster {
		informer := informerFactory.(metal3Informerfactory.SharedInformerFactory).Metal3().V1alpha1().BareMetalHosts().Informer()
		informer.AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) { sched.addBMNodeToCache(ctx, clusterId, obj) },
				DeleteFunc: func(obj interface{}) { sched.deleteBMNodeToCache(ctx, clusterId, obj) },
				UpdateFunc: func(oldObj interface{}, newObj interface{}) {
					sched.updateBMNodeInCache(ctx, clusterId, oldObj, newObj)
				},
			},
		)
		//factory.Start(make(<-chan struct{}))
	} else {
		// scheduled pod cache
		informerFactory.(informers.SharedInformerFactory).Core().V1().Pods().Informer().AddEventHandler(
			cache.FilteringResourceEventHandler{
				FilterFunc: func(obj interface{}) bool {
					switch t := obj.(type) {
					case *v1.Pod:
						return assignedPod(t)
					case cache.DeletedFinalStateUnknown:
						if _, ok := t.Obj.(*v1.Pod); ok {
							// The carried object may be stale, so we don't use it to check if
							// it's assigned or not. Attempting to cleanup anyways.
							return true
						}
						utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, sched))
						return false
					default:
						utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
						return false
					}
				},
				Handler: cache.ResourceEventHandlerFuncs{
					AddFunc:    func(obj interface{}) { sched.addPodToCache(ctx, clusterId, obj) },
					UpdateFunc: func(oldObj interface{}, newObj interface{}) { sched.updatePodInCache(ctx, clusterId, oldObj, newObj) },
					DeleteFunc: func(obj interface{}) { sched.deletePodFromCache(ctx, clusterId, obj) },
				},
			},
		)
		informerFactory.(informers.SharedInformerFactory).Core().V1().Nodes().Informer().AddEventHandler(
			cache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj interface{}) { sched.addNodeToCache(ctx, clusterId, obj) },
				UpdateFunc: func(oldObj interface{}, newObj interface{}) { sched.updateNodeInCache(ctx, clusterId, oldObj, newObj) },
				DeleteFunc: func(obj interface{}) { sched.deleteNodeFromCache(ctx, clusterId, obj) },
			},
		)
	}
}
