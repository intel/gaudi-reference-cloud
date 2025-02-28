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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type SwitchPortSyncState string

// +enum
const (
	LabelNameSwitchFQDN  = "switch_fqdn"
	LabelNameNetworkNode = "network_node"
	LabelFabricType      = "fabric_type"
)

const (
	FabricTypeFrontEnd    = "frontEnd"
	FabricTypeAccelerator = "accelerator"
	FabricTypeStorage     = "storage"
)

const (
	AccessMode = "access"
	TrunkMode  = "trunk"
)

const (
	NOOPVlanID       = -1
	NOOPBGPCommunity = -1
	NOOPPortChannel  = -1 // -1 really means something different to 0 (0 means not in a port channel, -1 means don't change the portchannel on the switch)
)

// SwitchPortSpec defines the desired state of SwitchPort
type SwitchPortSpec struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
	//+kubebuilder:default=-1
	VlanId int64 `json:"vlanId,omitempty"`
	//+kubebuilder:default=-1
	NativeVlan  int64     `json:"nativeVlan,omitempty"`
	Description string    `json:"description,omitempty"`
	TrunkGroups *[]string `json:"trunkGroups,omitempty"`
	// Default this to -1 so that when we upgrade CRDs the existing CRs will get -1 set.
	//+kubebuilder:default=-1
	PortChannel int64 `json:"portChannel"` // Must not omitempty because otherwise PATCHes to set this field to 0 will get overridden by the default.
}

// SwitchPortStatus defines the observed state of SwitchPort
type SwitchPortStatus struct {
	Name                                string      `json:"name"`
	Mode                                string      `json:"mode"`
	VlanId                              int64       `json:"vlanId,omitempty"`
	NativeVlan                          int64       `json:"nativeVlan,omitempty"`
	Description                         string      `json:"description,omitempty"`
	LinkStatus                          string      `json:"linkStatus,omitempty"`
	Bandwidth                           int         `json:"bandwidth,omitempty"`
	InterfaceType                       string      `json:"interfaceType,omitempty"`
	LineProtocolStatus                  string      `json:"lineProtocolStatus,omitempty"`
	TrunkGroups                         []string    `json:"trunkGroups,omitempty"`
	UntaggedVlan                        int64       `json:"untaggedVlan,omitempty"` // Unused. Would removal be a breaking change to protobuf encoding?
	PortChannel                         int64       `json:"portChannel,omitempty"`
	Duplex                              string      `json:"duplex,omitempty"`
	SwitchSideLastStatusChangeTimestamp int64       `json:"switchSideLastStatusChangeTimestamp,omitempty"`
	LastStatusChangeTime                metav1.Time `json:"lastStatusChangeTime,omitempty"`

	RavenDBVlanId int64 `json:"ravenDBVlanId,omitempty"` // deprecated
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=sp
//+kubebuilder:printcolumn:name="Interface",type=string,JSONPath=`.status.name`
//+kubebuilder:printcolumn:name="Desired_Mode",type=string,JSONPath=`.spec.mode`
//+kubebuilder:printcolumn:name="Observed_Mode",type=string,JSONPath=`.status.mode`
//+kubebuilder:printcolumn:name="Desired_VlanID",type=string,JSONPath=`.spec.vlanId`
//+kubebuilder:printcolumn:name="Observed_VlanID",type=string,JSONPath=`.status.vlanId`
//+kubebuilder:printcolumn:name="Desired_TG",type=string,JSONPath=`.spec.trunkGroups`
//+kubebuilder:printcolumn:name="Observed_TG",type=string,JSONPath=`.status.trunkGroups`
//+kubebuilder:printcolumn:name="Desired_PortChannel",type=string,JSONPath=`.spec.portChannel`
//+kubebuilder:printcolumn:name="Observed_PortChannel",type=string,JSONPath=`.status.portChannel`
//+kubebuilder:printcolumn:name="Network_Node",type=string,JSONPath=`.metadata.labels.network_node`
//+kubebuilder:printcolumn:name="Link_Status",type=string,JSONPath=`.status.linkStatus`
//+kubebuilder:printcolumn:name="Line_Protocol_Status",type=string,JSONPath=`.status.lineProtocolStatus`
//+kubebuilder:printcolumn:name="Fabric_Type",type=string,JSONPath=`.metadata.labels.fabric_type`
//+kubebuilder:printcolumn:name="Last_Status_Change_Time",type=string,JSONPath=`.status.lastStatusChangeTime`

// SwitchPort is the Schema for the switchports API
type SwitchPort struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwitchPortSpec   `json:"spec,omitempty"`
	Status SwitchPortStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SwitchPortList contains a list of SwitchPort
type SwitchPortList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwitchPort `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwitchPort{}, &SwitchPortList{})
}
