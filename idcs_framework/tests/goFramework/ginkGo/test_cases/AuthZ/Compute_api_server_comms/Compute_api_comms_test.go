package compute_api_comms_test

import (
	"context"
	"fmt"
	"goFramework/ginkGo/test_cases/testutils"
	"net/http"

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
	computeApiServerAddress := testutils.CheckEnvironmentAndGetLocalHost(grpcListenPort)
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

func CreateMachineImage(ctx context.Context, instanceCategories []pb.InstanceCategory, instanceTypes []string) string {
	By("Creating machine image client")
	computeApiServerAddress := testutils.CheckEnvironmentAndGetLocalHost(grpcListenPort)
	clientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).Should(Succeed())
	machineImageClient := pb.NewMachineImageServiceClient(clientConn)

	By("Creating MachineImage")
	machineImageName := uuid.NewString()
	machineImage := &pb.MachineImage{
		Metadata: &pb.MachineImage_Metadata{
			Name: machineImageName,
		},
		Spec: &pb.MachineImageSpec{
			DisplayName:        "Ubuntu 22.04 LTS (Jammy Jellyfish) v20230128",
			Description:        "Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.",
			Icon:               "https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png",
			InstanceCategories: instanceCategories,
			InstanceTypes:      instanceTypes,
			Md5Sum:             "764efa883dda1e11db47671c4a3bbd9e",
			Sha256Sum:          "98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4",
			Labels: map[string]string{
				"architecture": "X86_64",
				"family":       "ubuntu-2204-lts",
			},
			ImageCategories: []string{
				"AI",
			},
			Components: []*pb.MachineImageComponent{
				{
					Name:        "Ubuntu 22.04 LTS",
					Type:        "OS",
					Version:     "22.04",
					Description: "Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.",
					InfoUrl:     "https://discourse.ubuntu.com/t/jammy-jellyfish-release-notes/24668?_ga=2.61253994.716223186.1673128475-452578204.1673128475",
					ImageUrl:    "https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png",
				},
			},
		},
	}
	_, err = machineImageClient.Put(ctx, machineImage)
	Expect(err).Should(Succeed())
	return machineImageName
}

func CreateVmMachineImage(ctx context.Context) string {
	return CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{})
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

	createInstance := func(cloudAccountId string, runStrategy string, sshPublicKeyNames []string, instanceType string, machineImage string, vNet string) *openapi.ProtoInstance {
		runStrategy1, err := getRunStrategy(runStrategy)
		Expect(err).Should(Succeed())

		var createResp *openapi.ProtoInstance
		availabilityZone := availabilityZone
		interfaces := getInterfaces(vNet)

		createResp, _, err = api.InstanceServiceCreate(ctx, cloudAccountId).Body(
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
		cloudAccountId1 := cloudaccount.MustNewId()
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		sshPublicKeyNames1and2 := []string{sshPublicKeyName1, sshPublicKeyName2}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)
		interfaces := getInterfaces(vNet)
		statusPhase1, err := openapi.NewProtoInstancePhaseFromValue(statusPhase)
		Expect(err).Should(Succeed())

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)

		By("InstanceServiceCreate")
		createResp := createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet)
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
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyNames := []string{sshPublicKeyName1}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)

		for instanceIndex := 0; instanceIndex < 2; instanceIndex++ {
			desc := fmt.Sprintf(" (instance %v)", instanceIndex)
			By("InstanceServiceCreate" + desc)

			createdInstance := createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames, instanceType, machineImage, vNet)

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
		sshPublicKeyName := "name1-" + uuid.New().String()
		sshPublicKeyNames := []string{sshPublicKeyName}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName, pubKey1)

		By("InstanceServiceCreate")
		numRows := 4
		var createdInstances []*openapi.ProtoInstance
		for i := 0; i < numRows; i++ {
			createdInstance := createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames, instanceType, machineImage, vNet)
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
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyName2 := "name2-" + uuid.New().String()
		sshPublicKeyName3 := "name2-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}
		sshPublicKeyNames2 := []string{sshPublicKeyName2}
		sshPublicKeyNames3 := []string{sshPublicKeyName3}
		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId1)

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId1, sshPublicKeyName1, pubKey1)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName2, pubKey2)
		createSshPublicKey(cloudAccountId1, sshPublicKeyName3, pubKey3)

		By("InstanceServiceCreate")
		createResp := createInstance(cloudAccountId1, runStrategy1Str, sshPublicKeyNames1, instanceType, machineImage, vNet)
		resourceId := *createResp.Metadata.ResourceId
		resourceVersion1 := *createResp.Metadata.ResourceVersion

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
		getResp, _, err := api.InstanceServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(string(*getResp.Spec.RunStrategy)).Should(Equal(runStrategy2Str))
		Expect(getResp.Spec.SshPublicKeyNames).Should(Equal(sshPublicKeyNames2))
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

	It("InstanceServiceDelete should fail and return 404 NotFound when trying to delete non existing instance", func() {
		cloudAccountId, _ := baseline()
		resourceId := "7913dfb0-1eba-49fb-9c07-97bceb97b2ff"
		By("InstanceServiceDelete")
		_, httpResp, err := api.InstanceServiceDelete(ctx, cloudAccountId, resourceId).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusNotFound))
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
		machineImage := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_BareMetalHost}, []string{})
		instanceServiceReq.Spec.MachineImage = &machineImage
		_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("InstanceServiceCreate should fail when machine image is incompatible with instance type", func() {
		cloudAccountId, instanceServiceReq := baseline()
		machineImage := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{"large"})
		instanceServiceReq.Spec.MachineImage = &machineImage
		_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("InstanceServiceCreate should succeed when machine image is restricted to a single instance type", func() {
		cloudAccountId, instanceServiceReq := baseline()
		machineImage := CreateMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_VirtualMachine}, []string{*instanceServiceReq.Spec.InstanceType})
		instanceServiceReq.Spec.MachineImage = &machineImage
		_, _, err := api.InstanceServiceCreate(ctx, cloudAccountId).Body(instanceServiceReq).Execute()
		Expect(err).Should(Succeed())
	})
})
