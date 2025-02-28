// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation

package cmd

import (
	"fmt"
	"log"
	"strings"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// securityGroupCreateCmd represents the "sdnctl create security-group" command
var securityGroupCreateCmd = &cobra.Command{
	Use:   "securitygroup",
	Short: "Create securitygroup",
	Long:  `Create securitygroup <name> <vpc_id> <type: port|subnet> [port_ids | subnet_ids]`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(4)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create security group called")
		createSecurityGroup(args)
	},
}

// securityGroupDeleteCmd represents the "sdnctl delete security-group" command
var securityGroupDeleteCmd = &cobra.Command{
	Use:   "securitygroup",
	Short: "Delete securitygroup",
	Long:  `Delete securitygroup <security_group_uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete security group called")
		deleteSecurityGroup(args[0])
	},
}

// securityGroupGetCmd retrieves a specific security group by its ID
var securityGroupGetCmd = &cobra.Command{
	Use:   "securitygroup",
	Short: "Get security group",
	Long:  `Get security group <security group id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		getSecurityGroup(args[0])
	},
}

// securityGroupListCmd lists all security groups
var securityGroupListCmd = &cobra.Command{
	Use:   "securitygroup",
	Short: "List security groups",
	Long:  `List security groups`,
	Run: func(cmd *cobra.Command, args []string) {
		listSecurityGroups()
	},
}

var securityGroupUpdateCmd = &cobra.Command{
	Use:   "securitygroup",
	Short: "Update security groups",
	Long:  `Update securitygroup <security group id> <type: port|subnet> [port_ids | subnet_ids]`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(3)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update security group called")
		updateSecurityGroup(args)
	},
}

func init() {
	createCmd.AddCommand(securityGroupCreateCmd)
	deleteCmd.AddCommand(securityGroupDeleteCmd)
	getCmd.AddCommand(securityGroupGetCmd)
	listCmd.AddCommand(securityGroupListCmd)
	updateCmd.AddCommand(securityGroupUpdateCmd)
}

func createSecurityGroup(args []string) {
	name := args[0]
	vpcID := args[1]
	sg_type_str := args[2]
	var sg_type v1.SecurityGroupType
	var portIDs, subnetIDs []string
	var portUUIDs []*v1.PortId
	var subnetUUIDs []*v1.SubnetId

	if sg_type_str == "port" {
		sg_type = v1.SecurityGroupType_PORT
		portIDs = strings.Split(args[3], ",")
		for _, id := range portIDs {
			portUUIDs = append(portUUIDs, &v1.PortId{Uuid: id})
		}
	} else if sg_type_str == "subnet" {
		sg_type = v1.SecurityGroupType_SUBNET
		subnetIDs = strings.Split(args[3], ",")
		for _, id := range subnetIDs {
			subnetUUIDs = append(subnetUUIDs, &v1.SubnetId{Uuid: id})
		}
	} else {
		log.Fatalf("Invalid group type: %s", sg_type_str)
	}

	csreq := v1.CreateSecurityGroupRequest{
		SecurityGroupId: &v1.SecurityGroupId{
			Uuid: uuid.NewString(),
		},
		Name:      name,
		PortIds:   portUUIDs,
		SubnetIds: subnetUUIDs,
		Type:      sg_type,
		VpcId:     &v1.VPCId{Uuid: vpcID},
	}

	groupRet, err := c.CreateSecurityGroup(ctx, &csreq)
	if err != nil {
		log.Fatalf("CMD: The security group could not be created: %v\n", err)
		return
	}
	fmt.Printf("Security Group Created: %v\n", groupRet.SecurityGroupId.Uuid)
}

func deleteSecurityGroup(uuid string) {
	dsreq := v1.DeleteSecurityGroupRequest{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: uuid},
	}
	groupRet, err := c.DeleteSecurityGroup(ctx, &dsreq)
	if err != nil {
		log.Fatalf("The security group could not be deleted %v\n", err)
	}
	fmt.Printf("Security Group deleted was : %v\n", groupRet.SecurityGroupId.Uuid)
}

func getSecurityGroup(uuid string) {
	req := v1.GetSecurityGroupRequest{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: uuid},
	}
	ret, err := c.GetSecurityGroup(ctx, &req)
	if err != nil {
		log.Fatalf("Failed to get Security Group %v", err)
	}

	fmt.Printf(
		`Security Group info:
	ID: %v
	Name: %v
	VPC ID: %v
	Security Rule IDs: %v
	Port IDs: %v
	Subnet IDs: %v
	`,
		ret.SecurityGroup.Id.Uuid,
		ret.SecurityGroup.Name,
		ret.SecurityGroup.VpcId.Uuid,
		ret.SecurityGroup.SecurityRuleIds,
		ret.SecurityGroup.PortIds,
		ret.SecurityGroup.SubnetIds)
}

func listSecurityGroups() {
	securityGroupList, err := c.ListSecurityGroups(ctx, &v1.ListSecurityGroupsRequest{})
	if err != nil {
		log.Fatalf("could not get list of Security Groups: %v", err)
	}

	fmt.Println("Security Group List:")
	for _, securityGroupId := range securityGroupList.SecurityGroupIds {
		fmt.Printf("ID: %v\n", securityGroupId.Uuid)
	}
}

func updateSecurityGroup(args []string) {
	securityGroupID := args[0]
	sg_type_str := args[1]
	var portIDs, subnetIDs []string
	var portUUIDs []*v1.PortId
	var subnetUUIDs []*v1.SubnetId

	if sg_type_str == "port" {
		portIDs = strings.Split(args[2], ",")
		for _, id := range portIDs {
			portUUIDs = append(portUUIDs, &v1.PortId{Uuid: id})
		}
	} else if sg_type_str == "subnet" {
		subnetIDs = strings.Split(args[2], ",")
		for _, id := range subnetIDs {
			subnetUUIDs = append(subnetUUIDs, &v1.SubnetId{Uuid: id})
		}
	} else {
		log.Fatalf("Invalid group type: %s", sg_type_str)
	}

	usreq := v1.UpdateSecurityGroupRequest{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: securityGroupID},
		PortIds:         portUUIDs,
		SubnetIds:       subnetUUIDs,
	}

	// Call the update security group API
	groupRet, err := c.UpdateSecurityGroup(ctx, &usreq)
	if err != nil {
		log.Fatalf("CMD: The security group could not be updated: %v\n", err)
		return
	}
	fmt.Printf("Security Group Updated: %v\n", groupRet.SecurityGroupId.Uuid)
}
