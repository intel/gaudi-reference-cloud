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

type Protocol string

const (
	Protocol_TCP Protocol = "TCP"
	Protocol_UDP Protocol = "UDP"
)

type State string

const (
	RECONCILING State = "Reconciling"
	READY       State = "Active"
	DELETED     State = "Deleted"
	DELETING    State = "Deleting"
)

// FirewallRuleSpec defines the desired state of FirewallRule
type FirewallRuleSpec struct {
	SourceIPs     []string `json:"sourceips"`
	DestinationIP string   `json:"destip"`
	Protocol      Protocol `json:"protocol"`
	Port          string   `json:"port"`
}

// FirewallRuleStatus defines the observed state of FirewallRule
type FirewallRuleStatus struct {
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	State   State  `json:"state"`
	Message string `json:"message"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="DestIP",type="string",JSONPath=".spec.destip",description="Destination IP",priority=1
// +kubebuilder:printcolumn:name="Port",type="string",JSONPath=".spec.port",description="Source IP"
// +kubebuilder:printcolumn:name="Protocol",type="string",JSONPath=".spec.protocol",description="Protocol"
// +kubebuilder:printcolumn:name="SourceIPs",type="string",JSONPath=".spec.sourceips",description="Source IPs"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Current state of the FirewallRule"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of the FirewallRule"

// FirewallRule is the Schema for the firewallrules API
type FirewallRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallRuleSpec   `json:"spec,omitempty"`
	Status FirewallRuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FirewallRuleList contains a list of FirewallRule
type FirewallRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirewallRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirewallRule{}, &FirewallRuleList{})
}
