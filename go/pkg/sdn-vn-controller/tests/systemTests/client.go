// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package systemTests

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
)

func createVPC(ctx context.Context, name, tenant, region string) (string, error) {
	vpcReq := v1.CreateVPCRequest{
		VpcId: &v1.VPCId{
			Uuid: uuid.NewString(),
		},
		Name:     name,
		TenantId: tenant,
		RegionId: region,
	}

	vpcResp, err := client.CreateVPC(ctx, &vpcReq)
	if err != nil {
		return "", fmt.Errorf("failed to create VPC %s: %v", name, err)
	}

	vpcID := vpcResp.GetVpcId().GetUuid()
	log.Printf("VPC created: %s\n", vpcID)
	return vpcID, nil
}

func deleteVPC(ctx context.Context, vpcID string) error {
	req := v1.DeleteVPCRequest{
		VpcId: &v1.VPCId{Uuid: vpcID},
	}
	ret, err := client.DeleteVPC(ctx, &req)
	if err != nil {
		log.Fatalf("VPC could not be deleted %v", err)
		return err
	}
	fmt.Printf("VPC deleted: UUID %s\n", ret.VpcId.Uuid)
	return nil
}

func createSubnet(ctx context.Context, name, cidr, vpcID string) (string, error) {
	csreq := v1.CreateSubnetRequest{
		SubnetId: &v1.SubnetId{
			Uuid: uuid.NewString(),
		},
		Name:  name,
		Cidr:  cidr,
		VpcId: &v1.VPCId{Uuid: vpcID},
	}

	subnetRet, err := client.CreateSubnet(ctx, &csreq)
	if err != nil {
		return "", fmt.Errorf("failed to create subnet %s: %v", cidr, err)
	}
	//change that to chassisID when the field is changed to string in main
	subnetID := subnetRet.SubnetId.Uuid
	log.Printf("Subnet created with CIDR %s: UUID %s\n", cidr, subnetID)
	return subnetID, nil
}

func createPort(ctx context.Context, subnetID, ipAddr string, chassisID string, deviceID int, MAC string) (string, error) {
	portReq := &v1.CreatePortRequest{
		PortId: &v1.PortId{
			Uuid: uuid.NewString(),
		},
		SubnetId: &v1.SubnetId{
			Uuid: subnetID,
		},
		ChassisId:          chassisID,
		DeviceId:           uint32(deviceID),
		Internal_IPAddress: ipAddr,
		IsEnabled:          true,
		IsNAT:              false,
		MACAddress:         MAC,
	}

	portResp, err := client.CreatePort(ctx, portReq)
	if err != nil {
		return "", fmt.Errorf("failed to create port with IP %s: %v", ipAddr, err)
	}

	portID := portResp.GetPortId().GetUuid() // Retrieve UUID from the response
	log.Printf("Port created with IP %s in subnet %s: UUID %s\n", ipAddr, subnetID, portID)
	return portID, nil
}

func deleteSubnet(ctx context.Context, subnetID string) error {
	dsreq := v1.DeleteSubnetRequest{
		SubnetId: &v1.SubnetId{Uuid: subnetID},
	}
	subnetRet, err := client.DeleteSubnet(ctx, &dsreq)
	if err != nil {
		log.Fatalf("The subnet could not be deleted %v\n", err)
		return err
	}
	fmt.Printf("Subnet deleted: UUID %s\n", subnetRet.SubnetId.Uuid)
	return nil
}
func deletePort(ctx context.Context, portID string) error {
	dsreq := v1.DeletePortRequest{
		PortId: &v1.PortId{Uuid: portID},
	}
	portRet, err := client.DeletePort(ctx, &dsreq)
	if err != nil {
		log.Fatalf("The port could not be deleted %v\n", err)
		return err
	}
	fmt.Printf("Port deleted: UUID %s\n", portRet.PortId.Uuid)
	return nil
}

func createRouter(ctx context.Context, name, vpcID string) (string, error) {
	crreq := v1.CreateRouterRequest{
		RouterId: &v1.RouterId{
			Uuid: uuid.NewString(),
		},
		Name:             name,
		VpcId:            &v1.VPCId{Uuid: vpcID},
		AvailabilityZone: "AZ1",
	}

	routerResp, err := client.CreateRouter(ctx, &crreq)
	if err != nil {
		return "", fmt.Errorf("failed to create router %s: %v", name, err)
	}

	routerID := routerResp.GetRouterId().GetUuid()
	log.Printf("Router created: %s\n", routerID)
	return routerID, nil
}

func deleteRouter(ctx context.Context, routerID string) error {
	drreq := v1.DeleteRouterRequest{
		RouterId: &v1.RouterId{Uuid: routerID},
	}

	_, err := client.DeleteRouter(ctx, &drreq)
	if err != nil {
		return fmt.Errorf("failed to delete router %s: %v", routerID, err)
	}

	log.Printf("Router deleted: %s\n", routerID)
	return nil
}

func createRouterInterface(ctx context.Context, routerID, subnetID, interfaceIP, interfaceMAC string) (string, error) {
	cifreq := v1.CreateRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{
			Uuid: uuid.NewString(),
		},
		RouterId:      &v1.RouterId{Uuid: routerID},
		SubnetId:      &v1.SubnetId{Uuid: subnetID},
		Interface_IP:  interfaceIP,
		Interface_MAC: interfaceMAC,
	}

	routerIfResp, err := client.CreateRouterInterface(ctx, &cifreq)
	if err != nil {
		return "", fmt.Errorf("failed to create router interface for router %s and subnet %s: %v", routerID, subnetID, err)
	}

	routerInterfaceID := routerIfResp.RouterInterfaceId.Uuid
	log.Printf("Router interface created: %s\n", routerInterfaceID)
	return routerInterfaceID, nil
}

func deleteRouterInterface(ctx context.Context, routerInterfaceID string) error {
	difreq := v1.DeleteRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{Uuid: routerInterfaceID},
	}

	_, err := client.DeleteRouterInterface(ctx, &difreq)
	if err != nil {
		return fmt.Errorf("failed to delete router interface %s: %v", routerInterfaceID, err)
	}

	log.Printf("Router interface deleted: %s\n", routerInterfaceID)
	return nil
}

func createPortSecurityGroup(ctx context.Context, name, vpcID string, portIDs []string) (string, error) {
	var portUUIDs []*v1.PortId
	for _, id := range portIDs {
		portUUIDs = append(portUUIDs, &v1.PortId{Uuid: id})
	}

	csreq := v1.CreateSecurityGroupRequest{
		SecurityGroupId: &v1.SecurityGroupId{
			Uuid: uuid.NewString(),
		},
		Name:    name,
		PortIds: portUUIDs,
		VpcId:   &v1.VPCId{Uuid: vpcID},
		Type:    v1.SecurityGroupType_PORT,
	}

	groupRet, err := client.CreateSecurityGroup(ctx, &csreq)
	if err != nil {
		return "", fmt.Errorf("failed to create security group %s: %v", name, err)
	}

	groupID := groupRet.SecurityGroupId.Uuid
	log.Printf("Security Group created: %s\n", groupID)
	return groupID, nil
}

func createSubnetSecurityGroup(ctx context.Context, name, vpcID string, subnetIDs []string) (string, error) {

	var subnetUUIDs []*v1.SubnetId
	for _, id := range subnetIDs {
		subnetUUIDs = append(subnetUUIDs, &v1.SubnetId{Uuid: id})
	}

	csreq := v1.CreateSecurityGroupRequest{
		SecurityGroupId: &v1.SecurityGroupId{
			Uuid: uuid.NewString(),
		},
		Name:      name,
		SubnetIds: subnetUUIDs,
		VpcId:     &v1.VPCId{Uuid: vpcID},
		Type:      v1.SecurityGroupType_SUBNET,
	}

	groupRet, err := client.CreateSecurityGroup(ctx, &csreq)
	if err != nil {
		return "", fmt.Errorf("failed to create security group %s: %v", name, err)
	}

	groupID := groupRet.SecurityGroupId.Uuid
	log.Printf("Security Group created: %s\n", groupID)
	return groupID, nil
}

// deleteSecurityGroup deletes a security group by UUID.
func deleteSecurityGroup(ctx context.Context, securityGroupID string) error {
	dsreq := v1.DeleteSecurityGroupRequest{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: securityGroupID},
	}

	groupRet, err := client.DeleteSecurityGroup(ctx, &dsreq)
	if err != nil {
		return fmt.Errorf("failed to delete security group %s: %v", securityGroupID, err)
	}

	log.Printf("Security Group deleted: %s\n", groupRet.SecurityGroupId.Uuid)
	return nil
}

// createSecurityRule creates a security rule with specified parameters.
func createSecurityRule(ctx context.Context, name, vpcID, securityGroupId string, priority int, directionStr, actionStr, protocolStr string, sourceIPs, destinationIPs []string, sourcePortRangeStr, destinationPortRangeStr string) (string, error) {
	direction, err := parseDirection(directionStr)
	if err != nil {
		return "", err
	}

	action, err := parseAction(actionStr)
	if err != nil {
		return "", err
	}

	protocol, err := parseProtocol(protocolStr)
	if err != nil {
		return "", err
	}

	sourcePortRange, err := parsePortRange(sourcePortRangeStr)
	if err != nil {
		return "", err
	}

	destinationPortRange, err := parsePortRange(destinationPortRangeStr)
	if err != nil {
		return "", err
	}

	csreq := v1.CreateSecurityRuleRequest{
		SecurityRuleId: &v1.SecurityRuleId{
			Uuid: uuid.NewString(),
		},
		Name:                    name,
		VpcId:                   &v1.VPCId{Uuid: vpcID},
		SecurityGroupId:         &v1.SecurityGroupId{Uuid: securityGroupId},
		Priority:                uint32(priority),
		Direction:               direction,
		Action:                  action,
		Source_IPAddresses:      sourceIPs,
		Destination_IPAddresses: destinationIPs,
		Protocol:                &protocol,
		SourcePortRange:         sourcePortRange,
		DestinationPortRange:    destinationPortRange,
	}

	ruleRet, err := client.CreateSecurityRule(ctx, &csreq)
	if err != nil {
		return "", fmt.Errorf("failed to create security rule %s: %v", name, err)
	}

	ruleID := ruleRet.SecurityRuleId.Uuid
	log.Printf("Security Rule created: %s\n", ruleID)
	return ruleID, nil
}

// deleteSecurityRule deletes a security rule by UUID.
func deleteSecurityRule(ctx context.Context, securityRuleID string) error {
	dsreq := v1.DeleteSecurityRuleRequest{
		SecurityRuleId: &v1.SecurityRuleId{Uuid: securityRuleID},
	}

	ruleRet, err := client.DeleteSecurityRule(ctx, &dsreq)
	if err != nil {
		return fmt.Errorf("failed to delete security rule %s: %v", securityRuleID, err)
	}

	log.Printf("Security Rule deleted: %s\n", ruleRet.SecurityRuleId.Uuid)
	return nil
}

// Helper functions to parse direction, action, protocol, and port ranges.

func parseDirection(directionStr string) (v1.Direction, error) {
	switch strings.ToLower(directionStr) {
	case "ingress":
		return v1.Direction_INGRESS, nil
	case "egress":
		return v1.Direction_EGRESS, nil
	default:
		return v1.Direction_DIR_UNSPECIFIED, fmt.Errorf("unknown direction: %s", directionStr)
	}
}

func parseAction(actionStr string) (v1.SecurityAction, error) {
	switch strings.ToLower(actionStr) {
	case "allow":
		return v1.SecurityAction_ALLOW, nil
	case "deny":
		return v1.SecurityAction_DENY, nil
	default:
		return v1.SecurityAction_ACTION_UNSPECIFIED, fmt.Errorf("unknown action: %s", actionStr)
	}
}

func parseProtocol(protocolStr string) (v1.Protocol, error) {
	switch strings.ToLower(protocolStr) {
	case "tcp":
		return v1.Protocol_TCP, nil
	case "udp":
		return v1.Protocol_UDP, nil
	default:
		return 0, fmt.Errorf("unknown protocol: %s", protocolStr)
	}
}

func parsePortRange(rangeStr string) (*v1.PortRange, error) {
	if rangeStr == "" {
		return nil, nil
	}

	ports := strings.Split(rangeStr, "-")
	if len(ports) != 2 {
		return nil, fmt.Errorf("invalid port range: %s", rangeStr)
	}

	portMin, err := strconv.Atoi(ports[0])
	if err != nil {
		return nil, fmt.Errorf("invalid port range min: %v", err)
	}

	portMax, err := strconv.Atoi(ports[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port range max: %v", err)
	}

	return &v1.PortRange{
		Min: uint32(portMin),
		Max: uint32(portMax),
	}, nil
}

func updateSecurityRule(ctx context.Context, ruleID string, priority int, directionStr, actionStr, protocolStr string, sourceIPs, destinationIPs []string, sourcePortRangeStr, destinationPortRangeStr string) error {
	direction, err := parseDirection(directionStr)
	if err != nil {
		return err
	}

	action, err := parseAction(actionStr)
	if err != nil {
		return err
	}

	protocol, err := parseProtocol(protocolStr)
	if err != nil {
		return err
	}

	sourcePortRange, err := parsePortRange(sourcePortRangeStr)
	if err != nil {
		return err
	}

	destinationPortRange, err := parsePortRange(destinationPortRangeStr)
	if err != nil {
		return err
	}

	usreq := v1.UpdateSecurityRuleRequest{
		SecurityRuleId: &v1.SecurityRuleId{
			Uuid: ruleID,
		},
		Priority:                uint32(priority),
		Direction:               direction,
		Action:                  action,
		Source_IPAddresses:      sourceIPs,
		Destination_IPAddresses: destinationIPs,
		Protocol:                &protocol,
		SourcePortRange:         sourcePortRange,
		DestinationPortRange:    destinationPortRange,
	}

	_, err = client.UpdateSecurityRule(ctx, &usreq)
	if err != nil {
		return fmt.Errorf("failed to update security rule %s: %v", ruleID, err)
	}
	log.Printf("Security Rule updated: %s\n", ruleID)
	return nil
}

func updatePortSecurityGroup(ctx context.Context, securityGroupID string, portIDs []string) error {
	var portUUIDs []*v1.PortId
	for _, id := range portIDs {
		portUUIDs = append(portUUIDs, &v1.PortId{Uuid: id})
	}

	usreq := v1.UpdateSecurityGroupRequest{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: securityGroupID},
		PortIds:         portUUIDs,
	}

	// Call the update security group API
	groupRet, err := client.UpdateSecurityGroup(ctx, &usreq)
	if err != nil {
		return fmt.Errorf("the security group could not be updated: %v", err)
	}
	fmt.Printf("Security Group Updated: %v\n", groupRet.SecurityGroupId.Uuid)
	return nil
}

func updateSubnetSecurityGroup(ctx context.Context, securityGroupID string, subnetIDs []string) error {
	var subnetUUIDs []*v1.SubnetId
	for _, id := range subnetIDs {
		subnetUUIDs = append(subnetUUIDs, &v1.SubnetId{Uuid: id})
	}

	usreq := v1.UpdateSecurityGroupRequest{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: securityGroupID},
		SubnetIds:       subnetUUIDs,
	}

	// Call the update security group API
	groupRet, err := client.UpdateSecurityGroup(ctx, &usreq)
	if err != nil {
		return fmt.Errorf("the security group could not be updated: %v", err)
	}
	fmt.Printf("Security Group Updated: %v\n", groupRet.SecurityGroupId.Uuid)
	return nil
}
