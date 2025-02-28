// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation

package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// securityRuleCreateCmd represents the "sdnctl create security-rule" command
var securityRuleCreateCmd = &cobra.Command{
	Use:   "securityrule",
	Short: "Create securityrule",
	Long: `Create securityrule <name> <vpc_id> <security_group_uuid> <priority> <direction> <action> [source_IP_addresses] [remote_IP_addresses] [protocol] [source_port_range_min] [source_port_range_max]
	 [destination_port_range_min] [destination_port_range_max]`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.RangeArgs(6, 12)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create security rule called")
		createSecurityRule(args)
	},
}

var securityRuleGetCmd = &cobra.Command{
	Use:   "securityrule",
	Short: "Get securityrule",
	Long:  `Get securityrule <securityrule id>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		getSecurityRule(args[0])
	},
}

var securityRuleListCmd = &cobra.Command{
	Use:   "securityrule",
	Short: "List securityrule",
	Long:  `List securityrule`,
	Run: func(cmd *cobra.Command, args []string) {
		listSecurityRules()
	},
}

// securityRuleDeleteCmd represents the "sdnctl delete security-rule" command
var securityRuleDeleteCmd = &cobra.Command{
	Use:   "securityrule",
	Short: "Delete securityrule",
	Long:  `Delete security rule <security_rule_uuid>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		} else {
			return nil
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete security rule called")
		deleteSecurityRule(args[0])
	},
}

var securityRuleUpdateCmd = &cobra.Command{
	Use:   "securityrule",
	Short: "Update security rule",
	Long: `Update security rule <security_rule_uuid> <priority> <direction> <action> [source_IP_addresses] [destination_IP_addresses] [protocol] [source_port_range_min] [source_port_range_max]
	 [destination_port_range_min] [destination_port_range_max]`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.RangeArgs(4, 11)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update security rule called")
		updateSecurityRule(args)
	},
}

func init() {
	createCmd.AddCommand(securityRuleCreateCmd)
	deleteCmd.AddCommand(securityRuleDeleteCmd)
	getCmd.AddCommand(securityRuleGetCmd)
	listCmd.AddCommand(securityRuleListCmd)
	updateCmd.AddCommand(securityRuleUpdateCmd)
}

func createSecurityRule(args []string) {
	name := args[0]
	vpcId := args[1]
	securityGroupUUID := args[2]
	priority, err := strconv.Atoi(args[3])
	if err != nil {
		log.Fatalf("Invalid priority: %v\n", err)
		return
	}

	direction, err := parseDirection(args[4])
	if err != nil {
		log.Fatalf("Invalid direction: %v\n", err)
		return
	}

	action, err := parseAction(args[5])
	if err != nil {
		log.Fatalf("Invalid action: %v\n", err)
		return
	}

	var sourceIPAddresses []string
	var destinationIPAddresses []string
	var protocol v1.Protocol
	var sourcePortRange *v1.PortRange
	var destinationPortRange *v1.PortRange

	if len(args) > 6 {
		sourceIPAddresses = strings.Split(args[6], ",")
	}

	if len(args) > 7 {
		destinationIPAddresses = strings.Split(args[7], ",")
	}

	if len(args) > 8 {
		protocol, err = parseProtocol(args[8])
		if err != nil {
			log.Fatalf("Invalid protocol: %v\n", err)
			return
		}
	}

	if len(args) > 10 {
		portMin, err := strconv.Atoi(args[9])
		if err != nil {
			log.Fatalf("Invalid port range min: %v\n", err)
			return
		}
		portMax, err := strconv.Atoi(args[10])
		if err != nil {
			log.Fatalf("Invalid port range max: %v\n", err)
			return
		}
		sourcePortRange = &v1.PortRange{
			Min: uint32(portMin),
			Max: uint32(portMax),
		}
	}

	if len(args) > 12 {
		portMin, err := strconv.Atoi(args[11])
		if err != nil {
			log.Fatalf("Invalid port range min: %v\n", err)
			return
		}
		portMax, err := strconv.Atoi(args[12])
		if err != nil {
			log.Fatalf("Invalid port range max: %v\n", err)
			return
		}
		destinationPortRange = &v1.PortRange{
			Min: uint32(portMin),
			Max: uint32(portMax),
		}
	}

	csreq := v1.CreateSecurityRuleRequest{
		SecurityRuleId: &v1.SecurityRuleId{
			Uuid: uuid.NewString(),
		},
		Name:                    name,
		SecurityGroupId:         &v1.SecurityGroupId{Uuid: securityGroupUUID},
		VpcId:                   &v1.VPCId{Uuid: vpcId},
		Priority:                uint32(priority),
		Direction:               direction,
		Source_IPAddresses:      sourceIPAddresses,
		Destination_IPAddresses: destinationIPAddresses,
		Protocol:                &protocol,
		SourcePortRange:         sourcePortRange,
		DestinationPortRange:    destinationPortRange,
		Action:                  action,
	}

	ruleRet, err := c.CreateSecurityRule(ctx, &csreq)
	if err != nil {
		log.Fatalf("CMD: The security rule could not be created: %v\n", err)
		return
	}
	fmt.Printf("Security Rule Created: %v\n", ruleRet.SecurityRuleId.Uuid)
}

func deleteSecurityRule(uuid string) {
	dsreq := v1.DeleteSecurityRuleRequest{
		SecurityRuleId: &v1.SecurityRuleId{Uuid: uuid},
	}
	ruleRet, err := c.DeleteSecurityRule(ctx, &dsreq)
	if err != nil {
		log.Fatalf("The security rule could not be deleted %v\n", err)
		return
	}
	fmt.Printf("Security Rule deleted was : %v\n", ruleRet.SecurityRuleId.Uuid)
}

func parseDirection(direction string) (v1.Direction, error) {
	switch direction {
	case "ingress":
		return v1.Direction_INGRESS, nil
	case "egress":
		return v1.Direction_EGRESS, nil
	default:
		return 0, fmt.Errorf("invalid direction: %s", direction)
	}
}

func parseAction(action string) (v1.SecurityAction, error) {
	switch action {
	case "allow":
		return v1.SecurityAction_ALLOW, nil
	case "deny":
		return v1.SecurityAction_DENY, nil
	default:
		return 0, fmt.Errorf("invalid action: %s", action)
	}
}

func parseProtocol(protocol string) (v1.Protocol, error) {
	switch protocol {
	case "tcp":
		return v1.Protocol_TCP, nil
	case "udp":
		return v1.Protocol_UDP, nil
	case "icmp":
		return v1.Protocol_ICMP, nil
	default:
		return 0, fmt.Errorf("invalid protocol: %s", protocol)
	}
}

func getSecurityRule(uuid string) {
	req := v1.GetSecurityRuleRequest{
		SecurityRuleId: &v1.SecurityRuleId{Uuid: uuid},
	}
	ret, err := c.GetSecurityRule(ctx, &req)
	if err != nil {
		log.Fatalf("Failed to get Security Rule %v", err)
	}

	// Parse Direction
	directionStr, err := directionToString(ret.SecurityRule.Direction)
	if err != nil {
		log.Printf("Failed to parse direction: %v", err)
	}

	// Parse Action
	actionStr, err := actionToString(ret.SecurityRule.Action)
	if err != nil {
		log.Printf("Failed to parse action: %v", err)
	}

	// Parse Protocol
	protocolStr, err := protocolToString(ret.SecurityRule.Protocol)
	if err != nil {
		log.Printf("Failed to parse protocol: %v", err)
	}

	fmt.Printf(
		`Security Rule:
	ID: %v
	Name: %v
	Priority: %v
	Direction: %v
	Source IPs: %v
	Destination IPs: %v
	Protocol: %v
	Source Port Range: %v
	Destination Port Range: %v
	Action: %v
	VPC ID: %v
	`,
		ret.SecurityRule.Id.Uuid,
		ret.SecurityRule.Name,
		ret.SecurityRule.Priority,
		directionStr,
		ret.SecurityRule.Source_IPs,
		ret.SecurityRule.Destination_IPs,
		protocolStr,
		ret.SecurityRule.SourcePortRange,
		ret.SecurityRule.DestinationPortRange,
		actionStr,
		ret.SecurityRule.VpcId.Uuid)
}

func listSecurityRules() {
	securityRuleList, err := c.ListSecurityRules(ctx, &v1.ListSecurityRulesRequest{})
	if err != nil {
		log.Fatalf("could not get list of Security Rules: %v", err)
	}

	fmt.Println("Security Rule List:")
	for _, securityRuleId := range securityRuleList.SecurityRuleIds {
		fmt.Printf("ID: %v\n", securityRuleId.Uuid)
	}
}

func updateSecurityRule(args []string) {
	securityRuleUuid := args[0]
	priority, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatalf("Invalid priority: %v\n", err)
		return
	}

	direction, err := parseDirection(args[2])
	if err != nil {
		log.Fatalf("Invalid direction: %v\n", err)
		return
	}

	action, err := parseAction(args[3])
	if err != nil {
		log.Fatalf("Invalid action: %v\n", err)
		return
	}

	var sourceIPAddresses []string
	var destinationIPAddresses []string
	var protocol *v1.Protocol
	var sourcePortRange *v1.PortRange
	var destinationPortRange *v1.PortRange

	argIndex := 4

	if len(args) > argIndex {
		sourceIPAddresses = strings.Split(args[argIndex], ",")
		argIndex++
	}

	if len(args) > argIndex {
		destinationIPAddresses = strings.Split(args[argIndex], ",")
		argIndex++
	}

	if len(args) > argIndex {
		proto, err := parseProtocol(args[argIndex])
		if err != nil {
			log.Fatalf("Invalid protocol: %v\n", err)
			return
		}
		protocol = &proto
		argIndex++
	}

	if len(args) > argIndex+1 {
		portMin, err := strconv.Atoi(args[argIndex])
		if err != nil {
			log.Fatalf("Invalid source port range min: %v\n", err)
			return
		}
		portMax, err := strconv.Atoi(args[argIndex+1])
		if err != nil {
			log.Fatalf("Invalid source port range max: %v\n", err)
			return
		}
		sourcePortRange = &v1.PortRange{
			Min: uint32(portMin),
			Max: uint32(portMax),
		}
		argIndex += 2
	}

	if len(args) > argIndex+1 {
		portMin, err := strconv.Atoi(args[argIndex])
		if err != nil {
			log.Fatalf("Invalid destination port range min: %v\n", err)
			return
		}
		portMax, err := strconv.Atoi(args[argIndex+1])
		if err != nil {
			log.Fatalf("Invalid destination port range max: %v\n", err)
			return
		}
		destinationPortRange = &v1.PortRange{
			Min: uint32(portMin),
			Max: uint32(portMax),
		}
	}

	usreq := v1.UpdateSecurityRuleRequest{
		SecurityRuleId: &v1.SecurityRuleId{
			Uuid: securityRuleUuid,
		},
		Priority:                uint32(priority),
		Direction:               direction,
		Action:                  action,
		Source_IPAddresses:      sourceIPAddresses,
		Destination_IPAddresses: destinationIPAddresses,
		Protocol:                protocol,
		SourcePortRange:         sourcePortRange,
		DestinationPortRange:    destinationPortRange,
	}

	ruleRet, err := c.UpdateSecurityRule(ctx, &usreq)
	if err != nil {
		log.Fatalf("The security rule could not be updated: %v\n", err)
		return
	}
	fmt.Printf("Security Rule Updated: %v\n", ruleRet.SecurityRuleId.Uuid)
}

func directionToString(direction v1.Direction) (string, error) {
	switch direction {
	case v1.Direction_INGRESS:
		return "ingress", nil
	case v1.Direction_EGRESS:
		return "egress", nil
	default:
		return "", fmt.Errorf("invalid direction: %v", direction)
	}
}

func actionToString(action v1.SecurityAction) (string, error) {
	switch action {
	case v1.SecurityAction_ALLOW:
		return "allow", nil
	case v1.SecurityAction_DENY:
		return "deny", nil
	default:
		return "", fmt.Errorf("invalid action: %v", action)
	}
}

func protocolToString(protocol *v1.Protocol) (string, error) {
	if protocol == nil {
		return "", nil
	}
	switch *protocol {
	case v1.Protocol_TCP:
		return "tcp", nil
	case v1.Protocol_UDP:
		return "udp", nil
	case v1.Protocol_ICMP:
		return "icmp", nil
	default:
		return "", nil
	}
}
