// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package controllers

import (
	"context"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function to add a conditions to a given condition list.
func SetStatusCondition(conditions *[]cloudv1alpha1.BMEnrollmentCondition,
	newCondition cloudv1alpha1.BMEnrollmentCondition, updateStartTime bool) {

	if conditions == nil {
		conditions = &[]cloudv1alpha1.BMEnrollmentCondition{}
	}
	existingCondition := FindStatusCondition(*conditions, newCondition.Type)
	if existingCondition == nil {
		if newCondition.LastTransitionTime.IsZero() {
			newCondition.LastTransitionTime = metav1.Now()
		}
		*conditions = append(*conditions, newCondition)
	} else {
		if existingCondition.Status != newCondition.Status {
			existingCondition.Status = newCondition.Status
			if !newCondition.LastTransitionTime.IsZero() {
				existingCondition.LastTransitionTime = newCondition.LastTransitionTime
			} else {
				existingCondition.LastTransitionTime = metav1.Now()
			}
		}
		if updateStartTime {
			existingCondition.StartTime = metav1.Now()
		} else {
			existingCondition.StartTime = newCondition.StartTime
		}
		existingCondition.Reason = newCondition.Reason
		existingCondition.Message = newCondition.Message
		existingCondition.LastProbeTime = newCondition.LastProbeTime
	}
}

func SetStatusConditionIfMissing(bmEnrollment *cloudv1alpha1.BMEnrollment,
	conditionType cloudv1alpha1.BMEnrollmentConditionType, status k8sv1.ConditionStatus,
	reason cloudv1alpha1.ConditionReason, message string) {
	cond := FindStatusCondition(bmEnrollment.Status.Conditions, conditionType)
	if cond == nil {
		bmEnrollmentCondition := cloudv1alpha1.BMEnrollmentCondition{
			LastProbeTime:      metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Message:            message,
			Reason:             reason,
			StartTime:          metav1.Time{},
			Status:             status,
			Type:               conditionType,
		}
		SetStatusCondition(&bmEnrollment.Status.Conditions, bmEnrollmentCondition, false)
	}
}

// Helper to set a condition of given type to true and others to false
func SetBMEnrollmentCondition(ctx context.Context, bmEnrollment *cloudv1alpha1.BMEnrollment,
	conditionType cloudv1alpha1.BMEnrollmentConditionType, status k8sv1.ConditionStatus,
	reason cloudv1alpha1.ConditionReason, message string, updateStartTime bool) {

	startTime := metav1.Time{}
	if !updateStartTime {
		conditionInfo := FindStatusCondition(bmEnrollment.Status.Conditions, conditionType)
		startTime = conditionInfo.StartTime
	}
	condition := cloudv1alpha1.BMEnrollmentCondition{
		LastProbeTime:      metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Message:            message,
		Reason:             reason,
		StartTime:          startTime,
		Status:             status,
		Type:               conditionType,
	}

	SetStatusCondition(&bmEnrollment.Status.Conditions, condition, updateStartTime)
}

// Utility to find a condition of the given type.
func FindStatusCondition(conditions []cloudv1alpha1.BMEnrollmentCondition, conditionType cloudv1alpha1.BMEnrollmentConditionType) *cloudv1alpha1.BMEnrollmentCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

func (reconciler *BMEnrollmentReconciler) setInitialEnrollmentConditions(bmEnrollment *cloudv1alpha1.BMEnrollment) {
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionPreEnrollmentChecks, k8sv1.ConditionTrue, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionStarting, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionFailed, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionGetBMCInterface, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionUpdateBMCConfig, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionBMHStarting, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionBMHRegistering, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionBMHInspecting, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionBMHProvisioning, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionBMHDeprovisioning, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionBMHEnrolled, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionAddLabels, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
	SetStatusConditionIfMissing(bmEnrollment, cloudv1alpha1.BMEnrollmentConditionCompleted, k8sv1.ConditionFalse, cloudv1alpha1.ConditionReasonNone, "")
}
