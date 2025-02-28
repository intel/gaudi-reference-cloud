// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubernetes 1.24 kube-scheduler (https://github.com/kubernetes/kubernetes/tree/73da4d3652771d6c6dfe904fe8fae594a1a72e2b/pkg/scheduler).
// To see changes made, run diff-kube-scheduler.sh.

/*
Copyright 2015 The Kubernetes Authors.

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

package cache

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/framework"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	cleanAssumedPeriod = 1 * time.Second
)

// New returns a Cache implementation.
// It automatically starts a go routine that manages expiration of assumed pods.
// "ttl" is how long the assumed pod will get expired.
// "stop" is the channel that would close the background goroutine.
func New(ttl time.Duration, stop <-chan struct{}) Cache {
	cache := newCache(ttl, cleanAssumedPeriod, stop)
	cache.run()
	return cache
}

// nodeInfoListItem holds a NodeInfo pointer and acts as an item in a doubly
// linked list. When a NodeInfo is updated, it goes to the head of the list.
// The items closer to the head are the most recently updated items.
type nodeInfoListItem struct {
	info *framework.NodeInfo
	next *nodeInfoListItem
	prev *nodeInfoListItem
}

type cacheImpl struct {
	stop   <-chan struct{}
	ttl    time.Duration
	period time.Duration

	// This mutex guards all fields within this cache struct.
	mu sync.RWMutex
	// a set of assumed pod keys.
	// The key could further be used to get an entry in podStates.
	assumedPods sets.String
	// a map from pod key to podState.
	podStates map[string]*podState
	nodes     map[string]*nodeInfoListItem
	// headNode points to the most recently updated NodeInfo in "nodes". It is the
	// head of the linked list.
	headNode *nodeInfoListItem
	nodeTree *nodeTree
	// A map from image name to its imageState.
	imageStates map[string]*imageState
}

type podState struct {
	pod *v1.Pod
	// Used by assumedPod to determinate expiration.
	deadline *time.Time
	// Used to block cache from expiring assumedPod if binding still runs
	bindingFinished bool
}

type imageState struct {
	// Size of the image
	size int64
	// A set of node names for nodes having this image present
	nodes sets.String
}

// createImageStateSummary returns a summarizing snapshot of the given image's state.
func (cache *cacheImpl) createImageStateSummary(state *imageState) *framework.ImageStateSummary {
	return &framework.ImageStateSummary{
		Size:     state.size,
		NumNodes: len(state.nodes),
	}
}

func newCache(ttl, period time.Duration, stop <-chan struct{}) *cacheImpl {
	return &cacheImpl{
		ttl:    ttl,
		period: period,
		stop:   stop,

		nodes:       make(map[string]*nodeInfoListItem),
		nodeTree:    newNodeTree(nil),
		assumedPods: make(sets.String),
		podStates:   make(map[string]*podState),
		imageStates: make(map[string]*imageState),
	}
}

// newNodeInfoListItem initializes a new nodeInfoListItem.
func newNodeInfoListItem(ni *framework.NodeInfo) *nodeInfoListItem {
	return &nodeInfoListItem{
		info: ni,
	}
}

// moveNodeInfoToHead moves a NodeInfo to the head of "cache.nodes" doubly
// linked list. The head is the most recently updated NodeInfo.
// We assume cache lock is already acquired.
func (cache *cacheImpl) moveNodeInfoToHead(ctx context.Context, name string) {
	log := log.FromContext(ctx).WithName("cacheImpl.moveNodeInfoToHead")
	ni, ok := cache.nodes[name]
	if !ok {
		log.Error(nil, "No node info with given name found in the cache", "node", name)
		return
	}
	// if the node info list item is already at the head, we are done.
	if ni == cache.headNode {
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	if cache.headNode != nil {
		cache.headNode.prev = ni
	}
	ni.next = cache.headNode
	ni.prev = nil
	cache.headNode = ni
}

// removeNodeInfoFromList removes a NodeInfo from the "cache.nodes" doubly
// linked list.
// We assume cache lock is already acquired.
func (cache *cacheImpl) removeNodeInfoFromList(ctx context.Context, name string) {
	log := log.FromContext(ctx).WithName("cacheImpl.removeNodeInfoFromList")

	ni, ok := cache.nodes[name]
	if !ok {
		log.Error(nil, "No node info with given name found in the cache", "node", name)
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	// if the removed item was at the head, we must update the head.
	if ni == cache.headNode {
		cache.headNode = ni.next
	}
	delete(cache.nodes, name)
}

// Dump produces a dump of the current scheduler cache. This is used for
// debugging purposes only and shouldn't be confused with UpdateSnapshot
// function.
// This method is expensive, and should be only used in non-critical path.
func (cache *cacheImpl) Dump() *Dump {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	nodes := make(map[string]*framework.NodeInfo, len(cache.nodes))
	for k, v := range cache.nodes {
		nodes[k] = v.info.Clone()
	}

	return &Dump{
		Nodes:       nodes,
		AssumedPods: cache.assumedPods.Union(nil),
	}
}

func (cache *cacheImpl) DumpToLog(ctx context.Context) {
	log := log.FromContext(ctx).WithName("cacheImpl.DumpToLog")

	if log.V(1).Enabled() {
		cacheDump := cache.Dump()
		log.V(1).Info("Dump", "numAssumedPods", len(cacheDump.AssumedPods))
		nodeNames := make([]string, 0, len(cacheDump.Nodes))
		for nodeName := range cacheDump.Nodes {
			nodeNames = append(nodeNames, nodeName)
		}
		sort.Strings(nodeNames)
		for _, nodeName := range nodeNames {
			nodeInfo := cacheDump.Nodes[nodeName]
			unusedMilliCPU := nodeInfo.Allocatable.MilliCPU - nodeInfo.Requested.MilliCPU
			unusedMemory := nodeInfo.Allocatable.Memory - nodeInfo.Requested.Memory
			log.V(1).Info("Dump",
				"nodeName", nodeName,
				"numPods", len(nodeInfo.Pods),
				"unusedMilliCPU", unusedMilliCPU,
				"unusedMemoryMB", unusedMemory/1000000,
				"requestedMilliCPU", nodeInfo.Requested.MilliCPU,
				"requestedMemoryMB", nodeInfo.Requested.Memory/1000000,
				"labels", nodeInfo.Node().Labels,
			)
		}
	}
}

// UpdateSnapshot takes a snapshot of cached NodeInfo map. This is called at
// beginning of every scheduling cycle.
// The snapshot only includes Nodes that are not deleted at the time this function is called.
// nodeinfo.Node() is guaranteed to be not nil for all the nodes in the snapshot.
// This function tracks generation number of NodeInfo and updates only the
// entries of an existing snapshot that have changed after the snapshot was taken.
func (cache *cacheImpl) UpdateSnapshot(ctx context.Context, nodeSnapshot *Snapshot) error {
	log := log.FromContext(ctx).WithName("cacheImpl.UpdateSnapshot")

	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Get the last generation of the snapshot.
	snapshotGeneration := nodeSnapshot.generation

	// NodeInfoList and HavePodsWithAffinityNodeInfoList must be re-created if a node was added
	// or removed from the cache.
	updateAllLists := false
	// HavePodsWithAffinityNodeInfoList must be re-created if a node changed its
	// status from having pods with affinity to NOT having pods with affinity or the other
	// way around.
	updateNodesHavePodsWithAffinity := false
	// HavePodsWithRequiredAntiAffinityNodeInfoList must be re-created if a node changed its
	// status from having pods with required anti-affinity to NOT having pods with required
	// anti-affinity or the other way around.
	updateNodesHavePodsWithRequiredAntiAffinity := false

	// Start from the head of the NodeInfo doubly linked list and update snapshot
	// of NodeInfos updated after the last snapshot.
	for node := cache.headNode; node != nil; node = node.next {
		if node.info.Generation <= snapshotGeneration {
			// all the nodes are updated before the existing snapshot. We are done.
			break
		}
		if np := node.info.Node(); np != nil {
			existing, ok := nodeSnapshot.nodeInfoMap[np.Name]
			if !ok {
				updateAllLists = true
				existing = &framework.NodeInfo{}
				nodeSnapshot.nodeInfoMap[np.Name] = existing
			}
			clone := node.info.Clone()
			// We track nodes that have pods with affinity, here we check if this node changed its
			// status from having pods with affinity to NOT having pods with affinity or the other
			// way around.
			if (len(existing.PodsWithAffinity) > 0) != (len(clone.PodsWithAffinity) > 0) {
				updateNodesHavePodsWithAffinity = true
			}
			if (len(existing.PodsWithRequiredAntiAffinity) > 0) != (len(clone.PodsWithRequiredAntiAffinity) > 0) {
				updateNodesHavePodsWithRequiredAntiAffinity = true
			}
			// We need to preserve the original pointer of the NodeInfo struct since it
			// is used in the NodeInfoList, which we may not update.
			*existing = *clone
		}
	}
	// Update the snapshot generation with the latest NodeInfo generation.
	if cache.headNode != nil {
		nodeSnapshot.generation = cache.headNode.info.Generation
	}

	// Comparing to pods in nodeTree.
	// Deleted nodes get removed from the tree, but they might remain in the nodes map
	// if they still have non-deleted Pods.
	if len(nodeSnapshot.nodeInfoMap) > cache.nodeTree.numNodes {
		cache.removeDeletedNodesFromSnapshot(nodeSnapshot)
		updateAllLists = true
	}

	if updateAllLists || updateNodesHavePodsWithAffinity || updateNodesHavePodsWithRequiredAntiAffinity {
		cache.updateNodeInfoSnapshotList(ctx, nodeSnapshot, updateAllLists)
	}

	if len(nodeSnapshot.nodeInfoList) != cache.nodeTree.numNodes {
		errMsg := fmt.Sprintf("snapshot state is not consistent, length of NodeInfoList=%v not equal to length of nodes in tree=%v "+
			", length of NodeInfoMap=%v, length of nodes in cache=%v"+
			", trying to recover",
			len(nodeSnapshot.nodeInfoList), cache.nodeTree.numNodes,
			len(nodeSnapshot.nodeInfoMap), len(cache.nodes))
		log.Error(nil, errMsg)
		// We will try to recover by re-creating the lists for the next scheduling cycle, but still return an
		// error to surface the problem, the error will likely cause a failure to the current scheduling cycle.
		cache.updateNodeInfoSnapshotList(ctx, nodeSnapshot, true)
		return fmt.Errorf(errMsg)
	}

	return nil
}

func (cache *cacheImpl) updateNodeInfoSnapshotList(ctx context.Context, snapshot *Snapshot, updateAll bool) {
	log := log.FromContext(ctx).WithName("cacheImpl.updateNodeInfoSnapshotList")

	snapshot.havePodsWithAffinityNodeInfoList = make([]*framework.NodeInfo, 0, cache.nodeTree.numNodes)
	snapshot.havePodsWithRequiredAntiAffinityNodeInfoList = make([]*framework.NodeInfo, 0, cache.nodeTree.numNodes)
	if updateAll {
		// Take a snapshot of the nodes order in the tree
		snapshot.nodeInfoList = make([]*framework.NodeInfo, 0, cache.nodeTree.numNodes)
		nodesList, err := cache.nodeTree.list()
		if err != nil {
			log.Error(err, "Error occurred while retrieving the list of names of the nodes from node tree")
		}
		for _, nodeName := range nodesList {
			if nodeInfo := snapshot.nodeInfoMap[nodeName]; nodeInfo != nil {
				snapshot.nodeInfoList = append(snapshot.nodeInfoList, nodeInfo)
				if len(nodeInfo.PodsWithAffinity) > 0 {
					snapshot.havePodsWithAffinityNodeInfoList = append(snapshot.havePodsWithAffinityNodeInfoList, nodeInfo)
				}
				if len(nodeInfo.PodsWithRequiredAntiAffinity) > 0 {
					snapshot.havePodsWithRequiredAntiAffinityNodeInfoList = append(snapshot.havePodsWithRequiredAntiAffinityNodeInfoList, nodeInfo)
				}
			} else {
				log.Error(nil, "Node exists in nodeTree but not in NodeInfoMap, this should not happen", "node", nodeName)
			}
		}
	} else {
		for _, nodeInfo := range snapshot.nodeInfoList {
			if len(nodeInfo.PodsWithAffinity) > 0 {
				snapshot.havePodsWithAffinityNodeInfoList = append(snapshot.havePodsWithAffinityNodeInfoList, nodeInfo)
			}
			if len(nodeInfo.PodsWithRequiredAntiAffinity) > 0 {
				snapshot.havePodsWithRequiredAntiAffinityNodeInfoList = append(snapshot.havePodsWithRequiredAntiAffinityNodeInfoList, nodeInfo)
			}
		}
	}
}

// If certain nodes were deleted after the last snapshot was taken, we should remove them from the snapshot.
func (cache *cacheImpl) removeDeletedNodesFromSnapshot(snapshot *Snapshot) {
	toDelete := len(snapshot.nodeInfoMap) - cache.nodeTree.numNodes
	for name := range snapshot.nodeInfoMap {
		if toDelete <= 0 {
			break
		}
		if n, ok := cache.nodes[name]; !ok || n.info.Node() == nil {
			delete(snapshot.nodeInfoMap, name)
			toDelete--
		}
	}
}

// NodeCount returns the number of nodes in the cache.
// DO NOT use outside of tests.
func (cache *cacheImpl) NodeCount() int {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return len(cache.nodes)
}

// PodCount returns the number of pods in the cache (including those from deleted nodes).
// DO NOT use outside of tests.
func (cache *cacheImpl) PodCount() (int, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	// podFilter is expected to return true for most or all of the pods. We
	// can avoid expensive array growth without wasting too much memory by
	// pre-allocating capacity.
	count := 0
	for _, n := range cache.nodes {
		count += len(n.info.Pods)
	}
	return count, nil
}

func (cache *cacheImpl) AssumePod(ctx context.Context, pod *v1.Pod) error {
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if _, ok := cache.podStates[key]; ok {
		return fmt.Errorf("pod %v is in the cache, so can't be assumed", key)
	}

	return cache.addPod(ctx, pod, true)
}

func (cache *cacheImpl) FinishBinding(ctx context.Context, pod *v1.Pod) error {
	return cache.finishBinding(ctx, pod, time.Now())
}

// finishBinding exists to make tests determinitistic by injecting now as an argument
func (cache *cacheImpl) finishBinding(ctx context.Context, pod *v1.Pod, now time.Time) error {
	log := log.FromContext(ctx).WithName("cacheImpl.finishBinding")
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	log.V(3).Info("Finished binding for pod, can be expired", "pod", pod)
	currState, ok := cache.podStates[key]
	if ok && cache.assumedPods.Has(key) {
		dl := now.Add(cache.ttl)
		currState.bindingFinished = true
		currState.deadline = &dl
	}
	return nil
}

func (cache *cacheImpl) ForgetPod(ctx context.Context, pod *v1.Pod) error {
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.podStates[key]
	if ok && currState.pod.Spec.NodeName != pod.Spec.NodeName {
		return fmt.Errorf("pod %v was assumed on %v but assigned to %v", key, pod.Spec.NodeName, currState.pod.Spec.NodeName)
	}

	// Only assumed pod can be forgotten.
	if ok && cache.assumedPods.Has(key) {
		return cache.removePod(ctx, pod)
	}
	return fmt.Errorf("pod %v wasn't assumed so cannot be forgotten", key)
}

// Assumes that lock is already acquired.
func (cache *cacheImpl) addPod(ctx context.Context, pod *v1.Pod, assumePod bool) error {
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}
	n, ok := cache.nodes[pod.Spec.NodeName]
	if !ok {
		n = newNodeInfoListItem(framework.NewNodeInfo())
		cache.nodes[pod.Spec.NodeName] = n
	}
	n.info.AddPod(pod)
	cache.moveNodeInfoToHead(ctx, pod.Spec.NodeName)
	ps := &podState{
		pod: pod,
	}
	cache.podStates[key] = ps
	if assumePod {
		cache.assumedPods.Insert(key)
	}
	return nil
}

// Assumes that lock is already acquired.
func (cache *cacheImpl) updatePod(ctx context.Context, oldPod, newPod *v1.Pod) error {
	if err := cache.removePod(ctx, oldPod); err != nil {
		return err
	}
	return cache.addPod(ctx, newPod, false)
}

// Assumes that lock is already acquired.
// Removes a pod from the cached node info. If the node information was already
// removed and there are no more pods left in the node, cleans up the node from
// the cache.
func (cache *cacheImpl) removePod(ctx context.Context, pod *v1.Pod) error {
	log := log.FromContext(ctx).WithName("cacheImpl.removePod")

	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}

	n, ok := cache.nodes[pod.Spec.NodeName]
	if !ok {
		log.Error(nil, "Node not found when trying to remove pod", "node", pod.Spec.NodeName, "pod", pod)
	} else {
		if err := n.info.RemovePod(ctx, pod); err != nil {
			return err
		}
		if len(n.info.Pods) == 0 && n.info.Node() == nil {
			cache.removeNodeInfoFromList(ctx, pod.Spec.NodeName)
		} else {
			cache.moveNodeInfoToHead(ctx, pod.Spec.NodeName)
		}
	}

	delete(cache.podStates, key)
	delete(cache.assumedPods, key)
	return nil
}

func (cache *cacheImpl) AddPod(ctx context.Context, pod *v1.Pod) error {
	log := log.FromContext(ctx).WithName("cacheImpl.AddPod")
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.podStates[key]
	switch {
	case ok && cache.assumedPods.Has(key):
		log.V(3).Info("AddPod: Pod is in assumed pods", "pod", pod)
		if currState.pod.Spec.NodeName != pod.Spec.NodeName {
			// The pod was added to a different node than it was assumed to.
			log.Info("Pod was added to a different node than it was assumed", "pod", pod, "assumedNode", pod.Spec.NodeName, "currentNode", currState.pod.Spec.NodeName)
			if err = cache.updatePod(ctx, currState.pod, pod); err != nil {
				log.Error(err, "Error occurred while updating pod")
			}
		} else {
			delete(cache.assumedPods, key)
			cache.podStates[key].deadline = nil
			cache.podStates[key].pod = pod
		}
	case !ok:
		// Pod was expired. We should add it back.
		if err = cache.addPod(ctx, pod, false); err != nil {
			log.Error(err, "Error occurred while adding pod")
		}
	default:
		return fmt.Errorf("pod %v was already in added state", key)
	}
	return nil
}

func (cache *cacheImpl) UpdatePod(ctx context.Context, oldPod, newPod *v1.Pod) error {
	log := log.FromContext(ctx).WithName("cacheImpl.UpdatePod")
	key, err := framework.GetPodKey(oldPod)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.podStates[key]
	// An assumed pod won't have Update/Remove event. It needs to have Add event
	// before Update event, in which case the state would change from Assumed to Added.
	if ok && !cache.assumedPods.Has(key) {
		if currState.pod.Spec.NodeName != newPod.Spec.NodeName {
			log.Error(nil, "Pod updated on a different node than previously added to", "pod", oldPod,
				"currState.pod.Spec.NodeName", currState.pod.Spec.NodeName, "newPod.Spec.NodeName", newPod.Spec.NodeName)
			log.Error(nil, "scheduler cache is corrupted and can badly affect scheduling decisions")
			panic("scheduler cache is corrupted and can badly affect scheduling decisions")
		}
		return cache.updatePod(ctx, oldPod, newPod)
	}
	return fmt.Errorf("pod %v is not added to scheduler cache, so cannot be updated", key)
}

func (cache *cacheImpl) RemovePod(ctx context.Context, pod *v1.Pod) error {
	log := log.FromContext(ctx).WithName("cacheImpl.RemovePod")

	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.podStates[key]
	if !ok {
		return fmt.Errorf("pod %v is not found in scheduler cache, so cannot be removed from it", key)
	}
	if currState.pod.Spec.NodeName != pod.Spec.NodeName {
		log.Error(nil, "Pod was added to a different node than it was assumed", "pod", pod, "assumedNode", pod.Spec.NodeName, "currentNode", currState.pod.Spec.NodeName)
		if pod.Spec.NodeName != "" {
			// An empty NodeName is possible when the scheduler misses a Delete
			// event and it gets the last known state from the informer cache.
			log.Error(nil, "scheduler cache is corrupted and can badly affect scheduling decisions")
			os.Exit(1)
		}
	}
	return cache.removePod(ctx, currState.pod)
}

func (cache *cacheImpl) IsAssumedPod(pod *v1.Pod) (bool, error) {
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return false, err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	return cache.assumedPods.Has(key), nil
}

// GetPod might return a pod for which its node has already been deleted from
// the main cache. This is useful to properly process pod update events.
func (cache *cacheImpl) GetPod(pod *v1.Pod) (*v1.Pod, error) {
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return nil, err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	podState, ok := cache.podStates[key]
	if !ok {
		return nil, fmt.Errorf("pod %v does not exist in scheduler cache", key)
	}

	return podState.pod, nil
}

func (cache *cacheImpl) AddNode(ctx context.Context, node *v1.Node) *framework.NodeInfo {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.nodes[node.Name]
	if !ok {
		n = newNodeInfoListItem(framework.NewNodeInfo())
		cache.nodes[node.Name] = n
	} else {
		cache.removeNodeImageStates(n.info.Node())
	}
	cache.moveNodeInfoToHead(ctx, node.Name)

	cache.nodeTree.addNode(ctx, node)
	cache.addNodeImageStates(node, n.info)
	n.info.SetNode(node)
	return n.info.Clone()
}

func (cache *cacheImpl) UpdateNode(ctx context.Context, oldNode, newNode *v1.Node) *framework.NodeInfo {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.nodes[newNode.Name]
	if !ok {
		n = newNodeInfoListItem(framework.NewNodeInfo())
		cache.nodes[newNode.Name] = n
		cache.nodeTree.addNode(ctx, newNode)
	} else {
		cache.removeNodeImageStates(n.info.Node())
	}
	cache.moveNodeInfoToHead(ctx, newNode.Name)

	cache.nodeTree.updateNode(ctx, oldNode, newNode)
	cache.addNodeImageStates(newNode, n.info)
	n.info.SetNode(newNode)
	return n.info.Clone()
}

// RemoveNode removes a node from the cache's tree.
// The node might still have pods because their deletion events didn't arrive
// yet. Those pods are considered removed from the cache, being the node tree
// the source of truth.
// However, we keep a ghost node with the list of pods until all pod deletion
// events have arrived. A ghost node is skipped from snapshots.
func (cache *cacheImpl) RemoveNode(ctx context.Context, node *v1.Node) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.nodes[node.Name]
	if !ok {
		return fmt.Errorf("node %v is not found", node.Name)
	}
	n.info.RemoveNode()
	// We remove NodeInfo for this node only if there aren't any pods on this node.
	// We can't do it unconditionally, because notifications about pods are delivered
	// in a different watch, and thus can potentially be observed later, even though
	// they happened before node removal.
	if len(n.info.Pods) == 0 {
		cache.removeNodeInfoFromList(ctx, node.Name)
	} else {
		cache.moveNodeInfoToHead(ctx, node.Name)
	}
	if err := cache.nodeTree.removeNode(ctx, node); err != nil {
		return err
	}
	cache.removeNodeImageStates(node)
	return nil
}

// addNodeImageStates adds states of the images on given node to the given nodeInfo and update the imageStates in
// scheduler cache. This function assumes the lock to scheduler cache has been acquired.
func (cache *cacheImpl) addNodeImageStates(node *v1.Node, nodeInfo *framework.NodeInfo) {
	newSum := make(map[string]*framework.ImageStateSummary)

	for _, image := range node.Status.Images {
		for _, name := range image.Names {
			// update the entry in imageStates
			state, ok := cache.imageStates[name]
			if !ok {
				state = &imageState{
					size:  image.SizeBytes,
					nodes: sets.NewString(node.Name),
				}
				cache.imageStates[name] = state
			} else {
				state.nodes.Insert(node.Name)
			}
			// create the imageStateSummary for this image
			if _, ok := newSum[name]; !ok {
				newSum[name] = cache.createImageStateSummary(state)
			}
		}
	}
	nodeInfo.ImageStates = newSum
}

// removeNodeImageStates removes the given node record from image entries having the node
// in imageStates cache. After the removal, if any image becomes free, i.e., the image
// is no longer available on any node, the image entry will be removed from imageStates.
func (cache *cacheImpl) removeNodeImageStates(node *v1.Node) {
	if node == nil {
		return
	}

	for _, image := range node.Status.Images {
		for _, name := range image.Names {
			state, ok := cache.imageStates[name]
			if ok {
				state.nodes.Delete(node.Name)
				if len(state.nodes) == 0 {
					// Remove the unused image to make sure the length of
					// imageStates represents the total number of different
					// images on all nodes
					delete(cache.imageStates, name)
				}
			}
		}
	}
}

func (cache *cacheImpl) run() {
	go wait.Until(cache.cleanupExpiredAssumedPods, cache.period, cache.stop)
	go wait.Until(func() { cache.DumpToLog(context.Background()) }, 10*time.Second, cache.stop)
}

func (cache *cacheImpl) cleanupExpiredAssumedPods() {
	cache.cleanupAssumedPods(context.Background(), time.Now())
}

// cleanupAssumedPods exists for making test deterministic by taking time as input argument.
// It also reports metrics on the cache size for nodes, pods, and assumed pods.
func (cache *cacheImpl) cleanupAssumedPods(ctx context.Context, now time.Time) {
	log := log.FromContext(ctx).WithName("cacheImpl.cleanupAssumedPods")
	cache.mu.Lock()
	defer cache.mu.Unlock()
	defer cache.updateMetrics()

	// The size of assumedPods should be small
	for key := range cache.assumedPods {
		ps, ok := cache.podStates[key]
		if !ok {
			log.Error(nil, "Key found in assumed set but not in podStates, potentially a logical error")
			os.Exit(1)
		}
		if !ps.bindingFinished {
			log.V(3).Info("Could not expire cache for pod as binding is still in progress", "pod", ps.pod)
			continue
		}
		if now.After(*ps.deadline) {
			log.Info("Pod expired", "pod", ps.pod)
			if err := cache.removePod(ctx, ps.pod); err != nil {
				log.Error(err, "ExpirePod failed", "pod", ps.pod)
			}
		}
	}
}

// updateMetrics updates cache size metric values for pods, assumed pods, and nodes
func (cache *cacheImpl) updateMetrics() {
	metrics.CacheSize.WithLabelValues("assumed_pods").Set(float64(len(cache.assumedPods)))
	metrics.CacheSize.WithLabelValues("pods").Set(float64(len(cache.podStates)))
	metrics.CacheSize.WithLabelValues("nodes").Set(float64(len(cache.nodes)))
}
