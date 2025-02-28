/*
Copyright 2024.

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

// BMEnrollmentOperatorConfig is the Schema for the bmEnrollmentoperatorconfigs API
type BMEnrollmentOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the configurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`
	// Compute API server Address. Format should be "compute-api-server:80"
	ComputeApiServerAddress string `json:"computeApiServerAddress"`
	// DHCP Proxy config. Specific to baremetal virtual stack.
	DhcpProxy DhcpProxyConfig `json:"dhcpProxy"`
	//Max concurrent reconcile value
	MaxConcurrentReconciles int `json:"maxConcurrentReconciles"`
	// Men and Mice Configuration
	MenAndMice MenAndMiceConfig `json:"menAndMice"`
	// Netbox configutation
	Netbox NetboxConfig `json:"netbox"`
	// IDC region
	Region string `json:"region"`
	// set BIOS password
	SetBiosPassword bool `json:"setBiosPassword"`
	// Vault server address
	VaultAddress string `json:"vaultAddress"`
}

type DhcpProxyConfig struct {
	// enable or disable dhcp proxy
	Enabled bool `json:"enabled"`
	// dhcp proxy url
	URL string `json:"url"`
}

type MenAndMiceConfig struct {
	// enable or disable menandmice configuration
	Enabled bool `json:"enabled"`
	// menandmice url
	URL string `json:"url"`
	// menandmice server address
	ServerAddress string `json:"serverAddress"`
	// skip TLS verification
	InsecureSkipVerify bool `json:"insecureSkipVerify"`
	// TFTP server IP
	TftpServerIp string `json:"tftpServerIp"`
}

type NetboxConfig struct {
	// netbox address
	Address string `json:"address"`
	// skip TLS verification
	SkipTlsVerify bool `json:"skipTlsVerify"`
}

func init() {
	SchemeBuilder.Register(&BMEnrollmentOperatorConfig{})
}
