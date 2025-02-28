// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProductSpec defines the desired state of Product
type ProductSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ID          string `json:"id,omitempty"`
	VendorID    string `json:"vendorId,omitempty"`
	FamilyID    string `json:"familyId,omitempty"`
	Description string `json:"description,omitempty"`
	// ECCNs are five character alpha-numeric designations used on the
	// Commerce Control List (CCL) to identify dual-use items for export control purposes.
	ECCN string `json:"eccn,omitempty"`
	// Price Contract Quotation (PCQ)
	PCQ       string            `json:"pcq,omitempty"`
	MatchExpr string            `json:"matchExpr,omitempty"`
	Rates     []ProductRate     `json:"rates,omitempty"`
	Metadata  []ProductMetadata `json:"metadata,omitempty"`
}

type ProductMetadata struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type ProductRate struct {
	AccountType IDCAccountType `json:"accountType,omitempty"`
	Unit        string         `json:"unit,omitempty"`
	UsageExpr   string         `json:"usageExpr,omitempty"`
	Rate        string         `json:"rate,omitempty"`
}

type IDCAccountType string

// These are the valid states of vendor resource.
const (
	// IDC Enterprise Account
	EnterpriseAccountType IDCAccountType = "enterprise"

	// IDC Stanadrd Account
	StandardAccountType IDCAccountType = "standard"

	// IDC Premium Account
	PremiumAccountType IDCAccountType = "premium"

	// IDC Intel Account
	IntelAccountType IDCAccountType = "intel"
)

// ProductStatus defines the observed state of Product
type ProductStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	State ProductState `json:"state,omitempty"`
}

type ProductState string

// These are the valid states of vendor resource.
const (
	// ready means the product is successfully on-boarded to IDC.
	ProductStateReady ProductState = "ready"
	// error means there is an error on-boarding the product to IDC and
	// it can not be on-boarded.
	ProductStateError ProductState = "error"
	// provisioning means the product is being on-boarded to IDC and
	// is not yet ready.
	ProductStateProvisioning ProductState = "provisioning"
	// unspecified means the product status can not be determined
	ProductStateUndetermined ProductState = "undetermine"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Product is the Schema for the products API
type Product struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProductSpec   `json:"spec,omitempty"`
	Status ProductStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProductList contains a list of Product
type ProductList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Product `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Product{}, &ProductList{})
}
