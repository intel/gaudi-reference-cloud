/*
Copyright 2023.

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

package v1alpha1

type Pool struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// StandaloneNodeGroupOnly indicates that all the NodeGroups in this Pool are standalone(not connected to a spine).
	StandaloneNodeGroupOnly bool `json:"standaloneNodeGroupOnly,omitempty"`

	NetworkConfigStrategy *NetworkConfigStrategy `json:"networkConfigStrategy,omitempty"`

	SchedulingConfig *SchedulingConfig `json:"schedulingConfig,omitempty"`
}

type NetworkConfigStrategy struct {
	// FrontEndFabricStrategy defines the network configuration strategy for the front-end fabric
	FrontEndFabricStrategy *NetworkStrategy `json:"frontEndFabricStrategy,omitempty"`
	// AcceleratorFabricStrategy defines the network configuration strategy for the accelerator fabric
	AcceleratorFabricStrategy *NetworkStrategy `json:"acceleratorFabricStrategy,omitempty"`
	// StorageFabricStrategy defines the network configuration strategy for the storage fabric
	StorageFabricStrategy *NetworkStrategy `json:"storageFabricStrategy,omitempty"`
}

type SchedulingConfig struct {
	// MinimumSchedulableUnit defines the minimum unit can be scheduled for the items in this Pool. MSU can be NodeGroup or NetworkNode.
	MinimumSchedulableUnit MSU `json:"minimumSchedulableUnit,omitempty"`
}

type PoolList struct {
	Items []*Pool `json:"items,omitempty"`
}

type NetworkStrategy struct {
	// IsolationType defines what isolation solution will be used for a fabric.
	IsolationType IsolationType `json:"isolationType,omitempty"`
	// ProvisionConfig defines the default network configuration for the NodeGroups that are added to a Pool
	ProvisionConfig ProvisionConfig `json:"provisionConfig,omitempty"`
}

type ProvisionConfig struct {
	// If the value of DefaultVlanID is set, when a NodeGroup is assigned to a Pool, all its SwitchPort Vlan should be set to this default value.
	DefaultVlanID *int64 `json:"defaultVlanID,omitempty"`
	// If the value of DefaultBGPCommunity is set, when a NodeGroup is assigned to a Pool, all its Switch should be set to this default BGP community.
	DefaultBGPCommunity *int64 `json:"defaultBGPCommunity,omitempty"`
}

type MSU string

const (
	MSUNodeGroup   MSU = "NodeGroup"
	MSUNetworkNode MSU = "NetworkNode"
)

type IsolationType string

const (
	IsolationTypeVLAN IsolationType = "VLAN"
	IsolationTypeBGP  IsolationType = "BGP"
)

// NodeGroupToPoolMap is a simple mapping struct, to unmarshal the ingest configMap file
type NodeGroupToPoolMap struct {
	NodeGroupToPoolMap map[string]string `json:"nodeGroupToPoolMap,omitempty"`
}
