// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SshProxyTunnelSpec defines the desired state of SshProxyTunnel
type SshProxyTunnelSpec struct {

	// TargetAddresses specifies the list private ip of the instances which needs to be accessed from outside by the tenant
	TargetAddresses []string `json:"targetAddresses"`

	// TargetPorts specifies the list of ports at which the instance can be accessed
	TargetPorts []int `json:"targetPorts"`

	// SshPublicKeys specifies the list of ssh public keys of the tenant (id_rsa.pub) trying to access the instances
	SshPublicKeys []string `json:"sshPublicKeys"`
}

// SshProxyTunnelStatus defines the observed state of SshProxyTunnel
type SshProxyTunnelStatus struct {

	// ProxyUser specifies the user the tenant will use to access the proxy server
	ProxyUser string `json:"proxyUser"`

	// ProxyAddress specifies the proxy server where tenant ssh request will come to
	ProxyAddress string `json:"proxyAddress"`

	// ProxyPort specifies the port at which the proxy server will listen for incoming ssh connection from tenant
	ProxyPort int `json:"proxyPort"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+genclient

// SshProxyTunnel is the Schema for the sshproxytunnels API
type SshProxyTunnel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SshProxyTunnelSpec   `json:"spec,omitempty"`
	Status SshProxyTunnelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SshProxyTunnelList contains a list of SshProxyTunnel
type SshProxyTunnelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SshProxyTunnel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SshProxyTunnel{}, &SshProxyTunnelList{})
}
