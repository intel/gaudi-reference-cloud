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
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// SDNControllerConfig is the Schema for the sdncontrollerconfigs API
type SDNControllerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Configuration for the controller manager
	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	// Configuration for the CRD controllers
	ControllerConfig ControllerConfig `json:"controllerConfig" yaml:"controllerConfig"`
}

const (
	//	"eapi" mode talks to a switch directly for making changes.
	SwitchBackendModeEAPI = "eapi"
	//	"mock" mode is for testing. SDN-Controller will talk to a "Mock Switch" instead of a real switch.
	SwitchBackendModeMock = "mock"
	//	"readonly" mode is the same as the eapi client except not doing actual update to the switch.
	SwitchBackendModeReadOnly = "readonly"
)

const (
	SwitchImportSourceNetbox = "netbox"

	SwitchPortImportSourceBMH    = "bmh"
	SwitchPortImportSourceNetbox = "netbox"
)

const (
	// SDNManagerConfigMapName is the name of the SDN Controller Manager ConfigMap.
	// This name should match the value in the SDN deployment's spec.volumes.configMap.name.
	SDNManagerConfigMapName = "sdn-controller-manager-config"
)

// ControllerConfig
type ControllerConfig struct {
	// SwitchSecretsPath is the location of the shared eAPI secret file
	SwitchSecretsPath string `json:"switchSecretsPath" yaml:"switchSecretsPath" yaml:"switchSecretsPath"`
	// SwitchBackendMode specifies which type of switch backend that will be used.
	SwitchBackendMode string `json:"switchBackendMode" yaml:"switchBackendMode"`
	// DataCenter is a list of data center names with ":" as the separator. e.g., "fxhb3p3s:fxhb3p3r"
	DataCenter string `json:"dataCenter" yaml:"dataCenter"`
	// PortResyncPeriodInSec specifies the frequency of the periodic SwitchPort reconciliation.
	PortResyncPeriodInSec int `json:"portResyncPeriodInSec" yaml:"portResyncPeriodInSec"`
	// NetworkNodeResyncPeriodInSec specifies the frequency of the periodic networkNode reconciliation.
	NetworkNodeResyncPeriodInSec int `json:"networkNodeResyncPeriodInSec" yaml:"networkNodeResyncPeriodInSec"`
	// BMHResyncPeriodInSec specifies the frequency of the periodic BareMetalHost reconciliation.
	BMHResyncPeriodInSec int `json:"bmhResyncPeriodInSec" yaml:"bmhResyncPeriodInSec"`
	// SwitchResyncPeriodInSec specifies the frequency of the periodic Switch reconciliation.
	SwitchResyncPeriodInSec int `json:"switchResyncPeriodInSec" yaml:"switchResyncPeriodInSec"`
	// NodeGroupResyncPeriodInSec specifies the frequency of the periodic NodeGroup reconciliation.
	NodeGroupResyncPeriodInSec int `json:"nodeGroupResyncPeriodInSec" yaml:"nodeGroupResyncPeriodInSec"`
	// MaxConcurrentReconciles specifies the number of threads of the SwitchPort reconciler.
	MaxConcurrentReconciles int `json:"maxConcurrentReconciles" yaml:"maxConcurrentReconciles"`
	// MaxConcurrentSwitchReconciles specifies the number of threads of the Switch reconciler.
	MaxConcurrentSwitchReconciles int `json:"maxConcurrentSwitchReconciles" yaml:"maxConcurrentSwitchReconciles"`
	// MaxConcurrentNetworkNodeReconciles specifies the number of threads of the NetworkNode reconciler.
	MaxConcurrentNetworkNodeReconciles int `json:"maxConcurrentNetworkNodeReconciles" yaml:"maxConcurrentNetworkNodeReconciles"`
	// MaxConcurrentNodeGroupReconciles specifies the number of threads of the NodeGroup reconciler.
	MaxConcurrentNodeGroupReconciles int `json:"maxConcurrentNodeGroupReconciles" yaml:"maxConcurrentNodeGroupReconciles"`
	// BMHClusterKubeConfigFilePath is the location of the BMH Cluster kubeConfig file, or a list of paths of kubeConfig files separated by ";".
	BMHClusterKubeConfigFilePath string `json:"bmhClusterKubeConfigFilePath" yaml:"bmhClusterKubeConfigFilePath"`
	// EnableReadOnlyMode specifies if read only mode is enabled or not.
	EnableReadOnlyMode bool `json:"enableReadOnlyMode" yaml:"enableReadOnlyMode"`
	// SwitchImportPeriodInSec specifies the interval of getting switches from Switches/Netbox and then try to create SwitchCR for it.
	SwitchImportPeriodInSec int `json:"switchImportPeriodInSec" yaml:"switchImportPeriodInSec"`
	// InitNetworkNodeWithNOOPVlanID. Deprecated
	InitNetworkNodeWithNOOPVlanID bool `json:"initNetworkNodeWithNOOPVlanID" yaml:"initNetworkNodeWithNOOPVlanID"`
	// PoolsConfigFilePath is the location of the pool configuration file
	PoolsConfigFilePath string `json:"poolsConfigFilePath" yaml:"poolsConfigFilePath"`
	// NodeGroupToPoolMappingSource specifies which source SDN Controller should get the Group to Pool mappings. options are "local" and "crd"
	NodeGroupToPoolMappingSource string `json:"nodeGroupToPoolMappingSource" yaml:"nodeGroupToPoolMappingSource"`
	// NodeGroupToPoolMappingConfigFilePath specifies the location of the NodeGroup to Pool mapping file.
	// This is used for the LocalPoolMapping, we don't need this if we store the mapping in Netbox.
	NodeGroupToPoolMappingConfigFilePath string `json:"nodeGroupToPoolMappingConfigFilePath" yaml:"nodeGroupToPoolMappingConfigFilePath"`
	// UseDefaultValueInPoolForMovingNodeGroup specifies if NOOP Vlan or BGP value should be used a NodeGroup is created or moved to a new Pool.
	UseDefaultValueInPoolForMovingNodeGroup bool `json:"useDefaultValueInPoolForMovingNodeGroup" yaml:"useDefaultValueInPoolForMovingNodeGroup"`
	// SwitchImportSource specifies where to import the Switch data. options: "netbox" and "none"
	SwitchImportSource string `json:"switchImportSource" yaml:"switchImportSource"`
	// SwitchPortImportSource specifies where to import the SwitchPort data. options: "netbox", "bmh" and "none"
	SwitchPortImportSource string `json:"switchPortImportSource" yaml:"switchPortImportSource"`
	// StatusReportPeriodInSec specifies the interval of getting switch config and updating the status of Switch and SwitchPort CRs.
	StatusReportPeriodInSec int `json:"statusReportPeriodInSec" yaml:"statusReportPeriodInSec"`
	// StatusReportAcceleratedPeriodInSec specifies the interval of getting switch config and updating the status of Switch and SwitchPort CRs for a period just after changes have been made.
	StatusReportAcceleratedPeriodInSec int `json:"statusReportAcceleratedPeriodInSec" yaml:"statusReportAcceleratedPeriodInSec"`

	// AllowedTrunkGroups lists the trunkGroups that can be added to a port. An empty list allows ALL trunkGroups to be set.
	AllowedTrunkGroups []string `json:"allowedTrunkGroups" yaml:"allowedTrunkGroups"`
	// AllowedModes lists the modes that can be added to a port.
	AllowedModes []string `json:"allowedModes" yaml:"allowedModes"`
	// AllowedVlanIdss lists the vlan ids that can be added to a port.
	AllowedVlanIds string `json:"allowedVlanIds" yaml:"allowedVlanIds"`
	// AllowedNativeVlanIds lists the native vlan ids that can be added to a port.
	AllowedNativeVlanIds string `json:"allowedNativeVlanIds" yaml:"allowedNativeVlanIds"`
	// BGPCommunityIncomingGroupName is the name of the BGP group that the accelerator leaves use in their "ip community-list <BGPCommunityIncomingGroupName> permit 101:X" config
	BGPCommunityIncomingGroupName string `json:"bgpCommunityIncomingGroupName" yaml:"bgpCommunityIncomingGroupName"`
	// PortChannelsEnabled enables support for managing port-channels on the switch. Typically disabled for Tenant SDN, enabled for Provider SDN.
	PortChannelsEnabled bool `json:"portChannelsEnabled" yaml:"portChannelsEnabled"`
	// ProvisioningVlan lists the vlans that can be set as provisioning vlan.
	ProvisioningVlanIds string `json:"provisioningVlanIds" yaml:"provisioningVlanIds"`
	// AllowedCountAccInterfaces is the number of accelerator interfaces a BMH must have to be considered valid & imported as a NetworkNode. Comma-separated string of allowed numbers. eg. "0,24" = 0 for those without accelerators, 24 for a Gaudi2 or 3, etc.
	AllowedCountAccInterfaces string `json:"allowedCountAccInterfaces" yaml:"allowedCountAccInterfaces"`

	/* Netbox related */
	// NetboxServer
	NetboxServer string `json:"netboxServer" yaml:"netboxServer"`
	// NetboxTokenPath
	NetboxTokenPath string `json:"netboxTokenPath" yaml:"netboxTokenPath"`
	// NetboxProviderServersFilterFilePath - the filters that define which switches the Netbox Controller maintains
	NetboxProviderServersFilterFilePath string `json:"netboxProviderServersFilterFilePath" yaml:"netboxProviderServersFilterFilePath"`
	// note: filtering the interface also require the netboxProviderServersFilterFilePath value, as we need to first find the devices(switches), and then find the interfaces.
	// NetboxProviderInterfacesFilterFilePath - the filters that define which switch ports/interfaces the Netbox Controller maintains
	NetboxProviderInterfacesFilterFilePath string `json:"netboxProviderInterfacesFilterFilePath" yaml:"netboxProviderInterfacesFilterFilePath"`
	// NetboxSwitchesFilterFilePath - the filters that define which switches the Netbox Controller maintains
	NetboxSwitchesFilterFilePath string `json:"netboxSwitchesFilterFilePath" yaml:"netboxSwitchesFilterFilePath"`
	// NetboxClientInsecureSkipVerify specifies the InsecureSkipVerify setting
	NetboxClientInsecureSkipVerify bool `json:"netboxClientInsecureSkipVerify" yaml:"netboxClientInsecureSkipVerify"`
	// NetboxSwitchFQDNDomainName will be appended to hostnames loaded from Netbox. eg. "internal-placeholder.com" or "us-staging-3.cloud.intel.com"
	NetboxSwitchFQDNDomainName string `json:"netboxSwitchFQDNDomainName" yaml:"netboxSwitchFQDNDomainName"`
}

type SDNControllerRestConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata"`

	RestConfig RestConfig `json:"restConfig" yaml:"restConfig"`
}

type RestConfig struct {
	ListenPort           int      `json:"listenPort" yaml:"listenPort"`
	HealthPort           int      `json:"healthPort" yaml:"healthPort"`
	DataCenter           string   `json:"dataCenter" yaml:"dataCenter"`
	AllowedTrunkGroups   []string `json:"allowedTrunkGroups" yaml:"allowedTrunkGroups"`
	PortChannelsEnabled  bool     `json:"portChannelsEnabled" yaml:"portChannelsEnabled"`
	AllowedVlanIds       string   `json:"allowedVlanIds" yaml:"allowedVlanIds"`
	AllowedNativeVlanIds string   `json:"allowedNativeVlanIds" yaml:"allowedNativeVlanIds"`
}

func init() {
	SchemeBuilder.Register(&SDNControllerConfig{})
}
