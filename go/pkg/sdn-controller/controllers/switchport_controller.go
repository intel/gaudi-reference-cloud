// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright 2023.

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
	"fmt"
	"sort"
	"time"

	statusreporter "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/status_reporter"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	multierror "github.com/hashicorp/go-multierror"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	devicesManager "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/devices_manager"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SwitchPortReconciler reconciles a SwitchPort object
type SwitchPortReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder

	Conf           idcnetworkv1alpha1.SDNControllerConfig
	DevicesManager devicesManager.DevicesAccessManager
	StatusReporter *statusreporter.StatusReporter
}

const (
	ErrorRequeuePeriodInSec = 5
)

//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=switchports,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=switchports/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=switchports/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *SwitchPortReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("SwitchPortReconciler.Reconcile").WithValues(utils.LogFieldResourceId, req.Name).Start()
	defer span.End()
	result, reconcileErr := func() (ctrl.Result, error) {
		switchPortCR := &idcnetworkv1alpha1.SwitchPort{}
		err := r.Get(ctx, req.NamespacedName, switchPortCR)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("ignoring reconcile request because source CR was not found")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "unable to fetch SwitchPort CR")
			return ctrl.Result{}, err
		}

		// process the changes to the Switch Port
		processPortResult, processPortErr := func() (ctrl.Result, error) {
			if switchPortCR.ObjectMeta.DeletionTimestamp.IsZero() {
				return r.handleCreateOrUpdate(ctx, switchPortCR)
			} else {
				return r.handleDelete(ctx, switchPortCR)
			}
		}()

		return processPortResult, processPortErr
	}()

	if reconcileErr != nil {
		span.SetStatus(codes.Error, reconcileErr.Error())
		logger.Error(reconcileErr, "SwitchPortReconciler.Reconcile: error reconciling SwitchPort")
	}
	return result, reconcileErr
}

// SetupWithManager sets up the controller with the Manager.
func (r *SwitchPortReconciler) SetupWithManager(mgr ctrl.Manager) error {
	maxConcurrent := r.Conf.ControllerConfig.MaxConcurrentReconciles
	return ctrl.NewControllerManagedBy(mgr).
		For(&idcnetworkv1alpha1.SwitchPort{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		WithOptions(controller.Options{
			// This allows multiple ports / switches to be updated at once in parallel.
			// k8s guarantees only 1 reconcile at a time for any given port (see https://openkruise.io/blog/learning-concurrent-reconciling/).
			MaxConcurrentReconciles: maxConcurrent,
			RateLimiter: workqueue.NewMaxOfRateLimiter(
				workqueue.NewItemExponentialFailureRateLimiter(1*time.Second, 1000*time.Second),
				&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(float64(maxConcurrent)), maxConcurrent)},
			),
		}).
		Complete(r)
}

func (r *SwitchPortReconciler) handleDelete(ctx context.Context, switchPortCR *idcnetworkv1alpha1.SwitchPort) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *SwitchPortReconciler) handleCreateOrUpdate(ctx context.Context, switchPortCR *idcnetworkv1alpha1.SwitchPort) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("SwitchPortReconciler.handleCreateOrUpdate")
	var err error
	startTime := time.Now().UTC()
	defer func() {
		logger.V(1).Info(fmt.Sprintf("reconciliation completed for switchport %v", switchPortCR.Spec.Name), utils.LogFieldTimeElapsed, time.Since(startTime))
	}()

	switchFQDN, found := switchPortCR.GetLabels()[idcnetworkv1alpha1.LabelNameSwitchFQDN]
	if !found {
		// when we got an error during the reconciliation, we have a few ways to return the result:
		// 1. return an empty result, and a non-nil err. ie `return ctrl.Result{}, err`
		// 		this will requeue the reconciliation request with exponential backoff, which will start with a 5 milliseconds and end with waiting for 1000 second after a specific number of retries.
		//		this is the common way of returning the result when we got an error.
		// 2. retrun a non-zero result and a nil err. ie `return ctrl.Result{RequeueAfter: time.Duration(ErrorRequeuePeriodInSec) * time.Second}, nil`
		//		this will requeue the reconciliation request with a fixed duration, for example, 5 seconds for ErrorRequeuePeriodInSec.
		//		use this approach when some resource or dependency is not yet ready or something is still in progress.
		//		for example, during the startup of SDN, DevicesManager need to initialize the switch clients, before this is done, DevicesManager.GetSwitchClient() failure is expected and should be fixed in the next reconciliation.
		// 3. return immediately without an error. ie `return ctrl.Result{Requeue: true}, nil`
		//		this approach is not used in error-cases in the existing contorollers.

		// eapi/TACACS consideration: when the controller got error calling the eapi switch client(to update switch configuration), we should use approach 1 to return result,
		// as it generate less retry calls to the switches and void overwhelming the TACACS.
		// below is the estimation of total calls to the switches if we got a consistent error from calling eapi.
		// for exponential back off, in 1 hours, a total number of around 22 eapi calls will be created.
		// for 5 seconds fixed duration requeue, in 1 hour, a total number of 720 eapi calls will be created.

		// if we provide both a non-zero result and a non-nil error, the result will be ignored, and below warning message will be printed out.
		// Warning: Reconciler returned both a non-zero result and a non-nil error. The result will always be ignored if the error is non-nil and the non-nil error causes reqeueuing with exponential backoff. For more details, see: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile#Reconciler
		return ctrl.Result{}, fmt.Errorf("switch_fqdn label not provided")
	}

	switchClient, err := r.DevicesManager.GetSwitchClient(devicesManager.GetOption{
		SwitchFQDN: switchFQDN,
	})
	if err != nil {
		logger.Error(err, "DevicesManager.GetSwitchClient failed")
		return ctrl.Result{RequeueAfter: time.Duration(ErrorRequeuePeriodInSec) * time.Second}, nil
	}

	var shouldAccelerateSwitchStatusCheck = false

	// we will try update all the changes needed, as long as one of them failed, we will requeue it.
	var multiErrors *multierror.Error

	// Update "Mode" if needed
	if switchPortCR.Spec.Mode != switchPortCR.Status.Mode && switchPortCR.Spec.Mode != "" && switchPortCR.Spec.Mode != "routed" && switchPortCR.Status.Mode != "routed" { // Converting between "routed" & other modes not yet supported.
		err = switchClient.UpdateMode(ctx, sc.UpdateModeRequest{
			Mode:       switchPortCR.Spec.Mode,
			SwitchFQDN: switchFQDN,
			PortName:   switchPortCR.Spec.Name,
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.UpdateMode failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortCRName, switchPortCR.Spec.Name, utils.LogFieldMode, switchPortCR.Spec.Mode)
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeWarning, "switchClient.UpdateMode failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdateMode failed, %v", err))
		} else {
			logger.Info(fmt.Sprintf("successfully updated Mode from %v to %v for switch %v, port %v", switchPortCR.Status.Mode, switchPortCR.Spec.Mode, switchFQDN, switchPortCR.Spec.Name), utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeNormal, "successfully updated Mode", fmt.Sprintf("successfully updated Mode from %v to %v", switchPortCR.Status.Mode, switchPortCR.Spec.Mode))

			// Since we just made a change, we want to accelerate the status update (we know it is going to find a change)
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	// Update "Description" if needed
	if switchPortCR.Spec.Description != switchPortCR.Status.Description && switchPortCR.Spec.Description != "" {
		err = switchClient.UpdateDescription(ctx, sc.UpdateDescriptionRequest{
			Description: switchPortCR.Spec.Description,
			SwitchFQDN:  switchFQDN,
			PortName:    switchPortCR.Spec.Name,
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.Update Description failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortCRName, switchPortCR.Spec.Name, utils.LogFieldDescription, switchPortCR.Spec.Description)
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeWarning, "switchClient.UpdateDescription failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdateDescription failed, %v", err))
		} else {
			logger.Info(fmt.Sprintf("successfully updated description from %v to %v for switch %v, port %v", switchPortCR.Status.Description, switchPortCR.Spec.Description, switchFQDN, switchPortCR.Spec.Name), utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeNormal, "successfully updated description", fmt.Sprintf("successfully updated description from %v to %v", switchPortCR.Status.Description, switchPortCR.Spec.Description))

			// Since we just made a change, we want to accelerate the status update (we know it is going to find a change)
			logger.V(1).Info(fmt.Sprintf("just updated description on the switch. Accelerating status check for %s", switchFQDN))
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	// Update "Vlan" if needed
	if utils.ShouldUpdateSwitchPortVlan(*switchPortCR) {
		err = switchClient.UpdateVlan(ctx, sc.UpdateVlanRequest{
			SwitchFQDN: switchFQDN,
			PortName:   switchPortCR.Spec.Name,
			Vlan:       int32(switchPortCR.Spec.VlanId),
			Env:        r.Conf.RavenConfig.Environment,
			UpdateLLDP: true,
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.UpdateVlan failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortCRName, switchPortCR.Spec.Name, utils.LogFieldVlanID, switchPortCR.Spec.VlanId)
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeWarning, "switchClient.UpdateVlan failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdateVlan failed, %v", err))
		} else {
			logger.Info(fmt.Sprintf("successfully updated vlan from %v to %v for switch %v, port %v", switchPortCR.Status.VlanId, switchPortCR.Spec.VlanId, switchFQDN, switchPortCR.Spec.Name), utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeNormal, "successfully updated vlan", fmt.Sprintf("successfully updated vlan from %v to %v", switchPortCR.Status.VlanId, switchPortCR.Spec.VlanId))

			// Since we just made a change, we want to accelerate the status update (we know it is going to find a change)
			logger.V(1).Info(fmt.Sprintf("just updated vlan on the switch. Accelerating status check for %s", switchFQDN))
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	// Update "TrunkGroups" if needed
	var specTrunkGroup []string
	if switchPortCR.Spec.TrunkGroups != nil {
		specTrunkGroup = *switchPortCR.Spec.TrunkGroups
	}
	statusTrunkGroup := switchPortCR.Status.TrunkGroups
	sort.Strings(specTrunkGroup)
	sort.Strings(statusTrunkGroup)

	if switchPortCR.Spec.TrunkGroups != nil && !slices.Equal(specTrunkGroup, statusTrunkGroup) {
		err = switchClient.UpdateTrunkGroups(ctx, sc.UpdateTrunkGroupsRequest{
			SwitchFQDN:  switchFQDN,
			PortName:    switchPortCR.Spec.Name,
			TrunkGroups: specTrunkGroup,
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.UpdateTrunkGroups failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortCRName, switchPortCR.Spec.Name, utils.LogFieldTrunkGroups, switchPortCR.Spec.TrunkGroups)
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeWarning, "switchClient.UpdateTrunkGroups failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdateTrunkGroups failed, %v", err))
		} else {
			logmsg := fmt.Sprintf("successfully updated trunkGroups from %v to %v for switch %v, port %v", switchPortCR.Status.TrunkGroups, switchPortCR.Spec.TrunkGroups, switchFQDN, switchPortCR.Spec.Name)
			logger.Info(logmsg, utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeNormal, "successfully updated trunkGroups", logmsg)
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	// Update "NativeVlan" if needed
	if switchPortCR.Spec.NativeVlan != idcnetworkv1alpha1.NOOPVlanID && switchPortCR.Spec.NativeVlan != 0 && switchPortCR.Spec.NativeVlan != switchPortCR.Status.NativeVlan {
		err = switchClient.UpdateNativeVlan(ctx, sc.UpdateNativeVlanRequest{
			SwitchFQDN: switchFQDN,
			PortName:   switchPortCR.Spec.Name,
			NativeVlan: int32(switchPortCR.Spec.NativeVlan),
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.UpdateNativeVlan failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortCRName, switchPortCR.Spec.Name, utils.LogFieldNativeVlan, switchPortCR.Spec.NativeVlan)
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeWarning, "switchClient.UpdateNativeVlan failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdateNativeVlan failed, %v", err))
		} else {
			logmsg := fmt.Sprintf("successfully updated NativeVlan from %v to %v for switch %v, port %v", switchPortCR.Status.NativeVlan, switchPortCR.Spec.NativeVlan, switchFQDN, switchPortCR.Spec.Name)
			logger.Info(logmsg, utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(switchPortCR, corev1.EventTypeNormal, "successfully updated NativeVlan", logmsg)
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	if switchPortCR.Spec.PortChannel != switchPortCR.Status.PortChannel && switchPortCR.Spec.PortChannel != idcnetworkv1alpha1.NOOPPortChannel && r.Conf.ControllerConfig.PortChannelsEnabled {
		if switchPortCR.Spec.PortChannel == 0 {
			err = switchClient.RemoveSwitchPortFromPortChannel(ctx, sc.RemoveSwitchPortFromPortChannelRequest{
				SwitchPort: switchPortCR.Spec.Name,
			})
			if err != nil {
				logger.Error(fmt.Errorf("switchClient.RemoveSwitchPortFromPortChannel failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortCRName, switchPortCR.Spec.Name, utils.LogFieldVlanID, switchPortCR.Spec.VlanId)
				r.EventRecorder.Event(switchPortCR, corev1.EventTypeWarning, "switchClient.RemoveSwitchPortFromPortChannel failed", err.Error())
				multiErrors = multierror.Append(fmt.Errorf("switchClient.RemoveSwitchPortFromPortChannel failed, %v", err))
			} else {
				logger.Info(fmt.Sprintf("successfully removed switch port %v from port-channel %v on switch", switchPortCR.Spec.Name, switchPortCR.Status.PortChannel))
				r.EventRecorder.Event(switchPortCR, corev1.EventTypeNormal, "successfully removed switch port from PortChannel", fmt.Sprintf("successfully removed switch port %v from port-channel %v on switch", switchPortCR.Spec.Name, switchPortCR.Status.PortChannel))
				shouldAccelerateSwitchStatusCheck = true
			}
		} else {
			// Check PortChannel CR exists. Do not modify otherwise (to prevent portChannel being recreated on the switch automatically, without a corresponding portChannel CR)
			pcName, err := utils.PortChannelNumberAndSwitchFQDNToCRName(int(switchPortCR.Spec.PortChannel), switchFQDN)
			if err != nil {
				logger.Error(fmt.Errorf("PortChannelNumberToInterfaceName failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortCRName, switchPortCR.Spec.Name)
				r.EventRecorder.Event(switchPortCR, corev1.EventTypeWarning, "PortChannelNumberToInterfaceName failed", err.Error())
				multiErrors = multierror.Append(fmt.Errorf("PortChannelNumberToInterfaceName failed, %v", err))
			}
			existingPortChannelCR := &idcnetworkv1alpha1.PortChannel{}
			err = r.Get(ctx, types.NamespacedName{Name: pcName, Namespace: switchPortCR.Namespace}, existingPortChannelCR)
			if err != nil || existingPortChannelCR.Name == "" {
				errmsg := fmt.Sprintf("couldn't get portChannel %s from k8s. Porchannel CR must exist to add a switchport to it. %v", pcName, err)
				logger.Error(fmt.Errorf(errmsg), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortCRName, switchPortCR.Spec.Name, utils.LogFieldPortChannelInterfaceName, switchPortCR.Spec.PortChannel)
				r.EventRecorder.Event(switchPortCR, corev1.EventTypeWarning, "couldn't get portChannel from k8s", errmsg)
				multiErrors = multierror.Append(fmt.Errorf(errmsg, err))
			} else {
				err = switchClient.AssignSwitchPortToPortChannel(ctx, sc.AssignSwitchPortToPortChannelRequest{
					SwitchPort:        switchPortCR.Spec.Name,
					TargetPortChannel: switchPortCR.Spec.PortChannel,
				})
				if err != nil {
					logger.Error(fmt.Errorf("switchClient.AssignSwitchPortToPortChannel failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortCRName, switchPortCR.Spec.Name, utils.LogFieldPortChannel, switchPortCR.Spec.PortChannel)
					r.EventRecorder.Event(switchPortCR, corev1.EventTypeWarning, "switchClient.AssignSwitchPortToPortChannel failed", err.Error())
					multiErrors = multierror.Append(fmt.Errorf("switchClient.AssignSwitchPortToPortChannel failed, %v", err))
				} else {
					logger.Info(fmt.Sprintf("successfully assigned switch port %v to port-channel %v on switch", switchPortCR.Spec.Name, switchPortCR.Spec.PortChannel))
					r.EventRecorder.Event(switchPortCR, corev1.EventTypeNormal, "successfully updated switch port PortChannel", fmt.Sprintf("successfully assigned switch port %v to port-channel %v on switch", switchPortCR.Spec.Name, switchPortCR.Spec.PortChannel))
					shouldAccelerateSwitchStatusCheck = true
				}
			}
		}
	}

	if shouldAccelerateSwitchStatusCheck {
		logger.V(1).Info(fmt.Sprintf("just updated a switchport's config on the switch. Accelerating status check for %s", switchFQDN))
		r.StatusReporter.AccelerateStatusUpdate(switchFQDN)
	}

	err = multiErrors.ErrorOrNil()
	if err != nil {
		return ctrl.Result{}, err
	}
	// reconcile done, no error found, requeue after PortResyncPeriodInSec
	return ctrl.Result{RequeueAfter: time.Duration(r.Conf.ControllerConfig.PortResyncPeriodInSec) * time.Second}, nil
}
