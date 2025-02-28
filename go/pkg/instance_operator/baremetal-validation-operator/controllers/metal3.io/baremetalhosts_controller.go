// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package validation

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"

	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	baremetalv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Time after which the operator checks if the host is available to run validation tasks.
const (
	startValidationRequeueAfter    = 10 * time.Second
	checkValidationTaskStatusAfter = 30 * time.Second
	checkImageProvisioningAfter    = 30 * time.Second
	retryAfterError                = 10 * time.Second
	retryAfterVersionMismatch      = 1 * time.Second
	retryFirmwareUpgrade           = 60 * time.Second
	TimeoutBMHStateMinutes         = 180
)

const (
	NAME                          = "validationoperator"
	OptimisticLockErrorMsg        = "the object has been modified; please apply your changes to the latest version and try again"
	defaultOS                     = "ubuntu"
	FAILED_NODES_KEY              = "failedNodes"
	TEST_RESULT_KEY               = "testResult"
	IRONIC_CONFIGMAP_NAME         = "ironic-fw-update"
	IRONIC_CONFIGMAP_EVENT_PREFIX = IRONIC_CONFIGMAP_NAME + "-"
)

// Netbox Status constants
const (
	NetboxSuccess    = "Success %s"
	NetboxFailure    = "Failure %s"
	NetboxInProgress = "InProgress"
)

// BaremetalhostsReconciler reconciles Baremetalhost objects
type BaremetalhostsReconciler struct {
	client.Client
	Scheme               *runtime.Scheme
	Cfg                  *cloudv1alpha1.BmInstanceOperatorConfig
	Signer               *ssh.Signer
	ComputePrivateClient pb.InstancePrivateServiceClient
	ComputeGroupClient   pb.InstanceGroupServiceClient
	ImageFinder          *ImageFinder
	EventRecorder        record.EventRecorder
	NetBoxClient         dcim.DCIM
}

//+kubebuilder:rbac:groups=metal3.io,resources=baremetalhosts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=metal3.io,resources=baremetalhosts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=metal3.io,resources=baremetalhosts/finalizers,verbs=update
//+kubebuilder:rbac:groups=metal3.io,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=instances/finalizers,verbs=update

func (r *BaremetalhostsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BaremetalhostsReconciler.Reconcile").WithValues(logkeys.ResourceId, req.Name).Start()
	defer span.End()
	result, reconcileErr := func() (ctrl.Result, error) {
		// Fetch the BaremetalHost instance.
		bmhInstance, err := r.getBmhInstance(ctx, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if bmhInstance == nil {
			log.Info("Ignoring reconcile request because source instance was not found in cache")
			return ctrl.Result{}, nil
		}
		// Fetch current state
		stateHelper := getStateHelper(bmhInstance, r.Cfg)

		// Fetch instance Type
		instanceType, err := getInstanceType(bmhInstance)
		if err != nil {
			return ctrl.Result{}, err
		}

		if instanceType == "" {
			log.Info("Ignore reconcile request since validation of instance type is not set")
			return ctrl.Result{}, nil // ignore it
		}

		// Check if instanceType is enabled
		if !exists(r.Cfg.EnabledInstanceTypes, instanceType) {
			log.Info("Ignore reconcile request since validation of instance type is not enabled", logkeys.InstanceType, instanceType)
			stateHelper.updateValidationComplete(false)
			if err := r.updateBmh(ctx, bmhInstance); err != nil {
				return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
			}
			return ctrl.Result{}, nil
		}

		if strings.HasPrefix(req.Name, IRONIC_CONFIGMAP_EVENT_PREFIX) && r.Cfg.FeatureFlags.EnableFirmwareUpgrade {
			// the request is for updating firmware.
			if isProvisionedByValidationOperator(bmhInstance, r.Cfg.CloudAccountID) {
				log.Info("Validation is already in progress, retry firmware upgrade later")
				return ctrl.Result{
					RequeueAfter: retryFirmwareUpgrade,
				}, nil
			}
			desiredFwVersion, err := r.getDesiredFwVersion(ctx, instanceType, req.Namespace)
			if err != nil {
				log.Error(err, "Failed to fetch the desired firmware version")
				return ctrl.Result{}, err
			}
			currentFwVersion := GetLabel(bmenroll.FWVersionLabel, bmhInstance)
			if !strings.Contains(currentFwVersion, desiredFwVersion.BuildVersion) {
				stateHelper.TriggerValidationForFwUpgrade()
				if err := r.updateBmh(ctx, bmhInstance); err != nil {
					log.Error(err, "Failed to updateBmh to update firmware")
					return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
				}
				log.Info("Validation has been triggered to update firmware version")
				// Event to indicate Firmware upgrade has been triggered.
				r.EventRecorder.Event(bmhInstance, v1.EventTypeNormal, "TriggerFwUpgrade",
					"Firmware upgrade to "+desiredFwVersion.BuildVersion+" from "+currentFwVersion)
				return ctrl.Result{}, nil
			} else {
				log.Info("BMH already has updated firmware")
			}
		}

		// Skip validation only if it not for firmware upgrade
		if stateHelper.isSkipValidationEnabled() && !stateHelper.isTriggeredForFwUpgrade() {
			log.Info("Ignore reconcile request since skip validation label is present")
			stateHelper.updateValidationComplete(false)
			if err := r.updateBmh(ctx, bmhInstance); err != nil {
				return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
			}
			return ctrl.Result{}, nil
		}

		state := stateHelper.GetCurrentState()
		log.V(9).Info("Debug", logkeys.State, state, logkeys.Labels, bmhInstance.Labels)
		result, processErr := func() (ctrl.Result, error) {
			switch state {
			case STATE_BEGIN:
				//update state and trigger initialization.
				log.Info("State: Begin, new system enrolled for validation")
				if r.handledErroredBMH(ctx, bmhInstance) {
					if err := r.updateBmh(ctx, bmhInstance); err != nil {
						// Unable to update state, retry again.
						return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
					}
					r.EventRecorder.Event(bmhInstance, v1.EventTypeWarning, "InstanceValidationError",
						"Error observed "+GetLabel(bmenroll.CheckingFailedLabel, bmhInstance))
					// do not retrysince we have marked the BMH as validation failed.
					return ctrl.Result{}, nil
				}
				if !isProvisionedByValidationOperator(bmhInstance, r.Cfg.CloudAccountID) {
					if !isProvisioningAvailable(bmhInstance) {
						log.Info("Instance is not in \"Available\" state, retrying ... ")
						return ctrl.Result{RequeueAfter: startValidationRequeueAfter}, nil
					}
					log.Info("Instance is in \"Available\" state, triggering initialization.")

					instPrivate, err := r.createInstance(ctx, bmhInstance)
					if err != nil {
						if IsNonRetryable(err) {
							log.Error(err, "Failed to create instance due to a non-retryable error")
							stateHelper.updateVerificationTaskCompleted(true, "ImageMatchingFwNotFound")
							if err := r.updateBmh(ctx, bmhInstance); err != nil {
								// Unable to update state, retry again.
								return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
							}
							r.EventRecorder.Event(bmhInstance, v1.EventTypeWarning, "InstanceValidationError",
								"Error observed "+GetLabel(bmenroll.CheckingFailedLabel, bmhInstance))
							return ctrl.Result{}, nil
						}
						log.Error(err, "Failed to create instance, retrying...")
						return ctrl.Result{RequeueAfter: retryAfterError}, nil
					}
					log.Info("Instance Created", logkeys.InstancePrivateMetadata, instPrivate.Metadata, logkeys.InstancePrivateSpec, instPrivate.Spec)
				} else {
					// Instance provisioning was already triggered, just try to update the state.
					log.Info("Instance is already being provisioned due to validation operator, will update state to imaging started ")
				}
				stateHelper.updateValidationStarted("")
				if err := r.updateBmh(ctx, bmhInstance); err != nil {
					// Unable to update state, retry again.
					return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
				}
				r.EventRecorder.Event(bmhInstance, v1.EventTypeNormal, "ImagingStarted",
					"Imaging has been triggered as part of validation")
				return ctrl.Result{RequeueAfter: checkValidationTaskStatusAfter}, nil
			case STATE_BEGIN_INSTANCE_GROUP:
				log.Info("State: Begin, new instance group enrolled for validation")
				if r.handledErroredBMH(ctx, bmhInstance) {
					if err := r.updateBmh(ctx, bmhInstance); err != nil {
						// Unable to update state, retry again.
						return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
					}
					r.EventRecorder.Event(bmhInstance, v1.EventTypeWarning, "GroupValidationError",
						"Error observed "+GetLabel(bmenroll.CheckingFailedLabel, bmhInstance))
					// do not retrysince we have marked the BMH as validation failed.
					return ctrl.Result{}, nil
				}
				if !isProvisionedByValidationOperator(bmhInstance, r.Cfg.CloudAccountID) {

					isGroupAvailable, bmHosts, err := r.isGroupAvailable(ctx, bmhInstance)
					if err != nil {
						return ctrl.Result{}, err
					}
					if !isGroupAvailable {
						log.Info("Group is not available for instanceGroup provisioning, will retry again")
						return ctrl.Result{RequeueAfter: startValidationRequeueAfter}, nil
					}
					privateInstances, validationId, err := r.createInstanceGroup(ctx, &bmHosts)
					if err != nil {
						if IsNonRetryable(err) {
							log.Error(err, "Failed to create instance group due to a non-retryable error")
							stateHelper.updateGroupVerificationTaskCompleted(true, "ImageMatchingFwNotFound")
							if err := r.updateBmh(ctx, bmhInstance); err != nil {
								// Unable to update state, retry again.
								return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
							}
							r.EventRecorder.Event(bmhInstance, v1.EventTypeWarning, "GroupValidationError",
								"Error observed "+GetLabel(bmenroll.CheckingFailedLabel, bmhInstance))
							return ctrl.Result{}, nil
						}
						log.Error(err, "Failed to create instance group, retrying...")
						return ctrl.Result{RequeueAfter: retryAfterError}, nil
					}
					log.Info("Instance Group Created", logkeys.InstanceGroupName, privateInstances[0].Spec.InstanceGroup, logkeys.ValidationId, validationId)

					for _, h := range bmHosts {
						helper := getStateHelper(&h, r.Cfg)
						helper.updateValidationStarted(validationId)
						if err := r.updateBmh(ctx, &h); err != nil {
							log.Info("Failed to update validation started for host", logkeys.HostName, h.Name, logkeys.ValidationId, validationId)
						}
						r.EventRecorder.Event(&h, v1.EventTypeNormal, "InstanceGroupImaging", "Imaging has been triggered as part of validation")
					}
				} else {
					// Instance group provisioning was already triggered, just try to update the state.
					log.Info("Instance Group is already being provisioned due to validation operator, will update state to imaging started ")
					inst, err := r.getInstance(ctx, bmhInstance)
					if err != nil {
						log.Error(err, "Failed to get Instance during initialization")
						return ctrl.Result{RequeueAfter: retryAfterError}, nil
					}
					validationId, err := getValidationId(ctx, inst)
					if err != nil {
						return ctrl.Result{}, err
					}
					stateHelper.updateValidationStarted(validationId)
					if err := r.updateBmh(ctx, bmhInstance); err != nil {
						// Unable to update state, retry again.
						return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
					}
				}
				return ctrl.Result{RequeueAfter: checkValidationTaskStatusAfter}, nil

			case STATE_INTIALIZING:
				log.Info("State: System is initializing, check if Initialization state completed")
				// check if initialization is completed. if yes move to next state, else retry
				if !isProvisioned(bmhInstance) {
					// Instance provisioning in the previous step completed.
					log.Info("DEBUG: Instance is being provisioned, will check again later")
					return ctrl.Result{RequeueAfter: checkImageProvisioningAfter}, nil
				}
				// Check if instance is in ready state.
				inst, err := r.getInstance(ctx, bmhInstance)
				if err != nil {
					log.Error(err, "Failed to get Instance during initialization")
					return ctrl.Result{RequeueAfter: retryAfterError}, nil
				}
				log.Info("Instance has been provisioned", logkeys.InstanceName, inst.Name, logkeys.StatusPhase, inst.Status.Phase)
				if inst.Status.Phase == cloudv1alpha1.PhaseFailed {
					// instance is in a Failed state, update the state to indicate verification failure
					stateHelper.updateVerificationTaskCompleted(true, "Instance.creation.failed")
				} else {
					if inst.Status.Phase != cloudv1alpha1.PhaseReady {
						log.Info("Instance is not yet ready", logkeys.InstanceName, inst.Name, logkeys.StatusPhase, inst.Status.Phase)
						return ctrl.Result{RequeueAfter: checkImageProvisioningAfter}, nil
					}
					log.Info("DEBUG: Instance is ready")
					stateHelper.updateImagingCompleted()
					if fwVersions, err := r.getDesiredFwVersion(ctx, instanceType, bmhInstance.Namespace); err != nil {
						log.Error(err, "Failed to get firmware versions for the machine image", "MachineImage", inst.Spec.MachineImage)
					} else {
						// If there's no error, update the firmware label with the obtained versions
						stateHelper.updateFwLabel(fwVersions.BuildVersion)
					}
				}
				if err := r.updateBmh(ctx, bmhInstance); err != nil {
					return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
				}
				if err := r.updateNetboxValidationInProgress(ctx, bmhInstance.Name); err != nil {
					log.Error(err, "Failed to update Netbox with validation InProgress status")
				}
				// Update event only after BMH state is updated.
				if inst.Status.Phase == cloudv1alpha1.PhaseFailed {
					r.EventRecorder.Event(bmhInstance, v1.EventTypeWarning, "ImagingFailed", "ImagingFailed, marking verification as failed")
				} else {
					r.EventRecorder.Event(bmhInstance, v1.EventTypeNormal, "ImagingComplete", "Imaging has been completed and instance is in Ready phase")
				}
				return ctrl.Result{Requeue: true}, nil
			case STATE_INITIALIZED:
				// start the validation task
				log.Info("State: System is initialized, triggering validation task")
				inst, err := r.getInstance(ctx, bmhInstance)
				if err != nil {
					log.Error(err, "Failed to get Instance before triggering validation task")
					return ctrl.Result{RequeueAfter: retryAfterError}, nil
				}
				// Create a validator against the instance
				validator, err := CreateValidator(ctx, inst, bmhInstance, r.Signer, r.Cfg)
				var alreadyStarted bool = false
				if err != nil {
					log.Error(err, "Failed to create a Validator")
					if IsNonRetryable(err) {
						r.updateValidationFailed(bmhInstance, "validationArtifactNotFound")
						if err := r.updateBmh(ctx, bmhInstance); err != nil {
							// Unable to update state, retry again.
							return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
						}
						r.EventRecorder.Event(bmhInstance, v1.EventTypeWarning, "InstanceValidationError",
							"Error observed "+GetLabel(bmenroll.CheckingFailedLabel, bmhInstance))
						// do not retry
						return ctrl.Result{}, nil
					}
					return ctrl.Result{}, err
				} else {
					defer validator.Close(ctx)
					// calculate the desiredFwVersion
					desiredFwVersion, err := r.getDesiredFwVersion(ctx, validator.taskMeta.instanceType, validator.taskMeta.bmhNamespace)
					if err != nil {
						log.Error(err, "Failed to fetch the desired firmware version")
						return ctrl.Result{}, err
					}
					alreadyStarted, err = validator.startInstanceValidationTask(ctx, desiredFwVersion)
					if err != nil {
						log.Error(err, "Failed to start Validation task")
						return ctrl.Result{}, err
					}
					// Update the state to indicate verification started
					stateHelper.updateVerificationTaskStarted()
				}

				if err := r.updateBmh(ctx, bmhInstance); err != nil {
					return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
				}
				if !alreadyStarted { // do not generate duplicate events.
					r.EventRecorder.Event(bmhInstance, v1.EventTypeNormal, "InstanceValidationTaskStarted", "Validation task has started with "+
						validator.taskMeta.instanceArtifact)
				}
				// Check validation task status after the specified duration
				return ctrl.Result{RequeueAfter: checkValidationTaskStatusAfter}, nil
			case STATE_INTIALIZING_INSTANCE_GROUP:
				log.Info("State: System has completed instance validation., triggering instance group validation task")
				if !isMasterNode(bmhInstance) {
					log.Info("Not a master node, ignoring this event since validation will be done via master node.")
					return ctrl.Result{}, nil
				}
				// Check if all the instances have reached
				isGroupInitialized, masterInstance, memberIps, memberNames, err := r.isGroupInitialized(ctx, bmhInstance)
				if err != nil {
					log.Error(err, "Failed to check if instance group is Initialized")
					return ctrl.Result{RequeueAfter: retryAfterError}, nil
				}
				if !isGroupInitialized {
					log.Info("Group is not yet initialized, will retry")
					return ctrl.Result{RequeueAfter: startValidationRequeueAfter}, nil
				}
				// Create a validator against the master instance
				validator, err := CreateValidatorForGroup(ctx, masterInstance, bmhInstance, r.Signer, r.Cfg, memberIps, memberNames)
				if err != nil {
					log.Error(err, "Failed to create a Validator for instance group")
					if IsNonRetryable(err) {
						r.updateValidationFailed(bmhInstance, "validationArtifactNotFound")
						if err := r.updateBmh(ctx, bmhInstance); err != nil {
							// Unable to update state, retry again.
							return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
						}
						r.EventRecorder.Event(bmhInstance, v1.EventTypeWarning, "GroupValidationError",
							"Error observed "+GetLabel(bmenroll.CheckingFailedLabel, bmhInstance))
						// do not retry
						return ctrl.Result{}, nil
					}
					return ctrl.Result{}, err
				}
				defer validator.Close(ctx)
				// calculate the desiredFwVersion
				desiredFwVersion, err := r.getDesiredFwVersion(ctx, validator.taskMeta.instanceType, validator.taskMeta.bmhNamespace)
				if err != nil {
					log.Error(err, "Failed to fetch the desired firmware version")
					return ctrl.Result{}, err
				}
				alreadyStarted, err := validator.startGroupValidationTask(ctx, desiredFwVersion)
				if err != nil {
					log.Error(err, "Failed to start Group Validation task")
					return ctrl.Result{}, err
				}
				// Update the state
				stateHelper.updateGroupVerificationTaskStarted()
				if err := r.updateBmh(ctx, bmhInstance); err != nil {
					return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
				}
				if !alreadyStarted { // do not generate duplicate events.
					r.EventRecorder.Event(bmhInstance, v1.EventTypeNormal, "GroupValidationTaskStarted", "GroupValidation task has started with "+
						validator.taskMeta.clusterArtifact)
				}
				// Check validation task status after the specified duration
				return ctrl.Result{RequeueAfter: checkValidationTaskStatusAfter}, nil

			case STATE_VERIFYING:
				// check if verification is completed if yes move to next state, else retry
				log.Info("State: Verification in Progress, check if the Validation task has completed")
				// Fetch the instance and create a validator
				inst, err := r.getInstance(ctx, bmhInstance)
				if err != nil {
					log.Error(err, "Failed to get Instance while attempting to check validation task status")
					return ctrl.Result{RequeueAfter: retryAfterError}, nil
				}
				// Create a validator against the instance
				validator, err := CreateValidator(ctx, inst, bmhInstance, r.Signer, r.Cfg)
				if err != nil {
					log.Error(err, "Failed to create a Validator")
					return ctrl.Result{}, err
				}
				defer validator.Close(ctx)
				isComplete, taskStatus, err := validator.isValidationTaskCompleted(ctx)
				if err != nil {
					log.Error(err, "Failed to check if Validation task completed")
					return ctrl.Result{}, err
				}
				if !isComplete {
					// check again later
					return ctrl.Result{RequeueAfter: checkValidationTaskStatusAfter}, nil
				} else {
					log.Info("Validation completed", logkeys.TaskStatus, taskStatus)
					if taskStatus == FAILED {
						// Update the state of the BMH.
						stateHelper.updateVerificationTaskCompleted(true, "")
					} else {
						// Instance Validation is successful.
						if stateHelper.isInstanceGroup() && stateHelper.isFeatureGroupValidationEnabled() && !stateHelper.isSkipGroupValidationEnabled() {
							// check next master only if group validation feature flag is true and skip group validation is not enabled.
							isNextMaster, err := r.isNextMaster(ctx, bmhInstance)
							if err != nil {
								log.Error(err, "Failed to check if the current bmh can be the next master")
								return ctrl.Result{}, err
							}
							if isNextMaster {
								// No master found, mark the current node as the master.
								// The first node to complete instance validation is marked as the master.
								stateHelper.updateAsMasterNode()
							}
						}
						stateHelper.updateVerificationTaskCompleted(false, "")
					}
					stateHelper.markForDeletion() // mark bmh for deletion
					if err := r.updateBmh(ctx, bmhInstance); err != nil {
						return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
					}
					if isMasterNode(bmhInstance) {
						log.Info("Node has been marked as master node for group validation")
					}
				}
				if taskStatus == FAILED {
					r.EventRecorder.Event(bmhInstance, v1.EventTypeWarning, "InstanceValidationTaskFailed",
						"Validation task has failed")
				} else {
					r.EventRecorder.Event(bmhInstance, v1.EventTypeNormal, "InstanceValidationTaskCompleted",
						"Validation task completed successfully")
					// Clear the data once the state is updated and the event is recorded.
					validator.clearTestData(ctx)
				}
				// Requeue to complete cleanup, immediately.
				return ctrl.Result{Requeue: true}, nil
			case STATE_VERIFYING_INSTANCE_GROUP:
				clusterId := GetLabel(bmenroll.ClusterGroupID, bmhInstance)
				if clusterId == "" {
					log.Error(nil, "Node does not have cluster id, this is unexpected")
					return ctrl.Result{}, fmt.Errorf("clusterGroupId label not found for BMH %s", bmhInstance.Name)
				}
				validationId := GetLabel(bmenroll.ValidationIdLabel, bmhInstance)
				if validationId == "" {
					log.Error(nil, "Node does not have validation id, this is unexpected")
					return ctrl.Result{}, fmt.Errorf("ValidationIdLabel label not found for BMH %s", bmhInstance.Name)
				}
				// check if group verification is completed if yes move to next state, else retry
				log.Info("State: Group Verification in Progress, check if the Validation task has completed", logkeys.ClusterId, clusterId,
					logkeys.ValidationId, validationId)
				// Fetch the instance and create a validator
				inst, err := r.getInstance(ctx, bmhInstance)
				if err != nil {
					log.Error(err, "Failed to get Instance while attempting to check group validation task status")
					return ctrl.Result{RequeueAfter: retryAfterError}, nil
				}
				// Create a validator against the instance
				validator, err := CreateValidator(ctx, inst, bmhInstance, r.Signer, r.Cfg)
				if err != nil {
					log.Error(err, "Failed to create a Validator")
					return ctrl.Result{}, err
				}
				defer validator.Close(ctx)
				isComplete, taskStatus, err := validator.isValidationTaskCompleted(ctx)
				if err != nil {
					log.Error(err, "Failed to check if group Validation task completed")
					return ctrl.Result{}, err
				}
				if !isComplete {
					// check again later
					return ctrl.Result{RequeueAfter: checkValidationTaskStatusAfter}, nil
				} else {
					log.Info("Group Validation completed", logkeys.TaskStatus, taskStatus,
						logkeys.ClusterId, clusterId, logkeys.ValidationId, validationId, logkeys.InstanceName, inst.Name)

					resultMeta, err := validator.getValidationResultMeta(ctx)
					if err != nil {
						log.Error(err, "Failed to get the validation result metadata",
							logkeys.ClusterId, clusterId, logkeys.ValidationId, validationId, logkeys.InstanceName, inst.Name)
						//Log the error and continue without the result metadata.
					}

					var failedNodesList []string
					if resultMeta != nil {
						if failedNodes, ok := resultMeta[FAILED_NODES_KEY]; ok {
							log.Info("Validation result details", FAILED_NODES_KEY, failedNodes,
								logkeys.ClusterId, clusterId, logkeys.ValidationId, validationId, logkeys.InstanceName, inst.Name)
							failedNodesList = strings.Split(failedNodes, ",")
						}
					}

					// Get cluster members for the given cluster id
					bmhList, err := r.getClusterMembers(ctx, clusterId)
					if err != nil {
						return ctrl.Result{}, err
					}
					for _, h := range bmhList.Items {
						// Filter out cluster members that belong to a different validationId.
						currentValidationId := GetLabel(bmenroll.ValidationIdLabel, &h)
						if currentValidationId != validationId {
							log.Info("IgnoringBMH, this belongs to a different validation id", logkeys.HostName, h.Name,
								logkeys.ExpectedValidationId, validationId, logkeys.ObservedValidationId, currentValidationId)
							continue
						}
						if isMasterNode(&h) {
							//Skip updating master node here; it will be updated later.
							continue
						}
						status := taskStatus
						if slices.Contains(failedNodesList, h.Name) {
							status = FAILED
						}
						if err := r.updateGroupVerificationCompleted(ctx, &h, status); err != nil {
							return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
						}
					}
					status := taskStatus
					if slices.Contains(failedNodesList, bmhInstance.Name) {
						status = FAILED
					}
					// updating master node at the end to handle version mismatch errors.
					if err := r.updateGroupVerificationCompleted(ctx, bmhInstance, status); err != nil {
						return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
					}
				}
				// Requeue to complete cleanup, immediately.
				return ctrl.Result{Requeue: true}, nil
			case STATE_VERIFIED:
				// clean up and update so that the instance an be provisioned.
				log.Info("State: Verified, Clean up and ensure the instance can be provisioned by customers")

				inst, err := r.getInstance(ctx, bmhInstance)
				var resultMeta map[string]string
				if err != nil {
					if !apierrors.IsNotFound(err) {
						log.Error(err, "Error observed while attempting to fetch instance post verification, will retry")
						// retry only if it is a different from not found error
						return ctrl.Result{RequeueAfter: retryAfterError}, nil
					}
				} else { // able to fetch the instances, delete based on the feature flag
					if !r.Cfg.FeatureFlags.DeProvisionPostValidationFailure && stateHelper.IsFailed() {
						// Skip deprovisioning for failed tasks based on the Featureflag
						log.Info("Validation task failed, will skip deletion of instance")
					} else {
						validator, err := CreateValidator(ctx, inst, bmhInstance, r.Signer, r.Cfg)
						if err != nil {
							log.Error(err, "Failed to create a Validator during the verified phase")
							if !IsNonRetryable(err) { // retry only if it is a retryable error
								return ctrl.Result{}, err
							}
						} else {
							defer validator.Close(ctx)
							// attempt to fetch the validation result meta before deleting the instance.
							resultMeta = r.getValidationResultMeta(ctx, validator)
						}

						if !stateHelper.isSkipDeprovisioningEnabled() {
							log.Info("Attempting to delete the instance", logkeys.InstanceName, inst.Name)
							_, err = r.ComputePrivateClient.DeletePrivate(ctx, &pb.InstanceDeletePrivateRequest{
								Metadata: &pb.InstanceMetadataReference{
									CloudAccountId: r.Cfg.CloudAccountID,
									NameOrId: &pb.InstanceMetadataReference_ResourceId{
										ResourceId: inst.Name,
									},
								},
							})
							if err != nil && (status.Code(err) != codes.NotFound) {
								log.Error(err, "Deleting instance failed", logkeys.InstanceName, inst.Name)
								// retry only if it is a different from not found error
								return ctrl.Result{RequeueAfter: retryAfterError}, nil
							}
							r.EventRecorder.Event(bmhInstance, v1.EventTypeNormal, "CleanupTriggered",
								"Deprovision the baremetal as part of cleanup")
						} else {
							log.Info("Skipping deprovisioning of BMH", logkeys.InstanceName, inst.Name)
						}
					}
				}
				// update state indicating validation is completed
				stateHelper.updateValidationComplete(stateHelper.IsFailed())
				if err := r.updateBmh(ctx, bmhInstance); err != nil {
					return ctrl.Result{RequeueAfter: retryAfterVersionMismatch}, nil
				}

				// update the Netbox and events only if we are able to get resultMeta.
				if resultMeta != nil {
					r.EventRecorder.Event(bmhInstance, v1.EventTypeNormal, "ReportPath", resultMeta["bucket"]+resultMeta["uploadPath"])
					if err := r.updateNetboxValidationCompleted(ctx, bmhInstance.Name,
						resultMeta["uploadPath"], !stateHelper.IsFailed()); err != nil {
						log.Error(err, "Failed to update Netbox with validation completion status")
					}
					// log the results obtained from the BMH where the validation completed.
					if testResult, ok := resultMeta[TEST_RESULT_KEY]; ok {
						log.Info("Validation details", TEST_RESULT_KEY, testResult)
					}
				} else { // do not retry if Netbox updation fails.
					log.Info("Netbox was not updated with the status since we failed to fetch result metadata")
				}
				return ctrl.Result{}, nil // Validation completed return empty.
			case STATE_NOT_REQUIRED:
				log.Info("State: Validation not required or has been completed")
				return ctrl.Result{}, nil
			default:
				//nothing to do return.
				log.Info("State: Invalid phase, ignoring the reconcile event")
				return ctrl.Result{}, nil
			}
		}()
		return result, processErr
	}()
	if reconcileErr != nil {
		log.Error(reconcileErr, "InstanceReconciler.Reconcile: error reconciling Instance")
	}
	log.V(9).Info("END", logkeys.Result, result, logkeys.Error, reconcileErr)
	return result, reconcileErr
}

// Helper function to fetch the Validation result meta data. Returns a nil incase of an error.
func (r *BaremetalhostsReconciler) getValidationResultMeta(ctx context.Context, validator *Validator) map[string]string {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.getValidationResultMeta")

	resultMeta, err := validator.getValidationResultMeta(ctx)
	if err != nil {
		log.Info("Failed to get validation result meta information, ignoring it.", "error", err)
		return nil
	}
	return resultMeta
}

func (r *BaremetalhostsReconciler) updateGroupVerificationCompleted(ctx context.Context, h *baremetalv1alpha1.BareMetalHost,
	taskStatus TaskStatus) error {
	// get Statehelper
	stateHelper := getStateHelper(h, r.Cfg)
	state := stateHelper.GetCurrentState()
	if state == STATE_INTIALIZING_INSTANCE_GROUP || state == STATE_VERIFYING_INSTANCE_GROUP {
		// Update state
		if taskStatus == FAILED {
			stateHelper.updateGroupVerificationTaskCompleted(true, "")
		} else {
			stateHelper.updateGroupVerificationTaskCompleted(false, "")
		}
		if err := r.updateBmh(ctx, h); err != nil {
			return err
		}
		// Update Events
		if taskStatus == FAILED {
			r.EventRecorder.Event(h, v1.EventTypeWarning,
				"GroupValidationTaskFailed", "Validation task has failed")
		} else {
			r.EventRecorder.Event(h, v1.EventTypeNormal, "GroupValidationTaskCompleted",
				"Validation task completed successfully")
		}
	}
	return nil
}

// Update state to verification failed in case of ERROR scenario.
func (r *BaremetalhostsReconciler) handledErroredBMH(ctx context.Context, bmh *baremetalv1alpha1.BareMetalHost) bool {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.handledErroredBMH")
	// Helper function
	handleTimeout := func(startTime metav1.Time, timeoutMinutes int, operation string) bool {
		if startTime.IsZero() {
			return false
		}
		if isTimedOut(startTime.Time, timeoutMinutes) {
			errMsg := fmt.Sprintf("Timeout.waiting.for.%s.to.complete", operation)
			log.Info(fmt.Sprintf("Timeout waiting for BMH to %s", operation), logkeys.HostName, bmh.Name)
			r.updateValidationFailed(bmh, errMsg)
			return true
		}
		return false
	}
	// Check for deprovision timeout
	if bmh.Status.OperationHistory.Deprovision.End.IsZero() {
		if handleTimeout(bmh.Status.OperationHistory.Deprovision.Start, TimeoutBMHStateMinutes, "deprovision") {
			return true
		}
	}
	// Check for provision timeout
	if bmh.Status.OperationHistory.Provision.End.IsZero() {
		if handleTimeout(bmh.Status.OperationHistory.Provision.Start, TimeoutBMHStateMinutes, "provision") {
			return true
		}
	}
	return false
}

// Method to update BMH as validation failed.
func (r *BaremetalhostsReconciler) updateValidationFailed(bmh *baremetalv1alpha1.BareMetalHost, errMsg string) {
	stateHelper := getStateHelper(bmh, r.Cfg)
	if stateHelper.isInstanceGroup() {
		stateHelper.updateGroupVerificationTaskCompleted(true, errMsg)
	} else {
		stateHelper.updateVerificationTaskCompleted(true, errMsg)
	}
	stateHelper.markForDeletion()
}

// Method to update BareMetalHost
func (r *BaremetalhostsReconciler) updateBmh(ctx context.Context, bmh *baremetalv1alpha1.BareMetalHost) error {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.updateBmh")
	log.V(9).Info("Updation of bmh", logkeys.HostName, bmh.Name, logkeys.Labels, bmh.Labels)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of BMH before attempting update
		currentBmh := &baremetalv1alpha1.BareMetalHost{}
		if err := r.Get(ctx, types.NamespacedName{Name: bmh.Name, Namespace: bmh.Namespace}, currentBmh); err != nil {
			return err
		}
		// Update the labels
		currentBmh.Labels = bmh.Labels
		// Attempt to update the BMH resource
		return r.Update(ctx, currentBmh)
	})

	if err != nil {
		log.Error(err, "Failed to update BMH after retries", logkeys.HostName, bmh.Name)
		return fmt.Errorf("failed to update BMH %s after retries: %w", bmh.Name, err)
	}
	log.V(9).Info("BMH updated successfully", logkeys.HostName, bmh.Name)
	return nil
}

// Create Requests to provision an instance group for validation
func (r *BaremetalhostsReconciler) createMultiplePrivateRequests(ctx context.Context,
	bmhs *[]baremetalv1alpha1.BareMetalHost, instanceType, imageName, clusterId string) (*pb.InstanceCreateMultiplePrivateRequest, string, error) {
	hosts := *bmhs
	validationId := generateRandom8Digit()
	instanceGroupName := clusterId + "-validation-" + validationId
	cloudInit, err := r.getCloudInit(ctx)
	if err != nil {
		return nil, "", err
	}
	var requestList []*pb.InstanceCreatePrivateRequest
	for i := 0; i < len(hosts); i++ {
		resourceId, err := uuid.NewRandom()
		if err != nil {
			return nil, "", err
		}
		spec := getInstanceSpec(r, &hosts[i], instanceType, imageName)
		spec.InstanceGroup = instanceGroupName
		spec.ClusterGroupId = clusterId
		spec.UserData = cloudInit

		instanceCreatePvtReq := &pb.InstanceCreatePrivateRequest{
			Metadata: &pb.InstanceMetadataCreatePrivate{
				CloudAccountId: r.Cfg.CloudAccountID,
				ResourceId:     resourceId.String(),
				Name:           instanceGroupName + "-" + strconv.Itoa(i),
				SkipQuotaCheck: true,
			},
			Spec: spec,
		}
		requestList = append(requestList, instanceCreatePvtReq)

	}
	return &pb.InstanceCreateMultiplePrivateRequest{
		Instances: requestList,
	}, validationId, nil

}

// Create a cloud init to ensure all the instances provisioned for an instancegroup can perform passwordless ssh.
func (r *BaremetalhostsReconciler) getCloudInit(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.getCloudInit")
	privateKey, err := os.ReadFile(r.Cfg.SshConfig.PrivateKeyFilePath)
	if err != nil {
		log.Error(err, "Failed to read Private key", logkeys.KeyPath, r.Cfg.SshConfig.PrivateKeyFilePath)
		return "", err
	}
	cfg, err := util.NewEmptyCloudConfig(defaultOS)
	if err != nil {
		log.Error(err, "Failed to create CloudInit")
		return "", err
	}

	cfg.AddRunBinaryFile("/home/sdp/.ssh/id_rsa", privateKey, 0400)
	cfg.AddRunCmd("chown sdp:sdp /home/sdp/.ssh/id_rsa")
	cfg.SetRunCmd()
	cfg.SetPackages()

	cloudInitBytes, err := cfg.RenderYAML()
	if err != nil {
		log.Error(err, "Failed to render yaml cloud-init")
		return "", err
	}
	return string(cloudInitBytes), nil
}

func (r *BaremetalhostsReconciler) getDesiredFwVersion(ctx context.Context, instanceType, bmNamespace string) (*Version, error) {
	fwMap, err := r.ImageFinder.GetFirmwareVersionMap(ctx, bmNamespace)
	if err != nil {
		return nil, err
	}
	// default value
	desiredFwVersion := Version{
		BuildVersion:  "",
		SpiVersion:    "",
		FullFwVersion: "",
	}
	if fwMap != nil {
		if version, ok := fwMap.InstanceTypeFirmwareVersions[instanceType]; ok {
			desiredFwVersion = version
		}
	}
	return &desiredFwVersion, nil
}

func (r *BaremetalhostsReconciler) getLatestImage(ctx context.Context, instanceType, bmNamespace string) (string, error) {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.getLatestImage")
	desiredFwVersion, err := r.getDesiredFwVersion(ctx, instanceType, bmNamespace)
	if err != nil {
		return "", err
	}
	imageName, err := r.ImageFinder.GetLatestImage(ctx, instanceType, desiredFwVersion)
	if err != nil {
		return "", err
	}
	log.Info("Image found for instance type", logkeys.Namespace, bmNamespace, logkeys.InstanceType, instanceType, logkeys.Name, imageName,
		"desiredFirmwareVersion", desiredFwVersion)
	return imageName, nil
}

// Create an instance for the BMH using the compute private APIs
func (r *BaremetalhostsReconciler) createInstanceGroup(ctx context.Context,
	bmhs *[]baremetalv1alpha1.BareMetalHost) ([]*pb.InstancePrivate, string, error) {

	hosts := *bmhs
	clusterId := GetLabel(bmenroll.ClusterGroupID, &hosts[0])
	instanceType, err := getInstanceType(&hosts[0])
	if err != nil {
		return nil, "", err
	}
	imageName, err := r.getLatestImage(ctx, instanceType, hosts[0].Namespace)
	if err != nil {
		return nil, "", err
	}

	multipleReq, validationId, err := r.createMultiplePrivateRequests(ctx, bmhs, instanceType, imageName, clusterId)
	if err != nil {
		return nil, "", err
	}
	resp, err := r.ComputePrivateClient.CreateMultiplePrivate(ctx, multipleReq)
	if err != nil {
		return nil, "", err
	}
	// validate if instances count is non-zero
	if len(resp.Instances) == 0 {
		return nil, "", fmt.Errorf("CreateMultiplePrivate api call returned no instances")
	}
	return resp.Instances, validationId, nil
}

// Create an instance for the BMH using the compute private APIs
func (r *BaremetalhostsReconciler) createInstance(ctx context.Context,
	bmh *baremetalv1alpha1.BareMetalHost) (*pb.InstancePrivate, error) {
	resourceId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	instanceType, err := getInstanceType(bmh)
	if err != nil {
		return nil, err
	}
	imageName, err := r.getLatestImage(ctx, instanceType, bmh.Namespace)
	if err != nil {
		return nil, err
	}
	resp, err := r.ComputePrivateClient.CreateMultiplePrivate(ctx, &pb.InstanceCreateMultiplePrivateRequest{
		Instances: []*pb.InstanceCreatePrivateRequest{
			{
				Metadata: &pb.InstanceMetadataCreatePrivate{
					CloudAccountId: r.Cfg.CloudAccountID,
					ResourceId:     resourceId.String(),
					Name:           bmh.Name + "-validation",
					SkipQuotaCheck: true,
				},
				Spec: getInstanceSpec(r, bmh, instanceType, imageName),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	// validate if instances count is non-zero
	if len(resp.Instances) == 0 {
		return nil, fmt.Errorf("CreateMultiplePrivate api call returned no instances")
	}
	return resp.Instances[0], nil
}

// Fetch cloudv1alpha1.Instance corresponding to the BMH.
func (r *BaremetalhostsReconciler) getInstance(ctx context.Context,
	bmh *baremetalv1alpha1.BareMetalHost) (*cloudv1alpha1.Instance, error) {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.getInstance")
	inst := cloudv1alpha1.Instance{}
	if bmh.Spec.ConsumerRef == nil {
		log.Info("ConsumerRef of BMH is empty, no corresponding instance for BMH")
		return nil, apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: "Instance"}, bmh.Name)
	}
	key := client.ObjectKey{
		Namespace: bmh.Spec.ConsumerRef.Namespace,
		Name:      bmh.Spec.ConsumerRef.Name,
	}
	err := r.Client.Get(ctx, key, &inst)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// Get BMH instance from K8s.
// Returns (nil, nil) if not found.
func (r *BaremetalhostsReconciler) getBmhInstance(ctx context.Context, req ctrl.Request) (*baremetalv1alpha1.BareMetalHost, error) {
	log := log.FromContext(ctx).WithName("getBmhInstance")
	instance := &baremetalv1alpha1.BareMetalHost{}
	name := strings.TrimPrefix(req.Name, IRONIC_CONFIGMAP_EVENT_PREFIX)
	err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      name,
	}, instance)
	if apierrors.IsNotFound(err) || reflect.ValueOf(instance).IsZero() {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getInstance: %w", err)
	}
	log.V(9).Info("Debug:", "labels", instance.Labels, "generation", instance.Generation)
	return instance, nil
}

func getValidationId(ctx context.Context, inst *cloudv1alpha1.Instance) (string, error) {
	log := log.FromContext(ctx).WithName("getValidationId")
	if inst.Spec.InstanceGroup == "" {
		return "", fmt.Errorf("InstanceGroup is empty during group validation for %s", inst.Spec.NodeId)
	} else {
		split := strings.Split(inst.Spec.InstanceGroup, "-")
		validationId := split[len(split)-1]
		log.Info("Validation Details", logkeys.ValidationId, validationId, logkeys.InstanceNodeId, inst.Spec.NodeId, logkeys.InstanceName, inst.Name)
		return validationId, nil
	}
}

// List of ips of the other group members
func (r *BaremetalhostsReconciler) isGroupInitialized(ctx context.Context,
	bmh *baremetalv1alpha1.BareMetalHost) (bool, *cloudv1alpha1.Instance, *[]string, *[]string, error) {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.isGroupInitialized")
	clusterId := GetLabel(bmenroll.ClusterGroupID, bmh)
	if clusterId == "" {
		log.Error(nil, "Node does not have cluster id, this is unexpected")
		return false, nil, nil, nil, fmt.Errorf("clusterGroupId label not found for BMH %s", bmh.Name)
	}
	// Get cluster members for the given cluster id
	bmhList, err := r.getClusterMembers(ctx, clusterId)
	if err != nil {
		return false, nil, nil, nil, err
	}
	validationId := GetLabel(bmenroll.ValidationIdLabel, bmh)
	if validationId == "" {
		log.Error(nil, "Node does not have validation id, this is unexpected")
		return false, nil, nil, nil, fmt.Errorf("ValidationIdLabel label not found for BMH %s", bmh.Name)
	}
	var filteredHosts []baremetalv1alpha1.BareMetalHost
	for _, h := range bmhList.Items {
		log.Info("Processing", logkeys.HostName, h.Name)
		// Ignore bmh that do not require to be validated
		if !CheckLabelExists(bmenroll.ReadyToTestLabel, &h) {
			log.Info("Ignoring BMH, ReadyToTestLabel does not exist", logkeys.HostName, h.Name)
			continue
		}
		currentValidationId := GetLabel(bmenroll.ValidationIdLabel, &h)
		if currentValidationId != validationId {
			log.Info("IgnoringBMH, this belongs to a different validation id", logkeys.HostName, h.Name, logkeys.ExpectedValidationId, validationId,
				logkeys.ObservedValidationId, currentValidationId)
			continue
		}
		stateHelper := getStateHelper(&h, r.Cfg)
		state := stateHelper.GetCurrentState()
		if state == STATE_INTIALIZING_INSTANCE_GROUP || state == STATE_VERIFYING_INSTANCE_GROUP {
			// STATE_VERIFYING_INSTANCE_GROUP can be observed during a retry due to version mismatch.
			filteredHosts = append(filteredHosts, h)
		} else if state == STATE_VERIFYING || state == STATE_INITIALIZED || state == STATE_INTIALIZING {
			log.Info("InstanceGroup is not ready since an instance is still in a previous state", logkeys.HostName, h.Name, logkeys.State, state,
				logkeys.ValidationId, validationId)
			return false, nil, nil, nil, nil
		} else if state == STATE_BEGIN_INSTANCE_GROUP {
			log.Info("Instance is waiting for the current validation to complete", logkeys.HostName, h.Name, logkeys.ValidationId, validationId)
		} else {
			log.Error(nil, "Unexpected state observed", logkeys.HostName, h.Name, logkeys.State, state, logkeys.ValidationId, validationId)
			return false, nil, nil, nil, fmt.Errorf("unexpected state observed, will retry again")
		}
	}
	if len(filteredHosts) == 0 {
		return false, nil, nil, nil, fmt.Errorf("zero hosts observed, this is can happen only if the BMH are externally modified")
	}

	// grab the master instance and the memberIps
	var masterInstance *cloudv1alpha1.Instance
	var memberIPs []string
	var memberBMHs []string
	for _, h := range filteredHosts {

		inst, err := r.getInstance(ctx, &h)
		if err != nil {
			log.Error(err, "Failed to get Instance before triggering group validation task")
			return false, nil, nil, nil, err
		}
		if isMasterNode(&h) {
			masterInstance = inst
		} else {
			memberIPs = append(memberIPs, inst.Status.Interfaces[0].Addresses[0])
			memberBMHs = append(memberBMHs, h.Name)
		}
	}
	log.Info("Group of BMHs that are ready for interconnect validation", logkeys.Hosts, extractBMHNames(filteredHosts),
		logkeys.ValidationId, validationId)

	return true, masterInstance, &memberIPs, &memberBMHs, nil
}

// Check if BMH is available.
func (r *BaremetalhostsReconciler) isGroupAvailable(ctx context.Context,
	bmh *baremetalv1alpha1.BareMetalHost) (bool, []baremetalv1alpha1.BareMetalHost, error) {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.isGroupAvailable")
	var hosts []baremetalv1alpha1.BareMetalHost
	if isProvisioningAvailable(bmh) {
		// check if other bmhs are available
		clusterId := GetLabel(bmenroll.ClusterGroupID, bmh)
		if clusterId == "" {
			return false, nil, fmt.Errorf("clusterGroupId label not found for BMH %s", bmh.Name)
		}

		// Init a label selector and add the requirement
		hostList, err := r.getClusterMembers(ctx, clusterId)
		if err != nil {
			return false, nil, err
		}
		// Check if all the hosts are available.
		for _, h := range hostList.Items {
			log.Info("Processing", logkeys.HostName, h.Name)

			if isProvisionedByValidationOperator(&h, r.Cfg.CloudAccountID) {
				log.Info("Host is already being validated, ignoring it", logkeys.HostName, h.Name)
				continue
			}
			if isProvisioned(&h) || isProvisioning(&h) {
				log.Info("Ignoring BMH since it is provisioned/provisioning", logkeys.HostName, h.Name)
				continue
			}
			if CheckLabelExists(bmenroll.VerifiedLabel, &h) {
				log.Info("Ignoring BMH, already verified", logkeys.HostName, h.Name)
				continue
			}

			if !CheckLabelExists(bmenroll.ReadyToTestLabel, &h) {
				log.Info("Ignoring BMH, ReadyToTestLabel does not exist", logkeys.HostName, h.Name)
				continue
			}
			if CheckLabelExists(bmenroll.ImagingLabel, &h) {
				log.Info("Imaging has started by the validation operator, retry after sometime")
				continue
			}
			if !isProvisioningAvailable(&h) {
				// A bmh that is bound for testing is not in an available state, wait until it is available.
				log.Info("BMH that requires validation and is part of a cluster is not in an Available state, wait until it is available",
					logkeys.ClusterId, clusterId, logkeys.HostName, h.Name, logkeys.ProvisioningState, h.Status.Provisioning.State)
				return false, nil, nil
			}
			if CheckLabelExists(bmenroll.GateValidationLabel, &h) {
				// This flag gives the ability to wait until all the nodes are ready to start group validation.
				log.Info("Validation is gated for this cluster, retry after sometime")
				return false, nil, nil
			}
			hosts = append(hosts, h)
		}
		if len(hosts) < 2 {
			log.Info("Number of hosts for group validation is less than 2, retry after sometime")
			return false, nil, nil
		}
		log.Info("Group of BMHs available for validation ", logkeys.Hosts, extractBMHNames(hosts))
		return true, hosts, nil
	} else {
		log.Info("Provisioning state of the BMH is not available, retry after some time", logkeys.HostName, bmh.Name)
		return false, nil, nil
	}
}

// Function to get the names of BMHs
func extractBMHNames(hosts []baremetalv1alpha1.BareMetalHost) []string {
	var hostNames []string
	for _, h := range hosts {
		hostNames = append(hostNames, h.Name)
	}
	return hostNames
}

func (r *BaremetalhostsReconciler) getClusterMembers(ctx context.Context, clusterId string) (*baremetalv1alpha1.BareMetalHostList, error) {
	hostList := &baremetalv1alpha1.BareMetalHostList{}
	matchClusterIdReq, err := labels.NewRequirement(bmenroll.ClusterGroupID, selection.Equals, []string{clusterId})
	if err != nil {
		return hostList, fmt.Errorf("failed to create a requirement %w", err)
	}

	labelselector := labels.NewSelector()
	labelselector.Add(*matchClusterIdReq)

	opts := []client.ListOption{
		client.MatchingLabels{
			bmenroll.ClusterGroupID: clusterId,
		},
	}
	err = r.Client.List(ctx, hostList, opts...)
	return hostList, err
}

// Select the next master for group validation.
// If the group has only one member the next master is not chosen.
func (r *BaremetalhostsReconciler) isNextMaster(ctx context.Context, bmh *baremetalv1alpha1.BareMetalHost) (bool, error) {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.isNextMaster")
	clusterId := GetLabel(bmenroll.ClusterGroupID, bmh)
	if clusterId == "" {
		log.Error(nil, "Node does not have cluster id, this is unexpected")
		return false, fmt.Errorf("instance-group-id label not found for BMH %s", bmh.Name)
	}
	validationId := GetLabel(bmenroll.ValidationIdLabel, bmh)
	if validationId == "" {
		log.Error(nil, "Node does not have validation id, this is unexpected")
		return false, fmt.Errorf("validation-id label not found for BMH %s", bmh.Name)
	}
	// This returns a list of cluster members
	bmhList, err := r.getClusterMembers(ctx, clusterId)

	if err != nil {
		return false, err
	}

	isMasterAbsent := true
	for _, h := range bmhList.Items {
		if validationId == GetLabel(bmenroll.ValidationIdLabel, &h) {
			// hosts with the same validation id should be used for this
			stateHelper := getStateHelper(&h, r.Cfg)
			state := stateHelper.GetCurrentState()
			if state == STATE_INTIALIZING || state == STATE_INITIALIZED || state == STATE_INTIALIZING_INSTANCE_GROUP ||
				state == STATE_VERIFYING {
				if isMasterNode(&h) {
					isMasterAbsent = false
					break
				}
			}
		}
	}
	return isMasterAbsent, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BaremetalhostsReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	// Add an index for the metadata.name field
	if err := mgr.GetFieldIndexer().IndexField(ctx, &v1.ConfigMap{}, "metadata.name", func(rawObj client.Object) []string {
		configMap := rawObj.(*v1.ConfigMap)
		return []string{configMap.Name}
	}); err != nil {
		return err
	}
	// Watch for changes to ConfigMaps with the name "ironic-fw-update"
	configMapPredicate := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false // no event should be generated
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectNew.GetName() == IRONIC_CONFIGMAP_NAME
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false // no event should be generated
		},
		// GenericEvent can be triggered by tooling, currently disabled.
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
	// Map function to transform ConfigMap events to BaremetalHost reconcile requests
	mapFn := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, cfgObj client.Object) []ctrl.Request {
		return r.generateReconcileRequests(ctx, cfgObj)
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(&baremetalv1alpha1.BareMetalHost{}, builder.WithPredicates(predicate.LabelChangedPredicate{})).
		Watches(&v1.ConfigMap{}, mapFn, builder.WithPredicates(configMapPredicate)).
		WatchesRawSource(&source.Channel{Source: r.startupChannel(ctx, mgr)}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

// Create a Generic event to ensure we check all BMH on startup
func (r *BaremetalhostsReconciler) startupChannel(ctx context.Context, mgr ctrl.Manager) <-chan event.GenericEvent {
	log := log.FromContext(ctx).WithName("startupChannel")
	ch := make(chan event.GenericEvent)

	go func() {
		defer close(ch)
		// Wait for the cache to sync
		if !mgr.GetCache().WaitForCacheSync(ctx) {
			log.Error(nil, "Failed to wait for cache to sync")
		}
		var configMapList v1.ConfigMapList
		if err := r.List(ctx, &configMapList, client.MatchingFields{"metadata.name": IRONIC_CONFIGMAP_NAME}); err != nil {
			log.Error(err, "Unable to list ConfigMaps")
			return
		}

		for _, cfgObj := range configMapList.Items {
			requests := r.generateReconcileRequests(ctx, &cfgObj)
			for _, req := range requests {
				log.Info("Enqueueing Firmware upgrade reconcile request for BMH", "namespace", req.Namespace, "hostName", req.Name)
				// This will ensure the same event queue is used to check for firmware version on startup
				ch <- event.GenericEvent{
					Object: &baremetalv1alpha1.BareMetalHost{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: req.Namespace,
							Name:      req.Name,
						},
					},
				}
			}
		}
	}()

	return ch
}

func (r *BaremetalhostsReconciler) generateReconcileRequests(ctx context.Context, cfgObj client.Object) []ctrl.Request {
	log := log.FromContext(ctx).WithName("generateReconcileRequests")

	// List all BaremetalHosts in the same namespace as the ConfigMap
	var bmhList baremetalv1alpha1.BareMetalHostList
	var requests []ctrl.Request
	if err := r.List(ctx, &bmhList, client.InNamespace(cfgObj.GetNamespace())); err != nil {
		log.Error(err, "Unable to list BaremetalHosts in namespace", "namespace", cfgObj.GetNamespace())
		return requests
	}

	for _, bmh := range bmhList.Items {
		if isProvisioningAvailable(&bmh) ||
			isProvisionedByValidationOperator(&bmh, r.Cfg.CloudAccountID) ||
			isDeprovisioningByValidationOperator(&bmh) {
			// Generate reconcile events
			requests = append(requests, ctrl.Request{
				NamespacedName: client.ObjectKey{
					Namespace: bmh.Namespace,
					Name:      IRONIC_CONFIGMAP_EVENT_PREFIX + bmh.Name,
				},
			})
		}
	}
	return requests
}

// Update Netbox with validation Completed.
func (r *BaremetalhostsReconciler) updateNetboxValidationCompleted(ctx context.Context, deviceName,
	validationS3URL string, isValidationSuccessful bool) error {
	currentUTCTime := time.Now().UTC()
	// Add UTC timestamp to the status.
	validationStatus := fmt.Sprintf(NetboxSuccess, currentUTCTime)
	if !isValidationSuccessful {
		validationStatus = fmt.Sprintf(NetboxFailure, currentUTCTime)
	}

	return r.updateNetboxStatus(ctx, deviceName, validationS3URL, validationStatus)
}

// Update Netbox with validation InProgress.
func (r *BaremetalhostsReconciler) updateNetboxValidationInProgress(ctx context.Context, deviceName string) error {
	return r.updateNetboxStatus(ctx, deviceName, "", NetboxInProgress)
}

func (r *BaremetalhostsReconciler) updateNetboxStatus(ctx context.Context, deviceName,
	validationS3URL string, validationStatus string) error {
	log := log.FromContext(ctx).WithName("BaremetalhostsReconciler.updateNetboxStatus")
	log.V(9).Info("Fetching device id", logkeys.DeviceName, deviceName)
	deviceId, err := r.NetBoxClient.GetDeviceId(ctx, deviceName)
	if err != nil {
		return err
	}

	log.Info("Attempting to update Netbox", logkeys.DeviceName, deviceName, logkeys.DeviceId, deviceId, logkeys.ValidationStatus, validationStatus, logkeys.ValidationS3Bucket, r.Cfg.ValidationReportS3Config.BucketName,
		logkeys.ValidationS3URL, validationS3URL)
	customFields := dcim.DeviceCustomFields{
		BMValidationStatus:    validationStatus,
		BMValidationReportURL: "--",
	}
	if validationS3URL != "" {
		customFields.BMValidationReportURL = r.Cfg.ValidationReportS3Config.CloudfrontPrefix + validationS3URL + "/validation_logs.tar.gz"
	}

	if err := r.NetBoxClient.UpdateBMValidationStatus(ctx, deviceId, deviceName, customFields); err != nil {
		return fmt.Errorf("unable to update NetBox device status: %v", err)
	}
	return nil
}

// Helper methods

// Method to create InstanceSpec
func getInstanceSpec(r *BaremetalhostsReconciler, bmh *baremetalv1alpha1.BareMetalHost,
	instanceType, imageName string) *pb.InstanceSpecPrivate {

	spec := &pb.InstanceSpecPrivate{
		AvailabilityZone: r.Cfg.EnvConfiguration.AvailabilityZone,
		ClusterId:        bmh.Namespace,
		NodeId:           bmh.Name,
		InstanceType:     instanceType,
		MachineImage:     imageName,
		RunStrategy:      pb.RunStrategy_RerunOnFailure,
		SshPublicKeyNames: []string{
			NAME,
		},
		Interfaces: []*pb.NetworkInterfacePrivate{
			{
				Name: "eth0",
				VNet: r.Cfg.EnvConfiguration.AvailabilityZone + "-" + NAME,
			},
		},
	}
	// Get the network mode.
	networkMode := GetLabel(bmenroll.NetworkModeLabel, bmh)
	if networkMode == bmenroll.NetworkModeXBX {
		// Set the network mode to NetworkModeIgnore to skip accelator vnet creation.
		// This should happpen only for XBX network mode type.
		networkMode = bmenroll.NetworkModeIgnore
	}
	spec.NetworkMode = networkMode

	return spec
}

func isMasterNode(bmh *baremetalv1alpha1.BareMetalHost) bool {
	return CheckLabelExists(bmenroll.MasterNodeLabel, bmh)
}

func isProvisioningAvailable(bmh *baremetalv1alpha1.BareMetalHost) bool {
	return bmh.Status.Provisioning.State == baremetalv1alpha1.StateAvailable
}

func isDeprovisioningByValidationOperator(bmh *baremetalv1alpha1.BareMetalHost) bool {
	return bmh.Status.Provisioning.State == baremetalv1alpha1.StateDeprovisioning &&
		CheckLabelExists(bmenroll.VerifiedLabel, bmh)
}

// Check if BMH is provisioned.
func isProvisioned(bmh *baremetalv1alpha1.BareMetalHost) bool {
	return bmh.Status.Provisioning.State == baremetalv1alpha1.StateProvisioned
}

// Check if BMH is provisioning.
func isProvisioning(bmh *baremetalv1alpha1.BareMetalHost) bool {
	return bmh.Status.Provisioning.State == baremetalv1alpha1.StateProvisioning
}

// Check if BMH is being provisioned by Validation operator.
func isProvisionedByValidationOperator(bmh *baremetalv1alpha1.BareMetalHost, cloudAccountId string) bool {
	return bmh.Spec.ConsumerRef != nil && bmh.Spec.ConsumerRef.Namespace == cloudAccountId
}

// Helper method to get SSH Keys
func GetSSHKeys(ctx context.Context, publicKeyPath string) (string, error) {
	hostPublicKeyByte, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("unable to read host public key file %v: %w", publicKeyPath, err)
	}

	// Remove line breaks from the public key string
	hostPublicKeyString := strings.ReplaceAll(string(hostPublicKeyByte), "\n", "")

	// Remove semi colon from the public key string. A semicolon seems to get added while parsing the contents of host_public_key file
	// followed by subsequent conversion to string especially in ssh key verification of bm instance controller while using localhost
	// as bastion server for testing.
	// Ex: ssh-keyscan -t rsa localhost | awk '{print $2, $3}'> local/secrets/ssh-proxy-operator/host_public_key
	hostPublicKeyString = strings.ReplaceAll(hostPublicKeyString, ";", "")

	return hostPublicKeyString, nil
}
