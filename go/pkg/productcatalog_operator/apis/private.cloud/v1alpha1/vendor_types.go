// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VendorSpec defines the desired state of Vendor
type VendorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ID          string          `json:"id,omitempty"`
	Description string          `json:"description,omitempty"`
	Families    []ProductFamily `json:"families,omitempty"`
}

type ProductFamily struct {
	Name        string `json:"name,omitempty"`
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
}

// VendorStatus defines the observed state of Vendor
type VendorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	State VendorState `json:"state,omitempty"`
}

type VendorState string

// These are the valid states of vendor resource.
const (
	// ready means the vendor is successfully on-boarded to IDC.
	VendorStateReady VendorState = "ready"
	// error means there is an error on-boarding the vendor to IDC and
	// it can not be on-boarded.
	VendorStateError VendorState = "error"
	// provisioning means the vendor is being on-boarded to IDC and
	// is not yet ready.
	VendorStateProvisioning VendorState = "provisioning"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Vendor is the Schema for the vendors API
type Vendor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VendorSpec   `json:"spec,omitempty"`
	Status VendorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VendorList contains a list of Vendor
type VendorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Vendor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Vendor{}, &VendorList{})
}
