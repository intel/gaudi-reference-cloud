// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// VmInstanceSchedulerConfig is the schema for configuration for the VM Instance Scheduler.
type VmInstanceSchedulerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the configurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	// Directory containing KubeConfig files for all Harvester/KubeVirt clusters that this scheduler will manage.
	VmClustersKubeConfigDir string `json:"vmClustersKubeConfigDir"`
	// Directory containing KubeConfig files for metal3 that this scheduler will manage.
	// This value will be utilized solely for the testing environment for now
	BmKubeConfigDir string `json:"bmKubeConfigDir"`

	// GRPC server will listen on this port.
	ListenPort uint16 `json:"listenPort"`

	// Duration the scheduler will wait before expiring an assumed pod.
	// If the pod for an instance is not created within this duration after scheduling,
	// resources reserved for it will be unreserved.
	DurationToExpireAssumedPod metav1.Duration `json:"durationToExpireAssumedPod"`
	// Support BMaaS scheduler
	EnableBMaaSLocal bool `json:"enableBmaasLocal"`

	// Enable binpacking for BMaaS scheduler
	EnableBMaaSBinpack bool `json:"enableBmaasBinpack"`

	// The number of instances at the threshold will use the BGP network. if 0, no BGP will be used.
	BGPNetworkRequiredInstanceCountThreshold int `json:"bgpNetworkRequiredInstanceCountThreshold"`

	// Overcommit-config settings for cpu, memory and storage for the harvester cluster
	OvercommitConfig OvercommitConfig `json:"overcommitConfig"`

	// Region name.
	Region string `json:"region"`

	// Availability Zone
	AvailabilityZone string `json:"availabilityZone"`

	// The address of the Compute API Server in format "host:port"
	ComputeApiServerAddr string `json:"computeApiServerAddr"`
}

type OvercommitConfig struct {
	// CPU allocation ratio in terms of percentage. For example if we want 2:1 cpu overprovisioing ratio then cpu allocation ratio should be set to 200.
	CPU uint16 `json:"cpu"`
	// Memory allocation ratio in terms of percentage. For example if we want 2:1 memory overprovisioing ratio then memory allocation ratio should be set to 200.
	Memory uint16 `json:"memory"`
	// Storage allocation ratio in terms of percentage. For example if we want 3:1 storage overprovisioing ratio then storage allocation ratio should be set to 300.
	Storage uint16 `json:"storage"`
}

func init() {
	SchemeBuilder.Register(&VmInstanceSchedulerConfig{})
}
