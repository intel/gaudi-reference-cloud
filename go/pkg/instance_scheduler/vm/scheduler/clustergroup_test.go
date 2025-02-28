package scheduler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
)

func TestClusterGroup(t *testing.T) {
	tests := []struct {
		description              string
		group                    *ClusterGroup
		requestSize              int
		expectedFit              bool
		expectedIsFullyAvailable bool
		expectedUnusedNodes      int
	}{
		{
			description: "fully available group",
			group: &ClusterGroup{
				id:          "group-a",
				networkMode: "VB",
				currentCap:  8,
				maxCap:      8,
				invalid:     true,
			},
			requestSize:              4,
			expectedFit:              true,
			expectedIsFullyAvailable: true,
			expectedUnusedNodes:      4,
		},
		{
			description: "partially available group",
			group: &ClusterGroup{
				id:          "group-a",
				networkMode: "VB",
				currentCap:  4,
				maxCap:      8,
				invalid:     true,
			},
			requestSize:              4,
			expectedFit:              true,
			expectedIsFullyAvailable: false,
			expectedUnusedNodes:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Test: Fit
			fit := tt.group.Fit(tt.requestSize)
			assert.Equal(t, tt.expectedFit, fit)

			// Test: UnusedNodes
			unusedNodes := tt.group.UnusedNodes(tt.requestSize)
			assert.Equal(t, tt.expectedUnusedNodes, unusedNodes)
		})
	}
}

func TestClusterGroupInfos(t *testing.T) {
	const (
		groupNameA       = "group-a-size-4"
		groupNameB       = "group-b-size-2"
		invalidGroupName = "group-invalid"
	)

	createValidGroup := func(groupId string, groupSize int) []*v1.Node {
		nodes := []*v1.Node{}
		for i := 0; i < groupSize; i++ {
			nodes = append(nodes, &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%d", groupId, i),
					Labels: map[string]string{
						bmenroll.ClusterGroupID: groupId,
					},
				},
			})
		}

		return nodes
	}

	nodes := []*v1.Node{}

	groupA := createValidGroup(groupNameA, 4)
	nodes = append(nodes, groupA...)

	groupB := createValidGroup(groupNameB, 2)
	nodes = append(nodes, groupB...)

	invalidGroups := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					bmenroll.ClusterGroupID:   invalidGroupName,
					bmenroll.NetworkModeLabel: "V",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					bmenroll.ClusterGroupID:   invalidGroupName,
					bmenroll.NetworkModeLabel: "B",
				},
			},
		},
	}
	nodes = append(nodes, invalidGroups...)

	infos := NewClusterGroupInfos(nodes)
	assert.Equal(t, 2, len(infos.groups))
	assert.Contains(t, infos.groups, groupNameA)
	assert.Contains(t, infos.groups, groupNameB)
	assert.NotContains(t, infos.groups, invalidGroupName) // exclude group with invalid network

	nodes = []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
				Labels: map[string]string{
					bmenroll.ClusterGroupID:   groupNameA,
					bmenroll.NetworkModeLabel: "V",
					bmenroll.VerifiedLabel:    "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-2",
				Labels: map[string]string{
					bmenroll.ClusterGroupID:     groupNameA,
					bmenroll.NetworkModeLabel:   "V",
					bmenroll.UnschedulableLabel: "true",
				},
			},
		},
	}

	infos = NewClusterGroupInfos(nodes)
	assert.Equal(t, 1, len(infos.groups))
	assert.Equal(t, 1, infos.groups[groupNameA].currentCap)
	assert.Equal(t, 2, infos.groups[groupNameA].maxCap)
	assert.Equal(t, 1, infos.groups[groupNameA].unverifiedNodeCount)
	assert.Equal(t, 1, infos.groups[groupNameA].cordonedNodeCount)
	assert.Equal(t, 1, infos.groups[groupNameA].unavailableNodeCount)
}

func TestClusterGroupInfosWithGroupFiltersOption(t *testing.T) {
	nodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
				Labels: map[string]string{
					bmenroll.ClusterGroupID:   "group-a",
					bmenroll.NetworkModeLabel: bmenroll.NetworkModeXBX,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-2",
				Labels: map[string]string{
					bmenroll.ClusterGroupID: "group-b",
					bmenroll.VerifiedLabel:  "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-3",
				Labels: map[string]string{
					bmenroll.ClusterGroupID: "group-b",
					bmenroll.VerifiedLabel:  "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-4",
				Labels: map[string]string{
					bmenroll.ClusterGroupID: "group-c",
					bmenroll.VerifiedLabel:  "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-5",
				Labels: map[string]string{
					bmenroll.ClusterGroupID:     "group-c",
					bmenroll.VerifiedLabel:      "true",
					bmenroll.UnschedulableLabel: "true",
				},
			},
		},
	}

	infos := NewClusterGroupInfos(nodes, WithGroupFilters(FilterGroupsWithNetworkMode(bmenroll.NetworkModeXBX)))
	assert.Equal(t, 1, len(infos.groups))
	assert.Contains(t, infos.groups, "group-a")

	infos = NewClusterGroupInfos(nodes, WithGroupFilters(FilterGroupsWithMinimumCurrentCap(2)))
	assert.Equal(t, 1, len(infos.groups))
	assert.Contains(t, infos.groups, "group-b")
	assert.NotContains(t, infos.groups, "group-a") // group has size 1
	assert.NotContains(t, infos.groups, "group-c") // group has size 1 due to unschedulable node
}

func TestClusterGroupInfosWithGroupIdentifierOption(t *testing.T) {
	const (
		sc1       = "sc-1"
		sc2       = "sc-2"
		groupA    = "group-a"
		groupB    = "group-b"
		groupC    = "group-c"
		nodeName1 = "node-1"
		nodeName2 = "node-2"
		nodeName3 = "node-3"
	)

	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName1,
			Labels: map[string]string{
				bmenroll.SuperComputeGroupID: sc1,
				bmenroll.ClusterGroupID:      groupA,
				bmenroll.NetworkModeLabel:    bmenroll.NetworkModeXBX,
			},
		},
	}
	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName2,
			Labels: map[string]string{
				bmenroll.SuperComputeGroupID: sc1,
				bmenroll.ClusterGroupID:      groupB,
				bmenroll.NetworkModeLabel:    bmenroll.NetworkModeXBX,
			},
		},
	}
	node3 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName3,
			Labels: map[string]string{
				bmenroll.SuperComputeGroupID: sc2,
				bmenroll.ClusterGroupID:      groupC,
				bmenroll.NetworkModeLabel:    bmenroll.NetworkModeXBX,
			},
		},
	}

	nodes := []*v1.Node{node1, node2, node3}

	// Test: create group infos with cluster group identifier
	infos := NewClusterGroupInfos(nodes, WithGroupIdentifier(ClusterGroupIdentifier))
	assert.Equal(t, 3, len(infos.groups))
	assert.Contains(t, infos.groups, groupA)
	assert.Contains(t, infos.groups, groupB)
	assert.Contains(t, infos.groups, groupC)

	// Test: create group infos with super compute group identifier
	infos = NewClusterGroupInfos(nodes, WithGroupIdentifier(SuperComputeGroupIdentifier))
	assert.Equal(t, 5, len(infos.groups))

	assert.Contains(t, infos.groups, sc1)
	assert.Len(t, infos.groups[sc1].subGroups, 2)
	assert.Contains(t, infos.groups[sc1].subGroups, groupA)
	assert.Equal(t, infos.groups[sc1].subGroups[groupA], infos.groups[groupA])
	assert.Equal(t, infos.groups[sc1], infos.groups[groupA].parentGroup)
	assert.Contains(t, infos.groups[sc1].subGroups, groupB)
	assert.Equal(t, infos.groups[sc1].subGroups[groupB], infos.groups[groupB])
	assert.Equal(t, infos.groups[sc1], infos.groups[groupB].parentGroup)

	assert.Contains(t, infos.groups, sc2)
	assert.Equal(t, 1, len(infos.groups[sc2].subGroups))
	assert.Contains(t, infos.groups[sc2].subGroups, groupC)
	assert.Equal(t, infos.groups[sc2].subGroups[groupC], infos.groups[groupC])
	assert.Equal(t, infos.groups[sc2], infos.groups[groupC].parentGroup)

	assert.Contains(t, infos.groups, groupA)
	assert.Nil(t, infos.groups[groupA].subGroups)
	assert.Equal(t, 1, infos.groups[groupA].maxCap)

	assert.Contains(t, infos.groups, groupB)
	assert.Nil(t, infos.groups[groupB].subGroups)
	assert.Equal(t, 1, infos.groups[groupB].maxCap)

	assert.Contains(t, infos.groups, groupC)
	assert.Nil(t, infos.groups[groupC].subGroups)
	assert.Equal(t, 1, infos.groups[groupC].maxCap)

	// Test: remove node from group from nested groups
	assert.Equal(t, 2, infos.groups[sc1].maxCap)
	assert.Equal(t, 0, infos.groups[sc1].currentCap)
	assert.Equal(t, 2, infos.groups[sc1].unavailableNodeCount)
	assert.Len(t, infos.groups[sc1].nodes, 2)
	assert.Len(t, infos.groups[sc1].subGroups, 2)
	assert.Len(t, infos.groups[sc1].subGroups[groupA].nodes, 1)
	assert.Len(t, infos.groups[sc1].subGroups[groupB].nodes, 1)
	infos.deleteNode(node1)
	infos.deleteNode(node2)
	assert.NotContains(t, infos.groups, sc1)
	assert.Contains(t, infos.groups, sc2)

	// Test: remove subgroup from nested groups
	assert.Contains(t, infos.groups[sc2].subGroups, groupC)
	assert.Len(t, infos.groups[sc2].subGroups[groupC].nodes, 1)
	infos.deleteSubGroup(infos.groups[sc2], groupC)
	assert.NotContains(t, infos.groups, sc2)
}

func TestClusterGroupInfosWithNodeSelectorOption(t *testing.T) {
	nodes := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "device-1",
				Labels: map[string]string{
					bmenroll.ClusterGroupID: "group-a",
					bmenroll.VerifiedLabel:  "true",
					"good":                  "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "device-2",
				Labels: map[string]string{
					bmenroll.ClusterGroupID: "group-a",
					bmenroll.VerifiedLabel:  "true",
					"good":                  "false",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "device-3",
				Labels: map[string]string{
					bmenroll.ClusterGroupID: "group-b",
					bmenroll.VerifiedLabel:  "true",
					"good":                  "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "device-4",
				Labels: map[string]string{
					bmenroll.ClusterGroupID: "group-b",
					bmenroll.VerifiedLabel:  "true",
					"good":                  "true",
					"unschedulable":         "true",
				},
			},
		},
	}

	nodeSelector := &v1.NodeSelector{
		NodeSelectorTerms: []v1.NodeSelectorTerm{
			{
				MatchExpressions: []v1.NodeSelectorRequirement{
					{
						Key:      "good",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"true"},
					},
					{
						Key:      "unschedulable",
						Operator: v1.NodeSelectorOpNotIn,
						Values:   []string{"true"},
					},
				},
			},
		},
	}

	infos := NewClusterGroupInfos(nodes, WithNodeSelector(nodeSelector))
	assert.Equal(t, nodeSelector, infos.nodeSelector)
	assert.Equal(t, 2, len(infos.groups))
	assert.Equal(t, 1, infos.groups["group-a"].currentCap)
	assert.Equal(t, 2, infos.groups["group-a"].maxCap)
	assert.Equal(t, 1, infos.groups["group-b"].currentCap)
	assert.Equal(t, 2, infos.groups["group-b"].maxCap)
}

func TestFindGroups(t *testing.T) {
	group1 := ClusterGroup{
		id:          "group-1",
		networkMode: bmenroll.NetworkModeXBX,
		currentCap:  4,
		maxCap:      4,
	}
	group2 := ClusterGroup{
		id:          "group-2",
		networkMode: bmenroll.NetworkModeXBX,
		currentCap:  8,
		maxCap:      8,
	}
	group3 := ClusterGroup{
		id:          "group-3",
		networkMode: bmenroll.NetworkModeXBX,
		currentCap:  16,
		maxCap:      16,
	}

	// total current cap: 28
	// total max cap: 28
	groupinfos := &ClusterGroupInfos{
		groups: map[string]*ClusterGroup{
			group1.id: &group1,
			group2.id: &group2,
			group3.id: &group3,
		},
	}

	tests := []struct {
		description           string
		groupInfos            *ClusterGroupInfos
		requestedSize         int
		useRequestedSizeRange bool
		requestedSizeStart    int
		requestedSizeEnd      int
		preferredGroupIDs     []string
		expectedErr           error
		expectedGroupIDs      []string
	}{
		{
			description:       "requested size 0",
			groupInfos:        groupinfos,
			requestedSize:     0,
			preferredGroupIDs: []string{},
			expectedErr:       nil,
			expectedGroupIDs:  []string{},
		},
		{
			description:       "requested size 0 with preferred groups",
			groupInfos:        groupinfos,
			requestedSize:     0,
			preferredGroupIDs: []string{group1.id},
			expectedErr:       nil,
			expectedGroupIDs:  []string{},
		},
		{
			description:           "requested size 1-4",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    1,
			requestedSizeEnd:      4,
			preferredGroupIDs:     []string{},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id},
		},
		{
			description:           "requested size 5-8",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    5,
			requestedSizeEnd:      8,
			preferredGroupIDs:     []string{},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group2.id},
		},
		{
			description:           "requested size 9-12",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    9,
			requestedSizeEnd:      12,
			preferredGroupIDs:     []string{},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id, group2.id},
		},
		{
			description:           "requested size 13-16",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    13,
			requestedSizeEnd:      16,
			preferredGroupIDs:     []string{},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group3.id},
		},
		{
			description:           "requested size 17-20",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    17,
			requestedSizeEnd:      20,
			preferredGroupIDs:     []string{},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id, group3.id},
		},
		{
			description:           "requested size 21-24",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    21,
			requestedSizeEnd:      24,
			preferredGroupIDs:     []string{},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group2.id, group3.id},
		},
		{
			description:           "requested size 25-28",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    25,
			requestedSizeEnd:      28,
			preferredGroupIDs:     []string{},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id, group2.id, group3.id},
		},
		{
			description:       "requested size 29",
			groupInfos:        groupinfos,
			requestedSize:     29,
			preferredGroupIDs: []string{},
			expectedErr:       assert.AnError,
			expectedGroupIDs:  []string{},
		},
		{
			description:           "one preferred group | requested size 1-4",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    1,
			requestedSizeEnd:      4,
			preferredGroupIDs:     []string{group1.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id},
		},
		{
			description:           "one preferred group | requested size 5-8",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    5,
			requestedSizeEnd:      8,
			preferredGroupIDs:     []string{group2.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group2.id},
		},
		{
			description:           "multiple preferred groups [1,2] | requested size 1-4",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    1,
			requestedSizeEnd:      4,
			preferredGroupIDs:     []string{group1.id, group2.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id},
		},
		{
			description:           "multiple preferred groups [1,2] | requested size 5-12",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    5,
			requestedSizeEnd:      12,
			preferredGroupIDs:     []string{group1.id, group2.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id, group2.id},
		},
		{
			description:           "multiple preferred groups [1,3] | requested size 1-4",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    1,
			requestedSizeEnd:      4,
			preferredGroupIDs:     []string{group1.id, group3.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id},
		},
		{
			description:           "multiple preferred groups [1,3] | requested size 5-12",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    5,
			requestedSizeEnd:      12,
			preferredGroupIDs:     []string{group1.id, group3.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id, group3.id},
		},
		{
			description:           "multiple preferred groups [2,3] | requested size 1-8",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    1,
			requestedSizeEnd:      8,
			preferredGroupIDs:     []string{group2.id, group3.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group2.id},
		},
		{
			description:           "multiple preferred groups [2,3] | requested size 9-24",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    9,
			requestedSizeEnd:      24,
			preferredGroupIDs:     []string{group2.id, group3.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group2.id, group3.id},
		},
		{
			description:           "multiple preferred groups [1,2,3] | requested size 1-4",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    1,
			requestedSizeEnd:      4,
			preferredGroupIDs:     []string{group1.id, group2.id, group3.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id},
		},
		{
			description:           "multiple preferred groups [1,2,3] | requested size 5-12",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    5,
			requestedSizeEnd:      12,
			preferredGroupIDs:     []string{group1.id, group2.id, group3.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id, group2.id},
		},
		{
			description:           "multiple preferred groups [1,2,3] | requested size 13-28",
			groupInfos:            groupinfos,
			useRequestedSizeRange: true,
			requestedSizeStart:    13,
			requestedSizeEnd:      28,
			preferredGroupIDs:     []string{group1.id, group2.id, group3.id},
			expectedErr:           nil,
			expectedGroupIDs:      []string{group1.id, group2.id, group3.id},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			testFunc := func(size int) {
				groupIDs, err := tt.groupInfos.FindGroups(size, tt.preferredGroupIDs)
				if tt.expectedErr != nil {
					assert.Error(t, err)
					assert.Nil(t, groupIDs)
				} else {
					assert.NoError(t, err)
					assert.ElementsMatch(t, tt.expectedGroupIDs, groupIDs)
				}
			}
			if tt.useRequestedSizeRange {
				for size := tt.requestedSizeStart; size <= tt.requestedSizeEnd; size++ {
					testFunc(size)
				}
			} else {
				testFunc(tt.requestedSize)
			}
		})
	}
}

func TestFindGroupWithLeastCap(t *testing.T) {
	group1 := ClusterGroup{
		id:          "group-1",
		networkMode: bmenroll.NetworkModeXBX,
		currentCap:  8,
		maxCap:      10,
	}
	group2 := ClusterGroup{
		id:          "group-2",
		networkMode: bmenroll.NetworkModeXBX,
		currentCap:  6,
		maxCap:      8,
	}
	group3 := ClusterGroup{
		id:          "group-3",
		networkMode: bmenroll.NetworkModeXBX,
		currentCap:  6,
		maxCap:      8,
	}
	group4 := ClusterGroup{
		id:          "group-4",
		networkMode: bmenroll.NetworkModeXBX,
		currentCap:  10,
		maxCap:      12,
	}

	clusterGroupInfos := &ClusterGroupInfos{
		groups: map[string]*ClusterGroup{
			group1.id: &group1,
			group2.id: &group2,
			group3.id: &group3,
			group4.id: &group4,
		},
	}

	for i := 0; i < 3; i++ {
		// the expected group can be group2 or group3
		leastCapGroup, err := clusterGroupInfos.FindGroupWithLeastCap([]string{})
		assert.NoError(t, err)
		possibleGroupIDs := []string{group2.id, group3.id}
		assert.Contains(t, possibleGroupIDs, leastCapGroup.id)

		leastCapGroup, err = clusterGroupInfos.FindGroupWithLeastCap([]string{group2.id, group3.id})
		assert.NoError(t, err)
		possibleGroupIDs = []string{group2.id, group3.id}
		assert.Contains(t, possibleGroupIDs, leastCapGroup.id)
	}

	leastCapGroup, err := clusterGroupInfos.FindGroupWithLeastCap([]string{group1.id})
	assert.NoError(t, err)
	assert.Equal(t, group1.id, leastCapGroup.id)

	leastCapGroup, err = clusterGroupInfos.FindGroupWithLeastCap([]string{group1.id, group2.id})
	assert.NoError(t, err)
	assert.Equal(t, group2.id, leastCapGroup.id)
}
