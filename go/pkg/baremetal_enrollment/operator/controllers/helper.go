// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controllers

import (
	"strings"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "k8s.io/api/core/v1"
)

func startBMEnrollment(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionStarting && condition.Status == v1.ConditionFalse
}

func getBMCInterface(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface && condition.Status == v1.ConditionFalse
}

func updateBMCConfig(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionUpdateBMCConfig && condition.Status == v1.ConditionFalse
}

func enrollBareMetalHost(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionBMHStarting && condition.Status == v1.ConditionFalse
}

func registerBareMetalHost(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionBMHRegistering && condition.Status == v1.ConditionFalse
}

func inspectBareMetalHost(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionBMHInspecting && condition.Status == v1.ConditionFalse
}

func provisionBareMetalHost(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionBMHProvisioning && condition.Status == v1.ConditionFalse
}

func deprovisionBareMetalHost(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionBMHDeprovisioning && condition.Status == v1.ConditionFalse
}

func validateBareMetalHostEnrollment(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionBMHEnrolled && condition.Status == v1.ConditionFalse
}

func addBareMetalHostLabels(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionAddLabels && condition.Status == v1.ConditionFalse
}

func completedEnrollment(condition privatecloudv1alpha1.BMEnrollmentCondition) bool {
	return condition.Type == privatecloudv1alpha1.BMEnrollmentConditionCompleted && condition.Status == v1.ConditionFalse
}

func emptyBMCMACAddress(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	return bmEnrollment.Status.BMC.MACAddress == ""
}

func emptyBMCURL(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	return bmEnrollment.Status.BMC.Address == ""
}

func emptyEnrollmentBMCSecret(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	return bmEnrollment.Status.BMC.SecretName == ""
}

func bareMetalHostNamespaceAssigned(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	return bmEnrollment.Status.TargetBmNamespace != ""
}

func enrollmentCompleted(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionCompleted)
	return condition.Status == v1.ConditionTrue && bmEnrollment.Status.Phase == privatecloudv1alpha1.BMEnrollmentPhaseReady
}

func enrollmentFailed(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionFailed)
	return condition.Status == v1.ConditionTrue && bmEnrollment.Status.Phase == privatecloudv1alpha1.BMEnrollmentPhaseFailed
}

func createNewBMCUser(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	return (bmEnrollment.Status.BMC.CreateNewBMCUser != privatecloudv1alpha1.CreateBMCUserPassed) && (bmEnrollment.Status.BMC.CreateNewBMCUser != privatecloudv1alpha1.CreateBMCUserNotSupported)
}

func emptyBootAddress(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	return bmEnrollment.Status.BootMACAddress == ""
}

func emptyMetal3Address(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	return bmEnrollment.Status.BMC.Metal3Address == ""
}

func enableKCS(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	return (bmEnrollment.Status.BMC.KCS != privatecloudv1alpha1.KCSEnabled) && (bmEnrollment.Status.BMC.KCS != privatecloudv1alpha1.KCSNotSupported)
}

func enableHCI(bmEnrollment *privatecloudv1alpha1.BMEnrollment) bool {
	return (bmEnrollment.Status.BMC.HCI != privatecloudv1alpha1.HCIEnabled) && (bmEnrollment.Status.BMC.HCI != privatecloudv1alpha1.HCINotSupported)
}

func bareMetalHostHasError(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Status.ErrorMessage != ""
}

func bareMetalHostRegistrationInProgress(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Status.Provisioning.State == baremetalv1alpha1.StateRegistering || bmHost.Status.Provisioning.State == baremetalv1alpha1.StateNone
}

func bareMetalHostInspectionInProgress(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Status.Provisioning.State == baremetalv1alpha1.StateInspecting
}

// BMH registration is completed when provisioning state is transitioned to inspecting
func bareMetalHostRegistrationCompleted(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bareMetalHostInspectionInProgress(bmHost)
}

func bareMetalHostAvailable(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Status.Provisioning.State == baremetalv1alpha1.StateAvailable
}

func bareMetalHostPreparing(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Status.Provisioning.State == baremetalv1alpha1.StatePreparing
}

func bareMetalHostInspectionCompleted(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bareMetalHostAvailable(bmHost)
}

func bareMetalHostProvisioningInProgress(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Status.Provisioning.State == baremetalv1alpha1.StateProvisioning
}

func bareMetalHostProvisioned(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Status.Provisioning.State == baremetalv1alpha1.StateProvisioned
}

func bareMetalHostProvisioningCompleted(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bareMetalHostProvisioned(bmHost)
}

func bareMetalHostDeprovisioningInProgress(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Status.Provisioning.State == baremetalv1alpha1.StateDeprovisioning
}

func bareMetalHostDeprovisioningCompleted(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bareMetalHostAvailable(bmHost) || bareMetalHostPreparing(bmHost)
}

func bareMetalHostEnrolled(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bareMetalHostAvailable(bmHost) && bmHost.Status.OperationalStatus == baremetalv1alpha1.OperationalStatusOK && bmHost.Status.ErrorCount == 0 && bmHost.Status.ErrorMessage == ""
}

func emptyBareMetalHostImage(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Spec.Image == nil
}

func bareMetalHostConsumed(bmHost *baremetalv1alpha1.BareMetalHost) bool {
	return bmHost.Spec.ConsumerRef != nil
}

// formatLabelValue returns a label value with invalid characters removed
func formatLabelValue(value string) string {
	r := strings.NewReplacer(
		"(R)", "",
		"(", "",
		")", "",
		"@", "",
	)
	return strings.Join(strings.Fields(r.Replace(value)), "_")
}

func calculateTotalCpuCount(instance *pb.InstanceType) int {
	cpu := instance.Spec.Cpu
	return int(cpu.Cores * cpu.Sockets * cpu.Threads)
}
