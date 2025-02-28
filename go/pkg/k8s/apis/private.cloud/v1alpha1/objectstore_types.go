package v1alpha1

import (
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Object store spec defines the desired state of object store
type ObjectStoreSpec struct {

	// AZ not used right now
	AvailabilityZone string `json:"availabilityZone"`

	Versioned                 bool                      `json:"versioned"`
	Quota                     string                    `json:"quota"`
	BucketAccessPolicy        BucketAccessPolicy        `json:"bucketAccessPolicy,omitempty"`
	ObjectStoreBucketSchedule ObjectStoreBucketSchedule `json:"objectStoreBucketSchedule"`
}

type ObjectStoreBucketSchedule struct {
	ObjectStoreCluster ObjectStoreAssignedCluster `json:"objectStoreAssignedCluster"`
}

type ObjectStoreAssignedCluster struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
	Addr string `json:"addr"`
}

type ObjectStoreBucket struct {
	Name     string         `json:"name"`
	Id       string         `json:"id,omitempty"`
	Capacity BucketCapacity `json:"capacity,omitempty"`
}

type BucketCapacity struct {
	TotalBytes     string `json:"totalBytes"`
	AvailableBytes string `json:"AvailableBytes,omitempty"`
}

type BucketAccessPolicy int32

const (
	BucketAccessPolicyReadWrite   BucketAccessPolicy = 3
	BucketAccessPolicyReadOnly    BucketAccessPolicy = 2
	BucketAccessPolicyUnspecified BucketAccessPolicy = 1
	BucketAccessPolicyNone        BucketAccessPolicy = 0
)

// ObjectStoreStatus defines the observed state of the object Store
type ObjectStoreStatus struct {
	Phase      ObjectStorePhase       `json:"phase,omitempty"`
	Message    string                 `json:"message,omitempty"`
	Bucket     ObjectStoreBucket      `json:"bucket,omitempty"`
	Conditions []ObjectStoreCondition `json:"conditions,omitempty"`
}

type ObjectStoreCondition struct {
	Type   ObjectStoreConditionType `json:"type"`
	Status k8sv1.ConditionStatus    `json:"status"`
	// +nullable
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// +nullable
	LastTransitionTime metav1.Time                `json:"lastTransitionTime,omitempty"`
	Reason             ObjectStoreConditionReason `json:"reason,omitempty"`
	Message            string                     `json:"message,omitempty"`
}

type ObjectStoreConditionType string

// These are the valid conditions of a Object store.
const (
	// The Storage Controller has created or updated all output objects.
	ObjectStoreConditionAccepted ObjectStoreConditionType = "Accepted"

	// The ObjectStore bucket instance is running.
	ObjectStoreConditionRunning ObjectStoreConditionType = "Running"

	// The ObjectStore Bucket failed.
	ObjectStoreConditionFailed ObjectStoreConditionType = "Failed"
)

type ObjectStoreConditionReason string

const (
	ObjectStoreConditionReasonNotAccepted ObjectStoreConditionReason = "NotAccepted"
	ObjectStoreConditionReasonAccepted    ObjectStoreConditionReason = "Accepted"
)

type ObjectStorePhase string

// These are the valid phases of Object store bucket.
// The phase is always calculated from the conditions. It is never used to determine the state.
const (
	// PhaseProvisioning means the ObjectStore has started working on the request.
	ObjectStorePhasePhaseProvisioning ObjectStorePhase = "Provisioning"
	// This corresponds to the Ready condition of ObjectStore
	ObjectStorePhasePhaseReady ObjectStorePhase = "Ready"

	// PhaseTerminating means the ObjectStore and its associated resources are in the process of being deleted.
	ObjectStorePhasePhaseTerminating ObjectStorePhase = "Terminating"
	// PhaseFailed means that the ObjectStore crashed, deleted etc.
	ObjectStorePhasePhaseFailed ObjectStorePhase = "Failed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+genclient

// ObjectStore is the Schema for the Object Storages API
type ObjectStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ObjectStoreSpec   `json:"spec,omitempty"`
	Status ObjectStoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ObjectStorageList contains a list of ObjectStore
type ObjectStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObjectStore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ObjectStore{}, &ObjectStoreList{})
}
