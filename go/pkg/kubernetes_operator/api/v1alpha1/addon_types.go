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
	// ActiveAddonState represents an addon that was successfully put into the downstream cluster.
	ActiveAddonState AddonState = "Active"
	// UpdatingAddonState means that the addon will be put again into the downstream cluster due to
	// a change in its manifest.
	UpdatingAddonState AddonState = "Updating"
	// ErrorAddonState represents an error when putting the manifest into the downstream cluster.
	ErrorAddonState AddonState = "Error"
	// DeletingAddonState happens when the addon is removed from the cluster addon list.
	DeletingAddonState AddonState = "Deleting"

	// KubectlApplyAddonType specifies the kubectl addon provider and the apply action.
	KubectlApplyAddonType AddonType = "kubectl-apply"
	// KubectlReplaceAddonType specifies the kubectl addon provider and the replace action.
	KubectlReplaceAddonType AddonType = "kubectl-replace"
)

type AddonType string
type AddonState string

// AddonSpec defines the desired state of Addon
type AddonSpec struct {
	// ClusterName specifies the name of the cluster that owns this nodegroup.
	ClusterName string `json:"clusterName"`
	// Type specifies the provider and action to use for puting the addon into the cluster.
	Type AddonType `json:"type"`
	// Artifact specifies the url of the manifest that will be installed.
	Artifact string `json:"artifact"`
	// Args are the required variables to create a configured manifest out of the template artifact.
	// +optional
	Args map[string]string `json:"args"`
	// The IP of the kube-apiserver loadbalancer.
	APIServerLB string `json:"apiserverLB"`
	// The port of the kube-apiserver loadbalancer.
	APIServerLBPort string `json:"apiserverLBPort"`
}

// AddonStatus defines the observed state of Addon
type AddonStatus struct {
	// Name specifies the name of the addon.
	Name string `json:"name"`
	// State of the addon, one of Active, Updating, Error, Deleting.
	State AddonState `json:"state"`
	// Reason is a description of the current state.
	// +optional
	Reason string `json:"reason"`
	// Message is a more verbose description of the current state.
	// +optional
	Message string `json:"message"`
	// LastUpdate specifies when the state of the addon last updated.
	// +optional
	LastUpdate metav1.Time `json:"lastUpdate"`
	// Artifact is the last url used to download and install the manifest.
	Artifact string `json:"artifact"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Addon is the Schema for the addons API
type Addon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonSpec   `json:"spec,omitempty"`
	Status AddonStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AddonList contains a list of Addon
type AddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Addon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Addon{}, &AddonList{})
}
