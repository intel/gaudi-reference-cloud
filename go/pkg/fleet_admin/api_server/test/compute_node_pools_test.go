// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	mrand "math/rand"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compute Node Pools Integration Tests", Serial, func() {
	const (
		generalPoolId                    = "general"
		generalPoolName                  = "General Purpose"
		generalPoolAccountManagerAgsRole = "IDC.PoolAccountManager-General"
		habanaPoolId                     = "habana"
		habanaPoolName                   = "AI Workload Training Purpose"
		habanaPoolAccountManagerAgsRole  = "IDC.PoolAccountManager-Habana"
		createAdmin                      = "idcadmin@intel.com"
	)
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	It("Duplicate update pool should succeed", func() {
		cloudAccountId := cloudaccount.MustNewId()
		updateReq := NewUpdateComputeNodePoolsForCloudAccountRequest(cloudAccountId, []string{generalPoolId}, createAdmin)
		_, err := fleetAdminServiceClient.UpdateComputeNodePoolsForCloudAccount(ctx, updateReq)
		Expect(err).Should(Succeed())
		_, err = fleetAdminServiceClient.UpdateComputeNodePoolsForCloudAccount(ctx, updateReq)
		Expect(err).Should(Succeed())
	})

	It("Update and get pool should succeed", func() {
		By("UpdateComputeNodePoolsForCloudAccount")
		cloudAccountId := cloudaccount.MustNewId()
		updateReq := NewUpdateComputeNodePoolsForCloudAccountRequest(cloudAccountId, []string{generalPoolId}, createAdmin)
		_, err := fleetAdminServiceClient.UpdateComputeNodePoolsForCloudAccount(ctx, updateReq)
		Expect(err).Should(Succeed())

		By("SearchComputeNodePoolsForInstanceScheduling")
		getReq := NewSearchComputeNodePoolsForInstanceSchedulingRequest(cloudAccountId)
		getResp, err := fleetAdminServiceClient.SearchComputeNodePoolsForInstanceScheduling(ctx, getReq)
		Expect(err).Should(Succeed())
		Expect(len(getResp.ComputeNodePools)).Should(Equal(1))
		Expect(getResp.ComputeNodePools[0].PoolId).Should(Equal(generalPoolId))
	})

	It("Update without create admin should return error", func() {
		cloudAccountId := cloudaccount.MustNewId()
		updateReq := NewUpdateComputeNodePoolsForCloudAccountRequest(cloudAccountId, []string{generalPoolId}, "")
		_, err := fleetAdminServiceClient.UpdateComputeNodePoolsForCloudAccount(ctx, updateReq)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("create admin cannot be empty. provide a valid email address"))
	})

	It("should successfully report node statistics with valid data, verify all fields, and delete a node", func() {
		randInput := mrand.New(mrand.NewSource(41))
		By("Reporting valid node statistics")
		// Create request with 2 nodes
		numberOfNodes := 2
		createReq1 := NewReportNodeStatisticsRequest(numberOfNodes, randInput)

		// Report the node statistics using the client
		_, err := fleetAdminServiceClient.ReportNodeStatistics(ctx, createReq1)
		Expect(err).Should(Succeed())

		// Remove one node from the SchedulerNodeStatistics list
		By("Deleting one node from SchedulerNodeStatistics")
		createReq1.SchedulerNodeStatistics = createReq1.SchedulerNodeStatistics[1:]

		// Now, report the updated node statistics (with only one node)
		_, err = fleetAdminServiceClient.ReportNodeStatistics(ctx, createReq1)
		Expect(err).Should(Succeed())

		// Update Node Statistics
		By("Updating Node Statistics")
		// Modify the node's statistics
		createReq1.SchedulerNodeStatistics[0].NodeResources.FreeMilliCPU -= 500
		createReq1.SchedulerNodeStatistics[0].NodeResources.UsedMilliCPU += 500
		createReq1.SchedulerNodeStatistics[0].NodeResources.FreeMemoryBytes -= 1024
		createReq1.SchedulerNodeStatistics[0].NodeResources.UsedMemoryBytes += 1024

		// Remove an instance type from the node's InstanceTypeStatistics
		if len(createReq1.SchedulerNodeStatistics[0].InstanceTypeStatistics) > 0 {
			createReq1.SchedulerNodeStatistics[0].InstanceTypeStatistics = createReq1.SchedulerNodeStatistics[0].InstanceTypeStatistics[1:]
		}

		// Add a new instance type to the node's InstanceTypeStatistics
		newInstanceType := &pb.InstanceTypeStatistics{
			InstanceType:     "vm-spr-lrg",
			InstanceCategory: "VirtualMachine",
			RunningInstances: 5,
			MaxNewInstances:  10,
		}
		createReq1.SchedulerNodeStatistics[0].InstanceTypeStatistics = append(createReq1.SchedulerNodeStatistics[0].InstanceTypeStatistics, newInstanceType)

		// Report the updated node statistics
		_, err = fleetAdminServiceClient.ReportNodeStatistics(ctx, createReq1)
		Expect(err).Should(Succeed())
	})

})
