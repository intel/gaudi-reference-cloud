// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
)

func main() {

	var fqdn string
	var eapiFile string
	var eapiPort int
	log.BindFlags() // Use --zap-log-level X, higher X means more verbose (0 = info, 1 = debug)
	flag.StringVar(&eapiFile, "eapifile", "eapi", "eapi credential file path")
	// flag.StringVar(&fqdn, "fqdn", "internal-placeholder.com", "The address of the switch.")
	flag.StringVar(&fqdn, "fqdn", "", "The address of the switch.")
	flag.IntVar(&eapiPort, "port", 443, "The port of the eapi API on the switch.")
	flag.Parse()
	eapiFile = "/vault/secrets/eapi"
	log.SetDefaultLogger()

	allowedModes := []string{"access", "trunk"}
	allowedNativeVlanIds := []int{1, 55}
	allowedVlanIds := []int{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 222, 3999, 4008}
	provisioningVlanIds := []int{4008}

	// testEAPISwitchConnection(fqdn, eapiFile, eapiPort, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
	// testMultipleGet(fqdn, eapiFile, port, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)

	// Switch 1
	//testEAPIGetBGPCommunity(ip, eapiFile, 9022, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
	//testEAPIUpdateBGPCommunity(ip, eapiFile, 9022, 108, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)

	// Switch 2
	// testEAPIUpdateBGPCommunity(ip, eapiFile, 9023, 104, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)

	// testRavenSwitchClient()
	// testEAPIGetPortChannels(fqdn, eapiFile, eapiPort, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
	// testEAPICreatePortChannel(fqdn, eapiFile, eapiPort, 111, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
	// testDeletePortChannel(fqdn, eapiFile, eapiPort, 111, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
	// testAssignSwitchPortToPortChannel(fqdn, eapiFile, eapiPort, "Et5", 111, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
	testUpdateVlan(fqdn, eapiFile, eapiPort, "Ethernet27", int32(102), allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
	// testVlanUpdateForPortChannel(fqdn, eapiFile, eapiPort, 111, 100, allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
	// testUpdatePortChannelMode(fqdn, eapiFile, eapiPort, 111, "access", allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
	// testUpdatePortChannelMode(fqdn, eapiFile, eapiPort, 111, "trunk", allowedVlanIds, allowedNativeVlanIds, allowedModes, provisioningVlanIds)
}

func testRavenSwitchClient() {
	ctx := context.Background()
	cfg := idcnetworkv1alpha1.RavenConfiguration{
		Host:                "raven-devcloud.app.intel.com",
		Environment:         "rnd",
		CredentialsFilePath: "/vault/secrets/raven",
	}
	c, err := sc.NewRavenSwitchClient(cfg)
	if err != nil {
		fmt.Printf("NewRavenSwitchClient err: [%v] \n", err)
	}
	getres, err := c.GetSwitchPorts(ctx, sc.GetSwitchPortsRequest{
		SwitchFQDN: "fxhb3p3r-zal0113a.idcmgt.intel.com",
		Env:        "rnd",
	})
	if err != nil {
		fmt.Printf("GetSwitchPorts err: [%v] \n", err)
	}
	fmt.Printf("getres %v \n", getres)

	// err = c.UpdateVlan(ctx, sc.UpdateVlanRequest{
	//      SwitchFQDN: "fxhb3p3r-zal0113a.idcmgt.intel.com",
	//      PortName:   "Ethernet27/1",
	//      Vlan:       400,
	//      Env:        "rnd",
	// })
	// if err != nil {
	//      fmt.Printf("UpdateVlan err: [%v] \n", err)
	// }
	// fmt.Printf("UpdateVlan SUCCESS\n")
}

func testMultipleGet(fqdn string, eapi string, port int, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	c, err := sc.NewAristaClient(fqdn, eapi, port, "https", 120*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}

	cnt := 100
	var wg sync.WaitGroup
	fmt.Printf("time.now %v \n", time.Now().UTC())
	for i := 0; i < cnt; i++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			res, err := c.GetSwitchPorts(context.Background(), sc.GetSwitchPortsRequest{})
			if err != nil {
				fmt.Printf("%v GetSwitchPorts err: [%v] \n", time.Now().UTC(), err)
			}

			fmt.Printf("%v worker %v done, res: %v \n", time.Now().UTC(), w, res)
			// fmt.Printf("worker %v done\n", w)
		}(i)
		fmt.Printf("submit worker %v done\n", i)
	}
	wg.Wait()
	fmt.Printf("done\n")
}

func testEAPISwitchConnection(fqdn string, eapi string, port int, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, port, "https", 30*time.Second, true, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	getres, err := c.GetSwitchPorts(context.Background(), sc.GetSwitchPortsRequest{})
	if err != nil {
		fmt.Printf("GetSwitchPorts err: [%v] \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling GetSwitchPorts(): %v \n", time.Since(now))
	for _, p := range getres {
		fmt.Printf("Port Name: %v \n", p.Name)
	}
}

func testUpdateVlan(fqdn string, eapi string, eapiport int, spName string, vlan int32, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, eapiport, "https", 30*time.Second, true, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	err = c.UpdateVlan(context.Background(), sc.UpdateVlanRequest{
		Vlan:     vlan,
		PortName: spName,
	})
	if err != nil {
		fmt.Printf("UpdateVlan err: [%v] \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling GetSwitchPorts(): %v \n", time.Since(now))
}

func testEAPIGetBGPCommunity(fqdn string, eapi string, port int, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, port, "https", 30*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	community, err := c.GetBGPCommunity(context.Background(), sc.GetBGPCommunityRequest{
		BGPCommunityIncomingGroupName: "incoming_group",
	})
	if err != nil {
		fmt.Printf("GetBGPCommunity err: [%v] \n", err)
		panic(err)
	}

	fmt.Printf("Got BGP Community: %d \n", community)

	fmt.Printf("Time elapsed after calling GetBGPCommunity(): %v \n", time.Since(now))

}

func testEAPIUpdateBGPCommunity(fqdn string, eapi string, port int, bgpCommunity int, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, port, "https", 30*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	err = c.UpdateBGPCommunity(context.Background(), sc.UpdateBGPCommunityRequest{
		BGPCommunity:                  int32(bgpCommunity),
		BGPCommunityIncomingGroupName: "incoming_group",
	})
	if err != nil {
		fmt.Printf("UpdateBGPCommunity err: [%v] \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling UpdateBGPCommunity(): %v \n", time.Since(now))
}

func testEAPIGetPortChannels(fqdn string, eapi string, port int, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, port, "https", 30*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	req := sc.GetPortChannelsRequest{}
	pcs, err := c.GetPortChannels(context.Background(), req)
	if err != nil {
		fmt.Printf("GetPortChannels err: [%v] \n", err)
		panic(err)
	}

	for key, val := range pcs {
		fmt.Printf("port channels: %v, \n\t %+v \n", key, val)
	}

	fmt.Printf("Time elapsed after calling GetPortChannels(): %v \n", time.Since(now))

}

func testEAPICreatePortChannel(fqdn string, eapi string, port int, pc int64, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, port, "https", 30*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	req := sc.CreatePortChannelRequest{
		PortChannel: pc,
	}
	err = c.CreatePortChannel(context.Background(), req)
	if err != nil {
		fmt.Printf("CreatePortChannel err: [%v] \n", err)
		panic(err)
	}

	fmt.Printf("Time elapsed after calling CreatePortChannel(): %v \n", time.Since(now))

}

func testDeletePortChannel(fqdn string, eapi string, port int, pc int64, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, port, "https", 30*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	req := sc.DeletePortChannelRequest{
		TargetPortChannel: pc,
	}
	err = c.DeletePortChannel(context.Background(), req)
	if err != nil {
		fmt.Printf("CreatePortChannel err: [%v] \n", err)
		panic(err)
	}

	fmt.Printf("Time elapsed after calling CreatePortChannel(): %v \n", time.Since(now))

}

func testAssignSwitchPortToPortChannel(fqdn string, eapi string, eapiport int, sp string, pc int64, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, eapiport, "https", 30*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	req := sc.AssignSwitchPortToPortChannelRequest{
		SwitchPort:        sp,
		TargetPortChannel: pc,
	}
	err = c.AssignSwitchPortToPortChannel(context.Background(), req)
	if err != nil {
		fmt.Printf("AssignSwitchPortToPortChannel err: [%v] \n", err)
		panic(err)
	}

	fmt.Printf("Time elapsed after calling CreatePortChannel(): %v \n", time.Since(now))
}

/*
func testVlanUpdateForPortChannel(fqdn string, eapi string, eapiport int, pc int32, vlan int32, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, eapiport, "https", 30*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	req := sc.UpdatePortChannelVlanRequest{
		PortChannel: pc,
		Vlan:        vlan,
	}
	err = c.UpdatePortChannelVlan(context.Background(), req)
	if err != nil {
		fmt.Printf("UpdatePortChannelVlan err: [%v] \n", err)
	}

	fmt.Printf("Time elapsed after calling CreatePortChannel(): %v \n", time.Since(now))
}

func testUpdatePortChannelMode(fqdn string, eapi string, eapiport int, pc int32, mode string, allowedVlanIds []int, allowedNativeVlanIds []int, allowedModes []string, provisioningVlanIds []int) {
	now := time.Now().UTC()
	fmt.Printf("Now: %v \n", now)
	c, err := sc.NewAristaClient(fqdn, eapi, eapiport, "https", 30*time.Second, false, allowedVlanIds, allowedNativeVlanIds, allowedModes, nil, provisioningVlanIds)
	if err != nil {
		fmt.Printf("NewAristaClient error: %v \n", err)
		panic(err)
	}
	fmt.Printf("Time elapsed after calling NewAristaClient(): %v \n", time.Since(now))

	req := sc.UpdatePortChannelModeRequest{
		PortChannel: pc,
		Mode:        mode,
	}
	err = c.UpdatePortChannelMode(context.Background(), req)
	if err != nil {
		fmt.Printf("UpdatePortChannelVlan err: [%v] \n", err)
	}

	fmt.Printf("Time elapsed after calling CreatePortChannel(): %v \n", time.Since(now))
}
*/
