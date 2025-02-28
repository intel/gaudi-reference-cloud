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
	SDNControllerNamespace = "idcs-system"
)

// SwitchSpec defines the desired state of Switch
type SwitchSpec struct {
	FQDN        string     `json:"fqdn"`
	Ip          string     `json:"ip"`
	EAPIConf    *EAPIConf  `json:"eapiConf,omitempty"`
	BGP         *BGPConfig `json:"bgpConf,omitempty"`
	IpOverride  string     `json:"ipOverride,omitempty"`
	Maintenance string     `json:"maintenance,omitempty"`
}

type EAPIConf struct {
	// CredentialPath is the file that stores the eAPI credential for a switch. This file is expected to be injected by Vault.
	CredentialPath string `json:"credentialPath"`
	Port           int    `json:"port"`
	Transport      string `json:"transport"`
}

type EAPISecret struct {
	Credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"credentials"`
}

// SwitchStatus defines the observed state of Switch
type SwitchStatus struct {
	// Conditions            []metav1.Condition     `json:"conditions,omitempty"`
	SwitchBGPConfigStatus *SwitchBGPConfigStatus `json:"switchBGPConfigStatus,omitempty"`
	LastStatusUpdateTime  metav1.Time            `json:"lastStatusUpdateTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=sw
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Maintenance",type=string,JSONPath=`.spec.maintenance`
// +kubebuilder:printcolumn:name="IP_OVERRIDE",type=string,JSONPath=`.spec.ipOverride`
// +kubebuilder:printcolumn:name="desired_bgp_cmty",type=string,JSONPath=`.spec.bgpConf.bgpCommunity`
// +kubebuilder:printcolumn:name="last_observed_bgp_cmty",type=string,JSONPath=`.status.switchBGPConfigStatus.lastObservedBGPCommunity`
// +kubebuilder:printcolumn:name="Fabric_Type",type=string,JSONPath=`.metadata.labels.fabric_type`
// +kubebuilder:printcolumn:name="Last_Status_Update_Time",type=string,JSONPath=`.status.lastStatusUpdateTime`

// Switch is the Schema for the switches API
type Switch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwitchSpec   `json:"spec,omitempty"`
	Status SwitchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SwitchList contains a list of Switch
type SwitchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Switch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Switch{}, &SwitchList{})
}
