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

// subnetCreateCmd represents the "sdnctl create subnet" command
var subnetCreateCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Create subnet",
	Long:  `Create subnet <subnet name> <CIDR> <vpc id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(3)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create subnet called")
		createSwitch(args[0], args[1], args[2])
	},
}

// subnetDeleteCmd represents the "sdnctl delete subnet" command
var subnetDeleteCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Delete subnet",
	Long:  `Delete subnet <subnet uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete subnet called")
		deleteSwitch(args[0])
	},
}

// subnetGetCmd represents the "sdnctl get subnet" command
var subnetGetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Get subnet",
	Long:  `Get subnet <subnet uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get subnet called")
		getSubnet(args[0])
	},
}

// subnetListCmd represents the "sdnctl list subnet" command
var subnetListCmd = &cobra.Command{
	Use:   "subnet",
	Short: "List subnets",
	Long:  `List subnets`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list subnet called")
		listSwitches()
	},
}

func init() {
	createCmd.AddCommand(subnetCreateCmd)
	getCmd.AddCommand(subnetGetCmd)
	listCmd.AddCommand(subnetListCmd)
	deleteCmd.AddCommand(subnetDeleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// subnetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// subnetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func listSwitches() {
	subnetList, err := c.ListSubnets(ctx, &v1.ListSubnetsRequest{},
		grpc.WaitForReady(true))

	if err != nil {
		log.Fatalf("could not get list of subnets: %v\n", err)
	}
	fmt.Println("Subnet List is : ")
	for subnetid, lp := range subnetList.SubnetIds {
		fmt.Printf("%v   %v\n\n", subnetid, lp.Uuid)
	}
}

func getSubnet(uuid string) {
	gsReq := v1.GetSubnetRequest{
		SubnetId: &v1.SubnetId{Uuid: uuid},
	}
	subnetRet, err := c.GetSubnet(ctx, &gsReq)
	if err != nil {
		log.Fatalf("Failed to get subnet %v", err)
	}
	fmt.Printf("Subnet info: %s, %v\n", subnetRet.Subnet.Name, subnetRet.Subnet.Id)
}

func createSwitch(name string, cidr string, vpcId string) {
	csreq := v1.CreateSubnetRequest{

		// Generate a UUID for the Subnet
		SubnetId: &v1.SubnetId{
			Uuid: uuid.NewString(),
		},

		Name:  name,
		Cidr:  cidr,
		VpcId: &v1.VPCId{Uuid: vpcId},
	}
	subnetRet, err := c.CreateSubnet(ctx, &csreq)
	if err != nil {
		log.Fatalf("The subnet could not be created %v\n", err)
	}
	fmt.Printf("Subnet Created is : %v\n", subnetRet.SubnetId.Uuid)
}

func deleteSwitch(uuid string) {
	dsreq := v1.DeleteSubnetRequest{
		SubnetId: &v1.SubnetId{Uuid: uuid},
	}
	subnetRet, err := c.DeleteSubnet(ctx, &dsreq)
	if err != nil {
		log.Fatalf("The subnet could not be deleted %v\n", err)
	}
	fmt.Printf("Subnet Deleted was : %v\n", subnetRet.SubnetId.Uuid)
}
