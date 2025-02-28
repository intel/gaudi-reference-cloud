// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controllers

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	devicesmanager "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/devices_manager"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/pools"
	"github.com/prometheus/client_golang/prometheus"

	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bmenrollment "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	metal3ClientSet "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/clientset/versioned"
	metal3Informerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/informers/externalversions"
	metal3v1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/metal3client/listers/metal3.io/v1alpha1"
)

type BMHControllerManager struct {
	sync.Mutex

	controllers         map[string]BMHControllerIF // key: kubeconfig context name, val: controller
	failed              map[string]struct{}        // key: kubeconfig context name, val: empty struct
	failedRetryInterval time.Duration

	// cache the kubeconfig path for each controller, this will be used for recreating the controller. key: kubeconfig context name, val: kubeconfig path
	controllerKubeConfs map[string]string

	// common configurations for each controller.
	bmhConf               BMHControllerConf
	eventRecorder         record.EventRecorder
	newControllerFunc     func(ctx context.Context, bmhConf BMHControllerConf, name string, metal3Kubeconfig string) (BMHControllerIF, error)
	controllerNameGenFunc func(input string) string
}

func NewBMHControllerManager(
	bmhConf BMHControllerConf,
	newCtrlFunc func(ctx context.Context, bmhConf BMHControllerConf, name string, metal3Kubeconfig string) (BMHControllerIF, error),
	controllerNameGenFunc func(input string) string,
	metal3KubeconfigPaths string,
) *BMHControllerManager {
	// extract the metal3 kubeconfigs and cache them
	metal3KubeconfigFilePaths := strings.Split(metal3KubeconfigPaths, ";")
	controllerKubeConfs := make(map[string]string)
	for _, kubeconfigFilePath := range metal3KubeconfigFilePaths {
		controllerKubeConfs[controllerNameGenFunc(kubeconfigFilePath)] = kubeconfigFilePath
	}

	controllerMgr := &BMHControllerManager{
		controllers:           make(map[string]BMHControllerIF),
		failed:                make(map[string]struct{}),
		failedRetryInterval:   bmhConf.BMHControllerCreationRetryInterval,
		controllerKubeConfs:   controllerKubeConfs,
		bmhConf:               bmhConf,
		eventRecorder:         bmhConf.EventRecorder,
		newControllerFunc:     newCtrlFunc,
		controllerNameGenFunc: controllerNameGenFunc,
	}
	return controllerMgr
}

func (m *BMHControllerManager) GetController(name string) BMHControllerIF {
	m.Lock()
	defer m.Unlock()
	return m.controllers[name]
}

func (m *BMHControllerManager) GetFailedControllers() map[string]struct{} {
	m.Lock()
	defer m.Unlock()
	return m.failed
}

func (m *BMHControllerManager) GetRunnableControllers() map[string]BMHControllerIF {
	m.Lock()
	defer m.Unlock()
	return m.controllers
}

// CreateControllers create BMH Controllers
func (m *BMHControllerManager) CreateAllControllers(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("BMHManager.CreateControllers")
	m.Lock()
	// get all the metal3 servers kubeconfigs
	metal3KubeconfigFilePaths := m.controllerKubeConfs
	m.Unlock()

	for _, metal3Kubeconfig := range metal3KubeconfigFilePaths {
		err := m.createController(ctx, metal3Kubeconfig)
		if err != nil {
			logger.Info(fmt.Sprintf("create BMH Controller failed for kubeconfig file %s, reason: %s, will try to recreate it again later.", metal3Kubeconfig, err.Error()))
			objectRef := &corev1.ObjectReference{
				Kind:       "Namespace",
				Name:       idcnetworkv1alpha1.SDNControllerNamespace,
				Namespace:  idcnetworkv1alpha1.SDNControllerNamespace,
				APIVersion: "v1",
			}
			// create event
			m.eventRecorder.Event(objectRef, corev1.EventTypeWarning, "Unable to create BMH Controller", err.Error())
			// create the metrics when failed to create a BMH Controller that talks to a Metal3 server
			metrics.BMHControllerManagerErrors.With(prometheus.Labels{
				metrics.MetricsLabelErrorType:          metrics.ErrorTypeCreateBMHControllerFailed,
				metrics.MetricsLabelKubeconfigFilePath: metal3Kubeconfig,
			}).Set(1)
		} else {
			metrics.BMHControllerManagerErrors.With(prometheus.Labels{
				metrics.MetricsLabelErrorType:          metrics.ErrorTypeCreateBMHControllerFailed,
				metrics.MetricsLabelKubeconfigFilePath: metal3Kubeconfig,
			}).Set(0)
		}
	}
}

// createController create and add the controller to the "good controllers" list, if failed, add it to the "bad controllers" list.
func (m *BMHControllerManager) createController(ctx context.Context, metal3KubeconfigPath string) error {
	logger := log.FromContext(ctx).WithName("BMHManager.CreateController")
	logger.Info(fmt.Sprintf("creating BMHController for kubeconfig [%v]", metal3KubeconfigPath))

	var err error
	metal3KubeconfigPath = strings.TrimSpace(metal3KubeconfigPath)
	if len(metal3KubeconfigPath) == 0 {
		return fmt.Errorf("the provided metal3 kubeconfig file path is empty")
	}

	key := m.controllerNameGenFunc(metal3KubeconfigPath)

	controller, err := m.newControllerFunc(ctx, m.bmhConf, key, metal3KubeconfigPath)
	if err != nil {
		logger.Info(fmt.Sprintf("unable to create new BMHController for [%s], reason: %s", key, err.Error()))
		m.markControllerAsFailed(ctx, key)
		return err
	}

	m.markControllerAsSuccess(ctx, key, controller)
	logger.Info(fmt.Sprintf("successfully created BMHController [%v]", key))
	return nil
}

func (m *BMHControllerManager) StartAllControllers(ctx context.Context) {
	m.Lock()
	defer m.Unlock()
	for name, _ := range m.controllers {
		go m.startController(ctx, name)
	}
	// after starting all the controllers, spawn a go routine to periodically retry the failed ones.
	go m.retryFailedControllers(ctx)
}

// startController fetch the controller from the "good controllers" list, and start it.
func (m *BMHControllerManager) startController(ctx context.Context, name string) {
	logger := log.FromContext(ctx).WithName("BMHManager.startController")
	m.Lock()
	controller := m.controllers[name]
	m.Unlock()

	if controller == nil {
		logger.Info(fmt.Sprintf("BMHController [%s] is nil", name))
		m.markControllerAsFailed(ctx, name)
		return
	}

	// run the controller
	logger.Info(fmt.Sprintf("starting BMH Controller [%s]", name))
	err := controller.Run(ctx)
	// if we got an error when running the BMH Controller, don't panic, put it back to the retry list.
	if err != nil {
		logger.Info(fmt.Sprintf("failed running BMHController [%s], reason: %s", name, err.Error()))
		m.markControllerAsFailed(ctx, name)
	}
}

func (m *BMHControllerManager) markControllerAsFailed(ctx context.Context, name string) {
	m.Lock()
	defer m.Unlock()
	// remove this controller from the "good controllers" list
	delete(m.controllers, name)
	m.failed[name] = struct{}{}
}

func (m *BMHControllerManager) markControllerAsSuccess(ctx context.Context, name string, controller BMHControllerIF) {
	m.Lock()
	defer m.Unlock()
	delete(m.failed, name)
	m.controllers[name] = controller
}

func (m *BMHControllerManager) retryFailedControllers(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("BMHManager.retryFailedControllers")
	for {
		func() {

			time.Sleep(m.failedRetryInterval)

			// loop over the failed controllers and recreate them
			for name, _ := range m.failed {
				// get the conf for this controller
				conf, found := m.controllerKubeConfs[name]
				if !found {
					logger.Info(fmt.Sprintf("cannot find the kubeconfig for [%s] when trying to recreate the BMH Controller", name))
					continue
				}

				// recreate the BMH Controller
				err := m.createController(ctx, conf)
				if err != nil {
					logger.Info(fmt.Sprintf("CreateController failed for [%s], reason: %s", name, err.Error()))
					objectRef := &corev1.ObjectReference{
						Kind:       "Namespace",
						Name:       idcnetworkv1alpha1.SDNControllerNamespace,
						Namespace:  idcnetworkv1alpha1.SDNControllerNamespace,
						APIVersion: "v1",
					}
					// create event
					m.eventRecorder.Event(objectRef, corev1.EventTypeWarning, "Unable to create BMH Controller", err.Error())
					// update metrics
					metrics.BMHControllerManagerErrors.With(prometheus.Labels{
						metrics.MetricsLabelErrorType:          metrics.ErrorTypeCreateBMHControllerFailed,
						metrics.MetricsLabelKubeconfigFilePath: conf,
					}).Set(1)

					continue
				} else {
					metrics.BMHControllerManagerErrors.With(prometheus.Labels{
						metrics.MetricsLabelErrorType:          metrics.ErrorTypeCreateBMHControllerFailed,
						metrics.MetricsLabelKubeconfigFilePath: conf,
					}).Set(0)
				}
				go m.startController(ctx, name)
			}
		}()

	}
}

type BMHControllerIF interface {
	Run(context.Context) error
}

type BMHController struct {
	Name string

	sync.Mutex
	networkK8sClient client.Client
	informerFactory  metal3Informerfactory.SharedInformerFactory
	informer         cache.SharedIndexInformer
	lister           metal3v1alpha1.BareMetalHostLister
	queue            workqueue.RateLimitingInterface

	conf                               idcnetworkv1alpha1.SDNControllerConfig
	devicesManager                     devicesmanager.DevicesAccessManager
	PoolManager                        *pools.PoolManager
	BMHUsageReporter                   *pools.BMHUsageReporter
	GroupPoolMappingWatchIntervalInSec int

	EventRecorder record.EventRecorder
}

// BMHControllerConf has the common configurations for each BMH Controller
type BMHControllerConf struct {
	NwcpK8sClient                      client.Client
	Dam                                devicesmanager.DevicesAccessManager
	CtrlConf                           *idcnetworkv1alpha1.SDNControllerConfig
	GroupPoolMappingWatchIntervalInSec int
	PoolManager                        *pools.PoolManager
	BMHUsageReporter                   *pools.BMHUsageReporter
	EventRecorder                      record.EventRecorder

	BMHControllerCreationRetryInterval time.Duration
}

func NewBMHController(ctx context.Context, bmhConf BMHControllerConf, name string, metal3Kubeconfig string) (BMHControllerIF, error) {

	bmhController := &BMHController{}
	// load the kubeconfig
	var idcClusterRestClientConfig *rest.Config
	idcClusterRestClientConfig, err := utils.LoadKubeConfigFile(ctx, metal3Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("unable to LoadKubeConfigFile.  Kubeconfig servers: %s, reason %v", name, err)
	}

	if bmhConf.CtrlConf == nil {
		return nil, fmt.Errorf("SDNControllerConfig is not provided")
	}
	if bmhConf.Dam == nil {
		return nil, fmt.Errorf("DevicesAccessManager is not provided")
	}

	client, err := metal3ClientSet.NewForConfig(idcClusterRestClientConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get K8s REST config: %v", err)
	}

	informerFactory := metal3Informerfactory.NewSharedInformerFactory(client, time.Duration(bmhConf.CtrlConf.ControllerConfig.BMHResyncPeriodInSec)*time.Second)

	informer := informerFactory.Metal3().V1alpha1().BareMetalHosts().Informer()
	lister := informerFactory.Metal3().V1alpha1().BareMetalHosts().Lister()

	_, err = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    bmhController.handleBMHAdd,
		UpdateFunc: bmhController.handleBMHUpdate,
		DeleteFunc: bmhController.handleBMHDelete,
	})
	if err != nil {
		return nil, fmt.Errorf("AddEventHandler failed, reason: %s", err.Error())
	}

	bmhController.Name = name
	bmhController.networkK8sClient = bmhConf.NwcpK8sClient
	bmhController.informerFactory = informerFactory
	bmhController.devicesManager = bmhConf.Dam
	bmhController.informer = informer
	bmhController.conf = *bmhConf.CtrlConf
	bmhController.queue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "BMHEventQueue")
	bmhController.lister = lister

	bmhController.PoolManager = bmhConf.PoolManager
	bmhController.BMHUsageReporter = bmhConf.BMHUsageReporter
	bmhController.EventRecorder = bmhConf.EventRecorder

	return bmhController, nil
}

func (b *BMHController) GetQueue() workqueue.RateLimitingInterface {
	return b.queue
}

func (b *BMHController) Run(ctx context.Context) error {
	logger := log.FromContext(context.Background()).WithName("BMHController.Run")
	logger.Info("Starting BMHController")
	defer logger.Info("Shutting down BMHController")

	// start all the informers
	b.informerFactory.Start(ctx.Done())

	// Wait for all caches to be synced before processing events
	if !cache.WaitForCacheSync(ctx.Done(), b.informer.HasSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	// Process work queue items
	for {
		select {
		case <-ctx.Done():
			logger.Error(ctx.Err(), "context is done")
			return ctx.Err()
		// case err := <-errChan:
		// 	// we don't want to exit out when failed to get the mapping records, just log an error
		// 	logger.Error(err, "BMH Controller Run failed")
		default:
			if !b.processItem(ctx) {
				// if it got issue processing the items, log the error, don't exit out.
				logger.Info("processItem failed")
			}
		}
	}
}

func (b *BMHController) handleBMHAdd(obj interface{}) {
	logger := log.FromContext(context.Background()).WithName("BMHController.handleBMHAdd")
	logger.V(1).Info("BMH Add event observed")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		b.queue.Add(key)
	}
}

func (b *BMHController) handleBMHUpdate(oldObj, newObj interface{}) {
	logger := log.FromContext(context.Background()).WithName("BMHController.handleBMHUpdate")
	logger.V(1).Info("BMH Update event observed")
	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err == nil {
		b.queue.Add(key)
	}
}

func (b *BMHController) handleBMHDelete(obj interface{}) {
	logger := log.FromContext(context.Background()).WithName("BMHController.handleBMHDelete")
	logger.V(1).Info("BMH Delete event observed")
	// Handle tombstones
	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		obj = tombstone.Obj
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		b.queue.Add(key)
	}
}

func (b *BMHController) processItem(ctx context.Context) bool {
	logger := log.FromContext(context.Background()).WithName("BMHController.processItem")
	key, shutdown := b.queue.Get()
	if shutdown {
		return false
	}
	defer b.queue.Done(key)
	// Perform reconciliation for the item represented by the key
	err := b.reconcile(ctx, key.(string))
	if err != nil {
		logger.Error(err, fmt.Sprintf("reconcile failed, requeueing the BMH object %v, err: %v", key, err))
		// If reconciliation failed, requeue the item with a rate limit
		// b.queue.AddRateLimited(key)
		b.queue.AddAfter(key, 20*time.Second)
	} else {
		logger.V(1).Info(fmt.Sprintf("reconcile success, forgetting the BMH object %v", key))
		// If reconciliation succeeded, forget the item (resetting any rate limit counters for this key)
		b.queue.Forget(key)
	}
	return true
}

func (b *BMHController) reconcile(ctx context.Context, key string) error {
	logger := log.FromContext(ctx).WithName("BMHController.reconcile")

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	bareMetalHost, err := b.lister.BareMetalHosts(namespace).Get(name)
	objectExists := true
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(1).Info("BMH not found", utils.LogFieldBareMetalHost, name)
			objectExists = false
		} else {
			logger.Error(err, "unable to fetch BMH", "name", key)
			return err
		}
	}

	if bareMetalHost != nil && bareMetalHost.ObjectMeta.DeletionTimestamp.IsZero() {
		// if a BMH is not marked as deleted, take it as a Add or Update event.
		return b.handleBMHAddorUpdateEvents(ctx, bareMetalHost)
	} else if (bareMetalHost != nil && !bareMetalHost.ObjectMeta.DeletionTimestamp.IsZero()) || !objectExists {
		// if a BMH is marked as deleted, or a BMH is NOT found, we can delete the related NetworkNode.
		return b.handleBMHDeleteEvents(ctx, name)
	}
	return nil
}

const (
	ManufacturerWIWYNN = "WIWYNN"
	ManufacturerIntel  = "Intel Corporation"
)

func (b *BMHController) handleBMHAddorUpdateEvents(ctx context.Context, bareMetalHost *baremetalv1alpha1.BareMetalHost) error {
	logger := log.FromContext(ctx).WithName("BMHController.handleBMHAddorUpdateEvents").WithValues(utils.LogFieldBareMetalHost, bareMetalHost.Name)
	var err error

	b.BMHUsageReporter.ReportBMH(bareMetalHost)

	////////////////////////////
	// Handling NetworkNode
	////////////////////////////
	// check if the BareMetal host is in ready state.
	if bareMetalHost.Status.OperationalStatus != baremetalv1alpha1.OperationalStatusOK {
		logger.Info("bareMetalHost is not in `OperationalStatusOK` state.", "bareMetalHost", bareMetalHost.Name, "operationalStatus", bareMetalHost.Status.OperationalStatus)
		// the BMH is not ready yet, so just return nil.
		return nil
	}

	// try to create or update the NetworkNode for the given BMH
	networkNode, err := b.addOrUpdateNetworkNode(ctx, bareMetalHost)
	if err != nil {
		logger.Error(err, "addOrUpdateNetworkNode failed")
		return err
	}

	////////////////////////////
	// Handling NodeGroup
	////////////////////////////
	var desiredGroupName string
	// get the Group name from the BMH labels. This is the source of truth of the group id.
	if bareMetalHost.Labels != nil {
		desiredGroupName = bareMetalHost.Labels[idcnetworkv1alpha1.LabelBMHGroupID]
	}

	// get the current NodeGroup Name (that stored in the NetworkNode)
	currentGroupName := networkNode.Labels[idcnetworkv1alpha1.LabelGroupID]
	// if the desiredGroupName is not the same as currentGroupName(the one stored in the NetworkNode), we need to remove the NetworkNode from the "currentGroupName" NodeGroup.
	if len(currentGroupName) > 0 && desiredGroupName != currentGroupName {
		err = b.removeNetworkNodeAndSwitchesFromGroup(ctx, networkNode, currentGroupName)
		if err != nil {
			logger.Error(err, "removeNetworkNodeAndSwitchesFromGroup failed")
			return err
		}
	}

	// if the desired group is empty, return.
	if len(desiredGroupName) == 0 {
		logger.V(1).Info("BMH has no group id label")
		return nil
	}

	objectExists := true
	// check if there is an existing desiredGroupName CR
	nodeGroup, err := b.getNodeGroup(ctx, desiredGroupName)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		objectExists = false
	}

	// desiredGroupName NodeGroup not found, create a new one.
	if !objectExists {
		nodeGroup, err = b.createNodeGroup(ctx, desiredGroupName)
		if err != nil {
			logger.Error(err, "createNodeGroup failed")
			return err
		}
	}

	// try to add this NetworkNode and switches to a NodeGroup
	err = b.addNetworkNodeAndSwitchesToGroup(ctx, networkNode, nodeGroup)
	if err != nil {
		logger.Error(err, "addNetworkNodeAndSwitchesToGroup failed")
		return err
	}

	return nil
}

func (b *BMHController) getNodeGroup(ctx context.Context, groupName string) (*idcnetworkv1alpha1.NodeGroup, error) {
	logger := log.FromContext(ctx).WithName("BMHController.getNodeGroup").WithValues(utils.LogFieldNodeGroupName, groupName)

	nodeGroup := &idcnetworkv1alpha1.NodeGroup{}
	key := types.NamespacedName{Name: groupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := b.networkK8sClient.Get(ctx, key, nodeGroup)
	if err != nil {
		logger.V(1).Error(err, "failed to get NodeGroup")
		return nil, err
	}
	return nodeGroup, nil
}

// getOrCreateNodeToGroup creates an NodeGroup if it doesn't exist
func (b *BMHController) createNodeGroup(ctx context.Context, groupName string) (*idcnetworkv1alpha1.NodeGroup, error) {
	logger := log.FromContext(ctx).WithName("BMHController.getOrCreateNodeGroup").WithValues(utils.LogFieldNodeGroupName, groupName)

	// create a new NodeGroup
	nodeGroup := &idcnetworkv1alpha1.NodeGroup{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    map[string]string{},
			Namespace: idcnetworkv1alpha1.SDNControllerNamespace,
			Name:      groupName,
		},
		Spec: idcnetworkv1alpha1.NodeGroupSpec{
			NetworkNodes:            []string{},
			FrontEndLeafSwitches:    make([]string, 0),
			AcceleratorLeafSwitches: make([]string, 0),
			StorageLeafSwitches:     make([]string, 0),
			FrontEndFabricConfig:    &idcnetworkv1alpha1.FabricConfig{},
			AcceleratorFabricConfig: &idcnetworkv1alpha1.FabricConfig{},
			StorageFabricConfig:     &idcnetworkv1alpha1.FabricConfig{},
		},
	}

	err := b.networkK8sClient.Create(ctx, nodeGroup)
	if err != nil {
		logger.Error(err, "networkK8sClient create NodeGroup CR failed")
		return nil, err
	}

	return nodeGroup, nil
}

// addNetworkNodeAndSwitchesToGroup tries to add the NetworkNode and the switches to the NodeGroup.
func (b *BMHController) addNetworkNodeAndSwitchesToGroup(ctx context.Context, networkNode *idcnetworkv1alpha1.NetworkNode, nodeGroup *idcnetworkv1alpha1.NodeGroup) error {
	logger := log.FromContext(ctx).WithName("BMHController.addNetworkNodeAndSwitchesToGroup").WithValues(utils.LogFieldBareMetalHost, networkNode.Name)
	b.Lock()
	defer b.Unlock()

	nodeGroupCopy := nodeGroup.DeepCopy()
	nodeExists := false
	for _, node := range nodeGroupCopy.Spec.NetworkNodes {
		if node == networkNode.Name {
			// node already exists in the node list.
			nodeExists = true
		}
	}
	if !nodeExists {
		// this NetworkNode is not in the targetGroup's node list, add it now.
		nodeGroupCopy.Spec.NetworkNodes = append(nodeGroupCopy.Spec.NetworkNodes, networkNode.Name)

		// 02/27/2024: Lets NOT reset the NetworkNode, just leave it as is when adding it to the Group. Will keep the code but comment it out just in case we need it in the future.
		// try to reset a networkNode's default vlan values. Only do this when the NetworkNode is newly added to a NodeGroup!! otherwise the it will overwrite the existing values.
		// When MSU is NodeGroup, after adding a NetworkNode into the NodeGroup, the NodeGroup Controller will push its current Vlan config to the NetworkNode
		// when MSU is NetworkNode, NodeGroup Controller will not enforce the Vlan value to its children, so we want to set the default Vlan for each NetworkNode to the Pool defined value.
		// targetPool, err := b.PoolManager.GetPoolByGroupName(ctx, targetGroupName)
		// if err != nil {
		// 	return err
		// }
		// if targetPool != nil {
		// 	if targetPool.SchedulingConfig.MinimumSchedulableUnit == idcnetworkv1alpha1.MSUNetworkNode {
		// 		err = b.updateNetworkNodeWithDefaultVlan(ctx, networkNode.Name, targetPool)
		// 		if err != nil {
		// 			return err
		// 		}
		// 	}
		// }
	}

	// add the NetworkNode front-end switch to the NodeGroup
	func() {
		if networkNode.Spec.FrontEndFabric != nil {
			_, feSwitchFQDN := utils.ExtractSwitchAndPortNameFromSwitchPortCRName(networkNode.Spec.FrontEndFabric.SwitchPort)
			if len(feSwitchFQDN) == 0 {
				logger.Info("got an empty switch FQDN from NN [%v] front end SP [%v]", networkNode.Name, networkNode.Spec.FrontEndFabric.SwitchPort)
				return
			}

			feSwitchAlreadyExists := false
			for _, sw := range nodeGroupCopy.Spec.FrontEndLeafSwitches {
				if sw == feSwitchFQDN {
					feSwitchAlreadyExists = true
					break
				}
			}
			// if fe switch doesn't exist, add it to the NodeGroup, and update the switch CR labels.
			if !feSwitchAlreadyExists {
				nodeGroupCopy.Spec.FrontEndLeafSwitches = append(nodeGroupCopy.Spec.FrontEndLeafSwitches, feSwitchFQDN)
				err := b.updateSwitchMeta(ctx, feSwitchFQDN, nodeGroup.Name, idcnetworkv1alpha1.FabricTypeFrontEnd)
				if err != nil {
					logger.Error(err, "update switch CR metadata failed")
				}
			}
		}
	}()

	// add the NetworkNode accelerator switch/es to the NodeGroup
	func() {
		if networkNode.Spec.AcceleratorFabric != nil && len(networkNode.Spec.AcceleratorFabric.SwitchPorts) > 0 {
			existingAccelSwitches := make(map[string]struct{})
			for _, sw := range nodeGroupCopy.Spec.AcceleratorLeafSwitches {
				existingAccelSwitches[sw] = struct{}{}
			}
			for _, sp := range networkNode.Spec.AcceleratorFabric.SwitchPorts {
				_, accelSwitchFQDN := utils.ExtractSwitchAndPortNameFromSwitchPortCRName(sp)
				if len(accelSwitchFQDN) == 0 {
					logger.Info("got an empty switch FQDN from NN [%v] acc SP [%v]", networkNode.Name, sp)
					continue
				}

				if _, found := existingAccelSwitches[accelSwitchFQDN]; !found {
					// we found a new switch for a NodeGroup
					nodeGroupCopy.Spec.AcceleratorLeafSwitches = append(nodeGroupCopy.Spec.AcceleratorLeafSwitches, accelSwitchFQDN)
					existingAccelSwitches[accelSwitchFQDN] = struct{}{}
					// update the switch CR with the group and fabric type label
					err := b.updateSwitchMeta(ctx, accelSwitchFQDN, nodeGroup.Name, idcnetworkv1alpha1.FabricTypeAccelerator)
					if err != nil {
						logger.Error(err, "update switch CR metadata failed")
					}
				}
			}
		}
	}()

	// add the NetworkNode storage switch/es to the NodeGroup
	func() {
		if networkNode.Spec.StorageFabric != nil && len(networkNode.Spec.StorageFabric.SwitchPorts) > 0 {
			existingStrgSwitches := make(map[string]struct{})
			for _, sw := range nodeGroupCopy.Spec.StorageLeafSwitches {
				existingStrgSwitches[sw] = struct{}{}
			}
			for _, sp := range networkNode.Spec.StorageFabric.SwitchPorts {
				_, strgSwitchFQDN := utils.ExtractSwitchAndPortNameFromSwitchPortCRName(sp)
				if len(strgSwitchFQDN) == 0 {
					logger.Info("got an empty switch FQDN from NN [%v] storage SP [%v]", networkNode.Name, sp)
					continue
				}

				if _, found := existingStrgSwitches[strgSwitchFQDN]; !found {
					// we found a new switch for a NodeGroup
					nodeGroupCopy.Spec.StorageLeafSwitches = append(nodeGroupCopy.Spec.StorageLeafSwitches, strgSwitchFQDN)
					existingStrgSwitches[strgSwitchFQDN] = struct{}{}
					// update the switch CR with the group and fabric type label
					err := b.updateSwitchMeta(ctx, strgSwitchFQDN, nodeGroup.Name, idcnetworkv1alpha1.FabricTypeStorage)
					if err != nil {
						logger.Error(err, "update switch CR metadata failed")
					}
				}
			}
		}
	}()

	// update the NetworkNode with Group ID label
	latestNetworkNode := &idcnetworkv1alpha1.NetworkNode{}
	nnkey := types.NamespacedName{Name: networkNode.Name, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := b.networkK8sClient.Get(ctx, nnkey, latestNetworkNode)
	if err != nil {
		return err
	}
	latestNetworkNodeCopy := latestNetworkNode.DeepCopy()
	if latestNetworkNodeCopy.Labels == nil {
		latestNetworkNodeCopy.Labels = make(map[string]string)
	}
	latestNetworkNodeCopy.Labels[idcnetworkv1alpha1.LabelGroupID] = nodeGroup.Name

	nnpatch := client.MergeFromWithOptions(latestNetworkNode, client.MergeFromWithOptimisticLock{})
	err = b.networkK8sClient.Patch(ctx, latestNetworkNodeCopy, nnpatch)
	if err != nil {
		return err
	}

	// update the NodeGroup
	patch := client.MergeFromWithOptions(nodeGroup, client.MergeFromWithOptimisticLock{})
	if err := b.networkK8sClient.Patch(ctx, nodeGroupCopy, patch); err != nil {
		return err
	}
	return nil
}

func (b *BMHController) updateSwitchMeta(ctx context.Context, switchFQDN string, groupName string, fabricType string) error {
	existingSwitch := &idcnetworkv1alpha1.Switch{}
	key := types.NamespacedName{Name: switchFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := b.networkK8sClient.Get(ctx, key, existingSwitch)
	if err != nil {
		return err
	}
	switchCopy := existingSwitch.DeepCopy()
	if switchCopy.Labels == nil {
		switchCopy.Labels = make(map[string]string)
	}
	switchCopy.Labels[idcnetworkv1alpha1.LabelFabricType] = fabricType
	patch := client.MergeFrom(existingSwitch)
	err = b.networkK8sClient.Patch(ctx, switchCopy, patch)
	if err != nil {
		return err
	}
	return nil
}

// removeNetworkNodeAndSwitchesFromGroup
// removing a NetworkNode from a NodeGroup involving removing the node from the node list, and the related switches from the switch list. Also, we need to remove
// the Group ID label from the NetworkNode. We DON'T change the NetworkNode's Vlan values. While in the addNetworkNodeAndSwitchesToGroup(), when we add a NetworkNode
// to a NodeGroup, and if the NetworkNode is in a Pool whose MSU is "Node", we set the NetworkNode's Vlan to the Pool's default values.
func (b *BMHController) removeNetworkNodeAndSwitchesFromGroup(ctx context.Context, networkNode *idcnetworkv1alpha1.NetworkNode, groupName string) error {
	logger := log.FromContext(ctx).WithName("BMHController.removeNetworkNodeAndSwitchesFromGroup").WithValues(utils.LogFieldBareMetalHost, networkNode.Name)

	// try to get the NodeGroup
	existingNodeGroup := &idcnetworkv1alpha1.NodeGroup{}
	objectExists := true
	key := types.NamespacedName{Name: groupName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err := b.networkK8sClient.Get(ctx, key, existingNodeGroup)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Object not found
		objectExists = false
	}

	// NodeGroup doesn't exist, no action is needed
	if !objectExists {
		return nil
	}

	// remove the node from the NodeGroup
	newNodeList := make([]string, 0)
	for i := range existingNodeGroup.Spec.NetworkNodes {
		if existingNodeGroup.Spec.NetworkNodes[i] == networkNode.Name {
			continue
		}
		newNodeList = append(newNodeList, existingNodeGroup.Spec.NetworkNodes[i])
	}

	// figure out the latest fe and acc switches for a NodeGroup after removing a NetworkNode.
	feSwitches := make([]string, 0)
	accSwitches := make([]string, 0)
	strgSwitches := make([]string, 0)
	feSwitchesSet := make(map[string]struct{})
	accSwitchesSet := make(map[string]struct{})
	strgSwitchesSet := make(map[string]struct{})
	for i := range newNodeList {
		node := newNodeList[i]
		existingNetworkNodeCR := &idcnetworkv1alpha1.NetworkNode{}
		key := types.NamespacedName{Name: node, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err = b.networkK8sClient.Get(ctx, key, existingNetworkNodeCR)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			// Object not found, skip it
			continue
		}
		// collect the fe switches
		_, feSwitchName := utils.ExtractSwitchAndPortNameFromSwitchPortCRName(existingNetworkNodeCR.Spec.FrontEndFabric.SwitchPort)
		feSwitchesSet[feSwitchName] = struct{}{}
		// collect the acc switches
		if existingNetworkNodeCR.Spec.AcceleratorFabric != nil {
			for j := range existingNetworkNodeCR.Spec.AcceleratorFabric.SwitchPorts {
				_, accSwitchName := utils.ExtractSwitchAndPortNameFromSwitchPortCRName(existingNetworkNodeCR.Spec.AcceleratorFabric.SwitchPorts[j])
				accSwitchesSet[accSwitchName] = struct{}{}
			}
		}
		// collect the storage switches
		if existingNetworkNodeCR.Spec.StorageFabric != nil {
			for j := range existingNetworkNodeCR.Spec.StorageFabric.SwitchPorts {
				_, strgSwitchName := utils.ExtractSwitchAndPortNameFromSwitchPortCRName(existingNetworkNodeCR.Spec.StorageFabric.SwitchPorts[j])
				strgSwitchesSet[strgSwitchName] = struct{}{}
			}
		}
	}
	for swName, _ := range feSwitchesSet {
		feSwitches = append(feSwitches, swName)
	}
	for swName, _ := range accSwitchesSet {
		accSwitches = append(accSwitches, swName)
	}
	for swName, _ := range strgSwitchesSet {
		strgSwitches = append(strgSwitches, swName)
	}

	newNodeGroup := existingNodeGroup.DeepCopy()
	newNodeGroup.Spec.NetworkNodes = newNodeList
	newNodeGroup.Spec.FrontEndLeafSwitches = feSwitches
	newNodeGroup.Spec.AcceleratorLeafSwitches = accSwitches
	newNodeGroup.Spec.StorageLeafSwitches = strgSwitches
	patch := client.MergeFromWithOptions(existingNodeGroup, client.MergeFromWithOptimisticLock{})
	if err := b.networkK8sClient.Patch(ctx, newNodeGroup, patch); err != nil {
		logger.Error(err, "front end switch port CR patch update failed")
		return err
	}

	// remove the group id label from the current NetworkNode CR
	latestNetworkNodeCR := &idcnetworkv1alpha1.NetworkNode{}
	nnkey := types.NamespacedName{Name: networkNode.Name, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = b.networkK8sClient.Get(ctx, nnkey, latestNetworkNodeCR)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		return nil
	}
	latestNetworkNodeCRCopy := latestNetworkNodeCR.DeepCopy()
	nnPatch := client.MergeFrom(latestNetworkNodeCR)
	delete(latestNetworkNodeCRCopy.Labels, idcnetworkv1alpha1.LabelGroupID)
	if err := b.networkK8sClient.Patch(ctx, latestNetworkNodeCRCopy, nnPatch); err != nil {
		logger.Error(err, "NetworkNode CR patch update failed")
		return err
	}

	return nil
}

func (b *BMHController) checkSwitchCRExists(ctx context.Context, frontEndPortInfo SwitchPortInfo, accelPortsInfo []SwitchPortInfo, storagePortsInfo []SwitchPortInfo, bmhHostName string) error {
	switchesWanted := make(map[string]struct{})
	// find all the switches we need for this BMH
	switchesWanted[frontEndPortInfo.SwitchFQDN] = struct{}{}
	for _, accItem := range accelPortsInfo {
		switchesWanted[accItem.SwitchFQDN] = struct{}{}
	}
	for _, stgItem := range storagePortsInfo {
		switchesWanted[stgItem.SwitchFQDN] = struct{}{}
	}

	// if switches don't exist in k8s cluster, raise an alert
	missingSwitches := make([]string, 0)
	for switchFQDN, _ := range switchesWanted {
		existingSwitch := &idcnetworkv1alpha1.Switch{}
		key := types.NamespacedName{Name: switchFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}

		err := b.networkK8sClient.Get(ctx, key, existingSwitch)
		if err != nil {
			// if we got other errors, return directly.
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			// Object not found
			missingSwitches = append(missingSwitches, switchFQDN)
			// set the error flag metric for a specfic switch + BMH if it's not found.
			metrics.BMHControllerErrors.With(prometheus.Labels{
				metrics.MetricsLabelErrorType:  metrics.ErrorTypeMissingSwitchCR,
				metrics.MetricsLabelSwitchFQDN: switchFQDN,
				metrics.MetricsLabelHostName:   bmhHostName,
			}).Set(1)

		} else {
			// reset the value to 0 when the issue is gone or when it's normal
			// so even a switch is imported, it will still has a 0 value.
			// the alert or grafana diagram should only care about the one with value 1.
			metrics.BMHControllerErrors.With(prometheus.Labels{
				metrics.MetricsLabelErrorType:  metrics.ErrorTypeMissingSwitchCR,
				metrics.MetricsLabelSwitchFQDN: switchFQDN,
				metrics.MetricsLabelHostName:   bmhHostName,
			}).Set(0)
		}
	}

	if len(missingSwitches) > 0 {
		err := fmt.Errorf("switch CRs are missing: %v, BMH name: %s", missingSwitches, bmhHostName)
		// associate the event with the namespace, as the Switch CRs are missing
		objectRef := &corev1.ObjectReference{
			Kind:       "Namespace",
			Name:       idcnetworkv1alpha1.SDNControllerNamespace,
			Namespace:  idcnetworkv1alpha1.SDNControllerNamespace,
			APIVersion: "v1",
		}
		b.EventRecorder.Event(objectRef, corev1.EventTypeWarning, "Switch CRs Missing", err.Error())
		return err
	}
	return nil
}

func (b *BMHController) addOrUpdateNetworkNode(ctx context.Context, bareMetalHost *baremetalv1alpha1.BareMetalHost) (*idcnetworkv1alpha1.NetworkNode, error) {
	logger := log.FromContext(ctx).WithName("BMHController.addOrUpdateNetworkNode").WithValues(utils.LogFieldBareMetalHost, bareMetalHost.Name)
	// extract the ports information from the BMH lldp records.
	frontEndPortInfo, accelPortsInfo, storagePortsInfo, err := b.findSwitchPortsFromBMH(ctx, bareMetalHost)
	if err != nil {
		logger.Error(err, "error extracting port info from BMH")
		return nil, err
	}

	// check if the wanted switches have already been created/imported into the K8s cluster, if not, raise an alert.
	err = b.checkSwitchCRExists(ctx, *frontEndPortInfo, accelPortsInfo, storagePortsInfo, bareMetalHost.Name)
	if err != nil {
		logger.Error(err, "checkSwitchCRExists failed")
		return nil, err
	}

	// check if there is an existing NetworkNode CR.
	hostName := bareMetalHost.Name
	existingNetworkNodeCR := &idcnetworkv1alpha1.NetworkNode{}
	key := types.NamespacedName{Name: hostName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	objectExists := true
	err = b.networkK8sClient.Get(ctx, key, existingNetworkNodeCR)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		// Object not found
		objectExists = false
	}

	var newNetworkNodeCR *idcnetworkv1alpha1.NetworkNode

	// if we already have an existing NN, try to update it.
	if objectExists {
		// TODO: get the "Instance" that associated with the BMH

		newNetworkNodeCR = existingNetworkNodeCR.DeepCopy()

		//////////////////////////////////////
		// update front-end fabric information
		//////////////////////////////////////

		frontEndSwitchFqdn := frontEndPortInfo.SwitchFQDN
		frontEndSwitchPort := frontEndPortInfo.SwitchPort
		frontendPortCRName := utils.GeneratePortFullName(frontEndSwitchFqdn, frontEndSwitchPort)

		if newNetworkNodeCR.Spec.FrontEndFabric == nil {
			newNetworkNodeCR.Spec.FrontEndFabric = &idcnetworkv1alpha1.FrontEndFabric{}
		}
		newNetworkNodeCR.Spec.FrontEndFabric.SwitchPort = frontendPortCRName

		// if Mode is not set, then we will get the actual mode value from the switch.
		if len(newNetworkNodeCR.Spec.FrontEndFabric.Mode) == 0 {
			// get the actual frontend info from the switch
			frontEndSwitchPortDetails, err := b.getActualSwitchPortDetails(ctx, frontEndSwitchFqdn, frontEndSwitchPort)
			if err != nil {
				logger.Error(err, "getActualSwitchPortDetails failed")
				return nil, err
			}
			// set the NN frontend switch port mode.
			newNetworkNodeCR.Spec.FrontEndFabric.Mode = frontEndSwitchPortDetails.Mode
		}

		set := make(map[string]struct{})
		//////////////////////////////////////
		// update accelerator fabric information
		//////////////////////////////////////
		if len(accelPortsInfo) > 0 {
			accelPortCRNameList := make([]string, 0)
			for _, portInfo := range accelPortsInfo {
				switchFqdn := portInfo.SwitchFQDN
				switchPort := portInfo.SwitchPort
				accelPortCRName := utils.GeneratePortFullName(switchFqdn, switchPort)
				if _, found := set[accelPortCRName]; !found {
					accelPortCRNameList = append(accelPortCRNameList, accelPortCRName)
					set[accelPortCRName] = struct{}{}
				}
			}
			if newNetworkNodeCR.Spec.AcceleratorFabric == nil {
				newNetworkNodeCR.Spec.AcceleratorFabric = &idcnetworkv1alpha1.AcceleratorFabric{}
			}
			newNetworkNodeCR.Spec.AcceleratorFabric.SwitchPorts = accelPortCRNameList

			// if Mode is not set, then we will get the actual mode value from the switch.
			if len(newNetworkNodeCR.Spec.AcceleratorFabric.Mode) == 0 {
				_, mode, err := b.getFabricMajorityConf(ctx, accelPortsInfo)
				if err != nil {
					logger.Error(err, "get Accelerator Fabric mode failed")
					return nil, err
				}
				newNetworkNodeCR.Spec.AcceleratorFabric.Mode = mode
			}
		}

		//////////////////////////////////////
		// update storage fabric information
		//////////////////////////////////////
		if len(storagePortsInfo) > 0 {
			storagePortCRNameList := make([]string, 0)
			for _, portInfo := range storagePortsInfo {
				switchFqdn := portInfo.SwitchFQDN
				switchPort := portInfo.SwitchPort
				storagePortCRName := utils.GeneratePortFullName(switchFqdn, switchPort)
				if _, found := set[storagePortCRName]; !found {
					storagePortCRNameList = append(storagePortCRNameList, storagePortCRName)
					set[storagePortCRName] = struct{}{}
				}
			}
			if newNetworkNodeCR.Spec.StorageFabric == nil {
				newNetworkNodeCR.Spec.StorageFabric = &idcnetworkv1alpha1.StorageFabric{}
			}
			newNetworkNodeCR.Spec.StorageFabric.SwitchPorts = storagePortCRNameList

			if len(newNetworkNodeCR.Spec.StorageFabric.Mode) == 0 {
				// if Mode is not set, then we will get the actual mode value from the switch.
				_, mode, err := b.getFabricMajorityConf(ctx, storagePortsInfo)
				if err != nil {
					logger.Error(err, "get storage Fabric mode failed")
					return nil, err
				}
				newNetworkNodeCR.Spec.StorageFabric.Mode = mode
			}
		}

		// add the BMH namespace label
		if newNetworkNodeCR.Labels == nil {
			newNetworkNodeCR.Labels = make(map[string]string)
		}
		if _, found := newNetworkNodeCR.Labels[idcnetworkv1alpha1.LabelBMHNameSpace]; !found {
			newNetworkNodeCR.Labels[idcnetworkv1alpha1.LabelBMHNameSpace] = bareMetalHost.Namespace
		}

		// patch update the NetworkNode
		// TODO: Only do this if we actually made a change?
		patch := client.MergeFrom(existingNetworkNodeCR)
		if err := b.networkK8sClient.Patch(ctx, newNetworkNodeCR, patch); err != nil {
			logger.Error(err, "front end switch port CR patch update failed")
		}
		logger.V(1).Info("finished updating NetworkNode CR", utils.LogFieldNetworkNode, hostName)
	} else {
		// if the NetworkNode does not exist, create a new one
		newNetworkNodeCR = &idcnetworkv1alpha1.NetworkNode{
			ObjectMeta: metav1.ObjectMeta{
				Labels:    map[string]string{},
				Namespace: idcnetworkv1alpha1.SDNControllerNamespace,
				Name:      hostName,
			},
			Spec: idcnetworkv1alpha1.NetworkNodeSpec{
				FrontEndFabric: &idcnetworkv1alpha1.FrontEndFabric{},
			},
		}

		//////////////////////////////////////
		// update frontEnd fabric information
		//////////////////////////////////////
		frontEndSwitchFqdn := frontEndPortInfo.SwitchFQDN
		frontEndSwitchPort := frontEndPortInfo.SwitchPort
		frontendPortCRName := utils.GeneratePortFullName(frontEndSwitchFqdn, frontEndSwitchPort)
		newNetworkNodeCR.Spec.FrontEndFabric.SwitchPort = frontendPortCRName

		// get the actual frontend info from the switch
		frontEndSwitchPortDetails, err := b.getActualSwitchPortDetails(ctx, frontEndSwitchFqdn, frontEndSwitchPort)
		if err != nil {
			logger.Error(err, "getActualSwitchPortDetails failed")
			return nil, err
		}
		newNetworkNodeCR.Spec.FrontEndFabric.Mode = frontEndSwitchPortDetails.Mode

		set := make(map[string]struct{})
		//////////////////////////////////////
		// update accelerator fabric information
		//////////////////////////////////////
		if len(accelPortsInfo) > 0 {
			// collect accelerator fabric information
			accelPortCRNameList := make([]string, 0)
			for _, portInfo := range accelPortsInfo {
				switchFqdn := portInfo.SwitchFQDN
				switchPort := portInfo.SwitchPort
				accelPortCRName := utils.GeneratePortFullName(switchFqdn, switchPort)
				if _, found := set[accelPortCRName]; !found {
					accelPortCRNameList = append(accelPortCRNameList, accelPortCRName)
					set[accelPortCRName] = struct{}{}
				}
			}

			newNetworkNodeCR.Spec.AcceleratorFabric = &idcnetworkv1alpha1.AcceleratorFabric{
				SwitchPorts: accelPortCRNameList,
			}

			_, accelMode, err := b.getFabricMajorityConf(ctx, accelPortsInfo)
			if err != nil {
				logger.Error(err, "getFabricMajorityConf failed")
				return nil, err
			}
			newNetworkNodeCR.Spec.AcceleratorFabric.Mode = accelMode
		}

		//////////////////////////////////////
		// update storage fabric information
		//////////////////////////////////////
		if len(storagePortsInfo) > 0 {
			storagePortCRNameList := make([]string, 0)
			for _, portInfo := range storagePortsInfo {
				switchFqdn := portInfo.SwitchFQDN
				switchPort := portInfo.SwitchPort
				storagePortCRName := utils.GeneratePortFullName(switchFqdn, switchPort)
				if _, found := set[storagePortCRName]; !found {
					storagePortCRNameList = append(storagePortCRNameList, storagePortCRName)
					set[storagePortCRName] = struct{}{}
				}
			}

			newNetworkNodeCR.Spec.StorageFabric = &idcnetworkv1alpha1.StorageFabric{
				SwitchPorts: storagePortCRNameList,
			}

			_, storageMode, err := b.getFabricMajorityConf(ctx, storagePortsInfo)
			if err != nil {
				logger.Error(err, "get storage Fabric Vlan failed")
				return nil, err
			}
			newNetworkNodeCR.Spec.StorageFabric.Mode = storageMode

		}

		// add the BMH namespace label
		newNetworkNodeCR.Labels[idcnetworkv1alpha1.LabelBMHNameSpace] = bareMetalHost.Namespace
		err = b.networkK8sClient.Create(ctx, newNetworkNodeCR)
		if err != nil {
			logger.Error(err, "networkK8sClient create BMH CR failed")
		}
		logger.Info("created a new NetworkNode CR", utils.LogFieldNetworkNode, hostName)
	}

	return newNetworkNodeCR, nil
}

func (b *BMHController) getActualSwitchPortDetails(ctx context.Context, switchFqdn string, switchPortName string) (*idcnetworkv1alpha1.SwitchPortStatus, error) {
	logger := log.FromContext(ctx).WithName("BMHController.getFrontEndFabricPortVlan")
	// if the switch port is not initialized, then we will create one.
	// for vlan id, we will get it from the switch. Technically the Port's Vlan for a new enrolled BM should be 4008,
	// but getting it from the switch is more accurate and deterministic.
	// logger.Info("trying to get switch client from DeviceManager", utils.LogFieldSwitchFQDN, frontEndSwitchFqdn)
	swClient, err := b.devicesManager.GetSwitchClient(devicesmanager.GetOption{
		SwitchFQDN: switchFqdn,
	})
	if err != nil {
		logger.Info(fmt.Sprintf("failed to get switch client from DeviceManager for switch %v, please check if the Switch CR has been created and the switch is accessible.", switchFqdn))
		return nil, err
	}
	// fetch ports from the switch.
	portsFromSwitch, err := swClient.GetSwitchPorts(ctx, sc.GetSwitchPortsRequest{
		SwitchFQDN: switchFqdn,
	})
	if err != nil {
		logger.Info(fmt.Sprintf("failed to get ports info from the switch %v, %v", switchFqdn, err))
		return nil, err
	}

	switchPortDetails, found := portsFromSwitch[switchPortName]
	if !found {
		return nil, fmt.Errorf("cannot find port %v in the switch", switchPortName)
	}
	return switchPortDetails, nil
}

// getFabricVlan tried to get the current Vlan ID for the switch ports that a BMH connected to. If there are multiple Vlans, return the majority one.
func (b *BMHController) getFabricMajorityConf(ctx context.Context, portsInfo []SwitchPortInfo) (int64, string, error) {
	logger := log.FromContext(ctx).WithName("BMHController.getAcceleratorFabricVlan")
	// extract all the switches. key: switchFQDN, value: port list that connect to the BMH
	// a Gaudi node should have connected to 3 accelerator switches, and connected with 8 ports for each switch.
	var majorityVlan int64
	var majorityMode string
	switchList := make(map[string][]string, 0)
	for i := range portsInfo {
		_, found := switchList[portsInfo[i].SwitchFQDN]
		if !found {
			switchList[portsInfo[i].SwitchFQDN] = make([]string, 0)
		}
		switchList[portsInfo[i].SwitchFQDN] = append(switchList[portsInfo[i].SwitchFQDN], portsInfo[i].SwitchPort)
	}

	// get ports information for the switches
	vlanCnt := make(map[int64]int)
	modeCnt := make(map[string]int)
	total := 0
	for switchFQDN, portList := range switchList {
		switchClient, err := b.devicesManager.GetSwitchClient(devicesmanager.GetOption{SwitchFQDN: switchFQDN})
		if err != nil {
			return 0, "", fmt.Errorf("GetSwitchClient failed, %v", err)
		}
		allSwitchPorts, err := switchClient.GetSwitchPorts(ctx, sc.GetSwitchPortsRequest{
			SwitchFQDN: switchFQDN,
		})
		total += len(allSwitchPorts)
		// iterate the portList and find the VLAN for each of them
		for i := range portList {
			actualPortConf, found := allSwitchPorts[portList[i]]
			if !found || actualPortConf == nil {
				logger.Info(fmt.Sprintf("can't find port %v from switch %v", portList[i], switchFQDN))
				return 0, "", fmt.Errorf(fmt.Sprintf("can't find port %v from switch %v", portList[i], switchFQDN))
			}
			vlanCnt[actualPortConf.VlanId]++
			modeCnt[actualPortConf.Mode]++
		}
	}
	// get majority vlan
	vlanCntFreq := make([][]int64, total)
	for vlan, cnt := range vlanCnt {
		if vlanCntFreq[cnt] == nil {
			vlanCntFreq[cnt] = make([]int64, 0)
		}
		vlanCntFreq[cnt] = append(vlanCntFreq[cnt], vlan)
	}
	for i := total - 1; i >= 0; i-- {
		if len(vlanCntFreq[i]) > 0 {
			// return the highest count vlan, if there are multiple ones have the same count, just simply return the first one.
			logger.Info(fmt.Sprintf("majority-Vlan: %v", vlanCntFreq[i][0]))
			majorityVlan = vlanCntFreq[i][0]
			break
		}
	}

	// get majority mode
	modeCntFreq := make([][]string, total)
	for mode, cnt := range modeCnt {
		if modeCntFreq[cnt] == nil {
			modeCntFreq[cnt] = make([]string, 0)
		}
		modeCntFreq[cnt] = append(modeCntFreq[cnt], mode)
	}
	for i := total - 1; i >= 0; i-- {
		if len(modeCntFreq[i]) > 0 {
			// return the highest count vlan, if there are multiple ones have the same count, just simply return the first one.
			logger.Info(fmt.Sprintf("majority-mode: %v", modeCntFreq[i][0]))
			majorityMode = modeCntFreq[i][0]
			break
		}
	}

	// if we cannot find any vlan info for these switchPorts, there might be something wrong.
	return majorityVlan, majorityMode, nil
}

func (b *BMHController) handleBMHDeleteEvents(ctx context.Context, hostName string) error {
	logger := log.FromContext(ctx).WithName("BMHController.handleBMHDeleteEvents")
	var err error

	b.BMHUsageReporter.RemoveBMH(hostName)

	existingNetworkNode := &idcnetworkv1alpha1.NetworkNode{}
	// NetworkNode has the same name as the bareMetalHost.
	key := types.NamespacedName{Name: hostName, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = b.networkK8sClient.Get(ctx, key, existingNetworkNode)
	nnExists := true
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// Object not found
		logger.Info("NetworkNode CR does not exist, no need to remove it. ", utils.LogFieldNetworkNode, hostName)
		nnExists = false
	}

	// try to remove this NetworkNode from a NodeGroup.
	if nnExists {
		if existingNetworkNode.Labels != nil {
			// get the NodeGroup from the NN label
			groupName, found := existingNetworkNode.Labels[idcnetworkv1alpha1.LabelGroupID]
			if found && len(groupName) > 0 {
				err = b.removeNetworkNodeAndSwitchesFromGroup(ctx, existingNetworkNode, groupName)
				if err != nil {
					logger.Error(err, "removeNetworkNodeAndSwitchesFromGroup failed")
					return err
				}
			}
		}
	} else {
		// TODO
		// it is possible that the NetowkrNode CR doesn't exist, but the NetworkNode name still exist in a NodeGroup's nodes list
		// the only way is to scan all the NodeGroup to find it.
	}

	// delete the NetworkNode CR
	if nnExists {
		err = b.networkK8sClient.Delete(ctx, existingNetworkNode, &client.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("Delete NetworkNode CR %v failed, reason: %v", existingNetworkNode.Name, err)
		}
	}
	logger.Info("Delete NetworkNode done", utils.LogFieldNetworkNode, hostName)
	return nil
}

type SwitchPortInfo struct {
	SwitchFQDN string
	SwitchPort string // example:
}

const (
	StorageAnnotationPrefix = "storage.mac.cloud.intel.com"
)

// findSwitchPortsFromBMH extract the front end and accelerator fabric ports from the BMH object.
func (b *BMHController) findSwitchPortsFromBMH(ctx context.Context, bareMetalHost *baremetalv1alpha1.BareMetalHost) (*SwitchPortInfo, []SwitchPortInfo, []SwitchPortInfo, error) {
	logger := log.FromContext(ctx).WithName("BMHController.findSwitchPortsFromBMH")
	frontEndSwitchPortInfo := &SwitchPortInfo{}
	AcceleratorFabricSwitchPortsInfo := make([]SwitchPortInfo, 0)
	StorageFabricSwitchPortsInfo := make([]SwitchPortInfo, 0)
	if bareMetalHost == nil || bareMetalHost.Status.HardwareDetails == nil {
		return nil, nil, nil, fmt.Errorf("bareMetalHost.Status.HardwareDetails has no value")
	}
	nicsInformation := bareMetalHost.Status.HardwareDetails.NIC
	if len(nicsInformation) < 1 {
		return nil, nil, nil, fmt.Errorf("failed to find nic information in host %s", bareMetalHost.Name)
	}

	finishedEnrollment := false
	bmhLabels := bareMetalHost.Labels
	for labelKey, _ := range bmhLabels {
		if strings.HasPrefix(labelKey, "instance-type.cloud.intel.com/") {
			finishedEnrollment = true
			break
		}
	}
	if !finishedEnrollment {
		return nil, nil, nil, fmt.Errorf("bareMetalHost has not finished enrollment yet. Will not create NetworkNode")
	}
	// extract gaudi gpu mac and storage mac from annotation
	gaudiGPUMacs := make(map[string]struct{})
	storageMacs := make(map[string]struct{})
	for key, val := range bareMetalHost.GetAnnotations() {
		if strings.HasPrefix(key, bmenrollment.GPUAnnotationPrefix) {
			gaudiGPUMacs[val] = struct{}{}
		} else if strings.HasPrefix(key, StorageAnnotationPrefix) {
			storageMacs[val] = struct{}{}
		}
	}

	for _, nic := range nicsInformation {
		isFrontendMac := nic.MAC == bareMetalHost.Spec.BootMACAddress
		_, isGPUMac := gaudiGPUMacs[nic.MAC]
		_, isStorageMac := storageMacs[nic.MAC]
		// ignore the switch ports with key word "edgecore" in the description
		if strings.Contains(strings.ToLower(nic.LLDP.SwitchSystemDescription), "edgecore") {
			continue
		}
		// Ignore. LLDP does not match any of the MACs in the BMH annotations, which means server is connected to a fabric that SDN doesn't manage.
		if !(isFrontendMac || isGPUMac || isStorageMac) {
			logger.V(1).Info("LLDP interface does not match any of the MACs in the BMH annotations", "MAC", nic.MAC, "BMH", bareMetalHost.Name)
			continue
		}

		// If we see some "missing" accelerator interfaces, skip for now. Validation below will catch if there is the wrong number of interfaces at the end.
		switchFQDN := nic.LLDP.SwitchSystemName
		if nic.LLDP.SwitchSystemName == "" {
			continue
		}
		if nic.LLDP.SwitchPortId == "" {
			continue
		}

		// fail if the switch ports have an INVALID switch FQDN (this switch will not be able to be controlled by the SDN)
		err := utils.ValidateSwitchFQDN(switchFQDN, "")
		if err != nil {
			return nil, nil, nil, fmt.Errorf("invalid switchFQDN (%v) found in BMH's (%v) LLDP info. %v ", switchFQDN, bareMetalHost.Name, err)
		}

		if isFrontendMac {
			frontEndSwitchPortInfo.SwitchFQDN = nic.LLDP.SwitchSystemName
			frontEndSwitchPortInfo.SwitchPort = nic.LLDP.SwitchPortId
		} else if isGPUMac {
			AcceleratorFabricSwitchPortsInfo = append(AcceleratorFabricSwitchPortsInfo, SwitchPortInfo{
				SwitchFQDN: nic.LLDP.SwitchSystemName,
				SwitchPort: nic.LLDP.SwitchPortId,
			})
		} else if isStorageMac {
			StorageFabricSwitchPortsInfo = append(StorageFabricSwitchPortsInfo, SwitchPortInfo{
				SwitchFQDN: nic.LLDP.SwitchSystemName,
				SwitchPort: nic.LLDP.SwitchPortId,
			})
		}
	}

	// front-end switch port has to be valid
	if len(frontEndSwitchPortInfo.SwitchFQDN) == 0 {
		logger.Error(fmt.Errorf("got empty Switch name from [%v] LLDP nic information for frontEnd fabric ", bareMetalHost.Name), "")
		return nil, nil, nil, fmt.Errorf("got empty Switch name from [%v] LLDP nic information for frontEnd fabric", bareMetalHost.Name)
	}
	if len(frontEndSwitchPortInfo.SwitchPort) == 0 {
		logger.Error(fmt.Errorf("got empty Port name from [%v] LLDP nic information for frontEnd fabric ", bareMetalHost.Name), "")
		return nil, nil, nil, fmt.Errorf("got empty Port name from [%v] LLDP nic information for frontEnd fabric ", bareMetalHost.Name)
	}

	// There should be a valid number of accelerator interfaces (eg. 0 or 24).
	allowedCountAccInterfaces := strings.Split(b.conf.ControllerConfig.AllowedCountAccInterfaces, ",")
	if !slices.Contains(allowedCountAccInterfaces, strconv.Itoa(len(AcceleratorFabricSwitchPortsInfo))) {
		logger.Error(fmt.Errorf("unexpected number (%v) of accelerator interfaces found in BMH %v", len(AcceleratorFabricSwitchPortsInfo), bareMetalHost.Name), "")
		return nil, nil, nil, fmt.Errorf("unexpected number (%v) of accelerator interfaces found in BMH %v", len(AcceleratorFabricSwitchPortsInfo), bareMetalHost.Name)
	}

	if len(AcceleratorFabricSwitchPortsInfo) != len(gaudiGPUMacs) && len(AcceleratorFabricSwitchPortsInfo) != 0 { // allow 0, or ALL interfaces.
		logger.Error(fmt.Errorf("number (%v) of accelerator interfaces did not match number of GPU annotations (%v) on BMH %v", len(AcceleratorFabricSwitchPortsInfo), len(gaudiGPUMacs), bareMetalHost.Name), "")
		return nil, nil, nil, fmt.Errorf("number (%v) of accelerator interfaces did not match number of GPU annotations (%v) on BMH %v", len(AcceleratorFabricSwitchPortsInfo), len(gaudiGPUMacs), bareMetalHost.Name)
	}

	return frontEndSwitchPortInfo, AcceleratorFabricSwitchPortsInfo, StorageFabricSwitchPortsInfo, nil
}
