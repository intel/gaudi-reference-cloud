// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
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

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/hashicorp/go-multierror"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ddi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ipacmd"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mygofish"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/myssh"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets"
	helper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/units"
	"github.com/sethvargo/go-password/password"
)

// BMEnrollmentReconciler reconciles a BMEnrollment object
type BMEnrollmentReconciler struct {
	client.Client
	Cfg                       *privatecloudv1alpha1.BMEnrollmentOperatorConfig
	DDI                       ddi.DDI
	InstanceTypeServiceClient pb.InstanceTypeServiceClient
	NetBox                    dcim.DCIM
	Scheme                    *runtime.Scheme
	Vault                     secrets.SecretManager
}

type skipDisenrollment struct {
	enrollmentMsg string
}

func (e *skipDisenrollment) Error() string {
	return e.enrollmentMsg
}

type skipEnrollment struct {
	enrollmentMsg string
}

func (e *skipEnrollment) Error() string {
	return e.enrollmentMsg
}

//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=bmenrollments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=bmenrollments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=private.cloud.intel.com,resources=bmenrollments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BMEnrollment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (reconciler *BMEnrollmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.Reconcile").WithValues("resource", req.NamespacedName)
	log.Info("BEGIN")
	defer log.Info("END")

	bmEnrollment := &privatecloudv1alpha1.BMEnrollment{}
	err := reconciler.Get(ctx, req.NamespacedName, bmEnrollment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("BM enrollment resource not found. Ignoring")
			return ctrl.Result{}, nil
		}
		log.Info("Failed to get BM enrollment resource. Re-running reconcile.")
		return ctrl.Result{}, err
	}
	// enroll a node
	var processErr error
	var result reconcile.Result
	var skipEnrollmentErr *skipEnrollment
	var skipDisenrollmentErr *skipDisenrollment

	if bmEnrollment.ObjectMeta.DeletionTimestamp.IsZero() {
		// insert finalizer
		if err := reconciler.insertFinalizerIfMissing(ctx, bmEnrollment); err != nil {
			return ctrl.Result{}, err
		}
		// process Enrollment
		result, processErr = reconciler.processEnrollment(ctx, bmEnrollment, req)
		// Update BM enrollment status
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			if !errors.As(processErr, &skipEnrollmentErr) {
				processErr = multierror.Append(processErr, err)
			}
		}
		// Change processErr to nil if error type is skipEnrollment
		if errors.As(processErr, &skipEnrollmentErr) {
			log.Info("Setting error value to nil")
			processErr = nil
		}
		return result, processErr

	} else {
		// process Disenrollment
		result, processErr = reconciler.processDisenrollment(ctx, bmEnrollment, req)

		// Update BM enrollment status
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			if !errors.As(processErr, &skipDisenrollmentErr) {
				processErr = multierror.Append(processErr, err)
			}
		}
		// remove finalizer
		if (processErr == nil || !errors.As(processErr, &skipDisenrollmentErr)) && result.IsZero() {
			if err := reconciler.removeFinalizerIfExists(ctx, bmEnrollment, req); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Change processErr to nil if error type skipDisenrollment
		if errors.As(processErr, &skipDisenrollmentErr) {
			log.Info("Setting error value to nil")
			processErr = nil
		}
		return result, processErr
	}
}

func (reconciler *BMEnrollmentReconciler) insertFinalizerIfMissing(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.insertFinalizerIfMissing").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	if !controllerutil.ContainsFinalizer(bmEnrollment, EnrollmentFinalizer) {
		log.Info("Inserting finalizer", "finalizer", EnrollmentFinalizer)
		controllerutil.AddFinalizer(bmEnrollment, EnrollmentFinalizer)
		if err := reconciler.Update(ctx, bmEnrollment); err != nil {
			log.Error(err, "Failed to insert enrollment finalizer", "finalizer", EnrollmentFinalizer)
			return err
		}
		log.Info("Inserted finalizer", "finalizer", EnrollmentFinalizer)
	}
	return nil
}

func (reconciler *BMEnrollmentReconciler) processEnrollment(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.processEnrollment").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	// set conditions
	reconciler.setInitialEnrollmentConditions(bmEnrollment)
	// check if enrollment phase is set to failed
	if enrollmentFailed(bmEnrollment) {
		log.Info("enrollment is failed with error", "error", bmEnrollment.Status.ErrorMessage)
		return ctrl.Result{}, nil
	}
	// set starting conditions to True
	for _, condition := range bmEnrollment.Status.Conditions {
		// skip failed condition in process enrollment
		if condition.Type == privatecloudv1alpha1.BMEnrollmentConditionFailed {
			continue
		}
		result, err, task := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error, privatecloudv1alpha1.BMEnrollmentConditionType) {
			// NOTE:
			// When a task below returns ctrl.Result{}, nil, continue to the next task.
			// requeue or return error otherwise

			// enrollment pre checks
			if condition.Type == privatecloudv1alpha1.BMEnrollmentConditionPreEnrollmentChecks {
				result, err := reconciler.requireEnrollment(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionPreEnrollmentChecks
				}
			}
			// start enrollment task
			if startBMEnrollment(condition) {
				if err := reconciler.startBMEnrollment(ctx, bmEnrollment, req); err != nil {
					log.Info("failed to start the enrollment task")
					return ctrl.Result{}, err, privatecloudv1alpha1.BMEnrollmentConditionStarting
				}
			}
			// create BMC interface
			if getBMCInterface(condition) {
				result, err := reconciler.getBMCInterface(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface
				}
			}
			// update BMC configuration
			if updateBMCConfig(condition) {
				result, err := reconciler.updateBMCConfig(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionUpdateBMCConfig
				}
			}
			// enroll BMH host
			if enrollBareMetalHost(condition) {
				result, err := reconciler.enrollBareMetalHostToIronic(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionBMHStarting
				}
			}
			// register BMH host
			if registerBareMetalHost(condition) {
				result, err := reconciler.registerBareMetalHost(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionBMHRegistering
				}
			}
			// inspect BMH host
			if inspectBareMetalHost(condition) {
				result, err := reconciler.inspectBareMetalHost(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionBMHInspecting
				}
			}
			// provision BMH host
			if provisionBareMetalHost(condition) {
				result, err := reconciler.provisionBareMetalHost(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionBMHProvisioning
				}
			}
			// deprovision BMH host
			if deprovisionBareMetalHost(condition) {
				result, err := reconciler.deprovisionBareMetalHost(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionBMHDeprovisioning
				}
			}
			// complete BMH enrollment
			if validateBareMetalHostEnrollment(condition) {
				result, err := reconciler.validateBareMetalHostIronicEnrollment(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionBMHEnrolled
				}
			}
			// Add BMH Labels
			if addBareMetalHostLabels(condition) {
				result, err := reconciler.addBareMetalHostLabels(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionAddLabels
				}
			}
			// complete enrollment
			if completedEnrollment(condition) {
				result, err := reconciler.completeEnrollment(ctx, bmEnrollment, req)
				if err != nil || !result.IsZero() {
					return result, err, privatecloudv1alpha1.BMEnrollmentConditionCompleted
				}
			}
			return ctrl.Result{}, nil, ""
		}(ctx, bmEnrollment, req)

		if err != nil {
			log.Info("enrollment task failed", "task", task)
			return result, err
		}
		if !result.IsZero() {
			log.Info("requeueing enrollment task", "task", task)
			return result, nil
		}
	}
	return ctrl.Result{RequeueAfter: EnrollmentPeriodicRequeueAfter}, nil
}

// pre enrollment requirements checks
func (reconciler *BMEnrollmentReconciler) requireEnrollment(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.requireEnrollment").WithValues("deviceName", bmEnrollment.Spec.DeviceName)

	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionPreEnrollmentChecks)
	if condition.Status == v1.ConditionTrue {
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionPreEnrollmentChecks, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonPreEnrollmentChecksStarted, privatecloudv1alpha1.BMEnrollmentMessagePreEnrollmentChecksStarted, true)
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionPreEnrollmentChecks)
	}
	// timeout if pre checks are not completed within the timeout duration
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentGeneralTimeout)
	if err != nil {
		return result, err
	}

	result, processErr, status, reason, message := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (
		reconcile.Result, error, v1.ConditionStatus, privatecloudv1alpha1.ConditionReason, string) {

		// Enrollment namespace is not available yet, continue with enrollment
		if !bareMetalHostNamespaceAssigned(bmEnrollment) {
			log.Info("Target namespace is not assigned for the BareMetalHost. Continuing with the enrollment")
			return ctrl.Result{}, nil, v1.ConditionTrue,
				privatecloudv1alpha1.BMEnrollmentConditionReasonPreEnrollmentChecksCompleted,
				privatecloudv1alpha1.BMEnrollmentMessagePreEnrollmentChecksCompleted
		}
		// continue with the enrollment if BareMetalHost custom resource is not created
		bmHost, err := reconciler.getBareMetalHost(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("BareMetalHost is not found. Continuing with the enrollment")
				return ctrl.Result{}, nil, v1.ConditionTrue,
					privatecloudv1alpha1.BMEnrollmentConditionReasonPreEnrollmentChecksCompleted,
					privatecloudv1alpha1.BMEnrollmentMessagePreEnrollmentChecksCompleted
			} else {
				return ctrl.Result{}, err, v1.ConditionFalse,
					privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH,
					privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH
			}
		}
		// skip enrollment if baremetal host is consumed
		if !enrollmentCompleted(bmEnrollment) && bareMetalHostConsumed(bmHost) {
			log.Info("Skips enrollment as BareMetalHost is consumed")
			return ctrl.Result{},
				&skipEnrollment{enrollmentMsg: privatecloudv1alpha1.BMEnrollmentMessageConsumedBMH},
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonConsumedBMH,
				privatecloudv1alpha1.BMEnrollmentMessageConsumedBMH
		}
		// continue with enrollment when Completed condition is set to false
		if !enrollmentCompleted(bmEnrollment) {
			log.Info("Enrollment is in progress")
			return ctrl.Result{}, nil, v1.ConditionTrue,
				privatecloudv1alpha1.BMEnrollmentConditionReasonPreEnrollmentChecksCompleted,
				privatecloudv1alpha1.BMEnrollmentMessagePreEnrollmentChecksCompleted
		}
		return ctrl.Result{}, nil, v1.ConditionTrue,
			privatecloudv1alpha1.BMEnrollmentConditionReasonPreEnrollmentChecksCompleted,
			privatecloudv1alpha1.BMEnrollmentMessagePreEnrollmentChecksCompleted
	}(ctx, bmEnrollment)
	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionPreEnrollmentChecks, status, reason, message, false)
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("Enrollment pre checks completed")
	return result, processErr
}

// start enrollment of a node
func (reconciler *BMEnrollmentReconciler) startBMEnrollment(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.startBMEnrollment").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	log.Info("starting BM enrollment", "deviceName", bmEnrollment.Spec.DeviceName)
	// starting Enrollment task.
	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionStarting, v1.ConditionTrue,
		privatecloudv1alpha1.BMEnrollmentConditionReasonTaskStarted, privatecloudv1alpha1.BMEnrollmentMessageStarting, true)
	bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseStarting
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return err
	}
	return nil
}

// getBMCInterface:
// tasks:
// get BMC IP, MAC and credentials
// validate default BMC credentials
// create new BMC credentials and store it in vault
// create BMC interface with the new credentials
// validate new BMC credentials
func (reconciler *BMEnrollmentReconciler) getBMCInterface(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.getBMCInterface").WithValues("deviceName", bmEnrollment.Spec.DeviceName)

	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface)
	// set start time
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("starting get BMC interface task", "deviceName", bmEnrollment.Spec.DeviceName)
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonStartGetBMCInterface, privatecloudv1alpha1.BMEnrollmentMessageGetBMCInterface, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseGetBMCInterface
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface)
	}
	// timeout if task completed within the timeout duration
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentGeneralTimeout)
	if err != nil || !result.IsZero() {
		return result, err
	}
	result, processErr, status, reason, message := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (
		reconcile.Result, error, v1.ConditionStatus, privatecloudv1alpha1.ConditionReason, string) {
		// Get BMC MAC address
		if emptyBMCMACAddress(bmEnrollment) && reconciler.DDI != nil {
			log.Info("Getting BMC MAC Address")
			bmcMACAddress, err := reconciler.NetBox.GetBMCMACAddress(ctx, bmEnrollment.Spec.DeviceName, BMCInterfaceName)
			if bmcMACAddress == "" || err != nil {
				//ToDo: Update NetBox
				if err != nil {
					log.Error(err, "unable to get BMC MACAddress with error. Requeueing")
				} else {
					log.Info("BMC MAC address is not found. Empty string.. Requeueing")
				}
				return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
					privatecloudv1alpha1.BMEnrollmentConditionReasonFailedBMCMACAddress,
					privatecloudv1alpha1.BMEnrollmentMessageFailedBMCMACAddress
			}
			// update MAC Address in the status
			bmEnrollment.Status.BMC.MACAddress = bmcMACAddress
		}
		// Get BMC URL
		if emptyBMCURL(bmEnrollment) {
			bmcURL, err := reconciler.getBMCURL(ctx, bmEnrollment)
			if bmcURL == "" || err != nil {
				//ToDo: Update NetBox
				if err != nil {
					log.Error(err, "unable to get BMC URL with error. Requeueing")
				} else {
					log.Info("BMC URL is not found. Empty string.. Requeueing")
				}
				return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
					privatecloudv1alpha1.BMEnrollmentConditionReasonFailedBMCURLAddress,
					privatecloudv1alpha1.BMEnrollmentMessageFailedBMCURLAddress
			}
			// update bmc URL in the status
			bmEnrollment.Status.BMC.Address = bmcURL
		}

		// validate default bmc credentials, create and validate new bmc credentials
		if emptyEnrollmentBMCSecret(bmEnrollment) || createNewBMCUser(bmEnrollment) {
			// Get BMC default credentials
			log.Info("Getting default BMC credentials")
			bmcUsername, bmcPassword, err := reconciler.getDefaultBMCCredentials(ctx, bmEnrollment)
			if err != nil {
				log.Error(err, "failed to get default BMC credentials. Requeueing")
				return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
					privatecloudv1alpha1.BMEnrollmentConditionReasonFailedGetBMCCredentials,
					privatecloudv1alpha1.BMEnrollmentMessageFailedDefaultBMCCredentials
			}
			enrollmentSecret := reconciler.newBareMetalHostSecret(bmEnrollment, bmcUsername, bmcPassword, bmEnrollment.Namespace)
			if err = reconciler.createEnrollmentBMCSecret(ctx, enrollmentSecret, bmEnrollment); err != nil {
				log.Error(err, "failed to create k8s secret with BMC credentials. Requeueing")
				return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
					privatecloudv1alpha1.BMEnrollmentConditionReasonFailedCreateBMCSecret,
					privatecloudv1alpha1.BMEnrollmentMessageFailedCreateBMCSecret
			}
			// update BMC secret in the status
			bmEnrollment.Status.BMC.SecretName = enrollmentSecret.Name

			var bmcInterface bmc.Interface
			// create and validate BMC Interface with default BMC credentials
			if bmcInterface, err = reconciler.createBMCInterface(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface, bmcUsername, bmcPassword); err != nil {
				log.Error(err, "failed to create new BMC Interface. Requeueing")
				enrollmentCondition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface)
				return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
					enrollmentCondition.Reason, enrollmentCondition.Message
			}
			hwType := bmcInterface.GetHwType()
			bmEnrollment.Status.BMC.HardwareType = hwType.String()

			// generate new bmc user credentials, store it in vault and validate new credentials
			if bmcInterface.IsVirtual() || hwType == bmc.Gaudi2Wiwynn {
				log.Info("cannot create or change user accounts. ignoring BMC creds update")
				bmEnrollment.Status.BMC.CreateNewBMCUser = privatecloudv1alpha1.CreateBMCUserNotSupported
			} else {
				bmEnrollment.Status.BMC.CreateNewBMCUser = privatecloudv1alpha1.CreateBMCUserFailed
				// generate new BMC credentials
				newUsername, newPassword, err := reconciler.generateBMCCredentials(ctx)
				if err != nil {
					return reconcile.Result{}, fmt.Errorf("unable to generate new BMC creds: %v", err), v1.ConditionFalse,
						privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGenerateBMCCredentials,
						privatecloudv1alpha1.BMEnrollmentMessageFailedToGenerateBMCCredentials
				}
				// store new bmc user credentials in vault
				if err = reconciler.storeUserBMCCredentialsInVault(ctx, bmEnrollment, newUsername, newPassword); err != nil {
					log.Error(err, "unable to write to new bmc user credentials to vault. Requeueing")
					return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
						privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToStoreUserBMCCredentials,
						privatecloudv1alpha1.BMEnrollmentMessageFailedToStoreUserBMCCredentials
				}
				// store new bmc user in the host
				if createErr := reconciler.updateBMCCredentialsInRedfish(ctx, bmcInterface, newUsername, newPassword); createErr != nil {
					log.Info("Failed to update BMC Credentials, removing entry from Vault", "error", createErr)
					if err = reconciler.deleteBMCCredentialsFromVault(ctx, bmEnrollment, BMCUserSecretsPrefix); err != nil {
						log.Error(err, "failed to delete bmc credentials from vault. Requeueing")
						return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
							privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToDeleteBMCCredentials,
							privatecloudv1alpha1.BMEnrollmentMessageFailedToDeleteBMCCredentials
					}
					log.Error(err, "unable to update the new BMC user credentials in the host. Requeueing")
					return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
						privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToUpdateUserBMCCredentials,
						privatecloudv1alpha1.BMEnrollmentMessageFailedToUpdateUserBMCCredentials
				}
				// create and validate BMC Interface with user BMC credentials
				if _, err = reconciler.createBMCInterface(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface, newUsername, newPassword); err != nil {
					log.Error(err, "failed to create new BMC Interface. Requeueing")
					enrollmentCondition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface)
					return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
						enrollmentCondition.Reason, enrollmentCondition.Message
				}
				// Update secret username and password
				if err := reconciler.patchSecret(ctx, bmEnrollment, newUsername, newPassword); err != nil {
					log.Error(err, "failed to patch bmc user k8s secret with the new credentials. Requeueing")
					return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
						privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToUpdateUserBMCCredentials,
						privatecloudv1alpha1.BMEnrollmentMessageFailedToUpdateUserBMCCredentials
				}
				bmEnrollment.Status.BMC.CreateNewBMCUser = privatecloudv1alpha1.CreateBMCUserPassed
			}
		}
		log.Info("get BMC interface task completed")
		return reconcile.Result{}, nil, v1.ConditionTrue,
			privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedGetBMCInterface,
			privatecloudv1alpha1.BMEnrollmentMessageCompletedGetBMCMACAddress
	}(ctx, bmEnrollment)

	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionGetBMCInterface, status, reason, message, false)
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	return result, processErr
}

// updateBMCConfig:
// tasks:
// get boot MAC Address
// Sanitize boot order
// configure NTP
// verify PFR
// enable KCS
// enable HCI
func (reconciler *BMEnrollmentReconciler) updateBMCConfig(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.updateBMCConfig").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionUpdateBMCConfig)
	// set start time
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("start updating BMC Configuration", "deviceName", bmEnrollment.Spec.DeviceName)
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionUpdateBMCConfig, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonStartUpdatingBMCConfig, privatecloudv1alpha1.BMEnrollmentMessageStartUpdatingBMCConfig, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseUpdateBMCConfig
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionUpdateBMCConfig)
	}
	// timeout if task completed within the timeout duration
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentGeneralTimeout)
	if err != nil || !result.IsZero() {
		return result, err
	}
	// update BMC config
	result, processErr, status, reason, message := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (
		reconcile.Result, error, v1.ConditionStatus, privatecloudv1alpha1.ConditionReason, string) {
		bmcUsername, bmcPassword, err := reconciler.getSecretData(ctx, bmEnrollment)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to get BMC secret data: %v", err), v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMCCredentialsK8sSecretData,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMCCredentialsK8sSecretData
		}

		// create and validate BMC Interface
		var bmcInterface bmc.Interface
		if bmcInterface, err = reconciler.createBMCInterface(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionUpdateBMCConfig, bmcUsername, bmcPassword); err != nil {
			log.Error(err, "failed to create new BMC Interface. Requeueing")
			enrollmentCondition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionUpdateBMCConfig)
			return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
				enrollmentCondition.Reason, enrollmentCondition.Message
		}
		// get boot MAC address
		if emptyBootAddress(bmEnrollment) {
			bootMACAddress, err := reconciler.getBootMacAddress(ctx, bmEnrollment, bmcInterface)
			if err != nil {
				log.Error(err, "failed to get boot MAC address. Requeueing")
				return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
					privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBootMACAddress,
					privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBootMACAddress
			}
			bmEnrollment.Status.BootMACAddress = bootMACAddress
		}
		// Get BMC Address in metal3 format
		if emptyMetal3Address(bmEnrollment) {
			bmcMetal3Address, err := bmcInterface.GetHostBMCAddress()
			if err != nil {
				log.Error(err, "failed to get BMC address in metal3 supported format. Requeueing")
				return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
					privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetMetal3BMCAddress,
					privatecloudv1alpha1.BMEnrollmentMessageFailedToGetMetal3BMCAddress
			}
			bmEnrollment.Status.BMC.Metal3Address = bmcMetal3Address
		}

		// sanitize BMC boot order, set NTP and verify PFR
		if err = bmcInterface.SanitizeBMCBootOrder(ctx); err != nil {
			log.Error(err, "unable to update BMC Boot Order. Requeueing")
			return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToUpdateBMCBootOrder,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToUpdateBMCBootOrder
		}
		if err = bmcInterface.ConfigureNTP(ctx); err != nil {
			log.Error(err, "unable to update BMC NTP. Requeueing")
			return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToUpdateBMCNtp,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToUpdateBMCNtp
		}
		if err = bmcInterface.VerifyPlatformFirmwareResilience(ctx); err != nil {
			log.Error(err, "unable to verify the Platform Firmware Resilience. Requeueing")
			return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToVerifyBMCPfr,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToVerifyBMCPfr
		}
		// Enable KCS
		if enableKCS(bmEnrollment) {
			if err = bmcInterface.EnableKCS(ctx); err != nil {
				if errors.Is(err, bmc.ErrKCSNotSupported) {
					log.Info("Cannot enable KCS, skipping.")
					bmEnrollment.Status.BMC.KCS = privatecloudv1alpha1.KCSNotSupported
				} else {
					bmEnrollment.Status.BMC.KCS = privatecloudv1alpha1.KCSDisabled
					log.Error(err, "unable to enable KCS. Requeueing")
					return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
						privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToEnableKCS,
						privatecloudv1alpha1.BMEnrollmentMessageFailedToEnableKCS
				}
			} else {
				bmEnrollment.Status.BMC.KCS = privatecloudv1alpha1.KCSEnabled
			}
		}
		// Enable HCI
		if enableHCI(bmEnrollment) {
			if err = bmcInterface.EnableHCI(ctx); err != nil {
				if errors.Is(err, bmc.ErrHCINotSupported) {
					log.Info("Cannot enable HCI, skipping.")
					bmEnrollment.Status.BMC.HCI = privatecloudv1alpha1.HCINotSupported
				} else {
					bmEnrollment.Status.BMC.HCI = privatecloudv1alpha1.HCIDisabled
					log.Error(err, "unable to enable HCI. Requeueing")
					return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
						privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToEnableHCI,
						privatecloudv1alpha1.BMEnrollmentMessageFailedToEnableHCI
				}
			} else {
				bmEnrollment.Status.BMC.HCI = privatecloudv1alpha1.HCIEnabled
			}
		}
		log.Info("update BMC config task completed")
		return reconcile.Result{}, nil, v1.ConditionTrue,
			privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedUpdatingBMCConfig,
			privatecloudv1alpha1.BMEnrollmentMessageCompletedUpdatingBMCConfig
	}(ctx, bmEnrollment)

	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionUpdateBMCConfig, status, reason, message, false)
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	return result, processErr
}

// enroll BMH
// tasks:
// get target namespace
// get ironic IP
// create PXE record
// create BareMetalHost
// create BMH secret
func (reconciler *BMEnrollmentReconciler) enrollBareMetalHostToIronic(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.enrollBareMetalHost").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHStarting)
	// set start time
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("start creating the BareMetalHost", "deviceName", bmEnrollment.Spec.DeviceName)
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHStarting, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonStartBMHEnrolling, privatecloudv1alpha1.BMEnrollmentMessageStartBMHEnrolling, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseEnrolling
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHStarting)
	}
	// timeout if the enrollment condition failed.
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentGeneralTimeout)
	if err != nil || !result.IsZero() {
		return result, err
	}
	result, processErr, status, reason, message := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (
		reconcile.Result, error, v1.ConditionStatus, privatecloudv1alpha1.ConditionReason, string) {
		// get target namespace
		targetNamespace, err := reconciler.assignBareMetalHostNamespace(ctx)
		if err != nil {
			log.Error(err, "failed to get a target namespace. Requeueing")
			return reconcile.Result{RequeueAfter: GeneralRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetTargetNamespace,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetTargetNamespace
		}
		log.Info("target namespace", "namespace", targetNamespace)
		bmEnrollment.Status.TargetBmNamespace = targetNamespace.Name

		// get ironic IP
		ironicIP, exists := targetNamespace.ObjectMeta.Labels[Metal3NamespaceIronicIPKey]
		if !exists {
			log.Error(err, "failed to get the ironic IP. Requeueing", "namespace", targetNamespace.Name)
			return reconcile.Result{RequeueAfter: GeneralRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetIronicIP,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetIronicIP
		}
		log.Info("ironic IP address", "IP address", ironicIP)
		bmEnrollment.Status.IronicIPAddress = ironicIP

		// create PXE record
		err = reconciler.createPxeRecord(ctx, bmEnrollment)
		if err != nil {
			log.Error(err, "unable to create mac record in dhcp server. Requeueing")
			return reconcile.Result{RequeueAfter: GeneralRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToCreatePXERecord,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToCreatePXERecord
		}
		// create BareMetalHost
		err = reconciler.createBareMetalHost(ctx, bmEnrollment)
		if err != nil {
			log.Error(err, "failed to create BareMetalHost. Requeueing")
			return reconcile.Result{RequeueAfter: GeneralRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToCreateBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToCreateBMH
		}

		// create baremetal secret
		err = reconciler.createBMHSecret(ctx, bmEnrollment)
		if err != nil {
			log.Error(err, "failed to create BareMetalHost BMC secret. Requeueing")
			return reconcile.Result{RequeueAfter: GeneralRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToCreateBMHSecret,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToCreateBMHSecret
		}
		log.Info("enroll BareMetalHost task is completed")
		return ctrl.Result{}, nil, v1.ConditionTrue,
			privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedCreatingBMH,
			privatecloudv1alpha1.BMEnrollmentMessageCompletedCreatingBMH
	}(ctx, bmEnrollment)

	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHStarting, status, reason, message, false)
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	return result, processErr
}

// Register BareMetalHost
func (reconciler *BMEnrollmentReconciler) registerBareMetalHost(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.registerBareMetalHost").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHRegistering)
	// set start time
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("start BareMetalHost registration task", "deviceName", bmEnrollment.Spec.DeviceName)
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHRegistering, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonStartRegisteringBMH, privatecloudv1alpha1.BMEnrollmentMessageStartRegisteringBMH, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseEnrolling
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHRegistering)
	}
	// timeout if registration is not completed within the timeout duration
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentRegistrationTimeout)
	if err != nil || !result.IsZero() {
		return result, err
	}
	// register BareMetalHost
	result, processErr, status, reason, message := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (
		reconcile.Result, error, v1.ConditionStatus, privatecloudv1alpha1.ConditionReason, string) {
		// get BMH
		bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get %s during BareMetalHost registration: %v", bmEnrollment.Spec.DeviceName, err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH
		}
		// check for errors
		if bareMetalHostHasError(bmHost) {
			return ctrl.Result{}, fmt.Errorf("BareMetalHost registration failed with error: %v", bmHost.Status.ErrorMessage),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToRegisterBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToRegisterBMH
		}
		// requeue if BMH provisioning state is none or registering
		if bareMetalHostRegistrationInProgress(bmHost) {
			log.Info("BareMetalHost registration is in progress. Requeuing")
			return ctrl.Result{RequeueAfter: BMHRegisterRequeueAfter}, nil,
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonBMHRegistrationInProgress,
				privatecloudv1alpha1.BMEnrollmentMessageBMHRegistrationInProgress
		}

		// set registration status to completed when provisioning state is set to inspecting
		if bareMetalHostRegistrationCompleted(bmHost) {
			log.Info("BareMetalHost registration is completed")
			return ctrl.Result{}, nil, v1.ConditionTrue,
				privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedRegisteringBMH,
				privatecloudv1alpha1.BMEnrollmentMessageCompletedRegisteringBMH

		}
		return ctrl.Result{RequeueAfter: GeneralRequeueAfter}, nil, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonBMHRegistrationInProgress,
			privatecloudv1alpha1.BMEnrollmentMessageBMHRegistrationInProgress
	}(ctx, bmEnrollment)

	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHRegistering, status, reason, message, false)
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	return result, processErr
}

// Inspect BareMetalHost
func (reconciler *BMEnrollmentReconciler) inspectBareMetalHost(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.inspectBareMetalHost").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHInspecting)
	// set start time
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("start BareMetalHost inspection task", "deviceName", bmEnrollment.Spec.DeviceName)
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHInspecting, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonStartInspectingBMH, privatecloudv1alpha1.BMEnrollmentMessageStartInspectingBMH, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseEnrolling
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHInspecting)
	}
	// timeout if inspection is not completed within the timeout duration
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentInspectionTimeout)
	if err != nil || !result.IsZero() {
		return result, err
	}
	// inspect BareMetalHost
	result, processErr, status, reason, message := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (
		reconcile.Result, error, v1.ConditionStatus, privatecloudv1alpha1.ConditionReason, string) {
		// get BMH
		bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get %s during BareMetalHost inspection: %v", bmEnrollment.Spec.DeviceName, err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH
		}
		// check for errors
		if bareMetalHostHasError(bmHost) {
			return ctrl.Result{}, fmt.Errorf("BareMetalHost inspection failed with error: %v", bmHost.Status.ErrorMessage),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToInspectBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToInspectBMH
		}
		// requeue if BMH provisioning state is inspecting
		if bareMetalHostInspectionInProgress(bmHost) {
			log.Info("BareMetalHost inspection is in progress. Requeuing")
			return ctrl.Result{RequeueAfter: BMHInspectionRequeueAfter}, nil,
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonBMHInspectionInProgress,
				privatecloudv1alpha1.BMEnrollmentMessageBMHInspectionInProgress
		}
		// set inspection status to completed when provisioning state is set to available
		if bareMetalHostInspectionCompleted(bmHost) {
			log.Info("BareMetalHost inspection is completed")
			return ctrl.Result{}, nil, v1.ConditionTrue,
				privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedInspectingBMH,
				privatecloudv1alpha1.BMEnrollmentMessageCompletedInspectingBMH
		}
		return ctrl.Result{RequeueAfter: GeneralRequeueAfter}, nil, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonBMHInspectionInProgress,
			privatecloudv1alpha1.BMEnrollmentMessageBMHInspectionInProgress
	}(ctx, bmEnrollment)

	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHInspecting, status, reason, message, false)
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	return result, processErr
}

// Provisioning BareMetalHost
func (reconciler *BMEnrollmentReconciler) provisionBareMetalHost(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.provisionBareMetalHost").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHProvisioning)
	// set start time
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("start BareMetalHost provisioning task", "deviceName", bmEnrollment.Spec.DeviceName)
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHProvisioning, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonStartProvisioningBMH, privatecloudv1alpha1.BMEnrollmentMessageStartProvisioningBMH, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseEnrolling
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHProvisioning)
	}
	// timeout if provisioning is not completed within the timeout duration
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentProvisioningTimeout)
	if err != nil || !result.IsZero() {
		return result, err
	}
	//provision BareMetalHost
	result, processErr, status, reason, message := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (
		reconcile.Result, error, v1.ConditionStatus, privatecloudv1alpha1.ConditionReason, string) {
		// get BMH
		bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get %s during BareMetalHost provisioning: %v", bmEnrollment.Spec.DeviceName, err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH
		}
		// patch BareMetalHost with image to trigger provisioning
		imageData := map[string]interface{}{
			"spec": map[string]interface{}{
				"image": &baremetalv1alpha1.Image{
					URL:      fmt.Sprintf("http://%s:%s/images/cirros-disk.img", bmEnrollment.Status.IronicIPAddress, IronicHttpPortNb),
					Checksum: fmt.Sprintf("http://%s:%s/images/cirros-disk.img.md5sum", bmEnrollment.Status.IronicIPAddress, IronicHttpPortNb),
				},
			},
		}
		patch, err := json.Marshal(imageData)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to marshal BareMetalHost image data for patching: %v", err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToMarshalBMHImageData,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToMarshalBMHImageData
		}
		if bareMetalHostAvailable(bmHost) && emptyBareMetalHostImage(bmHost) {
			log.Info("BareMetalHost is available, patching BareMetalHost image to trigger provisioning")
			if err = reconciler.patchBareMetalHostImage(ctx, bmEnrollment, patch); err != nil {
				log.Error(err, "failed to Patch BareMetalHost")
				return ctrl.Result{RequeueAfter: BMHProvisionRequeueAfter}, nil, v1.ConditionFalse,
					privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToPatchBMH,
					privatecloudv1alpha1.BMEnrollmentMessageFailedToPatchBMH
			}
		}
		// get BMH after patching
		bmHost, err = reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get %s during BareMetalHost provisioning(after patching): %v", bmEnrollment.Spec.DeviceName, err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH
		}
		// check for errors
		if bareMetalHostHasError(bmHost) {
			return ctrl.Result{}, fmt.Errorf("BareMetalHost provisioning failed with error: %v", bmHost.Status.ErrorMessage),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToProvisionBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToProvisionBMH
		}
		// requeue if BMH state is set to provisioning
		if bareMetalHostProvisioningInProgress(bmHost) {
			log.Info("BareMetalHost provisioning is in progress. Requeuing")
			return ctrl.Result{RequeueAfter: BMHProvisionRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonBMHProvisioningInProgress,
				privatecloudv1alpha1.BMEnrollmentMessageBMHProvisioningInProgress
		}
		// set inspection status to completed when provisioning state is set to provisioned
		if bareMetalHostProvisioningCompleted(bmHost) {
			log.Info("BareMetalHost provisioning is completed")
			return ctrl.Result{}, nil, v1.ConditionTrue,
				privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedProvisioningBMH,
				privatecloudv1alpha1.BMEnrollmentMessageCompletedProvisioningBMH
		}
		return ctrl.Result{RequeueAfter: GeneralRequeueAfter}, nil, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonBMHProvisioningInProgress,
			privatecloudv1alpha1.BMEnrollmentMessageBMHProvisioningInProgress
	}(ctx, bmEnrollment)

	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHProvisioning, status, reason, message, false)
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	return result, processErr
}

// Deprovision BareMetalHost
func (reconciler *BMEnrollmentReconciler) deprovisionBareMetalHost(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.deprovisionBareMetalHost").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHDeprovisioning)
	// set start time
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("start BareMetalHost deprovisioning task", "deviceName", bmEnrollment.Spec.DeviceName)
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHDeprovisioning, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonStartDeprovisioningBMH, privatecloudv1alpha1.BMEnrollmentMessageStartDeprovisioningBMH, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseEnrolling
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHDeprovisioning)
	}
	// timeout if deprovisioning is not completed within the timeout duration
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentDeprovisioningTimeout)
	if err != nil || !result.IsZero() {
		return result, err
	}
	//deprovision BareMetalHost
	result, processErr, status, reason, message := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (
		reconcile.Result, error, v1.ConditionStatus, privatecloudv1alpha1.ConditionReason, string) {
		// get BMH
		bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get %s during BareMetalHost deprovisioning: %v", bmEnrollment.Spec.DeviceName, err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH
		}
		// patch BM host and remove image to trigger deprovisioning
		if bareMetalHostProvisioned(bmHost) {
			log.Info("BareMetalHost is provisioned. patching BareMetalHost image to trigger deprovisioning")
			patch := []byte(`{"spec":{"image": null}}`)
			if err = reconciler.patchBareMetalHostImage(ctx, bmEnrollment, patch); err != nil {
				log.Error(err, "failed to Patch BareMetalHost")
				return ctrl.Result{RequeueAfter: BMHDeprovisionRequeueAfter}, nil, v1.ConditionFalse,
					privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToPatchBMH,
					privatecloudv1alpha1.BMEnrollmentMessageFailedToPatchBMH
			}
		}
		// get BMH after patching to trigger deprovisioning
		bmHost, err = reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get %s during BareMetalHost deprovisioning(after patching): %v", bmEnrollment.Spec.DeviceName, err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH
		}
		// check for errors
		if bareMetalHostHasError(bmHost) {
			return ctrl.Result{}, fmt.Errorf("BareMetalHost deprovisioning failed with error: %v", bmHost.Status.ErrorMessage),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToDeprovisionBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToDeprovisionBMH
		}
		// requeue if BMH provisioning state is deprovisioning
		if bareMetalHostDeprovisioningInProgress(bmHost) {
			log.Info("BareMetalHost deprovisioning is in progress. Requeuing")
			return ctrl.Result{RequeueAfter: BMHDeprovisionRequeueAfter}, nil, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonBMHDeprovisioningInProgress,
				privatecloudv1alpha1.BMEnrollmentMessageBMHDeprovisioningInProgress
		}
		// set deprovisioning status to completed when provisioning state is set to available or preparing
		if bareMetalHostDeprovisioningCompleted(bmHost) {
			log.Info("BareMetalHost deprovisioning is completed")
			return ctrl.Result{}, nil, v1.ConditionTrue,
				privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedDeprovisioningBMH,
				privatecloudv1alpha1.BMEnrollmentMessageCompletedDeprovisioningBMH

		}
		return ctrl.Result{RequeueAfter: GeneralRequeueAfter}, nil, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonBMHDeprovisioningInProgress,
			privatecloudv1alpha1.BMEnrollmentMessageBMHDeprovisioningInProgress
	}(ctx, bmEnrollment)
	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHDeprovisioning, status, reason, message, false)
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	return result, processErr
}

// Validating BareMetalHost enrollment
func (reconciler *BMEnrollmentReconciler) validateBareMetalHostIronicEnrollment(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.completeBareMetalHostEnrollment").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHEnrolled)
	// set start time
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("BareMetalHost enrollment validation task", "deviceName", bmEnrollment.Spec.DeviceName)
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHEnrolled, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonStartBMHEnrollmentValidation,
			privatecloudv1alpha1.BMEnrollmentMessageStartBMHEnrollmentValidation, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseEnrolling
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionBMHEnrolled)
	}
	// time out
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentGeneralTimeout)
	if err != nil || !result.IsZero() {
		return result, err
	}
	// get BMH
	bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
	if err != nil {
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHEnrolled, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH, privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH, false)
		return ctrl.Result{}, fmt.Errorf("failed to get %s during BareMetalHost enrollment validation: %v", bmEnrollment.Spec.DeviceName, err)
	}
	// validate if BareMetalHost is successfully enrolled
	if bareMetalHostEnrolled(bmHost) {
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHEnrolled, v1.ConditionTrue,
			privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedBMHEnrollmentValidation,
			privatecloudv1alpha1.BMEnrollmentMessageCompletedBMHEnrollmentValidation, false)
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		} else {
			log.Info("BareMetalHost ironic enrollment validation is completed")
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{RequeueAfter: GeneralRequeueAfter}, nil
}

// Add required BareMetalHost labels
// set BIOS password(Intel platforms)
func (reconciler *BMEnrollmentReconciler) addBareMetalHostLabels(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.addBareMetalHostLabels").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionAddLabels)
	// set start time
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("starting BareMetalHost label task", "deviceName", bmEnrollment.Spec.DeviceName)
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionAddLabels, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonAddLabels, privatecloudv1alpha1.BMEnrollmentMessageAddLabels, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseBMHLabels
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionAddLabels)
	}
	// timeout if adding labels is not completed within the timeout duration
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentGeneralTimeout)
	if err != nil || !result.IsZero() {
		return result, err
	}
	// add required labels and annotations
	result, processErr, status, reason, message := func(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (
		reconcile.Result, error, v1.ConditionStatus, privatecloudv1alpha1.ConditionReason, string) {
		// Get BMC credentials
		bmcUsername, bmcPassword, err := reconciler.getSecretData(ctx, bmEnrollment)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to get BMC secret data: %v", err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMCCredentialsK8sSecretData,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMCCredentialsK8sSecretData
		}

		// create and validate BMC Interface
		bmcInterface, err := reconciler.createBMCInterface(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionAddLabels, bmcUsername, bmcPassword)
		if err != nil {
			enrollmentCondition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionAddLabels)
			return reconcile.Result{RequeueAfter: BMCTaskRequeueAfter}, nil, v1.ConditionFalse,
				enrollmentCondition.Reason, enrollmentCondition.Message
		}

		// get BMH
		bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get BareMetalHost %s when trying to get host IP %v", bmEnrollment.Spec.DeviceName, err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH
		}

		// Get BMH IP address
		bmHostIP, err := reconciler.findBareMetalHostIP(ctx, bmHost)
		if err != nil {
			return reconcile.Result{}, err, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetHostIP,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToGetHostIP
		}
		bmEnrollment.Status.HostIPAddress = bmHostIP

		// set BareMetalHost storage annotations
		if err = reconciler.addStorageAnnotations(ctx, bmEnrollment, bmcInterface); err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to add storage annotations to the BareMetalHost: %v", err), v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToAddStorageAnnotations,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToAddStorageAnnotations
		}

		// set BareMetalHost hardware labels
		if err = reconciler.updateHostHardwareLabels(ctx, bmEnrollment, bmcInterface); err != nil {
			return reconcile.Result{}, fmt.Errorf("failed to add hardware labels and annotations to the BareMetalHost: %v", err),
				v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToAddHardwareLabels,
				privatecloudv1alpha1.BMEnrollmentMessageFailedToAddHardwareLabels
		}
		log.Info("Added required labels and annotations to the BareMetalHost")
		return reconcile.Result{}, nil, v1.ConditionTrue,
			privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedLabels,
			privatecloudv1alpha1.BMEnrollmentMessageCompletedLabels
	}(ctx, bmEnrollment)
	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionAddLabels, status, reason, message, false)
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	return result, processErr
}

func (reconciler *BMEnrollmentReconciler) completeEnrollment(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.completeEnrollment").WithValues("deviceName", bmEnrollment.Spec.DeviceName)

	log.Info("completed enrollment", "deviceName", bmEnrollment.Spec.DeviceName)
	SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionCompleted, v1.ConditionTrue,
		privatecloudv1alpha1.BMEnrollmentConditionReasonCompletedEnrollment, privatecloudv1alpha1.BMEnrollmentMessageCompletedEnrollment, true)
	bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseReady
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// create new BMC Interface and validate
func (reconciler *BMEnrollmentReconciler) createBMCInterface(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment,
	conditionType privatecloudv1alpha1.BMEnrollmentConditionType, username string, password string) (bmc.Interface, error) {
	// get BMC interface with new bmc user credentials
	bmcInterface, err := bmc.New(&mygofish.MyGoFishManager{}, &bmc.Config{
		URL:      bmEnrollment.Status.BMC.Address,
		Username: username,
		Password: password,
	})
	if err != nil {
		SetBMEnrollmentCondition(ctx, bmEnrollment, conditionType, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonFailedNewBMCInterface, privatecloudv1alpha1.BMEnrollmentMessageFailedNewBMCInterface, false)
		return nil, fmt.Errorf("failed to create new BMC interface of user %s. error: %v", username, err)
	}

	hwType := bmcInterface.GetHwType()
	// verify BMC user credentials
	if hwType != bmc.Gaudi2Wiwynn {
		if err = reconciler.verifyBMCCredentials(ctx, bmEnrollment, bmcInterface); err != nil {
			SetBMEnrollmentCondition(ctx, bmEnrollment, conditionType, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedVerifyBMCCredentials, privatecloudv1alpha1.BMEnrollmentMessageFailedVerifyBMCCredentials, false)
			return nil, fmt.Errorf("failed to verify BMC credentials of user %s. error: %v", username, err)
		}
	}
	return bmcInterface, nil
}

// check if the enrollment condition is Timed out
func (reconciler *BMEnrollmentReconciler) isEnrollmentConditionTimedOut(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment,
	enrollmentTask privatecloudv1alpha1.BMEnrollmentConditionType, req ctrl.Request, timeout time.Duration) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.isEnrollmentTaskTimedOut").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, enrollmentTask)
	timeOfEnrollmentTimeOut := condition.StartTime.Add(timeout)
	if timeOfEnrollmentTimeOut.Before(time.Now()) && !enrollmentCompleted(bmEnrollment) {
		errMessage := fmt.Sprintf("%s. Task: %s", privatecloudv1alpha1.BMEnrollmentMessageEnrollmentTaskTimedOut, enrollmentTask)
		log.Error(fmt.Errorf(errMessage), errMessage, "timeOfRegistrationTimeOut", timeOfEnrollmentTimeOut,
			"current time", time.Now())
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionFailed, v1.ConditionTrue,
			privatecloudv1alpha1.BMEnrollmentConditionReasonTimedOut, errMessage, true)
		bmEnrollment.Status.ErrorMessage = errMessage
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseFailed
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, fmt.Errorf(errMessage)
	}
	return ctrl.Result{}, nil
}

// get BMC URL
func (reconciler *BMEnrollmentReconciler) getBMCURL(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (bmcURL string, err error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.getBMCURL")

	if reconciler.DDI == nil {
		log.Info("Getting BMC URL from NetBox")
		bmcURL, err = reconciler.NetBox.GetBMCURL(ctx, bmEnrollment.Spec.DeviceName)
		if err != nil {
			return "", fmt.Errorf("unable to get BMC URL with error:  %v", err)
		}
	} else {
		// find Range for BMC network based on Type = BMC and RackNAme
		ipRange, err := reconciler.DDI.GetRangeByName(ctx, bmEnrollment.Spec.RackName, MenAndMiceBMCType)
		if err != nil {
			return "", fmt.Errorf("failed to read Range: %v of Rack %s", err, bmEnrollment.Spec.RackName)
		}
		if len(ipRange.DhcpScopes) < 1 {
			return "", fmt.Errorf("failed to find dhcp Scopes in Range: %s", ipRange.Ref)
		}

		dhcpReservation, err := reconciler.DDI.GetDhcpReservationsByMacAddress(ctx, ipRange.DhcpScopes[0].Ref, bmEnrollment.Status.BMC.MACAddress)
		if err != nil {
			// TODO check Leases and convert to reservation
			// Currently delete Lease cause a failure in dhcp server
			return "", fmt.Errorf("failed to get dhcp reservation: %v", err)
		}
		if len(dhcpReservation.Addresses) < 1 {
			return "", fmt.Errorf("no addresses found for dhcpreservation: %s", dhcpReservation.Ref)
		}
		bmcURL = fmt.Sprintf("https://%s", dhcpReservation.Addresses[0])
	}

	log.Info("Found BMC URL", "bmcURL", bmcURL)
	return bmcURL, nil
}

// Get default BMC credentials
func (reconciler *BMEnrollmentReconciler) getDefaultBMCCredentials(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (bmcUsername string, bmcPassword string, err error) {
	// get default BMC credentials from Vault
	var secretPath string
	if bmEnrollment.Status.BMC.MACAddress == "" {
		secretPath = fmt.Sprintf("%s/deployed/virtual/default", bmEnrollment.Spec.Region)
	} else {
		secretPath = fmt.Sprintf("%s/deployed/%s/default", bmEnrollment.Spec.Region, bmEnrollment.Status.BMC.MACAddress)
	}
	bmcUsername, bmcPassword, err = reconciler.Vault.GetBMCCredentials(ctx, secretPath)
	if err != nil {
		return "", "", fmt.Errorf("unable to get BMC Credentials:  %v", err)
	}
	return bmcUsername, bmcPassword, nil
}

// generate new BMC credentials
func (reconciler *BMEnrollmentReconciler) generateBMCCredentials(ctx context.Context) (string, string, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.generateBMCCredentials")
	log.Info("Generating BMC credentials")

	newBMCUsername := helper.GetEnv(EnvBMCEnrollUsername, DefaultBMCEnrollUsername)
	log.Info("New Username", "newBMCUsername", newBMCUsername)

	// Customize the list of symbols of password generator.
	// Dell Gaudi doesn't following symbols in the password: &*_\\\"<>./
	passwordGenerator, err := password.NewGenerator(&password.GeneratorInput{
		// Supported symbols
		Symbols: "@%{}|[]?(),#+-=^~`",
	})
	if err != nil {
		return "", "", fmt.Errorf("could not generate a random password generator: %v", err)
	}
	// Generate a password that is 16 characters long with 5 digits, 5 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	randomBMCPassword, err := passwordGenerator.Generate(16, 5, 5, false, false)
	if err != nil {
		return "", "", fmt.Errorf("could not generate a random password: %v", err)
	}
	return newBMCUsername, randomBMCPassword, nil
}

// store new BMC user credentials in vault
func (reconciler *BMEnrollmentReconciler) storeUserBMCCredentialsInVault(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, newUsername, newPassword string) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.storeUserBMCCredentialsInVault")
	log.Info("Storing new User BMC deployed credentials")

	secretData := map[string]interface{}{
		"username": newUsername,
		"password": newPassword,
	}
	if err := reconciler.storeBMCCredentialsInVault(ctx, bmEnrollment, BMCUserSecretsPrefix, secretData); err != nil {
		return fmt.Errorf("unable to write to Vault client: %v", err)
	}
	return nil
}

// store credentials in vault
func (reconciler *BMEnrollmentReconciler) storeBMCCredentialsInVault(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, prefix string, secretData map[string]interface{}) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.storeBMCCredentialsInVault")
	log.Info("Storing new BMC deployed credentials")

	path, err := reconciler.getBMCSecretPath(bmEnrollment, prefix)
	if err != nil {
		return fmt.Errorf("unable to get BMC '%s': %v", bmEnrollment.Status.BMC.MACAddress, err)
	}
	log.Info("Storing credentials for MAC address", "bmcMACAddress", bmEnrollment.Status.BMC.MACAddress, "path", path)

	secret, err := reconciler.Vault.PutBMCSecrets(ctx, path, secretData)
	if err != nil {
		return fmt.Errorf("failed to write secrets from vault: %v", err)
	}

	log.Info("Returned data from write", "secret", secret)

	return nil
}

// get vault secret path
func (reconciler *BMEnrollmentReconciler) getBMCSecretPath(bmEnrollment *privatecloudv1alpha1.BMEnrollment, prefix string) (string, error) {
	if len(bmEnrollment.Status.BMC.MACAddress) < 12 { // Could be ff:ff:ff:ff:ff:ff, ff-ff-ff-ff-ff-ff, ffffffffffff
		return "", fmt.Errorf("BMC MAC address too short: %s", bmEnrollment.Status.BMC.MACAddress)
	}
	return fmt.Sprintf("%s/%s%s/%s", bmEnrollment.Spec.Region, BMCDeploymentSecretsPath, bmEnrollment.Status.BMC.MACAddress, prefix), nil
}

func (reconciler *BMEnrollmentReconciler) updateBMCCredentialsInRedfish(ctx context.Context, bmcInterface bmc.Interface, newUserName, newPassword string) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.updateBMCCredentialsInRedfish")
	log.Info("Updating BMC credentials")

	updateErr := bmcInterface.UpdateAccount(ctx, newUserName, newPassword)
	if updateErr == bmc.ErrAccountNotFound {
		if bmcInterface.IsVirtual() {
			log.Info("Skip creating an admin account on virtual BMC")
			return nil
		}
		if createdErr := bmcInterface.CreateAccount(ctx, newUserName, newPassword); createdErr != nil {
			return fmt.Errorf("unable to create Admin account's credentials: %v", createdErr)
		}
	} else if updateErr != nil {
		return fmt.Errorf("unable to update Admin account's credentials: %v", updateErr)
	}
	return nil
}

// delete bmc credentials from vault
func (reconciler *BMEnrollmentReconciler) deleteBMCCredentialsFromVault(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, prefix string) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.deleteBMCCredentialsFromVault")
	log.Info("Deleting new BMC deployed credentials")

	path, err := reconciler.getBMCSecretPath(bmEnrollment, prefix)
	if err != nil {
		return fmt.Errorf("unable to get BMC '%s': %v", bmEnrollment.Status.BMC.MACAddress, err)
	}

	log.Info("Deleting credentials for MAC address", "bmcMACAddress", bmEnrollment.Status.BMC.MACAddress, "path", path)
	err = reconciler.Vault.DeleteBMCSecrets(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete secrets from vault: %v", err)
	}

	log.Info("Deleted secrete from path", "path", path)

	return nil
}

// verify BMC credentials
func (reconciler *BMEnrollmentReconciler) verifyBMCCredentials(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, bmcInterface bmc.Interface) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.verifyBMCCredentials")
	log.Info("Verifying BMC default credentials")

	// Retrieve the service root
	service := bmcInterface.GetClient().GetService()

	//get a list of systems
	systems, err := service.Systems()
	if err != nil {
		return fmt.Errorf("failed to get Systems for BMC URL '%s': '%s'", bmEnrollment.Status.BMC.Address, err)
	}

	//iterate over the systems and print their details
	for _, system := range systems {
		log.Info("BMC", "System ID", system.ID(), "Name", system.Name(), "Power State", system.PowerState())
	}

	return nil
}

// get boot MAC address
func (reconciler *BMEnrollmentReconciler) getBootMacAddress(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, bmcInterface bmc.Interface) (string, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.getBootMacAddress")
	log.Info("Getting Boot MAC Address")

	hostMAC, err := bmcInterface.GetHostMACAddress(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get the host MAC address: %v", err)
	}
	if hostMAC == "" {
		// Try to extract it from Netbox
		hostMAC, err = reconciler.NetBox.GetBMCMACAddress(ctx, bmEnrollment.Spec.DeviceName, HostInterfaceName)
		if err != nil {
			return "", fmt.Errorf("unable to get Host MACAddress with error:  %v", err)
		}
		if hostMAC == "" {
			return "", fmt.Errorf("host MACAddress in Netbox is empty")
		}
	}

	normalizedMAC := helper.NormalizeMACAddress(hostMAC)
	log.Info("Normalizing MAC Address", "Original", hostMAC, "Normalized", normalizedMAC)

	return normalizedMAC, nil
}

// create Pxe record
func (reconciler *BMEnrollmentReconciler) createPxeRecord(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) error {
	// Using only in virtual environment in kind
	dhcpProxyUrl, present := os.LookupEnv(DhcpProxyUrlEnvVar)
	if present {
		_, err := http.Post(dhcpProxyUrl+"/record?mac="+bmEnrollment.Status.BootMACAddress+"&ironicip="+bmEnrollment.Status.IronicIPAddress, "Application/json", nil)
		if err != nil {
			return err
		}
	} else { // check if men and mice url is present
		if reconciler.DDI != nil {
			var ipRange *ddi.Range
			ipRange, err := reconciler.DDI.GetRangeByName(ctx, bmEnrollment.Spec.RackName, MenAndMiceProvisioningType)
			if err != nil {
				return fmt.Errorf("failed to read Range: %v of Rack %s", err, bmEnrollment.Spec.RackName)
			}
			if len(ipRange.DhcpScopes) < 1 {
				return fmt.Errorf("failed to find dhcp Scopes in Range: %s", ipRange.Ref)
			}
			var dhcpReservation *ddi.DhcpReservation
			var dhcpReservationRef string
			filename := fmt.Sprintf("http://%s:%s/%s", bmEnrollment.Status.IronicIPAddress, IronicHttpPortNb, IPXEProfileName)
			tftpServer := helper.GetEnv(TftpServerIPEnvVar, bmEnrollment.Status.IronicIPAddress)
			dhcpReservation, err = reconciler.DDI.GetDhcpReservationsByMacAddress(ctx, ipRange.DhcpScopes[0].Ref, bmEnrollment.Status.BootMACAddress)
			if err != nil {
				//get available IP address
				ipaddress, err := reconciler.DDI.GetAvailableIp(ctx, ipRange)
				if err != nil {
					return fmt.Errorf("failed to get an IP %v", err)
				}
				// Create a reservation
				dhcpReservationRef, err = reconciler.DDI.SetDhcpReservationByScope(ctx, ipRange, bmEnrollment.Status.BootMACAddress, ipaddress, bmEnrollment.Spec.DeviceName, filename, tftpServer)
				if err != nil {
					return fmt.Errorf("failed Set DhcpReservation %v", err)
				}
			} else {
				dhcpReservationRef = dhcpReservation.Ref
			}
			err = reconciler.DDI.UpdateDhcpReservationOptions(ctx, ipRange, bmEnrollment.Spec.DeviceName, filename, tftpServer, dhcpReservationRef, IPXEBinarayName)
			if err != nil {
				return fmt.Errorf("failed Update DhcpReservation ByMacAddress %v", err)
			}
		}
	}
	return nil
}

// get the BareMetalHost IP address
func (reconciler *BMEnrollmentReconciler) findBareMetalHostIP(ctx context.Context, bmHost *baremetalv1alpha1.BareMetalHost) (string, error) {
	ipAddress := ""
	// find BMH IP
	nicInformation := bmHost.Status.HardwareDetails.NIC
	for i := range nicInformation {
		if helper.NormalizeMACAddress(nicInformation[i].MAC) == helper.NormalizeMACAddress(bmHost.Spec.BootMACAddress) {
			ipAddress = nicInformation[i].IP
			break
		}
	}
	if ipAddress == "" {
		return "", fmt.Errorf("failed to fetch Host IP from BMH %s", bmHost.Name)
	}
	return ipAddress, nil
}

// get BIOS password from vault
func (reconciler *BMEnrollmentReconciler) getBIOSPasswordFromVault(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (string, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.getBIOSPasswordFromVault")
	log.Info("Check for existing BIOS password")

	path, err := reconciler.getBMCSecretPath(bmEnrollment, bmcBIOSSecretsPrefix)
	if err != nil {
		return "", fmt.Errorf("unable to get BMC Secret path '%s': %v", bmEnrollment.Status.BMC.MACAddress, err)
	}
	log.Info("requesting BIOS Password for MAC address", "bmcMACAddress", bmEnrollment.Status.BMC.MACAddress, "path", path)

	// Request the password from Vault
	biosPassword, err := reconciler.Vault.GetBMCBIOSPassword(ctx, path)
	if err != nil {
		return "", fmt.Errorf("unable to read from Vault client: %v", err)
	}

	return biosPassword, nil
}

// generate BIOS password
func (reconciler *BMEnrollmentReconciler) generateBIOSPassword(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.generateBIOSPassword")
	log.Info("Generating BIOS password")

	// BIOS/syscfg password rules
	// The password must have a length of 8–14 characters.
	// The password can have alphanumeric characters (a-z, A-Z, 0–9) and the following special characters:
	// ! @ # $ % ^ *( ) - _ + = ? '
	// Use two double quotes (“”) to represent a null password.
	passwordGenerator, err := password.NewGenerator(&password.GeneratorInput{
		Symbols: "!@#$%^*()-_+=?",
	})
	if err != nil {
		return "", fmt.Errorf("Could not create password generator: %v", err)
	}

	randomBIOSPassword, err := passwordGenerator.Generate(14, 3, 2, false, false)
	if err != nil {
		return "", fmt.Errorf("Could not generate a random password: %v", err)
	}

	return randomBIOSPassword, nil
}

func (reconciler *BMEnrollmentReconciler) storeBIOSPasswordInVault(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, newPassword string) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.storeBIOSPasswordInVault")
	log.Info("Storing new BIOS password")

	secretData := map[string]interface{}{
		"password": newPassword,
	}
	if err := reconciler.storeBMCCredentialsInVault(ctx, bmEnrollment, bmcBIOSSecretsPrefix, secretData); err != nil {
		return fmt.Errorf("unable to write to Vault client: %v", err)
	}

	return nil
}

// add storage annotations to the BaremetalHost resource
func (reconciler *BMEnrollmentReconciler) addStorageAnnotations(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, bmcInterface bmc.Interface) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.addStorageLabels")
	var storageMac string
	var isStorageNicPresent bool
	var err error
	storageIntfToMacMap := make(map[string]string)

	// get BMH
	bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
	if err != nil {
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionBMHDeprovisioning, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH, privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH, false)
		return fmt.Errorf("failed to get %s during BareMetalHost deprovisioning(after patching): %v", bmEnrollment.Spec.DeviceName, err)
	}
	if !bmcInterface.IsVirtual() {
		// Check for net1/0 which is the storage interface.
		isStorageNicPresent, storageMac, err = reconciler.getStorageNetworkDetails(ctx, bmEnrollment, StorageInterfaceName1)
		if err != nil {
			return fmt.Errorf("failed to fetch Storage MAC of interface %s from Netbox %s, %v", StorageInterfaceName1, bmHost.Name, err)
		}
		storageIntfToMacMap[StorageInterfaceName1] = storageMac
	} else {
		nicInformation := bmHost.Status.HardwareDetails.NIC
		for _, nic := range nicInformation {
			normalized_mac := helper.NormalizeMACAddress(nic.MAC)
			if normalized_mac != helper.NormalizeMACAddress(bmEnrollment.Status.BMC.MACAddress) &&
				normalized_mac != helper.NormalizeMACAddress(bmEnrollment.Status.BootMACAddress) {
				log.Info("Found storage nic", "nic", nic)
				isStorageNicPresent = true
				storageMac = normalized_mac
				break
			}
		}
		storageIntfToMacMap[StorageInterfaceName1] = storageMac
	}
	if isStorageNicPresent {
		if err := reconciler.updateStorageAnnotations(ctx, bmHost, storageIntfToMacMap); err != nil {
			return fmt.Errorf("error observed while updating StorageLabel: %v", err)
		}
	} else {
		log.Info("No Storage NIC Present for BMH")
	}
	return nil
}

// Get storage network details from Netbox
func (reconciler *BMEnrollmentReconciler) getStorageNetworkDetails(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, storageInterfaceName string) (bool, string, error) {
	log := log.FromContext(ctx).WithName("reconciler *BMEnrollmentReconciler.getStorageNetworkDetails")
	// Fetch the Storage MAC from netbox
	log.Info("Fetching Storage Network from netbox")
	storageMacAddress, err := reconciler.NetBox.GetBMCMACAddress(ctx, bmEnrollment.Spec.DeviceName, storageInterfaceName)
	if err != nil {
		log.Info("Error observed while attempting to fetch storage interface", "error", err)
		// log the error and set the storage as false
		return false, "", nil
	}
	if storageMacAddress == "" {
		return false, "", fmt.Errorf("host MACAddress of interface %s in Netbox is empty", storageInterfaceName)
	}
	return true, storageMacAddress, nil
}

// Upate storage annotations of BMH
func (reconciler *BMEnrollmentReconciler) updateStorageAnnotations(ctx context.Context, bmHost *baremetalv1alpha1.BareMetalHost, storageIntfToMacMap map[string]string) error {
	log := log.FromContext(ctx).WithName("reconciler *BMEnrollmentReconciler.updateStorageLabel")

	PatchHelper, err := patch.NewHelper(bmHost, reconciler.Client)
	if err != nil {
		return fmt.Errorf("failed to create helper to patch BareMetalHost with storage annotations: %v", err)
	}
	for intf, storageMacAddress := range storageIntfToMacMap {
		normalizedStorageMac := helper.NormalizeMACAddress(storageMacAddress)
		// Verify if storage MAC is different from boot MAC address
		if helper.NormalizeMACAddress(bmHost.Spec.BootMACAddress) == normalizedStorageMac {
			return fmt.Errorf("storage MAC %v from netbox matches BootMACAddress", storageMacAddress)
		}
		nicInformation := bmHost.Status.HardwareDetails.NIC
		var found bool
		// Verify if the storage nic is present in the BMH after inspection
		for _, nic := range nicInformation {
			normalized_mac := helper.NormalizeMACAddress(nic.MAC)
			if normalized_mac == normalizedStorageMac {
				found = true
			}
		}
		if !found {
			log.Info("Storage MAC in netbox is not present on the host", "storageMAC", storageMacAddress)
			return fmt.Errorf("storage MAC %v from netbox does not match BMH after inspection", storageMacAddress)
		}

		if bmHost.Annotations == nil {
			bmHost.Annotations = make(map[string]string)
		}
		intfIndex := strings.Replace(strings.Trim(intf, "net"), "/", "-", -1)
		bmHost.Annotations[fmt.Sprintf("%s/eth%s", StorageMACAnnotationPrefix, intfIndex)] = storageMacAddress
	}
	err = PatchHelper.Patch(ctx, bmHost)
	if err != nil {
		return fmt.Errorf("failed to update the BareMetalHost with storage annotations %w", err)
	}
	return nil
}

// updateHostHardwareLabels updates host's labels with hardware details from Ironic inspection and BMC data.
func (reconciler *BMEnrollmentReconciler) updateHostHardwareLabels(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, bmcInterface bmc.Interface) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.updateHostHardwareLabels")
	log.Info("Updating host's hardware labels")

	// get BMH
	bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
	if err != nil {
		return fmt.Errorf("failed to get BareMetalHost %s when trying to get host IP %v", bmEnrollment.Spec.DeviceName, err)
	}

	PatchHelper, err := patch.NewHelper(bmHost, reconciler.Client)
	if err != nil {
		return fmt.Errorf("failed to create helper to patch BareMetalHost with hardware labels: %v", err)
	}

	if bmHost.Status.HardwareDetails == nil {
		return fmt.Errorf("hardware details not found")
	}

	bmhMemorySize := (bmHost.Status.HardwareDetails.RAMMebibytes) / (units.GiB / units.MiB)

	cpu, err := bmcInterface.GetHostCPU(ctx)
	if err != nil {
		return fmt.Errorf("unable to get CPU information from BMC: %v", err)
	}

	gpuCount, gpuModel, err := bmcInterface.GPUDiscovery(ctx)
	if err != nil {
		return fmt.Errorf("unable to get GPU information from BMC: %v", err)
	}

	hbmMode, err := bmcInterface.HBMDiscovery(ctx)
	if err != nil {
		return fmt.Errorf("unable to get HBM information from BMC: %v", err)
	}

	hostLabels := bmHost.GetLabels()
	if hostLabels == nil {
		hostLabels = make(map[string]string)
	}
	hostLabels[CPUIDLabel] = strings.ToLower(cpu.CPUID)
	hostLabels[CPUCountLabel] = strconv.Itoa(bmHost.Status.HardwareDetails.CPU.Count)
	hostLabels[CPUSocketsLabel] = strconv.Itoa(cpu.Sockets)
	hostLabels[CPUCoresLabel] = strconv.Itoa(cpu.Cores)
	hostLabels[CPUThreadsLabel] = strconv.Itoa(cpu.Threads)
	hostLabels[CPUManufacturerLabel] = formatLabelValue(cpu.Manufacturer)
	hostLabels[GPUModelNameLabel] = strings.ToLower(gpuModel)
	hostLabels[GPUCountLabel] = strconv.Itoa(gpuCount)
	hostLabels[HBMModeLabel] = strings.ToLower(hbmMode)
	hostLabels[MemorySizeLabel] = strconv.Itoa(bmhMemorySize) + "Gi"
	hostLabels[NetworkModeLabel] = ""

	assignInstanceTypeLabelFailed := false
	var assignInstanceTypeErr error
	// function to assign InstanceTypeLabel to the host as a part of the enrollment
	if err := reconciler.assignInstanceTypeLabel(ctx, hostLabels); err != nil {
		log.Error(err, "failed to assign InstanceTypeLabel to BareMetalHost")
		assignInstanceTypeLabelFailed = true
		assignInstanceTypeErr = err
	}

	if !assignInstanceTypeLabelFailed {
		// trigger validation
		hostLabels[ReadyToTestLabel] = "true"
		// set instance group ID and cluster size
		if bmEnrollment.Spec.Cluster != "" {
			hostLabels[ClusterGroupID] = bmEnrollment.Spec.Cluster
			clusterSize, err := reconciler.NetBox.GetClusterSize(ctx, bmEnrollment.Spec.Cluster, bmEnrollment.Spec.AvailabilityZone)
			if err != nil {
				return fmt.Errorf("failed to get the cluster size. error: %s", err.Error())
			}
			hostLabels[ClusterSize] = strconv.FormatInt(clusterSize, 10)
			// set network mode
			clusterNetworkMode, err := reconciler.NetBox.GetClusterNetworkMode(ctx, bmEnrollment.Spec.Cluster, bmEnrollment.Spec.AvailabilityZone)
			if err != nil {
				return fmt.Errorf("failed to get the cluster network mode. error: %s", err.Error())
			}
			hostLabels[NetworkModeLabel] = clusterNetworkMode
		}
	}

	bmHost.SetLabels(hostLabels)

	hostAnnotations := bmHost.GetAnnotations()
	if hostAnnotations == nil {
		hostAnnotations = make(map[string]string)
	}
	// add annotation if system is gaudi
	switch bmcInterface.GetHwType() {
	case bmc.Gaudi2Smc, bmc.Gaudi2Wiwynn, bmc.Gaudi2Dell:
		if err = reconciler.getExtraEthernetMacAnnotations(ctx, bmEnrollment, hostAnnotations); err != nil {
			return fmt.Errorf("unable to habana GPUs mac addresses: %v", err)
		}
	default:
		break
	}
	// add model to annotation instead because its value format is not suitable for label
	hostAnnotations[CPUModelLabel] = bmHost.Status.HardwareDetails.CPU.Model
	bmHost.SetAnnotations(hostAnnotations)

	/*labelStr, err := json.Marshal(bmh.GetLabels())
	if err != nil {
		return err
	}
	annotationsStr, err := json.Marshal(bmh.GetAnnotations())
	if err != nil {
		return err
	}

	patch := []byte(fmt.Sprintf(`{"metadata":{"labels":%s, "annotations":%s}}`, labelStr, annotationsStr))

	_, err = t.dynamicClient.Resource(bmHostGVR).
		Namespace(bmh.GetNamespace()).
		Patch(ctx, bmh.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}*/

	// patch BaremetalHost with Updated labels and annotations
	err = PatchHelper.Patch(ctx, bmHost)
	if err != nil {
		return fmt.Errorf("failed to update the BareMetalHost with hardware labels and annotations %w", err)
	}

	if assignInstanceTypeLabelFailed {
		return fmt.Errorf("unable to assign InstanceType Label to BareMetalHost: %v", assignInstanceTypeErr)
	}

	return nil
}

// assign instance type label
func (reconciler *BMEnrollmentReconciler) assignInstanceTypeLabel(ctx context.Context, hostLabels map[string]string) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.assignInstanceTypeLabel")
	log.Info("Assigning InstanceTypeLabel to host")

	instanceTypesSearchResponse, err := reconciler.InstanceTypeServiceClient.Search(ctx, &pb.InstanceTypeSearchRequest{})
	if err != nil {
		return fmt.Errorf("unable to get instanceTypeList information from InstanceTypeClient: %v", err)
	}

	// First map for instanceType
	instanceTypeLabels := make(map[string]string)

	// Second map for hostLabelsSpecs
	hostLabelsSpecs := make(map[string]string)

	hostLabelsSpecs[CPUIDLabel] = hostLabels[CPUIDLabel]
	hostLabelsSpecs[CPUCountLabel] = hostLabels[CPUCountLabel]
	hostLabelsSpecs[GPUModelNameLabel] = hostLabels[GPUModelNameLabel]
	hostLabelsSpecs[GPUCountLabel] = hostLabels[GPUCountLabel]
	hostLabelsSpecs[HBMModeLabel] = hostLabels[HBMModeLabel]
	hostLabelsSpecs[MemorySizeLabel] = hostLabels[MemorySizeLabel]

	instanceTypeLabelAssigned := false

	for _, instanceType := range instanceTypesSearchResponse.Items {
		if instanceType.Spec.InstanceCategory == pb.InstanceCategory_BareMetalHost {
			instanceTypeLabels[CPUIDLabel] = strings.ToLower(instanceType.Spec.Cpu.Id)
			instanceTypeLabels[CPUCountLabel] = strconv.Itoa(calculateTotalCpuCount(instanceType))
			instanceTypeLabels[GPUModelNameLabel] = strings.ToLower(instanceType.Spec.Gpu.ModelName)
			instanceTypeLabels[GPUCountLabel] = strconv.Itoa(int(instanceType.Spec.Gpu.Count))
			instanceTypeLabels[HBMModeLabel] = strings.ToLower(instanceType.Spec.HbmMode)
			instanceTypeLabels[MemorySizeLabel] = instanceType.Spec.Memory.Size

			// match the specs
			specMatch := reflect.DeepEqual(instanceTypeLabels, hostLabelsSpecs)
			if specMatch {
				hostLabels[fmt.Sprintf(InstanceTypeLabel, instanceType.Spec.Name)] = "true"
				instanceTypeLabelAssigned = true
				break
			}
		}
	}

	if !instanceTypeLabelAssigned {
		return errors.New("failed to assign InstanceType Label")
	}

	return nil
}

// Add extra ethernet mac annotations(cluster network)
func (reconciler *BMEnrollmentReconciler) getExtraEthernetMacAnnotations(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, annotations map[string]string) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.GetExtraEthernetMacAnnotations")
	log.Info("getting extra ethernet information")
	sshManager := &myssh.MySSHManager{}
	hahabna, err := ipacmd.NewIpaCmdHelper(ctx, reconciler.Vault, sshManager, bmEnrollment.Status.HostIPAddress, bmEnrollment.Spec.Region)
	if err != nil {
		return fmt.Errorf("unable to initialize IpmCmd helper: %v", err)
	}
	habanabuses, err := hahabna.HabanaEthernetBusInfo(ctx)
	if err != nil {
		return fmt.Errorf("unable to get Habana Ethernet Bus Info: %v", err)
	}
	if err = hahabna.HabanaEthernetMacAddress(ctx, habanabuses); err != nil {
		return fmt.Errorf("unable to get Habana Ethernet MacAddresses: %v", err)
	}
	for _, item := range habanabuses {
		if len(item.MacAddresses) > 0 {
			for i, _ := range item.MacAddresses {
				annotations[fmt.Sprintf("%s/gpu-eth%s-%d", GPUAnnotationPrefix, item.ModuleID, i)] = item.MacAddresses[i]
			}
		}
	}
	return nil
}

// assign BareMetalHost namespace
func (reconciler *BMEnrollmentReconciler) assignBareMetalHostNamespace(ctx context.Context) (*v1.Namespace, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.assignBareMetalHostNamespace")
	namespaces, err := reconciler.listMetal3Namepaces(ctx)
	if err != nil {
		return nil, err
	}
	// find the namespace with the least number of baremetalhosts
	targetNamespace := &namespaces.Items[0]
	lowest := math.MaxInt64

	for i, ns := range namespaces.Items {
		hostList, err := reconciler.listBaremetalhosts(ctx, ns.Name)
		if err != nil {
			return nil, fmt.Errorf("unable to list baremetalhosts: %v", err)
		}
		current := len(hostList.Items)
		log.Info("Metal3 Namespace", "name", ns.Name, "host found", current)
		if current < lowest {
			lowest = current
			targetNamespace = &namespaces.Items[i]
		}
	}
	log.Info(fmt.Sprintf("Using %q namespace for the new host", targetNamespace.Name))
	return targetNamespace, nil
}

// get metal3 namespaces
func (reconciler *BMEnrollmentReconciler) listMetal3Namepaces(ctx context.Context) (*v1.NamespaceList, error) {
	selector, err := labels.Parse(fmt.Sprintf("%s=true", Metal3NamespaceSelectorKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse metal3 labels: %v", err)
	}
	nsList := &v1.NamespaceList{}
	err = reconciler.List(ctx, nsList, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return nil, fmt.Errorf("unable to list metal3 namespaces: %v", err)
	}
	if len(nsList.Items) == 0 {
		return nil, fmt.Errorf("no metal3 namespace found")
	}
	return nsList, nil
}

// list baremetalhosts
func (reconciler *BMEnrollmentReconciler) listBaremetalhosts(ctx context.Context, namespace string) (*baremetalv1alpha1.BareMetalHostList, error) {
	hostsList := &baremetalv1alpha1.BareMetalHostList{}
	err := reconciler.List(ctx, hostsList, client.InNamespace(namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list baremetalhosts in namespace %s: %v", namespace, err)
	}
	return hostsList, nil
}

// newBareMetalHost returns a new BareMetalHost with spec
func (reconciler *BMEnrollmentReconciler) createBareMetalHostSpec(bmEnrollment *privatecloudv1alpha1.BMEnrollment) *baremetalv1alpha1.BareMetalHost {
	rootDeviceHints := &baremetalv1alpha1.RootDeviceHints{}
	bootMode := baremetalv1alpha1.UEFI
	// set root device hint and boot mode for virtual hosts
	if bmEnrollment.Status.BMC.HardwareType == bmc.Virtual.String() {
		rootDeviceHints = &baremetalv1alpha1.RootDeviceHints{
			DeviceName: "/dev/vda",
		}
		bootMode = baremetalv1alpha1.Legacy
	}

	bmHost := &baremetalv1alpha1.BareMetalHost{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "metal3.io/v1alpha1",
			Kind:       "BareMetalHost",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bmEnrollment.Spec.DeviceName,
			Namespace: bmEnrollment.Status.TargetBmNamespace,
		},
		Spec: baremetalv1alpha1.BareMetalHostSpec{
			Online:         true,
			BootMode:       bootMode,
			BootMACAddress: bmEnrollment.Status.BootMACAddress,
			BMC: baremetalv1alpha1.BMCDetails{
				Address:                        bmEnrollment.Status.BMC.Metal3Address,
				CredentialsName:                bmEnrollment.Status.BMC.SecretName,
				DisableCertificateVerification: true,
			},
			RootDeviceHints: rootDeviceHints,
		},
	}

	return bmHost
}

// create BMH host
func (reconciler *BMEnrollmentReconciler) createBareMetalHost(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.createBareMetalHost")
	// create a new BareMetalHost
	bmHost := reconciler.createBareMetalHostSpec(bmEnrollment)
	log.Info("Creating BareMetalHost", "name", bmHost.Name, "namespace", bmEnrollment.Status.TargetBmNamespace)
	if err := reconciler.Create(ctx, bmHost); err != nil {
		if apierrors.IsAlreadyExists(err) {
			log.Info("BareMetalHost already exists", "host", bmHost.Name)
		} else {
			return fmt.Errorf("failed to create BareMetalHost: %v", err)
		}
	}
	return nil
}

// get BareMetalHost with backoff support
func (reconciler *BMEnrollmentReconciler) getBareMetalHostWithBackOff(ctx context.Context, name string, namespace string) (*baremetalv1alpha1.BareMetalHost, error) {
	host := &baremetalv1alpha1.BareMetalHost{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	existing := &baremetalv1alpha1.BareMetalHost{}

	// occasionally, bmh secret creation fails when the Get request below fails.
	// retrying with backoff when error message type is not found
	isRetryable := func(err error) bool {
		if apierrors.IsNotFound(err) {
			return true
		} else {
			return false
		}
	}

	backoff := wait.Backoff{
		Steps:    9,
		Duration: 10 * time.Millisecond,
		Factor:   3.0,
		Jitter:   0.1,
	}

	if err := retry.OnError(backoff, isRetryable, func() error {
		err := reconciler.Get(ctx, client.ObjectKeyFromObject(host), existing)
		return err
	}); err != nil {
		return nil, fmt.Errorf("error when trying to get BareMetalHost %s", name)
	}

	if reflect.ValueOf(existing).IsZero() {
		return nil, fmt.Errorf("failed to get BareMetalHost %s", name)
	}
	return existing, nil
}

// get BareMetalHost
func (reconciler *BMEnrollmentReconciler) getBareMetalHost(ctx context.Context, name string, namespace string) (*baremetalv1alpha1.BareMetalHost, error) {
	host := &baremetalv1alpha1.BareMetalHost{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	existing := &baremetalv1alpha1.BareMetalHost{}
	if err := reconciler.Get(ctx, client.ObjectKeyFromObject(host), existing); err != nil {
		return nil, err
	}
	if reflect.ValueOf(existing).IsZero() {
		return nil, fmt.Errorf("failed to get BareMetalHost %s", name)
	}
	return existing, nil
}

// Patch BareMetalHost
func (reconciler *BMEnrollmentReconciler) patchBareMetalHostImage(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, patch []byte) error {
	// get BMH
	bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
	if err != nil {
		return fmt.Errorf("failed to get BareMetalHost to patch it: %v", err)
	}
	if err = reconciler.Patch(ctx, bmHost, client.RawPatch(types.MergePatchType, patch)); err != nil {
		return fmt.Errorf("failed to patch BareMetalHost: %v", err)
	}
	return nil
}

// NewBareMetalHostSecret returns a new secret that contain BMC credentials for BareMetalHost
func (reconciler *BMEnrollmentReconciler) newBareMetalHostSecret(bmEnrollment *privatecloudv1alpha1.BMEnrollment, username string, password string, namespace string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-bmc-secret", bmEnrollment.Spec.DeviceName),
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"username": []byte(username),
			"password": []byte(password),
		},
	}
}

// create bmh secret
func (reconciler *BMEnrollmentReconciler) createBMHSecret(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.createBMHSecret")

	// get bmc credentials
	bmcUsername, bmcPassword, err := reconciler.getSecretData(ctx, bmEnrollment)
	if err != nil {
		return fmt.Errorf("failed to get BMC secret data: %v", err)
	}

	//bmh secret
	bmcSecret := reconciler.newBareMetalHostSecret(bmEnrollment, bmcUsername, bmcPassword, bmEnrollment.Status.TargetBmNamespace)
	newHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
	if err != nil {
		return fmt.Errorf("failed to get %s during BareMetalHost secret creation: %v", bmEnrollment.Spec.DeviceName, err)
	}
	bmcSecret.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: baremetalv1alpha1.GroupVersion.String(),
			Kind:       "BareMetalHost",
			Name:       newHost.GetName(),
			UID:        newHost.GetUID(),
		},
	}
	log.Info("creating BareMetalHost secret", "name", bmcSecret.Name, "namespace", bmEnrollment.Status.TargetBmNamespace)
	if err := reconciler.Create(ctx, bmcSecret); err != nil {
		if apierrors.IsAlreadyExists(err) {
			log.Info("BareMetalHost BMC secret already exists", "host", bmcSecret.Name)
		} else {
			return fmt.Errorf("failed to create BareMetalHost BMC secret: %v", err)
		}
	}
	return nil
}

// get k8s secret
func (reconciler *BMEnrollmentReconciler) getSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error) {
	existing := &v1.Secret{}
	if err := reconciler.Get(ctx, client.ObjectKeyFromObject(secret), existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// get enrollment secret data
func (reconciler *BMEnrollmentReconciler) getSecretData(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment) (string, string, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.getEnrollmentSecretData").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	var existing *v1.Secret
	var err error
	secret := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: bmEnrollment.Status.BMC.SecretName, Namespace: bmEnrollment.Namespace}}

	if existing, err = reconciler.getSecret(ctx, secret); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("secret not found", "name", secret.Name)
			return "", "", fmt.Errorf("secret not found: %v", err)
		} else {
			return "", "", err
		}
	}
	if existing.Data["username"] == nil {
		return "", "", fmt.Errorf("secret data username not found: %v", err)
	}
	if existing.Data["password"] == nil {
		return "", "", fmt.Errorf("secret data password not found: %v", err)
	}
	return string(existing.Data["username"]), string(existing.Data["password"]), nil
}

// Patch BMC secret data
func (reconciler *BMEnrollmentReconciler) patchSecret(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, username string, password string) error {
	secretInfo := v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: bmEnrollment.Status.BMC.SecretName, Namespace: bmEnrollment.Namespace}}

	secret, err := reconciler.getSecret(ctx, &secretInfo)
	if err != nil {
		return fmt.Errorf("failed to get secret for patching: %v", err)
	}
	data := map[string][]byte{
		"username": []byte(username),
		"password": []byte(password),
	}
	patch, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal secret data for patching: %v", err)
	}
	if err = reconciler.Patch(ctx, secret, client.RawPatch(types.StrategicMergePatchType, patch)); err != nil {
		return fmt.Errorf("failed to patch bmc secret: %v", err)
	}
	return nil
}

// create secret with default bmc credentials
func (reconciler *BMEnrollmentReconciler) createEnrollmentBMCSecret(ctx context.Context, secret *v1.Secret, bmEnrollment *privatecloudv1alpha1.BMEnrollment) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.createEnrollmentBMCSecret")
	// create secret with bmc credentials, if not present
	if _, err := reconciler.getSecret(ctx, secret); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("BM enrollment secret not found. create new enrollment secret", "name", secret.Name)
		} else {
			return fmt.Errorf("failed to get the enrollment bmc secret: %v", err)
		}
	} else {
		if err = reconciler.deleteSecret(ctx, secret); err != nil {
			return fmt.Errorf("failed to delete the enrollment bmc secret: %v", err)
		}
	}

	if err := controllerutil.SetControllerReference(bmEnrollment, secret, reconciler.Scheme); err != nil {
		return fmt.Errorf("failed to set bmc secret reference: %v", err)
	}
	log.Info("creating new BM enrollment secret", "name", secret.Name)
	if err := reconciler.Create(ctx, secret); err != nil {
		return fmt.Errorf("failed to create bmc secret: %v", err)
	}
	return nil
}

// delete k8s secret
func (reconciler *BMEnrollmentReconciler) deleteSecret(ctx context.Context, secret *v1.Secret) error {
	if err := reconciler.Delete(ctx, secret); err != nil {
		return err
	}
	return nil
}

// Disenroll a node
func (reconciler *BMEnrollmentReconciler) processDisenrollment(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.processDisenrollment").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	log.Info("Disenrolloing node")
	// set conditions
	SetStatusConditionIfMissing(bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks, v1.ConditionFalse, privatecloudv1alpha1.ConditionReasonNone, "")
	// check if Disenrollment phase is set to failed
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks)
	if enrollmentFailed(bmEnrollment) && condition.Reason != privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		log.Info("disenrollment failed with error", "error", bmEnrollment.Status.ErrorMessage)
		return ctrl.Result{}, &skipDisenrollment{enrollmentMsg: bmEnrollment.Status.ErrorMessage}
	}
	// disenrollment pre checks
	result, err := reconciler.preDisenrollmentChecks(ctx, bmEnrollment, req)
	if err != nil || !result.IsZero() {
		return result, err
	}
	if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (reconciler *BMEnrollmentReconciler) preDisenrollmentChecks(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.preDisenrollmentChecks").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	condition := FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks)
	if condition.Reason == privatecloudv1alpha1.BMEnrollmentConditionReasonNone {
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonPreDisenrollmentChecksStarted, privatecloudv1alpha1.BMEnrollmentMessagePreDisenrollmentChecksStarted, true)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseDisenrolling
		if err := reconciler.updateEnrollmentStatus(ctx, bmEnrollment, req); err != nil {
			return ctrl.Result{}, err
		}
		condition = FindStatusCondition(bmEnrollment.Status.Conditions, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks)
	}
	// timeout if pre disenrollment checks are not completed within the timeout duration
	result, err := reconciler.isEnrollmentConditionTimedOut(ctx, bmEnrollment, condition.Type, req, EnrollmentGeneralTimeout)
	if err != nil {
		return result, err
	}
	// get BMH
	bmHost, err := reconciler.getBareMetalHostWithBackOff(ctx, bmEnrollment.Spec.DeviceName, bmEnrollment.Status.TargetBmNamespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("BareMetalHost not found. Deleting Enrollment")
			if err = reconciler.Client.Delete(ctx, bmEnrollment); err != nil {
				return ctrl.Result{}, err
			} else {
				log.Info("BareMetal disenrollment completed")
				return ctrl.Result{}, nil
			}
		}
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToGetBMH, privatecloudv1alpha1.BMEnrollmentMessageFailedToGetBMH, false)
		return ctrl.Result{}, fmt.Errorf("failed to get BareMetalHost %s during disenrollment: %v", bmEnrollment.Spec.DeviceName, err)
	}
	// requeue if BareMetalHost is deprovisioning
	if bareMetalHostDeprovisioningInProgress(bmHost) {
		log.Info("BareMetalHost deprovisioning is in progress. Requeuing")
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonBMHDeprovisioningInProgress, privatecloudv1alpha1.BMEnrollmentMessageBMHDeprovisioningInProgress, false)
		return ctrl.Result{RequeueAfter: GeneralRequeueAfter}, nil
	}
	// Skip disenrollment if the node is consumed
	if bareMetalHostConsumed(bmHost) {
		log.Info("BareMetalHost is consumed")
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionFailed, v1.ConditionTrue,
			privatecloudv1alpha1.BMEnrollmentConditionReasonConsumedBMH, privatecloudv1alpha1.BMEnrollmentMessageDisenrollmentConsumedBMH, false)
		bmEnrollment.Status.Phase = privatecloudv1alpha1.BMEnrollmentPhaseFailed
		bmEnrollment.Status.ErrorMessage = privatecloudv1alpha1.BMEnrollmentMessageDisenrollmentConsumedBMH
		return ctrl.Result{}, &skipDisenrollment{enrollmentMsg: privatecloudv1alpha1.BMEnrollmentMessageDisenrollmentConsumedBMH}
	}
	// requeue if BareMetalHost is provisioning/provisioned
	if !bareMetalHostProvisioningInProgress(bmHost) || !bareMetalHostProvisioned(bmHost) {
		log.Info("Deleting BareMetalHost")
		if err = reconciler.Client.Delete(ctx, bmHost); err != nil {
			SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToDeleteBMH, privatecloudv1alpha1.BMEnrollmentMessageFailedToDeleteBMH, false)
			return ctrl.Result{}, err
		}
		log.Info("Deleting BareMetal enrollment")
		if err = reconciler.Client.Delete(ctx, bmEnrollment); err != nil {
			SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks, v1.ConditionFalse,
				privatecloudv1alpha1.BMEnrollmentConditionReasonFailedToDeleteEnrollment, privatecloudv1alpha1.BMEnrollmentMessageFailedToDeleteEnrollment, false)
			return ctrl.Result{}, err
		}
		if err == nil {
			log.Info("BareMetal disenrollment completed")
			return ctrl.Result{}, nil
		}
	} else {
		log.Info("BareMetalHost provisioning is in progress. Requeuing")
		SetBMEnrollmentCondition(ctx, bmEnrollment, privatecloudv1alpha1.BMEnrollmentConditionPreDisenrollmentChecks, v1.ConditionFalse,
			privatecloudv1alpha1.BMEnrollmentConditionReasonBMHProvisioningInProgress, privatecloudv1alpha1.BMEnrollmentMessageBMHProvisioningInProgress, false)
		return ctrl.Result{RequeueAfter: GeneralRequeueAfter}, nil
	}
	return ctrl.Result{RequeueAfter: GeneralRequeueAfter}, nil
}

// remove finalizer if it exists
func (reconciler *BMEnrollmentReconciler) removeFinalizerIfExists(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.removeFinalizerIfMissing").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	log.Info("Removing enrollment finalizer")
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestEnrollment := &privatecloudv1alpha1.BMEnrollment{}
		err := reconciler.Get(ctx, req.NamespacedName, latestEnrollment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("BM enrollment resource not found. Ignoring")
				return nil
			}
			log.Info("Failed to get BM enrollment resource. Re-running reconcile.")
			return err
		}
		// Remove finalizer
		if controllerutil.ContainsFinalizer(latestEnrollment, EnrollmentFinalizer) {
			controllerutil.RemoveFinalizer(latestEnrollment, EnrollmentFinalizer)
		}

		if !reflect.DeepEqual(bmEnrollment.GetFinalizers(), latestEnrollment.GetFinalizers()) {
			log.Info("enrollment finalizer mismatches", "currentEnrollmentFinalizers", bmEnrollment.GetFinalizers(),
				"latestEnrollmentFinalizers", latestEnrollment.GetFinalizers())

			if err := reconciler.Update(ctx, latestEnrollment); err != nil {
				return fmt.Errorf("RemoveFinalizers: update failed: %w", err)
			}
		} else {
			log.Info("enrollment finalizer doesn't need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update enrollment finalizers: %w", err)
	}
	return nil
}

// Update enrollment status.
func (reconciler *BMEnrollmentReconciler) updateEnrollmentStatus(ctx context.Context, bmEnrollment *privatecloudv1alpha1.BMEnrollment, req ctrl.Request) error {
	log := log.FromContext(ctx).WithName("BMEnrollmentReconciler.updateEnrollmentStatus").WithValues("deviceName", bmEnrollment.Spec.DeviceName)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestBMEnrollment := &privatecloudv1alpha1.BMEnrollment{}
		err := reconciler.Get(ctx, req.NamespacedName, latestBMEnrollment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Error(err, "latest BM enrollment resource not found to update the status. Ignoring")
				return err
			}
			log.Error(err, "Failed to get latest BM enrollment resource to update the status. Re-running reconcile.")
			return err
		}

		if !equality.Semantic.DeepEqual(bmEnrollment.Status, latestBMEnrollment.Status) {
			log.Info("BM enrollment status update mismatch. Retrying")
			// update latest instance status
			bmEnrollment.Status.DeepCopyInto(&latestBMEnrollment.Status)
			if err := reconciler.Status().Update(ctx, latestBMEnrollment); err != nil {
				return fmt.Errorf("updateEnrollmentStatus: %w", err)
			}
		} else {
			log.Info("BM enrollment status does not need to be changed")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update  BM enrollment status: %w", err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *BMEnrollmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&privatecloudv1alpha1.BMEnrollment{},
			// Reconcile if spec changes.
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: reconciler.Cfg.MaxConcurrentReconciles,
		}).Complete(reconciler)
}
