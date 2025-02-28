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
	ControlplaneNodegroupType NodegroupType = "controlplane"
	WorkerNodegroupType       NodegroupType = "worker"
	UnknownNodegroupType      NodegroupType = "unknown"

	ActiveNodegroupState   NodegroupState = "Active"
	UpdatingNodegroupState NodegroupState = "Updating"
	ErrorNodegroupState    NodegroupState = "Error"
	DeletingNodegroupState NodegroupState = "Deleting"
)

type NodegroupType string

type NodegroupState string

type UpgradeStrategy struct {
	// MaxUnavailable specifies the allow number nodes unavailable that the node group supports.
	// +optional
	MaxUnavailablePercent int `json:"maxUnavailablePercent"`
	// DrainBefore specifies if containers must be moved out to other nodes before a node deletion.
	// +optional
	DrainBefore bool `json:"drainBefore"`
}

type VNET struct {
	AvailabilityZone     string `json:"availabilityzone"`
	NetworkInterfaceVnet string `json:"networkvnet"`
}

type WekaStorage struct {
	// Enable specifies if weka storage is enabled for the nodegroup.
	// +optional
	Enable bool `json:"enable"`
	// ClusterId specifies the weka cluster id.
	// +optional
	ClusterId string `json:"clusterId"`
	// Number of cores used to register nodes.
	// +kubebuilder:default:=1
	// +optional
	NumCores int32 `json:"numCores"`
	// Weka mode to use for registering nodes.
	// +kubebuilder:default:="udp"
	// +optional
	Mode string `json:"mode"`
}

// NodegroupSpec defines the desired state of Nodegroup
type NodegroupSpec struct {
	// Cloud account Id specifies cloud account id
	// +optional
	CloudAccountId string `json:"cloudaccountid"`
	// +optional
	Region string `json:"region"`
	// +optional
	VNETS []VNET `json:"vnets"`
	// KubernetesVersion specifies the version of k8s to use for installation.
	KubernetesVersion string `json:"kubernetesVersion"`
	// ClusterName specifies the name of the cluster that owns this nodegroup.
	ClusterName string `json:"clusterName"`
	// NodegroupType specifies if nodes in this nodegroup are of type controlplane or worker.
	NodegroupType NodegroupType `json:"nodegroupType"`
	// InstanceType specifies the compute instance to be used based on CPU / Memory and Storage requirements.
	InstanceType string `json:"instanceType"`
	// ClusterType specifies the type of the cluster created like supercompute or generalpurpose clusters.
	// +optional
	ClusterType string `json:"clusterType"`
	// InstanceIMI specifies the Intel machine instance to be used for nodes creation.
	InstanceIMI string `json:"instanceIMI"`
	// SSHKey specifies the ssh key name or id to be used during instances provisioning.
	// +optional
	SSHKey []string `json:"sshKey"`
	// Count specifies the number of nodes that should exist in this node group.
	Count int `json:"count"`
	// KubernetesProvider specifies the provider to use for communicating with the
	// kubernetes cluster.
	KubernetesProvider string `json:"kubernetesProvider"`
	// NodeProvider specifies the provider to use for communicating with the
	// instance provider.
	NodeProvider string `json:"nodeProvider"`
	// The IP of the etcd loadbalancer.
	EtcdLB string `json:"etcdLB"`
	// The IP of the kube-apiserver loadbalancer.
	APIServerLB string `json:"apiserverLB"`
	// The port of the etcd loadbalancer.
	EtcdLBPort string `json:"etcdLBPort"`
	// The port of the kube-apiserver loadbalancer.
	APIServerLBPort string `json:"apiserverLBPort"`
	// Runtime specifies the container runtime used in worker nodes.
	// +kubebuilder:default:="Containerd"
	// +optional
	ContainerRuntime string `json:"runtime"`
	// RuntimeArgs specifies custom arguments to run the container runtime.
	// +optional
	ContainerRuntimeArgs map[string]string `json:"runtimeArgs"`
	// This is an url to a bash script that will be downloaded and executed with cloud init
	// during node provisioning.
	// +optional
	UserDataURL string `json:"userDataURL"`
	// This holds the information to register and deregister nodes from the weka storage provider.
	// +optional
	WekaStorage WekaStorage `json:"wekaStorage"`
}

type WekaStorageStatus struct {
	// The weka client id used to identify registered node.
	// +optional
	ClientId string `json:"clientId"`
	// The registration status.
	// +optional
	Status string `json:"status"`
	// The custom registration status.
	// +optional
	CustomStatus string `json:"customStatus"`
	// The registration message.
	// +optional
	Message string `json:"message"`
}

type NodeStatus struct {
	// Name holds the hostname of the node.
	// +optional
	Name string `json:"name"`
	// IpAddress holds the ip address of the node.
	// +optional
	IpAddress string `json:"ipAddress"`
	// StorageBackendIP holds the ip address of the storage backend node.
	// +optional
	StorageBackendIP string `json:"storageBackendIP"`
	// StorageBackendSubnet holds the subnet of the storage backend node.
	// +optional
	StorageBackendSubnet string `json:"storageBackendSubnet"`
	// StorageBackendGateway holds the gateway of the storage backend node.
	// +optional
	StorageBackendGateway string `json:"storageBackendGateway"`
	// InstanceIMI specifies the Intel machine instance used for creating the node.
	// +optional
	InstanceIMI string `json:"instanceIMI"`
	// Kubelet version running on the node.
	// +optional
	KubeletVersion string `json:"kubeletVersion"`
	// Kube-proxy version running on the node.
	// +optional
	KubeProxyVersion string `json:"kubeProxyVersion"`
	// CreationTime specfies the time where machine became a kubernetes node.
	// +optional
	CreationTime metav1.Time `json:"creationTime"`
	// State of the node, one of active, updating, error, deleting.
	State NodegroupState `json:"state"`
	// LastUpdate specifies when the state of the node last updated.
	// +optional
	LastUpdate metav1.Time `json:"lastUpdate"`
	// Reason is a description of the current state.
	// +optional
	Reason string `json:"reason"`
	// Message is a more verbose description of the current state.
	// +optional
	Message string `json:"message"`
	// Domain name of the instance.
	// +optional
	DNSName string `json:"dnsName"`
	// Unschedulable specifies if node can be used to schedule pods.
	// +optional
	Unschedulable bool `json:"unschedulable"`
	// AutoRepairDisabled specifies if node needs to be autorepaired or not.
	// +optional
	AutoRepairDisabled bool `json:"autoRepairDisabled"`
	// Specifies the gateway of the node.
	// +optional
	Gateway string `json:"gateway"`
	// Specifies the subnet of the node.
	// +optional
	Subnet string `json:"subnet"`
	// Specifies the netmask of the node.
	// +optional
	Netmask int32 `json:"netmask"`
	// Specifies the status of the weka registration action for the node.
	// +optional
	WekaStorageStatus WekaStorageStatus `json:"wekaStorage"`
}

// NodegroupStatus defines the observed state of Nodegroup
type NodegroupStatus struct {
	// Name holds the name of the nodegroup.
	// +optional
	Name string `json:"name"`
	// Count holds the current number of nodes in the nodegroup.
	// +optional
	Count int `json:"count"`
	// NodegroupType holds the type of nodegroup, one of controlplane or worker.
	// +optional
	Type NodegroupType `json:"type"`
	// State of the nodegroup, one of active, updating, error, deleting.
	// +optional
	State NodegroupState `json:"state"`
	// Nodes holds the status of the nodes.
	// +optional
	Nodes []NodeStatus `json:"nodes"`
	// Reason is a description of the current state.
	// +optional
	Reason string `json:"reason"`
	// Message is a more verbose description of the current state.
	// +optional
	Message string `json:"message"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string
//+kubebuilder:printcolumn:JSONPath=".status.type",name=Type,type=string
//+kubebuilder:printcolumn:JSONPath=".status.count",name=Count,type=integer
//+kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Nodegroup is the Schema for the nodegroups API
type Nodegroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodegroupSpec   `json:"spec,omitempty"`
	Status NodegroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodegroupList contains a list of Nodegroup
type NodegroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Nodegroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Nodegroup{}, &NodegroupList{})
}
