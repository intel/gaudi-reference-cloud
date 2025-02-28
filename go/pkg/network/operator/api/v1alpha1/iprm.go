// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright 2023.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PortStatus defines the observed state of Port
type PortStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Port is the Schema for the Port API
type Port struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   string `json:"spec"`
	Status string `json:"status"`
}

//+kubebuilder:object:root=true

// PortList contains a list of Port
type PortList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Port `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VPC{}, &VPCList{})
	SchemeBuilder.Register(&Subnet{}, &SubnetList{})
	SchemeBuilder.Register(&Port{}, &PortList{})
}
