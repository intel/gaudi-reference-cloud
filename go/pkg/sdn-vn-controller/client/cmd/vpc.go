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

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var vpcCreateCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Create VPC",
	Long:  `Create VPC <vpc id> <name> <tenant id> <region id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(4)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		createVPC(args[0], args[1], args[2], args[3])
	},
}

var vpcListCmd = &cobra.Command{
	Use:   "vpc",
	Short: "List VPC",
	Long:  `List VPC`,
	Run: func(cmd *cobra.Command, args []string) {
		listVPC()
	},
}

var vpcDeleteCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Delete VPC",
	Long:  `Delete VPC <vpc id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		deleteVPC(args[0])
	},
}

var vpcGetCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Get VPC",
	Long:  `Get VPC <vpc id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		getVPC(args[0])
	},
}

func init() {
	createCmd.AddCommand(vpcCreateCmd)
	listCmd.AddCommand(vpcListCmd)
	deleteCmd.AddCommand(vpcDeleteCmd)
	getCmd.AddCommand(vpcGetCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// vpcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// vpcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func listVPC() {
	VPCList, err := c.ListVPCs(ctx, &v1.ListVPCsRequest{},
		grpc.WaitForReady(true))

	if err != nil {
		log.Fatalf("could not get list of VPCs: %v", err)
	}
	fmt.Println("VPC List is : ")
	for vpcId, vpcIndo := range VPCList.VpcIds {
		fmt.Printf("%v:   ID %v: \n",
			vpcId, vpcIndo.Uuid)
	}

}

func createVPC(uuid string, name string, tenantId string, regionId string) {
	req := v1.CreateVPCRequest{
		VpcId:    &v1.VPCId{Uuid: uuid},
		Name:     name,
		TenantId: tenantId,
		RegionId: regionId,
	}
	ret, err := c.CreateVPC(ctx, &req)
	if err != nil {
		log.Fatalf("VPC could not be created %v", err)
	}
	fmt.Printf("VPC is : %v\n", ret.VpcId.Uuid)
}

func deleteVPC(uuid string) {
	req := v1.DeleteVPCRequest{
		VpcId: &v1.VPCId{Uuid: uuid},
	}
	ret, err := c.DeleteVPC(ctx, &req)
	if err != nil {
		log.Fatalf("VPC could not be deleted %v", err)
	}
	fmt.Printf("VPC Deleted was : %v\n", ret.VpcId.Uuid)
}

func getVPC(uuid string) {
	req := v1.GetVPCRequest{
		VpcId: &v1.VPCId{Uuid: uuid},
	}
	ret, err := c.GetVPC(ctx, &req)
	if err != nil {
		log.Fatalf("Failed to get VPC %v", err)
	}
	fmt.Printf(
		`VPC info:
ID: %v
Name: %v
Tenant ID: %v
`,
		ret.Vpc.Id, ret.Vpc.Name, ret.Vpc.TenantId)
}
