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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// LabelPool specifies the pool a NodeGroup belong to.
	LabelPool = "pool"
	// LabelMaintenance - values: true/false, when set to true, NodeGroup controller should not work on it.
	LabelMaintenance = "maintenance"
)

const (
	// maintenance works are in progress. for example, moving a NG from VV to VB
	NGMaintenancePhaseInProgress = "inProgress"
	// maintenance works are done, waiting for the NG/NN/SW to be ready. for example, when moving a NG from VV to VB, updating NG Spec is done
	// and waiting for the SwitchPort Vlan and BGP configuration to complete.
	NGMaintenancePhaseWaitingForReady = "waitingForReady"
)

/*
	NetworkNodes, FrontEndLeafSwitches and AcceleratorLeafSwitches are the devices that belong to a NodeGroup.
	NetworkNode CRD has the connections between the NetworkNode and the LeafSwitches. All these above represents
	the topology of a NodeGroup.

	FrontEndFabricConfig and AcceleratorFabricConfig define the desired values for network configuration(ie, Vlan or BGP).
*/
// NodeGroupSpec defines the desired state of NodeGroup
type NodeGroupSpec struct {
	NetworkNodes            []string `json:"networkNodes,omitempty"`
	FrontEndLeafSwitches    []string `json:"frontEndLeafSwitches,omitempty"`
	AcceleratorLeafSwitches []string `json:"acceleratorLeafSwitches,omitempty"`
	StorageLeafSwitches     []string `json:"storageLeafSwitches,omitempty"`

	FrontEndFabricConfig    *FabricConfig `json:"frontEndFabricConfig,omitempty"`
	AcceleratorFabricConfig *FabricConfig `json:"acceleratorFabricConfig,omitempty"`
	StorageFabricConfig     *FabricConfig `json:"storageFabricConfig,omitempty"`
}

type FabricConfig struct {
	VlanConf *VlanConfig `json:"vlanConf,omitempty"`
	BGPConf  *BGPConfig  `json:"bgpConf,omitempty"`
}

type ReconcileStrategy struct {
	EnforceConfig bool `json:"enforceConfig,omitempty"`
}

type VlanConfig struct {
	//+kubebuilder:default=-1
	VlanID int64 `json:"vlanID,omitempty"`
}

type VlanConfigStatus struct {
	Readiness              string                  `json:"readinessByNetworkNode,omitempty"`
	ReadinessByNetworkNode []NetworkNodeVlanStatus `json:"networkNodeVlanStatus,omitempty"`
	// LastObservedReadyVLAN is the last Vlan value in ready state(all Vlan configs are in-sync with the desired value)
	LastObservedReadyVLAN int64 `json:"lastObservedReadyVLAN,omitempty"`
	Ready                 bool  `json:"ready,omitempty"`
}

// NetworkNodeVlanStatus is a wrapper of the switchPort-vlan map and the NetworkNode it belongs to.
type NetworkNodeVlanStatus struct {
	// key: switch port Name. value: vlan id
	LastObservedVlanID map[string]int64 `json:"lastObservedVlanID,omitempty"`
	NetworkNodeName    string           `json:"networkNodeName,omitempty"`
}

type BGPConfig struct {
	//+kubebuilder:default=-1
	BGPCommunity int64 `json:"bgpCommunity,omitempty"`
}

type BGPConfigStatus struct {
	Readiness           string                  `json:"readinessBySwitch,omitempty"`
	SwitchBGPConfStatus []SwitchBGPConfigStatus `json:"switchBGPConfStatus,omitempty"`
	// LastObservedReadyBGP is the last BGP value in ready state(all switches' BGP configs are in-sync with the desired value)
	LastObservedReadyBGP int64 `json:"lastObservedReadyBGP,omitempty"`
	// Ready
	Ready bool `json:"ready,omitempty"`
}

type SwitchBGPConfigStatus struct {
	SwitchFQDN               string `json:"switchFQDN,omitempty"`
	LastObservedBGPCommunity int64  `json:"lastObservedBGPCommunity,omitempty"`
}

// NodeGroupStatus defines the observed state of NodeGroup
type NodeGroupStatus struct {
	NetworkNodesCount       int                 `json:"networkNodesCount,omitempty"`
	FrontEndSwitchCount     int                 `json:"frontEndSwitchCount,omitempty"`
	StorageSwitchCount      int                 `json:"storageSwitchCount,omitempty"`
	AccSwitchCount          int                 `json:"accSwitchCount,omitempty"`
	FrontEndFabricStatus    *FabricConfigStatus `json:"frontEndFabricStatus,omitempty"`
	AcceleratorFabricStatus *FabricConfigStatus `json:"acceleratorFabricStatus,omitempty"`
	StorageFabricStatus     *FabricConfigStatus `json:"storageFabricStatus,omitempty"`
}

type FabricConfigStatus struct {
	VlanConfigStatus VlanConfigStatus `json:"vlanConfigStatus,omitempty"`
	BGPConfigStatus  BGPConfigStatus  `json:"bgpConfigStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=ng
//+kubebuilder:printcolumn:name="pool",type=string,JSONPath=`.metadata.labels.pool`
//+kubebuilder:printcolumn:name="maintenance",type=string,JSONPath=`.metadata.labels.maintenance`
//+kubebuilder:printcolumn:name="NN_Count",type=string,JSONPath=`.status.networkNodesCount`
//+kubebuilder:printcolumn:name="FE_SwCount",type=string,JSONPath=`.status.frontEndSwitchCount`
//+kubebuilder:printcolumn:name="STG_SwCount",type=string,JSONPath=`.status.storageSwitchCount`
//+kubebuilder:printcolumn:name="ACC_SwCount",type=string,JSONPath=`.status.accSwitchCount`
//+kubebuilder:printcolumn:name="FE_Vlan",type=string,JSONPath=`.spec.frontEndFabricConfig.vlanConf.vlanID`
//+kubebuilder:printcolumn:name="FE_Ready(Nodes)",type=string,JSONPath=`.status.frontEndFabricStatus.vlanConfigStatus.readinessByNetworkNode`
//+kubebuilder:printcolumn:name="FE_BGP",type=string,JSONPath=`.spec.frontEndFabricConfig.bgpConf.bgpCommunity`
//+kubebuilder:printcolumn:name="FE_Ready(Switches)",type=string,JSONPath=`.status.bgpConfigStatus.bgpConfigStatus.readinessBySwitch`
//+kubebuilder:printcolumn:name="Acc_Vlan",type=string,JSONPath=`.spec.acceleratorFabricConfig.vlanConf.vlanID`
//+kubebuilder:printcolumn:name="Acc_Ready(Nodes)",type=string,JSONPath=`.status.acceleratorFabricStatus.vlanConfigStatus.readinessByNetworkNode`
//+kubebuilder:printcolumn:name="Acc_BGP",type=string,JSONPath=`.spec.acceleratorFabricConfig.bgpConf.bgpCommunity`
//+kubebuilder:printcolumn:name="Acc_Ready(Switches)",type=string,JSONPath=`.status.acceleratorFabricStatus.bgpConfigStatus.readinessBySwitch`
//+kubebuilder:printcolumn:name="Strg_Vlan",type=string,JSONPath=`.spec.storageFabricConfig.vlanConf.vlanID`
//+kubebuilder:printcolumn:name="Strg_Ready(Nodes)",type=string,JSONPath=`.status.storageFabricStatus.vlanConfigStatus.readinessByNetworkNode`
//+kubebuilder:printcolumn:name="Strg_BGP",type=string,JSONPath=`.spec.storageFabricConfig.bgpConf.bgpCommunity`
//+kubebuilder:printcolumn:name="Strg_Ready(Switches)",type=string,JSONPath=`.status.storageFabricStatus.bgpConfigStatus.readinessBySwitch`

// NodeGroup is the Schema for the nodegroups API
type NodeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeGroupSpec   `json:"spec,omitempty"`
	Status NodeGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeGroupList contains a list of NodeGroup
type NodeGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeGroup{}, &NodeGroupList{})
}
