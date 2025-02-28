// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Helper function to add a conditions to a given condition list.
func SetStatusCondition(conditions *[]cloudv1alpha1.StorageCondition,
	newCondition cloudv1alpha1.StorageCondition) {

	if conditions == nil {
		conditions = &[]cloudv1alpha1.StorageCondition{}
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
		existingCondition.Reason = newCondition.Reason
		existingCondition.Message = newCondition.Message
		existingCondition.LastProbeTime = newCondition.LastProbeTime
	}
}

func SetStatusConditionIfMissing(storage *cloudv1alpha1.Storage,
	conditionType cloudv1alpha1.StorageConditionType, status k8sv1.ConditionStatus,
	reason cloudv1alpha1.StorageConditionReason, message string) {
	cond := FindStatusCondition(storage.Status.Conditions, conditionType)
	if cond == nil {
		StorageCondition := cloudv1alpha1.StorageCondition{
			Status:             status,
			Message:            message,
			Type:               conditionType,
			LastTransitionTime: metav1.Now(),
			LastProbeTime:      metav1.Now(),
			Reason:             reason,
		}
		SetStatusCondition(&storage.Status.Conditions, StorageCondition)
	}
}

// Utility to find a condition of the given type.
func FindStatusCondition(conditions []cloudv1alpha1.StorageCondition, conditionType cloudv1alpha1.StorageConditionType) *cloudv1alpha1.StorageCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
