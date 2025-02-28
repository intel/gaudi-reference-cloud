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
	"reflect"
	"sort"
	"time"

	"go.opentelemetry.io/otel/codes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/strings/slices"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	multierror "github.com/hashicorp/go-multierror"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	devicesmanager "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/devices_manager"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/pools"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

// NetworkNodeReconciler reconciles a NetworkNode object
type NetworkNodeReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder

	DevicesManager       devicesmanager.DevicesAccessManager
	Conf                 idcnetworkv1alpha1.SDNControllerConfig
	PoolManager          *pools.PoolManager
	AllowedVlanIds       []int
	AllowedNativeVlanIds []int
}

type AlreadyOwnedError struct {
	switchPortName    string
	fabricType        string
	existingOwnerName string
}

func (e *AlreadyOwnedError) Error() string {
	return fmt.Sprintf("existing %s switchport %s is already owned by %v. Not updating it", e.fabricType, e.switchPortName, e.existingOwnerName)
}

//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=networknodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=networknodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=networknodes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *NetworkNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("NetworkNodeReconciler.Reconcile").WithValues(utils.LogFieldResourceId, req.Name).Start()
	defer span.End()
	result, reconcileErr := func() (ctrl.Result, error) {
		networkNode := &idcnetworkv1alpha1.NetworkNode{}
		err := r.Get(ctx, req.NamespacedName, networkNode)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("ignoring reconcile request because source CR was not found")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "unable to fetch NetworkNode CR")
			return ctrl.Result{}, err
		}

		// process the changes to the networkNode
		processPortResult, processPortErr := func() (ctrl.Result, error) {
			if networkNode.ObjectMeta.DeletionTimestamp.IsZero() {
				return r.handleCreateOrUpdate(ctx, networkNode)
			} else {
				return r.handleDelete(ctx, networkNode)
			}
		}()

		// update status
		updateStatusErr := r.updateStatusAndPersist(ctx, networkNode, processPortErr)

		// combine the errors
		processErr := multierror.Append(processPortErr, updateStatusErr)
		if len(processErr.Errors) == 0 {
			return processPortResult, nil
		}

		return processPortResult, processErr
	}()

	if reconcileErr != nil {
		span.SetStatus(codes.Error, reconcileErr.Error())
		logger.Error(reconcileErr, "NetworkNodeReconciler.Reconcile: error reconciling NetworkNode")
	}
	return result, reconcileErr
}

// SetupWithManager sets up the controller with the Manager.
func (r *NetworkNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&idcnetworkv1alpha1.NetworkNode{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Conf.ControllerConfig.MaxConcurrentNetworkNodeReconciles,
		}).
		Watches(&idcnetworkv1alpha1.SwitchPort{},
			&switchPortEventHandler{}).
		Complete(r)
}

func (r *NetworkNodeReconciler) handleDelete(ctx context.Context, networkNode *idcnetworkv1alpha1.NetworkNode) (ctrl.Result, error) {
	// TODO: do another round of switchPort CR deletion by checking the label, to make sure there is no orphan data.
	return ctrl.Result{}, nil
}

// TODO: removing an item from NetworkNode's acceleratorFabric.switchPorts won't remove the related SwitchPort CR, consider support this in the future.
func (r *NetworkNodeReconciler) handleCreateOrUpdate(ctx context.Context, networkNode *idcnetworkv1alpha1.NetworkNode) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("handleCreateOrUpdate").WithValues(utils.LogFieldNetworkNode, networkNode.Name)
	var err error
	multiErrors := &multierror.Error{}

	frontEndPortReadyCount := 0
	frontEndPortTotalCount := 1
	accelPortReadyCount := 0
	accelPortTotalCount := 0
	storagePortReadyCount := 0
	storagePortTotalCount := 0
	defer func() {
		r.setStatusReadiness(ctx, frontEndPortReadyCount, frontEndPortTotalCount, accelPortReadyCount, accelPortTotalCount, storagePortReadyCount, storagePortTotalCount, networkNode)
	}()

	// Is the front end fabric port previously created in the k8s?
	// 	no, create new SP
	//	yes,
	//		is the k8s SwitchPort spec's vlan the same as the desired vlan in the NetworkNode?
	//			yes,
	//				Is the k8s SwitchPort's status in-sync with the desired value?
	// 					yes, frontEndPortReadyCount++. continue to check the accelerator fabric
	// 					no, skip/continue (come back and check again)
	//			no, update the k8s SwitchPort spec
	//				append to "portsToUpdate"

	// FrontEnd Fabric
	func() {
		if networkNode.Spec.FrontEndFabric != nil {
			// try to get the SP CR from K8s.
			frontEndSwitchPortName := networkNode.Spec.FrontEndFabric.SwitchPort
			// the frontEndSwitchPortName is empty, no need to process
			if len(frontEndSwitchPortName) == 0 {
				return
			}

			key := types.NamespacedName{Name: frontEndSwitchPortName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			frontEndSwitchPortCR := &idcnetworkv1alpha1.SwitchPort{}
			objectExists := true
			err = r.Get(ctx, key, frontEndSwitchPortCR)
			if err != nil {
				if client.IgnoreNotFound(err) != nil {
					// got a real problem when talking to K8s API.
					multiErrors = multierror.Append(multiErrors, err)
				} else {
					// Object not found
					objectExists = false
				}
			}

			desiredFrontEndSwitchPortVlan := networkNode.Spec.FrontEndFabric.VlanId
			desiredFrontEndSwitchPortMode := networkNode.Spec.FrontEndFabric.Mode
			desiredFrontEndSwitchPortNativeVlan := networkNode.Spec.FrontEndFabric.NativeVlan
			// note:
			// `[]string{}` and `nil` matters here.
			// when a `[]string{}` is given, it means remove all the trunk group for a switch port
			// when `nil`, do nothing.
			var desiredFrontEndSwitchPortTrunkGroups []string
			if networkNode.Spec.FrontEndFabric.TrunkGroups != nil {
				desiredFrontEndSwitchPortTrunkGroups = *networkNode.Spec.FrontEndFabric.TrunkGroups
			}
			if !objectExists {
				// create a new SP
				labels := map[string]string{
					idcnetworkv1alpha1.LabelNameNetworkNode: networkNode.Name,
					idcnetworkv1alpha1.LabelFabricType:      idcnetworkv1alpha1.FabricTypeFrontEnd,
				}
				// vlanID is determined inside the BMH Controller and passed down to the NetworkNode controller.
				err = r.createSwitchPortCR(ctx, frontEndSwitchPortName, desiredFrontEndSwitchPortVlan, desiredFrontEndSwitchPortMode, desiredFrontEndSwitchPortNativeVlan, desiredFrontEndSwitchPortTrunkGroups, labels, networkNode)
				if err != nil {
					logger.Error(err, "failed to create switchPort CR for FrontEndFabric", utils.LogFieldSwitchPortCRName, frontEndSwitchPortName)
					// although we got an error here, we still want to finish all the other changes (eg, accelerator fabric changes)
					multiErrors = multierror.Append(multiErrors, err)
				} else {
					logger.Info(fmt.Sprintf("successfully created SwitchPort CR %v", frontEndSwitchPortName))
				}
			} else {
				err, ready := r.updateSwitchPortCRFromNetworkNode(ctx, networkNode, frontEndSwitchPortCR, desiredFrontEndSwitchPortVlan, desiredFrontEndSwitchPortMode, desiredFrontEndSwitchPortNativeVlan, desiredFrontEndSwitchPortTrunkGroups, idcnetworkv1alpha1.FabricTypeFrontEnd)
				if err != nil {
					multiErrors = multierror.Append(multiErrors, err)
					if alreadyOwnedErr, ok := err.(*AlreadyOwnedError); ok {
						logger.Error(alreadyOwnedErr, "ALERT! ")
						r.EventRecorder.Event(networkNode, corev1.EventTypeWarning, "SwitchPort already owned", alreadyOwnedErr.Error())
						// set the error flag metric for already owned switchport.
						metrics.NetworkNodeControllerErrors.With(prometheus.Labels{
							metrics.MetricsLabelErrorType: metrics.ErrorTypePortAlreadyOwned,
							metrics.MetricsLabelHostName:  networkNode.Name,
						}).Set(1)
					} else {
						// reset the value to 0 when the issue is gone or when it's normal
						metrics.NetworkNodeControllerErrors.With(prometheus.Labels{
							metrics.MetricsLabelErrorType: metrics.ErrorTypePortAlreadyOwned,
							metrics.MetricsLabelHostName:  networkNode.Name,
						}).Set(0)
					}
				}
				if ready {
					frontEndPortReadyCount++
				}

				// update status LastObservedVlanId, taking value from the SP's status' VlanId
			}
			frontEndReadiness := fmt.Sprintf("%d/%d", frontEndPortReadyCount, frontEndPortTotalCount)
			networkNode.Status.FrontEndFabricStatus.LastObservedVlanId = frontEndSwitchPortCR.Status.VlanId
			networkNode.Status.FrontEndFabricStatus.LastObservedMode = frontEndSwitchPortCR.Status.Mode
			networkNode.Status.FrontEndFabricStatus.LastObservedNativeVlan = frontEndSwitchPortCR.Status.NativeVlan
			networkNode.Status.FrontEndFabricStatus.LastObservedTrunkGroups = frontEndSwitchPortCR.Status.TrunkGroups
			networkNode.Status.FrontEndFabricStatus.Readiness = frontEndReadiness
		}
	}()

	// Accelerator Fabric
	func() {
		if networkNode.Spec.AcceleratorFabric != nil {
			// the AcceleratorFabric is empty, no need to process
			if len(networkNode.Spec.AcceleratorFabric.SwitchPorts) == 0 {
				return
			}

			logger.V(1).Info("reconciling NN's Accelerator Fabric")
			accelPortTotalCount = len(networkNode.Spec.AcceleratorFabric.SwitchPorts)
			// all switch ports of the acc fabric should share the same type of vlan, mode, native vlan and trunk groups.
			desiredAccelVlan := networkNode.Spec.AcceleratorFabric.VlanId
			desiredAccelMode := networkNode.Spec.AcceleratorFabric.Mode
			// note: we don't support setting trunk group for acc fabric for now, but may do so in the future.
			// desiredAccelNativeVlan := networkNode.Spec.AcceleratorFabric.NativeVlan
			// desiredAccelTrunkGroups := networkNode.Spec.AcceleratorFabric.TrunkGroups

			switchPortStatus := make([]*idcnetworkv1alpha1.AcceleratorFabricSwitchPortStatus, 0)
			for i := range networkNode.Spec.AcceleratorFabric.SwitchPorts {
				accelSwitchPortName := networkNode.Spec.AcceleratorFabric.SwitchPorts[i]

				key := types.NamespacedName{Name: accelSwitchPortName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
				accelSwitchPortCR := &idcnetworkv1alpha1.SwitchPort{}
				objectExists := true
				err = r.Get(ctx, key, accelSwitchPortCR)
				if err != nil {
					if client.IgnoreNotFound(err) != nil {
						// got a real problem when talking to K8s API.
						multiErrors = multierror.Append(multiErrors, err)
					} else {
						// Object not found
						objectExists = false
					}
				}

				if !objectExists {
					logger.Info("creating a new SP for Accelerator Fabric")
					// create a new SP
					labels := map[string]string{
						idcnetworkv1alpha1.LabelNameNetworkNode: networkNode.Name,
						idcnetworkv1alpha1.LabelFabricType:      idcnetworkv1alpha1.FabricTypeAccelerator,
					}
					// when creating a new SP CR, we get the vlan from the network switch and make it as the SP's Spec.Vlan.
					// It's fine if the switch port's vlan is not the same as the desiredAccelVlan, it will be corrected in the next reconciliation.
					err = r.createSwitchPortCR(ctx, accelSwitchPortName, networkNode.Spec.AcceleratorFabric.VlanId, desiredAccelMode, -1, nil, labels, networkNode)
					if err != nil {
						logger.Error(err, "failed to create switchPort CR for AcceleratorFabric", utils.LogFieldSwitchPortCRName, accelSwitchPortName)
						multiErrors = multierror.Append(multiErrors, err)
					} else {
						logger.Info(fmt.Sprintf("successfully created Accelerator Fabric SwitchPort CR %v", accelSwitchPortName))
					}
				} else {
					err, ready := r.updateSwitchPortCRFromNetworkNode(ctx, networkNode, accelSwitchPortCR, desiredAccelVlan, desiredAccelMode, -1, nil, idcnetworkv1alpha1.FabricTypeAccelerator)
					if err != nil {
						multiErrors = multierror.Append(multiErrors, err)
						if alreadyOwnedErr, ok := err.(*AlreadyOwnedError); ok {
							logger.Error(alreadyOwnedErr, "ALERT! ")
							r.EventRecorder.Event(networkNode, corev1.EventTypeWarning, "SwitchPort already owned", alreadyOwnedErr.Error())
							// set the error flag metric for already owned switchport.
							metrics.NetworkNodeControllerErrors.With(prometheus.Labels{
								metrics.MetricsLabelErrorType: metrics.ErrorTypePortAlreadyOwned,
								metrics.MetricsLabelHostName:  networkNode.Name,
							}).Set(1)
						} else {
							// reset the value to 0 when the issue is gone or when it's normal
							metrics.NetworkNodeControllerErrors.With(prometheus.Labels{
								metrics.MetricsLabelErrorType: metrics.ErrorTypePortAlreadyOwned,
								metrics.MetricsLabelHostName:  networkNode.Name,
							}).Set(0)
						}
					}
					if ready {
						accelPortReadyCount++
					}
				}
				// update status LastObservedVlanId, taking value from the SP's status' VlanId
				switchPortStatus = append(switchPortStatus, &idcnetworkv1alpha1.AcceleratorFabricSwitchPortStatus{
					SwitchPort:              accelSwitchPortName,
					LastObservedVlanId:      accelSwitchPortCR.Status.VlanId,
					LastObservedMode:        accelSwitchPortCR.Status.Mode,
					LastObservedNativeVlan:  accelSwitchPortCR.Status.NativeVlan,
					LastObservedTrunkGroups: accelSwitchPortCR.Status.TrunkGroups,
				})
			}
			accelFabricReadiness := fmt.Sprintf("%d/%d", accelPortReadyCount, accelPortTotalCount)
			networkNode.Status.AcceleratorFabricStatus.Readiness = accelFabricReadiness
			networkNode.Status.AcceleratorFabricStatus.SwitchPorts = switchPortStatus
		}
	}()

	// Storage Fabric
	func() {

		if networkNode.Spec.StorageFabric != nil {
			if len(networkNode.Spec.StorageFabric.SwitchPorts) == 0 {
				return
			}

			logger.V(1).Info("reconciling NN's Storage Fabric")
			storagePortTotalCount = len(networkNode.Spec.StorageFabric.SwitchPorts)
			desiredStorageVlan := networkNode.Spec.StorageFabric.VlanId
			desiredStorageMode := networkNode.Spec.StorageFabric.Mode
			// desiredStorageNativeVlan := networkNode.Spec.StorageFabric.NativeVlan
			// desiredStorageTrunkGroups := networkNode.Spec.StorageFabric.TrunkGroups

			switchPortStatus := make([]*idcnetworkv1alpha1.StorageFabricSwitchPortStatus, 0)
			for i := range networkNode.Spec.StorageFabric.SwitchPorts {
				storageSwitchPortName := networkNode.Spec.StorageFabric.SwitchPorts[i]

				key := types.NamespacedName{Name: storageSwitchPortName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
				storageSwitchPortCR := &idcnetworkv1alpha1.SwitchPort{}
				objectExists := true
				err = r.Get(ctx, key, storageSwitchPortCR)
				if err != nil {
					if client.IgnoreNotFound(err) != nil {
						// got a real problem when talking to K8s API.
						multiErrors = multierror.Append(multiErrors, err)
					} else {
						// Object not found
						objectExists = false
					}
				}

				if !objectExists {
					logger.Info("creating a new SP for Storage Fabric")
					// create a new SP
					labels := map[string]string{
						idcnetworkv1alpha1.LabelNameNetworkNode: networkNode.Name,
						idcnetworkv1alpha1.LabelFabricType:      idcnetworkv1alpha1.FabricTypeStorage,
					}
					// when creating a new SP CR, we get the vlan from the network switch and make it as the SP's Spec.Vlan.
					// It's fine if the switch port's vlan is not the same as the desiredStorageVlan, it will be corrected in the next reconciliation.
					err = r.createSwitchPortCR(ctx, storageSwitchPortName, desiredStorageVlan, desiredStorageMode, -1, nil, labels, networkNode)
					if err != nil {
						logger.Error(err, "failed to create switchPort CR for StorageFabric", utils.LogFieldSwitchPortCRName, storageSwitchPortName)
						multiErrors = multierror.Append(multiErrors, err)
					} else {
						logger.Info(fmt.Sprintf("successfully created Storage Fabric SwitchPort CR %v", storageSwitchPortName))
					}
				} else {
					err, ready := r.updateSwitchPortCRFromNetworkNode(ctx, networkNode, storageSwitchPortCR, desiredStorageVlan, desiredStorageMode, -1, nil, idcnetworkv1alpha1.FabricTypeStorage)
					if err != nil {
						multiErrors = multierror.Append(multiErrors, err)
						if alreadyOwnedErr, ok := err.(*AlreadyOwnedError); ok {
							logger.Error(alreadyOwnedErr, "ALERT! ")
							r.EventRecorder.Event(networkNode, corev1.EventTypeWarning, "SwitchPort already owned", alreadyOwnedErr.Error())
							// set the error flag metric for already owned switchport.
							metrics.NetworkNodeControllerErrors.With(prometheus.Labels{
								metrics.MetricsLabelErrorType: metrics.ErrorTypePortAlreadyOwned,
								metrics.MetricsLabelHostName:  networkNode.Name,
							}).Set(1)
						} else {
							// reset the value to 0 when the issue is gone or when it's normal
							metrics.NetworkNodeControllerErrors.With(prometheus.Labels{
								metrics.MetricsLabelErrorType: metrics.ErrorTypePortAlreadyOwned,
								metrics.MetricsLabelHostName:  networkNode.Name,
							}).Set(0)
						}
					}
					if ready {
						storagePortReadyCount++
					}
				}
				// update status LastObservedVlanId, taking value from the SP's status' VlanId
				switchPortStatus = append(switchPortStatus, &idcnetworkv1alpha1.StorageFabricSwitchPortStatus{
					SwitchPort:         storageSwitchPortName,
					LastObservedVlanId: storageSwitchPortCR.Status.VlanId,
				})
			}
			storageFabricReadiness := fmt.Sprintf("%d/%d", storagePortReadyCount, storagePortTotalCount)
			networkNode.Status.StorageFabricStatus.Readiness = storageFabricReadiness
			networkNode.Status.StorageFabricStatus.SwitchPorts = switchPortStatus
		}
	}()

	if len(multiErrors.Errors) > 0 {
		return ctrl.Result{}, multiErrors
	}

	return ctrl.Result{RequeueAfter: time.Duration(r.Conf.ControllerConfig.NetworkNodeResyncPeriodInSec) * time.Second}, nil
}

// updateSwitchPortCRFromNetworkNode check if the switch port config is in-sync with the NN, and will update the switchPort CR if needed.
func (r *NetworkNodeReconciler) updateSwitchPortCRFromNetworkNode(ctx context.Context, networkNode *idcnetworkv1alpha1.NetworkNode, switchPortCR *idcnetworkv1alpha1.SwitchPort, desiredSwitchPortVlan int64, desiredMode string, desiredNativeVlan int64, nnSpecTrunkGroup []string, fabricType string) (error, bool) {
	logger := log.FromContext(ctx).WithName("updateSwitchPortCRFromNetworkNode").WithValues(utils.LogFieldNetworkNode, networkNode.Name)
	// logger.V(1).Info(fmt.Sprintf("starting to see if %s SwitchPort CR %v needs an update for NetworkNode %v", fabricType, switchPortCR.Name, networkNode.Name))

	// update switchPortCR's spec to meet the desired value
	newSwitchPortCR := switchPortCR.DeepCopy()
	patch := client.MergeFrom(switchPortCR)
	ready := true

	// if there is no owner, set the current networkNode as the owner.
	if len(newSwitchPortCR.OwnerReferences) == 0 {
		trueVar := true
		ownerReference := metav1.OwnerReference{
			APIVersion: networkNode.APIVersion,
			Kind:       networkNode.Kind,
			Name:       networkNode.Name,
			UID:        networkNode.UID,
			Controller: &trueVar,
		}
		newSwitchPortCR.SetOwnerReferences([]metav1.OwnerReference{ownerReference})
	} else if newSwitchPortCR.OwnerReferences[0].UID != networkNode.UID {
		// Another NetworkNode owns this switchPort. Should error, to prevent stealing it from the existing NetworkNode, and TODO: alert for manual intervention.
		err := &AlreadyOwnedError{fabricType: fabricType, switchPortName: switchPortCR.Name, existingOwnerName: switchPortCR.OwnerReferences[0].Name}
		logger.Error(err, err.Error())
		return err, false
	}

	///////////////
	// check Mode
	///////////////
	// note: we will do nothing when the desiredMode is not provided, that means when the provided desiredMode in NN is empty, we will not remove the one in the SwitchPort CR.
	if len(desiredMode) != 0 {
		if desiredMode == switchPortCR.Spec.Mode {
			// switchPortCR Spec's mode field is already correct. next, check if switchPortCR status' mode is in-sync
			if switchPortCR.Spec.Mode == switchPortCR.Status.Mode {
				// switchPortCR status' Vlan is in-sync
			} else {
				ready = false
				logger.V(1).Info("mode update is still in progress")
				// switchPortCR status' Vlan is out-of-sync, SDN is still working on it, will come back and check again on next reconcile.
			}
		} else {
			// update the mode
			newSwitchPortCR.Spec.Mode = desiredMode
			ready = false
			logger.V(1).Info("start updating the mode")
		}
	}

	///////////////
	// check the vlan
	///////////////
	// if the desired port in the NetworkNode assigned with a no-op vlan, ignore the update.
	if desiredMode == idcnetworkv1alpha1.AccessMode && desiredSwitchPortVlan != idcnetworkv1alpha1.NOOPVlanID && desiredSwitchPortVlan != 0 {
		if desiredSwitchPortVlan == switchPortCR.Spec.VlanId {
			// switchPortCR Spec's vlan field is already correct. next, check if switchPortCR status' Vlan is in-sync
			if desiredSwitchPortVlan == switchPortCR.Status.VlanId {
				// switchPortCR status' Vlan is in-sync
			} else {
				logger.V(1).Info("vlan update is still in progress")
				ready = false
				// switchPortCR status' Vlan is out-of-sync, SDN is still working on it, will come back and check again on next reconcile.
			}
		} else {
			// update the vlan id
			newSwitchPortCR.Spec.VlanId = int64(desiredSwitchPortVlan)
			ready = false
			logger.V(1).Info("start updating the vlan")
			// emit event only for FE changes, as we only want to see one event each time node is provisioned.
			if fabricType == idcnetworkv1alpha1.FabricTypeFrontEnd {
				r.EventRecorder.Event(networkNode, corev1.EventTypeNormal, "updating vlan", fmt.Sprintf("updating %s SwitchPort's spec.vlan to %v", fabricType, desiredSwitchPortVlan))
			}
		}
	}

	///////////////
	// check the native vlan
	///////////////
	// if desiredMode == idcnetworkv1alpha1.TrunkMode && desiredNativeVlan != idcnetworkv1alpha1.NOOPValue {
	if desiredNativeVlan != idcnetworkv1alpha1.NOOPVlanID && desiredNativeVlan != 0 {
		if desiredNativeVlan == switchPortCR.Spec.NativeVlan {
			// switchPortCR Spec's vlan field is already correct. next, check if switchPortCR status' Vlan is in-sync
			if switchPortCR.Spec.NativeVlan == switchPortCR.Status.NativeVlan {
				// switchPortCR status' native Vlan is in-sync
			} else {
				ready = false
				logger.V(1).Info("native vlan update is still in progress")
				// switchPortCR status' native Vlan is out-of-sync, SDN is still working on it, will come back and check again on next reconcile.
			}
		} else {
			// update the native vlan id
			newSwitchPortCR.Spec.NativeVlan = desiredNativeVlan
			ready = false
			logger.V(1).Info("start updating the SwitchPort CR's native vlan")
		}
	}

	///////////////
	// check the trunk groups
	///////////////
	if nnSpecTrunkGroup != nil {
		var spSpecTrunkGroup []string
		if switchPortCR.Spec.TrunkGroups != nil {
			spSpecTrunkGroup = *switchPortCR.Spec.TrunkGroups
		}
		// spSpecTrunkGroup := switchPortCR.Spec.TrunkGroups
		spStatusTrunkGroup := switchPortCR.Status.TrunkGroups
		sort.Strings(nnSpecTrunkGroup)
		sort.Strings(spSpecTrunkGroup)
		sort.Strings(spStatusTrunkGroup)
		// if slices.Equal(desiredTrunkGroup, spSpecTrunkGroup) {
		if reflect.DeepEqual(nnSpecTrunkGroup, spSpecTrunkGroup) {
			// switchPortCR Spec's trunk groups field is already correct. next, check if switchPortCR status' trunk groups is in-sync
			if slices.Equal(spSpecTrunkGroup, spStatusTrunkGroup) {
				// switchPortCR status' trunk groups is in-sync
			} else {
				ready = false
				logger.V(1).Info("trunk group update is still in progress")
				// switchPortCR status' trunk groups is out-of-sync, SDN is still working on it, will come back and check again on next reconcile.
			}
		} else {
			newSwitchPortCR.Spec.TrunkGroups = &nnSpecTrunkGroup
			ready = false
			logger.V(1).Info("start updating trunk group")
		}
	} else {
		// if trunk groups field in a NN is nil, then we are not going to maintain the trunk groups, so whatever is set on the switch doesn't matter.
		newSwitchPortCR.Spec.TrunkGroups = nil
	}

	// update the other fields like labels.network_node and labels.fabric_type
	if newSwitchPortCR.Labels == nil {
		newSwitchPortCR.Labels = make(map[string]string)
	}
	newSwitchPortCR.Labels[idcnetworkv1alpha1.LabelNameNetworkNode] = networkNode.Name
	newSwitchPortCR.Labels[idcnetworkv1alpha1.LabelFabricType] = fabricType

	if !reflect.DeepEqual(switchPortCR, newSwitchPortCR) {
		if err := r.Patch(ctx, newSwitchPortCR, patch); err != nil {
			logger.Error(err, "switch port CR patch update failed", utils.LogFieldSwitchPortCRName, newSwitchPortCR.Name)
			return err, false
		} else {
			logger.V(1).Info(fmt.Sprintf("successfully updated %s SwitchPort CR %v, Vlan %v", fabricType, newSwitchPortCR.Name, newSwitchPortCR.Spec.VlanId))
		}
	} else {
		// logger.V(1).Info(fmt.Sprintf("not updating %s SwitchPort CR %v because nothing changed", fabricType, newSwitchPortCR.Name))
	}
	return nil, ready
}

func (r *NetworkNodeReconciler) setStatusReadiness(ctx context.Context, feReady, feTotal, accelReady, accelTotal int, strgReady, strgTotal int, networkNode *idcnetworkv1alpha1.NetworkNode) {
	if networkNode.Spec.FrontEndFabric != nil {
		frontEndReadiness := fmt.Sprintf("%d/%d", feReady, feTotal)
		networkNode.Status.FrontEndFabricStatus.Readiness = frontEndReadiness
	}

	if networkNode.Spec.AcceleratorFabric != nil {
		accelFabricReadiness := fmt.Sprintf("%d/%d", accelReady, accelTotal)
		networkNode.Status.AcceleratorFabricStatus.Readiness = accelFabricReadiness
	}

	if networkNode.Spec.StorageFabric != nil {
		strgFabricReadiness := fmt.Sprintf("%d/%d", strgReady, strgTotal)
		networkNode.Status.StorageFabricStatus.Readiness = strgFabricReadiness
	}
}

func (r *NetworkNodeReconciler) updateStatusAndPersist(ctx context.Context, networkNode *idcnetworkv1alpha1.NetworkNode, reconcileErr error) error {
	logger := log.FromContext(ctx).WithName("updateStatusAndPersist")

	// Persist status.
	if err := r.Status().Update(ctx, networkNode); err != nil {
		logger.Error(err, "update NetworkNode Status failed")
		return fmt.Errorf("update NetworkNode Status failed: %w", err)
	}
	logger.V(1).Info("update NetworkNode status success!")
	return nil
}

// createSwitchPortCR create the SwitchPort CR. It used the vlan ID from the input
func (r *NetworkNodeReconciler) createSwitchPortCR(ctx context.Context, switchPortCRName string, vlanId int64, desiredMode string, desiredNativeVlan int64, desiredTrunkGroups []string, labels map[string]string, networkNode *idcnetworkv1alpha1.NetworkNode) error {
	logger := log.FromContext(ctx).WithName("NetworkNodeReconciler.createSwitchPortCR")

	var err error
	logger.Info("trying to create switchPort CR...", utils.LogFieldSwitchPortCRName, switchPortCRName)

	// TODO: breakdown the CR name into port and switch fqdn
	switchPortName, switchFQDN := utils.PortFullNameToPortNameAndSwitchFQDN(switchPortCRName)

	switchPortCR := utils.NewSwitchPortTemplate(switchFQDN, switchPortName, vlanId, idcnetworkv1alpha1.NOOPPortChannel, labels)
	// set the networkNode as the SP owner
	trueVar := true
	ownerReference := metav1.OwnerReference{
		APIVersion: networkNode.APIVersion,
		Kind:       networkNode.Kind,
		Name:       networkNode.Name,
		UID:        networkNode.UID,
		Controller: &trueVar,
	}
	switchPortCR.SetOwnerReferences([]metav1.OwnerReference{ownerReference})

	switchPortCR.Spec.Mode = desiredMode
	switchPortCR.Spec.NativeVlan = desiredNativeVlan
	switchPortCR.Spec.TrunkGroups = &desiredTrunkGroups

	err = r.Create(ctx, switchPortCR, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create switchPort CR %v failed, reason: %v", switchPortCR.Name, err)
	}
	logger.Info("finished creating SwitchPort CR", "portName", switchPortCRName)
	return nil
}

// createSwitchPortCRWithVlanIDFromSwitch create the SwitchPort CR. It use the current vland ID from the switch.
func (r *NetworkNodeReconciler) createSwitchPortCRWithVlanIDFromSwitch(ctx context.Context, switchPortCRName string, labels map[string]string, networkNode *idcnetworkv1alpha1.NetworkNode) error {
	logger := log.FromContext(ctx).WithName("createSwitchPortCRWithVlanIDFromSwitch")

	var err error
	logger.Info("trying to create switchPort CR...", utils.LogFieldSwitchPortCRName, switchPortCRName)

	switchPortName, switchFQDN := utils.PortFullNameToPortNameAndSwitchFQDN(switchPortCRName)

	// if the switch port is not initialized, then we will create one.
	// for vlan id, we will get it from the switch. Technically the Port's Vlan for a new enrolled BM should be 4008,
	// but getting it from the switch is more accurate and deterministic.
	// logger.Info("trying to get switch client from DeviceManager", utils.LogFieldSwitchFQDN, frontEndSwitchFqdn)
	swClient, err := r.DevicesManager.GetSwitchClient(devicesmanager.GetOption{
		SwitchFQDN: switchFQDN,
	})
	if err != nil {
		logger.Info(fmt.Sprintf("failed to get switch client from DeviceManager for switch %v, please check if the Switch CR has been created and the switch is accessible.", switchFQDN))
		return err
	}
	// fetch ports from the switch.
	portsFromSwitch, err := swClient.GetSwitchPorts(ctx, sc.GetSwitchPortsRequest{
		SwitchFQDN: switchFQDN,
	})
	if err != nil {
		logger.Error(err, "failed to get ports info from the switch", utils.LogFieldSwitchFQDN, switchFQDN)
		return err
	}
	switchPortDetails, found := portsFromSwitch[switchPortName]
	if !found {
		logger.Error(fmt.Errorf("switchPortName not found"), "", utils.LogFieldSwitchPortName, switchPortName)
		return fmt.Errorf("cannot find port %v in the switch", switchPortName)
	}
	vlanId := switchPortDetails.VlanId
	mode := switchPortDetails.Mode
	portChannel := switchPortDetails.PortChannel

	// check if the switch port we are going to create is valid.
	err = utils.ValidatePort(int(vlanId), switchPortName, mode, r.AllowedVlanIds, r.Conf.ControllerConfig.AllowedModes)
	if err != nil {
		logger.Error(err, "SwitchPort is invalid ", utils.LogFieldSwitchPortName, switchPortName)
		return fmt.Errorf("SwitchPort is invalid %v", err)
	}

	// create the switchPort CR
	switchPortCR := utils.NewSwitchPortTemplate(switchFQDN, switchPortName, vlanId, portChannel, labels)
	// set the networkNode as the SP owner
	trueVar := true
	ownerReference := metav1.OwnerReference{
		APIVersion: networkNode.APIVersion,
		Kind:       networkNode.Kind,
		Name:       networkNode.Name,
		UID:        networkNode.UID,
		Controller: &trueVar,
	}
	switchPortCR.SetOwnerReferences([]metav1.OwnerReference{ownerReference})

	err = r.Create(ctx, switchPortCR, &client.CreateOptions{})
	if err != nil {
		logger.Error(err, "create switchPort CR failed", utils.LogFieldSwitchPortCRName, switchPortCR)
		return fmt.Errorf("create switchPort CR %v failed, reason: %v", switchPortCR.Name, err)
	}
	logger.Info("finished creating SwitchPort CR", "portName", switchPortCRName)
	return nil
}

type switchPortEventHandler struct{}

func (e *switchPortEventHandler) Create(ctx context.Context, evt event.CreateEvent, q workqueue.RateLimitingInterface) {
}

func (e *switchPortEventHandler) Update(ctx context.Context, evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	for _, req := range mapSwitchPortToEvent(evt.ObjectNew) {
		q.Add(req)
	}
}

func (e *switchPortEventHandler) Delete(ctx context.Context, evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	for _, req := range mapSwitchPortToEvent(evt.Object) {
		q.Add(req)
	}
}

func (e *switchPortEventHandler) Generic(ctx context.Context, evt event.GenericEvent, q workqueue.RateLimitingInterface) {
}

func mapSwitchPortToEvent(obj client.Object) []reconcile.Request {

	spObj := obj.(*idcnetworkv1alpha1.SwitchPort)
	// use label to find the networkNode
	nn, found := spObj.GetLabels()[idcnetworkv1alpha1.LabelNameNetworkNode]
	if !found {
		return nil
	}

	return []reconcile.Request{
		{NamespacedName: types.NamespacedName{
			Name:      nn,
			Namespace: obj.GetNamespace(),
		}},
	}
}
