// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import "time"

// Application configuration
type Config struct {
	// The address of the fleet admin service in format "host:port".
	FleetAdminServerAddr string `koanf:"fleetAdminServerAddr"`
	// The default value is false which applies patches to the local k8s cluster. If this is set to true,the remoteKubeConfigDir must be
	// specified enabling the resource patcher to apply patches to a remote k8s cluster
	ApplyRemoteKubeConfig bool `koanf:"applyRemoteKubeConfig"`
	// Path inside vault containing KubeConfig file of the remote k8s cluster.
	RemoteKubeConfigFilePath string `koanf:"remoteKubeConfigFilePath"`
	// Interval of time between attempts to obtain the resource patches that needs to be applied on the k8s cluster.
	PatchApplyInterval time.Duration `koanf:"patchApplyInterval"`
	// Set of label prefixes which are allowed to be updated on the remote k8s cluster.
	AllowedLabelPrefixes []string `koanf:"allowedLabelPrefixes"`
	// The cluster ID that the Instance Scheduler uses to identify the Harvester or Metal3 Kubernetes cluster
	ClusterId        string `koanf:"clusterId"`
	Region           string `koanf:"region"`
	AvailabilityZone string `koanf:"availabilityZone"`
}
