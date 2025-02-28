package scheduler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
)

func TestBinpackBmNodes(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"instance-category": string(cloudv1alpha1.InstanceCategoryBareMetalHost),
			},
		},
	}

	newNodeFromGroup := func(groupID string) *v1.Node {
		return &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "group-" + groupID,
				Labels: map[string]string{
					bmenroll.ClusterGroupID: groupID,
					bmenroll.VerifiedLabel:  "true",
				},
			},
		}
	}
	newNode := func(name string) *v1.Node {
		return &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	groupA1 := newNodeFromGroup("a")
	groupA2 := newNodeFromGroup("a")

	groupB1 := newNodeFromGroup("b")
	groupB2 := newNodeFromGroup("b")
	groupB3 := newNodeFromGroup("b")
	groupB4 := newNodeFromGroup("b")

	single1 := newNode("single-1")
	single2 := newNode("single-2")

	tests := []struct {
		name           string
		requestedSize  int
		pod            *v1.Pod
		availableNodes []*v1.Node
		expectedAny    []*v1.Node
	}{
		{
			name:          "single scheduling: prefer any single node if available",
			requestedSize: 1,
			pod:           pod,
			availableNodes: []*v1.Node{
				groupA1,
				groupB1,
				single1,
				single2,
			},
			expectedAny: []*v1.Node{single1, single2},
		},
		{
			requestedSize: 1,
			name:          "single scheduling: if no single node available, select a node from group with least available nodes",
			pod:           pod,
			availableNodes: []*v1.Node{
				groupA1, groupA2,
				groupB1, groupB2, groupB3, groupB4,
			},
			expectedAny: []*v1.Node{groupA1, groupA2},
		},
		{
			name:          "single scheduling: if no single node available, select a node from any groups with least available nodes",
			requestedSize: 1,
			pod:           pod,
			availableNodes: []*v1.Node{
				groupA1, groupA2,
				groupB1, groupB2,
			},
			expectedAny: []*v1.Node{groupA1, groupA2, groupB1, groupB2},
		},
		{
			name:           "single scheduling: no available nodes",
			requestedSize:  1,
			pod:            pod,
			availableNodes: []*v1.Node{},
			expectedAny:    []*v1.Node{},
		},
		{
			name:          "group scheduling: the requested size is too large",
			requestedSize: 8,
			pod:           pod,
			availableNodes: []*v1.Node{
				groupA1, groupA2,
				groupB1,
			},
			expectedAny: []*v1.Node{},
		},
		{
			name:          "group scheduling: select a node from group, select a node from group with least available nodes",
			requestedSize: 2,
			pod:           pod,
			availableNodes: []*v1.Node{
				groupA1, groupA2,
				groupB1, groupB2, groupB3, groupB4,
			},
			expectedAny: []*v1.Node{groupA1, groupA1},
		},
		{
			name:          "group scheduling: select a node from group, select a node from any groups with least available nodes",
			requestedSize: 2,
			pod:           pod,
			availableNodes: []*v1.Node{
				groupA1, groupA2,
				groupB1, groupB2,
			},
			expectedAny: []*v1.Node{groupA1, groupA1, groupB1, groupB2},
		},
		{
			name:           "group scheduling: no available nodes",
			requestedSize:  4,
			pod:            pod,
			availableNodes: []*v1.Node{},
			expectedAny:    []*v1.Node{},
		},
	}

	contain := func(nodes []*v1.Node, node *v1.Node) bool {
		for _, n := range nodes {
			if n.Name == node.Name {
				return true
			}
		}
		return false
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selectedNodes := binpack(context.Background(), tt.pod, tt.availableNodes, tt.requestedSize)
			if len(tt.expectedAny) == 0 {
				assert.Equal(t, len(selectedNodes), 0, "selected %v; want none", len(selectedNodes))
			} else {
				result := contain(tt.expectedAny, selectedNodes[0])
				assert.True(t, result, "selected %v; want any from %v", selectedNodes, tt.expectedAny)
			}
		})
	}
}
