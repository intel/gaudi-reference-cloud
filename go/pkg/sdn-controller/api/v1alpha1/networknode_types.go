// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
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
	DefaultAcceleratorFabricVlan = 100

	// LabelBMHGroupID is the label used in the BMH CR
	LabelBMHGroupID = "cloud.intel.com/instance-group-id"
	// LabelGroupID is the label used in the SDN NetworkNode CR
	LabelGroupID = "groupId"
	// LabelBMHNameSpace
	LabelBMHNameSpace = "bmhns"
)

// NetworkNodeSpec defines the desired state of NetworkNode
type NetworkNodeSpec struct {
	FrontEndFabric    *FrontEndFabric    `json:"frontEndFabric,omitempty"`
	AcceleratorFabric *AcceleratorFabric `json:"acceleratorFabric,omitempty"`
	StorageFabric     *StorageFabric     `json:"storageFabric,omitempty"`
}

// FrontEndFabric
type FrontEndFabric struct {
	//+kubebuilder:default=-1
	VlanId int64 `json:"vlanId,omitempty"`
	// SwitchPort is the name of a SwitchPort CR, ie, "ethernet27-1.internal-placeholder.com"
	SwitchPort string `json:"switchPort,omitempty"`
	// note: when Mode is "", controller will not push the change to the downstream SwitchPort CR
	Mode string `json:"mode,omitempty"`
	// note: when NativeVlan is -1, controller will not push the change to the downstream SwitchPort CR
	//+kubebuilder:default=-1
	NativeVlan int64 `json:"nativeVlan,omitempty"`
	// note: when TrunkGroups is nil, controller will not push the change to the downstream SwitchPort CR
	// when TrunkGroups is NOT nil and TrunkGroups is empty, remove trunk groups for this switch port, if any.
	TrunkGroups *[]string `json:"trunkGroups,omitempty"`
}

type FrontEndFabricStatus struct {
	Readiness               string   `json:"readiness,omitempty"`
	SwitchPort              string   `json:"switchPort,omitempty"`
	LastObservedVlanId      int64    `json:"lastObservedVlanId,omitempty"`
	LastObservedMode        string   `json:"lastObservedMode,omitempty"`
	LastObservedNativeVlan  int64    `json:"lastObservedNativeVlan,omitempty"`
	LastObservedTrunkGroups []string `json:"lastObservedTrunkGroups,omitempty"`
}

// AcceleratorFabric
type AcceleratorFabric struct {
	//+kubebuilder:default=-1
	VlanId int64 `json:"vlanId,omitempty"`
	// SwitchPortName is a list of SwitchPort CR names, ie, []{"ethernet27-1.internal-placeholder.com"}
	SwitchPorts []string `json:"switchPorts,omitempty"`
	Mode        string   `json:"mode,omitempty"`
	//+kubebuilder:default=-1
	NativeVlan  int64     `json:"nativeVlan,omitempty"`
	TrunkGroups *[]string `json:"trunkGroups,omitempty"`
}

type AcceleratorFabricStatus struct {
	Readiness   string                               `json:"readiness,omitempty"`
	SwitchPorts []*AcceleratorFabricSwitchPortStatus `json:"switchPorts,omitempty"`
}

type AcceleratorFabricSwitchPortStatus struct {
	SwitchPort              string   `json:"switchPort,omitempty"`
	LastObservedVlanId      int64    `json:"lastObservedVlanId,omitempty"`
	LastObservedMode        string   `json:"lastObservedMode,omitempty"`
	LastObservedNativeVlan  int64    `json:"lastObservedNativeVlan,omitempty"`
	LastObservedTrunkGroups []string `json:"lastObservedTrunkGroups,omitempty"`
}

// StorageFabric
type StorageFabric struct {
	//+kubebuilder:default=-1
	VlanId      int64    `json:"vlanId,omitempty"`
	SwitchPorts []string `json:"switchPorts,omitempty"`
	Mode        string   `json:"mode,omitempty"`
	//+kubebuilder:default=-1
	NativeVlan  int64     `json:"nativeVlan,omitempty"`
	TrunkGroups *[]string `json:"trunkGroups,omitempty"`
}

type StorageFabricStatus struct {
	Readiness   string                           `json:"readiness,omitempty"`
	SwitchPorts []*StorageFabricSwitchPortStatus `json:"switchPorts,omitempty"`
}

type StorageFabricSwitchPortStatus struct {
	SwitchPort              string   `json:"switchPort,omitempty"`
	LastObservedVlanId      int64    `json:"lastObservedVlanId,omitempty"`
	LastObservedMode        string   `json:"lastObservedMode,omitempty"`
	LastObservedNativeVlan  int64    `json:"lastObservedNativeVlan,omitempty"`
	LastObservedTrunkGroups []string `json:"lastObservedTrunkGroups,omitempty"`
}

// NetworkNodeStatus defines the observed state of NetworkNode
type NetworkNodeStatus struct {
	FrontEndFabricStatus    FrontEndFabricStatus    `json:"frontEndFabricStatus,omitempty"`
	AcceleratorFabricStatus AcceleratorFabricStatus `json:"acceleratorFabricStatus,omitempty"`
	StorageFabricStatus     StorageFabricStatus     `json:"storageFabricStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=nn
//+kubebuilder:printcolumn:name="FEMode",type=string,JSONPath=`.spec.frontEndFabric.mode`
//+kubebuilder:printcolumn:name="FEVLAN",type=string,JSONPath=`.spec.frontEndFabric.vlanId`
//+kubebuilder:printcolumn:name="FETG",type=string,JSONPath=`.spec.frontEndFabric.trunkGroups`
//+kubebuilder:printcolumn:name="FEReadiness",type=string,JSONPath=`.status.frontEndFabricStatus.readiness`
//+kubebuilder:printcolumn:name="AccelVLAN",type=string,JSONPath=`.spec.acceleratorFabric.vlanId`
//+kubebuilder:printcolumn:name="AccelReadiness",type=string,JSONPath=`.status.acceleratorFabricStatus.readiness`
//+kubebuilder:printcolumn:name="STRVLAN",type=string,JSONPath=`.spec.storageFabric.vlanId`
//+kubebuilder:printcolumn:name="STRReadiness",type=string,JSONPath=`.status.storageFabricStatus.readiness`
//+kubebuilder:printcolumn:name="group",type=string,JSONPath=`.metadata.labels.groupId`

// NetworkNode is the Schema for the networknodes API
type NetworkNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkNodeSpec   `json:"spec,omitempty"`
	Status NetworkNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetworkNodeList contains a list of NetworkNode
type NetworkNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkNode{}, &NetworkNodeList{})
}
