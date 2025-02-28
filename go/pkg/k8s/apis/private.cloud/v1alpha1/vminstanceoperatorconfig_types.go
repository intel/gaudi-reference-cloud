// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// VmInstanceOperatorConfig is the Schema for the vminstanceoperatorconfigs API.
// It stores the configuration for the Instance Operator.
type VmInstanceOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the configurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`
	// Configuration shared with BmInstanceOperatorConfig and VmInstanceOperatorConfig.
	InstanceOperator InstanceOperatorConfig `json:"instanceOperator"`
	// VmProvisionConfig related configuration
	VmProvisionConfig VmProvisionConfig `json:"vmProvisionConfig"`
	// List of Subnets Assigned for the storage server
	StorageServerSubnets []string `json:"storageServerSubnets"`
}

type CloudInitUser struct {
	Name                string   `json:"name,omitempty"`
	Sudo                string   `json:"sudo,omitempty"`
	Password            string   `json:"passwd,omitempty"`
	Lock_passwd         bool     `json:"lock_passwd,omitempty"`
	Ssh_authorized_keys []string `json:"ssh_authorized_keys,omitempty"`
	Shell               string   `json:"shell,omitempty"`
}

type CloudInitWriteFile struct {
	Owner       string `json:"owner,omitempty"`
	Path        string `json:"path,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	Content     string `json:"content,omitempty"`
}

type CloudInitUserData struct {
	Write_files         []CloudInitWriteFile `json:"write_files,omitempty"`
	Package_update      bool                 `json:"package_update,omitempty"`
	Manage_etc_hosts    string               `json:"manage_etc_hosts,omitempty"`
	Hostname            string               `json:"hostname,omitempty"`
	Fqdn                string               `json:"fqdn,omitempty"`
	RunCmd              [][]string           `json:"runcmd,omitempty"`
	Users               []string             `json:"users,omitempty"`
	Ssh_authorized_keys []string             `json:"ssh_authorized_keys,omitempty"`
}

type Route struct {
	Gateway     string `json:"gateway,omitempty"`
	Netmask     string `json:"netmask,omitempty"`
	Destination string `json:"destination,omitempty"`
}

type CloudInitSubnet struct {
	Type            string   `json:"type,omitempty"`
	Address         string   `json:"address,omitempty"`
	Gateway         string   `json:"gateway,omitempty"`
	Dns_nameservers []string `json:"dns_nameservers,omitempty"`
	Routes          []Route  `json:"routes,omitempty"`
}

type CloudInitSubnetConfig struct {
	Type    string            `json:"type,omitempty"`
	Name    string            `json:"name,omitempty"`
	Subnets []CloudInitSubnet `json:"subnets,omitempty"`
}

type CloudInitNetworkConfig struct {
	Version int64                   `json:"version,omitempty"`
	Config  []CloudInitSubnetConfig `json:"config,omitempty"`
}

type CloudInitNetworkData struct {
	Network CloudInitNetworkConfig `json:"network,omitempty"`
}

type CloudInitData struct {
	UserData    CloudInitUserData    `json:"userData"`
	NetworkData CloudInitNetworkData `json:"networkData"`
}

type VmProvisionConfig struct {
	// Path to the KubeConfig file for the Harvester/Kubevirt cluster.
	// If blank, use local Kubernetes cluster.
	KubeConfigFilePath string `json:"kubeConfigFilePath,omitempty"`
	// VM instances will be connected to this Harvester/Kubevirt ClusterNetwork.
	VMclusterNetwork string `json:"vmClusterNetwork"`
	// VM instances will be connected to this Storage ClusterNetwork.
	StorageClusterNetwork string `json:"storageClusterNetwork"`
	// A cloud-init config for user-data source
	CloudInitData CloudInitData `json:"cloudInitData"`
}

type HttpClientConfig struct {
	TimeoutInSecs       int64 `json:"timeoutInSecs"`
	KeepAlive           int64 `json:"keepAlive"`
	TLSHandshakeTimeout int64 `json:"tlsHandshakeTimeout"`
	InsecureSkipVerify  bool  `json:"insecureSkipVerify"`
}

func init() {
	SchemeBuilder.Register(&VmInstanceOperatorConfig{})
}
