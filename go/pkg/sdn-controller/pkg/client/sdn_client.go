// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(idcnetworkv1alpha1.AddToScheme(scheme))
}

// SDNClient
type SDNClient struct {
	config SDNClientConfig
	// dynamicClient       dynamic.Interface
	k8sClient           client.Client
	watchTimeoutSeconds int

	// pools and nodeGroupToPoolMapping are the fields for pool operation.
	// can be removed after moving these information to Netbox or other DBs.
	pools                  map[string]*idcnetworkv1alpha1.Pool
	nodeGroupToPoolMapping map[string]string
	mConfig                map[string]*idcnetworkv1alpha1.SDNControllerConfig
}

type SDNClientConfig struct {
	KubeConfig string

	// file paths for the Pool and Group-Pool mapping
	LocalPoolInfoFilePath          string
	NodeGroupToPoolMappingFilePath string
}

var switchPortGVR = idcnetworkv1alpha1.SchemeBuilder.GroupVersion.WithResource("switchports")
var networkNodeGVR = idcnetworkv1alpha1.SchemeBuilder.GroupVersion.WithResource("networknodes")
var nodeGroupsGVR = idcnetworkv1alpha1.SchemeBuilder.GroupVersion.WithResource("nodegroups")

func NewSDNClient(ctx context.Context, conf SDNClientConfig) (*SDNClient, error) {
	// if len(conf.KubeConfig) == 0 {
	// 	return nil, fmt.Errorf("kubeConfig file is not provided")
	// }

	sdnClient := &SDNClient{
		config: conf,
	}
	err := sdnClient.Reconnect(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewSDNClient failed: %w", err)
	}

	return sdnClient, nil
}

// NewSDNClientLocal is for SDN local subcomponent, and it also has the pool info with it.
func NewSDNClientLocal(ctx context.Context, conf SDNClientConfig) (*SDNClient, error) {
	sdnClient := &SDNClient{
		config: conf,
	}

	err := sdnClient.Reconnect(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewSDNClient failed: %w", err)
	}

	if len(conf.LocalPoolInfoFilePath) > 0 {
		pools, err := utils.GetPools(conf.LocalPoolInfoFilePath)
		if err != nil {
			return nil, err
		}
		sdnClient.pools = pools
	}

	if len(conf.NodeGroupToPoolMappingFilePath) > 0 {
		// read the NodeGroup to Pool mapping information
		groupToPoolMappings, err := utils.GetNodeGroupToPoolMapping(conf.NodeGroupToPoolMappingFilePath)
		if err != nil {
			return nil, err
		}
		sdnClient.nodeGroupToPoolMapping = groupToPoolMappings.NodeGroupToPoolMap
	}

	return sdnClient, nil
}

func k8sGetFailedShouldRetry(err error) bool {
	// do not retry if the object is not found
	if !apierrors.IsNotFound(err) {
		return true
	}
	return false
}

func k8sPatchFailedShouldRetry(err error) bool {
	var urlError *url.Error
	if errors.As(err, &urlError) {
		if strings.Contains(err.Error(), "connection lost") {
			return true
		}
		if strings.Contains(err.Error(), "connection refused") {
			return true
		}
		if strings.Contains(err.Error(), "i/o timeout") {
			return true
		}
	}

	if strings.Contains(err.Error(), "invalid argument") {
		return true
	}
	if strings.Contains(err.Error(), `"sdn-bmaas-role" not found`) {
		return true
	}
	if errors.Is(err, ErrK8sClientNil) {
		return true
	}

	return false
}

func (c *SDNClient) Reconnect(ctx context.Context) error {
	if len(c.config.KubeConfig) > 0 {
		k8sClient := utils.NewK8SClientFromConfAndScheme(ctx, c.config.KubeConfig, scheme)
		c.k8sClient = k8sClient
	} else {
		// if kubeconfig file is not provided, try to build it from the local file.
		c.k8sClient = utils.NewK8SClient()
	}
	if c.k8sClient == nil {
		return fmt.Errorf("unable to init K8s client")
	}
	return nil
}

func (c *SDNClient) GetSwitchPort(ctx context.Context, switchFQDN string, port string) (*idcnetworkv1alpha1.SwitchPort, error) {
	logger := log.FromContext(ctx).WithName("SDNClient.GetSwitchPort").WithValues(utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortName, port)
	err := utils.ValidatePortValue(port)
	if err != nil {
		return nil, fmt.Errorf("ValidatePortValue failed, %v", err)
	}
	err = utils.ValidateSwitchFQDN(switchFQDN, "")
	if err != nil {
		return nil, fmt.Errorf("ValidateSwitchFQDN failed, %v", err)
	}

	switchPortName := utils.GeneratePortFullName(switchFQDN, port)
	switchPort := &idcnetworkv1alpha1.SwitchPort{}
	key := types.NamespacedName{Name: switchPortName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = utils.ExecuteWithRetry(
		func() error {
			// check for nil object, it is possible that the k8s client will be nil.
			// if all the K8s nodes defined in the kubeconfig file fail to connect, then the Reconnect() will return a nil k8s client.
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Get(ctx, key, switchPort)
		},
		k8sGetFailedShouldRetry,
		func() error { return c.Reconnect(ctx) },
		3)
	if err != nil {
		return nil, fmt.Errorf("retried and failed to getSwitchPort too many times, %v", err)
	}

	logger.Info("get K8s SwitchPort success")
	return switchPort, nil

}

// UpdateVlan updates a single SwitchPort CR in the nwcp cluster.
// Deprecated: Updates should be made to the NetworkNode CR to update network settings for an entire server, not individual ports.
func (c *SDNClient) UpdateVlan(ctx context.Context, switchFQDN string, port string, vlan int64, description string) error {
	logger := log.FromContext(ctx).WithName("SDNClient.UpdateVlan").WithValues(utils.LogFieldSwitchFQDN, switchFQDN, utils.LogFieldSwitchPortName, port, utils.LogFieldVlanID, vlan)
	logger.Info("Warning: using deprecated SDNClient.UpdateVlan(). This function will be removed from the SDN client library soon. You should use SDNClient.UpdateNetworkNodeConfig() or SDNClient.UpdateNodeGroupStatus() instead.")

	// TODO: we do have these validations in the controller, but the validation results would NOT be returned to the caller.
	// We may want to implement the validations in the webhook in the future, so it return errors to the caller before modifying the K8s object.
	mConfig, err := c.GetNetworkManagerConfig(ctx)
	if err != nil {
		return fmt.Errorf("sdn client get manager configuration failed, %v", err)
	}

	var allowedVlanIds []int
	// there are cases that it got nothing from SDN, for instance, SDN is not yet updated to latest version, thus mConfig.ControllerConfig.AllowedVlanIds is empty.
	// in this case, check the default allowed vlans instead
	allowedVlanIdsStr := utils.DefaultAllowedVlanIdsStr
	if len(mConfig.ControllerConfig.AllowedVlanIds) > 0 {
		allowedVlanIdsStr = mConfig.ControllerConfig.AllowedVlanIds
	}
	allowedVlanIds, err = utils.ExpandVlanRanges(allowedVlanIdsStr)
	if err != nil {
		return fmt.Errorf("Error expanding valid VLAN range, %v", err)
	}

	err = utils.ValidatePortValue(port)
	if err != nil {
		return fmt.Errorf("ValidatePortValue failed, %v", err)
	}
	err = utils.ValidateVlanValue(int(vlan), allowedVlanIds)
	if err != nil {
		return fmt.Errorf("ValidateVlanValue failed, %v", err)
	}
	err = utils.ValidateSwitchFQDN(switchFQDN, "")
	if err != nil {
		return fmt.Errorf("ValidateSwitchFQDN failed, %v", err)
	}

	switchPortName := utils.GeneratePortFullName(switchFQDN, port)

	existingSwitchPort := &idcnetworkv1alpha1.SwitchPort{}
	key := types.NamespacedName{Name: switchPortName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = utils.ExecuteWithRetry(
		func() error {
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Get(ctx, key, existingSwitchPort)
		},
		k8sGetFailedShouldRetry,
		func() error { return c.Reconnect(ctx) },
		3)
	if err != nil {
		logger.Error(err, "k8sClient.Get SwitchPort failed")
		return fmt.Errorf("k8sClient.Get SwitchPort failed, %v", err)
	}

	newSwitchPort := existingSwitchPort.DeepCopy()
	patch := client.MergeFrom(existingSwitchPort)
	newSwitchPort.Spec.VlanId = vlan
	newSwitchPort.Spec.Description = description
	err = utils.ExecuteWithRetry(
		func() error {
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Patch(ctx, newSwitchPort, patch)
		},
		k8sPatchFailedShouldRetry,
		func() error { return c.Reconnect(ctx) }, 3)
	if err != nil {
		return fmt.Errorf("k8sClient.Patch failed: %v", err)
	}

	return fmt.Errorf("retried and failed to UpdateVlan too many times")
}

// NetworkNodeConfUpdateRequest is the request to be sent to the SDN to update the fabric config for a node.
type NetworkNodeConfUpdateRequest struct {
	NetworkNodeName                 string
	FrontEndFabricVlan              int64
	FrontEndFabricMode              string
	FrontEndFabricTrunkGroups       []string
	FrontEndFabricNativeVlan        int64
	AcceleratorFabricVlan           int64
	AcceleratorFabricBGPCommunityID int64
	StorageFabricVlan               int64
	Description                     string
}

// NetworkNodeConfStatusCheckRequest is the request to check if the target NetworkNode has the desired state ready
type NetworkNodeConfStatusCheckRequest struct {
	NetworkNodeName                  string
	DesiredFrontEndFabricVlan        int64
	DesiredAcceleratorFabricVlan     int64
	DesiredFrontEndFabricMode        string
	DesiredFrontEndFabricTrunkGroups []string
	DesiredFrontEndFabricNativeVlan  int64
	// Note: when DesiredAcceleratorFabricBGPCommunityID is specified, SDN client will try update the BGP Community ID for the ACC switch a node is connected to.
	DesiredAcceleratorFabricBGPCommunityID int64
	DesiredStorageFabricVlan               int64
}

type NetworkNodeConfStatusCheckResponse struct {
	NetworkNode idcnetworkv1alpha1.NetworkNode
	Status      string
}

// UpdateNetworkNodeConfig update the front-end and accelerator VLAN for a NetworkNode.
// Note: SDNClient is NOT responsible for checking if a field should be update or not. For example, SDNClient cannot prevent a caller updating the accelerator vlan for a non-Gaudi node.
func (c *SDNClient) UpdateNetworkNodeConfig(ctx context.Context, request NetworkNodeConfUpdateRequest) error {
	logger := log.FromContext(ctx).WithName("SDNClient.UpdateNetworkNodeConfig").WithValues(utils.LogFieldNetworkNode, request.NetworkNodeName)
	var err error

	mConfig, err := c.GetNetworkManagerConfig(ctx)
	if err != nil {
		return fmt.Errorf("sdn client get manager configuration failed, %v", err)
	}

	var allowedVlanIds []int
	var allowedNativeVlanIds []int

	// get the allowedVlanIds
	allowedVlanIdsStr := utils.DefaultAllowedVlanIdsStr
	if len(mConfig.ControllerConfig.AllowedVlanIds) > 0 {
		allowedVlanIdsStr = mConfig.ControllerConfig.AllowedVlanIds
	}
	allowedVlanIds, err = utils.ExpandVlanRanges(allowedVlanIdsStr)
	if err != nil {
		return fmt.Errorf("Error expanding valid VLAN range, %v", err)
	}

	// get the allowedNativeVlanIds
	allowedNativeVlanIdsStr := utils.DefaultAllowedNativeVlanIdsStr
	if len(mConfig.ControllerConfig.AllowedNativeVlanIds) > 0 {
		allowedNativeVlanIdsStr = mConfig.ControllerConfig.AllowedNativeVlanIds
	}
	allowedNativeVlanIds, err = utils.ExpandVlanRanges(allowedNativeVlanIdsStr)
	if err != nil {
		return fmt.Errorf("Error expanding valid NativeVLANs")
	}

	if len(request.NetworkNodeName) == 0 {
		return fmt.Errorf("NetworkNodeName is not provided")
	}

	// request.FrontEndFabricVlan is set
	if request.FrontEndFabricVlan != 0 {
		frontEndErr := utils.ValidateVlanValue(int(request.FrontEndFabricVlan), allowedVlanIds)
		if frontEndErr != nil {
			return fmt.Errorf("request.FrontEndFabricVlan is invalid, %v", frontEndErr)
		}
	}
	// request.FrontEndMode is set
	if len(request.FrontEndFabricMode) != 0 {
		frontEndModeErr := utils.ValidateModeValue(string(request.FrontEndFabricMode), mConfig.ControllerConfig.AllowedModes)
		if frontEndModeErr != nil {
			return fmt.Errorf("request.FrontEndFabricMode is invalid, %v", frontEndModeErr)
		}
	}
	// request.FrontEndTrunkGroups is set
	if request.FrontEndFabricTrunkGroups != nil {
		if len(request.FrontEndFabricTrunkGroups) == 0 {
			logger.Info("the provided FrontEndFabricTrunkGroups is empty, this will remove all the trunk groups for this front end port")
		}

		frontEndTrunkGroupsErr := utils.ValidateTrunkGroups([]string(request.FrontEndFabricTrunkGroups), nil) // TODO: Could pull list of allowed trunkGroups from SDN API?
		if frontEndTrunkGroupsErr != nil {
			return fmt.Errorf("request.FrontEndFabricTrunkGroups is invalid, %v", frontEndTrunkGroupsErr)
		}
	}
	// request.FrontEndNativeVlan is set
	if request.FrontEndFabricNativeVlan != 0 {
		frontEndNativeVlanErr := utils.ValidateVlanValue(int(request.FrontEndFabricNativeVlan), allowedNativeVlanIds)
		if frontEndNativeVlanErr != nil {
			return fmt.Errorf("request.FrontEndFabricNativeVlan is invalid, %v", frontEndNativeVlanErr)
		}
	}
	// request.AcceleratorFabricVlan is set
	if request.AcceleratorFabricVlan != 0 {
		accelErr := utils.ValidateVlanValue(int(request.AcceleratorFabricVlan), allowedVlanIds)
		if accelErr != nil {
			return fmt.Errorf("request.AcceleratorFabricVlan is invalid, %v", accelErr)
		}
	}
	// request.StorageFabricVlan is set
	if request.StorageFabricVlan != 0 {
		storageErr := utils.ValidateVlanValue(int(request.StorageFabricVlan), allowedVlanIds)
		if storageErr != nil {
			return fmt.Errorf("request.StorageFabricVlan is invalid, %v", storageErr)
		}
	}

	// request.AcceleratorFabricBGPCommunityID is set
	if request.AcceleratorFabricBGPCommunityID != 0 {
		accBGPErr := utils.ValidateBGPCommunityValue(int32(request.AcceleratorFabricBGPCommunityID))
		if accBGPErr != nil {
			return fmt.Errorf("request.AcceleratorFabricBGPCommunityID is invalid, %v", accBGPErr)
		}
	}

	existingNetworkNode := &idcnetworkv1alpha1.NetworkNode{}
	key := types.NamespacedName{Name: request.NetworkNodeName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = utils.ExecuteWithRetry(
		// the function we are going to execute
		func() error {
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Get(ctx, key, existingNetworkNode)
		},
		// the condition function that determine if it should trigger a retry
		k8sGetFailedShouldRetry,
		// actions should be taken when we need a retry
		func() error { return c.Reconnect(ctx) },
		3)
	if err != nil {
		return fmt.Errorf("k8sClient.Get NetworkNode failed: %v", err)
	}

	// TODO: comment this out for now, revisit if we still need this feature as we are updating the config at the NN level
	// check if the NG this NN belongs to is under maintenance, if so, block it from updating this NN.
	// if this NN is NOT belong to any group, ignore this check.
	// if existingNetworkNode.Labels != nil && len(existingNetworkNode.Labels[idcnetworkv1alpha1.LabelGroupID]) > 0 {
	// 	// get the NG
	// 	nodeGroupName := existingNetworkNode.Labels[idcnetworkv1alpha1.LabelGroupID]
	// 	nodeGroup := &idcnetworkv1alpha1.NodeGroup{}
	// 	ngKey := types.NamespacedName{Name: nodeGroupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	// 	err = c.k8sClient.Get(ctx, ngKey, nodeGroup)
	// 	if err != nil {
	// 		logger.Error(err, "k8sClient.Get nodeGroup failed")
	// 		return err
	// 	}

	// 	// check if the NG is in maintenance
	// 	if nodeGroup.Labels != nil && len(nodeGroup.Labels[idcnetworkv1alpha1.LabelMaintenance]) > 0 {
	// 		return fmt.Errorf("the networkNode's nodeGroup [%v] is under maintenance, please try again later", nodeGroup.Name)
	// 	}
	// }

	//////////////////////////////////
	// try to update the VLAN if provided
	//////////////////////////////////
	newNetworkNode := existingNetworkNode.DeepCopy()
	// return an error if we want to set Vlan value for a fabric does exist.
	if request.FrontEndFabricVlan != 0 {

		// if the request include vlan and trunk mode at the same time, reject it.
		// if request.FrontEndFabricMode == idcnetworkv1alpha1.TrunkMode {
		// 	return fmt.Errorf("cannot request vlan update at trunk mode, node name: %v", existingNetworkNode.Name)
		// }

		if newNetworkNode.Spec.FrontEndFabric == nil {
			return fmt.Errorf("networkNode %v has no frontend fabric configured", existingNetworkNode.Name)
		}
		newNetworkNode.Spec.FrontEndFabric.VlanId = request.FrontEndFabricVlan
	}
	if len(request.FrontEndFabricMode) != 0 {
		if newNetworkNode.Spec.FrontEndFabric == nil {
			return fmt.Errorf("networkNode %v has no frontend fabric configured", existingNetworkNode.Name)
		}
		newNetworkNode.Spec.FrontEndFabric.Mode = request.FrontEndFabricMode
	}
	if request.FrontEndFabricTrunkGroups != nil {
		if newNetworkNode.Spec.FrontEndFabric == nil {
			return fmt.Errorf("networkNode %v has no frontend fabric configured", existingNetworkNode.Name)
		}
		if len(request.FrontEndFabricTrunkGroups) == 0 {
			logger.Info("request.FrontEndFabricTrunkGroups is empty")
		}
		newNetworkNode.Spec.FrontEndFabric.TrunkGroups = &request.FrontEndFabricTrunkGroups
	} else {
		// note: when the client specify the DesiredFrontEndFabricTrunkGroups as nil or NOT specifying the DesiredFrontEndFabricTrunkGroups, we won't touch the NN.
		// newNetworkNode.Spec.FrontEndFabric.TrunkGroups = nil
	}

	if request.FrontEndFabricNativeVlan != 0 {
		if newNetworkNode.Spec.FrontEndFabric == nil {
			return fmt.Errorf("networkNode %v has no frontend fabric configured", existingNetworkNode.Name)
		}
		newNetworkNode.Spec.FrontEndFabric.NativeVlan = request.FrontEndFabricNativeVlan
	}
	if request.AcceleratorFabricVlan != 0 {
		if newNetworkNode.Spec.AcceleratorFabric == nil {
			return fmt.Errorf("networkNode %v has no accelerator fabric configured", existingNetworkNode.Name)
		}
		newNetworkNode.Spec.AcceleratorFabric.VlanId = request.AcceleratorFabricVlan
	}
	if request.StorageFabricVlan != 0 {
		if newNetworkNode.Spec.StorageFabric == nil {
			return fmt.Errorf("networkNode %v has no storage fabric configured", existingNetworkNode.Name)
		}
		newNetworkNode.Spec.StorageFabric.VlanId = request.StorageFabricVlan
	}

	patch := client.MergeFrom(existingNetworkNode)
	err = utils.ExecuteWithRetry(
		func() error {
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Patch(ctx, newNetworkNode, patch)
		},
		k8sPatchFailedShouldRetry,
		func() error { return c.Reconnect(ctx) }, 3)
	if err != nil {
		return fmt.Errorf("k8sClient.Patch failed: %v", err)
	}

	logger.Info(fmt.Sprintf("finished updating K8s NetworkNode CR for VLAN changes, request: %+v", request))

	//////////////////////////////////
	// try to update the BGP if provided
	//////////////////////////////////
	// we got a request that wants to update the BGP community ID for the switches.
	// first check if this NetworkNode is in a NodeGroup,
	// then make sure this NodeGroup is using BGP for isolation(ie, XBX, VBX etc..).
	// Note: this is to update the switches at the node level, if there are multiple nodes trying to update BGP for the same switch with different values, the last one win.
	if request.AcceleratorFabricBGPCommunityID != 0 {
		// identify the NodeGroup for this NetworkNode
		if existingNetworkNode.Labels == nil || len(existingNetworkNode.Labels[idcnetworkv1alpha1.LabelGroupID]) == 0 {
			return fmt.Errorf("failed to update BGP, networkNode doesn't have labels, it cannot identify the NodeGroup it belongs to")
		}

		// get the NG
		nodeGroupName := existingNetworkNode.Labels[idcnetworkv1alpha1.LabelGroupID]
		nodeGroup := &idcnetworkv1alpha1.NodeGroup{}
		ngKey := types.NamespacedName{Name: nodeGroupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = c.k8sClient.Get(ctx, ngKey, nodeGroup)
		if err != nil {
			logger.Error(err, "k8sClient.Get nodeGroup failed")
			return err
		}
		if nodeGroup.Labels == nil || len(nodeGroup.Labels[idcnetworkv1alpha1.LabelPool]) == 0 {
			return fmt.Errorf("failed to update BGP, nodeGroup doesn't have labels, it cannot identify the Pool it belongs to")
		}

		// verify if it's using BGP for isolation.
		poolName := nodeGroup.Labels[idcnetworkv1alpha1.LabelPool]
		pools, err := c.ListNetworkPoolConfigs(ctx)
		if err != nil {
			return fmt.Errorf("sdn client get pool configuration failed, %v", err)
		}
		isBGPIsolation := false
		for _, pool := range pools {
			if pool.Name == poolName {
				if pool.NetworkConfigStrategy != nil &&
					pool.NetworkConfigStrategy.AcceleratorFabricStrategy != nil &&
					pool.NetworkConfigStrategy.AcceleratorFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeBGP {
					isBGPIsolation = true
				}
			}
		}
		// confirmed that this node is in a NodeGroup that using BGP for isolation, next, update the BGP Community value for all the switches in this NodeGroup
		if isBGPIsolation {
			// reuse the update nodeGroup API
			updateRequest := NodeGroupConfUpdateRequest{
				NodeGroupName: nodeGroupName,
				DesiredAcceleratorFabricConfig: &idcnetworkv1alpha1.FabricConfig{
					BGPConf: &idcnetworkv1alpha1.BGPConfig{BGPCommunity: request.AcceleratorFabricBGPCommunityID},
				},
			}
			err := c.UpdateNodeGroupConfig(ctx, updateRequest)
			if err != nil {
				return fmt.Errorf("update BGP failed, %v", err)
			}
			logger.Info(fmt.Sprintf("finished updating K8s NodeGroup CR for BGP changes, request: %+v", request))
		} else {
			return fmt.Errorf("requesting a BGP update for a non-BGP isolation fabric")
		}
	}

	return nil
}

// GetNetworkNode returns the K8S NetworkNode CR for the given name.
func (c *SDNClient) GetNetworkNode(ctx context.Context, name string) (*idcnetworkv1alpha1.NetworkNode, error) {
	logger := log.FromContext(ctx).WithName("SDNClient.GetNetworkNode").WithValues(utils.LogFieldNetworkNode, name)

	networkNode := &idcnetworkv1alpha1.NetworkNode{}
	key := types.NamespacedName{Name: name, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := utils.ExecuteWithRetry(
		func() error {
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Get(ctx, key, networkNode)
		},
		k8sGetFailedShouldRetry,
		func() error { return c.Reconnect(ctx) },
		3)
	if err != nil {
		logger.V(1).Error(err, "get NetworkNode failed")
		return nil, fmt.Errorf("get NetworkNode failed, %v", err)
	}

	return networkNode, nil
}

const (
	UpdateNotStarted = "NotStarted"
	UpdateInProgress = "UpdateInProgress"
	UpdateCompleted  = "UpdateCompleted"
)

// CheckNetworkNodeStatus checks the progress
func (c *SDNClient) CheckNetworkNodeStatus(ctx context.Context, request NetworkNodeConfStatusCheckRequest) (NetworkNodeConfStatusCheckResponse, error) {
	logger := log.FromContext(ctx).WithName("SDNClient.CheckNetworkNodeStatus").WithValues(utils.LogFieldNetworkNode, request.NetworkNodeName)
	var err error
	result := NetworkNodeConfStatusCheckResponse{}
	if len(request.NetworkNodeName) == 0 {
		return result, fmt.Errorf("NetworkNodeName is not provided")
	}

	mConfig, err := c.GetNetworkManagerConfig(ctx)
	if err != nil {
		return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("sdn client get manager configuration failed, %v", err)
	}
	var allowedVlanIds []int
	var allowedNativeVlanIds []int

	// get the allowedVlanIds
	allowedVlanIdsStr := utils.DefaultAllowedVlanIdsStr
	if len(mConfig.ControllerConfig.AllowedVlanIds) > 0 {
		allowedVlanIdsStr = mConfig.ControllerConfig.AllowedVlanIds
	}
	allowedVlanIds, err = utils.ExpandVlanRanges(allowedVlanIdsStr)
	if err != nil {
		return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("Error expanding valid VLAN range, %v", err)
	}

	// get the allowedNativeVlanIds
	allowedNativeVlanIdsStr := utils.DefaultAllowedNativeVlanIdsStr
	if len(mConfig.ControllerConfig.AllowedNativeVlanIds) > 0 {
		allowedNativeVlanIdsStr = mConfig.ControllerConfig.AllowedNativeVlanIds
	}
	allowedNativeVlanIds, err = utils.ExpandVlanRanges(allowedNativeVlanIdsStr)
	if err != nil {
		return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("Error expanding valid VLAN range, %v", err)
	}

	if request.DesiredFrontEndFabricVlan != 0 {
		frontEndErr := utils.ValidateVlanValue(int(request.DesiredFrontEndFabricVlan), allowedVlanIds)
		if frontEndErr != nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("desiredFrontEndFabricVlan is invalid, %v", frontEndErr)
		}
	}
	if len(request.DesiredFrontEndFabricMode) != 0 {
		frontEndModeErr := utils.ValidateModeValue(request.DesiredFrontEndFabricMode, mConfig.ControllerConfig.AllowedModes)
		if frontEndModeErr != nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("desiredFrontEndFabricMode is invalid, %v", frontEndModeErr)
		}
	}
	if request.DesiredFrontEndFabricTrunkGroups != nil {
		frontEndTrunkGroupsErr := utils.ValidateTrunkGroups([]string(request.DesiredFrontEndFabricTrunkGroups), nil)
		if frontEndTrunkGroupsErr != nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("desiredFrontEndFabricTrunkGroups is invalid, %v", frontEndTrunkGroupsErr)
		}
	}
	if request.DesiredFrontEndFabricNativeVlan != 0 {
		frontEndNativeVlanErr := utils.ValidateVlanValue(int(request.DesiredFrontEndFabricNativeVlan), allowedNativeVlanIds)
		if frontEndNativeVlanErr != nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("desiredFrontEndFabricNativeVlan is invalid, %v", frontEndNativeVlanErr)
		}
	}
	if request.DesiredAcceleratorFabricVlan != 0 {
		accelErr := utils.ValidateVlanValue(int(request.DesiredAcceleratorFabricVlan), allowedVlanIds)
		if accelErr != nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("desiredAcceleratorFabricVlan is invalid, %v", accelErr)
		}
	}
	// request.StorageFabricVlan is set
	if request.DesiredStorageFabricVlan != 0 {
		storageErr := utils.ValidateVlanValue(int(request.DesiredStorageFabricVlan), allowedVlanIds)
		if storageErr != nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("DesiredStorageFabricVlan is invalid, %v", storageErr)
		}
	}

	// request.DesiredAcceleratorFabricBGPCommunityID is set
	if request.DesiredAcceleratorFabricBGPCommunityID != 0 {
		accBGPErr := utils.ValidateBGPCommunityValue(int32(request.DesiredAcceleratorFabricBGPCommunityID))
		if accBGPErr != nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("request.ValidateBGPCommunityValue is invalid, %v", accBGPErr)
		}
	}

	// get the networkNode
	networkNode := &idcnetworkv1alpha1.NetworkNode{}
	key := types.NamespacedName{Name: request.NetworkNodeName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = utils.ExecuteWithRetry(
		func() error {
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Get(ctx, key, networkNode)
		},
		k8sGetFailedShouldRetry,
		func() error { return c.Reconnect(ctx) },
		3)
	if err != nil {
		return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("get NetworkNode failed, %v", err)
	}

	//////////////////////////////////
	// check Specs
	//////////////////////////////////
	// check the Specs. As long as one of the item is in the desired state, we mark the whole request as UpdateNotStarted.
	// check the front-end fabric's Spec
	if request.DesiredFrontEndFabricVlan != 0 {
		if networkNode.Spec.FrontEndFabric == nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("networkNode %v has no frontend fabric configured", networkNode.Name)
		}

		// check vlan
		if request.DesiredFrontEndFabricVlan != networkNode.Spec.FrontEndFabric.VlanId {
			result.Status = UpdateNotStarted
			logger.Info(fmt.Sprintf("frontend vlan update NOT started, desired: [%v], current: [%v]", request.DesiredFrontEndFabricVlan, networkNode.Spec.FrontEndFabric.VlanId))
			return result, nil
		}
	}
	// check frontend mode Spec
	if len(request.DesiredFrontEndFabricMode) != 0 {
		if networkNode.Spec.FrontEndFabric == nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("networkNode %v has no frontend fabric configured", networkNode.Name)
		}
		if request.DesiredFrontEndFabricMode != networkNode.Spec.FrontEndFabric.Mode {
			result.Status = UpdateNotStarted
			logger.Info(fmt.Sprintf("frontend mode update NOT started, desired: [%v], current: [%v]", request.DesiredFrontEndFabricMode, networkNode.Spec.FrontEndFabric.Mode))
			return result, nil
		}
	}

	// check frontend native vlan Spec
	if request.DesiredFrontEndFabricNativeVlan != 0 {
		if networkNode.Spec.FrontEndFabric == nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("networkNode %v has no frontend fabric configured", networkNode.Name)
		}
		if request.DesiredFrontEndFabricNativeVlan != networkNode.Spec.FrontEndFabric.NativeVlan {
			result.Status = UpdateNotStarted
			logger.Info(fmt.Sprintf("frontend native vlan update NOT started, desired: [%v], current: [%v]", request.DesiredFrontEndFabricNativeVlan, networkNode.Spec.FrontEndFabric.NativeVlan))
			return result, nil
		}
	}

	// check frontend trunk groups Spec
	requestedTG := request.DesiredFrontEndFabricTrunkGroups
	// note: only when request.DesiredFrontEndFabricTrunkGroups is NOT nil, we check the status.
	// For DesiredFrontEndFabricTrunkGroups is nil or DesiredFrontEndFabricTrunkGroups field not provided in the request, we will ignore the status check.
	if requestedTG != nil {
		var nnTG []string
		if networkNode.Spec.FrontEndFabric.TrunkGroups != nil {
			nnTG = *networkNode.Spec.FrontEndFabric.TrunkGroups
		}
		sort.Strings(requestedTG)
		sort.Strings(nnTG)
		if !reflect.DeepEqual(requestedTG, nnTG) {
			result.Status = UpdateNotStarted
			logger.Info(fmt.Sprintf("frontend trunk groups update NOT started, desired: [%v], current: [%v]", requestedTG, nnTG))
			return result, nil
		}
	}

	// check the accelerator fabric's Spec
	if request.DesiredAcceleratorFabricVlan != 0 {
		if networkNode.Spec.AcceleratorFabric == nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("networkNode %v has no accelerator fabric configured", networkNode.Name)
		}

		if request.DesiredAcceleratorFabricVlan != networkNode.Spec.AcceleratorFabric.VlanId {
			result.Status = UpdateNotStarted
			logger.Info(fmt.Sprintf("acc trunk vlan update NOT started, desired: [%v], current: [%v]", request.DesiredAcceleratorFabricVlan, networkNode.Spec.AcceleratorFabric.VlanId))
			return result, nil
		}
	}

	// check the storage fabric's Spec
	if request.DesiredStorageFabricVlan != 0 {
		if networkNode.Spec.StorageFabric == nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("networkNode %v has no storage fabric configured", networkNode.Name)
		}

		if request.DesiredStorageFabricVlan != networkNode.Spec.StorageFabric.VlanId {
			result.Status = UpdateNotStarted
			logger.Info(fmt.Sprintf("storage trunk vlan update NOT started, desired: [%v], current: [%v]", request.DesiredStorageFabricVlan, networkNode.Spec.StorageFabric.VlanId))
			return result, nil
		}
	}

	//////////////////////////////////
	// check Status
	//////////////////////////////////
	result.Status = UpdateCompleted
	// check front-end fabric status
	if request.DesiredFrontEndFabricVlan != 0 {
		if request.DesiredFrontEndFabricVlan != networkNode.Status.FrontEndFabricStatus.LastObservedVlanId {
			result.Status = UpdateInProgress
			logger.Info(fmt.Sprintf("frontend vlan update is still in progress, desired: [%v], current: [%v]", request.DesiredFrontEndFabricVlan, networkNode.Status.FrontEndFabricStatus.LastObservedVlanId))
			// We can't return UpdateInProgress here, as we haven't check the BGP progress yet.
			// return result, nil
		}
	}
	// check front-end mode
	if len(request.DesiredFrontEndFabricMode) != 0 {
		if request.DesiredFrontEndFabricMode != networkNode.Status.FrontEndFabricStatus.LastObservedMode {
			result.Status = UpdateInProgress
			logger.Info(fmt.Sprintf("frontend mode update is still in progress, desired: [%v], current: [%v]", request.DesiredFrontEndFabricMode, networkNode.Status.FrontEndFabricStatus.LastObservedMode))
		}
	}
	// check front-end fabric native vlan
	if request.DesiredFrontEndFabricNativeVlan != 0 {
		if request.DesiredFrontEndFabricNativeVlan != networkNode.Status.FrontEndFabricStatus.LastObservedNativeVlan {
			result.Status = UpdateInProgress
			logger.Info(fmt.Sprintf("frontend native vlan update is still in progress, desired: [%v], current: [%v]", request.DesiredFrontEndFabricNativeVlan, networkNode.Status.FrontEndFabricStatus.LastObservedNativeVlan))
		}
	}

	// check front-end fabric trunk groups
	nnStatusTG := networkNode.Status.FrontEndFabricStatus.LastObservedTrunkGroups
	if requestedTG != nil {
		sort.Strings(requestedTG)
		sort.Strings(nnStatusTG)
		if !slices.Equal(requestedTG, nnStatusTG) {
			result.Status = UpdateInProgress
			logger.Info(fmt.Sprintf("frontend trunk group update is still in progress, desired: [%v], current: [%v]", requestedTG, nnStatusTG))
		}
	}

	// check accelerator fabric status
	if request.DesiredAcceleratorFabricVlan != 0 {
		for _, portInfo := range networkNode.Status.AcceleratorFabricStatus.SwitchPorts {
			// as long as there is 1 port is not ready, the whole NetworkNode is considered as not ready.
			if request.DesiredAcceleratorFabricVlan != portInfo.LastObservedVlanId {
				result.Status = UpdateInProgress
				// return result, nil
				logger.Info("acc vlan update is still in progress")
			}
		}
	}

	// check storage fabric status
	if request.DesiredStorageFabricVlan != 0 {
		for _, portInfo := range networkNode.Status.StorageFabricStatus.SwitchPorts {
			// as long as there is 1 port is not ready, the whole NetworkNode is considered as not ready.
			if request.DesiredStorageFabricVlan != portInfo.LastObservedVlanId {
				result.Status = UpdateInProgress
				logger.Info("storage vlan update is still in progress")
				// return result, nil
			}
		}
	}

	//////////////////////////////////
	// check BGP
	//////////////////////////////////
	if request.DesiredAcceleratorFabricBGPCommunityID != 0 {
		// identify the NodeGroup for this NetworkNode
		if networkNode.Labels == nil || len(networkNode.Labels[idcnetworkv1alpha1.LabelGroupID]) == 0 {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("failed to update BGP, networkNode doesn't have labels, it cannot identify the NodeGroup it belongs to")
		}

		// get the NG
		nodeGroupName := networkNode.Labels[idcnetworkv1alpha1.LabelGroupID]
		nodeGroup := &idcnetworkv1alpha1.NodeGroup{}
		ngKey := types.NamespacedName{Name: nodeGroupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = c.k8sClient.Get(ctx, ngKey, nodeGroup)
		if err != nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("k8sClient.Get nodeGroup failed, %v", err)
		}
		if nodeGroup.Labels == nil || len(nodeGroup.Labels[idcnetworkv1alpha1.LabelPool]) == 0 {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("failed to update BGP, nodeGroup doesn't have labels, it cannot identify the Pool it belongs to")
		}

		// verify if it's using BGP for isolation.
		poolName := nodeGroup.Labels[idcnetworkv1alpha1.LabelPool]
		pools, err := c.ListNetworkPoolConfigs(ctx)
		if err != nil {
			return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("sdn client get pool configuration failed, %v", err)
		}
		isBGPIsolation := false
		for _, pool := range pools {
			if pool.Name == poolName {
				if pool.NetworkConfigStrategy != nil &&
					pool.NetworkConfigStrategy.AcceleratorFabricStrategy != nil &&
					pool.NetworkConfigStrategy.AcceleratorFabricStrategy.IsolationType == idcnetworkv1alpha1.IsolationTypeBGP {
					isBGPIsolation = true
				}
			}
		}
		// confirmed that this node is in a NodeGroup that using BGP for isolation, next, update the BGP Community value for all the switches in this NodeGroup
		if isBGPIsolation {
			// reuse the update nodeGroup API
			updateRequest := NodeGroupConfUpdateRequest{
				NodeGroupName: nodeGroupName,
				DesiredAcceleratorFabricConfig: &idcnetworkv1alpha1.FabricConfig{
					BGPConf: &idcnetworkv1alpha1.BGPConfig{BGPCommunity: request.DesiredAcceleratorFabricBGPCommunityID},
				},
			}
			checkNGResult, err := c.CheckNodeGroupStatus(ctx, CheckNodeGroupStatusRequest(updateRequest))
			if err != nil {
				return NetworkNodeConfStatusCheckResponse{}, fmt.Errorf("CheckNodeGroupConfigStatus error: %v", err)
			}

			// if the BGP is not started, return UpdateNotStarted directly
			if checkNGResult.Status == UpdateNotStarted {
				result.Status = UpdateNotStarted
				return result, nil
			}
			// if BGP is in progress, it's ok to overwrite the overall status to UpdateInProgress
			if checkNGResult.Status == UpdateInProgress {
				result.Status = UpdateInProgress
			}
			// if BGP is UpdateCompleted, we don't need to update the overall status, as it defaults to UpdateCompleted
		} else {
			return result, fmt.Errorf("requesting a BGP status check for a non-BGP isolation fabric")
		}
	}

	logger.V(1).Info(fmt.Sprintf("finished checking K8s NetworkNode update request status,  request: %+v, result: %v,", request, result.Status))
	return result, nil
}

type NodeGroupConfUpdateRequest struct {
	NodeGroupName                  string
	DesiredFrontEndFabricConfig    *idcnetworkv1alpha1.FabricConfig
	DesiredAcceleratorFabricConfig *idcnetworkv1alpha1.FabricConfig
	DesiredStorageFabricConfig     *idcnetworkv1alpha1.FabricConfig
}

type CheckNodeGroupStatusRequest NodeGroupConfUpdateRequest

type NodeGroupConfStatusCheckResponse struct {
	NodeGroup idcnetworkv1alpha1.NodeGroup
	Status    string
}

var ErrK8sClientNil = fmt.Errorf("k8s client is nil")

// UpdateNodeGroupConfig update the front-end and accelerator VLAN/BGP for a NodeGroup.
// Note: SDNClient is NOT responsible for checking if a field should be update or not. For example, SDNClient cannot prevent a caller updating the accelerator vlan for a non-Gaudi node.
func (c *SDNClient) UpdateNodeGroupConfig(ctx context.Context, request NodeGroupConfUpdateRequest) error {
	logger := log.FromContext(ctx).WithName("SDNClient.UpdateNodeGroupConfig").WithValues(utils.LogFieldNodeGroup, request.NodeGroupName)
	var err error
	var allowedVlanIds []int

	mConfig, err := c.GetNetworkManagerConfig(ctx)
	if err != nil {
		return fmt.Errorf("sdn client get manager configuration failed, %v", err)
	}

	// get the allowedVlanIds
	allowedVlanIdsStr := utils.DefaultAllowedVlanIdsStr
	if len(mConfig.ControllerConfig.AllowedVlanIds) > 0 {
		allowedVlanIdsStr = mConfig.ControllerConfig.AllowedVlanIds
	}
	allowedVlanIds, err = utils.ExpandVlanRanges(allowedVlanIdsStr)
	if err != nil {
		return fmt.Errorf("Error expanding valid VLAN range, %v", err)
	}

	if len(request.NodeGroupName) == 0 {
		return fmt.Errorf("NodeGroupName is not provided")
	}

	if err := ValidateNodeGroupRequest(request, allowedVlanIds); err != nil {
		return err
	}

	// perform Patch update
	existingNodeGroup := &idcnetworkv1alpha1.NodeGroup{}
	key := types.NamespacedName{Name: request.NodeGroupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = utils.ExecuteWithRetry(
		func() error {
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Get(ctx, key, existingNodeGroup)
		},
		k8sGetFailedShouldRetry,
		func() error { return c.Reconnect(ctx) }, 3)
	if err != nil {
		return fmt.Errorf("k8sClient.Get NodeGroup failed, %v", err)
	}

	// check if the NG is in maintenance
	if existingNodeGroup.Labels != nil && len(existingNodeGroup.Labels[idcnetworkv1alpha1.LabelMaintenance]) > 0 {
		return fmt.Errorf("nodeGroup [%v] is under maintenance, please try again later", existingNodeGroup.Name)
	}

	newNodeGroup := existingNodeGroup.DeepCopy()

	if request.DesiredFrontEndFabricConfig != nil {

		if newNodeGroup.Spec.FrontEndFabricConfig == nil {
			return fmt.Errorf("nodeGroup %v has no frontend fabric configured", newNodeGroup.Name)
		}

		if request.DesiredFrontEndFabricConfig.VlanConf != nil {
			newNodeGroup.Spec.FrontEndFabricConfig.VlanConf = request.DesiredFrontEndFabricConfig.VlanConf
		}
		if request.DesiredFrontEndFabricConfig.BGPConf != nil {
			newNodeGroup.Spec.FrontEndFabricConfig.BGPConf = request.DesiredFrontEndFabricConfig.BGPConf
		}
	}

	if request.DesiredAcceleratorFabricConfig != nil {

		if newNodeGroup.Spec.AcceleratorFabricConfig == nil {
			return fmt.Errorf("nodeGroup %v has no accelerator fabric configured", newNodeGroup.Name)
		}

		if request.DesiredAcceleratorFabricConfig.VlanConf != nil {
			newNodeGroup.Spec.AcceleratorFabricConfig.VlanConf = request.DesiredAcceleratorFabricConfig.VlanConf
		}
		if request.DesiredAcceleratorFabricConfig.BGPConf != nil {
			newNodeGroup.Spec.AcceleratorFabricConfig.BGPConf = request.DesiredAcceleratorFabricConfig.BGPConf
		}
	}

	if request.DesiredStorageFabricConfig != nil {

		if newNodeGroup.Spec.StorageFabricConfig == nil {
			return fmt.Errorf("nodeGroup %v has no storage fabric configured", newNodeGroup.Name)
		}

		if request.DesiredStorageFabricConfig.VlanConf != nil {
			newNodeGroup.Spec.StorageFabricConfig.VlanConf = request.DesiredStorageFabricConfig.VlanConf
		}
		if request.DesiredStorageFabricConfig.BGPConf != nil {
			newNodeGroup.Spec.StorageFabricConfig.BGPConf = request.DesiredStorageFabricConfig.BGPConf
		}
	}

	patch := client.MergeFrom(existingNodeGroup)
	err = utils.ExecuteWithRetry(
		func() error {
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Patch(ctx, newNodeGroup, patch)
		},
		k8sPatchFailedShouldRetry,
		func() error { return c.Reconnect(ctx) }, 3)
	if err != nil {
		return fmt.Errorf("k8sClient.Patch NodeGroup failed: %v", err)
	}

	logger.Info("update K8s NodeGroup success", utils.LogFieldNodeGroupUpdateRequest, request)
	return nil
}

// CheckNodeGroupStatus check if a NodeGroup's current network config meet the desired values.
// The passed request should be identical to the one already passed to UpdateNodeGroupConfig, and is used to check the progress / status of that request.
func (c *SDNClient) CheckNodeGroupStatus(ctx context.Context, request CheckNodeGroupStatusRequest) (NodeGroupConfStatusCheckResponse, error) {
	logger := log.FromContext(ctx).WithName("SDNClient.CheckNodeGroupStatus").WithValues(utils.LogFieldNodeGroup, request.NodeGroupName)

	var allowedVlanIds []int

	mConfig, err := c.GetNetworkManagerConfig(ctx)
	if err != nil {
		return NodeGroupConfStatusCheckResponse{}, fmt.Errorf("sdn client get manager configuration failed, %v", err)
	}

	// get the allowedVlanIds
	allowedVlanIdsStr := utils.DefaultAllowedVlanIdsStr
	if len(mConfig.ControllerConfig.AllowedVlanIds) > 0 {
		allowedVlanIdsStr = mConfig.ControllerConfig.AllowedVlanIds
	}
	allowedVlanIds, err = utils.ExpandVlanRanges(allowedVlanIdsStr)
	if err != nil {
		return NodeGroupConfStatusCheckResponse{}, fmt.Errorf("Error expanding valid VLAN range, %v", err)
	}

	result := NodeGroupConfStatusCheckResponse{}
	if len(request.NodeGroupName) == 0 {
		return result, fmt.Errorf("NodeGroupName is not provided")
	}

	if err = ValidateNodeGroupRequest(NodeGroupConfUpdateRequest(request), allowedVlanIds); err != nil {
		return result, err
	}

	// get the NodeGroup
	nodeGroup := &idcnetworkv1alpha1.NodeGroup{}
	key := types.NamespacedName{Name: request.NodeGroupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = utils.ExecuteWithRetry(
		func() error {
			if c.k8sClient == nil {
				return ErrK8sClientNil
			}
			return c.k8sClient.Get(ctx, key, nodeGroup)
		},
		k8sGetFailedShouldRetry,
		func() error { return c.Reconnect(ctx) }, 3)
	if err != nil {
		return result, fmt.Errorf("k8sClient.Get NodeGroup failed: %v", err)
	}

	// check to see if it's started updating children yet
	if request.DesiredFrontEndFabricConfig != nil {
		if request.DesiredFrontEndFabricConfig.VlanConf != nil {
			if nodeGroup.Spec.FrontEndFabricConfig == nil ||
				nodeGroup.Spec.FrontEndFabricConfig.VlanConf.VlanID != request.DesiredFrontEndFabricConfig.VlanConf.VlanID {
				result.Status = UpdateNotStarted
				return result, nil
			}
		}
		if request.DesiredFrontEndFabricConfig.BGPConf != nil {
			if nodeGroup.Spec.FrontEndFabricConfig == nil ||
				nodeGroup.Spec.FrontEndFabricConfig.BGPConf.BGPCommunity != request.DesiredFrontEndFabricConfig.BGPConf.BGPCommunity {
				result.Status = UpdateNotStarted
				return result, nil
			}
		}
	}
	if request.DesiredAcceleratorFabricConfig != nil {
		if request.DesiredAcceleratorFabricConfig.VlanConf != nil {
			if nodeGroup.Spec.AcceleratorFabricConfig == nil ||
				nodeGroup.Spec.AcceleratorFabricConfig.VlanConf.VlanID != request.DesiredAcceleratorFabricConfig.VlanConf.VlanID {
				result.Status = UpdateNotStarted
				return result, nil
			}
		}
		if request.DesiredAcceleratorFabricConfig.BGPConf != nil {
			if nodeGroup.Spec.AcceleratorFabricConfig == nil ||
				nodeGroup.Spec.AcceleratorFabricConfig.BGPConf.BGPCommunity != request.DesiredAcceleratorFabricConfig.BGPConf.BGPCommunity {
				result.Status = UpdateNotStarted
				return result, nil
			}
		}
	}
	if request.DesiredStorageFabricConfig != nil {
		if request.DesiredStorageFabricConfig.VlanConf != nil {
			if nodeGroup.Spec.StorageFabricConfig == nil ||
				nodeGroup.Spec.StorageFabricConfig.VlanConf.VlanID != request.DesiredStorageFabricConfig.VlanConf.VlanID {
				result.Status = UpdateNotStarted
				return result, nil
			}
		}
		if request.DesiredStorageFabricConfig.BGPConf != nil {
			if nodeGroup.Spec.StorageFabricConfig == nil ||
				nodeGroup.Spec.StorageFabricConfig.BGPConf.BGPCommunity != request.DesiredStorageFabricConfig.BGPConf.BGPCommunity {
				result.Status = UpdateNotStarted
				return result, nil
			}
		}

		// if !reflect.DeepEqual(request.DesiredStorageFabricConfig, nodeGroup.Spec.StorageFabricConfig) {
		// 	result.Status = UpdateNotStarted
		// 	return result, nil
		// }
	}

	// check front-end fabric status
	if request.DesiredFrontEndFabricConfig != nil {
		if request.DesiredFrontEndFabricConfig.VlanConf != nil {
			// If any of the child NetworkNodes is not ready, update is in progress.
			if nodeGroup.Status.FrontEndFabricStatus == nil ||
				!nodeGroup.Status.FrontEndFabricStatus.VlanConfigStatus.Ready ||
				nodeGroup.Status.FrontEndFabricStatus.VlanConfigStatus.LastObservedReadyVLAN != request.DesiredFrontEndFabricConfig.VlanConf.VlanID {
				result.Status = UpdateInProgress
				return result, nil
			}
		}
		if request.DesiredFrontEndFabricConfig.BGPConf != nil {
			// If any of the child NetworkNodes is not ready, update is in progress.
			if nodeGroup.Status.FrontEndFabricStatus == nil ||
				!nodeGroup.Status.FrontEndFabricStatus.BGPConfigStatus.Ready ||
				nodeGroup.Status.FrontEndFabricStatus.BGPConfigStatus.LastObservedReadyBGP != request.DesiredFrontEndFabricConfig.BGPConf.BGPCommunity {
				result.Status = UpdateInProgress
				return result, nil
			}
		}
	}

	// check accelerator fabric status
	if request.DesiredAcceleratorFabricConfig != nil {
		if request.DesiredAcceleratorFabricConfig.VlanConf != nil {
			// If any of the child NetworkNodes is not ready, update is in progress.
			if nodeGroup.Status.AcceleratorFabricStatus == nil ||
				!nodeGroup.Status.AcceleratorFabricStatus.VlanConfigStatus.Ready ||
				nodeGroup.Status.AcceleratorFabricStatus.VlanConfigStatus.LastObservedReadyVLAN != request.DesiredAcceleratorFabricConfig.VlanConf.VlanID {
				result.Status = UpdateInProgress
				return result, nil
			}
		}
		if request.DesiredAcceleratorFabricConfig.BGPConf != nil {
			// If any of the child NetworkNodes is not ready, update is in progress.
			if nodeGroup.Status.AcceleratorFabricStatus == nil ||
				!nodeGroup.Status.AcceleratorFabricStatus.BGPConfigStatus.Ready ||
				nodeGroup.Status.AcceleratorFabricStatus.BGPConfigStatus.LastObservedReadyBGP != request.DesiredAcceleratorFabricConfig.BGPConf.BGPCommunity {
				result.Status = UpdateInProgress
				return result, nil
			}
		}
	}

	// check storage fabric status
	if request.DesiredStorageFabricConfig != nil {
		if request.DesiredStorageFabricConfig.VlanConf != nil {
			// If any of the child NetworkNodes is not ready, update is in progress.
			if nodeGroup.Status.StorageFabricStatus == nil ||
				!nodeGroup.Status.StorageFabricStatus.VlanConfigStatus.Ready ||
				nodeGroup.Status.StorageFabricStatus.VlanConfigStatus.LastObservedReadyVLAN != request.DesiredStorageFabricConfig.VlanConf.VlanID {
				result.Status = UpdateInProgress
				return result, nil
			}
		}
		if request.DesiredStorageFabricConfig.BGPConf != nil {
			// If any of the child NetworkNodes is not ready, update is in progress.
			if nodeGroup.Status.StorageFabricStatus == nil ||
				!nodeGroup.Status.StorageFabricStatus.BGPConfigStatus.Ready ||
				nodeGroup.Status.StorageFabricStatus.BGPConfigStatus.LastObservedReadyBGP != request.DesiredStorageFabricConfig.BGPConf.BGPCommunity {
				result.Status = UpdateInProgress
				return result, nil
			}
		}
	}

	logger.Info("check K8s NodeGroup success", utils.LogFieldNodeGroupUpdateRequest, request)
	// all fabrics are in-sync
	result.Status = UpdateCompleted
	return result, nil
}

func ValidateNodeGroupRequest(request NodeGroupConfUpdateRequest, allowedVlanIds []int) error {
	if request.DesiredFrontEndFabricConfig != nil {
		if err := ValidateFabricConfig(*request.DesiredFrontEndFabricConfig, allowedVlanIds); err != nil {
			return err
		}
	}
	if request.DesiredAcceleratorFabricConfig != nil {
		if err := ValidateFabricConfig(*request.DesiredAcceleratorFabricConfig, allowedVlanIds); err != nil {
			return err
		}
	}
	if request.DesiredStorageFabricConfig != nil {
		if err := ValidateFabricConfig(*request.DesiredStorageFabricConfig, allowedVlanIds); err != nil {
			return err
		}
	}
	return nil
}

func ValidateFabricConfig(fabricConfig idcnetworkv1alpha1.FabricConfig, allowedVlanIds []int) error {
	if fabricConfig.VlanConf != nil {
		err := utils.ValidateVlanValue(int(fabricConfig.VlanConf.VlanID), allowedVlanIds)
		if err != nil {
			return err
		}
	}
	if fabricConfig.BGPConf != nil {
		err := utils.ValidateBGPCommunityValue(int32(fabricConfig.BGPConf.BGPCommunity))
		if err != nil {
			return err
		}
	}

	// TODO: Validation of request with pool name.
	// Will be done by a webhook on the server-side.

	return nil
}

// TODO: comment this API out for now, as it's not used by anyone, and we need improvement on it.
// If we want to keep this API, then we need to first check what data source is used for the mapping, and then go fetch it.
// func (c *SDNClient) GetGroupToPoolMapping(ctx context.Context) (map[string]string, error) {
// 	logger := log.FromContext(ctx).WithName("SDNClient.GetGroupPoolMapping")

// 	configMap := &corev1.ConfigMap{}
// 	key := types.NamespacedName{Name: "group-pool-mapping-config", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
// 	err := c.k8sClient.Get(ctx, key, configMap)
// 	if err != nil {
// 		logger.Error(err, "Failed to get ConfigMap")
// 	}
// 	nodeGroupToPoolMapping := &idcnetworkv1alpha1.NodeGroupToPoolMap{}
// 	err = json.Unmarshal([]byte(configMap.Data["group_pool_mapping_config.json"]), nodeGroupToPoolMapping)
// 	if err != nil {
// 		logger.Error(err, "Failed to Unmarshal group_pool_mapping_config")
// 		return nil, err
// 	}

// 	return nodeGroupToPoolMapping.NodeGroupToPoolMap, nil
// }

func (c *SDNClient) ListNetworkPoolConfigs(ctx context.Context) ([]*idcnetworkv1alpha1.Pool, error) {
	logger := log.FromContext(ctx).WithName("SDNClient.ListNetworkPoolConfigs")

	configMap := &corev1.ConfigMap{}
	key := types.NamespacedName{Name: "sdn-pool-config", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := c.k8sClient.Get(ctx, key, configMap)
	if err != nil {
		logger.Error(err, "Failed to get sdn-pool-config config map")
	}
	poolConfigs := &idcnetworkv1alpha1.PoolList{}
	err = json.Unmarshal([]byte(configMap.Data["pool_config.json"]), poolConfigs)
	if err != nil {
		logger.Error(err, "Failed to Unmarshal pool_config.json")
		return nil, err
	}

	return poolConfigs.Items, nil
}

func (c *SDNClient) GetNetworkManagerConfig(ctx context.Context) (*idcnetworkv1alpha1.SDNControllerConfig, error) {
	logger := log.FromContext(ctx).WithName("SDNClient.GetNetworkManagerConfig")

	configMap := &corev1.ConfigMap{}
	key := types.NamespacedName{Name: "sdn-controller-manager-config", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := c.k8sClient.Get(ctx, key, configMap)
	if err != nil {
		logger.Error(err, "Failed to get sdn-controller-manager-config config map")
		return nil, err
	}

	managerConfig := &idcnetworkv1alpha1.SDNControllerConfig{}
	if configData, ok := configMap.Data["controller_manager_config.yaml"]; ok {
		jsonData, err := yaml.YAMLToJSON([]byte(configData))
		if err != nil {
			logger.Error(err, "Failed to convert YAML to JSON")
			return nil, err
		}
		err = json.Unmarshal([]byte(jsonData), managerConfig)
		if err != nil {
			logger.Error(err, "Failed to Unmarshal manager_config.json")
			return nil, err
		}
	} else {
		err = fmt.Errorf("manager_config.json not found in ConfigMap")
		logger.Error(err, "")
		return nil, err
	}
	return managerConfig, nil
}

// func (c *SDNClient) GetPoolConfigForNodeGroup(ctx context.Context, ng string) (*idcnetworkv1alpha1.Pool, error) {
// 	logger := log.FromContext(ctx).WithName("SDNClient.ListNetworkPoolConfigs")

// 	configMap := &corev1.ConfigMap{}
// 	key := types.NamespacedName{Name: "sdn-pool-config", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
// 	err := c.k8sClient.Get(ctx, key, configMap)
// 	if err != nil {
// 		logger.Error(err, "Failed to get sdn-pool-config config map")
// 	}
// 	poolConfigs := &idcnetworkv1alpha1.PoolList{}
// 	err = json.Unmarshal([]byte(configMap.Data["pool_config.json"]), poolConfigs)
// 	if err != nil {
// 		logger.Error(err, "Failed to Unmarshal pool_config.json")
// 		return nil, err
// 	}

// 	return poolConfigs.Items, nil
// }
