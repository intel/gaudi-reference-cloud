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
	fwv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	ilbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ilb_operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ActiveClusterState   ClusterState = "Active"
	UpdatingClusterState ClusterState = "Updating"
	ErrorClusterState    ClusterState = "Error"
	DeletingClusterState ClusterState = "Deleting"

	ActiveStorageState   StorageState = "Active"
	UpdatingStorageState StorageState = "Updating"
	ErrorStorageState    StorageState = "Error"
	DeletingStorageState StorageState = "Deleting"

	ReconcileFWState FirewallState = "Reconciling"
	ReadyFWState     FirewallState = "Ready"
	TerminateFWState FirewallState = "Terminated"
)

type ClusterState string

type StorageState string

type FirewallState string

type Network struct {
	// ServiceCIDR specifies the IP address range for k8s services.
	// +kubebuilder:default:="100.66.0.0/16"
	// +optional
	ServiceCIDR string `json:"serviceCIDR"`
	// PodCIDR specifies the IP address range for k8s pods.
	// +kubebuilder:default:="100.68.0.0/16"
	// +optional
	PodCIDR string `json:"podCIDR"`
	// ClusterDNS specifies the IP to use for the coredns service.
	// +kubebuilder:default:="100.66.0.10"
	// +optional
	ClusterDNS string `json:"clusterDNS"`
	// Region is cluster specific region
	// +optional
	Region string `json:"region"`
}

type AdvancedConfig struct {
	// KubeApiServerArgs specifies custom arguments for running the kube-apiserver component.
	// +optional
	KubeApiServerArgs string `json:"kubeApiServerArgs"`
	// KubeControllerManagerArgs specifies custom arguments for running the kube-controller-manager component.
	// +optional
	KubeControllerManagerArgs string `json:"kubeControllerManagerArgs"`
	// KubeSchedulerArgs specifies custom arguments for running the kube-scheduler component.
	// +optional
	KubeSchedulerArgs string `json:"kubeSchedulerArgs"`
	// KubeProxyArgs specifies custom arguments for running the kube-proxy component.
	// +optional
	KubeProxyArgs string `json:"kubeProxyArgs"`
	// KubeletArgs specifies custom arguments for running the kubelet component.
	// +optional
	KubeletArgs string `json:"kubeletArgs"`
}

// Storage specifies the configuration for the storage provider.
type Storage struct {
	// Provider specifies the storage provider to use for the cluster.
	// +optional
	Provider string `json:"provider"`
	// Size of the storage to be requested for the cluster.
	// +optional
	Size string `json:"size"`
	// Number of cores used to register nodes.
	// +kubebuilder:default:=1
	// +optional
	NumCores int32 `json:"numCores"`
	// Weka mode to use for registering nodes.
	// +kubebuilder:default:="udp"
	// +optional
	Mode string `json:"mode"`
}

// EtcdBackup specifies the configuration to take etcd backups.
type EtcdBackupConfig struct {
	// RetentionPolicy specifies for how long the backups will be kept.
	// +optional
	RetentionPolicy string `json:"retentionPolicy"`
	// Periodicity specifies how often to create backups.
	// +optional
	Periodicity string `json:"periodicity"`
	// S3BackupFolder specifies the local folder for temporary storing the backup.
	// +optional
	S3BackupFolder string `json:"s3BackupFolder"`
	// S3URL specifies the S3 URL.
	// +optional
	S3URL string `json:"s3URL"`
	// S3AccessKey specifies the S3 access key.
	// +optional
	S3AccessKey string `json:"s3AccessKey"`
	// S3SecretKey specifies the S3 secret key.
	// +optional
	S3SecretKey string `json:"s3SecretKey"`
	// S3UseSSL specifies if client connection needs to be secure.
	// +optional
	S3UseSSL bool `json:"s3UseSSL"`
	// S3BucketName specifies the name of the bucket where backups will be stored.
	// +optional
	S3BucketName string `json:"s3BucketName"`
	// S3ContentType specifies the content type for putting objects into S3.
	// +optional
	S3ContentType string `json:"s3ContentType"`
	// S3Path specifies the folder within the S3 bucket where backups will be stored.
	// +optional
	S3Path string `json:"s3Path"`
}

// NodegroupTemplateSpec defines the spec of a nodegroup.
type NodegroupTemplateSpec struct {
	// Name specifies the suffix used to name the nodegroup.
	Name string `json:"name"`
	// Runtime specifies the container runtime used in worker nodes.
	// +kubebuilder:default:="Containerd"
	// +optional
	ContainerRuntime string `json:"runtime"`
	// RuntimeArgs specifies custom arguments to run the container runtime.
	// +optional
	ContainerRuntimeArgs map[string]string `json:"runtimeArgs"`
	// KubernetesVersion specifies the version of k8s to use for installation.
	// +optional
	KubernetesVersion string `json:"kubernetesVersion"`
	// InstanceType specifies the compute instance to be used based on CPU / Memory and Storage requirements.
	InstanceType string `json:"instanceType"`
	// ClusterType specifies the type of the cluster created like supercompute or generalpurpose clusters.
	ClusterType string `json:"clusterType"`
	// InstanceIMI specifies the Intel machine instance to be used for nodes creation.
	InstanceIMI string `json:"instanceIMI"`
	// SSHKey specifies a list of ssh key name or id to be used during instances provisioning.
	// +optional
	SSHKey []string `json:"sshKey"`
	// Count specifies the number of nodes that should exist in this node group.
	Count int `json:"count"`
	// UpgradeStrategy specifies the maximum number of unavailable nodes and node drain enforcement during upgrades.
	// +optional
	UpgradeStrategy UpgradeStrategy `json:"upgradeStrategy"`
	// Taints specifies node group taints.
	// +optional
	Taints map[string]string `json:"taints"`
	// Labels specifies node group labels.
	// +optional
	Labels map[string]string `json:"labels"`
	// Annotations specifies the annotations that must be added to the nodegroup.
	// +optional
	Annotations map[string]string `json:"annotations"`
	// Virtual net configuration for compute instances.
	VNETS []VNET `json:"vnets"`
	// Cloudaccount ID to be used for compute instances.
	CloudAccountId string `json:"cloudaccountid"`
	// This is an url to a bash script that will be downloaded and executed with cloud init
	// during node provisioning.
	// +optional
	UserDataURL string `json:"userDataURL"`
}

type AddonTemplateSpec struct {
	// Name specifies the name of the provider.
	Name string `json:"name"`
	// Type specifies the provider and action to use for puting the addon into the cluster.
	Type AddonType `json:"type"`
	// Artifact specifies the url of the manifest that will be installed.
	Artifact string `json:"artifact"`
}

// ILBTemplateSpec defines the loadbalancers that need to be created for the cluster.
type ILBTemplateSpec struct {
	Name string `json:"name"`
	// +optional
	Description string `json:"description"`
	Port        int    `json:"port"`
	// +kubebuilder:default:="private"
	IPType string `json:"iptype"`
	// +kubebuilder:default:=""
	Persist string `json:"persist"`
	// +kubebuilder:default:="tcp"
	IPProtocol  string              `json:"ipprotocol"`
	Environment int                 `json:"environment"`
	Usergroup   int                 `json:"usergroup"`
	Pool        ILBPoolTemplateSpec `json:"pool"`
	Owner       string              `json:"owner"`
}

type ILBPoolTemplateSpec struct {
	Name string `json:"name"`
	// +optional
	Description string `json:"description"`
	Port        int    `json:"port"`
	// +kubebuilder:default:="least-connections-member"
	LoadBalancingMode string `json:"loadBalancingMode"`
	// +kubebuilder:default:=1
	MinActiveMembers int `json:"minActiveMembers"`
	// +kubebuilder:default:="i_tcp"
	Monitor string `json:"monitor"`
	// +kubebuilder:default:=0
	MemberConnectionLimit int `json:"memberConnectionLimit"`
	// +kubebuilder:default:=0
	MemberPriorityGroup int `json:"memberPriorityGroup"`
	// +kubebuilder:default:=1
	MemberRatio int `json:"memberRatio"`
	// +kubebuilder:default:="enabled"
	MemberAdminState string `json:"memberAdminState"`
}

type FirewallSpec struct {
	// Destination ip
	DestinationIp string   `json:"destinationIp"`
	Port          int      `json:"port"`
	Protocol      string   `json:"protocol"`
	SourceIps     []string `json:"sourceips"`
}

// ClusterSpec defines the desired state of Cluster.
type ClusterSpec struct {
	// KubernetesVersion specifies the version of k8s to use for installation.
	KubernetesVersion string `json:"kubernetesVersion"`
	// InstanceType specifies the compute instance to be used based on CPU / Memory and Storage requirements.
	InstanceType string `json:"instanceType"`
	// ClusterType specifies the type of the cluster created like supercompute or generalpurpose clusters.
	// +optional
	ClusterType string `json:"clusterType"`
	// InstanceIMI specifies the Intel machine instance to be used for nodes creation.
	InstanceIMI string `json:"instanceIMI"`
	// SSHKey specifies a list of ssh key name or id to be used during instances provisioning.
	// +optional
	SSHKey []string `json:"sshKey"`
	// Runtime specifies the container runtime used in worker nodes.
	// +kubebuilder:default:="Containerd"
	// +optional
	ContainerRuntime string `json:"runtime"`
	// RuntimeArgs specifies custom arguments to run the container runtime.
	// +optional
	ContainerRuntimeArgs map[string]string `json:"runtimeArgs"`
	// KubernetesProvider specifies the provider to use for provisoning the cluster.
	KubernetesProvider string `json:"kubernetesProvider"`
	// KubernetesProviderConfig specifies custom configuration used by the kubernetes provider.
	// +optional
	KubernetesProviderConfig map[string]string `json:"kubernetesProviderConfig"`
	// NodeProvider specifies the provider to use for provisioning the instances that will become nodes
	// in the cluster.
	NodeProvider string `json:"nodeProvider"`
	// Network specifies the container network configuration for the k8s cluster.
	// +kubebuilder:default:={serviceCIDR: "100.66.0.0/16", podCIDR: "100.68.0.0/16", clusterDNS: "100.66.0.10"}
	// +optional
	Network Network `json:"network"`
	// Nodegroups specifies a list of nodegroups.
	// +optional
	Nodegroups []NodegroupTemplateSpec `json:"nodegroups"`
	// Addons specifies the list of k8s addons that will be deployed in the k8s cluster.
	// +optional
	Addons []AddonTemplateSpec `json:"addons"`
	// Specifies the list of loadbalancers to create.
	// +optional
	ILBS []ILBTemplateSpec `json:"ilbs"`
	// EtcdBackupEnabled specifies if etcd backups should be taken.
	// +optional
	EtcdBackupEnabled bool `json:"etcdBackupEnabled"`
	// EtcdBackup specifies the configuration to take etcd backups.
	// +optional
	EtcdBackupConfig EtcdBackupConfig `json:"etcdBackupConfig"`
	// Advance configuration for k8s components.
	// +optional
	AdvancedConfig AdvancedConfig `json:"advancedConfig"`
	// Virtual net configuration for compute instances.
	VNETS []VNET `json:"vnets"`
	// Controlplane cloudaccount ID.
	CloudAccountId string `json:"cloudaccountid"`
	// Storage specifies the list of storage providers that need to be
	// configured in the cluster.
	// +optional
	Storage []Storage `json:"storage"`
	// Customer cloudaccount ID.
	// +optional
	CustomerCloudAccountId string `json:"customerCloudaccountid"`
	// Firewall sepecifies configuration for firewall.
	// +optional
	Firewall []FirewallSpec `json:"firewall"`
}

type StorageStatus struct {
	// Provider specifies the storage provider to use for the cluster.
	// +optional
	Provider string `json:"provider"`
	// Size of the storage to be requested for the cluster.
	// +optional
	Size string `json:"size"`
	// NamespaceCreated specifies if the storage has been created.
	// +optional
	NamespaceCreated bool `json:"namespaceCreated"`
	// NamespaceName specifies the namespace where the storage is created.
	// +optional
	NamespaceName string `json:"namespaceName"`
	// NamespaceState specifies the state of the created storage.
	// +optional
	NamespaceState StorageState `json:"namespaceState"`
	// General state of the storage, one of Active, Updating, deleting or Error.
	// +optional
	State StorageState `json:"state"`
	// LastUpdate specifies when the state of the storage last updated.
	// +optional
	LastUpdate metav1.Time `json:"lastUpdate"`
	// Reason is a description of the current state.
	// +optional
	Reason string `json:"reason"`
	// Message is a more verbose description of the current state.
	// +optional
	Message string `json:"message"`
	// Specifies if the storage class secret has been created in the downstream cluster.
	// +optional
	SecretCreated bool `json:"secretCreated"`
	// ClusterId is the id of the weka cluster.
	// +optional
	ClusterId string `json:"clusterId"`
	// ActiveAt specifies the time when the storage was created and seen as Active for the first time.
	// This is used to avoid the recreation of the storage if it was Active once.
	// +optional
	ActiveAt metav1.Time `json:"activeAt"`
	// CreatedAt specifies the time when the storage was created.
	// +optional
	CreatedAt metav1.Time `json:"createdAt"`
}

// Source ips status
type FirewallStatus struct {
	// +optional
	Firewallrulestatus fwv1alpha1.FirewallRuleStatus `json:"firewallrulestatus"`
	// +optional
	DestinationIp string `json:"destinationIp"`
	// +optional
	Port int `json:"port"`
	// +optional
	Protocol string `json:"protocol"`
	// +optional
	SourceIps []string `json:"sourceips"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// State of the cluster, one of Active, Updating, Deleting or Error.
	// +optional
	State ClusterState `json:"state"`
	// LastUpdate specifies when the state of the cluster last updated.
	// +optional
	LastUpdate metav1.Time `json:"lastUpdate"`
	// Reason is a description of the current state.
	// +optional
	Reason string `json:"reason"`
	// Message is a more verbose description of the current state.
	// +optional
	Message string `json:"message"`
	// List of nodegroups and nodes.
	// +optional
	Nodegroups []NodegroupStatus `json:"nodegroups"`
	// List of addons.
	// +optional
	Addons []AddonStatus `json:"addons"`
	// List of loadbalancers.
	// +optional
	ILBS []ilbv1alpha1.IlbStatus `json:"ilbs"`
	// List of storage providers.
	// +optional
	Storage []StorageStatus `json:"storage"`
	// List of source ips.
	// +optional
	Firewall []FirewallStatus `json:"firewall"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{}, &ilbv1alpha1.Ilb{}, &ilbv1alpha1.IlbList{}, &fwv1alpha1.FirewallRule{}, &fwv1alpha1.FirewallRuleList{})
}
