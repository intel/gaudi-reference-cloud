// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// LoadBalancerReplicatorConfig is the Schema for the loadbalancerreplicatorconfigs API.
// It stores the configuration for the Loadbalancer Replicator.
type LoadBalancerReplicatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the configurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	// Format should be "compute-api-server:80"
	ComputeApiServerAddr string `json:"computeApiServerAddr"`

	// The region in which the replicator is deployed.
	RegionId string `json:"regionId"`

	// The availability zone in which the replicator is deployed.
	AvailabilityZoneId string `json:"availabilityZoneId"`
}

func init() {
	SchemeBuilder.Register(&LoadBalancerReplicatorConfig{})
}
