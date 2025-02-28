// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StorageSpec defines the desired state of Storage FS
type StorageSpec struct {

	// Messages
	AvailabilityZone string                   `json:"availabilityZone"`
	StorageRequest   FilesystemStorageRequest `json:"request"`
	StorageClass     FilesystemStorageClass   `json:"storageClass,omitempty"`
	AccessModes      FilesystemAccessModes    `json:"accessModes"`
	MountProtocol    FilesystemMountProtocol  `json:"mountProtocol"`
	Encrypted        bool                     `json:"encrypted"`
	ProviderSchedule FilesystemSchedule       `json:"filesystemSchedule"`
	FilesystemType   FilesystemType           `json:"filesystemType"`
	Prefix           string                   `json:"prefix"`
}

type FilesystemSchedule struct {
	FilesystemName string            `json:"filesystemName"`
	Cluster        AssignedCluster   `json:"assignedCluster"`
	Namespace      AssignedNamespace `json:"assignedNamespace"`
	User           AssignedUser      `json:"assignedUser"`
}

type AssignedCluster struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	Addr    string `json:"addr"`
	Version string `json:"version"`
}

type AssignedNamespace struct {
	Name            string `json:"name"`
	CredentialsPath string `json:"credentialsPath"`
	User            string `json:"user"`
	Password        string `json:"password"`
}

type AssignedUser struct {
	CredentialsPath string `json:"credentialsPath"`
	User            string `json:"user"`
	Password        string `json:"password"`
}

type FilesystemStorageClass string

const (
	FilesystemStorageClassDefault        FilesystemStorageClass = "Default"
	FilesystemStorageClassGeneralPurpose FilesystemStorageClass = "GeneralPurpose"
	FilesystemStorageClassAIOptimized    FilesystemStorageClass = "AIOptimized"
)

type FilesystemType string

const (
	FilesystemTypeComputeGeneral    FilesystemType = "ComputeGeneral"
	FilesystemTypeComputeKubernetes FilesystemType = "ComputeKubernetes"
)

type FilesystemAccessModes string

const (
	FilesystemAccessModesReadWrite     FilesystemAccessModes = "ReadWrite"
	FilesystemAccessModesReadOnly      FilesystemAccessModes = "ReadOnly"
	FilesystemAccessModesReadWriteOnce FilesystemAccessModes = "ReadWriteOnce"
)

type FilesystemMountProtocol string

const (
	FilesystemMountProtocolWeka FilesystemMountProtocol = "Weka"
	FilesystemMountProtocolNFS  FilesystemMountProtocol = "NFS"
)

type FilesystemStorageRequest struct {
	// size
	Size string `json:"storage"`
}

// StorageStatus defines the observed state of Storage FS
type StorageStatus struct {
	Phase      FilesystemPhase       `json:"phase,omitempty"`
	Message    string                `json:"message,omitempty"`
	Mount      FilesystemMountStatus `json:"mount,omitempty"`
	Namespace  FilesystemNamespace   `json:"namespace,omitempty"`
	User       FilesystemUserStatus  `json:"user,omitempty"`
	Conditions []StorageCondition    `json:"conditions,omitempty"`
	Size       string                `json:"size,omitempty"`
}

type StorageCondition struct {
	Type   StorageConditionType  `json:"type"`
	Status k8sv1.ConditionStatus `json:"status"`
	// +nullable
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// +nullable
	LastTransitionTime metav1.Time            `json:"lastTransitionTime,omitempty"`
	Reason             StorageConditionReason `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
}

type StorageConditionType string

// These are the valid conditions of a Storage FS.
const (
	// The Storage Controller has created or updated all output objects.
	StorageConditionAccepted StorageConditionType = "Accepted"

	// The Storage FS instance is running.
	StorageConditionRunning StorageConditionType = "Running"

	// The Storage FS failed.
	StorageConditionFailed StorageConditionType = "Failed"

	// The Storage NS Success.
	StorageConditionNamespaceSuccess StorageConditionType = "Namespace Success"

	// The Storage FS success.
	StorageConditionFSSuccess StorageConditionType = "FileSystem Success"
	// The Storage in Deleting
	StorageConditionDeleting StorageConditionType = "Deleting"
	// The Storage IKS Path success.
	StorageConditionNamespaceK8sSuccess StorageConditionType = "Namespace only Path Success"
	StorageConditionUpdateFSSuccess     StorageConditionType = "Update FileSystem Success"
)

type StorageConditionReason string

const (
	StorageConditionReasonNone        StorageConditionReason = ""
	StorageConditionReasonNotAccepted StorageConditionReason = "NotAccepted"
	StorageConditionReasonAccepted    StorageConditionReason = "Accepted"
)

// These are the prefixes of StorageStatus.Message. The field may have additional details.
const (
	StorageMessageNew                     string = "Storage reconciliation has not started"
	StorageMessageProvisioningNotAccepted string = "Storage specification has not been accepted and is being provisioned"
	StorageMessageProvisioningAccepted    string = "Storage specification has been accepted and is being provisioned"
	StorageMessageRunning                 string = "Storage FS is running "
	StorageMessageFailed                  string = "The Storage FS is unavailable"
	StorageMessageDeleting                string = "The Storage FS in the process of being deleted"
)

type FilesystemMountStatus struct {
	FileSystemName string `json:"fileSystemName"`
	ClusterAddr    string `json:"clusterAddr"`
}

type FilesystemUserStatus struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type FilesystemNamespace struct {
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type FilesystemPhase string

// These are the valid phases of instances.
// The phase is always calculated from the conditions. It is never used to determine the state.
const (
	// PhaseProvisioning means the Storage FS has started working on the request.
	FilesystemPhaseProvisioning FilesystemPhase = "Provisioning"
	// This corresponds to the Ready condition of Storage FS.
	FilesystemPhaseReady FilesystemPhase = "Ready"

	// PhaseDeleting means the Storage FS and its associated resources are in the process of being deleted.
	FilesystemPhaseDeleting FilesystemPhase = "Deleting"
	// PhaseFailed means that the Storage FS crashed, deleted etc.
	FilesystemPhaseFailed FilesystemPhase = "Failed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+genclient

// Storage is the Schema for the Storages API
type Storage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageSpec   `json:"spec,omitempty"`
	Status StorageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StorageList contains a list of Storage
type StorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Storage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Storage{}, &StorageList{})
}
