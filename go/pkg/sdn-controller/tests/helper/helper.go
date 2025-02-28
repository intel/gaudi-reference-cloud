package helper

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"gopkg.in/yaml.v3"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/clab"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/mock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"

	sdnclient "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/client"
	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(idcnetworkv1alpha1.AddToScheme(scheme))
	utilruntime.Must(baremetalv1alpha1.AddToScheme(scheme))
}

type SDNTestHelper struct {
	ContainerLabManager *clab.ContainerLabManager
	MockBM              *mock.MockBM
	K8sClient           client.Client
	SDNClient           *sdnclient.SDNClient
	eAPISecretDir       string
}

type Config struct {
	TopologyDir    string
	NWCPKubeConfig string
	EAPISecretDir  string
}

func New(conf Config) *SDNTestHelper {
	ctx := context.Background()
	// create sdn client
	sdnClientConf := sdnclient.SDNClientConfig{
		KubeConfig: "",
		// KubeConfig: "../../../../../local/secrets/kubeconfig/kind-idc-us-dev-1a-network.yaml",
	}
	sdnClient, err := sdnclient.NewSDNClient(ctx, sdnClientConf)
	if err != nil {
		fmt.Printf("NewSDNClient error: %v \n", err)
	}

	clabMgr := clab.NewContainerLabManager(conf.TopologyDir, conf.EAPISecretDir)
	mockBM := mock.New(sdnClient)
	k8sClient := utils.NewK8SClientFromConfAndScheme(ctx, conf.NWCPKubeConfig, scheme)

	return &SDNTestHelper{
		ContainerLabManager: clabMgr,
		MockBM:              mockBM,
		K8sClient:           k8sClient,
		SDNClient:           sdnClient,
		eAPISecretDir:       conf.EAPISecretDir,
	}
}

func (p *SDNTestHelper) CreateK8sResourcesFromFile(crdyaml string) error {
	command := fmt.Sprintf("kubectl apply -f %s", crdyaml)

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl apply failed, Output: %s, error: %v\n", output, err)
		return fmt.Errorf("kubectl apply failed, Output: %s, error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteK8sResourcesFromFile(crdyaml string) error {
	command := fmt.Sprintf("kubectl delete --ignore-not-found=true -f %s", crdyaml)

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl delete failed: Output: %s, error: %v\n", output, err)
		return fmt.Errorf("kubectl delete failed, Output: %s, error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteAllEvents() error {
	command := fmt.Sprintf("kubectl delete event --all -n %s", idcnetworkv1alpha1.SDNControllerNamespace)

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl delete all events failed: output: %s error: %v\n", output, err)
		return fmt.Errorf("kubectl delete all events failed, Output: %s, error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteAllPortChannels() error {
	ctx := context.Background()

	// Get all portchannels & remove the finalizer
	allPortchannelCRs := &idcnetworkv1alpha1.PortChannelList{}
	err := p.K8sClient.List(ctx, allPortchannelCRs)
	if err != nil {
		return fmt.Errorf("list all portchannels failed, error: %v", err)
	}
	for _, portchannelCR := range allPortchannelCRs.Items {
		if len(portchannelCR.Finalizers) > 0 {
			bmhCRCopy := portchannelCR.DeepCopy()
			patch := client.MergeFrom(&portchannelCR)
			bmhCRCopy.Finalizers = nil
			if err := p.K8sClient.Patch(ctx, bmhCRCopy, patch); err != nil {
				fmt.Printf("Patch portchannelCR to remove finalizer failed: %v \n", err)
				continue
			}
		}
	}

	command := fmt.Sprintf("kubectl delete portchannel --all --force -n %s", idcnetworkv1alpha1.SDNControllerNamespace)
	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl delete all portchannels failed. Output: %s, error:  %v\n", output, err)
		return fmt.Errorf("kubectl delete all portchannels failed, output: %s error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteAllSwitches() error {
	command := fmt.Sprintf("kubectl delete switch --all -n %s", idcnetworkv1alpha1.SDNControllerNamespace)

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl delete all switches failed: output: %s error: %v\n", output, err)
		return fmt.Errorf("kubectl delete all switches failed, Output: %s, error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteAllNetworkNodes() error {
	command := fmt.Sprintf("kubectl delete networknode --all -n %s", idcnetworkv1alpha1.SDNControllerNamespace)

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl delete all networknode failed: output: %s error: %v\n", output, err)
		return fmt.Errorf("kubectl delete all networknode failed, Output: %s, error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteAllNodeGroupToPoolMappings() error {
	command := fmt.Sprintf("kubectl delete nodegrouptopoolmappings --all -n %s", idcnetworkv1alpha1.SDNControllerNamespace)

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl delete all nodegrouptopoolmappings failed: output: %s error: %v\n", output, err)
		return fmt.Errorf("kubectl delete all nodegrouptopoolmappings failed, Output: %s, error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteAllNodeGroups() error {
	command := fmt.Sprintf("kubectl delete nodegroup --all -n %s", idcnetworkv1alpha1.SDNControllerNamespace)

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl delete all nodegroup failed: output: %s error: %v\n", output, err)
		return fmt.Errorf("kubectl delete all nodegroup failed, Output: %s, error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteAllSwitchports() error {
	command := fmt.Sprintf("kubectl delete switchport --all -n %s", idcnetworkv1alpha1.SDNControllerNamespace)

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl delete all switchport failed: output: %s error: %v\n", output, err)
		return fmt.Errorf("kubectl delete all switchport failed, Output: %s, error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteAllBMH() error {
	ctx := context.Background()

	// Get all Bmhs & remove the finalizer
	allBmhCRs := &baremetalv1alpha1.BareMetalHostList{}
	err := p.K8sClient.List(ctx, allBmhCRs)
	if err != nil {
		return fmt.Errorf("list all BMHs failed, error: %v", err)
	}
	for _, BmhCR := range allBmhCRs.Items {
		if len(BmhCR.Finalizers) > 0 {
			bmhCRCopy := BmhCR.DeepCopy()
			patch := client.MergeFrom(&BmhCR)
			bmhCRCopy.Finalizers = nil
			if err := p.K8sClient.Patch(ctx, bmhCRCopy, patch); err != nil {
				fmt.Printf("Patch BmhCR to remove finalizer failed: %v \n", err)
				continue
			}
		}
	}

	command := fmt.Sprintf("kubectl delete bmh --all --all-namespaces")

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("kubectl delete all bmh failed: output: %s error: %v\n", output, err)
		return fmt.Errorf("kubectl delete all bmh failed, Output: %s, error: %v", output, err)
	}

	return nil
}

func (p *SDNTestHelper) DeleteAllK8sResources() error {
	multiErrors := &multierror.Error{}

	err := p.DeleteAllBMH()
	if err != nil {
		multiErrors = multierror.Append(multiErrors, err)
	}
	err = p.DeleteAllNodeGroups()
	if err != nil {
		multiErrors = multierror.Append(multiErrors, err)
	}
	err = p.DeleteAllNetworkNodes()
	if err != nil {
		multiErrors = multierror.Append(multiErrors, err)
	}
	err = p.DeleteAllSwitchports()
	if err != nil {
		multiErrors = multierror.Append(multiErrors, err)
	}
	err = p.DeleteAllSwitches()
	if err != nil {
		multiErrors = multierror.Append(multiErrors, err)
	}
	err = p.DeleteAllPortChannels()
	if err != nil {
		multiErrors = multierror.Append(multiErrors, err)
	}
	err = p.DeleteAllNodeGroupToPoolMappings()
	if err != nil {
		multiErrors = multierror.Append(multiErrors, err)
	}

	if len(multiErrors.Errors) > 0 {
		return multiErrors
	} else {
		return nil
	}
}

// CreateK8sResourcesForTopology creates the BMH and set the status to ready immediately, so this function should be used for test cases that don't care about the BM enrollment process.
func (p *SDNTestHelper) CreateK8sResourcesForTopology(topology *clab.Topology) error {
	ctx := context.Background()
	bmhs, switches, mappings, err := p.GenerateK8sResourcesFromClabTopology(topology)
	if err != nil {
		return fmt.Errorf("GenerateK8sResourcesFromTopology failed, %v", err)
	}

	return p.CreateK8sResources(ctx, bmhs, switches, mappings, true)
}

func (p *SDNTestHelper) CreateK8sResources(ctx context.Context, bmhs map[string]*baremetalv1alpha1.BareMetalHost, switches map[string]*idcnetworkv1alpha1.Switch, mappings map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping, waitForNNsCreation bool) error {
	logger := log.FromContext(ctx)
	logger.Info("Creating k8s resources...")
	// create the K8s resources
	for _, sw := range switches {
		err := p.K8sClient.Create(ctx, sw, &client.CreateOptions{})
		if err != nil {
			fmt.Printf("Create Switches failed: %v \n", err)
		}
	}

	for _, mapping := range mappings {
		err := p.K8sClient.Create(ctx, mapping, &client.CreateOptions{})
		if err != nil {
			fmt.Printf("Create Mappings failed: %v \n", err)
		}
	}

	for _, bmh := range bmhs {

		bmhCopy := bmh.DeepCopy()

		// create the BMH CRs
		err := p.K8sClient.Create(ctx, bmh, &client.CreateOptions{})
		if err != nil {
			fmt.Printf("Create BMH failed: %v \n", err)
		}

		// get the latest BMH and update the status
		latestBMH := &baremetalv1alpha1.BareMetalHost{}
		key := types.NamespacedName{Name: bmh.Name, Namespace: bmh.Namespace}

		err = p.K8sClient.Get(ctx, key, latestBMH)
		if err != nil {
			fmt.Printf("Get BMH failed: %v \n", err)
		}

		latestBMH.Status = bmhCopy.Status
		err = p.K8sClient.Status().Update(ctx, latestBMH)
		if err != nil {
			fmt.Printf("BMH Status().Update() failed: %v \n", err)
		}
	}

	if waitForNNsCreation {
		// wait until all the CRs(at least the NNs) have been created, this is to make it deterministic for any following steps that rely on the CRs generated from the BMH.
		logger.Info("Waiting until CRs have been created by SDN Controller...")
		resCh := make(chan struct{})
		go func() {
		OUTER:
			for {
				// check if NNs are created
				for _, bmh := range bmhs {
					nnCR := &idcnetworkv1alpha1.NetworkNode{}
					key := types.NamespacedName{Name: bmh.Name, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
					err := p.K8sClient.Get(ctx, key, nnCR)
					if err != nil {
						// fmt.Printf("NetworkNode are not ready yet, %v \n", err)
						continue OUTER
					}
				}
				logger.Info("NetworkNodes are created")
				resCh <- struct{}{}
			}
		}()
		select {
		case <-time.After(60 * time.Second):
			return fmt.Errorf("NetworkNode CRD creation timeout")
		case _, ok := <-resCh:
			if !ok {
				return fmt.Errorf("result channel is closed")
			}
			logger.Info("CRD creation completed")
			break
		}
	}

	return nil
}

// UpdateBMHAndStatus is a convenience-method for updating the BMH and updating the status at the same time.
// Status conflicts are not checked, the passed-in status will be applied over the top of the current CR.
func (p *SDNTestHelper) UpdateBMHAndStatus(ctx context.Context, bmhCR *baremetalv1alpha1.BareMetalHost) error {
	bmhCopy := bmhCR.DeepCopy() // Must copy because call to .Update() overwrites the .Status on bmhCR

	// First update the main CR.
	err := p.K8sClient.Update(ctx, bmhCR)
	if err != nil {
		return err
	}

	// Fetch the new CR so that its status can be updated.
	bmhKey := types.NamespacedName{Name: bmhCR.Name, Namespace: bmhCR.Namespace}
	latestBmhCR := &baremetalv1alpha1.BareMetalHost{}
	err = p.K8sClient.Get(ctx, bmhKey, latestBmhCR)
	if err != nil {
		return err
	}
	latestBmhCR.Status = bmhCopy.Status // Copy status from object that was passed in to the latest version

	// Finally update the status
	err = p.K8sClient.Status().Update(ctx, latestBmhCR)
	if err != nil {
		return err
	}

	return nil
}

func (p *SDNTestHelper) WaitForNodeGroupReady(nodeGroup string, waitForNodeCount int, waitForPool bool) error {
	ctx := context.Background()

	resCh := make(chan struct{})
	notReadyReason := ""
	go func() {
	OUTER:
		for {
			// check if nodeGroup is created & ready
			ngCR := &idcnetworkv1alpha1.NodeGroup{}
			key := types.NamespacedName{Name: nodeGroup, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
			err := p.K8sClient.Get(ctx, key, ngCR)
			if err != nil {
				notReadyReason = fmt.Sprintf("NodeGroup %s not ready yet, could not fetch CR from k8s: %v \n", nodeGroup, err)
				time.Sleep(100 * time.Millisecond)
				continue OUTER
			}
			if waitForPool && ngCR.Labels[idcnetworkv1alpha1.LabelPool] == "" {
				notReadyReason = fmt.Sprintf("NodeGroup %s not ready yet, not moved in to a pool \n", nodeGroup)
				time.Sleep(100 * time.Millisecond)
				continue OUTER
			}
			if ngCR.Labels[idcnetworkv1alpha1.LabelMaintenance] != "" {
				notReadyReason = fmt.Sprintf("NodeGroup %s not ready yet, in maintenance mode %s \n", nodeGroup, ngCR.Labels[idcnetworkv1alpha1.LabelMaintenance])
				time.Sleep(100 * time.Millisecond)
				continue OUTER
			}
			if ngCR.Status.AcceleratorFabricStatus != nil && ngCR.Status.AcceleratorFabricStatus.BGPConfigStatus.Ready != true {
				notReadyReason = fmt.Sprintf("NodeGroup %s some Status.AcceleratorFabricStatus.BGPConfigStatus.Ready not yet true, %v \n", nodeGroup, ngCR.Status.AcceleratorFabricStatus.BGPConfigStatus.Ready)
				time.Sleep(100 * time.Millisecond)
				continue OUTER
			}
			if ngCR.Status.FrontEndFabricStatus != nil && ngCR.Status.FrontEndFabricStatus.VlanConfigStatus.Ready != true {
				notReadyReason = fmt.Sprintf("NodeGroup %s some Status.FrontEndFabricStatus.VlanConfigStatus.Ready not yet true, %v \n", nodeGroup, ngCR.Status.FrontEndFabricStatus.VlanConfigStatus.Ready)
				time.Sleep(100 * time.Millisecond)
				continue OUTER
			}
			if ngCR.Status.StorageFabricStatus != nil && ngCR.Status.StorageFabricStatus.VlanConfigStatus.Ready != true {
				notReadyReason = fmt.Sprintf("NodeGroup %s some Status.StorageFabricStatus.VlanConfigStatus.Ready not yet true, %v \n", nodeGroup, ngCR.Status.StorageFabricStatus.VlanConfigStatus.Ready)
				time.Sleep(100 * time.Millisecond)
				continue OUTER
			}
			if ngCR.Status.NetworkNodesCount != waitForNodeCount {
				notReadyReason = fmt.Sprintf("Waiting for NodeGroup %s to get %d nodes (has %d) \n", nodeGroup, waitForNodeCount, ngCR.Status.NetworkNodesCount)
				time.Sleep(100 * time.Millisecond)
				continue OUTER
			}
			fmt.Printf("NodeGroup %s ready \n", nodeGroup)
			resCh <- struct{}{}
		}
	}()
	select {
	case <-time.After(60 * time.Second):
		return fmt.Errorf("timeout waiting for NodeGroup ready. nodeGroup: %s Reason: %s", nodeGroup, notReadyReason)
	case _, ok := <-resCh:
		if !ok {
			return fmt.Errorf("result channel is closed")
		}
		fmt.Println("NodeGroups creation completed")
		break
	}

	return nil
}

func (p *SDNTestHelper) DeleteK8sResourcesForClabTopology(topology *clab.Topology) error {
	ctx := context.Background()
	bmhs, sws, mappings, err := p.GenerateK8sResourcesFromClabTopology(topology)
	if err != nil {
		return fmt.Errorf("GenerateK8sResourcesFromTopology failed, %v", err)
	}

	return p.DeleteK8sResources(ctx, bmhs, sws, mappings)
}

func (p *SDNTestHelper) DeleteK8sResources(ctx context.Context, bmhs map[string]*baremetalv1alpha1.BareMetalHost, sws map[string]*idcnetworkv1alpha1.Switch, mappings map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping) error {
	for _, bmh := range bmhs {
		// try to remove the finalizer
		bmhCR := &baremetalv1alpha1.BareMetalHost{}
		key := types.NamespacedName{Name: bmh.Name, Namespace: bmh.Namespace}
		err := p.K8sClient.Get(ctx, key, bmhCR)
		if err != nil {
			fmt.Printf("get BMH failed: %v \n", err)
			continue
		}

		if len(bmhCR.Finalizers) > 0 {
			bmhCRCopy := bmhCR.DeepCopy()
			patch := client.MergeFrom(bmhCR)
			bmhCRCopy.Finalizers = nil
			if err := p.K8sClient.Patch(ctx, bmhCRCopy, patch); err != nil {
				fmt.Printf("Patch BMH failed: %v \n", err)
				continue
			}
		}

		err = p.K8sClient.Delete(ctx, bmh, &client.DeleteOptions{})
		if err != nil {
			fmt.Printf("Delete BMH failed: %v \n", err)
		}
	}

	for _, mapping := range mappings {
		err := p.K8sClient.Delete(ctx, mapping, &client.DeleteOptions{})
		if err != nil {
			fmt.Printf("Delete mappings failed: %v \n", err)
		}

		err = p.K8sClient.Delete(ctx, &idcnetworkv1alpha1.NodeGroup{
			ObjectMeta: v1.ObjectMeta{
				Name:      mapping.Name,
				Namespace: idcnetworkv1alpha1.SDNControllerNamespace,
			},
		})
		if err != nil {
			fmt.Printf("Delete nodeGroup failed: %v \n", err)
		}
	}

	for _, sw := range sws {
		err := p.K8sClient.Delete(ctx, sw, &client.DeleteOptions{})
		if err != nil {
			fmt.Printf("Delete Switches failed: %v \n", err)
		}
	}

	// Wait until all entities have been deleted.
	found := false
	for i := 0; i < 10; i++ {
		found = false
		for _, bmh := range bmhs {
			bmhCR := &baremetalv1alpha1.BareMetalHost{}
			key := types.NamespacedName{Name: bmh.Name, Namespace: bmh.Namespace}
			err := p.K8sClient.Get(ctx, key, bmhCR)
			if err == nil {
				found = true
			}
		}
		for _, mapping := range mappings {
			CR := &idcnetworkv1alpha1.NodeGroupToPoolMapping{}
			key := types.NamespacedName{Name: mapping.Name, Namespace: mapping.Namespace}
			err := p.K8sClient.Get(ctx, key, CR)
			if err == nil {
				found = true
			}
		}
		for _, sw := range sws {
			CR := &idcnetworkv1alpha1.Switch{}
			key := types.NamespacedName{Name: sw.Name, Namespace: sw.Namespace}
			err := p.K8sClient.Get(ctx, key, CR)
			if err == nil {
				found = true
			}
		}

		if found == false {
			fmt.Printf("All k8s resources deleted for topology\n")
			break
		} else {
			fmt.Printf("%d: still waiting for k8s resources to be deleted...\n", i)
			time.Sleep(1 * time.Second)
		}
	}

	if found {
		return fmt.Errorf("failed to delete all k8s resources")
	}

	return nil
}

// GenerateK8sResourcesFromClabTopology helps extract the k8s resources from a clab topology
func (p *SDNTestHelper) GenerateK8sResourcesFromClabTopology(topology *clab.Topology) (map[string]*baremetalv1alpha1.BareMetalHost, map[string]*idcnetworkv1alpha1.Switch, map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping, error) {

	nodes := topology.TopologySpec.Nodes
	links := topology.TopologySpec.Links

	nodeNameToNode := make(map[string]*clab.Node)
	// given ["accply1-leaf1:eth5", "server1-1:eth1"]
	// store it as "server1-1" : "accply1-leaf1:eth5"
	serverToFESwitchAndPortMap := make(map[string]map[string]struct{})
	serverToACCSwitchAndPortMap := make(map[string]map[string]struct{})
	serverToStorageSwitchAndPortMap := make(map[string]map[string]struct{})

	serverNameToGroup := make(map[string]string)

	for nodeName, node := range nodes {
		nodeNameToNode[nodeName] = node
		if clab.IsServerByName(nodeName) {
			groupName := extracGroupNameFromServerName(nodeName)
			serverNameToGroup[nodeName] = groupName
		}
	}

	for _, link := range links {
		if len(link.Endpoints) < 1 {
			continue
		}
		left := link.Endpoints[0]
		right := link.Endpoints[1]

		if strings.Contains(left, "server") {
			// if server is on the left side
			leftNodeName := strings.Split(left, ":")[0]

			if strings.Contains(right, "frontend") {
				if serverToFESwitchAndPortMap[leftNodeName] == nil {
					serverToFESwitchAndPortMap[leftNodeName] = map[string]struct{}{}
				}
				serverToFESwitchAndPortMap[leftNodeName][right] = struct{}{}
			} else if strings.Contains(right, "acc") {
				if serverToACCSwitchAndPortMap[leftNodeName] == nil {
					serverToACCSwitchAndPortMap[leftNodeName] = map[string]struct{}{}
				}
				serverToACCSwitchAndPortMap[leftNodeName][right] = struct{}{}
			} else if strings.Contains(right, "storage") {
				if serverToStorageSwitchAndPortMap[leftNodeName] == nil {
					serverToStorageSwitchAndPortMap[leftNodeName] = map[string]struct{}{}
				}
				serverToStorageSwitchAndPortMap[leftNodeName][right] = struct{}{}
			}
		} else if strings.Contains(right, "server") {
			// if server is on the right side
			rightNodeName := strings.Split(right, ":")[0]
			if strings.Contains(left, "frontend") {
				if serverToFESwitchAndPortMap[rightNodeName] == nil {
					serverToFESwitchAndPortMap[rightNodeName] = map[string]struct{}{}
				}
				serverToFESwitchAndPortMap[rightNodeName][left] = struct{}{}
			} else if strings.Contains(left, "acc") {
				if serverToACCSwitchAndPortMap[rightNodeName] == nil {
					serverToACCSwitchAndPortMap[rightNodeName] = map[string]struct{}{}
				}
				serverToACCSwitchAndPortMap[rightNodeName][left] = struct{}{}
			} else if strings.Contains(left, "storage") {
				if serverToStorageSwitchAndPortMap[rightNodeName] == nil {
					serverToStorageSwitchAndPortMap[rightNodeName] = map[string]struct{}{}
				}
				serverToStorageSwitchAndPortMap[rightNodeName][left] = struct{}{}
			}
		}
	}

	bmhs := make(map[string]*baremetalv1alpha1.BareMetalHost)
	sws := make(map[string]*idcnetworkv1alpha1.Switch)
	groupToPoolMappings := make(map[string]*idcnetworkv1alpha1.NodeGroupToPoolMapping)
	defaultGroupPool, found := topology.TopologySpec.Defaults.Labels["groupPool"]
	if !found {
		defaultGroupPool = "none" // Default default
	}
	// TODO: Support overriding pool per-node. Requires validation that no 2 nodes in the same group have different pools.

	for serverName, groupName := range serverNameToGroup {
		// generate the nodeGroupToPool mapping objects
		if defaultGroupPool != "none" {
			groupToPoolMappings[groupName] = &idcnetworkv1alpha1.NodeGroupToPoolMapping{
				ObjectMeta: v1.ObjectMeta{
					Name:      groupName,
					Namespace: "idcs-system",
				},
				Spec: idcnetworkv1alpha1.NodeGroupToPoolMappingSpec{
					Pool: defaultGroupPool,
				},
			}
		}

		/////////////////////////////////////
		// generate FE network configs
		/////////////////////////////////////
		feBootMAC := testutils.GenerateRandomMAC()
		switchAndPorts, found := serverToFESwitchAndPortMap[serverName]
		if !found {
			fmt.Printf("Did not find FE Switch for server %v \n", serverName)
			continue
		}
		var feSwitchShortName string
		var fePortShortName string
		// in most case, there should be only one FE link
		for serverToFESwitcgAndPort, _ := range switchAndPorts {
			SwitchAndPortNames := strings.Split(serverToFESwitcgAndPort, ":")
			if len(SwitchAndPortNames) != 2 {
				continue
			}

			feSwitchShortName = SwitchAndPortNames[0]
			fePortShortName = SwitchAndPortNames[1]
		}

		feSwitchFQDN := testutils.ConvertSWShortNameToFQDN(feSwitchShortName, topology.Name)
		fePortName, err := testutils.ConvertShortPortNameToLongFormat(fePortShortName)
		if err != nil {
			fmt.Printf("ConvertShortPortNameToLongFormat failed, %v \n", err)
			continue
		}

		ip := feSwitchFQDN
		node, found := nodeNameToNode[feSwitchShortName]
		if found {
			ip = node.IPv4Address
		}
		fesw := &idcnetworkv1alpha1.Switch{
			ObjectMeta: v1.ObjectMeta{
				Name:      feSwitchFQDN,
				Namespace: "idcs-system",
			},
			Spec: idcnetworkv1alpha1.SwitchSpec{
				FQDN: feSwitchFQDN,
				Ip:   ip,
			},
		}
		sws[feSwitchFQDN] = fesw

		bmh := &baremetalv1alpha1.BareMetalHost{

			ObjectMeta: v1.ObjectMeta{
				Name:        serverName,
				Namespace:   "metal3-1",
				Annotations: make(map[string]string),
				Labels: map[string]string{
					idcnetworkv1alpha1.LabelBMHGroupID:               groupName,
					"instance-type.cloud.intel.com/mock-bmh-sdn-e2e": "true",
				},
			},
			Spec: baremetalv1alpha1.BareMetalHostSpec{
				BootMACAddress: feBootMAC,
			},
			Status: baremetalv1alpha1.BareMetalHostStatus{
				HardwareDetails: &baremetalv1alpha1.HardwareDetails{
					NIC: []baremetalv1alpha1.NIC{
						baremetalv1alpha1.NIC{
							Name: fePortName,
							MAC:  feBootMAC,
							LLDP: baremetalv1alpha1.LLDP{
								SwitchPortId:     fePortName,
								SwitchSystemName: feSwitchFQDN,
							},
						},
					},
				},
				OperationalStatus: baremetalv1alpha1.OperationalStatusOK,
			},
		}

		/////////////////////////////////////
		// generate ACC network configs
		/////////////////////////////////////
		accLinks := serverToACCSwitchAndPortMap[serverName]
		accelSPCnt := 0
		for accLink, _ := range accLinks {
			// accply1-leaf1:eth1
			SwitchAndPortNames := strings.Split(accLink, ":")
			if len(SwitchAndPortNames) != 2 {
				continue
			}
			accSwitchShortName := SwitchAndPortNames[0]
			accPortShortName := SwitchAndPortNames[1]
			accSwitchFQDN := testutils.ConvertSWShortNameToFQDN(accSwitchShortName, topology.Name)
			accPortName, err := testutils.ConvertShortPortNameToLongFormat(accPortShortName)
			if err != nil {
				continue
			}

			ip := accSwitchFQDN
			node, found := nodeNameToNode[accSwitchShortName]
			if found {
				ip = node.IPv4Address
			}
			accsw := &idcnetworkv1alpha1.Switch{
				ObjectMeta: v1.ObjectMeta{
					Name:      accSwitchFQDN,
					Namespace: "idcs-system",
				},
				Spec: idcnetworkv1alpha1.SwitchSpec{
					FQDN: accSwitchFQDN,
					Ip:   ip,
				},
			}
			sws[accSwitchFQDN] = accsw

			mac := testutils.GenerateRandomMAC()
			gpu := fmt.Sprintf("gpu.mac.cloud.intel.com/gpu-%v", accelSPCnt)
			bmh.Annotations[gpu] = mac
			bmh.Status.HardwareDetails.NIC = append(bmh.Status.HardwareDetails.NIC, baremetalv1alpha1.NIC{
				Name: accPortName,
				MAC:  mac,
				LLDP: baremetalv1alpha1.LLDP{
					SwitchPortId:     accPortName,
					SwitchSystemName: accSwitchFQDN,
				},
			})
			accelSPCnt++
		}

		/////////////////////////////////////
		// generate Storage network configs
		/////////////////////////////////////
		strgLinks := serverToStorageSwitchAndPortMap[serverName]
		strgSPCnt := 0
		for strgLink, _ := range strgLinks {

			SwitchAndPortNames := strings.Split(strgLink, ":")
			if len(SwitchAndPortNames) != 2 {
				continue
			}
			strgSwitchShortName := SwitchAndPortNames[0]
			strgPortShortName := SwitchAndPortNames[1]
			strgSwitchFQDN := testutils.ConvertSWShortNameToFQDN(strgSwitchShortName, topology.Name)
			strgPortName, err := testutils.ConvertShortPortNameToLongFormat(strgPortShortName)
			if err != nil {
				continue
			}

			ip := strgSwitchFQDN
			node, found := nodeNameToNode[strgSwitchShortName]
			if found {
				ip = node.IPv4Address
			}

			strgsw := &idcnetworkv1alpha1.Switch{
				ObjectMeta: v1.ObjectMeta{
					Name:      strgSwitchFQDN,
					Namespace: "idcs-system",
				},
				Spec: idcnetworkv1alpha1.SwitchSpec{
					FQDN: strgSwitchFQDN,
					Ip:   ip,
				},
			}
			sws[strgSwitchFQDN] = strgsw

			mac := testutils.GenerateRandomMAC()
			gpu := fmt.Sprintf("storage.mac.cloud.intel.com/storage-eth-%v", strgSPCnt)
			bmh.Annotations[gpu] = mac
			bmh.Status.HardwareDetails.NIC = append(bmh.Status.HardwareDetails.NIC, baremetalv1alpha1.NIC{
				Name: strgPortName,
				MAC:  mac,
				LLDP: baremetalv1alpha1.LLDP{
					SwitchPortId:     strgPortName,
					SwitchSystemName: strgSwitchFQDN,
				},
			})
		}

		bmhs[bmh.Name] = bmh
	}

	return bmhs, sws, groupToPoolMappings, nil
}

func extracGroupNameFromServerName(nodeName string) string {
	if clab.IsServerByName(nodeName) {
		strs := strings.Split(strings.TrimPrefix(nodeName, clab.ServerNodePrefix), "-")
		if len(strs) == 2 {
			return strs[0]
		}
	}
	return ""
}

const (
	ConfigCheckShouldIgnore = `{{SKIP}}`
)

func (p *SDNTestHelper) CompareConfigs(expected, actual string) bool {
	expectedLines := strings.Split(expected, "!")
	actualLines := strings.Split(actual, "!")
	if len(expectedLines) != len(actualLines) {
		fmt.Printf("running-config number of lines mismatch. expectedLines[%v], actualLines[%v] \n %s \n", len(expectedLines), len(actualLines), actual)
		return false
	}

	diffCnt := 0
	skipCnt := 0
	for i := 0; i < len(expectedLines) && i < len(actualLines); i++ {
		if strings.Contains(expectedLines[i], ConfigCheckShouldIgnore) {
			skipCnt++
			continue
		}
		if expectedLines[i] != actualLines[i] {
			fmt.Printf("======== failed! unexpected switch config result observed ======== . \nexpected: \n%s \nactual: \n%s \n", expectedLines[i], actualLines[i])
			diffCnt++
		}
	}

	fmt.Printf("total [%d] lines are skipped \n", skipCnt)
	fmt.Printf("total [%d] lines have unexpected value \n", diffCnt)
	return diffCnt == 0
}

func (p *SDNTestHelper) DeploySDNWithConfig(configYaml string, deployRestAPI bool) error {
	ctx := context.Background()

	err := os.WriteFile("../../../../../deployment/helmfile/environments/test-e2e-sdn.yaml.gotmpl", []byte(configYaml), 0644)
	if err != nil {
		panic(err)
	}

	return p.deploySDN(ctx, deployRestAPI)
}

func (p *SDNTestHelper) DeploySDNWithConfigFile(configfile string, deployRestAPI bool) error {
	ctx := context.Background()

	cpCmd := fmt.Sprintf("cp %s ../../../../../deployment/helmfile/environments/test-e2e-sdn.yaml.gotmpl", configfile)
	fmt.Println(cpCmd)
	output, err := testutils.RunCommand(cpCmd)
	if err != nil {
		fmt.Println(output)
		return err
	}

	return p.deploySDN(ctx, deployRestAPI)
}

func (p *SDNTestHelper) deploySDN(ctx context.Context, deployRestAPI bool) error {
	logger := log.FromContext(ctx)
	logger.Info("Building & deploying SDN controller...")
	output, err := testutils.RunCommand("cd ../../../../.. && IDC_ENV=test-e2e-sdn make deploy-only-sdn-controller")
	if err != nil {
		fmt.Println(output)
		return err
	}

	if deployRestAPI { // Also requires "restAPI.enabled: true" to be set in the configFile. The deployRestAPI just skips the build for tests that don't need it.
		output, err = testutils.RunCommand("cd ../../../../.. && IDC_ENV=test-e2e-sdn make deploy-only-sdn-controller-rest")
		if err != nil {
			fmt.Println(output)
			return err
		}
	}

	// Wait for SDN-controller to go to "Running" state.
	err = p.waitForPodRunning(ctx, "sdn-controller")
	if err != nil {
		return err
	}

	if deployRestAPI {
		err = p.waitForPodRunning(ctx, "sdn-controller-rest")
		if err != nil {
			return err
		}
	}

	logger.Info("SDN controller deployed.")
	return nil
}

func (p *SDNTestHelper) waitForPodRunning(ctx context.Context, podName string) error {
	running := false
	for i := 0; i < 60; i++ {
		err, pod := p.GetPod(ctx, podName)
		if err != nil {
			fmt.Printf("%d: %v\n", i, err)
			time.Sleep(time.Second)
			continue
		}
		if pod.Status.Phase != "Running" {
			fmt.Printf("%d: %s pod is \"%s\" (not \"Running\")\n", i, podName, pod.Status.Phase)
			time.Sleep(time.Second)
			continue
		}

		// Running.
		running = true
		fmt.Printf("%d: %s pod is \"%s\"\n", i, podName, pod.Status.Phase)
		break
	}

	if !running {
		return fmt.Errorf("%s pod did not go to \"Running\" state", podName)
	}

	return nil
}

func (p *SDNTestHelper) ListSwitchPortEvents(ctx context.Context, switchPortName, namespace string) (*corev1.EventList, error) {
	events := &corev1.EventList{}

	err := p.K8sClient.List(ctx, events, client.InNamespace(namespace), client.MatchingFields{"involvedObject.kind": "SwitchPort", "involvedObject.name": switchPortName})
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (p *SDNTestHelper) ListSwitchEvents(ctx context.Context, switchName, namespace string) (*corev1.EventList, error) {
	events := &corev1.EventList{}

	err := p.K8sClient.List(ctx, events, client.InNamespace(namespace), client.MatchingFields{"involvedObject.kind": "Switch", "involvedObject.name": switchName})
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (p *SDNTestHelper) WaitForSwitchportEvent(ctx context.Context, switchPortName string, namespace string, desiredEventMessage string, timeoutSecs int) (bool, error) {

	for k := 0; k < timeoutSecs; k++ {
		events, err := p.ListSwitchPortEvents(ctx, switchPortName, namespace)

		if len(events.Items) == 0 || err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Sort events by creation timestamp in descending order
		sort.Slice(events.Items, func(i, j int) bool {
			return events.Items[i].CreationTimestamp.Time.After(events.Items[j].CreationTimestamp.Time)
		})

		// Check the message of the events
		for i := range events.Items {
			event := events.Items[i]
			if strings.Contains(event.Message, desiredEventMessage) {
				return true, nil
			}
		}
		time.Sleep(1 * time.Second)
	}

	return false, nil
}

func (p *SDNTestHelper) WaitForSwitchEvent(ctx context.Context, switchName string, namespace string, desiredEventMessage string, timeoutSecs int) (bool, error) {

	for k := 0; k < timeoutSecs; k++ {
		events, err := p.ListSwitchEvents(ctx, switchName, namespace)

		if len(events.Items) == 0 || err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Sort events by creation timestamp in descending order
		sort.Slice(events.Items, func(i, j int) bool {
			return events.Items[i].CreationTimestamp.Time.After(events.Items[j].CreationTimestamp.Time)
		})

		// Check the message of the events
		for i := range events.Items {
			event := events.Items[i]
			if strings.Contains(event.Message, desiredEventMessage) {
				return true, nil
			}
		}
		time.Sleep(1 * time.Second)
	}

	return false, nil
}

func (p *SDNTestHelper) CheckSDNIsRunningWithTestTenantConfig() error {
	ctx := context.Background()

	// Check SDN-controller is running with the expected config.
	err, sdnPod := p.GetPod(ctx, "sdn-controller")
	if err != nil {
		return err
	}
	if sdnPod.Status.Phase != "Running" {
		return fmt.Errorf("SDN Controller pod is \"%s\" (not \"Running\")", sdnPod.Status.Phase)
	}

	//Check config.
	sdnConfigMap := &corev1.ConfigMap{}
	key := types.NamespacedName{Name: "sdn-controller-manager-config", Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
	err = p.K8sClient.Get(ctx, key, sdnConfigMap)
	if err != nil {
		return err
	}

	cfg := &idcnetworkv1alpha1.SDNControllerConfig{}
	err = yaml.Unmarshal([]byte(sdnConfigMap.Data["controller_manager_config.yaml"]), cfg)
	if cfg.ControllerConfig.SwitchPortImportSource != "bmh" {
		return fmt.Errorf("prerequisite config SwitchPortImportSource expected to be \"bmh\". Was: %v", cfg.ControllerConfig.SwitchPortImportSource)
	}
	if !strings.Contains(cfg.ControllerConfig.DataCenter, "clab") {
		return fmt.Errorf("prerequisite config SwitchPortImportSource did not contain 'clab'. Was: %v", cfg.ControllerConfig.DataCenter)
	}
	if cfg.ControllerConfig.NodeGroupToPoolMappingSource != "crd" {
		return fmt.Errorf("prerequisite config NodeGroupToPoolMappingSource expected to be \"crd\". Was: %v", cfg.ControllerConfig.NodeGroupToPoolMappingSource)
	}
	if cfg.ControllerConfig.SwitchBackendMode != "eapi" {
		return fmt.Errorf("prerequisite config SwitchBackendMode expected to be \"eapi\". Was: %v", cfg.ControllerConfig.SwitchBackendMode)
	}
	if cfg.ControllerConfig.EnableReadOnlyMode != false {
		return fmt.Errorf("prerequisite config EnableReadOnlyMode expected to be false. Was: %v", cfg.ControllerConfig.EnableReadOnlyMode)
	}

	return nil
}

func (p *SDNTestHelper) GetContainerRestarts(ctx context.Context, podName string, containerName string) (error, int32) {
	err, pod := p.GetPod(ctx, podName)
	if err != nil {
		return err, 0
	}
	for i := range pod.Status.ContainerStatuses {
		ctr := pod.Status.ContainerStatuses[i]
		if ctr.Name == containerName {
			return nil, ctr.RestartCount
		}
	}
	return fmt.Errorf("container %s not found in pod %s \n", containerName, podName), 0
}

func (p *SDNTestHelper) GetPod(ctx context.Context, podName string) (error, *corev1.Pod) {
	pods := &corev1.PodList{}
	err := p.K8sClient.List(ctx, pods, client.InNamespace(idcnetworkv1alpha1.SDNControllerNamespace), client.MatchingLabels(map[string]string{"app.kubernetes.io/name": podName}))
	if err != nil {
		return err, nil
	}
	if len(pods.Items) < 1 {
		return fmt.Errorf("%s pod not yet found in the k8s cluster\n", podName), nil
	}
	if len(pods.Items) > 1 {
		return fmt.Errorf("multiple %s pods found running in the k8s cluster\n", podName), nil
	}
	pod := pods.Items[0]
	return nil, &pod
}

func (p *SDNTestHelper) PingInsideContainer(containerName, targetIP string, maxRetries int) bool {
	for i := 0; i < maxRetries; i++ {
		fmt.Printf("pinging %s from %s (attempt %d)...", targetIP, containerName, i+1)
		output, err := testutils.ExecCommandInContainer(containerName, []string{"ping", "-c", "4", targetIP})
		if err != nil {
			fmt.Printf("ExecCommandInContainer failed, %v \n", err)
		} else {
			if strings.Contains(output, "0% packet loss") {
				fmt.Printf("ping successful!\n")
				return true
			}
			fmt.Printf("ping output: %v \n", output)
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("ping failed after %d attempts \n", maxRetries)
	return false
}
