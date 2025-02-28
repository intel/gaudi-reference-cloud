// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/client"
)

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx)
	logger.Info("Started")

	conf := client.SDNClientConfig{
		//KubeConfig: "../../../../../local/secrets/restricted-kubeconfig/sdn-bmaas-kubeconfig.yaml",
		KubeConfig: "../../../../../local/secrets/kubeconfig/kind-idc-us-dev-1a-network.yaml",
		//LocalPoolInfoFilePath: "../config/big_cluster_conf/pools.json",
	}
	sdnclient, err := client.NewSDNClient(ctx, conf)
	if err != nil {
		fmt.Printf("NewSDNClient error: %v \n", err)
	}

	// testVlanUpdate(ctx, sdnclient)
	// testNetworkNodeUpdateAccelOnly(ctx, sdnclient)
	// testNetworkNodeUpdateStorageOnly(ctx, sdnclient)
	// testNetworkNodeUpdate(ctx, sdnclient)
	// testNodeGroupUpdate(ctx, sdnclient, 101, 1001, 101)
	// testNodeGroupUpdate(ctx, sdnclient, 4008, 1000, 4008)
	// testGetMapping(ctx, sdnclient)
	// testGetPoolConfigs(ctx, sdnclient)
	// testNetworkNodeUpdateWithoutBGP(ctx, sdnclient, false)
	// testNetworkNodeUpdateWithoutBGP(ctx, sdnclient, true)
	// testNetworkNodeUpdateWithBGP(ctx, sdnclient, false)
	// testNetworkNodeUpdateWithBGP(ctx, sdnclient, true)

	////////////////////////////////////////////////////
	// test examples for supporting trunking
	////////////////////////////////////////////////////
	// testNetworkNodeUpdateFrontEndAccessModeOnly(ctx, sdnclient)
	// testNetworkNodeUpdateFrontEndTrunkModeOnly(ctx, sdnclient)

	// an example when we want to deploy VM to a BM.
	exampleReserveBMforVM(ctx, sdnclient)
	// exampleReserveBMforVMWithNilTrunk(ctx, sdnclient)
	// exampleReserveBMforVMWithoutTrunk(ctx, sdnclient)
	// exampleReserveBMforVMWithMixedTrunkGroups(ctx, sdnclient)

	// release
	// exampleReleaseVM(ctx, sdnclient)
	// exampleReleaseVMTrunkGroupisNil(ctx, sdnclient)
	// exampleReleaseVMWithoutTrunkGroup(ctx, sdnclient)

	// testNNUpdateWithBGP(ctx, sdnclient)
	// testNNUpdateWithBGPReset(ctx, sdnclient)
}

func testVlanUpdate(ctx context.Context, sdnClient *client.SDNClient) {

	switchPort, err := sdnClient.GetSwitchPort(ctx, "internal-placeholder.com", "Ethernet27/1")
	if err != nil {
		fmt.Printf("GetSwitchPort error: %v \n", err)
	}
	fmt.Println(switchPort.Name)

	err = sdnClient.UpdateVlan(ctx, "internal-placeholder.com", "Ethernet27/1", 100, "test")
	if err != nil {
		fmt.Printf("updateVlan error: %v \n", err)
	}
}

func testNetworkNodeUpdateFrontEndAccessModeOnly(ctx context.Context, sdnClient *client.SDNClient) {
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:           "server1-1",
		FrontEndFabricMode:        "access",
		FrontEndFabricVlan:        102,
		FrontEndFabricTrunkGroups: []string{},
		FrontEndFabricNativeVlan:  1,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:                  "server1-1",
			DesiredFrontEndFabricMode:        "access",
			DesiredFrontEndFabricVlan:        102,
			DesiredFrontEndFabricTrunkGroups: []string{},
			DesiredFrontEndFabricNativeVlan:  1,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeConfigStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("testNetworkNodeUpdateFrontEndOnly SUCCESS \n")
}

func testNetworkNodeUpdateFrontEndTrunkModeOnly(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:           "server1-1",
		FrontEndFabricMode:        "trunk",
		FrontEndFabricTrunkGroups: []string{"Tenant_Nets"},
		FrontEndFabricNativeVlan:  1,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:                  "server1-1",
			DesiredFrontEndFabricMode:        "trunk",
			DesiredFrontEndFabricTrunkGroups: []string{"Tenant_Nets"},
			DesiredFrontEndFabricNativeVlan:  1,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeConfigStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("testNetworkNodeUpdateFrontEndOnly SUCCESS \n")
}

func testNetworkNodeUpdateAccelOnly(ctx context.Context, sdnClient *client.SDNClient) {

	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:       "pdx05-c01-bgan003",
		AcceleratorFabricVlan: 222,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:              "pdx05-c01-bgan003",
			DesiredAcceleratorFabricVlan: 222,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeConfigStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("testNetworkNodeUpdateAccelOnly SUCCESS \n")
}

func testNetworkNodeUpdate(ctx context.Context, sdnClient *client.SDNClient) {

	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:       "server1-1",
		FrontEndFabricVlan:    101,
		AcceleratorFabricVlan: 112,
		StorageFabricVlan:     123,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
		return
	}

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:              "server1-1",
			DesiredFrontEndFabricVlan:    101,
			DesiredAcceleratorFabricVlan: 112,
			DesiredStorageFabricVlan:     123,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeConfigStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("testNetworkNodeUpdate SUCCESS \n")
}

func testNNUpdateWithBGP(ctx context.Context, sdnClient *client.SDNClient) {

	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:                 "server1-1",
		FrontEndFabricVlan:              101,
		AcceleratorFabricBGPCommunityID: 1001,
		StorageFabricVlan:               102,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
		return
	}

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:                        "server1-1",
			DesiredFrontEndFabricVlan:              101,
			DesiredAcceleratorFabricBGPCommunityID: 1001,
			DesiredStorageFabricVlan:               102,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeConfigStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("testNetworkNodeUpdate SUCCESS \n")
}

func testNNUpdateWithBGPReset(ctx context.Context, sdnClient *client.SDNClient) {

	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:                 "server1-1",
		FrontEndFabricVlan:              4008,
		AcceleratorFabricBGPCommunityID: 1000,
		StorageFabricVlan:               4008,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
		return
	}

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:                        "server1-1",
			DesiredFrontEndFabricVlan:              4008,
			DesiredAcceleratorFabricBGPCommunityID: 1000,
			DesiredStorageFabricVlan:               4008,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeConfigStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("testNetworkNodeUpdate SUCCESS \n")
}

func testNetworkNodeUpdateWithoutBGP(ctx context.Context, sdnClient *client.SDNClient, reset bool) {
	vlanbase := 100
	for i := 1; i <= 8; i++ {
		node := fmt.Sprintf("g1n%d", i)
		vlan := int64(vlanbase + i)
		accvlan := int64(vlanbase + 100 + i)
		if reset {
			vlan = 4008
			accvlan = 100
		}

		err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
			NetworkNodeName:       node,
			FrontEndFabricVlan:    vlan,
			AcceleratorFabricVlan: accvlan,
			StorageFabricVlan:     vlan,
		})
		if err != nil {
			fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
			return
		}

		var result client.NetworkNodeConfStatusCheckResponse
		for result.Status != client.UpdateCompleted {
			result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
				NetworkNodeName:              node,
				DesiredFrontEndFabricVlan:    vlan,
				DesiredAcceleratorFabricVlan: accvlan,
				DesiredStorageFabricVlan:     vlan,
			})
			if err != nil {
				fmt.Printf("CheckNetworkNodeConfigStatus error: %v \n", err)
				return
			}
			time.Sleep(time.Second)
		}

	}
	fmt.Printf("testNetworkNodeUpdate SUCCESS \n")
}

func testNetworkNodeUpdateWithBGP(ctx context.Context, sdnClient *client.SDNClient, reset bool) {
	vlanbase := 100
	for i := 1; i <= 8; i++ {
		node := fmt.Sprintf("g1n%d", i)
		vlan := int64(vlanbase + i)
		bgp := int64(1234)
		if reset {
			vlan = 4008
			bgp = 1000
		}

		err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
			NetworkNodeName:                 node,
			FrontEndFabricVlan:              vlan,
			AcceleratorFabricBGPCommunityID: bgp,
			StorageFabricVlan:               vlan,
		})
		if err != nil {
			fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
			return
		}

		var result client.NetworkNodeConfStatusCheckResponse
		for result.Status != client.UpdateCompleted {
			result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
				NetworkNodeName:                        node,
				DesiredFrontEndFabricVlan:              vlan,
				DesiredAcceleratorFabricBGPCommunityID: bgp,
				DesiredStorageFabricVlan:               vlan,
			})
			if err != nil {
				fmt.Printf("CheckNetworkNodeConfigStatus error: %v \n", err)
				return
			}
			time.Sleep(time.Second)
		}

	}
	fmt.Printf("testNetworkNodeUpdate SUCCESS \n")
}

func testNodeGroupUpdate(ctx context.Context, sdnClient *client.SDNClient, fevlan, accbgp, stvlan int64) {
	updateRequest := client.NodeGroupConfUpdateRequest{
		NodeGroupName: "group-1",
		DesiredFrontEndFabricConfig: &idcnetworkv1alpha1.FabricConfig{
			VlanConf: &idcnetworkv1alpha1.VlanConfig{VlanID: fevlan},
			BGPConf:  nil,
		},
		DesiredAcceleratorFabricConfig: &idcnetworkv1alpha1.FabricConfig{
			VlanConf: nil,
			BGPConf:  &idcnetworkv1alpha1.BGPConfig{BGPCommunity: accbgp},
		},
		DesiredStorageFabricConfig: &idcnetworkv1alpha1.FabricConfig{
			VlanConf: &idcnetworkv1alpha1.VlanConfig{VlanID: stvlan},
			BGPConf:  nil,
		},
	}

	err := sdnClient.UpdateNodeGroupConfig(ctx, updateRequest)
	if err != nil {
		fmt.Printf("UpdateNodeGroupConfig error: %v \n", err)
		return
	}

	var result client.NodeGroupConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNodeGroupStatus(ctx, client.CheckNodeGroupStatusRequest(updateRequest))
		if err != nil {
			fmt.Printf("CheckNodeGroupConfigStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("testNodeGroupUpdate SUCCESS \n")
}

func testNetworkNodeUpdateStorageOnly(ctx context.Context, sdnClient *client.SDNClient) {

	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:   "pdx05-c01-bgan003",
		StorageFabricVlan: 444,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:          "pdx05-c01-bgan003",
			DesiredStorageFabricVlan: 444,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeConfigStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("testNetworkNodeUpdateStorageOnly SUCCESS \n")
}

// func testGetMapping(ctx context.Context, sdnClient *client.SDNClient) {
// 	cm, err := sdnClient.GetGroupToPoolMapping(ctx)
// 	if err != nil {
// 		fmt.Printf("GetGroupToPoolMapping failed %v \n", err)
// 	}
// 	fmt.Printf("%v \n", cm)
// }

/*
output example:
{"name":"VBX","description":"front-end VLAN, accel BGP","networkConfigStrategy":{"frontEndFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":4008}},"acceleratorFabricStrategy":{"isolationType":"BGP","provisionConfig":{"defaultVlanID":99,"defaultBGPCommunity":999}}},"schedulingConfig":{"minimumSchedulableUnit":"NodeGroup"}}
{"name":"VVX","description":"front-end VLAN, accel VLAN","networkConfigStrategy":{"frontEndFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":4008}},"acceleratorFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":100,"defaultBGPCommunity":1000}}},"schedulingConfig":{"minimumSchedulableUnit":"NetworkNode"}}
{"name":"VVX-standalone","description":"front-end VLAN, accel VLAN (no spine)","standaloneNodeGroupOnly":true,"networkConfigStrategy":{"frontEndFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":4008}},"acceleratorFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":100}}},"schedulingConfig":{"minimumSchedulableUnit":"NetworkNode"}}
{"name":"VBV","description":"front-end VLAN, accel BGP, storage VLAN","networkConfigStrategy":{"frontEndFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":4008}},"acceleratorFabricStrategy":{"isolationType":"BGP","provisionConfig":{"defaultVlanID":99,"defaultBGPCommunity":999}},"storageFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":4008}}},"schedulingConfig":{"minimumSchedulableUnit":"NodeGroup"}}
{"name":"VVV","description":"front-end VLAN, accel VLAN, storage VLAN","networkConfigStrategy":{"frontEndFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":4008}},"acceleratorFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":100,"defaultBGPCommunity":1000}},"storageFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":4008}}},"schedulingConfig":{"minimumSchedulableUnit":"NetworkNode"}}
{"name":"VXX","description":"front end VLAN only","standaloneNodeGroupOnly":true,"networkConfigStrategy":{"frontEndFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":4008}}},"schedulingConfig":{"minimumSchedulableUnit":"NetworkNode"}}
{"name":"XBX","description":"acc BGP only","networkConfigStrategy":{"acceleratorFabricStrategy":{"isolationType":"BGP","provisionConfig":{"defaultVlanID":99,"defaultBGPCommunity":999}}},"schedulingConfig":{"minimumSchedulableUnit":"NodeGroup"}}
{"name":"XVX","description":"acc VLAN only","networkConfigStrategy":{"acceleratorFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":100,"defaultBGPCommunity":1000}}},"schedulingConfig":{"minimumSchedulableUnit":"NetworkNode"}}
{"name":"VXV","description":"front end VLAN, storage VLAN","standaloneNodeGroupOnly":true,"networkConfigStrategy":{"frontEndFabricStrategy":{"isolationType":"VLAN","provisionConfig":{"defaultVlanID":4008}}},"schedulingConfig":{"minimumSchedulableUnit":"NetworkNode"}}
*/
func testGetPoolConfigs(ctx context.Context, sdnClient *client.SDNClient) {
	poolConfigs, err := sdnClient.ListNetworkPoolConfigs(ctx)
	if err != nil {
		fmt.Printf("ListNetworkPoolConfigs failed %v \n", err)
	}
	for _, pc := range poolConfigs {
		data, err := json.Marshal(pc)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to marshal pollConfig to json"))
		}
		fmt.Printf("%s \n", string(data))
	}
}

// ////////////////////////////
// trunking tests
// ////////////////////////////

// exampleReserveBMforVM is the example of reserving a BM Host for VMaaS.
/*
output example:

NN:
spec:
  frontEndFabric:
    mode: trunk
    nativeVlan: 1
    switchPort: ethernet5.clab-frontendonly-frontend-leaf1
    trunkGroups:
    - Tenant_Nets
    vlanId: 4008
status:
  acceleratorFabricStatus: {}
  frontEndFabricStatus:
    lastObservedMode: trunk
    lastObservedNativeVlan: 1
    lastObservedTrunkGroups:
    - Tenant_Nets
    readiness: 1/1
  storageFabricStatus: {}

SP:
spec:
  mode: trunk
  name: Ethernet5
  nativeVlan: 1
  trunkGroups:
  - Tenant_Nets
  vlanId: 4008
status:
  lastStatusChangeTime: "2024-07-24T16:37:48Z"
  lineProtocolStatus: up
  linkStatus: connected
  mode: trunk
  name: Ethernet5
  nativeVlan: 1
  trunkGroups:
  - Tenant_Nets
*/
func exampleReserveBMforVM(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:    "server1-2",
		FrontEndFabricMode: "trunk",
		// note: this is to set the trunk group to meet the below desired state(other trunk groups will be removed)
		FrontEndFabricTrunkGroups: []string{"Tenant_Nets"},
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:                  "server1-2",
			DesiredFrontEndFabricMode:        "trunk",
			DesiredFrontEndFabricTrunkGroups: []string{"Tenant_Nets"},
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("exampleReserveBMforVM SUCCESS \n")
}

func exampleReserveBMforVMWithMixedTrunkGroups(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:    "server1-2",
		FrontEndFabricMode: "trunk",
		// note: this is to set the trunk group to meet the below desired state(other trunk groups will be removed)
		// FrontEndFabricTrunkGroups: []string{"Tenant_Nets", "Provider_Nets"},
		FrontEndFabricTrunkGroups: []string{"Tenant_Nets", "Service_Nets"},
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:           "server1-2",
			DesiredFrontEndFabricMode: "trunk",
			// DesiredFrontEndFabricTrunkGroups: []string{"Tenant_Nets", "Provider_Nets"},
			DesiredFrontEndFabricTrunkGroups: []string{"Tenant_Nets", "Service_Nets"},
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("exampleReserveBMforVM SUCCESS \n")
}

func exampleReserveBMforVMWithNilTrunk(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:           "server1-2",
		FrontEndFabricMode:        "trunk",
		FrontEndFabricTrunkGroups: nil,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:                  "server1-2",
			DesiredFrontEndFabricMode:        "trunk",
			DesiredFrontEndFabricTrunkGroups: nil,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("exampleReserveBMforVMWithNilTrunk SUCCESS \n")
}

func exampleReserveBMforVMWithoutTrunk(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:    "server1-2",
		FrontEndFabricMode: "trunk",
		// FrontEndFabricTrunkGroups: nil,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:           "server1-2",
			DesiredFrontEndFabricMode: "trunk",
			// DesiredFrontEndFabricTrunkGroups: nil,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("exampleReserveBMforVMWithNilTrunk SUCCESS \n")
}

// if no trunk group is provided in the request, it should do nothing to the switch's trunk groups.
// the NN and SP's TrunkGroups fields should be `nil`, their status should reflect the actual configs.
// NOTE: this is NOT suggested, when moving to trunk mode, we SHOULD provide the trunk group in the request.

/*
output example:
NN:
spec:
  frontEndFabric:
    mode: trunk
    nativeVlan: 1
    switchPort: ethernet5.clab-frontendonly-frontend-leaf1
    vlanId: 4008
status:
  acceleratorFabricStatus: {}
  frontEndFabricStatus:
    lastObservedMode: trunk
    lastObservedNativeVlan: 1
    lastObservedTrunkGroups:
    - Tenant_Nets
    readiness: 1/1
  storageFabricStatus: {}


SP:
spec:
  mode: trunk
  name: Ethernet5
  nativeVlan: 1
  vlanId: 4008
status:
  lastStatusChangeTime: "2024-07-24T16:54:40Z"
  lineProtocolStatus: up
  linkStatus: connected
  mode: trunk
  name: Ethernet5
  nativeVlan: 1
  trunkGroups:
  - Tenant_Nets
*/

func moveToTrunkModeWithoutTG(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:    "server1-1",
		FrontEndFabricMode: "trunk",
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:           "server1-1",
			DesiredFrontEndFabricMode: "trunk",
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("moveToTrunkModeWithoutTG SUCCESS \n")
}

// exampleReleaseVM is the example of returning a Host from VMaaS to BMaaS
/*
output example:

NN:
spec:
  frontEndFabric:
    mode: access
    nativeVlan: 1
    switchPort: ethernet5.clab-frontendonly-frontend-leaf1
    trunkGroups: []
    vlanId: 4008
status:
  acceleratorFabricStatus: {}
  frontEndFabricStatus:
    lastObservedMode: access
    lastObservedNativeVlan: 1
    lastObservedVlanId: 4008
    readiness: 1/1
  storageFabricStatus: {}


SP:
spec:
  mode: access
  name: Ethernet5
  nativeVlan: 1
  trunkGroups: []
  vlanId: 4008
status:
  lastStatusChangeTime: "2024-07-24T16:52:18Z"
  lineProtocolStatus: up
  linkStatus: connected
  mode: access
  name: Ethernet5
  nativeVlan: 1
  vlanId: 4008
*/

func exampleReleaseVM(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:    "server1-2",
		FrontEndFabricMode: "access",
		// if vlan is not provided, SDN will set it with what it currently have in the SwitchPort CR.
		FrontEndFabricVlan: 4008,
		// optional, this will remove all the trunk group for this port
		FrontEndFabricTrunkGroups: []string{},
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:                  "server1-2",
			DesiredFrontEndFabricMode:        "access",
			DesiredFrontEndFabricVlan:        4008,
			DesiredFrontEndFabricTrunkGroups: []string{},
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("exampleReleaseVM SUCCESS \n")
}

func exampleReleaseVMTrunkGroupisNil(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:    "server1-2",
		FrontEndFabricMode: "access",
		// if vlan is not provided, SDN will set it with what it currently have in the SwitchPort CR.
		FrontEndFabricVlan: 4008,
		// optional, this will remove all the trunk group for this port
		FrontEndFabricTrunkGroups: nil,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:                  "server1-2",
			DesiredFrontEndFabricMode:        "access",
			DesiredFrontEndFabricVlan:        4008,
			DesiredFrontEndFabricTrunkGroups: nil,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("exampleReleaseVM SUCCESS \n")
}

func exampleReleaseVMWithoutTrunkGroup(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:    "server1-2",
		FrontEndFabricMode: "access",
		// if vlan is not provided, SDN will set it with what it currently have in the SwitchPort CR.
		FrontEndFabricVlan: 4008,
		// optional, this will remove all the trunk group for this port
		// FrontEndFabricTrunkGroups: nil,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:           "server1-2",
			DesiredFrontEndFabricMode: "access",
			DesiredFrontEndFabricVlan: 4008,
			// DesiredFrontEndFabricTrunkGroups: nil,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("exampleReleaseVM SUCCESS \n")
}

// if an empty trunk group is provided in the request, it will remove all the switch's trunk groups.
// the NN and SP's TrunkGroups fields should be `[]string{}`, their status should have no trunk group fields.
// we should not do so when converting to trunk mode, we SHOULD provide the trunk group in the request.
/*
output example:
NN:
spec:
  frontEndFabric:
    mode: trunk
    nativeVlan: 1
    switchPort: ethernet5.clab-frontendonly-frontend-leaf1
    trunkGroups: []
    vlanId: 4008
status:
  acceleratorFabricStatus: {}
  frontEndFabricStatus:
    lastObservedMode: trunk
    lastObservedNativeVlan: 1
    readiness: 1/1
  storageFabricStatus: {}

SP:
spec:
  mode: trunk
  name: Ethernet5
  nativeVlan: 1
  trunkGroups: []
  vlanId: 4008
status:
  lastStatusChangeTime: "2024-07-24T17:00:18Z"
  lineProtocolStatus: up
  linkStatus: connected
  mode: trunk
  name: Ethernet5
  nativeVlan: 1
*/
func moveToTrunkModeWithEmptyTG(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:           "server1-1",
		FrontEndFabricMode:        "trunk",
		FrontEndFabricTrunkGroups: []string{},
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:                  "server1-1",
			DesiredFrontEndFabricMode:        "trunk",
			DesiredFrontEndFabricTrunkGroups: []string{},
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("moveToTrunkModeWithEmptyTG SUCCESS \n")
}

// this release the VM with a `nil` trunk group, which mean sdn will leave what the trunk groups as is.
/*
output example:
NN:
spec:
  frontEndFabric:
    mode: access
    nativeVlan: 1
    switchPort: ethernet5.clab-frontendonly-frontend-leaf1
    vlanId: 4008
status:
  acceleratorFabricStatus: {}
  frontEndFabricStatus:
    lastObservedMode: access
    lastObservedNativeVlan: 1
    lastObservedTrunkGroups:
    - Tenant_Nets
    lastObservedVlanId: 4008
    readiness: 1/1
  storageFabricStatus: {}

SP:
spec:
  mode: access
  name: Ethernet5
  nativeVlan: 1
  vlanId: 4008
status:
  lastStatusChangeTime: "2024-07-24T17:07:48Z"
  lineProtocolStatus: up
  linkStatus: connected
  mode: access
  name: Ethernet5
  nativeVlan: 1
  trunkGroups:
  - Tenant_Nets
  vlanId: 4008
*/
func releaseVMWithoutSpecifyingTG(ctx context.Context, sdnClient *client.SDNClient) {
	fmt.Printf("running the test...\n")
	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:    "server1-1",
		FrontEndFabricMode: "access",
		FrontEndFabricVlan: 4008,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

	fmt.Printf("update done..\n")

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:           "server1-1",
			DesiredFrontEndFabricMode: "access",
			DesiredFrontEndFabricVlan: 4008,
		})
		if err != nil {
			fmt.Printf("CheckNetworkNodeStatus error: %v \n", err)
			return
		}
		time.Sleep(time.Second)
	}
	fmt.Printf("releaseVMWithoutSpecifyingTG SUCCESS \n")
}

// this will be rejected
func updateVlanAndTrunk(ctx context.Context, sdnClient *client.SDNClient) {

	err := sdnClient.UpdateNetworkNodeConfig(ctx, client.NetworkNodeConfUpdateRequest{
		NetworkNodeName:           "server1-1",
		FrontEndFabricMode:        "trunk",
		FrontEndFabricTrunkGroups: []string{"Tenant_Nets"},
		FrontEndFabricVlan:        101,
	})
	if err != nil {
		fmt.Printf("UpdateNetworkNodeConfig error: %v \n", err)
	}

}

// ////////////////////////////
// trunking tests end
// ////////////////////////////
