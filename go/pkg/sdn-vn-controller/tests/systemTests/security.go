// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package systemTests

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("[3] Security Network Tests", Ordered, func() {
	It("should perform security tests between VMs on different subnets", func() {
		ctx := context.Background()

		vm1 := &config.VMs[0]
		vm2 := &config.VMs[1]
		vm3 := &config.VMs[2]
		vm4 := &config.VMs[3]
		vm5 := &config.VMs[4]
		vm6 := &config.VMs[5]

		vpcID, err := createVPC(ctx, "vpc1", "tenant1", "region1")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(deleteVPC, ctx, vpcID)

		// Define subnet IDs and CIDRs,
		subnetCIDRs := []string{"10.0.0.0/24", "11.0.0.0/24", "12.0.0.0/24"}
		subnetIDs := make([]string, len(subnetCIDRs))

		for i, cidr := range subnetCIDRs {
			subnetName := fmt.Sprintf("subnet-%d", i+1)
			subnetID, err := createSubnet(ctx, subnetName, cidr, vpcID)
			Expect(err).NotTo(HaveOccurred())
			subnetIDs[i] = subnetID
			DeferCleanup(deleteSubnet, ctx, subnetID)
		}

		// Create ports and assign IPs for VMs based on specified server chassis IDs
		vmIPs := []string{"10.0.0.1", "10.0.0.2", "11.0.0.1", "11.0.0.2", "12.0.0.1", "12.0.0.2"}
		portIDs := []string{}
		for i := range config.VMs {
			// Use assignIP to set the IP address in the namespace
			err = config.VMs[i].assignIP(vmIPs[i], "/24")
			Expect(err).NotTo(HaveOccurred())

			subnetID := subnetIDs[i/2] // Use subnetID based on VM group (2 VMs per subnet)
			// Create port with VM's assigned IP and register deletion
			portID, err := createPort(ctx, subnetID, vmIPs[i], config.VMs[i].Chassis, config.VMs[i].DeviceID, config.VMs[i].MAC)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(deletePort, ctx, portID)
			portIDs = append(portIDs, portID)
		}

		routerID, err := createRouter(ctx, "r1", vpcID)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(deleteRouter, ctx, routerID)

		// Create Router Interfaces and register for cleanup
		routerInterfaceIDs := make([]string, len(subnetIDs))
		routerInterfaceIPs := []string{"10.0.0.254/24", "11.0.0.254/24", "12.0.0.254/24"}
		routerInterfaceMACs := []string{"00:11:22:33:44:55", "00:11:22:33:44:66", "00:11:22:33:44:77"}
		for i, subnetID := range subnetIDs {
			routerInterfaceID, err := createRouterInterface(ctx, routerID, subnetID, routerInterfaceIPs[i], routerInterfaceMACs[i])
			Expect(err).NotTo(HaveOccurred())
			routerInterfaceIDs[i] = routerInterfaceID
			DeferCleanup(deleteRouterInterface, ctx, routerInterfaceID)
		}

		// Add default routes to VMs to reach other subnets via the router
		for i := range config.VMs {
			gatewayIP := strings.Split(routerInterfaceIPs[i/2], "/")[0] // Use corresponding router interface IP as gateway
			err = config.VMs[i].addDefaultGateway(gatewayIP)
			Expect(err).NotTo(HaveOccurred())
		}

		// Start servers on VMs
		Expect(vm1.startServer(8080, "tcp")).To(Succeed())
		DeferCleanup(vm1.stopServers)

		Expect(vm4.startServer(8080, "tcp")).To(Succeed())
		DeferCleanup(vm4.stopServers)

		Expect(vm5.startServer(8080, "tcp")).To(Succeed())
		Expect(vm5.startServer(8081, "tcp")).To(Succeed())
		Expect(vm5.startServer(8080, "udp")).To(Succeed())
		DeferCleanup(vm5.stopServers)

		Expect(vm6.startServer(8080, "tcp")).To(Succeed())
		Expect(vm6.startServer(8080, "udp")).To(Succeed())
		DeferCleanup(vm6.stopServers)

		// Test TCP and UDP connections before applying any security rules
		fmt.Println("Testing TCP connection before applying security rule (should allow all)")
		Expect(vm1.testConnection(vm4, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).To(Succeed()) // UDP
		Expect(vm1.testConnection(vm6, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).To(Succeed())

		// Create Security Group with Cleanup and associate it with the ports
		fmt.Println("Creating Security Group and associating it with ports")
		securityGroupID, err := createPortSecurityGroup(ctx, "group1", vpcID, []string{portIDs[0], portIDs[2]})
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Group Created: %s\n", securityGroupID)
		var group1Deleted bool
		DeferCleanup(func() {
			if !group1Deleted {
				err := deleteSecurityGroup(ctx, securityGroupID)
				Expect(err).NotTo(HaveOccurred(), "Failed to clean up group1")
			}
		})

		// Creating Security Rule with Cleanup
		fmt.Println("Creating Security Rule to restrict access")
		rule1ID, err := createSecurityRule(ctx, "rule1", vpcID, securityGroupID, 100, "ingress", "deny", "tcp", []string{"10.0.0.0/24"}, []string{"12.0.0.1/32"}, "0-65535", "8080-8080")
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Rule Created: %s\n", rule1ID)
		var rule1Deleted bool
		DeferCleanup(func() {
			if !rule1Deleted {
				err := deleteSecurityRule(ctx, rule1ID)
				Expect(err).NotTo(HaveOccurred(), "Failed to clean up rule1")
			}
		})

		// Test TCP/UDP connections after applying the security rule
		fmt.Println("Testing TCP and UDP connections after applying security rule")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm4, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).To(Succeed())
		Expect(vm5.testConnection(vm1, 8080, "tcp")).To(Succeed())

		// Update the Security Rule
		fmt.Println("Updating Security Rule to change allowed and denied connections")
		err = updateSecurityRule(ctx, rule1ID, 100, "ingress", "deny", "udp", []string{"10.0.0.0/24"}, []string{"12.0.0.0/24"}, "0-65535", "8080-8080")
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Rule Updated: %s\n", rule1ID)

		fmt.Println("Testing TCP and UDP connections after updating the security rule")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).ToNot(Succeed())

		// Create the second security rule to test address sets
		fmt.Println("Creating the second security rule to test address sets")
		rule2ID, err := createSecurityRule(ctx, "rule2", vpcID, securityGroupID, 100, "ingress", "deny", "tcp",
			[]string{"10.0.0.0/24", "11.0.0.1/32", "11.0.0.2/32"}, // Source address set
			[]string{"12.0.0.1/32", "12.0.0.2/32"},                // Destination address set
			"0-65535", "8080-8080")
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Rule Created: %s\n", rule2ID)
		var rule2Deleted bool
		DeferCleanup(func() {
			if !rule2Deleted {
				err := deleteSecurityRule(ctx, rule2ID)
				Expect(err).NotTo(HaveOccurred(), "Failed to clean up rule2")
			}
		})

		// Update the security group to add the new rule and update ports
		fmt.Println("Updating the security group to update ports")
		err = updatePortSecurityGroup(ctx, securityGroupID, []string{portIDs[0], portIDs[1]})
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Group Updated: %s\n", securityGroupID)

		// Step 4: Test TCP/UDP connections after updating the security group
		fmt.Println("Testing TCP and UDP connections after updating the security group")

		Expect(vm1.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).ToNot(Succeed())

		// Update the security group to remove ports
		fmt.Println("Updating the security group to remove ports")
		err = updatePortSecurityGroup(ctx, securityGroupID, []string{})
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Group Updated: %s\n", securityGroupID)

		fmt.Println("Testing if all TCP/UDP connections are allowed after removing ports")
		// Perform the TCP and UDP connection tests to verify all connections are allowed
		Expect(vm1.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).To(Succeed())

		// Update the security group to include subnets
		fmt.Println("Creating a new security group to use subnets")
		securityGroupID2, err := createSubnetSecurityGroup(ctx, "group2", vpcID, []string{subnetIDs[0], subnetIDs[1]})
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Group Created: %s\n", securityGroupID2)
		var group2Deleted bool
		DeferCleanup(func() {
			if !group2Deleted {
				err := deleteSecurityGroup(ctx, securityGroupID2)
				Expect(err).NotTo(HaveOccurred(), "Failed to clean up group2")
			}
		})

		fmt.Println("Creating 3rd Security Rule for group2")
		rule3ID, err := createSecurityRule(ctx, "rule3", vpcID, securityGroupID2, 100, "ingress", "deny", "udp", []string{"10.0.0.1/32"}, []string{"12.0.0.0/24"}, "0-65535", "8080-8080")
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Rule Created: %s\n", rule3ID)
		var rule3Deleted bool
		DeferCleanup(func() {
			if !rule3Deleted {
				err := deleteSecurityRule(ctx, rule3ID)
				Expect(err).NotTo(HaveOccurred(), "Failed to clean up rule3")
			}
		})

		fmt.Println("Creating the 4th security rule to test address sets with group2")
		rule4ID, err := createSecurityRule(ctx, "rule4", vpcID, securityGroupID2, 100, "ingress", "deny", "tcp",
			[]string{"10.0.0.0/24", "11.0.0.1/32", "11.0.0.2/32"}, // Source address set
			[]string{"12.0.0.1/32", "12.0.0.2/32"},                // Destination address set
			"0-65535", "8080-8080")
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Rule Created: %s\n", rule4ID)
		var rule4Deleted bool
		DeferCleanup(func() {
			if !rule4Deleted {
				err := deleteSecurityRule(ctx, rule4ID)
				Expect(err).NotTo(HaveOccurred(), "Failed to clean up rule4")
			}
		})

		fmt.Println("Testing TCP/UDP connections after creating the security group and rules to use subnets")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).ToNot(Succeed())

		// Update the security group to test removing a subnet from the group
		fmt.Println("Updating the security group to test removing a subnet from the group")
		err = updateSubnetSecurityGroup(ctx, securityGroupID2, []string{subnetIDs[1]})
		Expect(err).NotTo(HaveOccurred(), "Expected security group update with subnet removal to succeed")

		fmt.Println("Testing TCP/UDP connections after removing a subnet from the security group")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).To(Succeed())

		// Update the security group to test adding a subnet to a group
		fmt.Println("Updating the security group to test adding a subnet to a group")
		err = updateSubnetSecurityGroup(ctx, securityGroupID2, []string{subnetIDs[0], subnetIDs[1]})
		Expect(err).NotTo(HaveOccurred(), "Expected security group update with added subnets to succeed")

		fmt.Println("Testing TCP/UDP connections after updating the security group")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).ToNot(Succeed())

		// Create new security group for port5
		fmt.Println("Creating new security group for port5")
		securityGroupID3, err := createPortSecurityGroup(ctx, "port5", vpcID, []string{portIDs[4]})
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Group Created: %s\n", securityGroupID3)
		var group3Deleted bool
		DeferCleanup(func() {
			if !group3Deleted {
				err := deleteSecurityGroup(ctx, securityGroupID3)
				Expect(err).NotTo(HaveOccurred(), "Failed to clean up group2")
			}
		})

		// Create a 5th security rule to test egress direction
		fmt.Println("Creating a 5th security rule to test egress direction and a new group")
		rule5ID, err := createSecurityRule(ctx, "rule5", vpcID, securityGroupID3, 100, "egress", "deny", "tcp", []string{"0.0.0.0/0"}, []string{"12.0.0.1/32"}, "0-65535", "8081-8081")
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Rule Created: %s\n", rule5ID)
		var rule5Deleted bool
		DeferCleanup(func() {
			if !rule5Deleted {
				err := deleteSecurityRule(ctx, rule5ID)
				Expect(err).NotTo(HaveOccurred(), "Failed to clean up rule3")
			}
		})
		Expect(vm1.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).ToNot(Succeed())

		fmt.Println("Updating the security group 2 to remove subnets")
		err = updateSubnetSecurityGroup(ctx, securityGroupID2, []string{})
		Expect(err).NotTo(HaveOccurred(), "Expected security group2 update to remove subnets to succeed")

		fmt.Println("Testing TCP/UDP connections after removing all subnets from the security group2")

		Expect(vm1.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).To(Succeed())

		fmt.Println("Deleting the 3rd security rule")
		err = deleteSecurityRule(ctx, rule3ID)
		Expect(err).NotTo(HaveOccurred())
		rule3Deleted = true

		fmt.Println("Deleting the 4th security rule")
		err = deleteSecurityRule(ctx, rule4ID)
		Expect(err).NotTo(HaveOccurred())
		rule4Deleted = true

		// Delete security group 2
		err = deleteSecurityGroup(ctx, securityGroupID2)
		Expect(err).NotTo(HaveOccurred(), "Failed to delete security group 2")
		group2Deleted = true

		// Create a fourth security rule to test priority
		fmt.Println("Creating a 6th security rule to test priority")
		rule6ID, err := createSecurityRule(ctx, "rule6", vpcID, securityGroupID, 120, "ingress", "allow", "tcp", []string{"10.0.0.2/32"}, []string{"12.0.0.1/32"}, "0-65535", "8080-8080")
		Expect(err).NotTo(HaveOccurred())
		fmt.Printf("Security Rule Created: %s\n", rule6ID)
		var rule6Deleted bool
		DeferCleanup(func() {
			if !rule6Deleted {
				err := deleteSecurityRule(ctx, rule6ID)
				Expect(err).NotTo(HaveOccurred(), "Failed to clean up rule4")
			}
		})
		err = updatePortSecurityGroup(ctx, securityGroupID, []string{portIDs[0], portIDs[2], portIDs[3]})
		Expect(err).NotTo(HaveOccurred(), "Expected security group1 update with high priority rule to succeed")

		// time.Sleep(60 * time.Second)

		fmt.Println("Testing TCP/UDP connections after adding the high priority rule to the security group")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).ToNot(Succeed())

		err = updatePortSecurityGroup(ctx, securityGroupID, []string{portIDs[0], portIDs[1], portIDs[2], portIDs[3]})
		Expect(err).NotTo(HaveOccurred(), "Expected security group1 update to succeed")

		// time.Sleep(60 * time.Second)
		fmt.Println("Add vm2 to port group1")
		fmt.Println("Testing TCP/UDP connections after adding vm2 to group1")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).ToNot(Succeed())

		// Delete the fourth security rule and test connections
		fmt.Println("Deleting the 6th security rule and testing connections")
		err = deleteSecurityRule(ctx, rule6ID)
		Expect(err).NotTo(HaveOccurred())
		rule6Deleted = true

		fmt.Println("Testing TCP/UDP connections after deleting the 6th (high priority) security rule")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).ToNot(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).ToNot(Succeed())

		// Delete the first security rule
		fmt.Println("Deleting the first security rule")
		err = deleteSecurityRule(ctx, rule1ID)
		Expect(err).NotTo(HaveOccurred())
		rule1Deleted = true

		fmt.Println("Testing TCP/UDP connections after deleting the UDP security tule rule1")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).To(Succeed())

		// Update the security group to test removing ports
		fmt.Println("Updating the security group to test removing ports")
		err = updatePortSecurityGroup(ctx, securityGroupID, []string{portIDs[2], portIDs[3]})
		Expect(err).NotTo(HaveOccurred(), "Expected security group update to remove ports to succeed")

		fmt.Println("Testing TCP/UDP connections after updating the security group")
		Expect(vm1.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).ToNot(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).To(Succeed())

		fmt.Println("Deleting the 2nd security rule")
		err = deleteSecurityRule(ctx, rule2ID)
		Expect(err).NotTo(HaveOccurred())
		rule2Deleted = true

		fmt.Println("Deleting the security groups to test all connections")
		// Delete security group 1
		err = deleteSecurityGroup(ctx, securityGroupID)
		Expect(err).NotTo(HaveOccurred(), "Failed to delete security group1")
		group1Deleted = true

		fmt.Println("Deleting the 5th security rule")
		err = deleteSecurityRule(ctx, rule5ID)
		Expect(err).NotTo(HaveOccurred())
		rule5Deleted = true

		// Delete security group 2
		err = deleteSecurityGroup(ctx, securityGroupID3)
		Expect(err).NotTo(HaveOccurred(), "Failed to delete security group3")
		group3Deleted = true
		fmt.Println("Testing if all TCP/UDP connections are allowed after deleting all security groups")

		// Perform the TCP and UDP connection tests to verify all connections are allowed
		Expect(vm1.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8081, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm2.testConnection(vm5, 8080, "udp")).To(Succeed())
		Expect(vm3.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm4.testConnection(vm5, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "tcp")).To(Succeed())
		Expect(vm1.testConnection(vm6, 8080, "udp")).To(Succeed())
	})
})
