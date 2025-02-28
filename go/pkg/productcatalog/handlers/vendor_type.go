// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import "time"

// VendorSpec defines the desired state of Vendor
type VendorSpec struct {
	ID          string          `json:"id,omitempty"`
	Description string          `json:"description,omitempty"`
	Families    []ProductFamily `json:"families,omitempty"`
}

type ProductFamily struct {
	Name         string    `json:"name,omitempty"`
	ID           string    `json:"id,omitempty"`
	Description  string    `json:"description,omitempty"`
	CreationTime time.Time `json:"creationTime"`
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

// Vendor is the Schema for the vendors API
type Vendor struct {
	Name         string       `json:"name"`
	CreationTime time.Time    `json:"creationTime"`
	Spec         VendorSpec   `json:"spec,omitempty"`
	Status       VendorStatus `json:"status,omitempty"`
}

// VendorList contains a list of Vendor
type VendorList struct {
	Items []Vendor `json:"items"`
}
