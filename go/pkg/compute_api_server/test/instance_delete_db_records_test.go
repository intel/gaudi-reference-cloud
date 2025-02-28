// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instance Delete records ", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	It("Instance Delete records", Serial, func() {
		api := openApiClient.InstanceServiceApi
		grpcClient := getInstanceGrpcClient()
		cloudAccountId1 := cloudaccount.MustNewId()
		sshPublicKeyName1 := "name1-" + uuid.NewString()
		sshPublicKeyNames := []string{sshPublicKeyName1}
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		availabilityZone := "us-dev-1a"
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)

		runStrategy1Str := "RerunOnFailure"
		runStrategy1, err := openapi.NewProtoRunStrategyFromValue(runStrategy1Str)
		Expect(err).Should(Succeed())
		interfaces := []openapi.ProtoNetworkInterface{{VNet: &vNet}}

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)

		By("InstanceServiceCreate1")
		createResp, _, err := api.InstanceServiceCreate(ctx, cloudAccountId1).Body(
			openapi.InstanceServiceCreateRequest{
				Metadata: &openapi.InstanceServiceCreateRequestMetadata{},
				Spec: &openapi.ProtoInstanceSpec{
					AvailabilityZone:  &availabilityZone,
					InstanceType:      &instanceType,
					MachineImage:      &machineImage,
					RunStrategy:       runStrategy1,
					SshPublicKeyNames: sshPublicKeyNames,
					Interfaces:        interfaces,
				},
			}).Execute()
		Expect(err).Should(Succeed())
		resourceId := *createResp.Metadata.ResourceId

		By("InstanceServiceCreate2")
		createResp, _, err = api.InstanceServiceCreate(ctx, cloudAccountId1).Body(
			openapi.InstanceServiceCreateRequest{
				Metadata: &openapi.InstanceServiceCreateRequestMetadata{},
				Spec: &openapi.ProtoInstanceSpec{
					AvailabilityZone:  &availabilityZone,
					InstanceType:      &instanceType,
					MachineImage:      &machineImage,
					RunStrategy:       runStrategy1,
					SshPublicKeyNames: sshPublicKeyNames,
					Interfaces:        interfaces,
				},
			}).Execute()
		Expect(err).Should(Succeed())
		nonDeletedResourceIds := []string{*createResp.Metadata.ResourceId}

		By("InstanceServiceCreate3")
		createResp, _, err = api.InstanceServiceCreate(ctx, cloudAccountId1).Body(
			openapi.InstanceServiceCreateRequest{
				Metadata: &openapi.InstanceServiceCreateRequestMetadata{},
				Spec: &openapi.ProtoInstanceSpec{
					AvailabilityZone:  &availabilityZone,
					InstanceType:      &instanceType,
					MachineImage:      &machineImage,
					RunStrategy:       runStrategy1,
					SshPublicKeyNames: sshPublicKeyNames,
					Interfaces:        interfaces,
				},
			}).Execute()
		Expect(err).Should(Succeed())
		nonDeletedResourceIds = append(nonDeletedResourceIds, *createResp.Metadata.ResourceId)

		By("InstanceServiceDelete")
		_, _, err = api.InstanceServiceDelete(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())

		By("RemoveFinalizer for an Instance")
		_, err = grpcClient.RemoveFinalizer(ctx, &pb.InstanceRemoveFinalizerRequest{
			Metadata: &pb.InstanceIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     resourceId,
			},
		})
		Expect(err).Should(Succeed())
		now := time.Now()
		currentTime := now.Add(5 * time.Minute)
		interval := 1 * time.Minute

		rowsAffected, err := grpcService.DeleteInstanceRecords(ctx, currentTime, interval)
		Expect(err).Should(Succeed())
		Expect(rowsAffected).Should(Equal(int64(1)))

		By("InstanceServiceSearch should return instance after delete is requested but before finalizer is removed")
		searchResp, _, err := api.InstanceServiceSearch(ctx, cloudAccountId1).Execute()
		Expect(err).Should(Succeed())
		searchRespResourceIds := []string{}
		for _, item := range searchResp.Items {
			searchRespResourceIds = append(searchRespResourceIds, *item.Metadata.ResourceId)
		}
		Expect(searchRespResourceIds).Should(ConsistOf(nonDeletedResourceIds))

	})

	It("Instance Delete many records", Serial, Pending, func() {
		api := openApiClient.InstanceServiceApi
		grpcClient := getInstanceGrpcClient()
		cloudAccountId1 := cloudaccount.MustNewId()
		sshPublicKeyName1 := "name1-" + uuid.NewString()
		sshPublicKeyNames := []string{sshPublicKeyName1}
		pubKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		availabilityZone := "us-dev-1a"
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)

		runStrategy1Str := "RerunOnFailure"
		runStrategy1, err := openapi.NewProtoRunStrategyFromValue(runStrategy1Str)
		Expect(err).Should(Succeed())
		interfaces := []openapi.ProtoNetworkInterface{{VNet: &vNet}}

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)

		countInstances := 3600
		for i := 0; i < countInstances; i++ {
			By("InstanceServiceCreate")
			createResp, _, err := api.InstanceServiceCreate(ctx, cloudAccountId1).Body(
				openapi.InstanceServiceCreateRequest{
					Metadata: &openapi.InstanceServiceCreateRequestMetadata{},
					Spec: &openapi.ProtoInstanceSpec{
						AvailabilityZone:  &availabilityZone,
						InstanceType:      &instanceType,
						MachineImage:      &machineImage,
						RunStrategy:       runStrategy1,
						SshPublicKeyNames: sshPublicKeyNames,
						Interfaces:        interfaces,
					},
				}).Execute()
			Expect(err).Should(Succeed())
			resourceId := *createResp.Metadata.ResourceId

			By("InstanceServiceDelete")
			_, _, err = api.InstanceServiceDelete(ctx, cloudAccountId1, resourceId).Execute()
			Expect(err).Should(Succeed())

			By("RemoveFinalizer for an Instance")
			_, err = grpcClient.RemoveFinalizer(ctx, &pb.InstanceRemoveFinalizerRequest{
				Metadata: &pb.InstanceIdReference{
					CloudAccountId: cloudAccountId1,
					ResourceId:     resourceId,
				},
			})
			Expect(err).Should(Succeed())
		}

		now := time.Now()
		currentTime := now.Add(5 * time.Minute)
		interval := 1 * time.Minute

		rowsAffected, err := grpcService.DeleteInstanceRecords(ctx, currentTime, interval)
		Expect(err).Should(Succeed())
		Expect(rowsAffected).Should(Equal(int64(countInstances)))

		By("InstanceServiceSearch should return no instances after all have been purged")
		searchResp, _, err := api.InstanceServiceSearch(ctx, cloudAccountId1).Execute()
		Expect(err).Should(Succeed())
		Expect(len(searchResp.Items)).Should(Equal(0))

	})
})
