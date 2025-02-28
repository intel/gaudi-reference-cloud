// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

// Configuration shared with BmInstanceOperatorConfig and VmInstanceOperatorConfig.
type InstanceOperatorConfig struct {
	// Http Client related configuration.
	HttpClient HttpClientConfig `json:"httpClient"`
	// Details for connecting to the Kubernetes API Server of the cluster running SSH Proxy Operator.
	SshProxyTunnelCluster SshProxyTunnelClusterConfig `json:"sshProxyTunnelCluster"`
	// Filter the Instance objects to be reconciled.
	Filter InstanceOperatorFilter `json:"filter"`
	// If true, this operator will add the Metering Monitor finalizer to all instances.
	EnableMeteringMonitorFinalizer bool `json:"enableMeteringMonitorFinalizer"`
	// Configuration for DHCP, DNS, IP Address Management (DDI).
	DdiConfig `json:"ddi,omitempty"`
	// Format should be "compute-api-server:80"
	ComputeApiServerAddr string `json:"computeApiServerAddr"`
	// Feature flags used by vm and bm instance operator
	OperatorFeatureFlags OperatorFeatureFlags `json:"operatorFeatureFlags"`
	// Quick Connect host, i.e. connect.us-region-1.devcloudtenant.io
	QuickConnectHost string `json:"quickConnectHost,omitempty"`
	// Storage cluster addr, i.e. vip1.vast-pdx09-1.us-qa1-1.cloud.intel.com:2049
	StorageClusterAddr string `json:"storageClusterAddr,omitempty"`
}

type SshProxyTunnelClusterConfig struct {
	// Path to the KubeConfig file for the Sshproxytunnel cluster.
	// If blank, use local Kubernetes cluster.
	KubeConfigFilePath string `json:"kubeConfigFilePath,omitempty"`
}

type InstanceOperatorFilter struct {
	// Instances must have all specified labels. Leave blank to have no filter.
	Labels map[string]string `json:"labels"`
}

// Configuration for DHCP, DNS, IP Address Management (DDI).
type DdiConfig struct {
	Method DdiMethod     `json:"method,omitempty"`
	Mmws   DdiConfigMmws `json:"mmws,omitempty"`
}

type DdiMethod string

// These are the valid values for DdiMethod.
const (
	DdiMethodDefault DdiMethod = ""
	// Use Men & Mice DDI Proxy for DDI.
	DdiMethodMmws DdiMethod = "mmws"
)

// Configuration for DDI using Men & Mice DDI Proxy.
type DdiConfigMmws struct {
	// URL of Men & Mice DDI Proxy.
	// Example: http://localhost:57930/mmws/api
	BaseUrl string `json:"baseUrl,omitempty"`
	// Path to the file containing the user name to authenticate to Men & Mice.
	UserNameFilePath string `json:"userNameFilePath,omitempty"`
	// Path to the file containing the password to authenticate to Men & Mice.
	PasswordFilePath string `json:"passwordFilePath,omitempty"`
}

type OperatorFeatureFlags struct {
	// Makes creation of quick connect client CA secret optional.
	EnableQuickConnectClientCA bool `json:"enableQuickConnectClientCA"`
	//Whether to create a VirtualMachine spec for a Harvester or KubeVirt cluster
	UseKubeVirtCluster bool `json:"useKubeVirtCluster"`
}
