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

// PortChannelSpec defines the desired state of PortChannel
type PortChannelSpec struct {
	Name string `json:"name"` // eg. Port-Channel51
	Mode string `json:"mode"`
	//+kubebuilder:default=-1
	VlanId int64 `json:"vlanId,omitempty"`
	//+kubebuilder:default=-1
	NativeVlan  int64     `json:"nativeVlan,omitempty"`
	Description string    `json:"description,omitempty"`
	LinkStatus  string    `json:"linkStatus,omitempty"`
	Bandwidth   int       `json:"bandwidth,omitempty"`
	TrunkGroups *[]string `json:"trunkGroups,omitempty"`
}

// PortChannelStatus defines the observed state of PortChannel
type PortChannelStatus struct {
	Name                 string      `json:"name"` // eg. Port-Channel51
	Mode                 string      `json:"mode"`
	VlanId               int64       `json:"vlanId,omitempty"`
	NativeVlan           int64       `json:"nativeVlan,omitempty"`
	Description          string      `json:"description,omitempty"`
	LinkStatus           string      `json:"linkStatus,omitempty"`
	Bandwidth            int         `json:"bandwidth,omitempty"`
	TrunkGroups          []string    `json:"trunkGroups,omitempty"`
	UntaggedVlan         int64       `json:"untaggedVlan,omitempty"`
	LastStatusChangeTime metav1.Time `json:"lastStatusChangeTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=portc;pchan
//+kubebuilder:printcolumn:name="Switch_FQDN",type=string,JSONPath=`.metadata.labels.switch_fqdn`
//+kubebuilder:printcolumn:name="Interface",type=string,JSONPath=`.status.name`
//+kubebuilder:printcolumn:name="Desired_Mode",type=string,JSONPath=`.spec.mode`
//+kubebuilder:printcolumn:name="Observed_Mode",type=string,JSONPath=`.status.mode`
//+kubebuilder:printcolumn:name="Desired_VlanID",type=string,JSONPath=`.spec.vlanId`
//+kubebuilder:printcolumn:name="Observed_VlanID",type=string,JSONPath=`.status.vlanId`
//+kubebuilder:printcolumn:name="Desired_TG",type=string,JSONPath=`.spec.trunkGroups`
//+kubebuilder:printcolumn:name="Observed_TG",type=string,JSONPath=`.status.trunkGroups`
//+kubebuilder:printcolumn:name="Link_Status",type=string,JSONPath=`.status.linkStatus`
//+kubebuilder:printcolumn:name="Last_Status_Change_Time",type=string,JSONPath=`.status.lastStatusChangeTime`

// PortChannel is the Schema for the portchannels API
type PortChannel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PortChannelSpec   `json:"spec,omitempty"`
	Status PortChannelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PortChannelList contains a list of PortChannel
type PortChannelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PortChannel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PortChannel{}, &PortChannelList{})
}
