// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// SshProxyOperatorConfig is the Schema for the sshproxyoperatorconfigs API.
// It stores the configuration for the SSH Proxy Operator.
type SshProxyOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the configurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	// Local path of the authorized_keys file that will be written.
	// If this operator will only SCP the file, this can be a temporary path.
	// Example: /tmp/.ssh/authorized_keys
	AuthorizedKeysFilePath string `json:"authorizedKeysFilePath,omitempty"`

	// User name that end user must use to connect to the SSH proxy server.
	// This is typically guest.
	ProxyUser string `json:"proxyUser,omitempty"`

	// FQDN or IP address that end user must use to connect to the SSH proxy server.
	// This is expected to be a load balancer.
	ProxyAddress string `json:"proxyAddress,omitempty"`

	// Port number that end user must use to connect to the SSH proxy server.
	// This is typically 22.
	ProxyPort int `json:"proxyPort,omitempty"`

	// A list of SCP URIs that the authorized_keys file will be copied to.
	// Format is: scp://guest@ssh-proxy-server-1.cloud.intel.com:22/home/guest/.ssh/authorized_keys
	AuthorizedKeysScpTargets []string `json:"authorizedKeysScpTargets,omitempty"`

	// Path to the SSH public key used by this operator to connect to the SCP target servers.
	// Example: /home/idc/.ssh/id_rsa.pub
	PublicKeyFilePath string `json:"publicKeyFilePath,omitempty"`

	// Path to the SSH private key used by this operator to connect to the SCP target servers.
	// Example: /home/idc/.ssh/id_rsa
	PrivateKeyFilePath string `json:"privateKeyFilePath,omitempty"`

	// Path to the ssh proxy server host public key used for verifying the SSH proxy server
	HostPublicKeyFilePath string `json:"hostPublicKeyFilePath"`
}

func init() {
	SchemeBuilder.Register(&SshProxyOperatorConfig{})
}
