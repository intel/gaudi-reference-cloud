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

var _ = Describe("[2] L3 Network Tests", Ordered, func() {
	It("should perform connectivity tests between VMs on different subnets", func() {
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
		for i := range config.VMs {
			// Use assignIP to set the IP address in the namespace
			err = config.VMs[i].assignIP(vmIPs[i], "/24")
			Expect(err).NotTo(HaveOccurred())

			subnetID := subnetIDs[i/2] // Use subnetID based on VM group (2 VMs per subnet)
			// Create port with VM's assigned IP and register deletion
			portID, err := createPort(ctx, subnetID, vmIPs[i], config.VMs[i].Chassis, config.VMs[i].DeviceID, config.VMs[i].MAC)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(deletePort, ctx, portID)
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

		// Test connectivity btween subnets
		Expect(vm1.testPingVM(vm2)).To(Succeed(), "Ping from vm1 to vm2 should succeed")
		Expect(vm1.testPingVM(vm3)).To(Succeed(), "Ping from vm1 to vm3 should succeed")
		Expect(vm1.testPingVM(vm4)).To(Succeed(), "Ping from vm1 to vm4 should succeed")
		Expect(vm1.testPingVM(vm5)).To(Succeed(), "Ping from vm1 to vm5 should succeed")
		Expect(vm1.testPingVM(vm6)).To(Succeed(), "Ping from vm1 to vm6 should succeed")
		Expect(vm3.testPingVM(vm5)).To(Succeed(), "Ping from vm1 to vm5 should succeed")
		Expect(vm4.testPingVM(vm6)).To(Succeed(), "Ping from vm1 to vm6 should succeed")

		// Pings to router interfaces
		Expect(vm1.testPing("10.0.0.254")).To(Succeed(), "Ping from vm1 to router interface 10.0.0.254 should succeed")
		Expect(vm3.testPing("11.0.0.254")).To(Succeed(), "Ping from vm3 to router interface 11.0.0.254 should succeed")
		Expect(vm5.testPing("12.0.0.254")).To(Succeed(), "Ping from vm5 to router interface 12.0.0.254 should succeed")
	})
})
