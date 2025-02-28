// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VastStorageSpec defines the desired state of VAST Storage FS
type VastStorageSpec struct {

	// Messages
	AvailabilityZone  string                       `json:"availabilityZone"`
	FilesystemName    string                       `json:"filesystemName"`
	FilesystemType    FilesystemType               `json:"filesystemType"`
	CSIVolumePrefix   string                       `json:"csiVolumePrefix"`
	StorageRequest    VASTFilesystemStorageRequest `json:"request"`
	StorageClass      FilesystemStorageClass       `json:"storageClass,omitempty"`
	ClusterAssignment ClusterAssignment            `json:"clusterAssignment"`
	MountConfig       MountConfig                  `json:"mountConfig"`
	Networks          Networks                     `json:"networks"`
}

type MountConfig struct {
	VolumePath    string                      `json:"volumePath"`
	MountProtocol VastFilesystemMountProtocol `json:"mountProtocol"`
}

type VASTFilesystemStorageRequest struct {
	// size
	Size string `json:"storage"`
}

type SecurityGroups struct {
	IPFilters []IPFilter `json:"ipFilters"`
}

type Networks struct {
	SecurityGroups SecurityGroups `json:"securityGroups"`
}

type ClusterAssignment struct {
	ClusterUUID    string `json:"clusterUUID"`
	ClusterVersion string `json:"clusterVersion"`
	NamespaceName  string `json:"namespaceName"`
}

type IPFilter struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
type VastFilesystemMountProtocol string

const (
	FilesystemMountProtocolSMB   VastFilesystemMountProtocol = "PROTOCOL_SMB"
	FilesystemMountProtocolNFSV3 VastFilesystemMountProtocol = "PROTOCOL_NFS_V3"
	FilesystemMountProtocolNFSV4 VastFilesystemMountProtocol = "PROTOCOL_NFS_V4"
)

type VolumeProperties struct {
	Size         string `json:"size,omitempty"`
	NamespaceId  int64  `json:"namespaceId"`
	FilesystemId int64  `json:"filesystemId"`
}

// StorageStatus defines the observed state of Storage FS
type VastStorageStatus struct {
	Phase       FilesystemPhase    `json:"phase,omitempty"`
	Message     string             `json:"message,omitempty"`
	VolumeProps VolumeProperties   `json:"volumeProps,omitempty"`
	Conditions  []StorageCondition `json:"conditions,omitempty"`
}

// These are the prefixes of StorageStatus.Message. The field may have additional details.
const (
	VastStorageMessageNew                     string = "Vast reconciliation has not started"
	VastStorageMessageProvisioningNotAccepted string = "VastStorage specification has not been accepted and is being provisioned"
	VastStorageMessageProvisioningAccepted    string = "VastStorage specification has been accepted and is being provisioned"
	VastStorageMessageRunning                 string = "VastStorage FS is running "
	VastStorageMessageFailed                  string = "The VastStorage FS is unavailable"
	VastStorageMessageDeleting                string = "The VastStorage FS in the process of being deleted"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+genclient

// VastStorage is the Schema for the VastStorages API
type VastStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VastStorageSpec   `json:"spec,omitempty"`
	Status VastStorageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VastStorageList contains a list of Storage
type VastStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VastStorage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VastStorage{}, &VastStorageList{})
}
