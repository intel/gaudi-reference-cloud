// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instance watch", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	It("Instance list and watch", Serial, func() {
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
		runStrategy2Str := "Halted"
		runStrategy2, err := openapi.NewProtoRunStrategyFromValue(runStrategy2Str)
		Expect(err).Should(Succeed())
		statusPhaseReadyStr := "Ready"
		statusMessageReadyStr := "Instance is running and has completed running startup scripts"
		interfaces := []openapi.ProtoNetworkInterface{{VNet: &vNet}}

		By("Instance SearchStreamPrivate should return only a bookmark (no instances)")
		stream, err := grpcClient.SearchStreamPrivate(ctx, &pb.InstanceSearchStreamPrivateRequest{})
		Expect(err).Should(Succeed())
		resp, err := stream.Recv()
		Expect(err).Should(Succeed())
		Expect(resp.Type).Should(Equal(pb.WatchDeltaType_Bookmark))
		resourceVersion := resp.Object.Metadata.ResourceVersion
		Expect(resourceVersion).ShouldNot(BeEmpty())
		_, err = stream.Recv()
		Expect(err).Should(Equal(io.EOF))
		Expect(stream.CloseSend()).Should(Succeed())

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)

		By("InstanceServiceCreate")
		numCreatedInstances := 4
		expectedResourceIds := make(map[string]*openapi.ProtoInstance)
		var createdInstances []*openapi.ProtoInstance
		for i := 0; i < numCreatedInstances; i++ {
			createdInstance, _, err := api.InstanceServiceCreate(ctx, cloudAccountId1).Body(
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
			expectedResourceIds[*createdInstance.Metadata.ResourceId] = createdInstance
			createdInstances = append(createdInstances, createdInstance)
		}
		Expect(len(expectedResourceIds)).Should(Equal(numCreatedInstances))

		By("InstanceServiceDelete record 0")
		deletedResourceId := *createdInstances[0].Metadata.ResourceId
		_, _, err = api.InstanceServiceDelete(ctx, cloudAccountId1, deletedResourceId).Execute()
		Expect(err).Should(Succeed())
		delete(expectedResourceIds, deletedResourceId)

		By("Instance SearchStreamPrivate should return all records except for record 0 which was deleted")
		stream, err = grpcClient.SearchStreamPrivate(ctx, &pb.InstanceSearchStreamPrivateRequest{})
		Expect(err).Should(Succeed())
		resourceVersion = ""
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			Expect(err).Should(Succeed())
			if resp.Type == pb.WatchDeltaType_Bookmark {
				resourceVersion = resp.Object.Metadata.ResourceVersion
			} else {
				Expect(resp.Type).Should(Equal(pb.WatchDeltaType_Updated))
				delete(expectedResourceIds, resp.Object.Metadata.ResourceId)
			}
		}
		Expect(len(expectedResourceIds)).Should(Equal(0))
		Expect(resourceVersion).ShouldNot(BeEmpty())

		By("Starting new watch from the ResourceVersion returned by SearchStreamPrivate")
		stream, err = grpcClient.Watch(ctx, &pb.InstanceWatchRequest{
			ResourceVersion: resourceVersion,
		})
		Expect(err).Should(Succeed())

		By("InstanceServiceUpdate RunStrategy record 1")
		updatedResourceId := *createdInstances[1].Metadata.ResourceId
		_, _, err = api.InstanceServiceUpdate(ctx, cloudAccountId1, updatedResourceId).Body(
			openapi.InstanceServiceUpdateRequest{
				Spec: &openapi.ProtoInstanceSpec{
					RunStrategy:       runStrategy2,
					SshPublicKeyNames: sshPublicKeyNames,
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())

		By("InstanceUpdateStatusRequest Phase record 0 (instance should be in Ready state to set Halted RunStrategy)")
		_, err = grpcClient.UpdateStatus(ctx, &pb.InstanceUpdateStatusRequest{
			Metadata: &pb.InstanceIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     updatedResourceId,
			},
			Status: &pb.InstanceStatusPrivate{
				Phase:   pb.InstancePhase(pb.InstancePhase_value[statusPhaseReadyStr]),
				Message: statusMessageReadyStr,
			},
		})
		Expect(err).Should(Succeed())

		By("InstanceServiceUpdate RunStrategy record 2")
		updatedResourceId = *createdInstances[1].Metadata.ResourceId
		_, _, err = api.InstanceServiceUpdate(ctx, cloudAccountId1, updatedResourceId).Body(
			openapi.InstanceServiceUpdateRequest{
				Spec: &openapi.ProtoInstanceSpec{
					RunStrategy:       runStrategy2,
					SshPublicKeyNames: sshPublicKeyNames,
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("After updating the instance, watch should return an Updated record (and possibly Bookmark messages)")
		Eventually(func(g Gomega) {
			resp, err = stream.Recv()
			g.Expect(err).Should(Succeed())
			g.Expect(resp.Type).Should(Equal(pb.WatchDeltaType_Updated))
			g.Expect(resp.Object.Metadata.ResourceId).Should(Equal(updatedResourceId))
		}, "1s")

		By("InstanceServiceDelete record 2")
		deletedResourceId = *createdInstances[2].Metadata.ResourceId
		_, _, err = api.InstanceServiceDelete(ctx, cloudAccountId1, deletedResourceId).Execute()
		Expect(err).Should(Succeed())

		By("After requesting deletion of an instance, watch should return an Updated record")
		Eventually(func(g Gomega) {
			resp, err = stream.Recv()
			g.Expect(err).Should(Succeed())
			g.Expect(resp.Type).Should(Equal(pb.WatchDeltaType_Updated))
			g.Expect(resp.Object.Metadata.ResourceId).Should(Equal(deletedResourceId))
		}, "1s")

		By("RemoveFinalizer simulates Instance Scheduler")
		_, err = grpcClient.RemoveFinalizer(ctx, &pb.InstanceRemoveFinalizerRequest{
			Metadata: &pb.InstanceIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     deletedResourceId,
			},
		})
		Expect(err).Should(Succeed())

		By("After finalizer is removed, watch should return a Deleted record")
		Eventually(func(g Gomega) {
			resp, err = stream.Recv()
			g.Expect(err).Should(Succeed())
			g.Expect(resp.Type).Should(Equal(pb.WatchDeltaType_Deleted))
			g.Expect(resp.Object.Metadata.ResourceId).Should(Equal(deletedResourceId))
		}, "1s")
	})
})
