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

// NodeGroupToPoolMappingSpec defines the desired state of NodeGroupToPoolMapping
type NodeGroupToPoolMappingSpec struct {
	Pool string `json:"pool,omitempty"`
}

// NodeGroupToPoolMappingStatus defines the observed state of NodeGroupToPoolMapping
type NodeGroupToPoolMappingStatus struct {
	LastChangeTime metav1.Time `json:"lastChangeTime,omitempty"`
}

//+kubebuilder:object:root=true
// +kubebuilder:resource:shortName=mapping
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Pool",type=string,JSONPath=`.spec.pool`
// +kubebuilder:printcolumn:name="Last_Change_Time",type=string,JSONPath=`.status.lastChangeTime`

// NodeGroupToPoolMapping is the Schema for the nodegrouptopoolmappings API
type NodeGroupToPoolMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeGroupToPoolMappingSpec   `json:"spec,omitempty"`
	Status NodeGroupToPoolMappingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeGroupToPoolMappingList contains a list of NodeGroupToPoolMapping
type NodeGroupToPoolMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeGroupToPoolMapping `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeGroupToPoolMapping{}, &NodeGroupToPoolMappingList{})
}
