// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var staticRouteCreateCmd = &cobra.Command{
	Use:   "staticroute",
	Short: "Create Static Route",
	Long:  `Create Static Route <router uuid> <ip prefix> <next hop>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(3)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create static route called")
		createStaticRoute(args[0], args[1], args[2])
	},
}

var staticRouteDeleteCmd = &cobra.Command{
	Use:   "staticroute",
	Short: "Delete Static Route",
	Long:  `Delete Static Route <route uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete static route called")
		deleteStaticRoute(args[0])
	},
}

var staticRouteListCmd = &cobra.Command{
	Use:   "staticroute",
	Short: "List Static Route",
	Long:  `List Static Route`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list static route called")
		listStaticRoute()
	},
}

var staticRouteGetCmd = &cobra.Command{
	Use:   "staticroute",
	Short: "Get Static Route",
	Long:  `Get Static Route <route uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get static route called")
		getStaticRoute(args[0])
	},
}

func init() {
	createCmd.AddCommand(staticRouteCreateCmd)
	listCmd.AddCommand(staticRouteListCmd)
	deleteCmd.AddCommand(staticRouteDeleteCmd)
	getCmd.AddCommand(staticRouteGetCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// routerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// routerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func listStaticRoute() {
	staticRouteList, err := c.ListStaticRoutes(ctx, &v1.ListStaticRoutesRequest{},
		grpc.WaitForReady(true))

	if err != nil {
		log.Fatalf("could not get list of static routes: %v", err)
	}
	fmt.Println("Static Route List is : ")
	for routeid, route := range staticRouteList.StaticRouteIds {
		fmt.Printf("%v:   ID %v: \n",
			routeid, route.Uuid)
	}

}

func createStaticRoute(routerUuid string, prefix string, nexthop string) {
	csreq := v1.CreateStaticRouteRequest{

		StaticRouteId: &v1.StaticRouteId{
			Uuid: uuid.NewString(),
		},
		RouterId: &v1.RouterId{
			Uuid: routerUuid,
		},
		Prefix:  prefix,
		Nexthop: nexthop,
	}
	staticRouteRet, err := c.CreateStaticRoute(ctx, &csreq)
	if err != nil {
		log.Fatalf("The static route could not be created %v", err)
	}
	fmt.Printf("Static route is : %v\n", staticRouteRet.StaticRouteId.Uuid)
}

func deleteStaticRoute(uuid string) {
	drreq := v1.DeleteStaticRouteRequest{
		StaticRouteId: &v1.StaticRouteId{Uuid: uuid},
	}
	ret, err := c.DeleteStaticRoute(ctx, &drreq)
	if err != nil {
		log.Fatalf("The static route could not be deleted %v", err)
	}
	fmt.Printf("Static Route Deleted was : %v\n", ret.StaticRouteId.Uuid)
}

func getStaticRoute(uuid string) {
	grReq := v1.GetStaticRouteRequest{
		StaticRouteId: &v1.StaticRouteId{Uuid: uuid},
	}
	ret, err := c.GetStaticRoute(ctx, &grReq)
	if err != nil {
		log.Fatalf("Failed to get the static route %v", err)
	}
	fmt.Printf(
		`Static Route info:
Router ID: %v
Prefix: %v
Next Hop: %v
`,
		ret.StaticRoute.RouterId.Uuid, ret.StaticRoute.Prefix, ret.StaticRoute.Nexthop)
}
