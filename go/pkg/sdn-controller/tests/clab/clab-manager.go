package clab

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"os"
	"path/filepath"
	"time"

	"io/ioutil"
	"strings"

	switchclients "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	testutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/tests/test_utils"
	"gopkg.in/yaml.v2"
)

const (
	ServerNodePrefix     = "server"
	FrontendSwitchPrefix = "frontend"
	ACCSwitchPrefix      = "acc"
	StorageSwitchPrefix  = "storage"
)

type Topology struct {
	Name         string        `yaml:"name"`
	TopologySpec *TopologySpec `yaml:"topology"`
}

type TopologySpec struct {
	Nodes    map[string]*Node `yaml:"nodes"`
	Links    []*Link          `yaml:"links"`
	Defaults TopologyDefaults `yaml:"defaults"`
}

type TopologyDefaults struct {
	Labels map[string]string `yaml:"labels"`
	Env    map[string]string `yaml:"env"`
}

type Node struct {
	Kind          string `yaml:"kind"`
	Image         string `yaml:"image"`
	StartupConfig string `yaml:"startup-config"`

	// custom field
	Number      string
	Name        string
	FQDN        string
	ContainerID string
	State       string
	IPv4Address string
	IPv6Address string
}

type Link struct {
	Endpoints []string `yaml:"endpoints"`
}

const (
	DefaultTopologyDir = "../../../../../../networking/containerlab/"
)

type ContainerLabManager struct {
	TopologiesDir string
	Topologies    map[string]*Topology
	// topologyName:switchFQDN:switchClient
	SwitchClients map[string]map[string]*switchclients.AristaClient
	EAPISecretDir string
	Logger        logr.Logger
}

func NewContainerLabManager(topologiesDir string, eAPISecretDir string) *ContainerLabManager {
	if len(topologiesDir) == 0 {
		topologiesDir = DefaultTopologyDir
	}
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("ContainerLabManager")
	return &ContainerLabManager{
		TopologiesDir: topologiesDir,
		Topologies:    make(map[string]*Topology),
		SwitchClients: make(map[string]map[string]*switchclients.AristaClient),
		EAPISecretDir: eAPISecretDir,
		Logger:        logger,
	}
}

func (c *ContainerLabManager) GetSwitchClient(topologyName, nodeName string) (*switchclients.AristaClient, error) {
	allSWClients, found := c.SwitchClients[topologyName]
	if !found {
		return nil, fmt.Errorf("topology %v not found", topologyName)
	}
	client, found := allSWClients[nodeName]
	if !found {
		return nil, fmt.Errorf("switch client for %v not found", nodeName)
	}

	return client, nil
}

func (c *ContainerLabManager) findTopologyFile(topologyName string) (string, error) {
	topologyFileName := fmt.Sprintf("%s.clab.yml", topologyName)
	topologyPath := filepath.Join(c.TopologiesDir, topologyName, topologyFileName)
	if _, err := os.Stat(topologyPath); os.IsNotExist(err) {
		return "", fmt.Errorf("topology %s does not exist", topologyName)
	}
	return topologyPath, nil
}

func (c *ContainerLabManager) Deploy(topologyFilePath string) error {
	c.Logger.Info("Deploying containerlab...")
	command := fmt.Sprintf("sudo containerlab deploy --reconfigure --topo %s", topologyFilePath)

	output, err := testutils.RunCommand(command)
	if err != nil {
		fmt.Printf("containerlab deploy failed: %v\n Output: %v", err, output)
		return fmt.Errorf("containerlab deploy failed, %v", err)
	}
	c.Logger.Info("Containerlab deployed.")

	return nil
}

func (c *ContainerLabManager) Connect(topologyFilePath string) (*Topology, error) {

	command := fmt.Sprintf("sudo containerlab inspect --topo %s --format json", topologyFilePath)

	output, err := testutils.RunCommandStdOut(command)
	if err != nil {
		fmt.Printf("containerlab inspect failed: %v\n", err)
		return nil, fmt.Errorf("containerlab inspect failed, %v", err)
	}

	nodes, err := parseJsonOutput(output)
	if err != nil {
		fmt.Printf("failed to parse output: %v\n", err)
		return nil, fmt.Errorf("failed to parse output: %v", err)
	}
	nodesMap := make(map[string]Node)
	for _, node := range nodes {
		nodesMap[node.Name] = node
	}

	topology, err := c.ReadTopology(topologyFilePath)
	if err != nil {
		return nil, fmt.Errorf("ReadTopology failed, %v", err)
	}

	// find out all the switches and create the clients
	swClients := make(map[string]*switchclients.AristaClient)
	for nodeName, node := range topology.TopologySpec.Nodes {
		// update the nodes
		node.Name = nodeName
		nodeFQDN := testutils.ConvertSWShortNameToFQDN(nodeName, topology.Name)
		node.FQDN = nodeFQDN
		// get the node info from the deploy result
		n, found := nodesMap[nodeFQDN]
		if found {
			node.IPv4Address = n.IPv4Address
			node.IPv6Address = n.IPv6Address
			node.ContainerID = n.ContainerID
		}

		if IsSwitch(node) {
			c.Logger.Info(fmt.Sprintf("creating switch client for [%v, %v]", node.Name, node.IPv4Address))
			// wait for the container and 443 port to be ready
			// err := testutils.CheckPortOpenWithTimeout(n.IPv4Address, 443, 120*time.Second)
			// if err != nil {
			// 	fmt.Printf("CheckPortOpenWithTimeout failed: %v \n", err)
			// }

			var connectStartTime = time.Now()
			var swClient *switchclients.AristaClient
			allowedModes := []string{"access", "trunk"}
			allowedNativeVlanIds := []int{1, 55}
			allowedVlanIds := []int{100, 101, 102, 103, 104, 105, 222, 3999, 4008}
			provisioningVlanIds := []int{4008}
			for i := 0; time.Since(connectStartTime) < (60 * time.Second); i++ {
				swClient, err = switchclients.NewAristaClient(node.IPv4Address, c.EAPISecretDir, 443, "https", 30*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, []string{}, provisioningVlanIds)
				if err != nil {
					fmt.Printf("NewAristaClient error: %v, retry: %d \n", err, i+1)
					time.Sleep(500 * time.Millisecond)
					continue
				}
				break
			}
			err = swClient.ValidateConnection()
			if err != nil {
				fmt.Printf("failed to validate switch client for [%v, %v] \n", node.Name, node.IPv4Address)
			}

			c.Logger.Info(fmt.Sprintf("finished creating switch client for [%v]", node.Name))

			// save the running-config as the startup-config
			//_, err = swClient.SaveRunningConfigAsStartupConfig(context.Background())
			//if err != nil {
			//	fmt.Printf("saveRunningConfigAsStartupConfig failed, %v \n", err)
			//	continue
			//}

			swClients[nodeName] = swClient
			swClients[nodeFQDN] = swClient
		}
	}
	if c.SwitchClients == nil {
		c.SwitchClients = make(map[string]map[string]*switchclients.AristaClient)
	}
	c.Topologies[topology.Name] = topology

	c.SwitchClients[topology.Name] = swClients
	return topology, nil
}

func (c *ContainerLabManager) Destroy(topologyFilePath string) error {
	command := fmt.Sprintf("sudo containerlab destroy --topo %s --cleanup", topologyFilePath)
	_, err := testutils.RunCommand(command)
	if err != nil {
		return fmt.Errorf("containerlab destroy failed, %v", err)
	}

	topology, err := c.ReadTopology(topologyFilePath)
	if err != nil {
		return fmt.Errorf("ReadTopology failed, %v", err)
	}

	delete(c.Topologies, topology.Name)

	delete(c.SwitchClients, topology.Name)
	return nil
}

func (c *ContainerLabManager) DestroyAll() error {
	c.Logger.Info("Destroying all containerlabs...")

	command := fmt.Sprintf("sudo containerlab destroy --all --cleanup")
	output, err := testutils.RunCommand(command)
	if err != nil {
		if !strings.Contains(output, "no containerlab labs found") { // ignore error if no labs found.
			return fmt.Errorf("containerlab destroy all failed. Output: %s, error: %v", output, err)
		}
	}

	c.Topologies = make(map[string]*Topology)
	c.SwitchClients = make(map[string]map[string]*switchclients.AristaClient)
	return nil
}

// ResetAllSwitches reset the config for all the switches. Deploy() is a cleaner way to reset all nodes.
func (c *ContainerLabManager) ResetAllSwitches(topologyFilePath string) error {
	topology, err := c.ReadTopology(topologyFilePath)
	if err != nil {
		return fmt.Errorf("ReadTopology failed, %v", err)
	}

	allSwClients, found := c.SwitchClients[topology.Name]
	if !found {
		return fmt.Errorf("cannot find switch client for %v", topology.Name)
	}

	for _, swClient := range allSwClients {
		_, err := swClient.RestoreRunningConfigFromStartupConfig(context.Background())
		if err != nil {
			fmt.Printf("restoreRunningConfigFromStartupConfig failed, %v", err)
			return err
		}
	}
	return nil
}

func (c *ContainerLabManager) DeployWithName(topologyName string) (*Topology, error) {
	topologyPath, err := c.findTopologyFile(topologyName)
	if err != nil {
		return nil, err
	}
	c.Deploy(topologyPath)
	if err != nil {
		return nil, err
	}
	return c.Connect(topologyPath)
}

func (c *ContainerLabManager) DestroyWithName(topologyName string) error {
	topologyPath, err := c.findTopologyFile(topologyName)
	if err != nil {
		return err
	}
	return c.Destroy(topologyPath)
}

func (c *ContainerLabManager) GetTopology(topologyName string) (*Topology, bool) {
	topo, ok := c.Topologies[topologyName]
	return topo, ok
}

func (c *ContainerLabManager) ReadTopology(topologyFile string) (*Topology, error) {
	data, err := ioutil.ReadFile(topologyFile)
	if err != nil {
		fmt.Printf("Error reading topology file: %v", err)
		return nil, err
	}

	topo := &Topology{}
	err = yaml.Unmarshal(data, topo)
	if err != nil {
		fmt.Printf("Error unmarshalling topology: %v", err)
		return nil, err
	}

	return topo, nil
}

func IsSwitch(node *Node) bool {
	return strings.Contains(node.Kind, "eos")
}
func IsServer(node *Node) bool {
	return strings.Contains(node.Kind, "linux")
}

func IsServerByName(nodeName string) bool {
	return strings.HasPrefix(nodeName, ServerNodePrefix)
}
func IsFrontEndSwitchByName(nodeName string) bool {
	return strings.HasPrefix(nodeName, FrontendSwitchPrefix)
}

func IsACCSwitchByName(nodeName string) bool {
	return strings.HasPrefix(nodeName, ACCSwitchPrefix)
}

func IsSTRGSwitchByName(nodeName string) bool {
	return strings.HasPrefix(nodeName, StorageSwitchPrefix)
}

func extracGroupNameFromServerName(nodeName string) string {
	if IsServerByName(nodeName) {
		strs := strings.Split(strings.TrimPrefix(nodeName, ServerNodePrefix), "-")
		if len(strs) == 2 {
			return strs[0]
		}
	}
	return ""
}

type clabContainer struct {
	LabName     string `json:"lab_name"`
	LabPath     string `json:"labPath"`
	Name        string `json:"name"`
	ContainerId string `json:"container_id"`
	Image       string `json:"image"`
	Kind        string `json:"kind"`
	State       string `json:"state"`
	Ipv4Address string `json:"ipv4_address"`
	Ipv6Address string `json:"ipv6_address"`
}

type clabContainers struct {
	Containers []clabContainer `json:"containers"`
}

func parseJsonOutput(output string) ([]Node, error) {
	var ctrs clabContainers
	var containers []Node
	err := json.Unmarshal([]byte(output), &ctrs)
	if err != nil {
		fmt.Printf("Failed to parse output from Containerlab inspect: %s", err.Error())
		return nil, err
	}

	for i, ctr := range ctrs.Containers {
		ipv4 := strings.TrimSpace(ctr.Ipv4Address)
		ipv4 = ipv4[:strings.Index(ipv4, "/")]
		ipv6 := strings.TrimSpace(ctr.Ipv6Address)
		ipv6 = ipv6[:strings.Index(ipv6, "/")]

		container := Node{
			Number:      strings.TrimSpace(fmt.Sprintf("%d", i)),
			Name:        strings.TrimSpace(ctr.Name),
			ContainerID: strings.TrimSpace(ctr.ContainerId),
			Image:       strings.TrimSpace(ctr.Image),
			Kind:        strings.TrimSpace(ctr.Kind),
			State:       strings.TrimSpace(ctr.State),
			IPv4Address: ipv4,
			IPv6Address: ipv6,
		}
		containers = append(containers, container)
	}

	return containers, nil
}
