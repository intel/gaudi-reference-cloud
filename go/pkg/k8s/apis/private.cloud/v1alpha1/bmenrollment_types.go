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
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BMEnrollmentConditionType string

// These are the valid conditions of an BMEnrollment.
const (

	// The BMEnrollment task is starting.
	BMEnrollmentConditionStarting BMEnrollmentConditionType = "Starting"

	// The BMEnrollment task is completed
	BMEnrollmentConditionCompleted BMEnrollmentConditionType = "Completed"

	// The BMEnrollment task is failed
	BMEnrollmentConditionFailed BMEnrollmentConditionType = "Failed"

	// BMenrollment - PreEnrollment Checks
	BMEnrollmentConditionPreEnrollmentChecks = "Pre-Enrollment-Checks"

	// BMenrollment - Create BMC Interface to communicate with host BMC
	BMEnrollmentConditionGetBMCInterface = "Get-BMC-Interface"

	// BMenrollment - BMH(BareMetalHost) enrollment starting
	BMEnrollmentConditionBMHStarting BMEnrollmentConditionType = "BMH-Enrollment-Starting"

	// BMEnrollment - BMH(BareMetalHost) in registering state
	BMEnrollmentConditionBMHRegistering BMEnrollmentConditionType = "BMH-Registering"

	// BMEnrollment - BMH(BareMetalHost) in inspecting state
	BMEnrollmentConditionBMHInspecting BMEnrollmentConditionType = "BMH-Inspecting"

	// BMEnrollment - BMH(BareMetalHost) in provisioning state
	BMEnrollmentConditionBMHProvisioning BMEnrollmentConditionType = "BMH-Provisioning"

	// BMEnrollment - BMH(BareMetalHost) in deprovisioning state
	BMEnrollmentConditionBMHDeprovisioning BMEnrollmentConditionType = "BMH-Deprovisioning"

	// BMEnrollment - BMH(BareMetalHost) enrollment completed
	BMEnrollmentConditionBMHEnrolled BMEnrollmentConditionType = "BMH-Enrolled"

	// BMEnrollment - BMC configuration update task
	BMEnrollmentConditionUpdateBMCConfig BMEnrollmentConditionType = "Update-BMC-Config"

	// BMEnrollment - Update BMH labels
	BMEnrollmentConditionAddLabels BMEnrollmentConditionType = "BMH-Add-Labels"

	// BMenrollment - PreDisenrollment Checks
	BMEnrollmentConditionPreDisenrollmentChecks = "Pre-Disenrollment-Checks"
)

const (
	//BMEnrollmentMessageFailed                                 string = "BM Enrollment failed"
	BMEnrollmentMessagePreEnrollmentChecksStarted             string = "starting pre BM enrollment checks"
	BMEnrollmentMessagePreEnrollmentChecksCompleted           string = "completed pre BM enrollment checks"
	BMEnrollmentMessageStarting                               string = "BM Enrollment is starting"
	BMEnrollmentMessageGetBMCInterface                        string = "Starting getBMCInterface task"
	BMEnrollmentMessageFailedBMCMACAddress                    string = "Failed to get the BMC MAC address"
	BMEnrollmentMessageFailedBMCURLAddress                    string = "Failed to get the BMC URL address"
	BMEnrollmentMessageFailedDefaultBMCCredentials            string = "Failed to get default BMC credentials"
	BMEnrollmentMessageFailedCreateBMCSecret                  string = "Failed to create secret with BMC credentials"
	BMEnrollmentMessageFailedNewBMCInterface                  string = "Failed to create new BMC interface"
	BMEnrollmentMessageFailedVerifyBMCCredentials             string = "Failed to verify BMC credentials"
	BMEnrollmentMessageFailedToGenerateBMCCredentials         string = "Failed to generate new BMC user credentials"
	BMEnrollmentMessageFailedToStoreUserBMCCredentials        string = "Failed to store user BMC credentials in vault"
	BMEnrollmentMessageFailedToDeleteBMCCredentials           string = "Failed to delete user BMC credentials from vault"
	BMEnrollmentMessageFailedToUpdateUserBMCCredentials       string = "Failed to update user BMC credentials in the host"
	BMEnrollmentMessageFailedToPatchK8SSecret                 string = "Failed to patch K8s secret with user BMC credentials"
	BMEnrollmentMessageCompletedGetBMCMACAddress              string = "GetBMCInterface task completed"
	BMEnrollmentMessageStartUpdatingBMCConfig                 string = "Start updating BMC config"
	BMEnrollmentMessageFailedToGetBMCCredentialsK8sSecretData string = "Failed to get BMC credentials secret data"
	BMEnrollmentMessageFailedToGetBootMACAddress              string = "Failed to get boot MAC address"
	BMEnrollmentMessageFailedToUpdateBMCBootOrder             string = "Failed to update BMC boot order"
	BMEnrollmentMessageFailedToGetMetal3BMCAddress            string = "Failed to Get BMC Address in Metal3 supported format"
	BMEnrollmentMessageFailedToUpdateBMCNtp                   string = "Failed to update NTP servers of BMC"
	BMEnrollmentMessageFailedToVerifyBMCPfr                   string = "Failed to verify platform resilience"
	BMEnrollmentMessageFailedToEnableKCS                      string = "Failed to enable KCS"
	BMEnrollmentMessageFailedToEnableHCI                      string = "Failed to enable HCI"
	BMEnrollmentMessageCompletedUpdatingBMCConfig             string = "Task to update the BMC configuration is completed"
	BMEnrollmentMessageStartBMHEnrolling                      string = "Start enrolling BareMetalHost"
	BMEnrollmentMessageFailedToGetTargetNamespace             string = "Failed to get target namespace"
	BMEnrollmentMessageFailedToGetIronicIP                    string = "Failed to get the ironic IP address"
	BMEnrollmentMessageFailedToCreatePXERecord                string = "Failed to create PXE record in the DHCP server"
	BMEnrollmentMessageFailedToCreateBMH                      string = "Failed to create BareMetalHost"
	BMEnrollmentMessageFailedToCreateBMHSecret                string = "Failed to create BareMetalHost secret"
	BMEnrollmentMessageCompletedCreatingBMH                   string = "Task to create the BareMetalHost is completed"
	BMEnrollmentMessageFailedToGetBMH                         string = "Failed to get the BareMetalHost"
	BMEnrollmentMessageBMHRegistrationInProgress              string = "BareMetalHost registration is in progress"
	BMEnrollmentMessageStartRegisteringBMH                    string = "Start registering BareMetalHost"
	BMEnrollmentMessageFailedToRegisterBMH                    string = "Failed to register BareMetalHost"
	BMEnrollmentMessageCompletedRegisteringBMH                string = "Completed BareMetalHost registration"
	BMEnrollmentMessageStartInspectingBMH                     string = "Start inspecting BareMetalHost"
	BMEnrollmentMessageFailedToInspectBMH                     string = "Failed to inspect BareMetalHost"
	BMEnrollmentMessageBMHInspectionInProgress                string = "BareMetalHost inspection is in progress"
	BMEnrollmentMessageCompletedInspectingBMH                 string = "Completed BareMetalHost inspection"
	BMEnrollmentMessageFailedToPatchBMH                       string = "Failed to patch BareMetalHost"
	BMEnrollmentMessageStartProvisioningBMH                   string = "Start provisioning BareMetalHost"
	BMEnrollmentMessageFailedToMarshalBMHImageData            string = "Failed to marshal BareMetalHostImageData"
	BMEnrollmentMessageBMHProvisioningInProgress              string = "BareMetalHost provisioning is in progress"
	BMEnrollmentMessageFailedToProvisionBMH                   string = "Failed to provision BareMetalHost"
	BMEnrollmentMessageCompletedProvisioningBMH               string = "Completed BareMetalHost provisioning"
	BMEnrollmentMessageStartDeprovisioningBMH                 string = "Start deprovisioning BareMetalHost"
	BMEnrollmentMessageFailedToDeprovisionBMH                 string = "Failed to deprovision BareMetalHost"
	BMEnrollmentMessageBMHDeprovisioningInProgress            string = "BareMetalHost deprovisioning is in progress"
	BMEnrollmentMessageCompletedDeprovisioningBMH             string = "Completed BareMetalHost deprovisioning"
	BMEnrollmentMessageStartBMHEnrollmentValidation           string = "Start validating BareMetalHost enrollment to Ironic"
	BMEnrollmentMessageCompletedBMHEnrollmentValidation       string = "Completed validating BareMetalHost enrollment to Ironic"
	BMEnrollmentMessageAddLabels                              string = "Start adding required labels to the BareMetalHost"
	BMEnrollmentMessageFailedToGetHostIP                      string = "Failed to find BareMetalHost IP address"
	BMEnrollmentMessageFailedToSetBIOSPassword                string = "Failed to set BIOS password"
	BMEnrollmentMessageFailedToAddStorageAnnotations          string = "Failed to add storage annotations to the BareMetalHost"
	BMEnrollmentMessageFailedToAddHardwareLabels              string = "Failed to add hardware labels to the BareMetalHost"
	BMEnrollmentMessageCompletedLabels                        string = "Added required labels to the BareMetalHost"
	BMEnrollmentMessageCompletedEnrollment                    string = "completed enrollment"
	BMEnrollmentMessageConsumedBMH                            string = "Failed to enroll BareMetalHost as it is consumed by an instance"
	BMEnrollmentMessageDisenrollmentConsumedBMH               string = "Failed to disenroll BareMetalHost as it is consumed by an instance"
	BMEnrollmentMessageEnrollmentTaskTimedOut                 string = "Failed to complete enrollment task within the configured timeout duration"
	BMEnrollmentMessagePreDisenrollmentChecksStarted          string = "Disenrollment checks started"
	BMEnrollmentMessageFailedToDeleteBMH                      string = "Failed to delete BareMetalHost"
	BMEnrollmentMessageFailedToDeleteEnrollment               string = "Failed to delete BareMetal enrollment"
)

type BMEnrollmentConditionReason string

const (
	BMEnrollmentConditionReasonNone                                   ConditionReason = ""
	BMEnrollmentConditionReasonNotAccepted                            ConditionReason = "NotAccepted"
	BMEnrollmentConditionReasonAccepted                               ConditionReason = "Accepted"
	BMEnrollmentConditionReasonPreEnrollmentChecksStarted             ConditionReason = "PreEnrollmentChecksStarted"
	BMEnrollmentConditionReasonPreEnrollmentChecksCompleted           ConditionReason = "PreEnrollmentChecksCompleted"
	BMEnrollmentConditionReasonTaskStarted                            ConditionReason = "EnrollmentTaskStarted"
	BMEnrollmentConditionReasonStartGetBMCInterface                   ConditionReason = "GetBMCInterfaceTaskStarted"
	BMEnrollmentConditionReasonFailedBMCMACAddress                    ConditionReason = "FailedToGetBMCMACAddress"
	BMEnrollmentConditionReasonFailedBMCURLAddress                    ConditionReason = "FailedToGetBMCURLAddress"
	BMEnrollmentConditionReasonFailedGetBMCCredentials                ConditionReason = "FailedToGetBMCCredentials"
	BMEnrollmentConditionReasonFailedCreateBMCSecret                  ConditionReason = "FailedToCreateBMCSecret"
	BMEnrollmentConditionReasonFailedNewBMCInterface                  ConditionReason = "FailedToCreateNewBMCInterface"
	BMEnrollmentConditionReasonFailedVerifyBMCCredentials             ConditionReason = "FailedToVerifyBMCCredentials"
	BMEnrollmentConditionReasonFailedToGenerateBMCCredentials         ConditionReason = "FailedToGenerateBMCCredentials"
	BMEnrollmentConditionReasonFailedToStoreUserBMCCredentials        ConditionReason = "FailedToStoreUserBMCCredentials"
	BMEnrollmentConditionReasonFailedToDeleteBMCCredentials           ConditionReason = "FailedToDeleteUserBMCCredentials"
	BMEnrollmentConditionReasonFailedToUpdateUserBMCCredentials       ConditionReason = "FailedToUpdateUserBMCCredentials"
	BMEnrollmentConditionReasonFailedToPatchK8SSecret                 ConditionReason = "FailedToPatchK8SSecret"
	BMEnrollmentConditionReasonCompletedGetBMCInterface               ConditionReason = "CompletedGetBMCInterface"
	BMEnrollmentConditionReasonStartUpdatingBMCConfig                 ConditionReason = "StartUpdatingBMCConfig"
	BMEnrollmentConditionReasonFailedToGetBMCCredentialsK8sSecretData ConditionReason = "FailedToGetBMCCredentialsK8sSecretData"
	BMEnrollmentConditionReasonFailedToGetBootMACAddress              ConditionReason = "FailedToGetBootMACAddress"
	BMEnrollmentConditionReasonFailedToUpdateBMCBootOrder             ConditionReason = "FailedToUpdateBMCBootOrder"
	BMEnrollmentConditionReasonFailedToGetMetal3BMCAddress            ConditionReason = "FailedToGetMetal3BMCAddress"
	BMEnrollmentConditionReasonFailedToUpdateBMCNtp                   ConditionReason = "FailedToUpdateBMCNtp"
	BMEnrollmentConditionReasonFailedToVerifyBMCPfr                   ConditionReason = "FailedToVerifyBMCPfr"
	BMEnrollmentConditionReasonFailedToEnableKCS                      ConditionReason = "FailedToEnableKCS"
	BMEnrollmentConditionReasonFailedToEnableHCI                      ConditionReason = "FailedToEnableHCI"
	BMEnrollmentConditionReasonCompletedUpdatingBMCConfig             ConditionReason = "CompletedUpdatingBMCConfig"
	BMEnrollmentConditionReasonStartBMHEnrolling                      ConditionReason = "StartBMHEnrolling"
	BMEnrollmentConditionReasonFailedToGetTargetNamespace             ConditionReason = "FailedToGetTargetNamespace"
	BMEnrollmentConditionReasonFailedToGetIronicIP                    ConditionReason = "FailedToGetIronicIP"
	BMEnrollmentConditionReasonFailedToCreatePXERecord                ConditionReason = "FailedToCreatePXERecord"
	BMEnrollmentConditionReasonFailedToCreateBMH                      ConditionReason = "FailedToCreateBMH"
	BMEnrollmentConditionReasonFailedToCreateBMHSecret                ConditionReason = "FailedToCreateBMHSecret"
	BMEnrollmentConditionReasonCompletedCreatingBMH                   ConditionReason = "CompletedBMHCreation"
	BMEnrollmentConditionReasonStartRegisteringBMH                    ConditionReason = "StartRegisteringBMH"
	BMEnrollmentConditionReasonFailedToGetBMH                         ConditionReason = "FailedToGetBMH"
	BMEnrollmentConditionReasonBMHRegistrationInProgress              ConditionReason = "BMHRegistrationInProgress"
	BMEnrollmentConditionReasonFailedToRegisterBMH                    ConditionReason = "FailedToRegisterBMH"
	BMEnrollmentConditionReasonCompletedRegisteringBMH                ConditionReason = "CompleteBMHRegistration"
	BMEnrollmentConditionReasonStartInspectingBMH                     ConditionReason = "StartInspectingBMH"
	BMEnrollmentConditionReasonFailedToInspectBMH                     ConditionReason = "FailedToInspectBMH"
	BMEnrollmentConditionReasonBMHInspectionInProgress                ConditionReason = "BMHInspectionInProgress"
	BMEnrollmentConditionReasonCompletedInspectingBMH                 ConditionReason = "CompletedBMHInspection"
	BMEnrollmentConditionReasonFailedToPatchBMH                       ConditionReason = "FailedToPatchBMH"
	BMEnrollmentConditionReasonStartProvisioningBMH                   ConditionReason = "StartProvisioningBMH"
	BMEnrollmentConditionReasonFailedToProvisionBMH                   ConditionReason = "FailedToProvisionBMH"
	BMEnrollmentConditionReasonFailedToMarshalBMHImageData            ConditionReason = "FailedToMarshalBMHImageData"
	BMEnrollmentConditionReasonBMHProvisioningInProgress              ConditionReason = "BMHProvisioningInProgress"
	BMEnrollmentConditionReasonCompletedProvisioningBMH               ConditionReason = "CompletedBMHProvisioning"
	BMEnrollmentConditionReasonStartDeprovisioningBMH                 ConditionReason = "StartDeprovisioningBMH"
	BMEnrollmentConditionReasonFailedToDeprovisionBMH                 ConditionReason = "FailedToDeprovisionBMH"
	BMEnrollmentConditionReasonBMHDeprovisioningInProgress            ConditionReason = "BMHDeprovisioningInProgress"
	BMEnrollmentConditionReasonCompletedDeprovisioningBMH             ConditionReason = "CompletedBMHDeprovisioning"
	BMEnrollmentConditionReasonStartBMHEnrollmentValidation           ConditionReason = "StartBMHEnrollmentValidation"
	BMEnrollmentConditionReasonCompletedBMHEnrollmentValidation       ConditionReason = "CompletedBMHEnrollmentValidation"
	BMEnrollmentConditionReasonAddLabels                              ConditionReason = "AddLabels"
	BMEnrollmentConditionReasonFailedToGetHostIP                      ConditionReason = "FailedToGetHostIP"
	BMEnrollmentConditionReasonFailedToSetBIOSPassword                ConditionReason = "FailedToSetBIOSPassword"
	BMEnrollmentConditionReasonFailedToAddStorageAnnotations          ConditionReason = "FailedToAddStorageAnnotations"
	BMEnrollmentConditionReasonFailedToAddHardwareLabels              ConditionReason = "FailedToAddHardwareLabels"
	BMEnrollmentConditionReasonCompletedLabels                        ConditionReason = "CompletedLabels"
	BMEnrollmentConditionReasonCompletedEnrollment                    ConditionReason = "CompletedEnrollment"
	BMEnrollmentConditionReasonConsumedBMH                            ConditionReason = "ConsumedBMH"
	BMEnrollmentConditionReasonTimedOut                               ConditionReason = "EnrollmentTaskTimedOut"
	BMEnrollmentConditionReasonPreDisenrollmentChecksStarted          ConditionReason = "PreDisenrollmentChecksStarted"
	BMEnrollmentConditionReasonFailedToDeleteBMH                      ConditionReason = "FailedToDeleteBMH "
	BMEnrollmentConditionReasonFailedToDeleteEnrollment               ConditionReason = "FailedToDeleteBMHEnrollment"
)

type BMEnrollmentPhase string

const (
	BMEnrollmentPhaseStarting        BMEnrollmentPhase = "Starting"
	BMEnrollmentPhaseReady           BMEnrollmentPhase = "Ready"
	BMEnrollmentPhaseFailed          BMEnrollmentPhase = "Failed"
	BMEnrollmentPhaseGetBMCInterface BMEnrollmentPhase = "Get-BMC-Interface"
	BMEnrollmentPhaseEnrolling       BMEnrollmentPhase = "BMH-Enrolling"
	BMEnrollmentPhaseUpdateBMCConfig BMEnrollmentPhase = "Update-BMC-Config"
	BMEnrollmentPhaseBMHLabels       BMEnrollmentPhase = "Add-BMH-Labels"
	BMEnrollmentPhaseDisenrolling    BMEnrollmentPhase = "Disenrolling"
)

const (
	CreateBMCUserFailed       string = "failed"
	CreateBMCUserPassed       string = "passed"
	CreateBMCUserNotSupported string = "notSupported"
)

type KCSStatus string

const (
	KCSEnabled      KCSStatus = "enabled"
	KCSDisabled     KCSStatus = "disabled"
	KCSNotSupported KCSStatus = "notSupported"
)

type HCIStatus string

const (
	HCIEnabled      HCIStatus = "enabled"
	HCIDisabled     HCIStatus = "disabled"
	HCINotSupported HCIStatus = "notSupported"
)

type BMEnrollmentCondition struct {
	Type   BMEnrollmentConditionType `json:"type"`
	Status k8sv1.ConditionStatus     `json:"status"`
	// +nullable
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// +nullable
	LastTransitionTime metav1.Time     `json:"lastTransitionTime,omitempty"`
	Reason             ConditionReason `json:"reason,omitempty"`
	// +nullable
	StartTime metav1.Time `json:"startTime,omitempty"`
	Message   string      `json:"message,omitempty"`
}

// BMC status
type BMC struct {
	// BMC Address
	Address string `json:"address"`
	// Create BMC User
	CreateNewBMCUser string `json:"createNewBMCUser"`
	// BMC hardware type
	HardwareType string `json:"hardwareType"`
	// Host BMC Address(Metal3 format)
	Metal3Address string `json:"metal3Address"`
	// HCI status
	HCI HCIStatus `json:"hci"`
	// KCS status
	KCS KCSStatus `json:"kcs"`
	// BMC MAC address
	MACAddress string `json:"macAddress"`
	// BMC secret Name
	SecretName string `json:"secretName"`
}

// BMEnrollmentSpec defines the desired state of BMEnrollment
type BMEnrollmentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Availability Zone
	AvailabilityZone string `json:"az"`
	// Cluster Name
	Cluster string `json:"cluster,omitempty"`
	// Device ID
	DeviceID int64 `json:"deviceID"`
	// Device Name
	DeviceName string `json:"deviceName"`
	// Rack Name
	RackName string `json:"rack"`
	// IDC Region
	Region string `json:"region"`
}

// BMEnrollmentStatus defines the observed state of BMEnrollment
type BMEnrollmentStatus struct {
	// BMC status
	BMC BMC `json:"bmc"`
	// host boot MAC address
	BootMACAddress string `json:"bootMACAddress"`
	// BM Enrollment conditions
	Conditions []BMEnrollmentCondition `json:"conditions,omitempty"`
	// enrollment error message
	ErrorMessage string `json:"errorMessage"`
	// Host IP Address
	HostIPAddress string `json:"hostIPAddress"`
	// Ironic IP
	IronicIPAddress string `json:"ironicIPAddress"`
	// BM enrollment Phase
	Phase BMEnrollmentPhase `json:"phase"`
	// Target BM namespace
	TargetBmNamespace string `json:"targetBmNamespace"`
}

//+kubebuilder:object:root=true
// +kubebuilder:resource:shortName=bme
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AZ",type=string,JSONPath=`.spec.az`
// +kubebuilder:printcolumn:name="DeviceID",type=integer,JSONPath=`.spec.deviceID`
// +kubebuilder:printcolumn:name="Rack",type=string,JSONPath=`.spec.rack`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// BMEnrollment is the Schema for the bmenrollments API
type BMEnrollment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BMEnrollmentSpec   `json:"spec,omitempty"`
	Status BMEnrollmentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BMEnrollmentList contains a list of BMEnrollment
type BMEnrollmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BMEnrollment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BMEnrollment{}, &BMEnrollmentList{})
}
