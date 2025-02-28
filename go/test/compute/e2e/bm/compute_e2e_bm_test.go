// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bm

import (
	"context"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Compute E2E BM Test", Serial, Label("large"), func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	It("Instance creation happy path", func() {
		cloudAccountId := cloudAccount
		log.Info("Created CloudAccount", "cloudAccountId", cloudAccountId)
		instanceType := "bm-virtual-sc"
		machineImage := "ubuntu-22.04-server-cloudimg-amd64-latest"

		createPrivateInstance := true
		var instanceResourceId string
		if createPrivateInstance {
			createInstanceReq := computeTestHelper.NewCreateInstancePrivateRequestGrpc(cloudAccountId, []string{sshPublicKeyName}, instanceType, machineImage, vNetName, "us-dev-1b")
			createInstanceResp, err := computeTestHelper.CreateInstancePrivateGrpc(ctx, createInstanceReq)
			Expect(err).Should(Succeed())
			instanceResourceId = createInstanceResp.Metadata.ResourceId
			log.Info("Created Instance", "createInstanceResp", createInstanceResp)

		} else {
			createInstanceReq := computeTestHelper.NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName, "us-dev-1b")
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

		By("Waiting for instance to start")
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
