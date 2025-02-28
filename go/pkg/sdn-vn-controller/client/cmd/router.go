// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"net"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// routerCmd represents the create router command
var routerCreateCmd = &cobra.Command{
	Use:   "router",
	Short: "Create Router",
	Long:  `Create Router <router name> <vpc id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create router called")
		createRouter(args[0], args[1])
	},
}

// routerDeleteCmd represents the delete router command
var routerDeleteCmd = &cobra.Command{
	Use:   "router",
	Short: "Delete Router",
	Long:  `Delete Router <router uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete router called")
		deleteRouter(args[0])
	},
}

// routerGetCmd represents the get router command
var routerGetCmd = &cobra.Command{
	Use:   "router",
	Short: "Get Router",
	Long:  `Get Router <router uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get router called")
		getRouter(args[0])
	},
}

// routerListCmd represents the list router command
var routerListCmd = &cobra.Command{
	Use:   "router",
	Short: "List Router",
	Long:  `List Router`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list router called")
		listRouters()
	},
}

func init() {
	createCmd.AddCommand(routerCreateCmd)
	getCmd.AddCommand(routerGetCmd)
	listCmd.AddCommand(routerListCmd)
	deleteCmd.AddCommand(routerDeleteCmd)

	routerCreateCmd.AddCommand(routerCreateInterfaceCmd)
	routerDeleteCmd.AddCommand(routerDeleteInterfaceCmd)
	routerListCmd.AddCommand(routerListInterfacesCmd)
	routerGetCmd.AddCommand(routerGetInterfaceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// routerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// routerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func listRouters() {
	routerList, err := c.ListRouters(ctx, &v1.ListRoutersRequest{},
		grpc.WaitForReady(true))

	if err != nil {
		log.Fatalf("could not get list of routers: %v", err)
	}
	fmt.Println("Router List is : ")
	for routerid, lr := range routerList.RouterIds {
		// fmt.Printf("%v   %v\n\n", routerid, lr)
		fmt.Printf("%v:   ID %v: \n",
			routerid, lr.Uuid)
	}
}

func getRouter(uuid string) {
	grReq := v1.GetRouterRequest{
		RouterId: &v1.RouterId{Uuid: uuid},
	}
	routerRet, err := c.GetRouter(ctx, &grReq)
	if err != nil {
		log.Fatalf("Failed to get the router %v", err)
	}
	fmt.Printf("Router info: %s, %v\n", routerRet.Router.Name, routerRet.Router.Id)
}

func createRouter(name string, vpcId string) {
	crreq := v1.CreateRouterRequest{

		RouterId: &v1.RouterId{
			Uuid: uuid.NewString(),
		},

		Name: name,
		//			RouterInterfaces: []*v1.RouterInterface{&routerInterface},
		VpcId:            &v1.VPCId{Uuid: vpcId},
		AvailabilityZone: "AZ1",
	}
	routerRet, err := c.CreateRouter(ctx, &crreq)
	if err != nil {
		log.Fatalf("The router could not be created %v", err)
	}
	fmt.Printf("Router is : %v\n", routerRet.RouterId.Uuid)
}

func deleteRouter(uuid string) {
	drreq := v1.DeleteRouterRequest{
		RouterId: &v1.RouterId{Uuid: uuid},
	}
	routerRet, err := c.DeleteRouter(ctx, &drreq)
	if err != nil {
		log.Fatalf("The router could not be deleted %v", err)
	}
	fmt.Printf("Router Deleted was : %v\n", routerRet.RouterId.Uuid)
}

// routerCreateInterfaceCmd represents the create router interface command
var routerCreateInterfaceCmd = &cobra.Command{
	Use:   "interface",
	Short: "Create Router Interface",
	Long:  `Create Router Interface <router uuid> <subnet uuid> <IP addr> <MAC addr>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(4)(cmd, args); err != nil {
			return err
		} else {

			_, _, err := net.ParseCIDR(args[2]) // returns (IP, *IPNet, error)
			if err != nil {
				return fmt.Errorf("invalid IP prefix specified: %s", args[2])
			}

			_, err2 := net.ParseMAC(args[3]) // returns (HardwareAddr, error)
			if err2 != nil {
				return fmt.Errorf("invalid MAC address specified: %s", args[3])
			}
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create router interface called")
		createRouterInterface(args)
	},
}

// routerDeleteInterfaceCmd represents the delete router interface command
var routerDeleteInterfaceCmd = &cobra.Command{
	Use:   "interface",
	Short: "Delete Router Interface",
	Long:  `Delete Router Interface <router interface uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete router interface called")
		deleteRouterInterface(args)
	},
}

var routerListInterfacesCmd = &cobra.Command{
	Use:   "interface",
	Short: "List Router Interface",
	Long:  `List Router Interface`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list router interface called")
		listRouterInterfaces()
	},
}

var routerGetInterfaceCmd = &cobra.Command{
	Use:   "interface",
	Short: "Get Router Interface",
	Long:  `Get Router Interface <router interface uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get router interface called")
		getRouterInterface(args[0])
	},
}

// Create Router Interface <router uuid> <subnet uuid> <IP addr> <MAC addr>
func createRouterInterface(args []string) {
	cifreq := v1.CreateRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{
			Uuid: uuid.NewString(),
		},
		RouterId:      &v1.RouterId{Uuid: args[0]},
		SubnetId:      &v1.SubnetId{Uuid: args[1]},
		Interface_IP:  args[2],
		Interface_MAC: args[3],
	}
	routerIfRet, err := c.CreateRouterInterface(ctx, &cifreq)
	if err != nil {
		log.Fatalf("The router interface could not be created %v\n", err)
	}
	fmt.Printf("Router interface uuid is : %v\n", routerIfRet.RouterInterfaceId.Uuid)
}

// Delete Router Interface <router uuid> <subnet uuid>
func deleteRouterInterface(args []string) {
	difreq := v1.DeleteRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{Uuid: args[0]},
	}
	routerIfRet, err := c.DeleteRouterInterface(ctx, &difreq)
	if err != nil {
		log.Fatalf("The router interface could not be deleted %v\n", err)
	}
	fmt.Printf("Router interface with uuid %v has been deleted\n", routerIfRet.RouterInterfaceId)
}

func listRouterInterfaces() {
	resp, err := c.ListRouterInterfaces(ctx, &v1.ListRouterInterfacesRequest{})
	if err != nil {
		log.Fatalf("The router interfaces could not be listed %v\n", err)
	}
	fmt.Println("Router Interface List is : ")
	for ifid, lp := range resp.RouterInterfaceIds {
		fmt.Printf("%v   %v\n\n", ifid, lp.Uuid)
	}
}

func getRouterInterface(uuid string) {
	req := v1.GetRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{Uuid: uuid},
	}
	resp, err := c.GetRouterInterface(ctx, &req)
	if err != nil {
		log.Fatalf("Failed to get the router interface %v", err)
	}
	iface := resp.GetRouterInterface()
	fmt.Printf("Router Interface info: \n  %v \n  Router %v \n  Subnet %v \n  IP %v \n  MAC %v \n",
		iface.Id, iface.RouterId, iface.SubnetId, iface.Interface_IP, iface.Interface_MAC,
	)
}
