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
	"time"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"go.opentelemetry.io/otel/codes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	devicesmanager "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/devices_manager"
	statusreporter "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/status_reporter"
	switchclients "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	SwitchFinalizer = "idcnetwork.intel.com/switchfinalizer"

	ErrorRequeueSec = 5
)

// SwitchReconciler reconciles a Switch object
type SwitchReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder

	Conf                 idcnetworkv1alpha1.SDNControllerConfig
	DevicesAccessManager devicesmanager.DevicesAccessManager
	StatusReporter       *statusreporter.StatusReporter
}

//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=devices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=devices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=devices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *SwitchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("SwitchReconciler.Reconcile").WithValues(utils.LogFieldResourceId, req.Name).Start()
	defer span.End()

	result, reconcileErr := func() (ctrl.Result, error) {
		sw := &idcnetworkv1alpha1.Switch{}
		err := r.Get(ctx, req.NamespacedName, sw)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("ignoring reconcile request because source CR was not found")
				// if we found non-existing object, do one more round of cleanup to make sure all resources are removed.
				cleanupErr := r.cleanup(req.Name)
				if cleanupErr != nil {
					logger.Error(err, "cleanup resources failed")
					return ctrl.Result{RequeueAfter: time.Duration(ErrorRequeueSec) * time.Second}, nil
				}
				return ctrl.Result{}, nil
			}
			logger.Error(err, "unable to fetch Switch CR")
			return ctrl.Result{}, err
		}

		processResult, processErr := func() (ctrl.Result, error) {
			if sw.ObjectMeta.DeletionTimestamp.IsZero() {
				return r.handleCreateOrUpdate(ctx, sw)
			} else {
				return r.handleDelete(ctx, sw)
			}
		}()

		return processResult, processErr
	}()

	if reconcileErr != nil {
		span.SetStatus(codes.Error, reconcileErr.Error())
		logger.Error(reconcileErr, "SwitchReconciler.Reconcile: error reconciling Switch")
	}
	return result, reconcileErr
}

// SetupWithManager sets up the controller with the Manager.
func (r *SwitchReconciler) SetupWithManager(mgr ctrl.Manager) error {

	b := ctrl.NewControllerManagedBy(mgr)
	b.For(
		&idcnetworkv1alpha1.Switch{},
		builder.WithPredicates(predicate.GenerationChangedPredicate{}),
	).WithOptions(controller.Options{
		MaxConcurrentReconciles: r.Conf.ControllerConfig.MaxConcurrentSwitchReconciles,
	})
	err := b.Complete(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *SwitchReconciler) handleCreateOrUpdate(ctx context.Context, sw *idcnetworkv1alpha1.Switch) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("SwitchReconciler.handleCreateOrUpdate")

	// Add the finalizer to the object
	if !utils.ContainsFinalizer(sw.ObjectMeta.Finalizers, SwitchFinalizer) {
		sw.ObjectMeta.Finalizers = append(sw.ObjectMeta.Finalizers, SwitchFinalizer)
	}

	startDeviceManager := time.Now().UTC()

	isMaintenanceMode := sw.Spec.Maintenance == "true"
	if isMaintenanceMode {
		logger.Info(fmt.Sprintf("Switch: %s is in maintenance mode, Skipping DeviceManager operations for the switch", sw.Spec.FQDN))
		err := r.cleanup(sw.Spec.FQDN)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("cleanup of resource %v in maintenance mode failed, %v", sw.Spec.FQDN, err)
		}

		return ctrl.Result{}, nil
	}

	// add the sw to the device access manager
	err := r.DevicesAccessManager.AddOrUpdateSwitch(sw, false)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("DevicesAccessManager.AddOrUpdateSwitch failed, switch: %v, %v", sw.Spec.FQDN, err)
	}
	logger.V(1).Info(fmt.Sprintf("finished AddOrUpdateSwitch, timeElapsed %v, switch: %v", time.Since(startDeviceManager), sw.Spec.FQDN))

	// add the sw to the status reporter.
	err = r.StatusReporter.AddSwitch(sw.Spec.FQDN)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("StatusReporter.AddSwitch failed, switch: %v, %v", sw.Spec.FQDN, err)
	}

	switchClient, err := r.DevicesAccessManager.GetSwitchClient(devicesmanager.GetOption{
		SwitchFQDN: sw.Spec.FQDN,
	})
	if err != nil {
		logger.Error(err, "DevicesManager.GetSwitchClient failed")
		return ctrl.Result{RequeueAfter: time.Duration(ErrorRequeueSec) * time.Second}, nil
	}

	// reconcile BGP if needed
	if sw.Spec.BGP != nil && sw.Status.SwitchBGPConfigStatus != nil && sw.Spec.BGP.BGPCommunity != idcnetworkv1alpha1.NOOPBGPCommunity && sw.Spec.BGP.BGPCommunity != 0 {

		if sw.Spec.BGP.BGPCommunity != sw.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity {
			logger.V(1).Info(fmt.Sprintf("Switch Spec.BGP.BGPCommunity [%v] != observed [%v], performing BGP update", sw.Spec.BGP.BGPCommunity, sw.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity))
			req := switchclients.UpdateBGPCommunityRequest{
				BGPCommunity:                  int32(sw.Spec.BGP.BGPCommunity),
				BGPCommunityIncomingGroupName: r.Conf.ControllerConfig.BGPCommunityIncomingGroupName,
			}
			err = switchClient.UpdateBGPCommunity(ctx, req)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("UpdateBGPCommunity failed, request: %v, reason: %v", req, err)
			}
			logger.Info(fmt.Sprintf("successfully updated BGP on switch from %v to %v", sw.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity, sw.Spec.BGP.BGPCommunity), utils.LogFieldSwitchFQDN, sw.Spec.FQDN)

			// Since we just made a change, we want to accelerate the status update (we know it is going to find a change)
			logger.V(1).Info(fmt.Sprintf("just updated BGP on the switch. Accelerating status check for %s", sw.Spec.FQDN))
			r.StatusReporter.AccelerateStatusUpdate(sw.Spec.FQDN)
		}
	}

	return ctrl.Result{RequeueAfter: time.Duration(r.Conf.ControllerConfig.SwitchResyncPeriodInSec) * time.Second}, nil
}

func (r *SwitchReconciler) handleDelete(ctx context.Context, sw *idcnetworkv1alpha1.Switch) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("SwitchReconciler.handleDelete")
	// Note: Switch CRD is currently used only for device access management. Removing a Switch doesn't involve the removal of Port CR.
	// Will leave it to the BM enrollment/de-enrollment process to maintain the creation/deletion of a Port.
	err := r.cleanup(sw.Spec.FQDN)
	if err != nil {
		logger.Error(err, "cleanup resources failed")
		return ctrl.Result{RequeueAfter: time.Duration(ErrorRequeueSec) * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

func (r *SwitchReconciler) cleanup(swFQDN string) error {

	err := r.DevicesAccessManager.DeleteSwitch(swFQDN)
	if err != nil {
		return fmt.Errorf("DevicesAccessManager.DeleteSwitch failed, switch: %v, %v", swFQDN, err)
	}

	err = r.StatusReporter.RemoveSwitch(swFQDN)
	if err != nil {
		return fmt.Errorf("StatusReporter.StopReportStatusFor failed, switch: %v, %v", swFQDN, err)
	}
	return nil
}
