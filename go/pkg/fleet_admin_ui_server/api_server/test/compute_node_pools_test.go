// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ = Describe("Compute Node Pools UI Integration Tests", Serial, func() {
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

	validateRequest := func(cloudAccountId, poolId, adminEmail, expectedErrSubstring string) {
		addReq := NewAddCloudAccountToComputeNodePool(cloudAccountId, poolId, adminEmail)
		_, err := fleetAdminUIServiceClient.AddCloudAccountToComputeNodePool(ctx, addReq)

		Expect(err).To(HaveOccurred())
		Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
		Expect(err.Error()).To(ContainSubstring(expectedErrSubstring))
	}

	It("AddCloudAccountToComputeNodePool: Admin Email validation", func() {
		By("validating that an empty admin email returns an error")
		validateRequest(cloudaccount.MustNewId(), generalPoolId, "", "createAdmin cannot be empty")

		By("validating that an invalid admin email returns an error")
		validateRequest(cloudaccount.MustNewId(), generalPoolId, "test", "invalid createAdmin format")
	})

	It("AddCloudAccountToComputeNodePool: CloudAccountId validation", func() {
		By("validating that an empty cloudAccountId returns an error")
		validateRequest("", generalPoolId, createAdmin, "invalid CloudAccountId")

		By("validating that a cloudAccountId with less than 12 digits returns an error")
		validateRequest("1234567", generalPoolId, createAdmin, "invalid CloudAccountId")

		By("validating that a cloudAccountId with more than 12 digits returns an error")
		validateRequest("1234567891011", generalPoolId, createAdmin, "invalid CloudAccountId")

		By("validating that a cloudAccountId with non-numeric characters returns an error")
		validateRequest("1234abcd8888", generalPoolId, createAdmin, "invalid CloudAccountId")

		By("validating that a cloudAccountId with special characters returns an error")
		validateRequest("1234@6789101", generalPoolId, createAdmin, "invalid CloudAccountId")
	})

	It("AddCloudAccountToComputeNodePool: PoolId validation", func() {
		By("validating that an empty poolId returns an error")
		validateRequest(cloudaccount.MustNewId(), "", createAdmin, "Invalid PoolId")

		By("validating that a poolId exceeding 42 characters returns an error")
		validateRequest(cloudaccount.MustNewId(), strings.Repeat("A", 43), createAdmin, "must not exceed 42 characters")

		By("validating that a poolId containing special characters returns an error")
		validateRequest(cloudaccount.MustNewId(), "pool234@", createAdmin, "make sure it begins and ends with an alphanumeric character")
	})

	It("Deleting a pool for a cloudAccount should succeed", func() {
		By("AddCloudAccountToComputeNodePool")
		cloudAccountId := cloudaccount.MustNewId()
		addReq := NewAddCloudAccountToComputeNodePool(cloudAccountId, generalPoolId, createAdmin)
		_, err := fleetAdminUIServiceClient.AddCloudAccountToComputeNodePool(ctx, addReq)
		Expect(err).Should(Succeed())

		By("DeleteCloudAccountFromComputeNodePool")
		deleteReq := NewDeleteCloudAccountFromComputeNodePool(cloudAccountId, generalPoolId)
		_, err = fleetAdminUIServiceClient.DeleteCloudAccountFromComputeNodePool(ctx, deleteReq)
		Expect(err).Should(Succeed())
	})

	It("Deleting a non existing pool for a cloudAccount should succeed", func() {
		By("AddCloudAccountToComputeNodePool")
		cloudAccountId := cloudaccount.MustNewId()
		addReq := NewAddCloudAccountToComputeNodePool(cloudAccountId, generalPoolId, createAdmin)
		_, err := fleetAdminUIServiceClient.AddCloudAccountToComputeNodePool(ctx, addReq)
		Expect(err).Should(Succeed())

		By("DeleteCloudAccountFromComputeNodePool")
		deleteReq := NewDeleteCloudAccountFromComputeNodePool(cloudAccountId, habanaPoolId)
		_, err = fleetAdminUIServiceClient.DeleteCloudAccountFromComputeNodePool(ctx, deleteReq)
		Expect(err).Should(Succeed())
	})

	It("Creating a pool should succeed", func() {
		By("PutComputeNodePool")
		createReq := NewPutComputeNodePool(generalPoolId, generalPoolName, generalPoolAccountManagerAgsRole)
		_, err := fleetAdminUIServiceClient.PutComputeNodePool(ctx, createReq)
		Expect(err).Should(Succeed())
	})

	It("Updating a pool should succeed", func() {
		By("PutComputeNodePool")
		createReq := NewPutComputeNodePool(generalPoolId, generalPoolName, generalPoolAccountManagerAgsRole)
		_, err := fleetAdminUIServiceClient.PutComputeNodePool(ctx, createReq)
		Expect(err).Should(Succeed())

		By("PutComputeNodePool")
		updatePoolName := generalPoolName + " Updated"
		updateReq := NewPutComputeNodePool(generalPoolId, updatePoolName, generalPoolAccountManagerAgsRole)
		_, err = fleetAdminUIServiceClient.PutComputeNodePool(ctx, updateReq)
		Expect(err).Should(Succeed())
	})

	It("Get cloudaccount(s) for a poolId should succeed", func() {
		By("AddCloudAccountToComputeNodePool")
		cloudAccountId1 := cloudaccount.MustNewId()
		addReq1 := NewAddCloudAccountToComputeNodePool(cloudAccountId1, generalPoolId, createAdmin)
		_, err := fleetAdminUIServiceClient.AddCloudAccountToComputeNodePool(ctx, addReq1)
		Expect(err).Should(Succeed())

		By("AddCloudAccountToComputeNodePool")
		addReq2 := NewAddCloudAccountToComputeNodePool(cloudAccountId1, habanaPoolId, createAdmin)
		_, err = fleetAdminUIServiceClient.AddCloudAccountToComputeNodePool(ctx, addReq2)
		Expect(err).Should(Succeed())

		By("AddCloudAccountToComputeNodePool")
		cloudAccountId2 := cloudaccount.MustNewId()
		addReq3 := NewAddCloudAccountToComputeNodePool(cloudAccountId2, generalPoolId, createAdmin)
		_, err = fleetAdminUIServiceClient.AddCloudAccountToComputeNodePool(ctx, addReq3)
		Expect(err).Should(Succeed())

		By("SearchCloudAccountsForComputeNodePool")
		getReq1 := NewSearchCloudAccountsForComputeNodePool(generalPoolId)
		getResp1, err := fleetAdminUIServiceClient.SearchCloudAccountsForComputeNodePool(ctx, getReq1)
		Expect(err).Should(Succeed())
		Expect(len(getResp1.CloudAccountsForComputeNodePool)).Should(Equal(2))

		By("SearchCloudAccountsForComputeNodePool")
		getReq2 := NewSearchCloudAccountsForComputeNodePool(habanaPoolId)
		getResp2, err := fleetAdminUIServiceClient.SearchCloudAccountsForComputeNodePool(ctx, getReq2)
		Expect(err).Should(Succeed())
		Expect(len(getResp2.CloudAccountsForComputeNodePool)).Should(Equal(1))
	})

	It("Get cloudaccount(s) for a non exisiting poolId should return no cloudaccounts", func() {
		By("AddCloudAccountToComputeNodePool")
		cloudAccountId := cloudaccount.MustNewId()
		addReq := NewAddCloudAccountToComputeNodePool(cloudAccountId, generalPoolId, createAdmin)
		_, err := fleetAdminUIServiceClient.AddCloudAccountToComputeNodePool(ctx, addReq)
		Expect(err).Should(Succeed())

		By("SearchCloudAccountsForComputeNodePool")
		getReq := NewSearchCloudAccountsForComputeNodePool(habanaPoolId)
		resp, err := fleetAdminUIServiceClient.SearchCloudAccountsForComputeNodePool(ctx, getReq)
		Expect(err).Should(Succeed())
		Expect(len(resp.CloudAccountsForComputeNodePool)).Should(Equal(0))
	})

	It("Search Pools for Pool Account Manager should return all the pools for AGS roles that the user has", func() {
		By("PutComputeNodePool")
		createReq1 := NewPutComputeNodePool(generalPoolId, generalPoolName, generalPoolAccountManagerAgsRole)
		_, err := fleetAdminUIServiceClient.PutComputeNodePool(ctx, createReq1)
		Expect(err).Should(Succeed())

		By("PutComputeNodePool")
		createReq2 := NewPutComputeNodePool(habanaPoolId, habanaPoolName, habanaPoolAccountManagerAgsRole)
		_, err = fleetAdminUIServiceClient.PutComputeNodePool(ctx, createReq2)
		Expect(err).Should(Succeed())

		By("SearchComputeNodePoolsForPoolAccountManager")
		searchCtx, err := NewSearchComputeNodePoolsForPoolAccountManager(ctx, []string{generalPoolAccountManagerAgsRole, habanaPoolAccountManagerAgsRole})
		Expect(err).Should(Succeed())
		searchResp, err := fleetAdminUIServiceClient.SearchComputeNodePoolsForPoolAccountManager(searchCtx, &emptypb.Empty{})
		Expect(err).Should(Succeed())
		Expect(len(searchResp.ComputeNodePoolsForPoolAccountManager)).Should(Equal(2))
	})
})
