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
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	"go.opentelemetry.io/otel/codes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	multierror "github.com/hashicorp/go-multierror"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	devicesManager "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/devices_manager"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
)

const (
	AfterCreateRequeuePeriodInSec = 1
)

const (
	PortChannelFinalizer = "idcnetwork.intel.com/portchannelfinalizer"
)

// PortChannelReconciler reconciles a PortChannel object
type PortChannelReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder

	Conf           idcnetworkv1alpha1.SDNControllerConfig
	DevicesManager devicesManager.DevicesAccessManager
	StatusReporter *statusreporter.StatusReporter
}

//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=portchannels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=portchannels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=portchannels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PortChannel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *PortChannelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("PortChannelReconciler.Reconcile").WithValues(utils.LogFieldResourceId, req.Name).Start()
	defer span.End()
	result, reconcileErr := func() (ctrl.Result, error) {
		portChannel := &idcnetworkv1alpha1.PortChannel{}
		err := r.Get(ctx, req.NamespacedName, portChannel)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("ignoring reconcile request because source CR was not found")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "unable to fetch PortChannel CR")
			return ctrl.Result{}, err
		}

		// process the changes to the Port channel
		processPortResult, processPortErr := func() (ctrl.Result, error) {
			if portChannel.ObjectMeta.DeletionTimestamp.IsZero() {
				return r.handleCreateOrUpdate(ctx, portChannel)
			} else {
				return r.handleDelete(ctx, portChannel)
			}
		}()

		// combine the errors
		processErr := multierror.Append(processPortErr)
		if len(processErr.Errors) == 0 {
			return processPortResult, nil
		}

		return processPortResult, processErr
	}()

	if reconcileErr != nil {
		span.SetStatus(codes.Error, reconcileErr.Error())
		logger.Error(reconcileErr, "PortChannelReconciler.Reconcile: error reconciling PortChannel")
	}

	return result, reconcileErr
}

// SetupWithManager sets up the controller with the Manager.
func (r *PortChannelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&idcnetworkv1alpha1.PortChannel{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		WithOptions(controller.Options{
			// This allows multiple ports / switches to be updated at once in parallel.
			// k8s guarantees only 1 reconcile at a time for any given port (see https://openkruise.io/blog/learning-concurrent-reconciling/).
			MaxConcurrentReconciles: r.Conf.ControllerConfig.MaxConcurrentReconciles,
		}).
		Complete(r)
}

//func (r *PortChannelReconciler) updateStatusAndPersist(ctx context.Context, portChannel *idcnetworkv1alpha1.PortChannel, reconcileErr error) error {
//	logger := log.FromContext(ctx).WithName("updateStatusAndPersist").WithValues(utils.LogFieldPortChannelCRName, portChannel.Name)
//	if reconcileErr == nil {
//		portChannel.Status.VlanId = portChannel.Spec.VlanId
//		portChannel.Status.Mode = portChannel.Spec.Mode
//		portChannel.Status.TrunkGroups = portChannel.Spec.TrunkGroups
//		portChannel.Status.NativeVlan = portChannel.Spec.NativeVlan
//	}
//
//	// Persist status.
//	if err := r.Status().Update(ctx, portChannel); err != nil {
//		logger.Error(err, "update PortChannel Status failed")
//		return fmt.Errorf("update PortChannel Status failed: %w", err)
//	}
//	//logger.V(1).Info("update portChannel status success!")
//	return nil
//}

func (r *PortChannelReconciler) handleCreateOrUpdate(ctx context.Context, portChannelCR *idcnetworkv1alpha1.PortChannel) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("portchannelController.handleCreateOrUpdate")
	var err error
	startTime := time.Now().UTC()
	defer func() {
		logger.V(1).Info(fmt.Sprintf("reconciliation ended for portchannel %v", portChannelCR.Spec.Name), utils.LogFieldTimeElapsed, time.Since(startTime))
	}()

	// Add the finalizer to the object
	if !utils.ContainsFinalizer(portChannelCR.ObjectMeta.Finalizers, PortChannelFinalizer) {
		portChannelCR.ObjectMeta.Finalizers = append(portChannelCR.ObjectMeta.Finalizers, PortChannelFinalizer)
		err = r.Update(ctx, portChannelCR) // Should we combine this with any further modifications to the CR?
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer to portchannel, %v", err)
		}
	}

	switchFQDN, found := portChannelCR.GetLabels()[idcnetworkv1alpha1.LabelNameSwitchFQDN]
	if !found {
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

	portChannelNumber, err := utils.PortChannelInterfaceNameToNumber(portChannelCR.Spec.Name)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("utils.PortChannelInterfaceNameToNumber failed, %v", err)
	}

	// If the port channel has no status, then it probably doesn't exist on the switch. Create it.
	// Unless it was "originally discovered" from the switch and never modified, in which case we don't want to recreate a switchport that was manually deleted from the switch.
	if portChannelCR.Status.Name == "" && r.sdnIsOwner(portChannelCR) {
		err := switchClient.CreatePortChannel(ctx, sc.CreatePortChannelRequest{
			PortChannel: int64(portChannelNumber),
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.CreatePortChannel failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name)
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeWarning, "switchClient.CreatePortChannel failed", err.Error())
			return ctrl.Result{}, fmt.Errorf("switchClient.CreatePortChannel failed, %v", err)
		}

		r.StatusReporter.AccelerateStatusUpdate(switchFQDN)
		return ctrl.Result{RequeueAfter: time.Duration(AfterCreateRequeuePeriodInSec) * time.Second}, nil
	}

	// Update "Mode" if needed
	var multiErrors *multierror.Error
	if portChannelCR.Spec.Mode != portChannelCR.Status.Mode && portChannelCR.Spec.Mode != "" {
		/*err = switchClient.UpdatePortChannelMode(ctx, sc.UpdatePortChannelModeRequest{
			Mode:        portChannelCR.Spec.Mode,
			PortChannel: int32(portChannelNumber),
		})*/
		err = switchClient.UpdateMode(ctx, sc.UpdateModeRequest{
			Mode:     portChannelCR.Spec.Mode,
			PortName: portChannelCR.Spec.Name,
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.UpdatePortChannelMode failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name, utils.LogFieldMode, portChannelCR.Spec.Mode)
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeWarning, "switchClient.UpdatePortChannelMode failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdatePortChannelMode failed, %v", err))
		} else {
			logger.Info(fmt.Sprintf("successfully updated portchannel Mode from %v to %v for switch %v, port %v", portChannelCR.Status.Mode, portChannelCR.Spec.Mode, switchFQDN, portChannelCR.Spec.Name), utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeNormal, "successfully updated portchannel Mode", fmt.Sprintf("successfully updated Mode from %v to %v", portChannelCR.Status.Mode, portChannelCR.Spec.Mode))

			// Since we just made a change, we want to accelerate the status update (we know it is going to find a change)
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	// Update "Description" if needed
	if portChannelCR.Spec.Description != portChannelCR.Status.Description && portChannelCR.Spec.Description != "" {
		err = switchClient.UpdateDescription(ctx, sc.UpdateDescriptionRequest{
			Description: portChannelCR.Spec.Description,
			PortName:    portChannelCR.Spec.Name,
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.UpdatePortChannelDescription failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name, utils.LogFieldDescription, portChannelCR.Spec.Description)
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeWarning, "switchClient.UpdatePortChannelDescription failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdatePortChannelDescription failed, %v", err))
		} else {
			logger.Info(fmt.Sprintf("successfully updated portChannel Description from %v to %v for switch %v, port %v", portChannelCR.Status.Description, portChannelCR.Spec.Description, switchFQDN, portChannelCR.Spec.Name), utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeNormal, "successfully updated description", fmt.Sprintf("successfully updated description from %v to %v", portChannelCR.Status.Description, portChannelCR.Spec.Description))

			// Since we just made a change, we want to accelerate the status update (we know it is going to find a change)
			logger.V(1).Info(fmt.Sprintf("just updated description on the switch. Accelerating status check for %s", switchFQDN))
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	// Update "Vlan" if needed
	if utils.ShouldUpdatePortChannelVlan(*portChannelCR) {
		err = switchClient.UpdateVlan(ctx, sc.UpdateVlanRequest{
			SwitchFQDN: switchFQDN,
			PortName:   portChannelCR.Spec.Name,
			Vlan:       int32(portChannelCR.Spec.VlanId),
			Env:        r.Conf.RavenConfig.Environment,
			UpdateLLDP: false,
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.UpdateVlan failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name, utils.LogFieldVlanID, portChannelCR.Spec.VlanId)
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeWarning, "switchClient.UpdateVlan failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdateVlan failed, %v", err))
		} else {
			logger.Info(fmt.Sprintf("successfully updated vlan from %v to %v for switch %v, port %v", portChannelCR.Status.VlanId, portChannelCR.Spec.VlanId, switchFQDN, portChannelCR.Spec.Name), utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeNormal, "successfully updated vlan", fmt.Sprintf("successfully updated vlan from %v to %v", portChannelCR.Status.VlanId, portChannelCR.Spec.VlanId))

			// Since we just made a change, we want to accelerate the status update (we know it is going to find a change)
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	// Update "TrunkGroups" if needed
	var specTrunkGroup []string
	if portChannelCR.Spec.TrunkGroups != nil {
		specTrunkGroup = *portChannelCR.Spec.TrunkGroups
	}
	statusTrunkGroup := portChannelCR.Status.TrunkGroups
	sort.Strings(specTrunkGroup)
	sort.Strings(statusTrunkGroup)

	if portChannelCR.Spec.TrunkGroups != nil && !slices.Equal(specTrunkGroup, statusTrunkGroup) {

		err = switchClient.UpdateTrunkGroups(ctx, sc.UpdateTrunkGroupsRequest{
			SwitchFQDN:  switchFQDN,
			PortName:    portChannelCR.Spec.Name,
			TrunkGroups: specTrunkGroup,
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.UpdateTrunkGroups failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name, utils.LogFieldTrunkGroups, portChannelCR.Spec.TrunkGroups)
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeWarning, "switchClient.UpdateTrunkGroups failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdateTrunkGroups failed, %v", err))
		} else {
			logmsg := fmt.Sprintf("successfully updated trunkGroups from %v to %v for switch %v, portchannel %v", portChannelCR.Status.TrunkGroups, portChannelCR.Spec.TrunkGroups, switchFQDN, portChannelCR.Spec.Name)
			logger.Info(logmsg, utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeNormal, "successfully updated portchannel trunkGroups", logmsg)
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	// Update "NativeVlan" if needed
	if portChannelCR.Spec.NativeVlan != idcnetworkv1alpha1.NOOPVlanID && portChannelCR.Spec.NativeVlan != 0 && portChannelCR.Spec.NativeVlan != portChannelCR.Status.NativeVlan {
		err = switchClient.UpdateNativeVlan(ctx, sc.UpdateNativeVlanRequest{
			SwitchFQDN: switchFQDN,
			PortName:   portChannelCR.Spec.Name,
			NativeVlan: int32(portChannelCR.Spec.NativeVlan),
		})
		if err != nil {
			logger.Error(fmt.Errorf("switchClient.UpdateNativeVlan failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name, utils.LogFieldNativeVlan, portChannelCR.Spec.NativeVlan)
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeWarning, "switchClient.UpdateNativeVlan failed", err.Error())
			multiErrors = multierror.Append(fmt.Errorf("switchClient.UpdateNativeVlan failed, %v", err))
		} else {
			logmsg := fmt.Sprintf("successfully updated NativeVlan from %v to %v for switch %v, portchannel %v", portChannelCR.Status.NativeVlan, portChannelCR.Spec.NativeVlan, switchFQDN, portChannelCR.Spec.Name)
			logger.Info(logmsg, utils.LogFieldTimeElapsed, time.Since(startTime))
			r.EventRecorder.Event(portChannelCR, corev1.EventTypeNormal, "successfully updated portchannel NativeVlan", logmsg)
			shouldAccelerateSwitchStatusCheck = true
		}
	}

	if shouldAccelerateSwitchStatusCheck {
		logger.V(1).Info(fmt.Sprintf("just updated a portchannel's config on the switch. Accelerating status check for %s", switchFQDN))
		r.StatusReporter.AccelerateStatusUpdate(switchFQDN)
	}

	err = multiErrors.ErrorOrNil()
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Duration(r.Conf.ControllerConfig.PortResyncPeriodInSec) * time.Second}, nil
}

func (r *PortChannelReconciler) handleDelete(ctx context.Context, portChannelCR *idcnetworkv1alpha1.PortChannel) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("portchannelController.handleDelete")
	var err error
	startTime := time.Now().UTC()
	defer func() {
		logger.V(1).Info(fmt.Sprintf("reconciliation (deletion) ended for portchannel %v", portChannelCR.Spec.Name), utils.LogFieldTimeElapsed, time.Since(startTime))
	}()

	switchFQDN, found := portChannelCR.GetLabels()[idcnetworkv1alpha1.LabelNameSwitchFQDN]
	if !found {
		return ctrl.Result{}, fmt.Errorf("switch_fqdn label not provided")
	}

	// Only delete the PortChannel on the switch if we ever actually took control of it.
	// Creating & immediately deleting a portChannel CR shouldn't remove a portChannel that already existed on the Switch.
	// If "desired" fields are all empty, do not remove it from any switchPorts or delete it from the switch.
	if !r.sdnIsOwner(portChannelCR) {
		portChannelCR.ObjectMeta.Finalizers = utils.RemoveFinalizer(portChannelCR.ObjectMeta.Finalizers, PortChannelFinalizer)
		if err := r.Update(ctx, portChannelCR); err != nil {
			logger.Error(fmt.Errorf("removeFinalizer failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name)
			return ctrl.Result{}, err
		}

		// Schedule a status update for the switch
		r.StatusReporter.AccelerateStatusUpdate(switchFQDN)
		return ctrl.Result{}, nil
	}

	switchClient, err := r.DevicesManager.GetSwitchClient(devicesManager.GetOption{
		SwitchFQDN: switchFQDN,
	})
	if err != nil {
		logger.Error(err, "DevicesManager.GetSwitchClient failed")
		return ctrl.Result{RequeueAfter: time.Duration(ErrorRequeuePeriodInSec) * time.Second}, nil
	}

	pcNumber, err := utils.PortChannelInterfaceNameToNumber(portChannelCR.Spec.Name)
	if err != nil {
		logger.Error(fmt.Errorf("PortChannelInterfaceNameToNumber failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name)
		return ctrl.Result{}, fmt.Errorf("PortChannelInterfaceNameToNumber failed, %v", err)
	}

	// Remove portchannel from all the switchports that reference this PortChannel.
	labelSelector := labels.SelectorFromSet(labels.Set{
		idcnetworkv1alpha1.LabelNameSwitchFQDN: switchFQDN,
	})

	listOpts := &client.ListOptions{
		Namespace:     portChannelCR.Namespace,
		LabelSelector: labelSelector,
	}
	switchPortCRsOnThisSwitch := &idcnetworkv1alpha1.SwitchPortList{}
	err = r.List(ctx, switchPortCRsOnThisSwitch, listOpts)
	if err != nil {
		logger.Error(err, "networkK8sClient.List switchports failed", utils.LogFieldSwitchFQDN, switchFQDN)
		return ctrl.Result{}, fmt.Errorf("networkK8sClient.List switchports failed, %v", err)
	}

	for _, spCR := range switchPortCRsOnThisSwitch.Items {
		if spCR.Spec.PortChannel == int64(pcNumber) {
			spCR.Spec.PortChannel = 0 // Explicity set the desired portChannel to 0 (meaning remove from portChannel).
			err := r.Update(ctx, &spCR)
			if err != nil {
				logger.Error(fmt.Errorf("failed to updating switchPortCR %s to remove portChannel, %v", spCR.Name, err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name)
				return ctrl.Result{}, fmt.Errorf("failed to update switchPortCR %s to remove portChannel, %v", spCR.Name, err)
			}
		}
	}

	// Remove the portChannel from the switch
	err = switchClient.DeletePortChannel(ctx, sc.DeletePortChannelRequest{
		TargetPortChannel: int64(pcNumber),
	})
	if err != nil {
		logger.Error(fmt.Errorf("switchClient.DeletePortChannel failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name)
		return ctrl.Result{}, fmt.Errorf("switchClient.DeletePortChannel failed, %v", err)
	}

	// Remove the finalizer now that deletion is completed
	portChannelCR.ObjectMeta.Finalizers = utils.RemoveFinalizer(portChannelCR.ObjectMeta.Finalizers, PortChannelFinalizer)
	if err := r.Update(ctx, portChannelCR); err != nil {
		logger.Error(fmt.Errorf("removeFinalizer failed, %v", err), "", utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldPortChannelCRName, portChannelCR.Spec.Name)
		return ctrl.Result{}, err
	}

	// Schedule a status update for the switch
	r.StatusReporter.AccelerateStatusUpdate(switchFQDN)

	return ctrl.Result{}, nil
}

func (r *PortChannelReconciler) sdnIsOwner(portChannelCR *idcnetworkv1alpha1.PortChannel) bool {
	return portChannelCR.Spec.Mode != "" ||
		(portChannelCR.Spec.VlanId != 0 && portChannelCR.Spec.VlanId != idcnetworkv1alpha1.NOOPVlanID) ||
		portChannelCR.Spec.TrunkGroups != nil
}
