// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func createSshPublicKey(cloudAccountId string, name string, sshPublicKey string) {
	_, _, err := openApiClient.SshPublicKeyServiceApi.SshPublicKeyServiceCreate(context.Background(), cloudAccountId).Body(
		openapi.SshPublicKeyServiceCreateRequest{
			Metadata: &openapi.SshPublicKeyServiceCreateRequestMetadata{
				Name: &name,
			},
			Spec: &openapi.ProtoSshPublicKeySpec{
				SshPublicKey: &sshPublicKey,
			},
		}).Execute()
	Expect(err).Should(Succeed())
}

func CreateInstanceType(ctx context.Context, name string) string {
	By("Creating instance type client")
	var instanceTypeName string
	computeApiServerAddress := fmt.Sprintf("localhost:%d", grpcListenPort)
	clientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).Should(Succeed())
	instanceTypeClient := pb.NewInstanceTypeServiceClient(clientConn)

	By("Creating an InstanceType")
	if name == "" {
		instanceTypeName = uuid.NewString()
	} else {
		instanceTypeName = name

	}
	instanceType := &pb.InstanceType{
		Metadata: &pb.InstanceType_Metadata{
			Name: instanceTypeName,
		},
		Spec: &pb.InstanceTypeSpec{
			Name:             instanceTypeName,
			InstanceCategory: pb.InstanceCategory_VirtualMachine,
			Cpu: &pb.CpuSpec{
				Cores:     8,
				Sockets:   1,
				Threads:   1,
				Id:        "0x806F2",
				ModelName: "Intel® Xeon 4th Gen® Scalable processor formerly known as Sapphire Rapids",
			},
			Description: "Intel® Xeon 4th Gen® Scalable processor formerly known as Sapphire Rapids",
			Disks: []*pb.DiskSpec{
				{Size: "20Gi"},
			},
			DisplayName: "Small VM",
			Memory: &pb.MemorySpec{
				DimmCount: 1,
				Speed:     3200,
				DimmSize:  "16Gi",
				Size:      "16Gi",
			},
		},
	}
	_, err = instanceTypeClient.Put(ctx, instanceType)
	Expect(err).Should(Succeed())
	return instanceTypeName
}

func CreateVmMachineImage(ctx context.Context) string {
	// virtualMachine image name must be less than 32 characters
	vmMachineImageName := uuid.NewString()[:30]
	_, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{}, vmMachineImageName, false)
	Expect(err).Should(Succeed())
	return vmMachineImageName
}

func NewPutVNetRequest(name string) *openapi.VNetServicePutRequest {
	region := "us-dev-1"
	availabilityZone := "us-dev-1a"
	prefixLength := int32(24)
	return &openapi.VNetServicePutRequest{
		Metadata: &openapi.VNetServicePutRequestMetadata{
			Name: &name,
		},
		Spec: &openapi.ProtoVNetSpec{
			Region:           &region,
			AvailabilityZone: &availabilityZone,
			PrefixLength:     &prefixLength,
		},
	}
}

func CreateVNet(ctx context.Context, cloudAccountId string) string {
	By("Creating VNet")
	vNetName := uuid.NewString()
	createVNetReq := NewPutVNetRequest(vNetName)
	_, _, err := openApiClient.VNetServiceApi.VNetServicePut(ctx, cloudAccountId).Body(*createVNetReq).Execute()
	Expect(err).Should(Succeed())
	return vNetName
}

var _ = Describe("GRPC-REST server with OpenAPI client", func() {
	const (
		pubKey1          = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
		pubKey2          = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDCgjoBPq3LUKCo2M/FZdT8N86GjVDVoWeTKz/tc87vV8l+9r1vIYDp2VKSfbnEYOMEPEQtbTmeMRmaRHgBK36eKhJbs3gmX749e2izIyp4OX6fxsoD0zfMJpw/R/ZBk500wnCveV49hV9SV1RugKltxA2BxjVrOX1oTBdpMoudwTcYfstNOFnBD7jjt8FqknvD8q2lxo7T4pdrFeFOJ7YQLwcDRSAdONoWqNvSiwmUQMuOS/tR6PIz9NlwNdM7EglPczNYS+Zt7oeml9s3dAbj8gaNCBZi+gXUbTbjbw0y9y8feyX+w+u0hRxPmMigsRYIV6kYaIusRtXxIdu27iV9AYQ9jiG76A2xJhCf6VAGITYU1x1lJasodzUS8C7R53YqQr0P9qSL+QHnjdJ7yNEZQDNpIVNAF1hOoLc+OE7YYrsdz/dHpoRirJePuzNciZxbJX02tNtpwkZZEoGGyJByCksFf0RVPAqz8+qX1yzuykFfllpkCiyHsRktiEuNVrc3sw829NDuaog7K7yiXU4rOM2yuoh9mzExIyNpxB6MlgN3Ms1L9nsdBZ9rZvZIy6ptXeCjyvx2NbGdoWxvujtKNE/zzIFfBFk+AMZjWI+ZLid/En3JtfCcJ4YfexbQKgp1G0hdwU2m9IN0p5a9o78uIr9AIes6VT2D8Zd5UdB7/Q== user2@example.org"
		pubKey3          = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDHXHvjaIYztky6u3J1vfe8SvSvHDJhCSz8PbeSBqBpbVgR94xBpteKWuBi4z35MRp5sT4UTAXhtN+Tz+lag5vupHRPeBQuidCFrEMotaL7bmjiSpOj3bREJsyO7pRgaNeApH/Y9FL/Sh3EBZZMzdm2igTWM9vYlfYGXCBEeO3MiygnTq3DmqADbZjCLk3UWguM7kiubqzOIAdNZ7HzU4sBtljiXVigz8rGb8mKM1449wBoUqG3vs+I/CcQMd0GnZqVzrxZRWVhnpX7vCI78C77v4Fvhu1Z+R795T6TNRV2zEFJPmzs42scmSU6qgmQ8V5lxrP0Z5AVhS3WK6oi5Tnip+7BbbZDBCYDQAHZ68YszYjktV4jCac6rLS6DfF/LeopyvNWtzNbqG7CxHQx9RCeW4A1J3pdeCE/w3eCXf/iWtPSH3yUjUYEh4NCK+QyVuJA7CXF2bE0DIDPlLcMsTSCxGTlf29x/30wIgBAD5/lvlSDiaCejwXmBtP3UHvuXWQMcNvJ7JOJdowXWQF+W5Pm+8GMnSEePv4DqfyNrmBAShBjKfqUVcl9UZ6Kab8Lp54W/Z1rmj+CK+tiEKEmB8OIGtRCjCBxpGWtR5BrjQCRE1ye56406E94yyJvI21XYPnSjnmgl/p4y9GBRCO9sD3pojgeALcmluTC7N58rNXOLw== user3@example.org"
		runStrategy1Str  = "RerunOnFailure"
		runStrategy2Str  = "Halted"
		runStrategy3Str  = "Always"
		availabilityZone = "us-dev-1a"
		address1         = "2.2.2.2"
		statusPhase      = "Ready"
		statusMessage    = "instance is ready"
		proxyAddress     = "1.2.3.4"
		proxyUser        = "guest1"
		proxyPort        = int32(22)
	)

	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	var api *openapi.InstanceServiceApiService

	getInterfaces := func(vNet string) []openapi.ProtoNetworkInterface {
		return []openapi.ProtoNetworkInterface{{VNet: &vNet}}
	}

	getRunStrategy := func(runStrategy string) (*openapi.ProtoRunStrategy, error) {
		return openapi.NewProtoRunStrategyFromValue(runStrategy)
	}

	getPrivateInterfaces := func(vNet string) []*pb.NetworkInterfacePrivate {
		return []*pb.NetworkInterfacePrivate{{VNet: vNet}}
	}

	createPrivateInstance := func(cloudAccountId string, runStrategy string, sshPublicKeyNames []string, instanceType string, machineImage string, vNet string,
		labels map[string]string, instanceGroup string, skipQuotaCheck bool) (*pb.InstancePrivate, error) {

		grpcClient := getInstanceGrpcClient()
		var createResp *pb.InstancePrivate
		availabilityZone := availabilityZone
		interfaces := getPrivateInterfaces(vNet)

		createResp, err := grpcClient.CreatePrivate(ctx, &pb.InstanceCreatePrivateRequest{
			Metadata: &pb.InstanceMetadataCreatePrivate{
				CloudAccountId: cloudAccountId,
				Labels:         labels,
				SkipQuotaCheck: skipQuotaCheck,
			},
			Spec: &pb.InstanceSpecPrivate{
				AvailabilityZone:  availabilityZone,
				InstanceType:      instanceType,
				MachineImage:      machineImage,
				SshPublicKeyNames: sshPublicKeyNames,
				Interfaces:        interfaces,
				ServiceType:       pb.InstanceServiceType_KubernetesAsAService,
			},
		})
		return createResp, err

	}

	createMultiplePrivateInstances := func(cloudAccountId string, runStrategy string, sshPublicKeyNames []string, instanceType string, machineImage string, vNet string,
		labels map[string]string, instanceGroup string, instanceCount int, skipQuotaCheck bool) (*pb.InstanceCreateMultiplePrivateResponse, error) {

		grpcClient := getInstanceGrpcClient()
		availabilityZone := availabilityZone
		interfaces := getPrivateInterfaces(vNet)
		var requestList []*pb.InstanceCreatePrivateRequest

		for i := 0; i < instanceCount; i++ {
			resourceId, _ := uuid.NewRandom()
			privateRequest := &pb.InstanceCreatePrivateRequest{
				Metadata: &pb.InstanceMetadataCreatePrivate{
					CloudAccountId: cloudAccountId,
					ResourceId:     resourceId.String(),
					Name:           instanceGroup + "-" + strconv.Itoa(i),
					Labels:         labels,
					SkipQuotaCheck: skipQuotaCheck,
				},
				Spec: &pb.InstanceSpecPrivate{
					AvailabilityZone:  availabilityZone,
					InstanceType:      instanceType,
					MachineImage:      machineImage,
					SshPublicKeyNames: sshPublicKeyNames,
					Interfaces:        interfaces,
					ServiceType:       pb.InstanceServiceType_KubernetesAsAService,
					InstanceGroup:     instanceGroup,
					ClusterId:         "cluster1",
					NodeId:            "node1",
				},
			}

			requestList = append(requestList, privateRequest)
		}

		multiplePrivateRequest := &pb.InstanceCreateMultiplePrivateRequest{
			Instances: requestList,
		}

		return grpcClient.CreateMultiplePrivate(ctx, multiplePrivateRequest)
	}

	createInstance := func(cloudAccountId string, runStrategy string, sshPublicKeyNames []string, instanceType string, machineImage string, vNet string, labels map[string]string, instanceGroup string, userData string) *openapi.ProtoInstance {
		runStrategy1, err := getRunStrategy(runStrategy)
		Expect(err).Should(Succeed())

		var createResp *openapi.ProtoInstance
		availabilityZone := availabilityZone
		interfaces := getInterfaces(vNet)

		createResp, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(
			openapi.InstanceServiceCreateRequest{
				Metadata: &openapi.InstanceServiceCreateRequestMetadata{
					Labels: &labels,
				},
				Spec: &openapi.ProtoInstanceSpec{
					AvailabilityZone:  &availabilityZone,
					InstanceType:      &instanceType,
					MachineImage:      &machineImage,
					RunStrategy:       runStrategy1,
					SshPublicKeyNames: sshPublicKeyNames,
					Interfaces:        interfaces,
					InstanceGroup:     &instanceGroup,
					UserData:          &userData,
				},
			}).Execute()
		Expect(err).Should(Succeed())
		return createResp

	}

	baseline := func() (cloudAccountId string, req openapi.InstanceServiceCreateRequest) {
		cloudAccountId = cloudaccount.MustNewId()
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		availabilityZone := availabilityZone
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId)

		runStrategy1, err := getRunStrategy(runStrategy1Str)
		Expect(err).Should(Succeed())
		interfaces := getInterfaces(vNet)

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId, sshPublicKeyName1, pubKey1)

		By("InstanceServiceCreate")
		req = openapi.InstanceServiceCreateRequest{
			Metadata: &openapi.InstanceServiceCreateRequestMetadata{},
			Spec: &openapi.ProtoInstanceSpec{
				AvailabilityZone:  &availabilityZone,
				InstanceType:      &instanceType,
				MachineImage:      &machineImage,
				RunStrategy:       runStrategy1,
				SshPublicKeyNames: sshPublicKeyNames1,
				Interfaces:        interfaces,
			},
		}
		return
	}

	BeforeEach(func() {
		clearDatabase(ctx)
		api = openApiClient.InstanceServiceApi
	})

	It("Create, get, update, delete without name should succeed", func() {
		grpcClient := getInstanceGrpcClient()
		addresses1 := []string{address1}
		labels := make(map[string]string)
		cloudAccountId1 := cloudaccount.MustNewId()
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		sshPublicKeyNames1and2 := []string{sshPublicKeyName1, sshPublicKeyName2}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		userData := "hostname: localhost"
		vNet := CreateVNet(ctx, cloudAccountId1)
		interfaces := getInterfaces(vNet)
		statusPhase1, err := openapi.NewProtoInstancePhaseFromValue(statusPhase)
		Expect(err).Should(Succeed())

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

		By("InstanceServiceCreate")
		createResp := createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, "", userData)
		Expect(*createResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		Expect(*createResp.Metadata.Name).Should(Equal(*createResp.Metadata.ResourceId))
		Expect(*createResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
		Expect(*createResp.Spec.InstanceType).Should(Equal(instanceType))
		Expect(*createResp.Spec.MachineImage).Should(Equal(machineImage))
		Expect(string(*createResp.Spec.RunStrategy)).Should(Equal(runStrategy1Str))
		Expect(createResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
		Expect(createResp.Spec.Interfaces[0].VNet).Should(Equal(interfaces[0].VNet))
		resourceId := *createResp.Metadata.ResourceId

		By("InstanceServiceGet")
		getResp, _, err := api.InstanceServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*getResp.Metadata.ResourceId).Should(Equal(resourceId))
		Expect(*getResp.Metadata.Name).Should(Equal(resourceId))
		Expect(*getResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
		Expect(*getResp.Spec.InstanceType).Should(Equal(instanceType))
		Expect(*getResp.Spec.MachineImage).Should(Equal(machineImage))
		Expect(string(*getResp.Spec.RunStrategy)).Should(Equal(runStrategy1Str))
		Expect(getResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
		Expect(getResp.Spec.Interfaces[0].VNet).Should(Equal(interfaces[0].VNet))

		By("UpdateStatus")
		_, err = grpcClient.UpdateStatus(ctx, &pb.InstanceUpdateStatusRequest{
			Metadata: &pb.InstanceIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     resourceId,
			},
			Status: &pb.InstanceStatusPrivate{
				Phase:   pb.InstancePhase(pb.InstancePhase_value[statusPhase]),
				Message: statusMessage,
				Interfaces: []*pb.InstanceInterfaceStatusPrivate{
					{
						VNet:      vNet,
						Addresses: addresses1,
					},
				},
				SshProxy: &pb.SshProxyTunnelStatus{
					ProxyUser:    proxyUser,
					ProxyAddress: proxyAddress,
					ProxyPort:    proxyPort,
				},
			},
		})
		Expect(err).Should(Succeed())

		By("InstanceServiceGet should return updated Status")
		getResp, _, err = api.InstanceServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*getResp.Metadata.ResourceId).Should(Equal(resourceId))
		Expect(*getResp.Metadata.Name).Should(Equal(resourceId))
		Expect(*getResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
		Expect(*getResp.Spec.InstanceType).Should(Equal(instanceType))
		Expect(*getResp.Spec.MachineImage).Should(Equal(machineImage))
		Expect(string(*getResp.Spec.RunStrategy)).Should(Equal(runStrategy1Str))
		Expect(getResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
		Expect(getResp.Spec.Interfaces[0].VNet).Should(Equal(interfaces[0].VNet))
		Expect(getResp.Status.Phase).Should(Equal(statusPhase1))
		Expect(*getResp.Status.Message).Should(Equal(statusMessage))
		Expect(getResp.Status.Interfaces[0].VNet).Should(Equal(interfaces[0].VNet))
		Expect(getResp.Status.Interfaces[0].Addresses).Should(Equal(addresses1))
		Expect(*getResp.Status.SshProxy.ProxyAddress).Should(Equal(proxyAddress))
		Expect(*getResp.Status.SshProxy.ProxyUser).Should(Equal(proxyUser))
		Expect(*getResp.Status.SshProxy.ProxyPort).Should(Equal(proxyPort))

		By("InstanceServiceUpdate SshPublicKeys")
		runStrategy1, err := getRunStrategy(runStrategy1Str)
		Expect(err).Should(Succeed())

		_, _, err = api.InstanceServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.InstanceServiceUpdateRequest{
				Spec: &openapi.ProtoInstanceSpec{
					RunStrategy:       runStrategy1,
					SshPublicKeyNames: sshPublicKeyNames1and2,
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("InstanceServiceGet should return updated SshPublicKeys")
		getResp, _, err = api.InstanceServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*getResp.Metadata.ResourceId).Should(Equal(resourceId))
		Expect(*getResp.Metadata.Name).Should(Equal(resourceId))
		Expect(*getResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
		Expect(*getResp.Spec.InstanceType).Should(Equal(instanceType))
		Expect(*getResp.Spec.MachineImage).Should(Equal(machineImage))
		Expect(string(*getResp.Spec.RunStrategy)).Should(Equal(runStrategy1Str))
		Expect(getResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1and2))
		Expect(getResp.Spec.Interfaces[0].VNet).Should(Equal(interfaces[0].VNet))
		Expect(getResp.Status.Phase).Should(Equal(statusPhase1))
		Expect(*getResp.Status.Message).Should(Equal(statusMessage))
		Expect(getResp.Status.Interfaces[0].VNet).Should(Equal(interfaces[0].VNet))
		Expect(getResp.Status.Interfaces[0].Addresses).Should(Equal(addresses1))
		Expect(*getResp.Status.SshProxy.ProxyAddress).Should(Equal(proxyAddress))
		Expect(*getResp.Status.SshProxy.ProxyUser).Should(Equal(proxyUser))
		Expect(*getResp.Status.SshProxy.ProxyPort).Should(Equal(proxyPort))

		By("InstanceServiceUpdate RunStrategy")
		runStrategy2, err := getRunStrategy(runStrategy2Str)
		Expect(err).Should(Succeed())
		_, _, err = api.InstanceServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.InstanceServiceUpdateRequest{
				Spec: &openapi.ProtoInstanceSpec{
					RunStrategy:       runStrategy2,
					SshPublicKeyNames: sshPublicKeyNames1and2,
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("InstanceServiceGet should return updated RunStrategy")
		getResp, _, err = api.InstanceServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*getResp.Metadata.ResourceId).Should(Equal(resourceId))
		Expect(*getResp.Metadata.Name).Should(Equal(resourceId))
		Expect(*getResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
		Expect(*getResp.Spec.InstanceType).Should(Equal(instanceType))
		Expect(*getResp.Spec.MachineImage).Should(Equal(machineImage))
		Expect(string(*getResp.Spec.RunStrategy)).Should(Equal(runStrategy2Str))
		Expect(getResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1and2))
		Expect(getResp.Spec.Interfaces[0].VNet).Should(Equal(interfaces[0].VNet))
		Expect(getResp.Status.Phase).Should(Equal(statusPhase1))
		Expect(*getResp.Status.Message).Should(Equal(statusMessage))
		Expect(getResp.Status.Interfaces[0].VNet).Should(Equal(interfaces[0].VNet))
		Expect(getResp.Status.Interfaces[0].Addresses).Should(Equal(addresses1))
		Expect(*getResp.Status.SshProxy.ProxyAddress).Should(Equal(proxyAddress))
		Expect(*getResp.Status.SshProxy.ProxyUser).Should(Equal(proxyUser))
		Expect(*getResp.Status.SshProxy.ProxyPort).Should(Equal(proxyPort))

		By("InstanceServiceDelete")
		_, _, err = api.InstanceServiceDelete(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())

		By("InstanceServiceGet should succeed after delete is requested but before finalizer is removed")
		_, _, err = api.InstanceServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())

		By("RemoveFinalizer simulates Instance Scheduler")
		_, err = grpcClient.RemoveFinalizer(ctx, &pb.InstanceRemoveFinalizerRequest{
			Metadata: &pb.InstanceIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     resourceId,
			},
		})
		Expect(err).Should(Succeed())

		By("InstanceServiceGet should return NotFound after finalizer is removed")
		_, httpResponse, err := api.InstanceServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResponse.StatusCode).Should(Equal(http.StatusNotFound))
	})

	It("InstanceServiceDelete twice with same name should succeed", func() {
		grpcClient := getInstanceGrpcClient()
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyNames := []string{sshPublicKeyName1}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		userData := "hostname: localhost"
		vNet := CreateVNet(ctx, cloudAccountId1)

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)

		for instanceIndex := 0; instanceIndex < 2; instanceIndex++ {
			desc := fmt.Sprintf(" (instance %v)", instanceIndex)
			By("InstanceServiceCreate" + desc)

			createdInstance := createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames, instanceType, machineImage, vNet, labels, "", userData)

			By("InstanceServiceDelete request deletion by name" + desc)
			_, _, err := api.InstanceServiceDelete2(ctx, cloudAccountId1, *createdInstance.Metadata.Name).Execute()
			Expect(err).Should(Succeed())

			By("RemoveFinalizer simulates Instance Scheduler" + desc)
			_, err = grpcClient.RemoveFinalizer(ctx, &pb.InstanceRemoveFinalizerRequest{
				Metadata: &pb.InstanceIdReference{
					CloudAccountId: cloudAccountId1,
					ResourceId:     *createdInstance.Metadata.ResourceId,
				},
			})
			Expect(err).Should(Succeed())
		}
	})

	It("InstanceServiceSearch should succeed", func() {
		grpcClient := getInstanceGrpcClient()
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)
		sshPublicKeyName := "name1-" + uuid.New().String()
		sshPublicKeyNames := []string{sshPublicKeyName}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		userData := "hostname: localhost"
		vNet := CreateVNet(ctx, cloudAccountId1)

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName, pubKey1)

		By("InstanceServiceCreate")
		numRows := 4
		var createdInstances []*openapi.ProtoInstance
		for i := 0; i < numRows; i++ {
			createdInstance := createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames, instanceType, machineImage, vNet, labels, "", userData)
			createdInstances = append(createdInstances, createdInstance)
		}

		By("InstanceServiceSearch")
		searchResp, _, err := api.InstanceServiceSearch(ctx, cloudAccountId1).Execute()
		log.Info("Search", "searchResp", searchResp)
		Expect(err).Should(Succeed())
		Expect(len(searchResp.Items)).Should(Equal(numRows))
		for i := 0; i < numRows; i++ {
			Expect(*searchResp.Items[i].Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		}

		By("InstanceServiceSearch should return OMITTED as a user data")
		searchResp, _, err = api.InstanceServiceSearch(ctx, cloudAccountId1).Execute()
		Expect(err).Should(Succeed())
		Expect(*searchResp.Items[0].Spec.UserData).Should(Equal("OMITTED"))

		By("InstanceServiceDelete should delete first instance")
		numDeletedRows := 1
		deletedResourceId := *createdInstances[0].Metadata.ResourceId
		_, _, err = api.InstanceServiceDelete(ctx, cloudAccountId1, deletedResourceId).Execute()
		Expect(err).Should(Succeed())

		By("InstanceServiceSearch should return instance after delete is requested but before finalizer is removed")
		searchResp, _, err = api.InstanceServiceSearch(ctx, cloudAccountId1).Execute()
		Expect(err).Should(Succeed())
		Expect(len(searchResp.Items)).Should(Equal(numRows))

		By("RemoveFinalizer simulates Instance Scheduler")
		_, err = grpcClient.RemoveFinalizer(ctx, &pb.InstanceRemoveFinalizerRequest{
			Metadata: &pb.InstanceIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     deletedResourceId,
			},
		})
		Expect(err).Should(Succeed())

		By("InstanceServiceSearch should not return deleted instance after finalizer is removed")
		searchResp, _, err = api.InstanceServiceSearch(ctx, cloudAccountId1).Execute()
		Expect(err).Should(Succeed())
		Expect(len(searchResp.Items)).Should(Equal(numRows - numDeletedRows))
	})

	It("InstanceServiceUpdate concurrency control with resource version", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyName3 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		sshPublicKeyNames2 := []string{sshPublicKeyName2}
		sshPublicKeyNames3 := []string{sshPublicKeyName3}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		userData := "hostname: localhost"
		vNet := CreateVNet(ctx, cloudAccountId1)

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName3, pubKey3)

		By("InstanceServiceCreate")
		createResp := createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, "", userData)
		resourceId := *createResp.Metadata.ResourceId

		By("Set instance phase to Ready (simulates VM Instance Operator)")
		grpcClient := getInstanceGrpcClient()
		_, err := grpcClient.UpdateStatus(ctx, &pb.InstanceUpdateStatusRequest{
			Metadata: &pb.InstanceIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     resourceId,
			},
			Status: &pb.InstanceStatusPrivate{
				Phase:   pb.InstancePhase(pb.InstancePhase_value[statusPhase]),
				Message: statusMessage,
			},
		})
		Expect(err).Should(Succeed())

		getResp, _, err := api.InstanceServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		resourceVersion1 := *getResp.Metadata.ResourceVersion

		By("InstanceServiceUpdate with stored resource version should succeed")
		runStrategy2, err := getRunStrategy(runStrategy2Str)
		Expect(err).Should(Succeed())

		_, _, err = api.InstanceServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.InstanceServiceUpdateRequest{
				Metadata: &openapi.InstanceServiceUpdateRequestMetadata{
					ResourceVersion: &resourceVersion1,
				},
				Spec: &openapi.ProtoInstanceSpec{
					RunStrategy:       runStrategy2,
					SshPublicKeyNames: sshPublicKeyNames2,
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("InstanceServiceUpdate with old resource version should fail")
		runStrategy3, err := getRunStrategy(runStrategy3Str)
		Expect(err).Should(Succeed())

		_, _, err = api.InstanceServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.InstanceServiceUpdateRequest{
				Metadata: &openapi.InstanceServiceUpdateRequestMetadata{
					ResourceVersion: &resourceVersion1,
				},
				Spec: &openapi.ProtoInstanceSpec{
					RunStrategy:       runStrategy3,
					SshPublicKeyNames: sshPublicKeyNames3,
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())

		By("InstanceServiceGet should return last updated values")
		getResp, _, err = api.InstanceServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(string(*getResp.Spec.RunStrategy)).Should(Equal(runStrategy2Str))
		Expect(getResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames2))
	})

	It("InstanceServiceSearch should return instances filtered with labels", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		labels1 := map[string]string{"key1": "value1"}
		labels2 := map[string]string{"key1": "value1", "key2": "value2"}
		labels3 := map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}
		invalidLables := map[string]string{"@0919@": "##IDC&rele@$e"}
		labels4 := map[string]string{"key4": "value4"}
		sshPublicKeyName := "name1-" + uuid.New().String()
		sshPublicKeyNames := []string{sshPublicKeyName}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		userData := "hostname: localhost"
		vNet := CreateVNet(ctx, cloudAccountId1)

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName, pubKey1)

		By("InstanceServiceCreate instance with labels")
		createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames, instanceType, machineImage, vNet, labels1, "", userData)
		createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames, instanceType, machineImage, vNet, labels2, "", userData)
		createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames, instanceType, machineImage, vNet, labels3, "", userData)

		By("InstanceServiceSearch with no labels")
		searchResp, _, err := api.InstanceServiceSearch(ctx, cloudAccountId1).Execute()
		log.Info("Search", "searchResp", searchResp)
		Expect(err).Should(Succeed())
		Expect(len(searchResp.Items)).Should(Equal(3))

		By("InstanceServiceSearch2 with 1 label")
		searchResp2, _, err := api.InstanceServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.InstanceServiceSearch2Request{
				Metadata: &openapi.InstanceServiceSearch2RequestMetadata{
					Labels: &labels1,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).Should(Succeed())
		Expect(len(searchResp2.Items)).Should(Equal(3))

		By("InstanceServiceSearch2 with 2 labels")
		searchResp2, _, err = api.InstanceServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.InstanceServiceSearch2Request{
				Metadata: &openapi.InstanceServiceSearch2RequestMetadata{
					Labels: &labels2,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).Should(Succeed())
		Expect(len(searchResp2.Items)).Should(Equal(2))

		By("InstanceServiceSearch2 with 3 labels")
		searchResp2, _, err = api.InstanceServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.InstanceServiceSearch2Request{
				Metadata: &openapi.InstanceServiceSearch2RequestMetadata{
					Labels: &labels3,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).Should(Succeed())
		Expect(len(searchResp2.Items)).Should(Equal(1))

		By("InstanceServiceSearch2 with labels with no existing instance for those labels")
		searchResp2, _, err = api.InstanceServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.InstanceServiceSearch2Request{
				Metadata: &openapi.InstanceServiceSearch2RequestMetadata{
					Labels: &labels4,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).Should(Succeed())
		Expect(len(searchResp2.Items)).Should(Equal(0))

		By("InstanceServiceSearch2 with invalid labels should return error")
		searchResp2, _, err = api.InstanceServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.InstanceServiceSearch2Request{
				Metadata: &openapi.InstanceServiceSearch2RequestMetadata{
					Labels: &invalidLables,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).ShouldNot(Succeed())
	})

	It("InstanceServiceCreate should succeed when instance metadata is empty", func() {
		cloudAccountId, instanceServiceReq := baseline()
		instanceServiceReq.Metadata = &openapi.InstanceServiceCreateRequestMetadata{}
		_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).Should(Succeed())
	})

	It("InstanceServiceCreate should fail and return 400 BadRequest when instance body is empty", func() {
		cloudAccountId, _ := baseline()
		instanceServiceReq := openapi.InstanceServiceCreateRequest{}
		_, httpResp, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("InstanceServiceCreate should fail and return 400 BadRequest when instance spec is empty", func() {
		cloudAccountId, instanceServiceReq := baseline()
		instanceServiceReq.Spec = &openapi.ProtoInstanceSpec{}
		_, httpResp, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("InstanceServiceCreate should fail and return 400 BadRequest when MachineImage is empty", func() {
		cloudAccountId, instanceServiceReq := baseline()
		invalidMachineImage := ""
		instanceServiceReq.Spec.MachineImage = &invalidMachineImage
		_, httpResp, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("InstanceServiceCreate should fail and return 400 BadRequest when MachineImage does not exist", func() {
		cloudAccountId, instanceServiceReq := baseline()
		invalidMachineImage := "InvalidMachineImage"
		instanceServiceReq.Spec.MachineImage = &invalidMachineImage
		_, httpResp, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("InstanceServiceCreate should fail and return 400 BadRequest when InstanceType does not exist", func() {
		cloudAccountId, instanceServiceReq := baseline()
		invalidInstanceType := "InvalidInstanceType"
		instanceServiceReq.Spec.InstanceType = &invalidInstanceType
		_, httpResp, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("InstanceServiceCreate should fail and return 400 BadRequest when InstanceType is empty", func() {
		cloudAccountId, instanceServiceReq := baseline()
		emptyString := ""
		instanceServiceReq.Spec.InstanceType = &emptyString
		_, httpResp, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("InstanceServiceCreate should fail and return 409 Conflict when creating twice with the same name", func() {
		cloudAccountId, instanceServiceReq := baseline()
		instanceName := uuid.NewString()
		instanceServiceReq.Metadata.Name = &instanceName
		_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).Should(Succeed())
		_, httpResp, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusConflict))
	})

	It("InstanceServiceCreate should fail and return 400 BadRequest when creating above the quota limit", func() {
		tnstanceTypes := []string{"vm-spr-med", "bm-icp-gaudi2"}
		for _, instanceTypeStr := range tnstanceTypes {
			cloudAccountId, instanceServiceReq := baseline()
			instanceName := uuid.NewString()
			instanceServiceReq.Metadata.Name = &instanceName
			instanceType := CreateInstanceType(ctx, instanceTypeStr)
			instanceServiceReq.Spec.InstanceType = &instanceType
			_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
			Expect(err).Should(Succeed())
			instanceName2 := uuid.NewString()
			instanceServiceReq.Metadata.Name = &instanceName2
			_, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
			Expect(err).Should(Succeed())

			// The standard user's maximum quota for vm-spr-med and bm-icp-gaudi2 is set at 2 for this test.
			instanceName3 := uuid.NewString()
			instanceServiceReq.Metadata.Name = &instanceName3
			_, httpResp, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
			Expect(err).ShouldNot(Succeed())
			Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
		}
	})

	It("InstanceServiceDelete should fail and return 404 NotFound when trying to delete non existing instance", func() {
		cloudAccountId, _ := baseline()
		resourceId := "7913dfb0-1eba-49fb-9c07-97bceb97b2ff"
		By("InstanceServiceDelete")
		_, httpResp, err := api.InstanceServiceDelete(ctx, cloudAccountId, resourceId).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusNotFound))
	})

	It("InstanceServiceDelete should fail when trying to delete an instance with non empty instanceGroup", func() {
		cloudAccountId := cloudaccount.MustNewId()
		sshPublicKeyName := "name1-" + uuid.New().String()
		sshPublicKeyNames := []string{sshPublicKeyName}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		userData := "hostname: localhost"
		vNet := CreateVNet(ctx, cloudAccountId)
		labels := make(map[string]string)
		instanceGroup := "idc-instance-group-" + uuid.New().String()

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId, sshPublicKeyName, pubKey1)

		By("InstanceServiceCreate instance with instance group")
		createResp := createInstance(cloudAccountId, runStrategy1Str, sshPublicKeyNames, instanceType, machineImage, vNet, labels, instanceGroup, userData)
		Expect(*createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		resourceId := *createResp.Metadata.ResourceId

		By("InstanceServiceDelete")
		_, _, err := api.InstanceServiceDelete(ctx, cloudAccountId, resourceId).Execute()
		Expect(err).ShouldNot(Succeed())
	})
	It("InstanceServiceCreate should fail when instanceName is beginning and ending with -", func() {
		cloudAccountId, instanceServiceReq := baseline()
		instanceName := "-my-instance1"
		instanceServiceReq.Metadata.Name = &instanceName
		_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())

		instanceName = "my-instance-1-"
		instanceServiceReq.Metadata.Name = &instanceName
		_, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("InstanceServiceCreate should fail when instanceName is invalid", func() {
		cloudAccountId, instanceServiceReq := baseline()
		instanceName := "my_instance1"
		instanceServiceReq.Metadata.Name = &instanceName
		_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())

		instanceName = "my$instance!1"
		instanceServiceReq.Metadata.Name = &instanceName
		_, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())

		instanceName = "My_instance_1"
		instanceServiceReq.Metadata.Name = &instanceName
		_, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("InstanceServiceCreate should fail when instanceName exceeeds 63 characters", func() {
		cloudAccountId, instanceServiceReq := baseline()
		instanceName := "instance_name_exceeding_63_characters_in_it_s_name_for_testing__"
		instanceServiceReq.Metadata.Name = &instanceName
		_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("InstanceServiceCreate should fail when label name has an invalid character", func() {
		cloudAccountId, instanceServiceReq := baseline()
		instanceServiceReq.Metadata.Labels = &map[string]string{
			"invalid!!!.label.com/name": "value",
		}
		_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("InstanceServiceCreate should fail when machine image is incompatible with instance category", func() {
		cloudAccountId, instanceServiceReq := baseline()
		machineImage := uuid.NewString()
		_, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_BareMetalHost}, []string{}, machineImage, false)
		Expect(err).Should(Succeed())
		instanceServiceReq.Spec.MachineImage = &machineImage
		_, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("InstanceServiceCreate should fail when machine image is incompatible with instance type", func() {
		cloudAccountId, instanceServiceReq := baseline()
		machineImage := uuid.NewString()[:30]
		_, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{"large"}, machineImage, false)
		Expect(err).Should(Succeed())
		instanceServiceReq.Spec.MachineImage = &machineImage
		_, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("InstanceServiceCreate should succeed when machine image is restricted to a single instance type", func() {
		cloudAccountId, instanceServiceReq := baseline()
		machineImage := uuid.NewString()[:30]
		_, err := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{*instanceServiceReq.Spec.InstanceType}, machineImage, false)
		Expect(err).Should(Succeed())
		instanceServiceReq.Spec.MachineImage = &machineImage
		_, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).Should(Succeed())
	})

	It("Create private instance should succeed", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)
		skipQuotaCheck := false

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

		By("InstanceServiceCreate")
		createResp, err := createPrivateInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, "", skipQuotaCheck)
		Expect(err).Should(Succeed())
		Expect(createResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		Expect(createResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
		Expect(createResp.Spec.InstanceType).Should(Equal(instanceType))
		Expect(createResp.Spec.MachineImage).Should(Equal(machineImage))
		Expect(createResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
	})

	It("Create private instance under quota limit should succeed", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		skipQuotaCheck := false
		// The standard user's maximum quota for vm-spr-sml is set at 5 for this test.
		for i := 1; i <= 3; i++ {
			labels := make(map[string]string)
			sshPublicKeyName1 := "name1-" + uuid.New().String()
			sshPublicKeyName2 := "name2-" + uuid.New().String()
			sshPublicKeyNames1 := []string{sshPublicKeyName1}
			instanceType := CreateInstanceType(ctx, "vm-spr-sml")
			machineImage := CreateVmMachineImage(ctx)
			vNet := CreateVNet(ctx, cloudAccountId1)
			// interfaces := getInterfaces(vNet)

			By("SshPublicKeyServiceCreate")
			createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
			createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

			By("InstancePrivateServiceCreatePrivate")
			createResp, err := createPrivateInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, "", skipQuotaCheck)
			Expect(err).Should(Succeed())
			Expect(createResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
			Expect(createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(createResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
			Expect(createResp.Spec.InstanceType).Should(Equal(instanceType))
			Expect(createResp.Spec.MachineImage).Should(Equal(machineImage))
			Expect(createResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
		}
	})

	It("Create private instance above quota limit should fail", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		skipQuotaCheck := false
		// The standard user's maximum quota for vm-spr-sml is set at 5 for this test.
		maxLimit := 5
		instanceCount := 6
		for i := 1; i <= instanceCount; i++ {
			labels := make(map[string]string)
			sshPublicKeyName1 := "name1-" + uuid.New().String()
			sshPublicKeyName2 := "name2-" + uuid.New().String()
			sshPublicKeyNames1 := []string{sshPublicKeyName1}
			instanceType := CreateInstanceType(ctx, "vm-spr-sml")
			machineImage := CreateVmMachineImage(ctx)
			vNet := CreateVNet(ctx, cloudAccountId1)

			By("SshPublicKeyServiceCreate")
			createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
			createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

			By("InstancePrivateServiceCreate")
			createResp, err := createPrivateInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, "", skipQuotaCheck)
			if i > maxLimit {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).Should(Succeed())
				Expect(createResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
				Expect(createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
				Expect(createResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
				Expect(createResp.Spec.InstanceType).Should(Equal(instanceType))
				Expect(createResp.Spec.MachineImage).Should(Equal(machineImage))
				Expect(createResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
			}
		}
	})

	It("Create private instance should skip quota limit if skipQuotaCheck is true", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		skipQuotaCheck := true
		// The standard user's maximum quota for vm-spr-sml is set at 5 for this test.
		instanceCount := 8
		for i := 1; i <= instanceCount; i++ {
			labels := make(map[string]string)
			sshPublicKeyName1 := "name1-" + uuid.New().String()
			sshPublicKeyName2 := "name2-" + uuid.New().String()
			sshPublicKeyNames1 := []string{sshPublicKeyName1}
			instanceType := CreateInstanceType(ctx, "vm-spr-sml")
			machineImage := CreateVmMachineImage(ctx)
			vNet := CreateVNet(ctx, cloudAccountId1)

			By("SshPublicKeyServiceCreate")
			createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
			createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

			By("InstancePrivateServiceCreatePrivate")
			createResp, err := createPrivateInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, "", skipQuotaCheck)
			Expect(err).Should(Succeed())
			Expect(createResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
			Expect(createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(createResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
			Expect(createResp.Spec.InstanceType).Should(Equal(instanceType))
			Expect(createResp.Spec.MachineImage).Should(Equal(machineImage))
			Expect(createResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
		}
	})

	It("Create MultiplePrivate instances under quota limit should succeed", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)
		instanceGroup := "testgroup"

		// The standard user's maximum quota for vm-spr-sml is set at 5 for this test.
		instanceCount := 5
		skipQuotaCheck := false

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

		By("InstancePrivateServiceCreateMultiplePrivate")
		createResp, err := createMultiplePrivateInstances(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, instanceGroup, instanceCount, skipQuotaCheck)
		Expect(err).Should(Succeed())
		Expect(len(createResp.Instances)).Should(Equal(instanceCount))
		for _, instanceResp := range createResp.Instances {
			Expect(instanceResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
			Expect(instanceResp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(instanceResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
			Expect(instanceResp.Spec.InstanceType).Should(Equal(instanceType))
			Expect(instanceResp.Spec.MachineImage).Should(Equal(machineImage))
			Expect(instanceResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
		}
	})

	It("Create MultiplePrivate instances above quota limit should fail", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)
		instanceGroup := "testgroup"

		// The standard user's maximum quota for vm-spr-sml is set at 5 for this test.
		instanceCount := 6
		skipQuotaCheck := false

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

		By("InstancePrivateServiceCreateMultiplePrivate")
		_, err := createMultiplePrivateInstances(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, instanceGroup, instanceCount, skipQuotaCheck)
		Expect(err).ShouldNot(Succeed())
	})

	It("Create MultiplePrivate instances should skip quota limit if skipQuotaCheck is true", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)
		instanceGroup := "testgroup"

		// The standard user's maximum quota for vm-spr-sml is set at 5 for this test.
		instanceCount := 8
		skipQuotaCheck := true

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

		By("InstancePrivateServiceCreateMultiplePrivate")
		createResp, err := createMultiplePrivateInstances(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, instanceGroup, instanceCount, skipQuotaCheck)
		Expect(err).Should(Succeed())
		Expect(len(createResp.Instances)).Should(Equal(instanceCount))
		for _, instanceResp := range createResp.Instances {
			Expect(instanceResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
			Expect(instanceResp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(instanceResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
			Expect(instanceResp.Spec.InstanceType).Should(Equal(instanceType))
			Expect(instanceResp.Spec.MachineImage).Should(Equal(machineImage))
			Expect(instanceResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
		}
	})

	It("Create private instance should have correct instanceGroupSize", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		skipQuotaCheck := true
		labels := make(map[string]string)
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

		By("InstancePrivateServiceCreatePrivate")
		createResp, err := createPrivateInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, "", skipQuotaCheck)
		Expect(err).Should(Succeed())
		Expect(createResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		Expect(createResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
		Expect(createResp.Spec.InstanceType).Should(Equal(instanceType))
		Expect(createResp.Spec.MachineImage).Should(Equal(machineImage))
		Expect(createResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
		Expect(createResp.Spec.InstanceGroupSize).Should(Equal(int32(1)))
	})

	It("Create MultiplePrivate instances should have correct instanceGroupSize", func() {
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)
		instanceGroup := "testgroup"

		// The standard user's maximum quota for vm-spr-sml is set at 5 for this test.
		instanceCount := 5
		skipQuotaCheck := false

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

		By("InstancePrivateServiceCreateMultiplePrivate")
		createResp, err := createMultiplePrivateInstances(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet, labels, instanceGroup, instanceCount, skipQuotaCheck)
		Expect(err).Should(Succeed())
		Expect(len(createResp.Instances)).Should(Equal(instanceCount))
		for _, instanceResp := range createResp.Instances {
			Expect(instanceResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
			Expect(instanceResp.Metadata.ResourceId).ShouldNot(BeEmpty())
			Expect(instanceResp.Spec.AvailabilityZone).Should(Equal(availabilityZone))
			Expect(instanceResp.Spec.InstanceType).Should(Equal(instanceType))
			Expect(instanceResp.Spec.MachineImage).Should(Equal(machineImage))
			Expect(instanceResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames1))
			Expect(instanceResp.Spec.InstanceGroupSize).Should(Equal(int32(instanceCount)))
		}
	})
})
