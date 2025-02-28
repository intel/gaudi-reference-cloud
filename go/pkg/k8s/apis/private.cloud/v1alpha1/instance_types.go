// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstanceSpec defines the desired state of Instance.
// JSON representation should match InstanceSpecPrivate in /public_api/proto/compute.proto
// unless there is special handling in go/pkg/instance_replicator/convert/convert.go.
type InstanceSpec struct {
	AvailabilityZone          string                      `json:"availabilityZone"`
	InstanceType              string                      `json:"instanceType,omitempty"`
	MachineImage              string                      `json:"machineImage,omitempty"`
	RunStrategy               RunStrategy                 `json:"runStrategy,omitempty"`
	SshPublicKeyNames         []string                    `json:"sshPublicKeyNames,omitempty"`
	Interfaces                []InterfaceSpec             `json:"interfaces,omitempty"`
	InstanceTypeSpec          InstanceTypeSpec            `json:"instanceTypeSpec"`
	MachineImageSpec          MachineImageSpec            `json:"machineImageSpec"`
	SshPublicKeySpecs         []SshPublicKeySpec          `json:"sshPublicKeySpecs"`
	SuperComputeGroupId       string                      `json:"superComputeGroupId,omitempty"`
	ClusterGroupId            string                      `json:"clusterGroupId"`
	ClusterId                 string                      `json:"clusterId"`
	Region                    string                      `json:"region"`
	NodeId                    string                      `json:"nodeId"`
	ServiceType               string                      `json:"serviceType,omitempty"`
	InstanceName              string                      `json:"instanceName,omitempty"`
	Labels                    map[string]string           `json:"labels,omitempty"`
	TopologySpreadConstraints []TopologySpreadConstraints `json:"topologySpreadConstraints,omitempty"`
	Partition                 string                      `json:"partition,omitempty"`
	UserData                  string                      `json:"userData,omitempty"`
	InstanceGroup             string                      `json:"instanceGroup,omitempty"`
	InstanceGroupSize         int32                       `json:"instanceGroupSize,omitempty"`
	ComputeNodePools          []string                    `json:"computeNodePools,omitempty"`
	QuickConnectEnabled       string                      `json:"quickConnectEnabled,omitempty"`
}

// RunStrategy is a label for the requested Instance Running State.
type RunStrategy string

// These are the valid Instance run strategies.
const (
	// Instance will initially be running and restarted if a failure occurs.
	// It will not be restarted upon successful completion.
	RunStrategyRerunOnFailure RunStrategy = "RerunOnFailure"
	// Instance should never be running.
	RunStrategyHalted RunStrategy = "Halted"
	// Instance should always be running.
	RunStrategyAlways RunStrategy = "Always"
)

type InterfaceSpec struct {
	// Interface name such as eth0.
	Name string `json:"name,omitempty"`
	// VNet name such as us-west-1a-default.
	VNet        string   `json:"vNet,omitempty"`
	DnsName     string   `json:"dnsName"`
	Nameservers []string `json:"nameservers"`
}

type InstanceTypeSpec struct {
	Name             string           `json:"name"`
	DisplayName      string           `json:"displayName"`
	Description      string           `json:"description"`
	InstanceCategory InstanceCategory `json:"instanceCategory,omitempty"`
	Cpu              CpuSpec          `json:"cpu"`
	Gpu              GpuSpec          `json:"gpu"`
	Memory           MemorySpec       `json:"memory"`
	Disks            []DiskSpec       `json:"disks,omitempty"`
	HBMMode          string           `json:"hbmMode"`
}

type InstanceCategory string

const (
	InstanceCategoryVirtualMachine InstanceCategory = "VirtualMachine"
	InstanceCategoryBareMetalHost  InstanceCategory = "BareMetalHost"
)

type CpuSpec struct {
	Cores     uint32 `json:"cores"`
	Id        string `json:"id"`
	ModelName string `json:"modelName"`
	Sockets   uint32 `json:"sockets"`
	Threads   uint32 `json:"threads"`
}

type GpuSpec struct {
	ModelName string `json:"modelName"`
	Count     uint32 `json:"count"`
}

type MemorySpec struct {
	Size      string `json:"size"`
	DimmSize  string `json:"dimmSize"`
	DimmCount uint32 `json:"dimmCount"`
	Speed     uint32 `json:"speed"`
}

type DiskSpec struct {
	Size string `json:"size"`
}

type MachineImageSpec struct {
	Name      string `json:"name"`
	UserName  string `json:"userName"`
	Md5sum    string `json:"md5sum,omitempty"`
	Sha256sum string `json:"sha256sum,omitempty"`
	Sha512sum string `json:"sha512sum,omitempty"`
}

type SshPublicKeySpec struct {
	SshPublicKey string `json:"sshPublicKey"`
}

type TopologySpreadConstraints struct {
	LabelSelector LabelSelector `json:"labelSelector"`
}

type LabelSelector struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	// +kubebuilder:default:=Provisioning
	// A summary of the instance status for display to users. For automation, use conditions instead.
	Phase      InstancePhase             `json:"phase,omitempty"`
	Message    string                    `json:"message,omitempty"`
	Interfaces []InstanceInterfaceStatus `json:"interfaces,omitempty"`
	Conditions []InstanceCondition       `json:"conditions,omitempty"`
	SshProxy   SshProxyTunnelStatus      `json:"sshProxy,omitempty"`
	UserName   string                    `json:"userName,omitempty"`
}

type InstanceInterfaceStatus struct {
	Name         string `json:"name,omitempty"`
	VNet         string `json:"vNet,omitempty"`
	DnsName      string `json:"dnsName"`
	PrefixLength int    `json:"prefixLength"`
	// List of IP addresses.
	Addresses []string `json:"addresses,omitempty"`
	Subnet    string   `json:"subnet"`
	Gateway   string   `json:"gateway"`
	VlanId    int      `json:"vlanId"`
}

type InstanceConditionType string

// These are the valid conditions of an instance.
const (
	// The Instance Controller has created or updated all output objects.
	InstanceConditionAccepted InstanceConditionType = "Accepted"

	// The instance is started (powered on).
	// Corresponds to Kubevirt VirtualMachineInstanceReady.
	// This condition is false when instance is starting, stopping or stopped
	InstanceConditionRunning InstanceConditionType = "Running"

	// The instance is running and has completed running startup scripts for the first time post instance provisioning.
	// For virtual machines, this becomes True when the QEMU Guest Agent is connected to the host but may be changed in the future.
	InstanceConditionStartupComplete InstanceConditionType = "StartupComplete"

	// For virtual machines, the QEMU Guest Agent is connected to the host.
	InstanceConditionAgentConnected InstanceConditionType = "AgentConnected"

	// The SSH proxy is ready to tunnel connections to the instance.
	InstanceConditionSshProxyReady InstanceConditionType = "SshProxyReady"

	// The instance crashed, failed, or is otherwise unavailable.
	InstanceConditionFailed InstanceConditionType = "Failed"

	// Instance KCS status
	InstanceConditionKcsEnabled InstanceConditionType = "KCSEnabled"

	// Instance HCI status
	InstanceConditionHCIEnabled InstanceConditionType = "HCIEnabled"

	// The instance is stopped.
	InstanceConditionStopped InstanceConditionType = "Stopped"

	// The instance is stopping.
	InstanceConditionStopping InstanceConditionType = "Stopping"

	// The instance is starting.
	InstanceConditionStarting InstanceConditionType = "Starting"

	// The instance has completed startup and is available to use post an Instance powercycle.
	InstanceConditionStarted InstanceConditionType = "Started"

	// The instance has verified SSH access and has completed SSH ping
	InstanceConditionVerifiedSshAccess InstanceConditionType = "VerifiedSshAccess"
)

// These are the prefixes of InstanceStatus.Message. The field may have additional details.
const (
	InstanceMessageNew                     string = "Instance reconciliation has not started"
	InstanceMessageProvisioningNotAccepted string = "Instance specification has not been accepted and is being provisioned"
	InstanceMessageProvisioningAccepted    string = "Instance specification has been accepted and is being provisioned"
	InstanceMessageRunning                 string = "Instance is powered on but has not completed running startup scripts"
	InstanceMessageStartupComplete         string = "Instance is running and has completed running startup scripts"
	InstanceMessageFailed                  string = "The instance crashed, failed, or is otherwise unavailable"
	InstanceMessageTerminating             string = "The instance and its associated resources are in the process of being deleted"
	InstanceMessageStopped                 string = "The instance is stopped and is unavailable"
	InstanceMessageStopping                string = "The instance is stopping"
	InstanceMessageStarting                string = "The instance is starting"
	InstanceMessageStarted                 string = "The instance has completed startup and is available to use"
)

type ConditionReason string

const (
	ConditionReasonNone        ConditionReason = ""
	ConditionReasonNotAccepted ConditionReason = "NotAccepted"
	ConditionReasonAccepted    ConditionReason = "Accepted"
)

type InstanceCondition struct {
	Type   InstanceConditionType `json:"type"`
	Status k8sv1.ConditionStatus `json:"status"`
	// +nullable
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// +nullable
	LastTransitionTime metav1.Time     `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason `json:"reason,omitempty"`
	Message            string          `json:"message,omitempty"`
}

type InstancePhase string

// These are the valid phases of instances.
// The phase is always calculated from the conditions. It is never used to determine the state.
const (
	// PhaseProvisioning means the instance has started working on the request.
	PhaseProvisioning InstancePhase = "Provisioning"
	// This corresponds to the StartupComplete condition.
	PhaseReady InstancePhase = "Ready"
	// PhaseStopping means the instance is in the process of being stopped.
	PhaseStopping InstancePhase = "Stopping"
	// PhaseStopped means the is stopped.
	PhaseStopped InstancePhase = "Stopped"
	// PhaseTerminating means the instances and its associated resources are in the process of being deleted.
	PhaseTerminating InstancePhase = "Terminating"
	// PhaseFailed means that the instance crashed, disappeared unexpectedly or got deleted from the cluster before it was ever started.
	// It also is used to indicate errors like image pull error, data volume error and PVC errors.
	PhaseFailed InstancePhase = "Failed"
	// PhaseStarting means that the instance has been provisioned and is in the process of booting.
	PhaseStarting InstancePhase = "Starting"
	// PhaseStarted means that the instance has completed startup and is available to use.
	PhaseStarted InstancePhase = "Started"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+genclient

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta `json:",inline"`
	// This contains metadata such as name, namespace, labels and annotations.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}

func (instance *Instance) NewEvent(reason, message string, related *corev1.ObjectReference) corev1.Event {
	t := metav1.Now()
	return corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: reason + "-",
			Namespace:    instance.ObjectMeta.Namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:       "Instance",
			Namespace:  instance.Namespace,
			Name:       instance.Name,
			UID:        instance.UID,
			APIVersion: SchemeGroupVersion.String(),
		},
		Reason:  reason,
		Message: message,
		Source: corev1.EventSource{
			Component: "instance-controller",
		},
		FirstTimestamp:      t,
		LastTimestamp:       t,
		Count:               1,
		Type:                corev1.EventTypeNormal,
		ReportingController: "instance-operator",
		Related:             related,
	}
}
