// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package validation

import (
	"strconv"

	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	baremetalv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
)

type ValidationState string

const (
	STATE_BEGIN                      ValidationState = "Begin"
	STATE_BEGIN_INSTANCE_GROUP       ValidationState = "BeginInstanceGroup"
	STATE_INTIALIZING_INSTANCE_GROUP ValidationState = "InitializingInstanceGroup"
	STATE_INTIALIZING                ValidationState = "Initializing"
	STATE_INITIALIZED_INSTANCE_GROUP ValidationState = "InitializedInstanceGroup"
	STATE_INITIALIZED                ValidationState = "Initialized"
	STATE_VERIFYING                  ValidationState = "Verifying"
	STATE_VERIFYING_INSTANCE_GROUP   ValidationState = "VerifyingInstanceGroup"
	STATE_VERIFIED                   ValidationState = "Verified"
	STATE_NOT_REQUIRED               ValidationState = "ValidationNotRequired"
)

type StateHelper struct {
	bmh *baremetalv1alpha1.BareMetalHost
	cfg *cloudv1alpha1.BmInstanceOperatorConfig
}

// Create a StateHelper
func getStateHelper(bmh *baremetalv1alpha1.BareMetalHost, cfg *cloudv1alpha1.BmInstanceOperatorConfig) *StateHelper {
	return &StateHelper{
		bmh: bmh,
		cfg: cfg,
	}
}

func (p *StateHelper) GetCurrentState() ValidationState {
	if p.isReadyToTest() {
		if p.isVerified() {
			return STATE_VERIFIED
		} else if p.isGroupVerificationInProgress() {
			return STATE_VERIFYING_INSTANCE_GROUP
		} else if p.isInstanceVerificationInProgress() {
			return STATE_VERIFYING
		} else if p.isGroupInitialized() {
			return STATE_INTIALIZING_INSTANCE_GROUP
		} else if p.isInitialized() {
			return STATE_INITIALIZED
		} else if p.isInitializing() {
			return STATE_INTIALIZING
		} else if p.isInstanceGroup() {
			// Enable Instance group only if GroupValidation flag is set to true and instanceType is enabled
			if p.isFeatureGroupValidationEnabled() {
				return STATE_BEGIN_INSTANCE_GROUP
			} else {
				return STATE_BEGIN
			}
		} else {
			return STATE_BEGIN
		}
	} else {
		return STATE_NOT_REQUIRED
	}
}

func (p *StateHelper) isVerified() bool {
	return checkBoolLabel(bmenroll.CheckingCompletedLabel, p.bmh)
}

func (p *StateHelper) isInstanceVerificationInProgress() bool {
	return checkBoolLabel(bmenroll.CheckingLabel, p.bmh)
}

func (p *StateHelper) isGroupVerificationInProgress() bool {
	return checkBoolLabel(bmenroll.CheckingGroupLabel, p.bmh)
}

func (p *StateHelper) isGroupInitialized() bool {
	return checkBoolLabel(bmenroll.InstanceValidationCompletedLabel, p.bmh)
}
func (p *StateHelper) isInitialized() bool {
	return checkBoolLabel(bmenroll.ImagingCompletedLabel, p.bmh)
}

func (p *StateHelper) isInitializing() bool {
	return checkBoolLabel(bmenroll.ImagingLabel, p.bmh)
}

func (p *StateHelper) isInstanceGroup() bool {
	return CheckLabelExists(bmenroll.ClusterGroupID, p.bmh)
}

func (p *StateHelper) isSkipGroupValidationEnabled() bool {
	return CheckLabelExists(bmenroll.SkipGroupValidationLabel, p.bmh)
}

func (p *StateHelper) isSkipValidationEnabled() bool {
	return CheckLabelExists(bmenroll.SkipValidationLabel, p.bmh)
}

func (p *StateHelper) isSkipDeprovisioningEnabled() bool {
	return CheckLabelExists(bmenroll.SkipDeprovisionLabel, p.bmh)
}

func (p *StateHelper) isFeatureGroupValidationEnabled() bool {
	isEnabled := false
	if p.cfg.FeatureFlags.GroupValidation {
		// Group validation is enabled
		instanceType, err := getInstanceType(p.bmh)
		if err == nil && exists(p.cfg.FeatureFlags.EnabledGroupInstanceTypes, instanceType) {
			isEnabled = true
		}
	}
	return isEnabled
}

func (p *StateHelper) isReadyToTest() bool {
	return checkBoolLabel(bmenroll.ReadyToTestLabel, p.bmh)
}

func (p *StateHelper) markForDeletion() {
	p.addLabel(bmenroll.DeletionLabel, "true")
}

func (p *StateHelper) IsFailed() bool {
	return CheckLabelExists(bmenroll.CheckingFailedLabel, p.bmh)
}

func (p *StateHelper) TriggerValidationForFwUpgrade() {
	p.addLabel(bmenroll.ReadyToTestLabel, "true")
	p.addLabel(bmenroll.FWVersionUpdateTriggerLabel, "true")
	delete(p.bmh.Labels, bmenroll.VerifiedLabel)
	delete(p.bmh.Labels, bmenroll.CheckingFailedLabel)
}

func (p *StateHelper) isTriggeredForFwUpgrade() bool {
	return checkBoolLabel(bmenroll.FWVersionUpdateTriggerLabel, p.bmh)
}

// Update bmh labels to indicate validation is completed
func (p *StateHelper) updateValidationComplete(isFailed bool) {

	delete(p.bmh.Labels, bmenroll.CheckingCompletedLabel)
	delete(p.bmh.Labels, bmenroll.ReadyToTestLabel)
	delete(p.bmh.Labels, bmenroll.FWVersionUpdateTriggerLabel)
	if !isFailed {
		// add verified label only if validation completed successfully.
		p.addLabel(bmenroll.VerifiedLabel, "true")
	}
}

// Update bmh labels to indicate overall validation has started with imaging
func (p *StateHelper) updateValidationStarted(validationId string) {
	// clean up Verified Label if it exists
	delete(p.bmh.Labels, bmenroll.VerifiedLabel)
	//Imaging started
	p.addLabel(bmenroll.ImagingLabel, "true")
	p.addLabel(bmenroll.ValidationIdLabel, validationId)
}

func (p *StateHelper) updateAsMasterNode() {
	//Update as master Node.
	p.addLabel(bmenroll.MasterNodeLabel, "true")
}

// Update bmh label to indicate Imaging has completed
func (p *StateHelper) updateImagingCompleted() {
	delete(p.bmh.Labels, bmenroll.ImagingLabel)
	p.addLabel(bmenroll.ImagingCompletedLabel, "true")
}

func (p *StateHelper) updateFwLabel(fwVersion string) {
	p.addLabel(bmenroll.FWVersionLabel, fwVersion)
}

// Update bmh labels to indicate Checking has started
func (p *StateHelper) updateVerificationTaskStarted() {
	delete(p.bmh.Labels, bmenroll.ImagingCompletedLabel)
	p.addLabel(bmenroll.CheckingLabel, "true")
}

// Update bmh labels to indicate Group Verification has started
func (p *StateHelper) updateGroupVerificationTaskStarted() {
	delete(p.bmh.Labels, bmenroll.InstanceValidationCompletedLabel)
	p.addLabel(bmenroll.CheckingGroupLabel, "true")
}

// Update bmh labels to indicate checking step has completed
func (p *StateHelper) updateVerificationTaskCompleted(isFailed bool, msg string) {
	delete(p.bmh.Labels, bmenroll.ImagingLabel)
	delete(p.bmh.Labels, bmenroll.CheckingLabel)
	// clean up the label when operator decides to skip the validation.
	delete(p.bmh.Labels, bmenroll.ImagingCompletedLabel)
	// clean up Verified Label if it exists (This is required when the validation errors due to BMH error state)
	delete(p.bmh.Labels, bmenroll.VerifiedLabel)

	if isFailed {
		// On failure mark the checking as completed, skip the instance group verification.
		delete(p.bmh.Labels, bmenroll.MasterNodeLabel)
		delete(p.bmh.Labels, bmenroll.ValidationIdLabel)

		p.addLabel(bmenroll.CheckingCompletedLabel, "true")
		if msg == "" {
			p.addLabel(bmenroll.CheckingFailedLabel, "instanceValidationFailed")
		} else {
			p.addLabel(bmenroll.CheckingFailedLabel, msg)
		}
	} else {
		// go to the new instance validation completed state only if feature flag is set.
		if p.isInstanceGroup() && p.isFeatureGroupValidationEnabled() && !p.isSkipGroupValidationEnabled() {
			// update instance validation completed to move to STATE_INITIALIZED_INSTANCE_GROUP state.
			p.addLabel(bmenroll.InstanceValidationCompletedLabel, "true")
		} else {
			delete(p.bmh.Labels, bmenroll.ValidationIdLabel)
			// mark the checking label as completed. We directly move to the STATE_VERIFIED state.
			p.addLabel(bmenroll.CheckingCompletedLabel, "true")
		}
	}
}

// Update bmh labels to indicate group checking step has completed
func (p *StateHelper) updateGroupVerificationTaskCompleted(isFailed bool, msg string) {
	delete(p.bmh.Labels, bmenroll.InstanceValidationCompletedLabel) // this can be present in the member nodes
	delete(p.bmh.Labels, bmenroll.CheckingGroupLabel)
	delete(p.bmh.Labels, bmenroll.MasterNodeLabel)
	delete(p.bmh.Labels, bmenroll.ValidationIdLabel)
	// clean up Verified Label if it exists (This is required when the valition errors due to BMH error state)
	delete(p.bmh.Labels, bmenroll.VerifiedLabel)
	if isFailed {
		if msg == "" {
			// On failure add the checking failed label with the reason as groupValidation failure.
			p.addLabel(bmenroll.CheckingFailedLabel, "groupValidationFailed")
		} else {
			p.addLabel(bmenroll.CheckingFailedLabel, msg)
		}
	}
	p.addLabel(bmenroll.CheckingCompletedLabel, "true")

}

// add Label to bmh
func (p *StateHelper) addLabel(key, value string) {
	if p.bmh == nil || p.bmh.Labels == nil {
		return
	}
	p.bmh.Labels[key] = value
}

// check if the desired label exists
// return false if it does not exist or bmh.Labels is nil
func checkBoolLabel(key string, bmh *baremetalv1alpha1.BareMetalHost) bool {
	if bmh.Labels == nil {
		return false
	}
	value, ok := bmh.Labels[key]
	if ok {
		res, err := strconv.ParseBool(value)
		if err != nil {
			return false
		}
		return res
	}
	// label does not exist.
	return false
}
