// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delete Deactivated Instances records ", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	BeforeEach(func() {
		clearDatabase(ctx)
	})

	It("Delete Deactivated Instances records", Serial, func() {
		api := openApiClient.InstanceServiceApi
		sshPublicKeyName := "name1-" + uuid.NewString()
		sshPublicKeyNames := []string{sshPublicKeyName}
		pubKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		availabilityZone := "us-dev-1a"
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId)
		instanceType1 := CreateInstanceType(ctx, instanceTypeName1)
		instanceType2 := CreateInstanceType(ctx, instanceTypeName2)

		runStrategy1Str := "RerunOnFailure"
		runStrategy1, err := openapi.NewProtoRunStrategyFromValue(runStrategy1Str)
		Expect(err).Should(Succeed())
		interfaces1 := []openapi.ProtoNetworkInterface{{VNet: &vNet}}

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId, sshPublicKeyName, pubKey)

		By("InstanceServiceCreate1")
		_, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(
			openapi.InstanceServiceCreateRequest{
				Metadata: &openapi.InstanceServiceCreateRequestMetadata{},
				Spec: &openapi.ProtoInstanceSpec{
					AvailabilityZone:  &availabilityZone,
					InstanceType:      &instanceType1,
					MachineImage:      &machineImage,
					RunStrategy:       runStrategy1,
					SshPublicKeyNames: sshPublicKeyNames,
					Interfaces:        interfaces1,
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("InstanceServiceCreate2")
		_, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(
			openapi.InstanceServiceCreateRequest{
				Metadata: &openapi.InstanceServiceCreateRequestMetadata{},
				Spec: &openapi.ProtoInstanceSpec{
					AvailabilityZone:  &availabilityZone,
					InstanceType:      &instanceType2,
					MachineImage:      &machineImage,
					RunStrategy:       runStrategy1,
					SshPublicKeyNames: sshPublicKeyNames,
					Interfaces:        interfaces1,
				},
			}).Execute()
		Expect(err).Should(Succeed())

		numOfDeactivatedInstancesDeleted, err := grpcService.DeletedDeactivatedInstances(ctx)
		Expect(err).Should(Succeed())
		Expect(numOfDeactivatedInstancesDeleted).Should(Equal(int64(2)))
	})
})
