package main

import (
	"context"
	"fmt"

	nc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/netbox"
)

func main() {
	ctx := context.Background()
	netboxClient, err := nc.NewSDNNetBoxClient("https://internal-placeholder.com", "xxx", true)
	// netboxClient, err := nc.NewSDNNetBoxClient("https://internal-placeholder.com", "xxx", true)
	// netboxClient, err := nc.NewSDNNetBoxClient("https://internal-placeholder.com", "xxx", true)
	if err != nil {
		fmt.Printf("unable to initialize NetBox client: %v \n", err)
	}

	// switchesFilter := &nc.DevicesFilter{}
	// switchesListReq := nc.ListDevicesRequest{
	// 	Filter: switchesFilter,
	// }
	// switches, err := netboxClient.ListDevices(ctx, switchesListReq)
	// if err != nil {
	// 	fmt.Printf("failed to fetch switches from Netbox, %v \n", err)
	// }
	// fmt.Printf("switches: [%+v] \n", switches)

	InterfacesFilter := &nc.InterfacesFilter{}
	ListInterfaceRequest := nc.ListInterfaceRequest{
		Filter: InterfacesFilter,
	}
	interfaces, err := netboxClient.ListInterfaces(ctx, ListInterfaceRequest)
	if err != nil {
		fmt.Printf("ListInterfaces failed, %v \n", err)
	}
	fmt.Printf("interfaces: [%+v] \n", interfaces)

	// ipRequest := nc.ListIPAddressesRequest{
	// 	Filter: &nc.IPAddressesFilter{
	// 		// InterfacesId: []int32{mgmt.GetId()},
	// 	},
	// }

	// ips, err := netboxClient.ListIPAddresses(ctx, ipRequest)
	// if err != nil {
	// 	fmt.Printf("ListIPAddresses failed, %v", err)
	// }
	// fmt.Printf("ips: [%+v] \n", ips)

}
