package scheduler

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"

	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
)

const (
	// node label used for marking nodes that are assigned to a tenant
	nodeAssignedLabel = "scheduling.cloud.intel.com/node-assigned"

	ClusterGroupType      = "cluster"
	SuperComputeGroupType = "supercompute"
)

// ClusterGroup is a group of nodes in a network
type ClusterGroup struct {
	// id is the unique identifier of the cluster group.
	id string
	// type of the group
	groupType string
	// networkMode is the network configuration of the cluster group.
	networkMode string
	// currentCap is the current capacity of the cluster group.
	currentCap int
	// maxCap is the maximum capacity that the cluster group can reach.
	maxCap int
	// assigned indicates whether the cluster group is assigned to a tenant.
	assigned bool
	// assignedNodeCount is the number of nodes assigned to a tenant.
	assignedNodeCount int
	// unavailableNodeCount is the number of nodes that are unavailable in the cluster group.
	unavailableNodeCount int
	// cordonedNodeCount is the number of nodes marked as unschedulable in the cluster group.
	cordonedNodeCount int
	// unverifiedNodeCount is the number of nodes that are unverified by the node validator.
	unverifiedNodeCount int
	// filteredNodeCount is the number of nodes that are filtered out by the node selector.
	filteredNodeCount int
	// invalid indicates whether the cluster group is invalid and should not be considered.
	invalid bool
	// nodes contains the aggregated nodes in the cluster group.
	nodes map[string]*v1.Node
	// parentGroup is the larger group that this group is part of.
	parentGroup *ClusterGroup
	// subGroups is the smaller groups that are part of this group. Support only one nested level.
	subGroups map[string]*ClusterGroup
	// mutex for the group
	mu sync.RWMutex
}

func (g *ClusterGroup) Fit(requestedSize int) bool {
	return requestedSize <= g.currentCap
}

func (g *ClusterGroup) UnusedNodes(requestedSize int) int {
	return g.currentCap - requestedSize
}

func (g *ClusterGroup) SubtractGroup(group *ClusterGroup) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.maxCap -= group.maxCap
	g.currentCap -= group.currentCap
	g.unavailableNodeCount -= group.unavailableNodeCount
	g.assignedNodeCount -= group.assignedNodeCount
	g.cordonedNodeCount -= group.cordonedNodeCount
	g.unverifiedNodeCount -= group.unverifiedNodeCount
	g.filteredNodeCount -= group.filteredNodeCount

	for _, node := range group.nodes {
		delete(g.nodes, node.Name)
	}
}

func (g *ClusterGroup) GetLogKeyValues() []any {
	g.mu.RLock()
	defer g.mu.RUnlock()

	parentGroupID := ""
	if g.parentGroup != nil {
		parentGroupID = g.parentGroup.id
	}

	subGroupIDs := []string{}
	for _, subGroup := range g.subGroups {
		subGroupIDs = append(subGroupIDs, subGroup.id)
	}

	return []any{
		"id", g.id,
		"type", g.groupType,
		"capacity", fmt.Sprintf("%d/%d", g.currentCap, g.maxCap),
		"unavailableNodeCount", g.unavailableNodeCount,
		"assigned", g.assigned,
		"assignedNodeCount", g.assignedNodeCount,
		"filteredNodeCount", g.filteredNodeCount,
		"unverifiedNodeCount", g.unverifiedNodeCount,
		"cordonedNodeCount", g.cordonedNodeCount,
		"networkMode", g.networkMode,
		"parentGroupID", parentGroupID,
		"subGroupCount", len(g.subGroups),
		"subGroupIDs", subGroupIDs,
	}
}

// ClusterGroupInfos provides an aggregated information about available cluster groups
type ClusterGroupInfos struct {
	mu              sync.RWMutex
	groups          map[string]*ClusterGroup
	groupIdentifier GroupIdentifier
	groupFilters    []ClusterGroupInfosFilter
	nodeSelector    *v1.NodeSelector
}

func NewClusterGroupInfos(nodes []*v1.Node, opts ...ClusterGroupInfosOption) *ClusterGroupInfos {
	groups := make(map[string]*ClusterGroup)
	infos := &ClusterGroupInfos{groups: groups}
	for _, opt := range opts {
		opt(infos)
	}
	for _, node := range nodes {
		infos.addNode(node)
	}
	for _, group := range groups {
		if group.assignedNodeCount > 0 {
			group.assigned = true
		}
		infos.filterGroup(group)
	}
	return infos
}

func (i *ClusterGroupInfos) addNode(node *v1.Node) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.groupIdentifier == nil {
		i.groupIdentifier = ClusterGroupIdentifier
	}

	groups, err := i.groupIdentifier(node)
	if err != nil {
		return
	}

	for _, group := range groups {
		// add group if not exists
		if _, exist := i.groups[group.id]; !exist {
			i.groups[group.id] = group
		}
		existingGroup := i.groups[group.id]

		// add subgroups if not exists
		for _, sub := range group.subGroups {
			if _, exist := existingGroup.subGroups[sub.id]; !exist {
				existingGroup.subGroups[sub.id] = sub
			}
			// update parent group
			sub.parentGroup = existingGroup
		}

		i.addNodeToGroup(existingGroup, node)
	}
}

func (i *ClusterGroupInfos) deleteNode(node *v1.Node) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, group := range i.groups {
		// delete bottom-up
		for _, subGroup := range group.subGroups {
			i.deleteNodeFromGroup(subGroup, node)
			// cleanup group if no nodes left
			if len(subGroup.nodes) == 0 {
				delete(group.subGroups, subGroup.id)
			}
		}
		i.deleteNodeFromGroup(group, node)
		if len(group.nodes) == 0 {
			delete(i.groups, group.id)
		}
	}
}

func (i *ClusterGroupInfos) addNodeToGroup(group *ClusterGroup, node *v1.Node) {
	if group == nil {
		return
	}

	group.mu.Lock()
	defer group.mu.Unlock()

	// update group stats
	group.maxCap++
	if unavailable := i.updateGroupUnavailableNodeCount(group, node, true); !unavailable {
		group.currentCap++
	}
	i.updateGroupNetworkMode(group, node)

	if group.nodes == nil {
		group.nodes = make(map[string]*v1.Node)
	}
	group.nodes[node.Name] = node
}

func (i *ClusterGroupInfos) deleteNodeFromGroup(group *ClusterGroup, node *v1.Node) {
	if group == nil {
		return
	}

	group.mu.Lock()
	defer group.mu.Unlock()

	if _, exist := group.nodes[node.Name]; !exist {
		return
	}

	// update group stats
	group.maxCap--
	if unavailable := i.updateGroupUnavailableNodeCount(group, node, false); !unavailable {
		group.currentCap--
	}

	delete(group.nodes, node.Name)
}

func (i *ClusterGroupInfos) deleteSubGroup(group *ClusterGroup, subGroupID string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	subGroup, exist := group.subGroups[subGroupID]
	if !exist {
		return
	}

	group.SubtractGroup(subGroup)
	delete(group.subGroups, subGroup.id)

	// clean up group if no nodes left
	if group.maxCap == 0 {
		delete(i.groups, group.id)
	}
}

func (i *ClusterGroupInfos) updateGroupNetworkMode(group *ClusterGroup, node *v1.Node) {
	if mode, exists := node.Labels[bmenroll.NetworkModeLabel]; exists {
		if group.networkMode != "" && group.networkMode != mode {
			// invalidate group with inconsistent network mode information
			group.invalid = true
			return
		}
		group.networkMode = mode
	}
}

func (i *ClusterGroupInfos) updateGroupUnavailableNodeCount(group *ClusterGroup, node *v1.Node, add bool) bool {
	delta := 1
	if !add {
		delta = -1
	}

	unavailable := false
	if value, exists := node.Labels[bmenroll.UnschedulableLabel]; exists && value == "true" {
		unavailable = true
		group.cordonedNodeCount += delta
	}
	if value, exists := node.Labels[bmenroll.VerifiedLabel]; !exists || value != "true" {
		unavailable = true
		group.unverifiedNodeCount += delta
	}
	if value, exists := node.Labels[nodeAssignedLabel]; exists && value == "true" {
		unavailable = true
		group.assignedNodeCount += delta
	}
	if i.nodeSelector != nil {
		passed, err := i.filterNode(node)
		if err != nil || !passed {
			unavailable = true
			group.filteredNodeCount += delta
		}
	}
	if unavailable {
		group.unavailableNodeCount += delta
	}

	return unavailable
}

func (i *ClusterGroupInfos) filterNode(node *v1.Node) (bool, error) {
	selector := nodeaffinity.NewLazyErrorNodeSelector(i.nodeSelector)
	return selector.Match(node)
}

func (i *ClusterGroupInfos) filterGroup(group *ClusterGroup) {
	if group.invalid {
		delete(i.groups, group.id)
		return
	}
	for _, passed := range i.groupFilters {
		if !passed(group) {
			delete(i.groups, group.id)
			return
		}
	}
}

// FindGroups returns the group IDs of the groups that fit the requested size.
func (i *ClusterGroupInfos) FindGroups(requestedSize int, preferredGroupIds []string) ([]string, error) {
	if requestedSize == 0 {
		return []string{}, nil
	}

	i.mu.RLock()
	defer i.mu.RUnlock()

	groups := i.sortGroupsByCapDecreasing()

	remaining := requestedSize
	prioritizedGroups := sets.NewString(preferredGroupIds...)
	var selectedGroups, unfitGroups []*ClusterGroup

	// prioritize preferredGroupIds
	for _, preferredId := range preferredGroupIds {
		for _, group := range groups {
			if group.id != preferredId {
				continue
			}
			prioritizedGroups.Insert(group.id)
			selectedGroups = append(selectedGroups, group)
			remaining -= group.currentCap
			if remaining < 0 {
				return getGroupIDs(selectedGroups), nil
			}
			break
		}
		if remaining == 0 {
			return getGroupIDs(selectedGroups), nil
		}
	}

	// find new groups to fullfil the request
	for _, group := range groups {
		if prioritizedGroups.Has(group.id) {
			continue
		}
		if group.currentCap <= remaining {
			selectedGroups = append(selectedGroups, group)
			remaining -= group.currentCap
		} else {
			unfitGroups = append(unfitGroups, group)
		}
		if remaining == 0 {
			return getGroupIDs(selectedGroups), nil
		}
	}

	if remaining > 0 {
		// reconsider the last unfit group to reduce unused nodes
		if len(unfitGroups) > 0 {
			lastUnfitGroup := unfitGroups[len(unfitGroups)-1]

			// check if the group should fit all nodes
			if remaining <= lastUnfitGroup.currentCap {
				unused := lastUnfitGroup.UnusedNodes(remaining)
				newUnused := lastUnfitGroup.UnusedNodes(requestedSize)
				if newUnused >= 0 && newUnused < unused {
					return []string{lastUnfitGroup.id}, nil
				}

				// check if th group should fit the remaining nodes
				if len(selectedGroups) > 0 {
					lastSelectedGroup := selectedGroups[len(selectedGroups)-1]
					unused = lastUnfitGroup.UnusedNodes(remaining)
					newUnused = lastUnfitGroup.UnusedNodes(remaining + lastSelectedGroup.currentCap)
					if newUnused >= 0 && newUnused < unused {
						// remove the last selected group
						selectedGroups = selectedGroups[:len(selectedGroups)-1]
						remaining += lastSelectedGroup.currentCap
					}
				}

				// fit the remaining in this group
				selectedGroups = append(selectedGroups, lastUnfitGroup)
				remaining -= lastUnfitGroup.currentCap
				if remaining <= 0 {
					return getGroupIDs(selectedGroups), nil
				}
			}
		}
		return nil, fmt.Errorf("the requested size %d exceeds or does not fit the available group capacity", requestedSize)
	}

	return getGroupIDs(selectedGroups), nil
}

// find the group with the least capacity.
func (i *ClusterGroupInfos) FindGroupWithLeastCap(preferredGroupIds []string) (*ClusterGroup, error) {
	if len(i.groups) == 0 {
		return nil, fmt.Errorf("no groups found")
	}

	i.mu.RLock()
	defer i.mu.RUnlock()

	groups := i.sortGroupsByCapIncreasing()

	// prioritize preferred groups
	preferredGroups := []*ClusterGroup{}

	for _, group := range groups {
		for _, preferredId := range preferredGroupIds {
			if group.id == preferredId {
				preferredGroups = append(preferredGroups, group)
			}
		}
	}
	if len(preferredGroups) > 0 {
		groups = preferredGroups
	}

	// get the groups with the least capacity, possibly multiple groups
	leastCap := groups[0].currentCap
	leastCapGroups := []*ClusterGroup{}
	for _, group := range groups {
		if group.currentCap == leastCap {
			leastCapGroups = append(leastCapGroups, group)
		} else {
			break
		}
	}

	// randomly select a group
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	selectedGroup := leastCapGroups[r.Intn(len(leastCapGroups))]

	return selectedGroup, nil
}

func (i *ClusterGroupInfos) sortGroupsByCap(decreasingOrder bool) []*ClusterGroup {
	var groups []*ClusterGroup
	for _, group := range i.groups {
		if group.currentCap > 0 {
			groups = append(groups, group)
		}
	}
	sort.Slice(groups, func(i, j int) bool {
		if decreasingOrder {
			return groups[i].currentCap > groups[j].currentCap
		} else {
			return groups[i].currentCap < groups[j].currentCap
		}
	})
	return groups
}

// sortGroupsByCapIncreasing sorts the groups from the least to the most capacity.
func (i *ClusterGroupInfos) sortGroupsByCapIncreasing() []*ClusterGroup {
	return i.sortGroupsByCap(false)
}

// sortGroupsByCapDecreasing sorts the groups from the most to the least capacity.
func (i *ClusterGroupInfos) sortGroupsByCapDecreasing() []*ClusterGroup {
	return i.sortGroupsByCap(true)
}

func getGroupIDs(groups []*ClusterGroup) []string {
	ids := []string{}
	for _, group := range groups {
		ids = append(ids, group.id)
	}
	return ids
}

// GroupIdentifier is a function that identifies the group(s) of a node.
type GroupIdentifier func(node *v1.Node) ([]*ClusterGroup, error)

func ClusterGroupIdentifier(node *v1.Node) ([]*ClusterGroup, error) {
	if node.Labels == nil {
		return nil, fmt.Errorf("node labels are missing")
	}
	groups := []*ClusterGroup{}
	groupID, exist := node.Labels[bmenroll.ClusterGroupID]
	if exist {
		groups = append(groups, &ClusterGroup{
			id:        groupID,
			groupType: ClusterGroupType,
		})
	}

	return groups, nil
}

func SuperComputeGroupIdentifier(node *v1.Node) ([]*ClusterGroup, error) {
	if node.Labels == nil {
		return nil, fmt.Errorf("node labels are missing")
	}

	groups := []*ClusterGroup{}
	subGroups := map[string]*ClusterGroup{}

	groupID, exist := node.Labels[bmenroll.ClusterGroupID]
	if exist {
		clustergroup := &ClusterGroup{
			id:        groupID,
			groupType: ClusterGroupType,
		}
		subGroups[clustergroup.id] = clustergroup
		groups = append(groups, clustergroup)
	}

	supercomputeGroupID, exist := node.Labels[bmenroll.SuperComputeGroupID]
	if exist {
		superComputeGroup := &ClusterGroup{
			id:        supercomputeGroupID,
			groupType: SuperComputeGroupType,
			subGroups: subGroups,
		}
		groups = append(groups, superComputeGroup)
	}

	return groups, nil
}

type ClusterGroupInfosOption func(infos *ClusterGroupInfos)

func WithGroupIdentifier(identifier GroupIdentifier) ClusterGroupInfosOption {
	return func(i *ClusterGroupInfos) {
		i.groupIdentifier = identifier
	}
}

func WithNodeSelector(nodeSelector *v1.NodeSelector) ClusterGroupInfosOption {
	return func(i *ClusterGroupInfos) {
		i.nodeSelector = nodeSelector
	}
}

type ClusterGroupInfosFilter func(group *ClusterGroup) bool

func WithGroupFilters(filters ...ClusterGroupInfosFilter) ClusterGroupInfosOption {
	return func(i *ClusterGroupInfos) {
		i.groupFilters = filters
	}
}

func FilterGroupsWithNetworkMode(networkMode string) ClusterGroupInfosFilter {
	return func(group *ClusterGroup) bool {
		return group.networkMode == networkMode
	}
}

func FilterGroupsWithMinimumCurrentCap(min int) ClusterGroupInfosFilter {
	return func(group *ClusterGroup) bool {
		return group.Fit(min)
	}
}
