// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controllers

import "time"

const (
	BMCDeploymentSecretsPath   = "deployed/"
	BMCUserSecretsPrefix       = "user/"
	bmcBIOSSecretsPrefix       = "bios/"
	EnvBMCEnrollUsername       = "BMC_ENROLL_USERNAME"
	DefaultBMCEnrollUsername   = "bmaas"
	Metal3NamespaceSelectorKey = "cloud.intel.com/bmaas-metal3-namespace"
	Metal3NamespaceIronicIPKey = "ironicIP"

	MenAndMiceBMCType          = "BMC"
	MenAndMiceProvisioningType = "Provisioning"

	MenAndMiceUrlEnvVar           = "MEN_AND_MICE_URL"
	MenAndMiceServerAddressEnvVar = "MEN_AND_MICE_SERVER_ADDRESS"
	TftpServerIPEnvVar            = "TFTP_SERVER"
	IPXEProfileName               = "boot.ipxe"
	IronicHttpPortNb              = "6180"
	IPXEBinarayName               = "snponly.efi"

	SetBiosPasswordEnvVar = "SET_BIOS_PASSWORD"
	DhcpProxyUrlEnvVar    = "DHCP_PROXY_URL"

	BMCInterfaceName           = "BMC"
	HostInterfaceName          = "net0/0"
	StorageInterfaceName1      = "net0/1"
	StorageInterfaceName2      = "net1/0"
	ComputeApiServerAddrEnvVar = "COMPUTE_API_SERVER_ADDRESS"
)

// Hardware specification labels
const (
	//
	// GPU Prefix Annotation
	GPUAnnotationPrefix    = "gpu.mac.cloud.intel.com"
	GPUIPsAnnotationPrefix = "gpu.ip.cloud.intel.com"
	// Storage Prefix Annotation
	StorageMACAnnotationPrefix = "storage.mac.cloud.intel.com"
	// CPUManufacturerLabel is the manufacturer of the CPU
	CPUManufacturerLabel = "cloud.intel.com/host-cpu-manufacturer"
	// CPUIDLabel is a unique number that defines a CPU version
	CPUIDLabel = "cloud.intel.com/host-cpu-id"
	// CPUModelLabel is the model name of the CPU
	CPUModelLabel = "cloud.intel.com/host-cpu-model"
	// CPUCountLabel is the total number of logical cores or hyperthreads available on host
	CPUCountLabel = "cloud.intel.com/host-cpu-count"
	// CPUSocketsLabel is total number of enabled CPU sockets on host
	CPUSocketsLabel = "cloud.intel.com/host-cpu-sockets"
	// CPUCoresLabel is the total number of physical cores available on host
	CPUCoresLabel = "cloud.intel.com/host-cpu-cores"
	// CPUThreadsLabel is the total number of CPU threads available on host
	CPUThreadsLabel = "cloud.intel.com/host-cpu-threads"
	// GPU Count Label
	GPUCountLabel = "cloud.intel.com/host-gpu-count"
	// GPU  Device ID Label
	GPUModelNameLabel = "cloud.intel.com/host-gpu-model"
	// HBM mode Label
	HBMModeLabel = "cloud.intel.com/hbm-mode"
	// latest associated resource
	LastAssociatedInstance = "last-associated-instance"
	// Label key applied to Instance to identify the instance group.
	ClusterGroup     = "instance-group"
	LastClusterGroup = "last-cluster-group"
	// Label key applied to the memory size label to match the memory of BM and InstanceTypes
	MemorySizeLabel = "cloud.intel.com/host-memory-size"
	// Label key applied to BareMetalHost to identify hosts that connect to the same cluster fabric.
	// Hosts with the same value should connect to the same cluster fabric and can be consumed by instances in
	// the same instance group.
	ClusterGroupID = "cloud.intel.com/instance-group-id"
	// Label key applied to BareMetalHost.
	// applied value define the number of nodes in the cluster
	ClusterSize = "cloud.intel.com/cluster-size"
	// Label key applied to Instance to label the Instance Type.
	// applied value define the type of instance type
	InstanceTypeLabel = "instance-type.cloud.intel.com/%s"
	// Labels used by the validation operator
	// Label used to trigger the validation process by the validation operator.
	ReadyToTestLabel = "cloud.intel.com/ready-to-test"
	// Label set by the validation operator to indicate that the validation completed successfully.
	VerifiedLabel = "cloud.intel.com/verified"
	// Label that indicates that the checking/validation failed. The value will have type of validation that failed
	CheckingFailedLabel = "cloud.intel.com/validation-check-failed"
	// Label used by the scheduler to prevent node from being scheduled.
	UnschedulableLabel = "cloud.intel.com/unschedulable"
	// Label used to indicate the network mode of the node.
	NetworkModeLabel = "cloud.intel.com/network-mode"
	// network mode for non-spine-leaf accelerator fabric isolation type with VLAN.
	NetworkModeVVXStandalone = "VVX-standalone"
	// network mode for spine-leaf accelerator fabric isolation type with BGP.
	NetworkModeXBX = "XBX"
	// network mode for partitioned-leaf accelerator fabric isolation type with VLAN.
	NetworkModeVVV = "VVV"

	// Label which indicates the state of the validation process
	// Imaging is in progress for validation.
	ImagingLabel = "cloud.intel.com/validation-imaging"
	// Wait for all the instances in the Instance group to complete individual validation
	WaitForInstanceValidation = "cloud.intel.com/group-wait-for-InstanceValidation"
	// Image process has completed and the bmh is initialized.
	ImagingCompletedLabel = "cloud.intel.com/validation-imaging-completed"
	// Instance validation has completed for all bmhs in the instance group.
	InstanceValidationCompletedLabel = "cloud.intel.com/validation-instance-completed"
	// Checking/Validation process for InstanceGroup is in progress.
	CheckingGroupLabel = "cloud.intel.com/group-validation-checking"
	// Checking/Validation process is in progress.
	CheckingLabel = "cloud.intel.com/validation-checking"
	// Checking/Validation process has completed.
	CheckingCompletedLabel = "cloud.intel.com/validation-checking-completed"
	// Checking/Validation process for InstanceGroup has completed.
	CheckingCompletedGroupLabel = "cloud.intel.com/validation-checking-completed-group"
	// Deletion for validation process
	DeletionLabel = "cloud.intel.com/deletion-for-validation"
	// Label to indicate a node is a master node for cluster validation
	MasterNodeLabel = "cloud.intel.com/validation-master-node"
	// Label to gate cluster validation.
	GateValidationLabel = "cloud.intel.com/validation-gating"
	// Label to represent the validation id.
	ValidationIdLabel = "cloud.intel.com/validation-id"
)

const (
	EnrollmentFinalizer = "private.cloud.intel.com/baremetal-enrollment-operator"
)

const (
	BMCTaskRequeueAfter             = 30 * time.Second
	GeneralRequeueAfter             = 30 * time.Second
	BMHRegisterRequeueAfter         = 2 * time.Second
	BMHInspectionRequeueAfter       = 20 * time.Second
	BMHProvisionRequeueAfter        = 20 * time.Second
	BMHDeprovisionRequeueAfter      = 30 * time.Second
	EnrollmentPeriodicRequeueAfter  = 30 * time.Minute
	EnrollmentGeneralTimeout        = 30 * time.Minute
	EnrollmentRegistrationTimeout   = 15 * time.Minute
	EnrollmentInspectionTimeout     = 30 * time.Minute
	EnrollmentProvisioningTimeout   = 180 * time.Minute
	EnrollmentDeprovisioningTimeout = 180 * time.Minute
)
