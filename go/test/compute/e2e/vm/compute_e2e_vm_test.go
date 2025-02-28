// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vm

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compute E2E VM Test", Serial, Label("large"), func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	It("Instance creation happy path", func() {
		cloudAccountId := cloudAccount
		log.Info("Created CloudAccount", "cloudAccountId", cloudAccountId)
		instanceType := "vm-spr-sml"
		machineImage := "ubuntu-2204-jammy-v20230122"
		availabilityZone := "us-dev-1a"
		vNetName, _ := computeTestHelper.CreateVNet(ctx, cloudAccountId, "us-dev-1a-default", availabilityZone)
		log.Info("Created VNet", "vNetName", vNetName)

		sshPublicKeyName, _, publicKey, _ := computeTestHelper.CreateSshPublicKey(ctx, cloudAccountId)
		log.Info("Created SSH Key Pair", "sshPublicKeyName", sshPublicKeyName, "publicKey", publicKey)

		createPrivateInstance := true
		var instanceResourceId string
		if createPrivateInstance {
			createInstanceReq := computeTestHelper.NewCreateInstancePrivateRequestGrpc(cloudAccountId, []string{sshPublicKeyName}, instanceType, machineImage, vNetName, availabilityZone)
			createInstanceResp, err := computeTestHelper.CreateInstancePrivateGrpc(ctx, createInstanceReq)
			Expect(err).Should(Succeed())
			instanceResourceId = createInstanceResp.Metadata.ResourceId
			log.Info("Created Instance", "createInstanceResp", createInstanceResp)

		} else {
			createInstanceReq := computeTestHelper.NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName, availabilityZone)
			createInstanceResp, err := computeTestHelper.CreateInstance(ctx, cloudAccountId, createInstanceReq)
			Expect(err).Should(Succeed())
			instanceResourceId = *createInstanceResp.Metadata.ResourceId
			log.Info("Created Instance", "createInstanceResp", createInstanceResp)
		}
		_, err := computeTestHelper.GetInstance(ctx, cloudAccountId, instanceResourceId)
		Expect(err).Should(Succeed())

		By("Waiting for instance to have SSH proxy address")
		Eventually(func(g Gomega) {
			instance, err := computeTestHelper.GetInstance(ctx, cloudAccountId, instanceResourceId)
			log.Info("Get Instance", "instance", instance)
			g.Expect(err).Should(Succeed())
			g.Expect(instance.Status).ShouldNot(BeNil())
			g.Expect(instance.Status.SshProxy).ShouldNot(BeNil())
			g.Expect(*instance.Status.SshProxy.ProxyAddress).ShouldNot(BeEmpty())
		}, "30s", "2s").Should(Succeed())

		By("Waiting for VM to start")
		skipWaitForInstanceReady := false
		Eventually(func(g Gomega) {
			instance, err := computeTestHelper.GetInstance(ctx, cloudAccountId, instanceResourceId)
			log.Info("Get Instance", "instance", instance)
			g.Expect(err).Should(Succeed())
			if skipWaitForInstanceReady {
				g.Expect(*instance.Status.Message).Should(ContainSubstring("Instance specification has been accepted and is being provisioned."))
			} else {
				g.Expect(string(*instance.Status.Phase)).Should(Equal("Ready"))
			}
		}, "4m", "2s").Should(Succeed())

		// TODO: SSH into instance.

		// Instance deletion
		response := computeTestHelper.DeleteInstanceViaResty(ctx, instance_endpoint, cloudAccountId, instanceResourceId)
		Expect(response.StatusCode()).To(Equal(200))

		By("Waiting for instance to be not found")
		Eventually(func(g Gomega) {
			g.Expect(computeTestHelper.CheckInstanceNotFound(ctx, cloudAccountId, instanceResourceId)).Should(Succeed())
		}, "3m", "1s").Should(Succeed())
	})
})
