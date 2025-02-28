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

package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	v1 "k8s.io/api/core/v1"
	utilnode "k8s.io/component-helpers/node/topology"
)

// nodeTree is a tree-like data structure that holds node names in each zone. Zone names are
// keys to "NodeTree.tree" and values of "NodeTree.tree" are arrays of node names.
// NodeTree is NOT thread-safe, any concurrent updates/reads from it must be synchronized by the caller.
// It is used only by schedulerCache, and should stay as such.
type nodeTree struct {
	tree     map[string][]string // a map from zone (region-zone) to an array of nodes in the zone.
	zones    []string            // a list of all the zones in the tree (keys)
	numNodes int
}

// newNodeTree creates a NodeTree from nodes.
func newNodeTree(nodes []*v1.Node) *nodeTree {
	nt := &nodeTree{
		tree: make(map[string][]string),
	}
	for _, n := range nodes {
		nt.addNode(context.Background(), n)
	}
	return nt
}

// addNode adds a node and its corresponding zone to the tree. If the zone already exists, the node
// is added to the array of nodes in that zone.
func (nt *nodeTree) addNode(ctx context.Context, n *v1.Node) {
	log := log.FromContext(ctx).WithName("nodeTree.addNode")

	zone := utilnode.GetZoneKey(n)
	if na, ok := nt.tree[zone]; ok {
		for _, nodeName := range na {
			if nodeName == n.Name {
				log.Info("Node already exists in the NodeTree", logkeys.Node, n)
				return
			}
		}
		nt.tree[zone] = append(na, n.Name)
	} else {
		nt.zones = append(nt.zones, zone)
		nt.tree[zone] = []string{n.Name}
	}
	log.V(2).Info("Added node in listed group to NodeTree", logkeys.Node, n, logkeys.AvailabilityZone, zone)
	nt.numNodes++
}

// removeNode removes a node from the NodeTree.
func (nt *nodeTree) removeNode(ctx context.Context, n *v1.Node) error {
	log := log.FromContext(ctx).WithName("nodeTree.removeNode")

	zone := utilnode.GetZoneKey(n)
	if na, ok := nt.tree[zone]; ok {
		for i, nodeName := range na {
			if nodeName == n.Name {
				nt.tree[zone] = append(na[:i], na[i+1:]...)
				if len(nt.tree[zone]) == 0 {
					nt.removeZone(zone)
				}
				log.V(2).Info("Removed node in listed group from NodeTree", logkeys.Node, n, logkeys.AvailabilityZone, zone)
				nt.numNodes--
				return nil
			}
		}
	}
	log.Error(nil, "Node in listed group was not found", logkeys.Node, n, logkeys.AvailabilityZone, zone)
	return fmt.Errorf("node %q in group %q was not found", n.Name, zone)
}

// removeZone removes a zone from tree.
// This function must be called while writer locks are hold.
func (nt *nodeTree) removeZone(zone string) {
	delete(nt.tree, zone)
	for i, z := range nt.zones {
		if z == zone {
			nt.zones = append(nt.zones[:i], nt.zones[i+1:]...)
			return
		}
	}
}

// updateNode updates a node in the NodeTree.
func (nt *nodeTree) updateNode(ctx context.Context, old, new *v1.Node) {
	log := log.FromContext(ctx).WithName("nodeTree.updateNode")
	var oldZone string
	if old != nil {
		oldZone = utilnode.GetZoneKey(old)
	}
	newZone := utilnode.GetZoneKey(new)
	// If the zone ID of the node has not changed, we don't need to do anything. Name of the node
	// cannot be changed in an update.
	if oldZone == newZone {
		return
	}
	if old == nil {
		panic("old is nil.")
	}
	// No error checking. We ignore whether the old node exists or not.
	if err := nt.removeNode(ctx, old); err != nil {
		log.Error(err, logkeys.Error)
	}
	nt.addNode(ctx, new)
}

// list returns the list of names of the node. NodeTree iterates over zones and in each zone iterates
// over nodes in a round robin fashion.
func (nt *nodeTree) list() ([]string, error) {
	if len(nt.zones) == 0 {
		return nil, nil
	}
	nodesList := make([]string, 0, nt.numNodes)
	numExhaustedZones := 0
	nodeIndex := 0
	for len(nodesList) < nt.numNodes {
		if numExhaustedZones >= len(nt.zones) { // all zones are exhausted.
			return nodesList, errors.New("all zones exhausted before reaching count of nodes expected")
		}
		for zoneIndex := 0; zoneIndex < len(nt.zones); zoneIndex++ {
			na := nt.tree[nt.zones[zoneIndex]]
			if nodeIndex >= len(na) { // If the zone is exhausted, continue
				if nodeIndex == len(na) { // If it is the first time the zone is exhausted
					numExhaustedZones++
				}
				continue
			}
			nodesList = append(nodesList, na[nodeIndex])
		}
		nodeIndex++
	}
	return nodesList, nil
}
