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

	"go.opentelemetry.io/otel/codes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/pools"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
)

const (
	NodeGroupUpdateInProgressRequeueTimeInSec = 3
)

// NodeGroupReconciler reconciles a NodeGroup object
type NodeGroupReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder

	PoolManager *pools.PoolManager
	Conf        idcnetworkv1alpha1.SDNControllerConfig
}

//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=nodegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=nodegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=idcnetwork.intel.com,resources=nodegroups/finalizers,verbs=update

func (r *NodeGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("NodeGroupReconciler.Reconcile").WithValues(utils.LogFieldResourceId, req.Name).Start()
	defer span.End()
	result, reconcileErr := func() (ctrl.Result, error) {

		ng := &idcnetworkv1alpha1.NodeGroup{}
		err := r.Get(ctx, req.NamespacedName, ng)
		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("ignoring reconcile request because source CR was not found")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "unable to fetch NodeGroup CR")
			return ctrl.Result{}, err
		}

		// get the pool info for this NodeGroup
		isUnderMaintenance, found := ng.GetLabels()[idcnetworkv1alpha1.LabelMaintenance]
		if found && isUnderMaintenance == idcnetworkv1alpha1.NGMaintenancePhaseInProgress {
			logger.V(1).Info(fmt.Sprintf("the nodeGroup %v is under maintenance, will skip the reconciliation", ng.Name))
			return ctrl.Result{}, nil
		}

		processPortResult, processPortErr := func() (ctrl.Result, error) {
			if ng.ObjectMeta.DeletionTimestamp.IsZero() {
				return r.handleCreateOrUpdate(ctx, ng)
			} else {
				return r.handleDelete(ctx, ng)
			}
		}()

		// update status
		updateStatusErr := r.updateStatusAndPersist(ctx, ng, processPortErr)

		// combine the errors
		processErr := multierror.Append(processPortErr, updateStatusErr)
		if len(processErr.Errors) == 0 {
			return processPortResult, nil
		}

		return processPortResult, processErr

	}()

	if reconcileErr != nil {
		span.SetStatus(codes.Error, reconcileErr.Error())
		logger.Error(reconcileErr, "NodeGroupReconciler.Reconcile: error reconciling NodeGroup")
	}

	return result, reconcileErr
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&idcnetworkv1alpha1.NodeGroup{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Conf.ControllerConfig.MaxConcurrentNodeGroupReconciles,
		}).
		Watches(&idcnetworkv1alpha1.NetworkNode{},
			&networkNodeEventHandler{}).
		Watches(&idcnetworkv1alpha1.Switch{},
			&switchEventHandler{}).
		Complete(r)
}

func (r *NodeGroupReconciler) updateStatusAndPersist(ctx context.Context, ng *idcnetworkv1alpha1.NodeGroup, reconcileErr error) error {
	logger := log.FromContext(ctx).WithName("updateStatusAndPersist")

	ng.Status.NetworkNodesCount = len(ng.Spec.NetworkNodes)
	ng.Status.AccSwitchCount = len(ng.Spec.AcceleratorLeafSwitches)
	ng.Status.FrontEndSwitchCount = len(ng.Spec.FrontEndLeafSwitches)
	ng.Status.StorageSwitchCount = len(ng.Spec.StorageLeafSwitches)
	// Persist status.
	if err := r.Status().Update(ctx, ng); err != nil {
		return fmt.Errorf("update NodeGroup Status failed: %w", err)
	}

	logger.V(1).Info("update NodeGroup status success!")
	return nil
}

func (r *NodeGroupReconciler) updateGroupMemberMeta(ctx context.Context, ng *idcnetworkv1alpha1.NodeGroup) error {
	logger := log.FromContext(ctx).WithName("updateMemberLabels")

	// update nodes' labels
	for i := range ng.Spec.NetworkNodes {
		nodeName := ng.Spec.NetworkNodes[i]
		existingNetworkNodeCR := &idcnetworkv1alpha1.NetworkNode{}
		key := types.NamespacedName{Name: nodeName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err := r.Get(ctx, key, existingNetworkNodeCR)
		if err != nil {
			logger.Error(err, fmt.Sprintf("failed to Get NetworkNode %v", nodeName))
			continue
		}

		if existingNetworkNodeCR.GetLabels() != nil {
			existingGroupID, found := existingNetworkNodeCR.GetLabels()[idcnetworkv1alpha1.LabelGroupID]
			if found {
				if existingGroupID != ng.Name {
					// if a label exists, raise an alert.
					logger.Error(fmt.Errorf("trying to update networkNode %v's GroupID to %v, but found an existing one %v", nodeName, ng.Name, existingGroupID), "")
				}
				continue
			}
		}
		// no Group ID label, create one for the NetworkNode
		networkNodeCRCopy := existingNetworkNodeCR.DeepCopy()
		patch := client.MergeFrom(existingNetworkNodeCR)
		if networkNodeCRCopy.GetLabels() == nil {
			networkNodeCRCopy.Labels = make(map[string]string)
		}
		networkNodeCRCopy.GetLabels()[idcnetworkv1alpha1.LabelGroupID] = ng.Name
		if err := r.Patch(ctx, networkNodeCRCopy, patch); err != nil {
			logger.Error(err, "networkNode CR patch update failed")
		}
	}

	// update switch's labels
	// note: it's not guaranteed that the front-end switches are dedicated for a NodeGroup, so we only update the ACC switches label.
	for i := range ng.Spec.AcceleratorLeafSwitches {
		swName := ng.Spec.AcceleratorLeafSwitches[i]
		existingSwitchCR := &idcnetworkv1alpha1.Switch{}
		key := types.NamespacedName{Name: swName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err := r.Get(ctx, key, existingSwitchCR)
		if err != nil {
			logger.Error(err, fmt.Sprintf("failed to Get Switch CR %v", swName))
			continue
		}

		if existingSwitchCR.GetLabels() != nil {
			existingGroupID, found := existingSwitchCR.GetLabels()[idcnetworkv1alpha1.LabelGroupID]
			if found {
				if existingGroupID != ng.Name {
					// if a label with other GroupID exists, raise an alert.
					logger.Error(fmt.Errorf("trying to update Switch %v's GroupID to %v, but found an existing one %v", swName, ng.Name, existingGroupID), "")
				}
				// continue
			}
		}
		// no Group ID label, create one for the Switch CR
		switchCRCopy := existingSwitchCR.DeepCopy()
		patch := client.MergeFrom(existingSwitchCR)
		if switchCRCopy.GetLabels() == nil {
			switchCRCopy.Labels = make(map[string]string)
		}
		switchCRCopy.GetLabels()[idcnetworkv1alpha1.LabelGroupID] = ng.Name
		if err := r.Patch(ctx, switchCRCopy, patch); err != nil {
			logger.Error(err, "switch CR patch update failed")
		}
	}

	logger.V(1).Info("update NodeGroup member meta success!")
	return nil
}

// TODO: code deduplication
func (r *NodeGroupReconciler) handleCreateOrUpdate(ctx context.Context, ng *idcnetworkv1alpha1.NodeGroup) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("NodeGroupReconciler.handleCreateOrUpdate").WithValues(utils.LogFieldNodeGroup, ng.Name)
	// get the pool name for this NodeGroup
	poolName, found := ng.GetLabels()[idcnetworkv1alpha1.LabelPool]
	if !found || len(poolName) == 0 {
		logger.V(1).Info("the pool label is empty")
		return ctrl.Result{}, nil
	}
	// get the pool details
	pool, err := r.PoolManager.GetPoolByName(ctx, poolName)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get configuration for pool [%v], %v", poolName, err)
	}

	// update group member(nodes, front-end and acc switches) labels
	err = r.updateGroupMemberMeta(ctx, ng)
	if err != nil {
		// we don't want to exit if update meta failed.
		logger.Error(err, "")
	}

	// if the MSU for a pool is NetworkNode, we should not enforce the NodeGroup Spec's desired values to the NetworkNode.
	if pool.SchedulingConfig.MinimumSchedulableUnit == idcnetworkv1alpha1.MSUNetworkNode {
		logger.V(1).Info(fmt.Sprintf("pool MSU is %v, no need to process it at the NodeGroup Controller", idcnetworkv1alpha1.MSUNetworkNode))
		return ctrl.Result{}, nil
	}
	// Note: if the MSU is NetworkNode, all the logic that will cause an update to NetworkNode CR should stop here.

	allAreInSync := true
	/////////////////////////////////
	// handling the front end fabric
	/////////////////////////////////
	if ng.Spec.FrontEndFabricConfig != nil && pool.NetworkConfigStrategy != nil && pool.NetworkConfigStrategy.FrontEndFabricStrategy != nil {
		// make sure the status is not nil
		if ng.Status.FrontEndFabricStatus == nil {
			ng.Status.FrontEndFabricStatus = &idcnetworkv1alpha1.FabricConfigStatus{}
		}
		if ng.Spec.FrontEndFabricConfig.VlanConf != nil {
			// if IsolationType is VLAN, but VlanConf.VlanID is empty, it's ok, the NodeGroup hasn't been assigned with values yet.
			// if ng.Spec.FrontEndFabricConfig.VlanConf == nil {
			// 	// this is ok, when a NodeGroup is initialized, its VlanConf or BGPConf will be empty.
			// 	// logger.Error(fmt.Errorf("FrontEndFabricConfig.VlanConf is empty"), fmt.Sprintf("group %v FrontEndFabric.IsolationType is VLAN, but VlanConf is not provided", ng.Name))
			// 	return ctrl.Result{}, nil
			// }
			// TODO add validation instead of checking zero value
			if ng.Spec.FrontEndFabricConfig.VlanConf.VlanID == 0 {
				logger.Error(fmt.Errorf("vlan ID is 0"), fmt.Sprintf("group %v FrontEndFabric.IsolationType is VLAN, but Vlan ID is 0", ng.Name))
				return ctrl.Result{}, fmt.Errorf("vlan ID is 0")
			}

			// recreate the status for each reconcile loop
			ng.Status.FrontEndFabricStatus.VlanConfigStatus.ReadinessByNetworkNode = make([]idcnetworkv1alpha1.NetworkNodeVlanStatus, 0)

			//////////////////////////////////////////////
			// ensuring the desired FE Vlan Configuration
			//////////////////////////////////////////////
			totalNodesCount := len(ng.Spec.NetworkNodes)
			frontEndNodesReadyCount := 0
			desiredVlan := ng.Spec.FrontEndFabricConfig.VlanConf.VlanID
			// if the vlan ID is a NOOPVlanID, do nothing.
			if desiredVlan != idcnetworkv1alpha1.NOOPVlanID {
				// handle all networkNodes' front-end vlan
				for i := range ng.Spec.NetworkNodes {
					// get the NN
					nodeName := ng.Spec.NetworkNodes[i]

					existingNetworkNodeCR := &idcnetworkv1alpha1.NetworkNode{}
					key := types.NamespacedName{Name: nodeName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
					objectExists := true
					err := r.Get(ctx, key, existingNetworkNodeCR)
					if err != nil {
						if client.IgnoreNotFound(err) != nil {
							// if we have issue fetching a NN, log the error and continue to the next.
							logger.Error(err, fmt.Sprintf("failed to Get NetworkNode %v", nodeName))
							continue
						}
						// Object not found
						objectExists = false
					}

					if !objectExists {
						// if the NN doesn't exist, just log an error (creating the NN is NOT the responsibility of NodeGroup Controller)
						logger.Error(err, fmt.Sprintf("NetworkNode %v doesn't exist", nodeName))
					} else {
						networkNodeCRCopy := existingNetworkNodeCR.DeepCopy()

						if networkNodeCRCopy.Spec.FrontEndFabric == nil {
							networkNodeCRCopy.Spec.FrontEndFabric = &idcnetworkv1alpha1.FrontEndFabric{}
						}

						if networkNodeCRCopy.Spec.FrontEndFabric.VlanId == desiredVlan {
							// if the current NN's vlan is the same as the desired vlan, check if the status field is also in-sync, and update the readiness accordingly.
							if networkNodeCRCopy.Status.FrontEndFabricStatus.LastObservedVlanId == desiredVlan {
								frontEndNodesReadyCount++
							}
						} else {
							// if the current NN's vlan is not the same as the desired vlan, update it.
							patch := client.MergeFrom(existingNetworkNodeCR)
							networkNodeCRCopy.Spec.FrontEndFabric.VlanId = desiredVlan

							if err := r.Patch(ctx, networkNodeCRCopy, patch); err != nil {
								logger.Error(err, "networkNode CR patch update failed")
							} else {
								logger.Info(fmt.Sprintf("successfully update Vlan to %v for networkNode CR %v", desiredVlan, nodeName))
							}
						}

						nnVlanStatus := idcnetworkv1alpha1.NetworkNodeVlanStatus{
							NetworkNodeName: nodeName,
							LastObservedVlanID: map[string]int64{
								networkNodeCRCopy.Spec.FrontEndFabric.SwitchPort: networkNodeCRCopy.Status.FrontEndFabricStatus.LastObservedVlanId,
							},
						}
						ng.Status.FrontEndFabricStatus.VlanConfigStatus.ReadinessByNetworkNode = append(ng.Status.FrontEndFabricStatus.VlanConfigStatus.ReadinessByNetworkNode, nnVlanStatus)
					}
				}
			}

			// update the status fields
			frontEndReadiness := fmt.Sprintf("%d/%d", frontEndNodesReadyCount, totalNodesCount)
			ng.Status.FrontEndFabricStatus.VlanConfigStatus.Readiness = frontEndReadiness

			ready := false
			if frontEndNodesReadyCount == totalNodesCount {
				ready = true
				ng.Status.FrontEndFabricStatus.VlanConfigStatus.LastObservedReadyVLAN = desiredVlan
			}
			ng.Status.FrontEndFabricStatus.VlanConfigStatus.Ready = ready

			if frontEndNodesReadyCount != totalNodesCount {
				allAreInSync = false
			}
		}
		if ng.Spec.FrontEndFabricConfig.BGPConf != nil {
			//////////////////////////////////////////////
			// ensuring the desired FE BGP Configuration
			//////////////////////////////////////////////

			// TODO add validation instead of checking zero value
			if ng.Spec.FrontEndFabricConfig.BGPConf.BGPCommunity == 0 {
				logger.Error(fmt.Errorf("BGPCommunity is 0"), fmt.Sprintf("group %v FrontEndFabric.IsolationType is BGP, but BGPConf is 0", ng.Name))
				return ctrl.Result{}, fmt.Errorf("BGPCommunity is 0")
			}

			// update the status details
			ng.Status.FrontEndFabricStatus.BGPConfigStatus.SwitchBGPConfStatus = make([]idcnetworkv1alpha1.SwitchBGPConfigStatus, 0)
			totalSwitchesCount := len(ng.Spec.FrontEndLeafSwitches)
			frontEndSwitchesReadyCount := 0

			desiredBGPCommunity := ng.Spec.FrontEndFabricConfig.BGPConf.BGPCommunity
			// if the BGP is a NOOPBGPCommunity, do nothing.
			if desiredBGPCommunity != idcnetworkv1alpha1.NOOPBGPCommunity {
				// handle all switches' front-end BGP community
				for i := range ng.Spec.FrontEndLeafSwitches {
					// get the NN
					swFQDN := ng.Spec.FrontEndLeafSwitches[i]

					existingSwitchCR := &idcnetworkv1alpha1.Switch{}
					key := types.NamespacedName{Name: swFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
					objectExists := true
					err := r.Get(ctx, key, existingSwitchCR)
					if err != nil {
						if client.IgnoreNotFound(err) != nil {
							// if we have issue fetching a Switch, log the error and continue to the next.
							logger.Error(err, fmt.Sprintf("failed to Get Switch CR %v", swFQDN))
							continue
						}
						// Object not found
						objectExists = false
					}

					if !objectExists {
						logger.Error(err, fmt.Sprintf("switch CR %v doesn't exist", swFQDN))
					} else {

						switchCRCopy := existingSwitchCR.DeepCopy()
						if switchCRCopy.Spec.BGP == nil {
							logger.Error(err, fmt.Sprintf("switch CR %v doesn't have BGP configured", swFQDN))
							switchCRCopy.Spec.BGP = &idcnetworkv1alpha1.BGPConfig{}
						}

						if switchCRCopy.Spec.BGP != nil && switchCRCopy.Spec.BGP.BGPCommunity == desiredBGPCommunity {
							// if the current switches' BGP community is the same as the desired community , check if the status field is also in-sync, and update the readiness accordingly.
							if switchCRCopy.Status.SwitchBGPConfigStatus != nil && switchCRCopy.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity == desiredBGPCommunity {
								frontEndSwitchesReadyCount++
							}
						} else {
							// if the current Switch's BGP community is not the same as the desired one, update it.
							patch := client.MergeFrom(existingSwitchCR)
							switchCRCopy.Spec.BGP.BGPCommunity = desiredBGPCommunity
							if err := r.Patch(ctx, switchCRCopy, patch); err != nil {
								logger.Error(err, "switch CR patch update failed")
							} else {
								logger.Info(fmt.Sprintf("successfully update BGP community to %v for Switch CR %v", desiredBGPCommunity, swFQDN))
							}
						}

						// update status fields to reflex the latest BGP value
						switchBGPConfigStatus := idcnetworkv1alpha1.SwitchBGPConfigStatus{
							SwitchFQDN: switchCRCopy.Name,
						}
						if switchCRCopy.Status.SwitchBGPConfigStatus != nil {
							switchBGPConfigStatus.LastObservedBGPCommunity = switchCRCopy.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity
						}
						ng.Status.FrontEndFabricStatus.BGPConfigStatus.SwitchBGPConfStatus = append(ng.Status.FrontEndFabricStatus.BGPConfigStatus.SwitchBGPConfStatus, switchBGPConfigStatus)
					}
				}
			}

			// update status fields
			frontEndReadiness := fmt.Sprintf("%d/%d", frontEndSwitchesReadyCount, totalSwitchesCount)
			ng.Status.FrontEndFabricStatus.BGPConfigStatus.Readiness = frontEndReadiness

			ready := false
			if frontEndSwitchesReadyCount == totalSwitchesCount {
				ready = true
				ng.Status.FrontEndFabricStatus.BGPConfigStatus.LastObservedReadyBGP = desiredBGPCommunity
			}
			ng.Status.FrontEndFabricStatus.BGPConfigStatus.Ready = ready

			// update allAreInSync to determine if further reconciliation is needed
			if frontEndSwitchesReadyCount != totalSwitchesCount {
				allAreInSync = false
			}
		}
	}

	/////////////////////////////////
	// handling the accelerator fabric
	/////////////////////////////////
	if ng.Spec.AcceleratorFabricConfig != nil && pool.NetworkConfigStrategy != nil && pool.NetworkConfigStrategy.AcceleratorFabricStrategy != nil {
		// make sure the status is not nil
		if ng.Status.AcceleratorFabricStatus == nil {
			ng.Status.AcceleratorFabricStatus = &idcnetworkv1alpha1.FabricConfigStatus{}
		}
		if ng.Spec.AcceleratorFabricConfig.VlanConf != nil {
			//////////////////////////////////////////////
			// ensuring the desired ACC VLAN Configuration
			//////////////////////////////////////////////

			// update the status details
			ng.Status.AcceleratorFabricStatus.VlanConfigStatus.ReadinessByNetworkNode = make([]idcnetworkv1alpha1.NetworkNodeVlanStatus, 0)

			totalNodesCount := len(ng.Spec.NetworkNodes)
			accelNodesReadyCount := 0
			desiredVlan := ng.Spec.AcceleratorFabricConfig.VlanConf.VlanID
			// if the vlan ID is a NOOPVlanID, do nothing.
			if desiredVlan != idcnetworkv1alpha1.NOOPVlanID {
				// handle all networkNodes' front-end vlan
				for i := range ng.Spec.NetworkNodes {
					// get the NN
					nodeName := ng.Spec.NetworkNodes[i]

					existingNetworkNodeCR := &idcnetworkv1alpha1.NetworkNode{}
					key := types.NamespacedName{Name: nodeName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
					objectExists := true
					err := r.Get(ctx, key, existingNetworkNodeCR)
					if err != nil {
						if client.IgnoreNotFound(err) != nil {
							// if we have issue fetching a NN, log the error and continue to the next.
							logger.Error(err, fmt.Sprintf("failed to Get NetworkNode %v", nodeName))
							continue
						}
						// Object not found
						objectExists = false
					}

					if !objectExists {
						// if the NN doesn't exist, just log an error (creating the NN is NOT the responsibility of NodeGroup Controller)
						logger.Error(err, fmt.Sprintf("NetworkNode %v doesn't exist", nodeName))
					} else {
						networkNodeCRCopy := existingNetworkNodeCR.DeepCopy()
						allACCVlanReady := true

						if networkNodeCRCopy.Spec.AcceleratorFabric == nil {
							networkNodeCRCopy.Spec.AcceleratorFabric = &idcnetworkv1alpha1.AcceleratorFabric{}
						}

						if networkNodeCRCopy.Spec.AcceleratorFabric.VlanId == desiredVlan {
							// if the current NN's vlan is the same as the desired vlan, check if the status field is also in-sync, and update the readiness accordingly.
							for _, sp := range networkNodeCRCopy.Status.AcceleratorFabricStatus.SwitchPorts {
								if sp.LastObservedVlanId != desiredVlan {
									allACCVlanReady = false
									break
								}
							}
							if allACCVlanReady {
								accelNodesReadyCount++
							}
						} else {
							// if the current NN's vlan is not the same as the desired vlan, update it.
							patch := client.MergeFrom(existingNetworkNodeCR)
							networkNodeCRCopy.Spec.AcceleratorFabric.VlanId = desiredVlan

							if err := r.Patch(ctx, networkNodeCRCopy, patch); err != nil {
								logger.Error(err, "networkNode CR patch update failed")
							} else {
								logger.Info(fmt.Sprintf("successfully update Vlan to %v for networkNode CR %v", desiredVlan, nodeName))
							}
						}

						lastObservedACCVlanID := make(map[string]int64)
						for _, accSPInfo := range networkNodeCRCopy.Status.AcceleratorFabricStatus.SwitchPorts {
							lastObservedACCVlanID[accSPInfo.SwitchPort] = accSPInfo.LastObservedVlanId
						}
						nnVlanStatus := idcnetworkv1alpha1.NetworkNodeVlanStatus{
							NetworkNodeName:    nodeName,
							LastObservedVlanID: lastObservedACCVlanID,
						}
						ng.Status.AcceleratorFabricStatus.VlanConfigStatus.ReadinessByNetworkNode = append(ng.Status.AcceleratorFabricStatus.VlanConfigStatus.ReadinessByNetworkNode, nnVlanStatus)
					}
				}
			}
			accelVlanReadiness := fmt.Sprintf("%d/%d", accelNodesReadyCount, totalNodesCount)
			ng.Status.AcceleratorFabricStatus.VlanConfigStatus.Readiness = accelVlanReadiness

			ready := false
			if accelNodesReadyCount == totalNodesCount {
				ready = true
				ng.Status.AcceleratorFabricStatus.VlanConfigStatus.LastObservedReadyVLAN = desiredVlan
			}
			ng.Status.AcceleratorFabricStatus.VlanConfigStatus.Ready = ready

			if accelNodesReadyCount != totalNodesCount {
				allAreInSync = false
			}
		}

		if ng.Spec.AcceleratorFabricConfig.BGPConf != nil {

			// update the status details
			ng.Status.AcceleratorFabricStatus.BGPConfigStatus.SwitchBGPConfStatus = make([]idcnetworkv1alpha1.SwitchBGPConfigStatus, 0)

			totalSwitchesCount := len(ng.Spec.AcceleratorLeafSwitches)
			accelSwitchesReadyCount := 0

			desiredBGPCommunity := ng.Spec.AcceleratorFabricConfig.BGPConf.BGPCommunity
			// if the BGP is a NOOPBGPCommunity, do nothing.
			if desiredBGPCommunity != idcnetworkv1alpha1.NOOPBGPCommunity {
				// handle all switches' ACC BGP community
				for i := range ng.Spec.AcceleratorLeafSwitches {
					// get the NN
					swFQDN := ng.Spec.AcceleratorLeafSwitches[i]

					existingSwitchCR := &idcnetworkv1alpha1.Switch{}
					key := types.NamespacedName{Name: swFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
					objectExists := true
					err := r.Get(ctx, key, existingSwitchCR)
					if err != nil {
						if client.IgnoreNotFound(err) != nil {
							// if we have issue fetching a Switch, log the error and continue to the next.
							logger.Error(err, fmt.Sprintf("failed to Get Switch CR %v", swFQDN))
							continue
						}
						// Object not found
						objectExists = false
					}

					if !objectExists {
						// if the NN doesn't exist, just log an error (creating the NN is NOT the responsibility of NodeGroup Controller)
						logger.Error(err, fmt.Sprintf("switch CR %v doesn't exist", swFQDN))
					} else {
						switchCRCopy := existingSwitchCR.DeepCopy()
						if switchCRCopy.Spec.BGP == nil {
							logger.Error(err, fmt.Sprintf("switch CR %v doesn't have BGP configured", swFQDN))
							switchCRCopy.Spec.BGP = &idcnetworkv1alpha1.BGPConfig{}
						}
						if switchCRCopy.Spec.BGP != nil && switchCRCopy.Spec.BGP.BGPCommunity == desiredBGPCommunity {
							// if the current switches' BGP community is the same as the desired community , check if the status field is also in-sync, and update the readiness accordingly.
							if switchCRCopy.Status.SwitchBGPConfigStatus != nil && switchCRCopy.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity == desiredBGPCommunity {
								accelSwitchesReadyCount++
							}
						} else {
							// if the current Switch's BGP community is not the same as the desired one, update it.
							patch := client.MergeFrom(existingSwitchCR)
							switchCRCopy.Spec.BGP.BGPCommunity = desiredBGPCommunity

							if err := r.Patch(ctx, switchCRCopy, patch); err != nil {
								logger.Error(err, "switch CR patch update failed")
							} else {
								logger.Info(fmt.Sprintf("successfully update BGP community to %v for Switch CR %v", desiredBGPCommunity, swFQDN))
							}
						}

						switchBGPConfigStatus := idcnetworkv1alpha1.SwitchBGPConfigStatus{
							SwitchFQDN: switchCRCopy.Name,
						}
						if switchCRCopy.Status.SwitchBGPConfigStatus != nil {
							switchBGPConfigStatus.LastObservedBGPCommunity = switchCRCopy.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity
						}
						ng.Status.AcceleratorFabricStatus.BGPConfigStatus.SwitchBGPConfStatus = append(ng.Status.AcceleratorFabricStatus.BGPConfigStatus.SwitchBGPConfStatus, switchBGPConfigStatus)
					}
				}
			}

			accReadiness := fmt.Sprintf("%d/%d", accelSwitchesReadyCount, totalSwitchesCount)
			ng.Status.AcceleratorFabricStatus.BGPConfigStatus.Readiness = accReadiness

			ready := false
			if accelSwitchesReadyCount == totalSwitchesCount {
				ready = true
				ng.Status.AcceleratorFabricStatus.BGPConfigStatus.LastObservedReadyBGP = desiredBGPCommunity
			}
			ng.Status.AcceleratorFabricStatus.BGPConfigStatus.Ready = ready

			if accelSwitchesReadyCount != totalSwitchesCount {
				allAreInSync = false
			}
		}
	}

	/////////////////////////////////
	// handling the storage fabric
	/////////////////////////////////
	if ng.Spec.StorageFabricConfig != nil && pool.NetworkConfigStrategy != nil && pool.NetworkConfigStrategy.StorageFabricStrategy != nil {
		// make sure the status is not nil
		if ng.Status.StorageFabricStatus == nil {
			ng.Status.StorageFabricStatus = &idcnetworkv1alpha1.FabricConfigStatus{}
		}
		if ng.Spec.StorageFabricConfig.VlanConf != nil {
			//////////////////////////////////////////////
			// ensuring the desired Storage VLAN Configuration
			//////////////////////////////////////////////

			// update the status details
			ng.Status.StorageFabricStatus.VlanConfigStatus.ReadinessByNetworkNode = make([]idcnetworkv1alpha1.NetworkNodeVlanStatus, 0)

			totalNodesCount := len(ng.Spec.NetworkNodes)
			strgNodesReadyCount := 0
			desiredVlan := ng.Spec.StorageFabricConfig.VlanConf.VlanID
			// if the vlan ID is a NOOPVlanID, do nothing.
			if desiredVlan != idcnetworkv1alpha1.NOOPVlanID {
				// handle all networkNodes' storage vlan
				for i := range ng.Spec.NetworkNodes {
					// get the NN
					nodeName := ng.Spec.NetworkNodes[i]

					existingNetworkNodeCR := &idcnetworkv1alpha1.NetworkNode{}
					key := types.NamespacedName{Name: nodeName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
					objectExists := true
					err := r.Get(ctx, key, existingNetworkNodeCR)
					if err != nil {
						if client.IgnoreNotFound(err) != nil {
							// if we have issue fetching a NN, log the error and continue to the next.
							logger.Error(err, fmt.Sprintf("failed to Get NetworkNode %v", nodeName))
							continue
						}
						// Object not found
						objectExists = false
					}

					if !objectExists {
						// if the NN doesn't exist, just log an error (creating the NN is NOT the responsibility of NodeGroup Controller)
						logger.Error(err, fmt.Sprintf("NetworkNode %v doesn't exist", nodeName))
					} else {
						networkNodeCRCopy := existingNetworkNodeCR.DeepCopy()
						allStrgVlanReady := true

						if networkNodeCRCopy.Spec.StorageFabric == nil {
							networkNodeCRCopy.Spec.StorageFabric = &idcnetworkv1alpha1.StorageFabric{}
						}

						if networkNodeCRCopy.Spec.StorageFabric.VlanId == desiredVlan {
							// if the current NN's vlan is the same as the desired vlan, check if the status field is also in-sync, and update the readiness accordingly.
							for _, sp := range networkNodeCRCopy.Status.StorageFabricStatus.SwitchPorts {
								if sp.LastObservedVlanId != desiredVlan {
									allStrgVlanReady = false
									break
								}
							}
							if allStrgVlanReady {
								strgNodesReadyCount++
							}
						} else {
							// if the current NN's vlan is not the same as the desired vlan, update it.
							patch := client.MergeFrom(existingNetworkNodeCR)
							networkNodeCRCopy.Spec.StorageFabric.VlanId = desiredVlan

							if err := r.Patch(ctx, networkNodeCRCopy, patch); err != nil {
								logger.Error(err, "networkNode CR patch update failed")
							} else {
								logger.Info(fmt.Sprintf("successfully update Vlan to %v for networkNode CR %v", desiredVlan, nodeName))
							}
						}

						lastObservedStrgVlanID := make(map[string]int64)
						for _, strgSPInfo := range networkNodeCRCopy.Status.StorageFabricStatus.SwitchPorts {
							lastObservedStrgVlanID[strgSPInfo.SwitchPort] = strgSPInfo.LastObservedVlanId
						}
						nnVlanStatus := idcnetworkv1alpha1.NetworkNodeVlanStatus{
							NetworkNodeName:    nodeName,
							LastObservedVlanID: lastObservedStrgVlanID,
						}
						ng.Status.StorageFabricStatus.VlanConfigStatus.ReadinessByNetworkNode = append(ng.Status.StorageFabricStatus.VlanConfigStatus.ReadinessByNetworkNode, nnVlanStatus)
					}
				}
			}
			strgVlanReadiness := fmt.Sprintf("%d/%d", strgNodesReadyCount, totalNodesCount)
			ng.Status.StorageFabricStatus.VlanConfigStatus.Readiness = strgVlanReadiness

			ready := false
			if strgNodesReadyCount == totalNodesCount {
				ready = true
				ng.Status.StorageFabricStatus.VlanConfigStatus.LastObservedReadyVLAN = desiredVlan
			}
			ng.Status.StorageFabricStatus.VlanConfigStatus.Ready = ready

			if strgNodesReadyCount != totalNodesCount {
				allAreInSync = false
			}
		}
		if ng.Spec.StorageFabricConfig.BGPConf != nil {

			// update the status details
			ng.Status.StorageFabricStatus.BGPConfigStatus.SwitchBGPConfStatus = make([]idcnetworkv1alpha1.SwitchBGPConfigStatus, 0)

			totalSwitchesCount := len(ng.Spec.StorageLeafSwitches)
			strgSwitchesReadyCount := 0

			desiredBGPCommunity := ng.Spec.StorageFabricConfig.BGPConf.BGPCommunity
			// if the BGP is a NOOPBGPCommunity, do nothing.
			if desiredBGPCommunity != idcnetworkv1alpha1.NOOPBGPCommunity {
				// handle all switches' storage BGP community
				for i := range ng.Spec.StorageLeafSwitches {
					// get the NN
					swFQDN := ng.Spec.StorageLeafSwitches[i]

					existingSwitchCR := &idcnetworkv1alpha1.Switch{}
					key := types.NamespacedName{Name: swFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
					objectExists := true
					err := r.Get(ctx, key, existingSwitchCR)
					if err != nil {
						if client.IgnoreNotFound(err) != nil {
							// if we have issue fetching a Switch, log the error and continue to the next.
							logger.Error(err, fmt.Sprintf("failed to Get Switch CR %v", swFQDN))
							continue
						}
						// Object not found
						objectExists = false
					}

					if !objectExists {
						// if the NN doesn't exist, just log an error (creating the NN is NOT the responsibility of NodeGroup Controller)
						logger.Error(err, fmt.Sprintf("switch CR %v doesn't exist", swFQDN))
					} else {
						switchCRCopy := existingSwitchCR.DeepCopy()
						if switchCRCopy.Spec.BGP == nil {
							logger.Error(err, fmt.Sprintf("switch CR %v doesn't have BGP configured", swFQDN))

							switchCRCopy.Spec.BGP = &idcnetworkv1alpha1.BGPConfig{}
						}
						if switchCRCopy.Spec.BGP.BGPCommunity == desiredBGPCommunity {
							// if the current switches' BGP community is the same as the desired community , check if the status field is also in-sync, and update the readiness accordingly.
							if switchCRCopy.Status.SwitchBGPConfigStatus != nil && switchCRCopy.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity == desiredBGPCommunity {
								strgSwitchesReadyCount++
							}
						} else {
							// if the current Switch's BGP community is not the same as the desired one, update it.
							patch := client.MergeFrom(existingSwitchCR)
							switchCRCopy.Spec.BGP.BGPCommunity = desiredBGPCommunity

							if err := r.Patch(ctx, switchCRCopy, patch); err != nil {
								logger.Error(err, "switch CR patch update failed")
							} else {
								logger.Info(fmt.Sprintf("successfully update BGP community to %v for Switch CR %v", desiredBGPCommunity, swFQDN))
							}
						}

						switchBGPConfigStatus := idcnetworkv1alpha1.SwitchBGPConfigStatus{
							SwitchFQDN: switchCRCopy.Name,
						}
						if switchCRCopy.Status.SwitchBGPConfigStatus != nil {
							switchBGPConfigStatus.LastObservedBGPCommunity = switchCRCopy.Status.SwitchBGPConfigStatus.LastObservedBGPCommunity
						}
						ng.Status.StorageFabricStatus.BGPConfigStatus.SwitchBGPConfStatus = append(ng.Status.StorageFabricStatus.BGPConfigStatus.SwitchBGPConfStatus, switchBGPConfigStatus)
					}
				}
			}

			storageReadiness := fmt.Sprintf("%d/%d", strgSwitchesReadyCount, totalSwitchesCount)
			ng.Status.StorageFabricStatus.BGPConfigStatus.Readiness = storageReadiness

			ready := false
			if strgSwitchesReadyCount == totalSwitchesCount {
				ready = true
				ng.Status.StorageFabricStatus.BGPConfigStatus.LastObservedReadyBGP = desiredBGPCommunity
			}
			ng.Status.StorageFabricStatus.BGPConfigStatus.Ready = ready

			if strgSwitchesReadyCount != totalSwitchesCount {
				allAreInSync = false
			}
		}
	}

	if allAreInSync {
		logger.Info("nodeGroup is in-sync")
		return ctrl.Result{RequeueAfter: time.Duration(r.Conf.ControllerConfig.NodeGroupResyncPeriodInSec) * time.Second}, nil
	}
	// update in progress, wait for a while and check the progress in the next round
	return ctrl.Result{RequeueAfter: time.Duration(NodeGroupUpdateInProgressRequeueTimeInSec) * time.Second}, nil
}

func (r *NodeGroupReconciler) handleDelete(ctx context.Context, sw *idcnetworkv1alpha1.NodeGroup) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("NodeGroupReconciler.handleVBPoolDelete")

	logger.Info("success")
	return ctrl.Result{}, nil
}

type switchEventHandler struct{}

func (e *switchEventHandler) Create(ctx context.Context, evt event.CreateEvent, q workqueue.RateLimitingInterface) {
}

func (e *switchEventHandler) Update(ctx context.Context, evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	for _, req := range mapSwitchToNodeGroupEvent(evt.ObjectNew) {
		q.Add(req)
	}
}

func (e *switchEventHandler) Delete(ctx context.Context, evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	for _, req := range mapSwitchToNodeGroupEvent(evt.Object) {
		q.Add(req)
	}
}

func (e *switchEventHandler) Generic(ctx context.Context, evt event.GenericEvent, q workqueue.RateLimitingInterface) {
}

func mapSwitchToNodeGroupEvent(obj client.Object) []reconcile.Request {
	swObj, ok := obj.(*idcnetworkv1alpha1.Switch)
	if !ok {
		return nil
	}

	// use label to find the networkNode
	ng, found := swObj.GetLabels()[idcnetworkv1alpha1.LabelGroupID]
	if !found {
		return nil
	}

	return []reconcile.Request{
		{NamespacedName: types.NamespacedName{
			Name:      ng,
			Namespace: obj.GetNamespace(),
		}},
	}
}

type networkNodeEventHandler struct{}

func (e *networkNodeEventHandler) Create(ctx context.Context, evt event.CreateEvent, q workqueue.RateLimitingInterface) {
}

func (e *networkNodeEventHandler) Update(ctx context.Context, evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	for _, req := range mapNetworkNodeToNodeGroupEvent(evt.ObjectNew) {
		q.Add(req)
	}
}

func (e *networkNodeEventHandler) Delete(ctx context.Context, evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	for _, req := range mapNetworkNodeToNodeGroupEvent(evt.Object) {
		q.Add(req)
	}
}

func (e *networkNodeEventHandler) Generic(ctx context.Context, evt event.GenericEvent, q workqueue.RateLimitingInterface) {
}

func mapNetworkNodeToNodeGroupEvent(obj client.Object) []reconcile.Request {
	nnObj, ok := obj.(*idcnetworkv1alpha1.NetworkNode)
	if !ok {
		return nil
	}

	// use label to find the networkNode
	ng, found := nnObj.GetLabels()[idcnetworkv1alpha1.LabelGroupID]
	if !found {
		return nil
	}

	return []reconcile.Request{
		{NamespacedName: types.NamespacedName{
			Name:      ng,
			Namespace: obj.GetNamespace(),
		}},
	}
}
