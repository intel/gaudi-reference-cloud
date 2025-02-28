package pools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
)

type PoolManager struct {
	conf                               *idcnetworkv1alpha1.SDNControllerConfig
	groupPoolMappingWatchIntervalInSec int

	NwcpK8sClient     client.WithWatch
	poolMapping       PoolMappingReader
	poolConfig        PoolConfigReader
	nodeUsageReporter NodeUsageReporter
}

type PoolResourceManagerConf struct {
	NwcpK8sClient                      client.WithWatch
	CtrlConf                           *idcnetworkv1alpha1.SDNControllerConfig
	GroupPoolMappingWatchIntervalInSec int
	NodeUsageReporter                  NodeUsageReporter
	Mgr                                manager.Manager
}

func NewPoolManager(pmConf *PoolResourceManagerConf) (*PoolManager, error) {
	var err error
	var mappingReader PoolMappingReader
	mappingReader, err = GetGroupPoolMappingReader(pmConf)
	if err != nil {
		return nil, err
	}

	poolConfig, err := NewLocalPoolConfig(pmConf.CtrlConf)
	if err != nil {
		return nil, err
	}

	return &PoolManager{
		groupPoolMappingWatchIntervalInSec: 10,
		conf:                               pmConf.CtrlConf,
		NwcpK8sClient:                      pmConf.NwcpK8sClient,
		poolMapping:                        mappingReader,
		poolConfig:                         poolConfig,
		nodeUsageReporter:                  pmConf.NodeUsageReporter,
	}, nil
}

func (l *PoolManager) Start(ctx context.Context) error {
	err := l.reconcile(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (l *PoolManager) isNodeGroupAvailable(ctx context.Context, nodeGroup *idcnetworkv1alpha1.NodeGroup) (bool, error) {
	for _, networkNode := range nodeGroup.Spec.NetworkNodes {
		isNodeReserved, err := l.nodeUsageReporter.IsNodeReserved(networkNode)
		if err != nil {
			// if we cannot determine if a node is available, return false and error.
			return false, err
		}
		if isNodeReserved {
			// as long as one of the node is not available, we return false.
			return false, nil
		}
	}
	return true, nil
}

func (l *PoolManager) reconcile(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("PoolManager.reconcile")
	logger.V(1).Info("start")

	// check if the reader provide a watcher, if so, use watch approach, otherwise use Poll approach
	mappingEventCh, err := l.poolMapping.WatchGroupToPoolMappings()
	if err == nil && mappingEventCh != nil {
		err = l.reconcileWithWatch(ctx, mappingEventCh)
		if err != nil {
			return err
		}
	} else {
		err = l.reconcileWithPoll(ctx)
		if err != nil {
			return err
		}
	}
	return err
}

const (
	SendRespTimeOutInSec = 30
)

// reconcileWithWatch watch and process the event from the mapping channel provided by the mapping reader.
func (l *PoolManager) reconcileWithWatch(ctx context.Context, mappingCh chan MappingEvent) error {
	logger := log.FromContext(ctx).WithName("PoolManager.reconcileWithWatch")
	for {
		select {
		case event, ok := <-mappingCh:
			if !ok {
				return fmt.Errorf("mapping channel is closed")
			}

			func() {
				nodeGroupName := event.NodeGroup
				targetPoolName := event.Pool
				var err error
				var processStatus MappingEventProcessStatus = MAPPING_EVENT_PROCESS_UNKNOWN
				defer func() {
					res := MappingHandleResult{}
					res.ResultStatus = processStatus
					if err != nil {
						res.ErrorMessage = err.Error()
					}

					select {
					case event.ResCh <- res:
						logger.V(1).Info(fmt.Sprintf("successfully send back the reconciliation result to the mapping controller. res: [%+v], nodeGroup [%v], targetPoolName: [%v]", res, nodeGroupName, targetPoolName))
					case <-time.After(SendRespTimeOutInSec * time.Second):
						logger.Error(fmt.Errorf("timeout sending process result"), "")
					}
				}()

				// get the nodeGroup CR
				nodeGroup := &idcnetworkv1alpha1.NodeGroup{}
				key := types.NamespacedName{Name: nodeGroupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
				err = l.NwcpK8sClient.Get(ctx, key, nodeGroup)
				if err != nil {
					// do nothing if the NG object doesn't exist
					if apierrors.IsNotFound(err) {
						processStatus = MAPPING_EVENT_PROCESS_NOOP
					} else {
						logger.Error(err, "NwcpK8sClient Get NodeGroup failed")
						processStatus = MAPPING_EVENT_PROCESS_FAILED
					}
					return
				}

				processStatus, err = l.reconcileGroupPoolMapping(ctx, nodeGroup, targetPoolName)
				if err != nil {
					logger.Error(err, fmt.Sprintf("reconcileGroupPoolMapping failed, nodeGroup: %v, targetPool: %v", nodeGroup.Name, targetPoolName))
				}
			}()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// reconcileWithPoll periodically fetch the mappings from the sources.
func (l *PoolManager) reconcileWithPoll(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("PoolManager.watchGroupPoolMapping")
	ticker := time.NewTicker(time.Duration(l.groupPoolMappingWatchIntervalInSec) * time.Second)
	defer ticker.Stop()

	for {
		// check if mappings have been changed
		mapping, err := l.GetGroupToPoolMapping(ctx)
		if err != nil {
			logger.Error(err, "GetGroupToPoolMapping failed")
			// sleep for a while we had issue getting the mappings.
			time.Sleep(time.Second)
			continue
		}
		allNodeGroups := &idcnetworkv1alpha1.NodeGroupList{}
		err = l.NwcpK8sClient.List(ctx, allNodeGroups, &client.ListOptions{})
		if err != nil {
			logger.Error(err, "list NodeGroup failed")
		}
	GROUPS:
		for _, nodeGroup := range allNodeGroups.Items {
			// get the targetPoolName from the mappings
			targetPoolName := mapping[nodeGroup.Name]

			_, err = l.reconcileGroupPoolMapping(ctx, &nodeGroup, targetPoolName)
			if err != nil {
				logger.Error(err, fmt.Sprintf("reconcileGroupPoolMapping failed, nodeGroup: %v, targetPool: %v", nodeGroup.Name, targetPoolName))
				continue GROUPS
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			continue
		}
	}
}

func (l *PoolManager) reconcileGroupPoolMapping(ctx context.Context, nodeGroup *idcnetworkv1alpha1.NodeGroup, targetPoolName string) (MappingEventProcessStatus, error) {
	logger := log.FromContext(ctx).WithName("PoolManager.reconcileGroupPoolMapping")
	var err error

	defer func() {
		logger.V(1).Info(fmt.Sprintf("finished reconcileGroupPoolMapping for nodeGroup [%v], targetPoolName: [%v]", nodeGroup.Name, targetPoolName))
	}()

	// get the currentPoolName
	var currentPoolName string
	var currentMaintenance string
	if nodeGroup.GetLabels() != nil {
		currentPoolName = nodeGroup.GetLabels()[idcnetworkv1alpha1.LabelPool]
		currentMaintenance = nodeGroup.GetLabels()[idcnetworkv1alpha1.LabelMaintenance]
	}

	// return early if no change is needed
	if currentPoolName == targetPoolName && currentMaintenance == "" {
		return MAPPING_EVENT_PROCESS_NOOP, nil
	}

	////////////////////////////////
	// check NodeGroup/NetworkNode availability before the move
	////////////////////////////////
	isNGAvailable, err := l.isNodeGroupAvailable(ctx, nodeGroup)
	if err != nil {
		logger.Error(fmt.Errorf("trying to move NodeGroup %v from pool [%v] to [%v] but failed, failed to check NodeGroup availability, %v", nodeGroup.Name, currentPoolName, targetPoolName, err), "isNodeGroupAvailable failed")
		return MAPPING_EVENT_PROCESS_FAILED, err
	}
	if !isNGAvailable {
		logger.Error(fmt.Errorf("trying to move NodeGroup %v from pool [%v] to [%v] but failed, not all nodes are available", nodeGroup.Name, currentPoolName, targetPoolName), "isNodeGroupAvailable failed")
		return MAPPING_EVENT_PROCESS_FAILED, err
	}

	// if there is no mapping record for this nodeGroup, then remove the pool label from the NG.
	// TODO: when a NG is moved out of a pool, currently only label is removed. What else may needed?
	if len(targetPoolName) == 0 {
		if len(currentPoolName) != 0 {
			err = l.updateNodeGroupLabel(ctx, nodeGroup.Name,
				map[string]string{
					idcnetworkv1alpha1.LabelPool:        "",
					idcnetworkv1alpha1.LabelMaintenance: "",
				})
			if err != nil {
				logger.Error(err, "remove pool label failed")
				return MAPPING_EVENT_PROCESS_FAILED, err
			}
		}
		// targetPoolName is empty
		// if currentPoolName is not empty. we should have reset the labels above. return noop.
		// if currentPoolName is empty, then nothing need to be done. return noop
		logger.V(1).Info(fmt.Sprintf("target pool is empty, finished clearing the pool label for nodeGroup [%v]", nodeGroup.Name))
		return MAPPING_EVENT_PROCESS_NOOP, nil
	}

	// get the targetPool which contains the network strategy definitions.
	targetPool, err := l.GetPoolByName(ctx, targetPoolName)
	if err != nil {
		logger.Error(err, fmt.Sprintf("GetPoolByName failed, pool name: %v", targetPoolName))
		return MAPPING_EVENT_PROCESS_FAILED, err
	}
	if targetPool == nil || targetPool.NetworkConfigStrategy == nil {
		logger.Error(err, fmt.Sprintf("pool or NetworkConfigStrategy definition is nil, pool name: %v", targetPoolName))
		return MAPPING_EVENT_PROCESS_FAILED, err
	}

	if currentPoolName != targetPoolName {
		logger.Info(fmt.Sprintf("starting the process of moving nodeGroup [%v] from [%v] to [%v]", nodeGroup.Name, currentPoolName, targetPoolName))

		////////////////////////////////
		// validate if the NG can fit into the target Pool
		////////////////////////////////
		// there are cases that a NG cannot be moved to a target pool. examples:
		// - the NetworkNodes don't have ACC connections, but the target pool defines
		canMove := l.canNodeGroupFitIntoTargetPool(ctx, nodeGroup, targetPool)
		if !canMove {
			logger.Error(fmt.Errorf("trying to move NodeGroup %v from pool [%v] to [%v] but failed, NodeGroup cannot fit into the target Pool", nodeGroup.Name, currentPoolName, targetPoolName), "canNodeGroupFitIntoTargetPool failed")
			return MAPPING_EVENT_PROCESS_FAILED, fmt.Errorf("trying to move NodeGroup %v from pool [%v] to [%v] but failed, NodeGroup cannot fit into the target Pool", nodeGroup.Name, currentPoolName, targetPoolName)
		}

		////////////////////////////////
		// lock the NG before any further update.
		////////////////////////////////
		err = l.updateNodeGroupLabel(ctx, nodeGroup.Name,
			map[string]string{
				idcnetworkv1alpha1.LabelMaintenance: idcnetworkv1alpha1.NGMaintenancePhaseInProgress, // make this in maintenance mode, so the NG controller will not work on it.
			})
		if err != nil {
			logger.Error(err, "updateNodeGroupLabel failed")
			return MAPPING_EVENT_PROCESS_FAILED, err
		}
		logger.Info(fmt.Sprintf("nodeGroup %v is locked for maintenance", nodeGroup.Name))

		////////////////////////////////
		// update the NetworkNodes
		////////////////////////////////
		var updateNodeErrs *multierror.Error
		for _, nodeName := range nodeGroup.Spec.NetworkNodes {
			updateNodeErr := l.updateNetworkNode(ctx, nodeName, targetPoolName)
			if updateNodeErr != nil {
				logger.Error(updateNodeErr, "updateNetworkNode failed")
				updateNodeErrs = multierror.Append(updateNodeErr)
				continue
			}
			logger.Info(fmt.Sprintf("finished updating NetworkNode %v for moving from Pool %v to %v", nodeName, currentPoolName, targetPoolName))
		}

		////////////////////////////////
		// update the Switches
		////////////////////////////////
		var updateSwitchesErrs *multierror.Error
		for _, feSwitch := range nodeGroup.Spec.FrontEndLeafSwitches {
			updateSwitchesErr := l.updateSwitch(ctx, targetPool.NetworkConfigStrategy.FrontEndFabricStrategy, feSwitch)
			if updateSwitchesErr != nil {
				logger.Error(updateSwitchesErr, "update frontend Switch failed")
				updateSwitchesErrs = multierror.Append(updateSwitchesErr)
				continue
			}
			logger.Info(fmt.Sprintf("finished updating Switch %v for moving from Pool %v to %v", feSwitch, currentPoolName, targetPoolName))
		}

		for _, accSwitch := range nodeGroup.Spec.AcceleratorLeafSwitches {
			updateSwitchesErr := l.updateSwitch(ctx, targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy, accSwitch)
			if updateSwitchesErr != nil {
				logger.Error(updateSwitchesErr, "update ACC Switch failed")
				updateSwitchesErrs = multierror.Append(updateSwitchesErr)
				continue
			}
			logger.Info(fmt.Sprintf("finished updating Switch %v for moving from Pool %v to %v", accSwitch, currentPoolName, targetPoolName))
		}

		for _, strgSwitch := range nodeGroup.Spec.StorageLeafSwitches {
			updateSwitchesErr := l.updateSwitch(ctx, targetPool.NetworkConfigStrategy.StorageFabricStrategy, strgSwitch)
			if updateSwitchesErr != nil {
				logger.Error(updateSwitchesErr, "update STRG Switch failed")
				updateSwitchesErrs = multierror.Append(updateSwitchesErr)
				continue
			}
			logger.Info(fmt.Sprintf("finished updating Switch %v for moving from Pool %v to %v", strgSwitch, currentPoolName, targetPoolName))
		}

		// move to the next NG if update failed
		if updateNodeErrs != nil && len(updateNodeErrs.Errors) > 0 {
			logger.Error(updateNodeErrs, "updateNetworkNode finished with errors")
			return MAPPING_EVENT_PROCESS_FAILED, updateNodeErrs
		}

		if updateSwitchesErrs != nil && len(updateSwitchesErrs.Errors) > 0 {
			logger.Error(updateSwitchesErrs, "updateSwitch finished with errors")
			return MAPPING_EVENT_PROCESS_FAILED, updateSwitchesErrs
		}

		////////////////////////////////
		// update the nodeGroup
		////////////////////////////////
		updateNGErr := l.updateNodeGroup(ctx, nodeGroup.Name, targetPoolName)
		if updateNGErr != nil {
			logger.Error(updateNGErr, "updateNodeGroup failed")
			return MAPPING_EVENT_PROCESS_FAILED, updateNGErr
		}
		logger.Info(fmt.Sprintf("finished updating NodeGroup %v Spec and Status (labels not updated yet) for moving from Pool %v to %v", nodeGroup.Name, currentPoolName, targetPoolName))

		////////////////////////////////
		// update the pool label to WaitingForReady
		////////////////////////////////
		// we have updated the NN, SW and NG CRD above, but the controllers need time to executed the update. So we mark the label as NGMaintenancePhaseWaitingForReady,
		// and come back for the next round to check the readiness.
		err = l.updateNodeGroupLabel(ctx, nodeGroup.Name,
			map[string]string{
				idcnetworkv1alpha1.LabelMaintenance: idcnetworkv1alpha1.NGMaintenancePhaseWaitingForReady, // set maintenance phase to waiting for ready
				idcnetworkv1alpha1.LabelPool:        targetPool.Name,
			})
		if err != nil {
			logger.Error(err, "update LabelMaintenance & LabelPool failed")
			return MAPPING_EVENT_PROCESS_FAILED, err
		}
		logger.Info(fmt.Sprintf("moving NodeGroup %v from Pool %v to %v, maintenance label set to waitingForReady", nodeGroup.Name, currentPoolName, targetPoolName))

		return MAPPING_EVENT_PROCESS_IN_PROGRESS, nil
	} else {
		// if currentPoolName == targetPoolName, and the NG matches the default state of the target pool, cleanup the maintenance label.
		var maintenanceLabelVal string
		if nodeGroup.Labels != nil {
			maintenanceLabelVal = nodeGroup.Labels[idcnetworkv1alpha1.LabelMaintenance]
		}
		if len(maintenanceLabelVal) != 0 {
			// check if the resources match the default state of the target pool
			moveToTargetPoolDone := l.isNodeGroupInDefaultState(ctx, nodeGroup, targetPool)
			if !moveToTargetPoolDone {
				logger.Info(fmt.Sprintf("moving NodeGroup %v to %v is in progress, will check it again later...", nodeGroup.Name, targetPoolName))
				return MAPPING_EVENT_PROCESS_IN_PROGRESS, nil
			}

			// status is insync, we can mark this NG as ready
			err = l.updateNodeGroupLabel(ctx, nodeGroup.Name,
				map[string]string{idcnetworkv1alpha1.LabelMaintenance: ""})
			if err != nil {
				logger.Error(err, "update label LabelReady failed")
				return MAPPING_EVENT_PROCESS_FAILED, err
			}

			// if currentPoolName == targetPoolName, and the maintenance label was not empty, but the state match the desired.
			// we assume a pool adjustment happened previously.
			logger.Info(fmt.Sprintf("finished moving NodeGroup %v to %v.", nodeGroup.Name, targetPoolName))
			return MAPPING_EVENT_PROCESS_SUCCESS, nil
		} else {
			// if currentPoolName == targetPoolName, and the maintenance label is empty, the this is a noop.
			return MAPPING_EVENT_PROCESS_NOOP, nil
		}
	}
}

func (l *PoolManager) updateNodeGroupLabel(ctx context.Context, nodeGroupName string, labels map[string]string) error {
	latestNodeGroup := &idcnetworkv1alpha1.NodeGroup{}
	key := types.NamespacedName{Name: nodeGroupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := l.NwcpK8sClient.Get(ctx, key, latestNodeGroup)
	if err != nil {
		return fmt.Errorf("NwcpK8sClient Get NodeGroup failed, %v", err)
	}
	latestNodeGroupForLabelUpdate := latestNodeGroup.DeepCopy()
	if latestNodeGroupForLabelUpdate.Labels == nil {
		latestNodeGroupForLabelUpdate.Labels = make(map[string]string)
	}
	for k, v := range labels {
		latestNodeGroupForLabelUpdate.Labels[k] = v
	}
	patch := client.MergeFrom(latestNodeGroup)
	err = l.NwcpK8sClient.Patch(ctx, latestNodeGroupForLabelUpdate, patch)
	if err != nil {
		return fmt.Errorf("NwcpK8sClient Patch update label failed, %v", err)
	}
	return nil
}

func (l *PoolManager) setDefaultVlanForNodeGroup(ctx context.Context, fabricConfig *idcnetworkv1alpha1.FabricConfig, msu idcnetworkv1alpha1.MSU, defaultVlan int64) {
	if fabricConfig.VlanConf == nil {
		fabricConfig.VlanConf = &idcnetworkv1alpha1.VlanConfig{}
	}
	if msu == idcnetworkv1alpha1.MSUNetworkNode {
		// for MSU is networkNode, we don't do anything at the NodeGroup level, so will leave the NodeGroup's vlan value as empty.
	} else if msu == idcnetworkv1alpha1.MSUNodeGroup {
		// by default set vlan to NOOP
		fabricConfig.VlanConf.VlanID = idcnetworkv1alpha1.NOOPVlanID
		// set it to the Pool defined default value if it's enabled
		if l.conf.ControllerConfig.UseDefaultValueInPoolForMovingNodeGroup {
			fabricConfig.VlanConf.VlanID = defaultVlan
		}
	}
}

func (l *PoolManager) setDefaultBGPForNodeGroup(ctx context.Context, fabricConfig *idcnetworkv1alpha1.FabricConfig, msu idcnetworkv1alpha1.MSU, defaultBGPCmty int64) {
	if fabricConfig.BGPConf == nil {
		fabricConfig.BGPConf = &idcnetworkv1alpha1.BGPConfig{}
	}
	if msu == idcnetworkv1alpha1.MSUNetworkNode {
	} else if msu == idcnetworkv1alpha1.MSUNodeGroup {
		fabricConfig.BGPConf.BGPCommunity = idcnetworkv1alpha1.NOOPBGPCommunity
		if l.conf.ControllerConfig.UseDefaultValueInPoolForMovingNodeGroup {
			fabricConfig.BGPConf.BGPCommunity = defaultBGPCmty
		}
	}
}

// GetPoolDetails accepts a Pool name and return the details of Pool configuration.
func (l *PoolManager) GetPoolByName(ctx context.Context, poolName string) (*idcnetworkv1alpha1.Pool, error) {
	return l.poolConfig.GetPoolConfigByName(ctx, poolName)
}

// GetPoolDetails accepts a Pool name and return the details of Pool configuration.
func (l *PoolManager) GetGroupToPoolMapping(ctx context.Context) (map[string]string, error) {
	return l.poolMapping.GetGroupToPoolMappings(ctx)
}

// GetPoolNameByGroupName returns the Pool that a NodeGroup belongs to. This implementation read it from the local file system.
// if nil, nil is returned, that means the NodeGroup to Pool mapping is not found.
func (l *PoolManager) GetPoolByGroupName(ctx context.Context, groupName string) (*idcnetworkv1alpha1.Pool, error) {
	poolName, err := l.poolMapping.GetPoolByGroupName(ctx, groupName)
	if err != nil {
		return nil, err
	}
	// there is no Group to Pool mapping configured
	if len(poolName) == 0 {
		return nil, nil
	}

	pool, err := l.poolConfig.GetPoolConfigByName(ctx, poolName)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (l *PoolManager) removeBGPSettingForSwitches(ctx context.Context, switches []string) error {
	for _, sw := range switches {
		existingSwitch := &idcnetworkv1alpha1.Switch{}
		key := types.NamespacedName{Name: sw, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err := l.NwcpK8sClient.Get(ctx, key, existingSwitch)
		if err != nil {
			return err
		}
		patch := client.MergeFromWithOptions(existingSwitch, client.MergeFromWithOptimisticLock{})

		newSwitch := existingSwitch.DeepCopy()
		newSwitch.Spec.BGP = nil
		if err := l.NwcpK8sClient.Patch(ctx, newSwitch, patch); err != nil {
			return err
		}

		newSwitchForStatus := existingSwitch.DeepCopy()
		newSwitchForStatus.Status.SwitchBGPConfigStatus = nil
		err = l.NwcpK8sClient.Status().Patch(ctx, newSwitchForStatus, patch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *PoolManager) resetVlanSettingForNetworkNodes(ctx context.Context, nodes []string, fabricType string, vlan int64) error {

	for _, nn := range nodes {
		existingNetworkNode := &idcnetworkv1alpha1.NetworkNode{}
		key := types.NamespacedName{Name: nn, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err := l.NwcpK8sClient.Get(ctx, key, existingNetworkNode)
		if err != nil {
			return err
		}
		newNetworkNode := existingNetworkNode.DeepCopy()
		newNetworkNodeForStatus := existingNetworkNode.DeepCopy()
		patch := client.MergeFromWithOptions(existingNetworkNode, client.MergeFromWithOptimisticLock{})
		if fabricType == idcnetworkv1alpha1.FabricTypeFrontEnd {
			newNetworkNode.Spec.FrontEndFabric.VlanId = vlan
			// newNetworkNode.Spec.FrontEndFabric.VlanId = idcnetworkv1alpha1.NOOPVlanID
			newNetworkNodeForStatus.Status.FrontEndFabricStatus = idcnetworkv1alpha1.FrontEndFabricStatus{}
		} else if fabricType == idcnetworkv1alpha1.FabricTypeAccelerator {
			newNetworkNode.Spec.AcceleratorFabric.VlanId = vlan
			// newNetworkNode.Spec.AcceleratorFabric.VlanId = idcnetworkv1alpha1.NOOPVlanID
			newNetworkNodeForStatus.Status.AcceleratorFabricStatus = idcnetworkv1alpha1.AcceleratorFabricStatus{}
		} else if fabricType == idcnetworkv1alpha1.FabricTypeStorage {
			newNetworkNode.Spec.StorageFabric.VlanId = vlan
			// newNetworkNode.Spec.StorageFabric.VlanId = idcnetworkv1alpha1.NOOPVlanID
			newNetworkNodeForStatus.Status.StorageFabricStatus = idcnetworkv1alpha1.StorageFabricStatus{}
		}

		err = l.NwcpK8sClient.Patch(ctx, newNetworkNode, patch)
		if err != nil {
			return err
		}

		err = l.NwcpK8sClient.Status().Patch(ctx, newNetworkNodeForStatus, patch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *PoolManager) updateNetworkNodeWithDefaultVlan(ctx context.Context, nnName string, targetPool *idcnetworkv1alpha1.Pool) error {
	logger := log.FromContext(ctx).WithName("PoolManager.updateNetworkNodeWithDefaultVlan").WithValues(utils.LogFieldNetworkNode, nnName)

	if targetPool.NetworkConfigStrategy == nil {
		return fmt.Errorf("targetPool.NetworkConfigStrategy is nil")
	}

	existingNetworkNode := &idcnetworkv1alpha1.NetworkNode{}
	key := types.NamespacedName{Name: nnName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := l.NwcpK8sClient.Get(ctx, key, existingNetworkNode)
	if err != nil {
		return err
	}
	newNetwokNode := existingNetworkNode.DeepCopy()
	newNetwokNodeForStatus := existingNetworkNode.DeepCopy()

	if targetPool.NetworkConfigStrategy.FrontEndFabricStrategy != nil {
		if newNetwokNode.Spec.FrontEndFabric == nil {
			newNetwokNode.Spec.FrontEndFabric = &idcnetworkv1alpha1.FrontEndFabric{}
		}

		// NetworkNode is only for Vlan changes
		if targetPool.NetworkConfigStrategy.FrontEndFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeVLAN {
			if l.conf.ControllerConfig.UseDefaultValueInPoolForMovingNodeGroup {
				logger.V(1).Info(fmt.Sprintf("resetting Spec.FrontEndFabric.VlanId to %v", targetPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID))
				newNetwokNode.Spec.FrontEndFabric.VlanId = *targetPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID
			}
		} else if targetPool.NetworkConfigStrategy.FrontEndFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeBGP {
			// if the frontend fabric is using BGP, clear the vlan status,
			newNetwokNode.Spec.FrontEndFabric.VlanId = 0
			newNetwokNodeForStatus.Status.FrontEndFabricStatus = idcnetworkv1alpha1.FrontEndFabricStatus{}
		}
	}

	if targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy != nil {
		if newNetwokNode.Spec.AcceleratorFabric == nil {
			newNetwokNode.Spec.AcceleratorFabric = &idcnetworkv1alpha1.AcceleratorFabric{}
		}

		if targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeVLAN {
			if l.conf.ControllerConfig.UseDefaultValueInPoolForMovingNodeGroup {
				if l.conf.ControllerConfig.UseDefaultValueInPoolForMovingNodeGroup {
					logger.V(1).Info(fmt.Sprintf("resetting Spec.AcceleratorFabric.VlanId to %v", targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID))
					newNetwokNode.Spec.AcceleratorFabric.VlanId = *targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID
				}
			} else if targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeBGP {
				// TODO:  if the acc fabric is using BGP, clear the vlan status,
				newNetwokNode.Spec.AcceleratorFabric.VlanId = 0
				newNetwokNodeForStatus.Status.AcceleratorFabricStatus = idcnetworkv1alpha1.AcceleratorFabricStatus{}
			}
		}
	}

	if targetPool.NetworkConfigStrategy.StorageFabricStrategy != nil {
		if newNetwokNode.Spec.StorageFabric == nil {
			newNetwokNode.Spec.StorageFabric = &idcnetworkv1alpha1.StorageFabric{}
		}

		if targetPool.NetworkConfigStrategy.StorageFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeVLAN {
			if l.conf.ControllerConfig.UseDefaultValueInPoolForMovingNodeGroup {
				logger.V(1).Info(fmt.Sprintf("resetting Spec.StorageFabricStrategy.VlanId to %v", targetPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID))
				newNetwokNode.Spec.StorageFabric.VlanId = *targetPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID
			}
		} else if targetPool.NetworkConfigStrategy.StorageFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeBGP {
			// if the storage fabric is using BGP, clear the vlan status,
			newNetwokNode.Spec.StorageFabric.VlanId = 0
			newNetwokNodeForStatus.Status.StorageFabricStatus = idcnetworkv1alpha1.StorageFabricStatus{}
		}
	}

	err = l.NwcpK8sClient.Status().Update(ctx, newNetwokNodeForStatus)
	if err != nil {
		return err
	}

	patch := client.MergeFrom(existingNetworkNode)
	err = l.NwcpK8sClient.Patch(ctx, newNetwokNode, patch)
	if err != nil {
		return err
	}
	logger.Info("successfully reset Vlan values for NetworkNode!!")
	return nil
}

func (l *PoolManager) updateSwitchMeta(ctx context.Context, switchFQDN string, groupName string, fabricType string) error {
	existingSwitch := &idcnetworkv1alpha1.Switch{}
	key := types.NamespacedName{Name: switchFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := l.NwcpK8sClient.Get(ctx, key, existingSwitch)
	if err != nil {
		return err
	}
	switchCopy := existingSwitch.DeepCopy()
	if switchCopy.Labels == nil {
		switchCopy.Labels = make(map[string]string)
	}
	switchCopy.Labels[idcnetworkv1alpha1.LabelFabricType] = fabricType
	if fabricType == idcnetworkv1alpha1.FabricTypeAccelerator {
		// only ACC switches are dedicated to a group
		switchCopy.Labels[idcnetworkv1alpha1.LabelGroupID] = groupName
	}
	patch := client.MergeFrom(existingSwitch)
	err = l.NwcpK8sClient.Patch(ctx, switchCopy, patch)
	if err != nil {
		return err
	}
	return nil
}

func (l *PoolManager) isNodeGroupVlanStatusReady(ctx context.Context, fabricStatus *idcnetworkv1alpha1.FabricConfigStatus, networkConfigStrategy *idcnetworkv1alpha1.NetworkStrategy) bool {
	// check nodeGroup vlan status is in ready state
	if fabricStatus == nil ||
		!fabricStatus.VlanConfigStatus.Ready ||
		fabricStatus.VlanConfigStatus.LastObservedReadyVLAN != *networkConfigStrategy.ProvisionConfig.DefaultVlanID {
		return false
	}
	return true
}

func (l *PoolManager) isNodeGroupBGPStatusReady(ctx context.Context, fabricStatus *idcnetworkv1alpha1.FabricConfigStatus, networkConfigStrategy *idcnetworkv1alpha1.NetworkStrategy) bool {
	// check nodeGroup BGP status is in ready state
	if fabricStatus == nil ||
		!fabricStatus.BGPConfigStatus.Ready ||
		fabricStatus.BGPConfigStatus.LastObservedReadyBGP != *networkConfigStrategy.ProvisionConfig.DefaultBGPCommunity {
		return false
	}
	return true
}

func (l *PoolManager) canNodeGroupFitIntoTargetPool(ctx context.Context, nodeGroup *idcnetworkv1alpha1.NodeGroup, targetPool *idcnetworkv1alpha1.Pool) bool {
	logger := log.FromContext(ctx).WithName("PoolManager.canNodeGroupFitIntoTargetPool")
	canMove := true
	if targetPool.NetworkConfigStrategy != nil && targetPool.NetworkConfigStrategy.FrontEndFabricStrategy != nil &&
		targetPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID != nil {
		if len(nodeGroup.Spec.FrontEndLeafSwitches) == 0 {
			logger.Info(fmt.Sprintf("cannot move NG %v to pool [%v], NG has no frontend fabric switch", nodeGroup.Name, targetPool.Name))
			canMove = false
		}
	}
	if targetPool.NetworkConfigStrategy != nil && targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy != nil &&
		targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID != nil {
		if len(nodeGroup.Spec.AcceleratorLeafSwitches) == 0 {
			logger.Info(fmt.Sprintf("cannot move NG %v to pool [%v], NG has no accelerator fabric switch", nodeGroup.Name, targetPool.Name))
			canMove = false
		}
	}
	if targetPool.NetworkConfigStrategy != nil && targetPool.NetworkConfigStrategy.StorageFabricStrategy != nil &&
		targetPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID != nil {
		if len(nodeGroup.Spec.StorageLeafSwitches) == 0 {
			logger.Info(fmt.Sprintf("cannot move NG %v to pool [%v], NG has no storage fabric switch", nodeGroup.Name, targetPool.Name))
			canMove = false
		}
	}

	return canMove
}

// isNodeGroupInDefaultState check if the provided nodeGroup's current network config is the same as the defined default values.
// Node: this function won't validate the configs that are NOT defined in the current Pool config. For example, if this is a VV Pool. it won't check if BGP value is correct or not.
func (l *PoolManager) isNodeGroupInDefaultState(ctx context.Context, nodeGroup *idcnetworkv1alpha1.NodeGroup, currentPool *idcnetworkv1alpha1.Pool) bool {
	logger := log.FromContext(ctx).WithName("PoolManager.isNodeGroupInDefaultState")
	_ = logger
	if currentPool.NetworkConfigStrategy == nil {
		logger.Error(fmt.Errorf("currentPool.NetworkConfigStrategy is nil"), "")
		return false
	}

	// TODO: here we set the default value as true, and check all the false conditions. The down side is it requires more checks to cover all the false conditions,
	// but know what conditions are causing the failure.
	// consider set this to false by default, and mark it as true if conditions are met?
	validationResult := true
	// for MSU is NodeGroup, check at the NodeGroup level
	if currentPool.SchedulingConfig.MinimumSchedulableUnit == idcnetworkv1alpha1.MSUNodeGroup {
		if currentPool.NetworkConfigStrategy.FrontEndFabricStrategy != nil {
			if nodeGroup.Spec.FrontEndFabricConfig == nil {
				logger.Info("currentPool.NetworkConfigStrategy.FrontEndFabricStrategy is NOT nil, but nodeGroup.Spec.FrontEndFabricConfig is nil")
				validationResult = false
			} else {
				// check front end vlan config if it's defined in the Pool
				if currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID != nil {

					if nodeGroup.Spec.FrontEndFabricConfig.VlanConf == nil {
						logger.Info("currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID is NOT nil, but nodeGroup.Spec.FrontEndFabricConfig.VlanConf is nil")
						validationResult = false
					} else {
						// if spec's vlan is set to -1, no need to perform validation
						if nodeGroup.Spec.FrontEndFabricConfig.VlanConf.VlanID != idcnetworkv1alpha1.NOOPVlanID {
							if nodeGroup.Spec.FrontEndFabricConfig.VlanConf.VlanID != *currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID {
								logger.Info(fmt.Sprintf("nodeGroup %v frontend Vlan need to set to default value %v before moving to other Pool", nodeGroup.Name, *currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID))
								validationResult = false
							}

							// check nodeGroup vlan status is in ready state. When spec's vlan is -1 or 0, then status most likely is nil, so no need to check that.
							if nodeGroup.Status.FrontEndFabricStatus == nil {
								logger.Info("nodeGroup.Status.FrontEndFabricStatus is nil")
								validationResult = false
							} else {
								if !l.isNodeGroupVlanStatusReady(ctx, nodeGroup.Status.FrontEndFabricStatus, currentPool.NetworkConfigStrategy.FrontEndFabricStrategy) {
									logger.Info(fmt.Sprintf("nodeGroup %v frontend Vlan status is not ready. want: %v, current: %v ", nodeGroup.Name, *currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID, nodeGroup.Status.FrontEndFabricStatus.VlanConfigStatus.LastObservedReadyVLAN))
									validationResult = false
								}
							}
						}
					}
				}

				// check front end BGP config if it's defined in the Pool
				if currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultBGPCommunity != nil {
					if nodeGroup.Spec.FrontEndFabricConfig.BGPConf == nil {
						logger.Info("currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultBGPCommunity is NOT nil, but nodeGroup.Spec.FrontEndFabricConfig.BGPConf is nil")
						validationResult = false
					} else {

						if nodeGroup.Spec.FrontEndFabricConfig.BGPConf.BGPCommunity != idcnetworkv1alpha1.NOOPBGPCommunity {

							if nodeGroup.Spec.FrontEndFabricConfig.BGPConf.BGPCommunity != *currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultBGPCommunity {
								logger.Info(fmt.Sprintf("nodeGroup %v frontend BGP need to set to default value %v before moving to other Pool", nodeGroup.Name, *currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultBGPCommunity))
								validationResult = false
							}

							// check nodeGroup BGP status is in ready state
							if nodeGroup.Status.FrontEndFabricStatus == nil {
								logger.Info("nodeGroup.Status.FrontEndFabricStatus is nil")
								validationResult = false
							} else {
								if !l.isNodeGroupBGPStatusReady(ctx, nodeGroup.Status.FrontEndFabricStatus, currentPool.NetworkConfigStrategy.FrontEndFabricStrategy) {
									logger.Info(fmt.Sprintf("nodeGroup %v frontend BGP status is not ready. want: %v, current: %v ", nodeGroup.Name, *currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultBGPCommunity, nodeGroup.Status.FrontEndFabricStatus.BGPConfigStatus.LastObservedReadyBGP))
									validationResult = false
								}
							}
						}
					}

				}
			}
		}

		if currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy != nil {

			if nodeGroup.Spec.AcceleratorFabricConfig == nil {
				logger.Info("currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy is NOT nil, but nodeGroup.Spec.AcceleratorFabricConfig is nil")
				validationResult = false
			} else {

				if currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID != nil {
					if nodeGroup.Spec.AcceleratorFabricConfig.VlanConf == nil {
						logger.Info("nodeGroup.Spec.AcceleratorFabricConfig.VlanConf is nil")
						validationResult = false
					} else {

						if nodeGroup.Spec.AcceleratorFabricConfig.VlanConf.VlanID != idcnetworkv1alpha1.NOOPVlanID {

							if nodeGroup.Spec.AcceleratorFabricConfig.VlanConf.VlanID != *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID {
								logger.Info(fmt.Sprintf("nodeGroup %v accelerator Vlan need to set to default value %v before moving to other Pool", nodeGroup.Name, *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID))
								validationResult = false
							}

							// check nodeGroup vlan status is in ready state
							if nodeGroup.Status.AcceleratorFabricStatus == nil {
								logger.Info("nodeGroup.Status.AcceleratorFabricStatus is nil")
								validationResult = false
							} else {
								if !l.isNodeGroupVlanStatusReady(ctx, nodeGroup.Status.AcceleratorFabricStatus, currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy) {
									logger.Info(fmt.Sprintf("nodeGroup %v ACC Vlan status is not ready. want: %v, current: %v ", nodeGroup.Name, *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID, nodeGroup.Status.AcceleratorFabricStatus.VlanConfigStatus.LastObservedReadyVLAN))
									validationResult = false
								}
							}
						}
					}
				}
				if currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultBGPCommunity != nil {

					if nodeGroup.Spec.AcceleratorFabricConfig.BGPConf == nil {
						logger.Info("currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultBGPCommunity is NOT nil, but nodeGroup.Spec.AcceleratorFabricConfig.BGPConf is nil")
						validationResult = false
					} else {

						if nodeGroup.Spec.AcceleratorFabricConfig.BGPConf.BGPCommunity != idcnetworkv1alpha1.NOOPBGPCommunity {

							if nodeGroup.Spec.AcceleratorFabricConfig.BGPConf.BGPCommunity != *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultBGPCommunity {
								logger.Info(fmt.Sprintf("nodeGroup %v accelerator BGP need to set to default value %v before moving to other Pool", nodeGroup.Name, *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultBGPCommunity))
								validationResult = false
							}

							// check nodeGroup BGP status is in ready state
							if nodeGroup.Status.AcceleratorFabricStatus == nil {
								logger.Info("nodeGroup.Status.AcceleratorFabricStatus is nil")
								validationResult = false
							} else {
								if !l.isNodeGroupBGPStatusReady(ctx, nodeGroup.Status.AcceleratorFabricStatus, currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy) {
									logger.Info(fmt.Sprintf("nodeGroup %v ACC BGP status is not ready. want: %v, current: %v ", nodeGroup.Name, *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultBGPCommunity, nodeGroup.Status.AcceleratorFabricStatus.BGPConfigStatus.LastObservedReadyBGP))
									validationResult = false
								}
							}
						}
					}
				}
			}
		}

		if currentPool.NetworkConfigStrategy.StorageFabricStrategy != nil {

			if nodeGroup.Spec.StorageFabricConfig == nil {
				logger.Info("currentPool.NetworkConfigStrategy.StorageFabricStrategy is NOT nil, but nodeGroup.Spec.StorageFabricConfig is nil")
				validationResult = false
			} else {

				if currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID != nil {

					if nodeGroup.Spec.StorageFabricConfig.VlanConf == nil {
						logger.Info("currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID is NOT nil, but nodeGroup.Spec.StorageFabricConfig.VlanConf is nil")
						validationResult = false
					} else {

						if nodeGroup.Spec.StorageFabricConfig.VlanConf.VlanID != idcnetworkv1alpha1.NOOPVlanID {

							if nodeGroup.Spec.StorageFabricConfig.VlanConf.VlanID != *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID {
								logger.Info(fmt.Sprintf("nodeGroup %v storage Vlan need to set to default value %v before moving to other Pool", nodeGroup.Name, *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID))
								validationResult = false
							}

							// check nodeGroup vlan status is in ready state
							if nodeGroup.Status.StorageFabricStatus == nil {
								logger.Info(" nodeGroup.Status.StorageFabricStatus is nil")
								validationResult = false
							} else {
								if !l.isNodeGroupVlanStatusReady(ctx, nodeGroup.Status.StorageFabricStatus, currentPool.NetworkConfigStrategy.StorageFabricStrategy) {
									logger.Info(fmt.Sprintf("nodeGroup %v storage Vlan status is not ready. want: %v, current: %v ", nodeGroup.Name, *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID, nodeGroup.Status.StorageFabricStatus.VlanConfigStatus.LastObservedReadyVLAN))
									validationResult = false
								}
							}
						}

					}
				}

				if currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultBGPCommunity != nil {
					if nodeGroup.Spec.StorageFabricConfig.BGPConf == nil {
						logger.Info("currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultBGPCommunity is NOT nil, but nodeGroup.Spec.StorageFabricConfig.BGPConf is nil")
						validationResult = false
					} else {

						if nodeGroup.Spec.StorageFabricConfig.BGPConf.BGPCommunity != idcnetworkv1alpha1.NOOPBGPCommunity {

							if nodeGroup.Spec.StorageFabricConfig.BGPConf.BGPCommunity != *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultBGPCommunity {
								logger.Info(fmt.Sprintf("nodeGroup %v storage BGP need to set to default value %v before moving to other Pool", nodeGroup.Name, *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultBGPCommunity))
								validationResult = false
							}

							// check nodeGroup BGP status is in ready state
							if nodeGroup.Status.StorageFabricStatus == nil {
								logger.Info("nodeGroup.Status.StorageFabricStatus is nil")
								validationResult = false
							} else {
								if !l.isNodeGroupBGPStatusReady(ctx, nodeGroup.Status.StorageFabricStatus, currentPool.NetworkConfigStrategy.StorageFabricStrategy) {
									logger.Info(fmt.Sprintf("nodeGroup %v storage BGP status is not ready. want: %v, current: %v ", nodeGroup.Name, *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultBGPCommunity, nodeGroup.Status.StorageFabricStatus.BGPConfigStatus.LastObservedReadyBGP))
									validationResult = false
								}
							}
						}
					}
				}
			}

		}
	} else if currentPool.SchedulingConfig.MinimumSchedulableUnit == idcnetworkv1alpha1.MSUNetworkNode {
		// for MSU is NetworkNode, iterate all the NetworkNodes to make sure they are in the default state.
		for _, nn := range nodeGroup.Spec.NetworkNodes {
			existingNetworkNode := &idcnetworkv1alpha1.NetworkNode{}
			key := types.NamespacedName{Name: nn, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err := l.NwcpK8sClient.Get(ctx, key, existingNetworkNode)
			if err != nil {
				logger.Error(err, "NwcpK8sClient Get NetworkNode failed")
				validationResult = false
			}

			if currentPool.NetworkConfigStrategy.FrontEndFabricStrategy != nil {
				if existingNetworkNode.Spec.FrontEndFabric == nil {
					logger.Info("currentPool.NetworkConfigStrategy.FrontEndFabricStrategy is NOT nil, but NetworkNode.Spec.FrontEndFabric is nil")
					validationResult = false
				} else {
					if existingNetworkNode.Spec.FrontEndFabric.VlanId != idcnetworkv1alpha1.NOOPVlanID {
						if currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeVLAN {

							if existingNetworkNode.Spec.FrontEndFabric.VlanId != *currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID {
								logger.Info(fmt.Sprintf("networkNode %v frontend Vlan need to set to default value %v before moving to other Pool", nn, *currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID))
								validationResult = false
							}

							// check the status
							if existingNetworkNode.Status.FrontEndFabricStatus.LastObservedVlanId != *currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.ProvisionConfig.DefaultVlanID {
								// the front end vlan is not ready yet
								validationResult = false
							}

						} else if currentPool.NetworkConfigStrategy.FrontEndFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeBGP {
							// it should NOT be BGP for MSU NetworkNode
							logger.Info("MSU is NetworkNode, but BGP IsolationType is defined for front-end fabric")
							// return false
							validationResult = false
						}
					}
				}
			}

			if currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy != nil {
				if existingNetworkNode.Spec.AcceleratorFabric == nil {
					logger.Info("currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy is NOT nil, but NetworkNode.Spec.AcceleratorFabric is nil")
					validationResult = false
				} else {
					if existingNetworkNode.Spec.AcceleratorFabric.VlanId != idcnetworkv1alpha1.NOOPVlanID {
						if currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeVLAN {

							if existingNetworkNode.Spec.AcceleratorFabric.VlanId != *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID {
								logger.Info(fmt.Sprintf("networkNode %v accelerator Vlan need to set to default value %v before moving to other Pool", nn, *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID))
								validationResult = false
							}

							for _, spStatus := range existingNetworkNode.Status.AcceleratorFabricStatus.SwitchPorts {
								if spStatus.LastObservedVlanId != *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID {
									logger.Info(fmt.Sprintf("networkNode %v accelerator fabric SwitchPort %v's Vlan hasn't been updated to %v", nn, spStatus.SwitchPort, *currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.ProvisionConfig.DefaultVlanID))
									validationResult = false
								}
							}

						} else if currentPool.NetworkConfigStrategy.AcceleratorFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeBGP {
							// it should NOT be BGP for MSU NetworkNode
							logger.Info("MSU is NetworkNode, but BGP IsolationType is defined for ACC fabric")
							validationResult = false

						}
					}
				}
			}

			if currentPool.NetworkConfigStrategy.StorageFabricStrategy != nil {
				if existingNetworkNode.Spec.StorageFabric == nil {
					logger.Info("currentPool.NetworkConfigStrategy.StorageFabricStrategy is NOT nil, but NetworkNode.Spec.StorageFabric is nil")
					validationResult = false
				} else {
					if existingNetworkNode.Spec.StorageFabric.VlanId != idcnetworkv1alpha1.NOOPVlanID {
						if currentPool.NetworkConfigStrategy.StorageFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeVLAN {

							if existingNetworkNode.Spec.StorageFabric.VlanId != *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID {
								logger.Info(fmt.Sprintf("networkNode %v storage Vlan's Spec need to set from default value %v before moving to other Pool", nn, *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID))
								validationResult = false
							}

							for _, spStatus := range existingNetworkNode.Status.StorageFabricStatus.SwitchPorts {
								if spStatus.LastObservedVlanId != *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID {
									logger.Info(fmt.Sprintf("networkNode %v storage fabric SwitchPort %v's Vlan [%v] has to be updated to [%v] before moving to other pool", nn, spStatus.SwitchPort, spStatus.LastObservedVlanId, *currentPool.NetworkConfigStrategy.StorageFabricStrategy.ProvisionConfig.DefaultVlanID))
									validationResult = false
								}
							}
						} else if currentPool.NetworkConfigStrategy.StorageFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeBGP {
							// it should NOT be BGP for MSU NetworkNode
							logger.Info("MSU is NetworkNode, but BGP IsolationType is defined for Storage fabric")
							validationResult = false
						}
					}
				}
			}
		}
	}
	return validationResult
}

func (l *PoolManager) updateNodeGroup(ctx context.Context, nodeGroupName string, desiredPoolName string) error {
	logger := log.FromContext(ctx).WithName("PoolManager.updateNodeGroup")

	targetPool, err := l.GetPoolByName(ctx, desiredPoolName)
	if err != nil {
		return fmt.Errorf("GetPoolByName failed, pool name: %v, %v", desiredPoolName, err)
	}

	if targetPool.NetworkConfigStrategy == nil {
		return fmt.Errorf("targetPool.NetworkConfigStrategy is nil")
	}

	// get the latest NodeGroup. (we may have done some update before, so better fetch the latest object.)
	existingNodeGroup := &idcnetworkv1alpha1.NodeGroup{}
	key := types.NamespacedName{Name: nodeGroupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = l.NwcpK8sClient.Get(ctx, key, existingNodeGroup)
	if err != nil {
		return err
	}

	newNodeGroupForSpecPatch := existingNodeGroup.DeepCopy()
	newNodeGroupForStatusUpdate := existingNodeGroup.DeepCopy()

	// reset the Status fields
	newNodeGroupForStatusUpdate.Status.FrontEndFabricStatus = nil
	newNodeGroupForStatusUpdate.Status.AcceleratorFabricStatus = nil
	newNodeGroupForStatusUpdate.Status.StorageFabricStatus = nil

	// if the MSU is NetworkNode, remove all the Spec and Status fields, as we don't need to push the VLAN norther BGP config from NG to its NN or SW.
	if targetPool.SchedulingConfig.MinimumSchedulableUnit == idcnetworkv1alpha1.MSUNetworkNode {
		// reset the Spec fields
		newNodeGroupForSpecPatch.Spec.FrontEndFabricConfig = nil
		newNodeGroupForSpecPatch.Spec.AcceleratorFabricConfig = nil
		newNodeGroupForSpecPatch.Spec.StorageFabricConfig = nil
		// reset the Status fields
		// newNodeGroupForStatusUpdate.Status.FrontEndFabricStatus = nil
		// newNodeGroupForStatusUpdate.Status.AcceleratorFabricStatus = nil
		// newNodeGroupForStatusUpdate.Status.StorageFabricStatus = nil

	} else if targetPool.SchedulingConfig.MinimumSchedulableUnit == idcnetworkv1alpha1.MSUNodeGroup {
		// if the MSU is NodeGroup, then update the NG based on the Pool config template.
		////////////////
		// FE
		////////////////
		newNodeGroupForSpecPatch.Spec.FrontEndFabricConfig, err = l.doUpdateNodeGroup(targetPool.NetworkConfigStrategy.FrontEndFabricStrategy, newNodeGroupForSpecPatch.Spec.FrontEndFabricConfig)
		if err != nil {
			logger.Error(err, "doUpdateNodeGroup for frontend fabric failed")
			return err
		}

		////////////////
		// ACC
		////////////////
		newNodeGroupForSpecPatch.Spec.AcceleratorFabricConfig, err = l.doUpdateNodeGroup(targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy, newNodeGroupForSpecPatch.Spec.AcceleratorFabricConfig)
		if err != nil {
			logger.Error(err, "doUpdateNodeGroup for acc fabric failed")
			return err
		}

		////////////////
		// STRG
		////////////////
		newNodeGroupForSpecPatch.Spec.StorageFabricConfig, err = l.doUpdateNodeGroup(targetPool.NetworkConfigStrategy.StorageFabricStrategy, newNodeGroupForSpecPatch.Spec.StorageFabricConfig)
		if err != nil {
			logger.Error(err, "doUpdateNodeGroup for storage fabric failed")
			return err
		}
	}

	err = l.NwcpK8sClient.Status().Update(ctx, newNodeGroupForStatusUpdate)
	if err != nil {
		logger.Error(err, "update NodeGroup status failed")
	}

	patch := client.MergeFrom(existingNodeGroup)
	err = l.NwcpK8sClient.Patch(ctx, newNodeGroupForSpecPatch, patch)
	if err != nil {
		return err
	}

	return nil
}

func (l *PoolManager) doUpdateNodeGroup(networkStrategyTemplate *idcnetworkv1alpha1.NetworkStrategy, fabricConfig *idcnetworkv1alpha1.FabricConfig) (*idcnetworkv1alpha1.FabricConfig, error) {
	if networkStrategyTemplate == nil {
		return nil, nil
	}
	if fabricConfig == nil {
		fabricConfig = &idcnetworkv1alpha1.FabricConfig{}
	}

	if networkStrategyTemplate.ProvisionConfig.DefaultVlanID != nil {
		// Vlan is set. update the NG spec.
		if fabricConfig.VlanConf == nil {
			fabricConfig.VlanConf = &idcnetworkv1alpha1.VlanConfig{}
		}
		fabricConfig.VlanConf.VlanID = *networkStrategyTemplate.ProvisionConfig.DefaultVlanID
	} else {
		// Vlan is NOT set, remove the related spec and status fields.
		fabricConfig.VlanConf = nil
		// remove the VLAN status field
		// frontendStatusConfig.VlanConfigStatus = idcnetworkv1alpha1.VlanConfigStatus{}
	}
	if networkStrategyTemplate.ProvisionConfig.DefaultBGPCommunity != nil {
		// BGP is set. update the NG spec.
		if fabricConfig.BGPConf == nil {
			fabricConfig.BGPConf = &idcnetworkv1alpha1.BGPConfig{}
		}
		fabricConfig.BGPConf.BGPCommunity = *networkStrategyTemplate.ProvisionConfig.DefaultBGPCommunity
	} else {
		// BGP is NOT set, remove the related spec and status fields.
		fabricConfig.BGPConf = nil
		// remove the BGP status field
		// frontendStatusConfig.BGPConfigStatus = idcnetworkv1alpha1.BGPConfigStatus{}
	}
	return fabricConfig, nil
}

func (l *PoolManager) updateNetworkNode(ctx context.Context, networkNodeName string, targetPoolName string) error {
	logger := log.FromContext(ctx).WithName("PoolManager.updateNetworkNode")

	targetPool, err := l.GetPoolByName(ctx, targetPoolName)
	if err != nil {
		return fmt.Errorf("GetPoolByName failed, pool name: %v, %v", targetPoolName, err)
	}

	if targetPool.NetworkConfigStrategy == nil {
		return fmt.Errorf("targetPool.NetworkConfigStrategy is nil")
	}

	// get the latest NodeGroup. (we may have done some update before, so better fetch the latest object.)
	existingNetworkNode := &idcnetworkv1alpha1.NetworkNode{}
	key := types.NamespacedName{Name: networkNodeName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = l.NwcpK8sClient.Get(ctx, key, existingNetworkNode)
	if err != nil {
		return err
	}

	newNetworkNodeForSpecPatch := existingNetworkNode.DeepCopy()
	newNetworkNodeForStatusUpdate := existingNetworkNode.DeepCopy()

	// if the MSU is NetworkNode, set the NN based on the pool template.
	// !!! Note: this code will replace the existing NN values with the default values.
	// we have the logic that check if the current NG/NN are in default state before adjusting a NG's pool, so this won't update NN that being used, make sure this logic is working as expected.

	////////////////
	// FE
	////////////////
	feNetworkStrategy := targetPool.NetworkConfigStrategy.FrontEndFabricStrategy
	if feNetworkStrategy != nil && feNetworkStrategy.ProvisionConfig.DefaultVlanID != nil {
		if newNetworkNodeForSpecPatch.Spec.FrontEndFabric == nil {
			newNetworkNodeForSpecPatch.Spec.FrontEndFabric = &idcnetworkv1alpha1.FrontEndFabric{}
		}
		// Vlan is set. update the NN spec. Only do this if MSU is NetworkNode, when MSU is NodeGroup, the values in the NodeGroup will be pushed to the NN.
		if targetPool.SchedulingConfig.MinimumSchedulableUnit == idcnetworkv1alpha1.MSUNetworkNode {
			newNetworkNodeForSpecPatch.Spec.FrontEndFabric.VlanId = *feNetworkStrategy.ProvisionConfig.DefaultVlanID
		}
	} else {
		// feNetworkStrategy is nil or feNetworkStrategy Vlan is NOT set, remove the related spec and status fields.
		// newNetworkNodeForSpecPatch.Spec.FrontEndFabric = nil
		newNetworkNodeForSpecPatch.Spec.FrontEndFabric = &idcnetworkv1alpha1.FrontEndFabric{}
		// remove the VLAN status field
		newNetworkNodeForStatusUpdate.Status.FrontEndFabricStatus = idcnetworkv1alpha1.FrontEndFabricStatus{}

	}

	////////////////
	// ACC
	////////////////
	accNetworkStrategy := targetPool.NetworkConfigStrategy.AcceleratorFabricStrategy
	if accNetworkStrategy != nil && accNetworkStrategy.ProvisionConfig.DefaultVlanID != nil {
		if newNetworkNodeForSpecPatch.Spec.AcceleratorFabric == nil {
			newNetworkNodeForSpecPatch.Spec.AcceleratorFabric = &idcnetworkv1alpha1.AcceleratorFabric{}
		}
		// Vlan is set. update the NN spec.
		if targetPool.SchedulingConfig.MinimumSchedulableUnit == idcnetworkv1alpha1.MSUNetworkNode {
			newNetworkNodeForSpecPatch.Spec.AcceleratorFabric.VlanId = *accNetworkStrategy.ProvisionConfig.DefaultVlanID
		}
	} else {
		// Vlan is NOT set, remove the related spec and status fields.
		// newNetworkNodeForSpecPatch.Spec.AcceleratorFabric = nil
		newNetworkNodeForSpecPatch.Spec.AcceleratorFabric = &idcnetworkv1alpha1.AcceleratorFabric{}

		// remove the VLAN status field
		newNetworkNodeForStatusUpdate.Status.AcceleratorFabricStatus = idcnetworkv1alpha1.AcceleratorFabricStatus{}
	}

	////////////////
	// STRG
	////////////////
	strgNetworkStrategy := targetPool.NetworkConfigStrategy.StorageFabricStrategy
	if strgNetworkStrategy != nil && strgNetworkStrategy.ProvisionConfig.DefaultVlanID != nil {
		if newNetworkNodeForSpecPatch.Spec.StorageFabric == nil {
			newNetworkNodeForSpecPatch.Spec.StorageFabric = &idcnetworkv1alpha1.StorageFabric{}
		}
		// Vlan is set. update the NN spec.
		if targetPool.SchedulingConfig.MinimumSchedulableUnit == idcnetworkv1alpha1.MSUNetworkNode {
			newNetworkNodeForSpecPatch.Spec.StorageFabric.VlanId = *strgNetworkStrategy.ProvisionConfig.DefaultVlanID
		}
	} else {
		// Vlan is NOT set, remove the related spec and status fields.
		// newNetworkNodeForSpecPatch.Spec.StorageFabric = nil
		newNetworkNodeForSpecPatch.Spec.StorageFabric = &idcnetworkv1alpha1.StorageFabric{}

		// remove the VLAN status field
		newNetworkNodeForStatusUpdate.Status.StorageFabricStatus = idcnetworkv1alpha1.StorageFabricStatus{}
	}

	err = l.NwcpK8sClient.Status().Update(ctx, newNetworkNodeForStatusUpdate)
	if err != nil {
		logger.Error(err, "update NodeGroup status failed")
	}

	// only when all the above work are successfully done, we update the Pool label.
	patch := client.MergeFrom(existingNetworkNode)
	err = l.NwcpK8sClient.Patch(ctx, newNetworkNodeForSpecPatch, patch)
	if err != nil {
		return err
	}

	return nil
}

func (l *PoolManager) updateSwitch(ctx context.Context, networkStrategy *idcnetworkv1alpha1.NetworkStrategy, switchFQDN string) error {
	logger := log.FromContext(ctx).WithName("PoolManager.updateSwitch")

	existingSwitch := &idcnetworkv1alpha1.Switch{}
	key := types.NamespacedName{Name: switchFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := l.NwcpK8sClient.Get(ctx, key, existingSwitch)
	if err != nil {
		return err
	}

	newSwitchForSpecPatch := existingSwitch.DeepCopy()
	newSwitchForStatusUpdate := existingSwitch.DeepCopy()

	if networkStrategy == nil {
		newSwitchForSpecPatch.Spec.BGP = nil
		newSwitchForStatusUpdate.Status.SwitchBGPConfigStatus = nil
	} else {
		if networkStrategy.ProvisionConfig.DefaultBGPCommunity != nil {
			if newSwitchForSpecPatch.Spec.BGP == nil {
				newSwitchForSpecPatch.Spec.BGP = &idcnetworkv1alpha1.BGPConfig{}
			}
			newSwitchForSpecPatch.Spec.BGP.BGPCommunity = *networkStrategy.ProvisionConfig.DefaultBGPCommunity
		} else {
			// BGP is NOT set, remove the related spec and status fields.
			newSwitchForSpecPatch.Spec.BGP = nil
			newSwitchForStatusUpdate.Status.SwitchBGPConfigStatus = nil
		}

	}

	err = l.NwcpK8sClient.Status().Update(ctx, newSwitchForStatusUpdate)
	if err != nil {
		logger.Error(err, "update NodeGroup status failed")
	}

	// only when all the above work are successfully done, we update the Pool label.
	patch := client.MergeFrom(existingSwitch)
	err = l.NwcpK8sClient.Patch(ctx, newSwitchForSpecPatch, patch)
	if err != nil {
		return err
	}

	return nil
}

type PoolConfigReader interface {
	GetPoolConfigByName(ctx context.Context, poolName string) (*idcnetworkv1alpha1.Pool, error)
	ListPoolConfigs(ctx context.Context) ([]*idcnetworkv1alpha1.Pool, error)
}

type LocalPoolConfig struct {
	poolConfigFilePath string
	// key: pool name, val: pool
	pools map[string]*idcnetworkv1alpha1.Pool
}

func NewLocalPoolConfig(conf *idcnetworkv1alpha1.SDNControllerConfig) (*LocalPoolConfig, error) {
	// get the pool configuration
	poolByteValue, err := os.ReadFile(conf.ControllerConfig.PoolsConfigFilePath)
	if err != nil {
		return nil, err
	}
	var poolList idcnetworkv1alpha1.PoolList
	err = json.Unmarshal(poolByteValue, &poolList)
	if err != nil {
		return nil, err
	}
	pools := make(map[string]*idcnetworkv1alpha1.Pool)
	for i := range poolList.Items {
		pools[poolList.Items[i].Name] = poolList.Items[i]
	}

	return &LocalPoolConfig{
		poolConfigFilePath: conf.ControllerConfig.PoolsConfigFilePath,
		pools:              pools,
	}, nil
}

// GetPoolConfigByName returns the Pool that a NodeGroup belongs to. This implementation read it from the local file system.
func (l *LocalPoolConfig) GetPoolConfigByName(ctx context.Context, poolName string) (*idcnetworkv1alpha1.Pool, error) {
	pool, found := l.pools[poolName]
	if !found || pool == nil {
		return nil, fmt.Errorf("cannot find pool [%v] from the Pool configuration file %v", poolName, l.poolConfigFilePath)
	}
	return pool, nil
}

// ListPoolConfigs returns all the pools
func (l *LocalPoolConfig) ListPoolConfigs(ctx context.Context) ([]*idcnetworkv1alpha1.Pool, error) {
	res := make([]*idcnetworkv1alpha1.Pool, 0)
	for _, pool := range l.pools {
		res = append(res, pool)
	}
	return res, nil
}

// NodeUsageReporter knows the usage of a node, for example, if a node is reserved or not.
type NodeUsageReporter interface {
	IsNodeReserved(nodeName string) (bool, error)
}

func NewBMHUsageReporter() *BMHUsageReporter {
	bmhs := make(map[string]*baremetalv1alpha1.BareMetalHost)
	return &BMHUsageReporter{
		BMHs: bmhs,
	}
}

type BMHUsageReporter struct {
	sync.Mutex
	// key: BMH name, value: BMH instance
	// it's possible that these BMHs are from different metal3 clusters, but we assume that BMaaS will ensure there is no duplicated BMH across the clusters.
	BMHs map[string]*baremetalv1alpha1.BareMetalHost
}

func (um *BMHUsageReporter) ReportBMH(bmh *baremetalv1alpha1.BareMetalHost) {
	um.Lock()
	defer um.Unlock()
	um.BMHs[bmh.Name] = bmh
}

func (um *BMHUsageReporter) RemoveBMH(bmhName string) {
	um.Lock()
	defer um.Unlock()
	delete(um.BMHs, bmhName)
}

func (um *BMHUsageReporter) IsNodeReserved(nodeName string) (bool, error) {
	um.Lock()
	defer um.Unlock()
	bmh, found := um.BMHs[nodeName]
	if !found {
		return false, fmt.Errorf("bmh %v not found", nodeName)
	}
	if bmh.Spec.ConsumerRef != nil && len(bmh.Spec.ConsumerRef.Name) > 0 {
		return true, nil
	}
	return false, nil
}
