// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/client"
)

func main() {
	var numConcurrentThreads int
	var nwcpKubeconfig string
	var requestType string
	var durationInSec int
	var totalRequests int
	var reset bool
	flag.BoolVar(&reset, "reset", false, "")
	flag.StringVar(&requestType, "t", "ng", "creating ng or nn")
	// run for this amount of time, and then stop
	flag.IntVar(&durationInSec, "d", 30, "run this test for d seconds")
	// execute this amount of requests, and then stop
	flag.IntVar(&totalRequests, "r", 10, "run r requests")
	// spawn 'c' number of threads that keep working on updating ng or nn.
	flag.IntVar(&numConcurrentThreads, "c", 8, "number of concurrent threads")
	// flag.StringVar(&nwcpKubeconfig, "nwcpKubeconfig", "/home/jzhen/.kube/config.idc-staging-nwcp-devel", "nwcp kubeconfig file path")
	flag.StringVar(&nwcpKubeconfig, "nwcpKubeconfig", "/home/jzhen/.kube/config", "nwcp kubeconfig file path")
	// flag.StringVar(&nwcpKubeconfig, "nwcpKubeconfig", "", "nwcp kubeconfig file path")
	flag.Parse()

	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx)
	logger.Info("Started")

	conf := client.SDNClientConfig{
		// KubeConfig: "../../../../../local/secrets/kubeconfig/kind-idc-us-dev-1a-network.yaml",
		KubeConfig: nwcpKubeconfig,
	}
	sdnClient, err := client.NewSDNClient(ctx, conf)
	if err != nil {
		fmt.Printf("NewSDNClient error: %v \n", err)
		os.Exit(1)
	}

	if requestType == "ng" {
		testNG(ctx, sdnClient, numConcurrentThreads, totalRequests, durationInSec)
	} else if requestType == "nn" {
		testNN(ctx, sdnClient, numConcurrentThreads, totalRequests, durationInSec)
	}

}

func testNG(ctx context.Context, sdnClient *client.SDNClient, numWorkers int, totalRounds int, durationInSec int) {
	var totalTimeConsumed time.Duration
	var totalNumRequests int
	resList := make([]time.Duration, 0)
	resChan := make(chan time.Duration, 0)
	stopChan := make(chan bool, 0)
	fmt.Printf("number of concurrent NodeGroup workers %v \n", numWorkers)

	// each worker will update the NodeGroup and then reset it back to default state.

	start := time.Now().UTC()
	for i := 1; i <= numWorkers; i++ {
		vlan := int64(100 + i)
		bgp := int64(1000 + i)
		group := fmt.Sprintf("group-%d", i)
		go func(vlan, bgp int64, g string, worker int, rChan chan time.Duration, stopCh chan bool) {
			for {
				select {
				case <-stopCh:
					fmt.Printf("stopping work load generation...\n")
					return
				default:
					var request1, request2 client.NodeGroupConfUpdateRequest

					request1 = client.NodeGroupConfUpdateRequest{
						NodeGroupName: group,
						DesiredFrontEndFabricConfig: &idcnetworkv1alpha1.FabricConfig{
							VlanConf: &idcnetworkv1alpha1.VlanConfig{VlanID: vlan},
							BGPConf:  nil,
						},
						DesiredAcceleratorFabricConfig: &idcnetworkv1alpha1.FabricConfig{
							VlanConf: nil,
							BGPConf:  &idcnetworkv1alpha1.BGPConfig{BGPCommunity: bgp},
						},
					}

					request2 = client.NodeGroupConfUpdateRequest{
						NodeGroupName: group,
						DesiredFrontEndFabricConfig: &idcnetworkv1alpha1.FabricConfig{
							VlanConf: &idcnetworkv1alpha1.VlanConfig{VlanID: 4008},
							BGPConf:  nil,
						},
						DesiredAcceleratorFabricConfig: &idcnetworkv1alpha1.FabricConfig{
							VlanConf: nil,
							BGPConf:  &idcnetworkv1alpha1.BGPConfig{BGPCommunity: 1000},
						},
					}

					// update
					res1, err := updateNodeGroup(ctx, sdnClient, request1)
					if err != nil {
						fmt.Printf("updateVB error: %v \n", err)
						os.Exit(1)
					}
					rChan <- res1

					// reset
					res2, err := updateNodeGroup(ctx, sdnClient, request2)
					if err != nil {
						fmt.Printf("updateVB error: %v \n", err)
						os.Exit(1)
					}
					rChan <- res2
					time.Sleep(time.Second)
				}
			}
		}(vlan, bgp, group, i, resChan, stopChan)
	}

	go func(rChan chan time.Duration) {
		for res := range rChan {
			totalTimeConsumed += res
			totalNumRequests++
			resList = append(resList, res)
		}
	}(resChan)

	if totalRounds > 0 {
		// coverity[INFINITE_LOOP:FALSE]
		for totalNumRequests < totalRounds {
			time.Sleep(time.Millisecond * 100)
		}
	} else if durationInSec > 0 {
		<-time.After(time.Second * time.Duration(durationInSec))
	}
	testDuration := time.Since(start)

	// stop the generator
	stopChan <- true
	time.Sleep(1 * time.Second)

	// reset
	for i := 1; i <= numWorkers; i++ {
		group := fmt.Sprintf("group-%d", i)

		var request client.NodeGroupConfUpdateRequest

		request = client.NodeGroupConfUpdateRequest{
			NodeGroupName: group,
			DesiredFrontEndFabricConfig: &idcnetworkv1alpha1.FabricConfig{
				VlanConf: &idcnetworkv1alpha1.VlanConfig{VlanID: 4008},
				BGPConf:  nil,
			},
			DesiredAcceleratorFabricConfig: &idcnetworkv1alpha1.FabricConfig{
				VlanConf: nil,
				BGPConf:  &idcnetworkv1alpha1.BGPConfig{BGPCommunity: 1000},
			},
		}

		// reset
		_, err := updateNodeGroup(ctx, sdnClient, request)
		if err != nil {
			fmt.Printf("reset error: %v \n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("testElapsed\t\t%v\n", testDuration)
	fmt.Printf("numRequests\t\t%v\n", totalNumRequests)
	fmt.Printf("durations\t\t%v\n", resList)
	fmt.Printf("average\t\t\t%v\n", totalTimeConsumed/time.Duration(totalNumRequests))
}

func updateNodeGroup(ctx context.Context, sdnClient *client.SDNClient, request client.NodeGroupConfUpdateRequest) (time.Duration, error) {
	start := time.Now().UTC()
	var updateDuration time.Duration
	var checkDuration time.Duration
	err := sdnClient.UpdateNodeGroupConfig(ctx, request)
	if err != nil {
		return 0, err
	}

	updateDuration = time.Since(start)
	start2 := time.Now().UTC()
	var result client.NodeGroupConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNodeGroupStatus(ctx, client.CheckNodeGroupStatusRequest(request))
		if err != nil {
			return 0, err
		}
	}
	checkDuration = time.Since(start2)
	fmt.Printf("update NodeGroup duration %v, wait for ready duration %v \n", updateDuration, checkDuration)
	return time.Since(start), nil
}

func testNN(ctx context.Context, sdnClient *client.SDNClient, numWorkers int, totalRequests int, durationInSec int) {
	var totalTimeConsumed time.Duration
	var totalNumRequests int
	resList := make([]time.Duration, 0)
	resChan := make(chan time.Duration, 0)
	fmt.Printf("number of concurrent NodeGroup workers %v \n", numWorkers)

	start := time.Now().UTC()
	for i := 0; i < numWorkers; i++ {
		groupID := i/8 + 1
		nodeID := i%8 + 1
		networkNodeName := fmt.Sprintf("g%vn%v", groupID, nodeID)
		vlan := int64(100 + i + 1)

		go func(vlan int64, nn string, worker int, rChan chan time.Duration) {
			for {

				// update
				request1 := client.NetworkNodeConfUpdateRequest{
					NetworkNodeName:       nn,
					FrontEndFabricVlan:    vlan,
					AcceleratorFabricVlan: vlan,
				}
				request1StatusCheck := client.NetworkNodeConfStatusCheckRequest{
					NetworkNodeName:              nn,
					DesiredFrontEndFabricVlan:    vlan,
					DesiredAcceleratorFabricVlan: vlan,
				}

				res1, err := updateNetworkNode(ctx, sdnClient, request1, request1StatusCheck)
				if err != nil {
					fmt.Printf("updateNetworkNode error: %v \n", err)
					os.Exit(1)
				}
				rChan <- res1

				// reset
				request2 := client.NetworkNodeConfUpdateRequest{
					NetworkNodeName:       nn,
					FrontEndFabricVlan:    4008,
					AcceleratorFabricVlan: 100,
				}
				request2StatusCheck := client.NetworkNodeConfStatusCheckRequest{
					NetworkNodeName:              nn,
					DesiredFrontEndFabricVlan:    4008,
					DesiredAcceleratorFabricVlan: 100,
				}

				res2, err := updateNetworkNode(ctx, sdnClient, request2, request2StatusCheck)
				if err != nil {
					fmt.Printf("updateNetworkNode error: %v \n", err)
					os.Exit(1)
				}
				rChan <- res2

				time.Sleep(time.Second)
			}
		}(vlan, networkNodeName, i, resChan)
	}

	go func(rChan chan time.Duration) {
		for res := range rChan {
			totalTimeConsumed += res
			totalNumRequests++
			resList = append(resList, res)
		}
	}(resChan)

	if totalRequests > 0 {
		// coverity[INFINITE_LOOP:FALSE]
		for totalNumRequests < totalRequests {
			time.Sleep(time.Millisecond * 100)
		}
	} else if durationInSec > 0 {
		<-time.After(time.Second * time.Duration(durationInSec))
	}
	testDuration := time.Since(start)

	// reset
	for i := 0; i < numWorkers; i++ {
		groupID := i/8 + 1
		nodeID := i%8 + 1

		networkNodeName := fmt.Sprintf("g%vn%v", groupID, nodeID)

		request := client.NetworkNodeConfUpdateRequest{
			NetworkNodeName:       networkNodeName,
			FrontEndFabricVlan:    4008,
			AcceleratorFabricVlan: 100,
		}
		requestStatusCheck := client.NetworkNodeConfStatusCheckRequest{
			NetworkNodeName:              networkNodeName,
			DesiredFrontEndFabricVlan:    4008,
			DesiredAcceleratorFabricVlan: 100,
		}

		_, err := updateNetworkNode(ctx, sdnClient, request, requestStatusCheck)
		if err != nil {
			fmt.Printf("updateNetworkNode error: %v \n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("testElapsed\t\t%v\n", testDuration)
	fmt.Printf("numRequests\t\t%v\n", totalNumRequests)
	fmt.Printf("durations\t\t%v\n", resList)
	fmt.Printf("average\t\t\t%v\n", totalTimeConsumed/time.Duration(totalNumRequests))
}

func updateNetworkNode(ctx context.Context, sdnClient *client.SDNClient, request client.NetworkNodeConfUpdateRequest, checkRequest client.NetworkNodeConfStatusCheckRequest) (time.Duration, error) {
	start := time.Now().UTC()

	err := sdnClient.UpdateNetworkNodeConfig(ctx, request)
	if err != nil {
		return 0, err
	}

	var result client.NetworkNodeConfStatusCheckResponse
	for result.Status != client.UpdateCompleted {
		result, err = sdnClient.CheckNetworkNodeStatus(ctx, checkRequest)
		if err != nil {
			return 0, err
		}
	}
	return time.Since(start), nil
}
