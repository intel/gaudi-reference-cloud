// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bm

import (
	"context"
	"encoding/json"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

func CreateInstanceGroupFlow(ctx context.Context, instanceCount int32) {
	log := log.FromContext(ctx)
	_ = log

	cloudAccountId := cloudAccount
	log.Info("Created CloudAccount", "cloudAccountId", cloudAccountId)
	instanceType := "bm-virtual-sc"
	instanceGroupName := "automate-bm-group-" + computeTestHelper.GetRandomStringWithLimit(5)
	machineImage := "ubuntu-22.04-server-cloudimg-amd64-latest"

	createInstanceGroupReq := computeTestHelper.NewCreateInstanceGroupRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName, "us-dev-1b", instanceGroupName, instanceCount)
	createInstanceGroupResp, err := computeTestHelper.CreateInstanceGroup(ctx, cloudAccountId, createInstanceGroupReq)
	Expect(err).Should(Succeed())
	log.Info("Created Instance Group", "createInstanceGroupResp", createInstanceGroupResp)

	// Get list of instance Ids from the Instance Group
	searchInstanceGroupBody := computeTestHelper.NewSearchInstanceGroupRequest(instanceGroupName)
	response_body, _ := computeTestHelper.SearchInstanceGroup(ctx, cloudAccountId, searchInstanceGroupBody)
	respJsonBytes, _ := json.Marshal(response_body)
	respJsonString := string(respJsonBytes)
	ids := gjson.Get(respJsonString, "items.#.metadata.resourceId").Array()
	listOfIds := []string{}
	for _, id := range ids {
		listOfIds = append(listOfIds, id.String())
	}
	log.Info("List of instance ids in the group", "list of Ids", listOfIds)

	By("Waiting for instances in instance group to have SSH proxy address")
	for _, id := range listOfIds {
		currentResourceId := id
		Eventually(func(g Gomega) {
			instance, err := computeTestHelper.GetInstance(ctx, cloudAccountId, currentResourceId)
			log.Info("Get Instance in Instance Group", "instance", instance)
			g.Expect(err).Should(Succeed())
			g.Expect(instance.Status).ShouldNot(BeNil())
			g.Expect(instance.Status.SshProxy).ShouldNot(BeNil())
			g.Expect(*instance.Status.SshProxy.ProxyAddress).ShouldNot(BeEmpty())
		}, "30s", "2s").Should(Succeed())
	}

	By("Waiting for instances in instance group to start")
	for _, id := range listOfIds {
		currentResourceId := id
		skipWaitForInstanceReady := false
		Eventually(func(g Gomega) {
			instance, err := computeTestHelper.GetInstance(ctx, cloudAccountId, currentResourceId)
			log.Info("Get Instance in Instance Group", "instance", instance)
			g.Expect(err).Should(Succeed())
			if skipWaitForInstanceReady {
				g.Expect(*instance.Status.Message).Should(ContainSubstring("Instance specification has been accepted and is being provisioned."))
			} else {
				g.Expect(string(*instance.Status.Phase)).Should(Equal("Ready"))
			}
		}, "20m", "2s").Should(Succeed())
	}

	// TODO: SSH into instance.

	// Instance group deletion
	response := computeTestHelper.DeleteInstanceGroupViaResty(ctx, instancegroupEndpoint, cloudAccountId, instanceGroupName)
	Expect(response.StatusCode()).To(Equal(200))

	By("Waiting for instance group to be not found")
	Eventually(func(g Gomega) {
		g.Expect(computeTestHelper.CheckInstanceGroupNotFound(ctx, cloudAccountId)).Should(Succeed())
	}, "3m", "1s").Should(Succeed())
}

var _ = Describe("Compute E2E BM Instance Group Test - (non-bgp)", Ordered, ContinueOnFailure, Label("large"), func() {
	ctx := context.Background()

	It("Instance Group (non-bgp) creation happy path", func() {
		CreateInstanceGroupFlow(ctx, 2)
	})
})

var _ = Describe("Compute E2E BM Instance Group Test - BGP", Ordered, ContinueOnFailure, Label("large"), func() {
	ctx := context.Background()

	It("Instance Group (bgp) creation happy path", func() {
		CreateInstanceGroupFlow(ctx, 6)
	})
})
