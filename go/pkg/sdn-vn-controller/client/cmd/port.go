// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"strconv"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// portCreateCmd represents the "sdnctl create port" command
var portCreateCmd = &cobra.Command{
	Use:   "port",
	Short: "Create port",
	Long:  `Create port <switch_uuid> <chassis_id> <device_id> <ip_addr> <MAC> [is_nat]`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 5 || len(args) > 6 {
			return fmt.Errorf("requires 5 or 6 arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create port called")
		createPort(args)
	},
}

// portDeleteCmd represents the "sdnctl delete port" command
var portDeleteCmd = &cobra.Command{
	Use:   "port",
	Short: "Delete port",
	Long:  `Delete port <port_uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete port called")
		deletePort(args[0])
	},
}

// portUpdateCmd represents the "sdnctl create port" command
var portUpdateCmd = &cobra.Command{
	Use:   "port",
	Short: "Update port",
	Long:  `Update port <port_uuid> <is_nat> <admin_state>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(3)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Update port called")
		updatePort(args[0], args[1], args[2])
	},
}

var portListCmd = &cobra.Command{
	Use:   "port",
	Short: "List port",
	Long:  `List port`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list port called")
		listPorts()
	},
}

var portGetCmd = &cobra.Command{
	Use:   "port",
	Short: "Get port",
	Long:  `Get port <port_uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get port called")
		getPort(args[0])
	},
}

func init() {
	createCmd.AddCommand(portCreateCmd)
	deleteCmd.AddCommand(portDeleteCmd)
	updateCmd.AddCommand(portUpdateCmd)
	listCmd.AddCommand(portListCmd)
	getCmd.AddCommand(portGetCmd)
}

func createPort(args []string) {
	subnet_id := args[0]
	chassis_id := args[1]
	device_id := args[2]
	ip_addr := args[3]
	mac := args[4]

	deviceIdUint, err := strconv.ParseUint(device_id, 10, 32)
	if err != nil {
		log.Fatalf("Invalid device_id: %v\n", err)
		return
	}

	// Default isNAT to false if not provided
	isNat := false
	if len(args) == 6 {
		isNatBool, err := strconv.ParseBool(args[5])
		if err != nil {
			log.Fatalf("Invalid boolean value for isNAT param: %s\n", args[4])
			return
		}
		isNat = isNatBool
	}

	csreq := v1.CreatePortRequest{
		PortId: &v1.PortId{
			Uuid: uuid.NewString(),
		},
		SubnetId: &v1.SubnetId{
			Uuid: subnet_id,
		},
		ChassisId:          chassis_id,
		DeviceId:           uint32(deviceIdUint),
		Internal_IPAddress: ip_addr,
		IsEnabled:          true,
		IsNAT:              isNat,
		MACAddress:         mac,
	}

	portRet, err := c.CreatePort(ctx, &csreq)
	if err != nil {
		log.Fatalf("CMD:The port could not be created: %v\n", err)
		return
	}
	fmt.Printf("Port Created: %v\n", portRet.PortId.Uuid)
}

func deletePort(uuid string) {
	dsreq := v1.DeletePortRequest{
		PortId: &v1.PortId{Uuid: uuid},
	}
	portRet, err := c.DeletePort(ctx, &dsreq)
	if err != nil {
		log.Fatalf("The port could not be deleted %v\n", err)
	}
	fmt.Printf("Port deleted was : %v\n", portRet.PortId.Uuid)
}

func updatePort(uuid string, isnat string, IsEnabled string) {
	fmt.Printf("Update port : %s\n", uuid)
	isNatBool, err := strconv.ParseBool(isnat)
	if err != nil {
		log.Fatalf("Invalid boolean value for IsNat param: %s\n", isnat)
		return
	}

	IsEnabledBool, err := strconv.ParseBool(IsEnabled)
	if err != nil {
		log.Fatalf("Invalid boolean value for IsEnabled param: %s\n", IsEnabled)
		return
	}

	csreq := v1.UpdatePortRequest{
		PortId:    &v1.PortId{Uuid: uuid},
		IsNAT:     &isNatBool,
		IsEnabled: &IsEnabledBool,
	}

	portRet, err := c.UpdatePort(ctx, &csreq)
	if err != nil {
		log.Fatalf("CMD:The port could not be updated: %v\n", err)
		return
	}
	fmt.Printf("Port updated: %v\n", portRet.PortId.Uuid)
}

func listPorts() {
	resp, err := c.ListPorts(ctx, &v1.ListPortsRequest{})
	if err != nil {
		log.Fatalf("The ports could not be listed %v\n", err)
	}
	fmt.Println("Port List is : ")
	for portid, lp := range resp.PortIds {
		fmt.Printf("%v   %v\n\n", portid, lp.Uuid)
	}
}

func getPort(uuid string) {
	req := &v1.GetPortRequest{
		PortId: &v1.PortId{Uuid: uuid},
	}
	resp, err := c.GetPort(ctx, req)
	if err != nil {
		log.Fatalf("Failed to get the port %v", err)
	}
	port := resp.GetPort()
	fmt.Printf("Port info: \n  %v \n  Subnet %v \n  IP %v \n  Chassis %v \n  Device %v \n",
		port.Id, port.SubnetId, port.IPAddress, port.ChassisId, port.DeviceId,
	)
}
