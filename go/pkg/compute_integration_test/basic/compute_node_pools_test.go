// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package basic

import (
	"context"
	"net/http"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	fleetadmintest "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin/api_server/test"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compute Node Pools Tests", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	It("Instance creation should succeed with an explicit allowed pool for a Cloud Account", func() {
		e := NewBasicTestEnv(BasicTestEnvOptions{})
		defer e.Stop()
		e.Start()

		By("Created node with compute node pool " + e.computeNodePoolId)

		cloudAccountId := cloudaccount.MustNewId()
		e.namespace = cloudAccountId
		instanceType := CreateInstanceType(ctx)
		machineImage := CreateMachineImage(ctx)
		vNetName := CreateVNet(ctx, cloudAccountId)
		sshPublicKeyName := CreateSshPublicKey(ctx, cloudAccountId)

		By("UpdateComputeNodePoolsForCloudAccount")
		updateReq := fleetadmintest.NewUpdateComputeNodePoolsForCloudAccountRequest(cloudAccountId, []string{e.computeNodePoolId}, "idcadmin@intel.com")
		_, err := e.fleetAdminServiceClient.UpdateComputeNodePoolsForCloudAccount(ctx, updateReq)
		Expect(err).Should(Succeed())

		By("InstanceServiceCreate")
		createInstanceReq := NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName)
		createResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
		Expect(err).Should(Succeed())
		e.instanceResourceId = *createResp.Metadata.ResourceId
	})

	It("Instance creation should fail if no nodes are in the pools allowed by the Cloud Account", func() {
		e := NewBasicTestEnv(BasicTestEnvOptions{})
		defer e.Stop()
		e.Start()

		cloudAccountId := cloudaccount.MustNewId()
		e.namespace = cloudAccountId
		instanceType := CreateInstanceType(ctx)
		machineImage := CreateMachineImage(ctx)
		vNetName := CreateVNet(ctx, cloudAccountId)
		sshPublicKeyName := CreateSshPublicKey(ctx, cloudAccountId)

		By("UpdateComputeNodePoolsForCloudAccount")
		updateReq := fleetadmintest.NewUpdateComputeNodePoolsForCloudAccountRequest(cloudAccountId, []string{"pool-without-nodes"}, "idcadmin@intel.com")
		_, err := e.fleetAdminServiceClient.UpdateComputeNodePoolsForCloudAccount(ctx, updateReq)
		Expect(err).Should(Succeed())

		By("InstanceServiceCreate")
		createInstanceReq := NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName)
		_, httpResp, err := openApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		// gRPC ResourceExhausted status code will be translated to HTTP status code 429 (Too Many Requests).
		Expect(httpResp.StatusCode).Should(Equal(http.StatusTooManyRequests))
	})
})
