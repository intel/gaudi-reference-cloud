package test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var poolPort = int32(9112)

var _ = Describe("GRPC-REST server with OpenAPI client (LoadBalancer)", func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	var api *openapi.LoadBalancerServiceApiService

	createLoadBalancerSelector := func(cloudAccountId string, port int32, sourceIPs []string, labels map[string]string) (*openapi.ProtoLoadBalancer, error) {

		var createResp *openapi.ProtoLoadBalancer

		createResp, _, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(
			openapi.LoadBalancerServiceCreateRequest{
				Metadata: &openapi.LoadBalancerServiceCreateRequestMetadata{
					Labels: &labels,
				},
				Spec: &openapi.ProtoLoadBalancerSpec{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &port,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &poolPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: sourceIPs,
					},
				},
			}).Execute()
		return createResp, err
	}

	createInstance := func(cloudAccountId string) *openapi.ProtoInstance {
		const (
			pubKey1         = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
			runStrategy1Str = "RerunOnFailure"
			skipQuotaCheck  = false
		)

		availabilityZone := "us-dev-1a"

		getRunStrategy := func(runStrategy string) (*openapi.ProtoRunStrategy, error) {
			return openapi.NewProtoRunStrategyFromValue(runStrategy)
		}

		getInterfaces := func(vNet string) []openapi.ProtoNetworkInterface {
			return []openapi.ProtoNetworkInterface{{VNet: &vNet}}
		}

		var apiInstance *openapi.InstanceServiceApiService
		apiInstance = openApiClient.InstanceServiceApi

		// Create an instance
		sshPublicKeyName1 := "name1-" + uuid.New().String()
		sshPublicKeyNames1 := []string{sshPublicKeyName1}

		By("SshPublicKeyServiceCreate")
		createSshPublicKey(cloudAccountId, sshPublicKeyName1, pubKey1)

		instanceType := CreateInstanceType(ctx, "vm-spr-sml")
		machineImage := CreateVmMachineImage(ctx)
		vNet := CreateVNet(ctx, cloudAccountId)
		labels := map[string]string{"label1": "value1"}
		//availabilityZone := availabilityZone
		instanceGroup := "idc-instance-group-" + uuid.New().String()
		userData := "hostname: localhost"

		var createResp *openapi.ProtoInstance
		interfaces := getInterfaces(vNet)

		runStrategy1, err := getRunStrategy(runStrategy1Str)
		Expect(err).Should(Succeed())

		createResp, _, err = apiInstance.InstanceServiceCreate(ctx, cloudAccountId).Body(
			openapi.InstanceServiceCreateRequest{
				Metadata: &openapi.InstanceServiceCreateRequestMetadata{
					Labels: &labels,
				},
				Spec: &openapi.ProtoInstanceSpec{
					AvailabilityZone:  &availabilityZone,
					InstanceType:      &instanceType,
					MachineImage:      &machineImage,
					RunStrategy:       runStrategy1,
					SshPublicKeyNames: sshPublicKeyNames1,
					Interfaces:        interfaces,
					InstanceGroup:     &instanceGroup,
					UserData:          &userData,
				},
			}).Execute()
		Expect(err).Should(Succeed())

		getInstanceResp, _, err := apiInstance.InstanceServiceGet(ctx, cloudAccountId, *createResp.Metadata.ResourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getInstanceResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId))
		Expect(*getInstanceResp.Metadata.Name).Should(Equal(*createResp.Metadata.ResourceId))

		return createResp
	}

	createLoadBalancerStatic := func(cloudAccountId string, port int32, sourceIPs []string, labels map[string]string) *openapi.ProtoLoadBalancer {

		var createResp *openapi.ProtoLoadBalancer

		createInstanceResp := createInstance(cloudAccountId)

		createResp, _, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(
			openapi.LoadBalancerServiceCreateRequest{
				Metadata: &openapi.LoadBalancerServiceCreateRequestMetadata{
					Labels: &labels,
				},
				Spec: &openapi.ProtoLoadBalancerSpec{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceResourceIds: []string{
								*createInstanceResp.Metadata.ResourceId,
							},
							Port: &poolPort,
						},
						Port: &port,
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: sourceIPs,
					},
				},
			}).Execute()
		Expect(err).Should(Succeed())

		return createResp
	}

	baselineInstanceSelector := func() (cloudAccountId string, req openapi.LoadBalancerServiceCreateRequest) {
		const (
			monitorType = "tcp"
		)

		port := int32(8080)

		cloudAccountId = cloudaccount.MustNewId()

		labels := make(map[string]string)

		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue(monitorType)
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceCreate")
		req = openapi.LoadBalancerServiceCreateRequest{
			Metadata: &openapi.LoadBalancerServiceCreateRequestMetadata{
				Labels: &labels,
			},
			Spec: &openapi.ProtoLoadBalancerSpec{
				Listeners: []openapi.ProtoLoadBalancerListener{{
					Pool: &openapi.ProtoLoadBalancerPool{
						InstanceSelectors: &map[string]string{
							"foo": "bar",
						},
						Monitor: monitorType1,
					},
					Port: &port,
				}},
				Security: &openapi.ProtoLoadBalancerSecurity{
					Sourceips: []string{
						"any",
					},
				},
			},
		}
		return
	}

	BeforeEach(func() {
		clearDatabase(ctx)
		api = openApiClient.LoadBalancerServiceApi
	})

	It("Create, get, update, delete should succeed", func() {
		const (
			monitorType = "tcp"
			port        = int32(8080)
		)
		sourceIPs := []string{"1.2.3.4", "10.252.0.0/27"}

		grpcClient := getLoadBalancerGrpcClient()

		labels := make(map[string]string)
		labels["foo"] = "bar"

		cloudAccountId1 := cloudaccount.MustNewId()

		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue(monitorType)
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceCreate - Single Listener - Static")
		createResp := createLoadBalancerStatic(cloudAccountId1, port, sourceIPs, labels)
		Expect(*createResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		Expect(*createResp.Metadata.Name).Should(Equal(*createResp.Metadata.ResourceId))
		Expect(*createResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(BeEmpty())
		Expect(len(createResp.Spec.Listeners[0].Pool.InstanceResourceIds)).Should(Equal(1))
		Expect(*createResp.Spec.Listeners[0].Pool.Port).Should(Equal(poolPort))
		Expect(createResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*createResp.Spec.Listeners[0].Port).Should(Equal(port))
		Expect(createResp.Spec.Security.Sourceips).Should(Equal(sourceIPs))

		resourceId := *createResp.Metadata.ResourceId
		Expect(resourceId).ShouldNot(BeEmpty())

		By("LoadBalancerServiceCreate - Single Listener - InstanceSelector")
		createResp, _ = createLoadBalancerSelector(cloudAccountId1, port, sourceIPs, labels)
		Expect(*createResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		Expect(*createResp.Metadata.Name).Should(Equal(*createResp.Metadata.ResourceId))
		Expect(*createResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(Equal(map[string]string{"foo": "bar"}))
		Expect(createResp.Spec.Listeners[0].Pool.InstanceResourceIds).Should(BeEmpty())
		Expect(*createResp.Spec.Listeners[0].Pool.Port).Should(Equal(poolPort))
		Expect(createResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*createResp.Spec.Listeners[0].Port).Should(Equal(port))
		Expect(createResp.Spec.Security.Sourceips).Should(Equal(sourceIPs))

		resourceId = *createResp.Metadata.ResourceId
		Expect(resourceId).ShouldNot(BeEmpty())

		By("LoadBalancerServiceGet - Single Listener")
		getResp, _, err := api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*getResp.Metadata.ResourceId).Should(Equal(resourceId))
		Expect(*getResp.Metadata.Name).Should(Equal(resourceId))
		Expect(*getResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(Equal(map[string]string{"foo": "bar"}))
		Expect(*getResp.Spec.Listeners[0].Pool.Port).Should(Equal(poolPort))
		Expect(getResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*getResp.Spec.Listeners[0].Port).Should(Equal(port))
		Expect(getResp.Spec.Security.Sourceips).Should(Equal(sourceIPs))

		By("UpdateStatus")
		statusStateActive := "Pending"
		statusVIP := "1.2.3.4"
		statusMessage := "Load Balancer is pending creation..."

		_, err = grpcClient.UpdateStatus(ctx, &pb.LoadBalancerUpdateStatusRequest{
			Metadata: &pb.LoadBalancerIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     resourceId,
			},
			Status: &pb.LoadBalancerStatusPrivate{
				Conditions: &pb.LoadBalancerConditionsStatus{
					Listeners: []*pb.LoadBalancerConditionsListenerStatus{{
						Port:          port,
						PoolCreated:   false,
						VipCreated:    false,
						VipPoolLinked: false,
					}},
					FirewallRuleCreated: false,
				},
				Listeners: []*pb.LoadBalancerListenerStatus{{
					Name:        "",
					VipID:       0,
					Message:     statusMessage,
					PoolMembers: nil,
					PoolID:      0,
				}},
				State: statusStateActive,
				Vip:   statusVIP,
			},
		})
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceGet should return updated Status")
		getResp, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		Expect(*getResp.Metadata.ResourceId).Should(Equal(resourceId))
		Expect(*getResp.Metadata.Name).Should(Equal(resourceId))
		Expect(string(*getResp.Status.State)).Should(Equal(statusStateActive))
		Expect(*getResp.Status.Vip).Should(Equal(statusVIP))

		By("LoadBalancerServiceUpdate Port")
		newSpecPort := int32(9191)
		_, _, err = api.LoadBalancerServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &newSpecPort,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &poolPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: sourceIPs,
					},
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceGet should return updated Port")
		getResp, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(Equal(map[string]string{"foo": "bar"}))
		Expect(*getResp.Spec.Listeners[0].Pool.Port).Should(Equal(poolPort))
		Expect(getResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*getResp.Spec.Listeners[0].Port).Should(Equal(newSpecPort))

		By("LoadBalancerServiceUpdate Pool.Port")
		newSpecPort = int32(9191)
		newPoolPort := int32(13244)
		_, _, err = api.LoadBalancerServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &newSpecPort,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &newPoolPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: sourceIPs,
					},
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceGet should return updated Pool.Port")
		getResp, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(Equal(map[string]string{"foo": "bar"}))
		Expect(*getResp.Spec.Listeners[0].Pool.Port).Should(Equal(newPoolPort))
		Expect(getResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*getResp.Spec.Listeners[0].Port).Should(Equal(newSpecPort))

		By("LoadBalancerServiceUpdate Pool.InstanceSelector")
		newSpecPort = int32(9191)
		_, _, err = api.LoadBalancerServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &newSpecPort,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"new": "val",
								"var": "two",
							},
							Port: &newSpecPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: sourceIPs,
					},
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceGet should return updated Pool.InstanceSelector")
		getResp, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(Equal(map[string]string{"new": "val", "var": "two"}))
		Expect(getResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*getResp.Spec.Listeners[0].Port).Should(Equal(newSpecPort))

		By("LoadBalancerServiceUpdate Pool.InstanceResourceIds")
		newSpecPort = int32(9191)

		createInstanceResp1 := createInstance(cloudAccountId1)
		createInstanceResp2 := createInstance(cloudAccountId1)

		_, _, err = api.LoadBalancerServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &newSpecPort,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceResourceIds: []string{
								createInstanceResp1.Metadata.GetResourceId(),
								createInstanceResp2.Metadata.GetResourceId(),
							},
							Port: &newSpecPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: sourceIPs,
					},
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceGet should return updated Pool.InstanceResourceIds")
		getResp, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(BeEmpty())
		Expect(len(getResp.Spec.Listeners[0].Pool.InstanceResourceIds)).Should(Equal(2))
		Expect(getResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*getResp.Spec.Listeners[0].Port).Should(Equal(newSpecPort))

		By("LoadBalancerServiceUpdate Monitor")
		newMonitor := "tcp"
		newMonitor1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue(newMonitor)
		Expect(err).Should(Succeed())

		_, _, err = api.LoadBalancerServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &newSpecPort,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Monitor: newMonitor1,
							Port:    &newSpecPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: sourceIPs,
					},
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceGet should return updated Monitor")
		getResp, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(Equal(map[string]string{"foo": "bar"}))
		Expect(getResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*getResp.Spec.Listeners[0].Port).Should(Equal(newSpecPort))

		By("LoadBalancerServiceUpdate SourceIPs")
		_, _, err = api.LoadBalancerServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &newSpecPort,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Monitor: newMonitor1,
							Port:    &poolPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{"1.2.3.4", "9.8.7.6/32"},
					},
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceGet should return updated SourceIPs")
		getResp, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(*getResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(Equal(map[string]string{"foo": "bar"}))

		Expect(getResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*getResp.Spec.Listeners[0].Port).Should(Equal(newSpecPort))
		Expect(getResp.Spec.Security.Sourceips).Should(Equal([]string{"1.2.3.4", "9.8.7.6/32"}))

		By("LoadBalancerDelete")
		_, _, err = api.LoadBalancerServiceDelete(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceGet should succeed after delete is requested but before finalizer is removed")
		_, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())

		By("RemoveFinalizer simulates LoadBalancer operator")
		_, err = grpcClient.RemoveFinalizer(ctx, &pb.LoadBalancerRemoveFinalizerRequest{
			Metadata: &pb.LoadBalancerIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     resourceId,
			},
		})
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceGet should return NotFound after finalizer is removed")
		_, httpResponse, err := api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResponse.StatusCode).Should(Equal(http.StatusNotFound))
	})

	It("LoadBalancerServiceSearch should succeed", func() {
		const (
			monitorType = "tcp"
			port        = int32(8080)
		)

		grpcClient := getLoadBalancerGrpcClient()
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)

		By("LoadBalancerServiceCreate")
		numRows := 4
		var createdLoadBalancers []*openapi.ProtoLoadBalancer
		for i := 0; i < numRows; i++ {
			createdLoadBalancer, err := createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels)
			Expect(err).Should(Succeed())
			createdLoadBalancers = append(createdLoadBalancers, createdLoadBalancer)
		}

		By("LoadBalancerServiceSearch")
		searchResp, _, err := api.LoadBalancerServiceSearch(ctx, cloudAccountId1).Execute()
		log.Info("Search", "searchResp", searchResp)
		Expect(err).Should(Succeed())
		Expect(len(searchResp.Items)).Should(Equal(numRows))
		for i := 0; i < numRows; i++ {
			Expect(*searchResp.Items[i].Metadata.CloudAccountId).Should(Equal(cloudAccountId1))
		}

		By("LoadBalancerServiceDelete should delete first lb")
		numDeletedRows := 1
		deletedResourceId := *createdLoadBalancers[0].Metadata.ResourceId
		_, _, err = api.LoadBalancerServiceDelete(ctx, cloudAccountId1, deletedResourceId).Execute()
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceSearch should return load balancer after delete is requested but before finalizer is removed")
		searchResp, _, err = api.LoadBalancerServiceSearch(ctx, cloudAccountId1).Execute()
		Expect(err).Should(Succeed())
		Expect(len(searchResp.Items)).Should(Equal(numRows))

		By("RemoveFinalizer simulates Instance Scheduler")
		_, err = grpcClient.RemoveFinalizer(ctx, &pb.LoadBalancerRemoveFinalizerRequest{
			Metadata: &pb.LoadBalancerIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     deletedResourceId,
			},
		})
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceSearch should not return deleted instance after finalizer is removed")
		searchResp, _, err = api.LoadBalancerServiceSearch(ctx, cloudAccountId1).Execute()
		Expect(err).Should(Succeed())
		Expect(len(searchResp.Items)).Should(Equal(numRows - numDeletedRows))
	})

	It("LoadBalancerServiceUpdate concurrency control with resource version", func() {
		const (
			monitorType = "tcp"
			port        = int32(8080)
		)

		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)

		By("LoadBalancerServiceCreate")
		createResp, err := createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels)
		Expect(err).Should(Succeed())
		resourceId := *createResp.Metadata.ResourceId
		Expect(resourceId).ShouldNot(BeEmpty())

		By("Set load balancer state to Active (Simulates the LB Operator)")
		grpcClient := getLoadBalancerGrpcClient()
		statusStateActive := "Active"
		statusMessage := "Load Balancer is pending creation..."

		_, err = grpcClient.UpdateStatus(ctx, &pb.LoadBalancerUpdateStatusRequest{
			Metadata: &pb.LoadBalancerIdReference{
				CloudAccountId: cloudAccountId1,
				ResourceId:     resourceId,
			},
			Status: &pb.LoadBalancerStatusPrivate{
				Conditions: &pb.LoadBalancerConditionsStatus{
					Listeners: []*pb.LoadBalancerConditionsListenerStatus{{
						Port:          port,
						PoolCreated:   false,
						VipCreated:    false,
						VipPoolLinked: false,
					}},
					FirewallRuleCreated: false,
				},
				Listeners: []*pb.LoadBalancerListenerStatus{{
					Name:        "",
					VipID:       0,
					Message:     statusMessage,
					PoolMembers: nil,
					PoolID:      0,
				}},
				State: statusStateActive,
			},
		})
		Expect(err).Should(Succeed())

		getResp, _, err := api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		resourceVersion1 := *getResp.Metadata.ResourceVersion

		By("LoadBalancerServiceUpdate with stored resource version should succeed")
		newMonitor := "tcp"
		newMonitor1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue(newMonitor)
		Expect(err).Should(Succeed())
		newSpecPort := int32(9191)

		_, _, err = api.LoadBalancerServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Metadata: &openapi.LoadBalancerServiceUpdateRequestMetadata{
					ResourceVersion: &resourceVersion1,
				},
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &newSpecPort,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Monitor: newMonitor1,
							Port:    &newSpecPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{
							"1.1.1.1",
						},
					},
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceUpdate with old resource version should fail")

		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue(monitorType)
		Expect(err).Should(Succeed())

		_, _, err = api.LoadBalancerServiceUpdate(ctx, cloudAccountId1, resourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Metadata: &openapi.LoadBalancerServiceUpdateRequestMetadata{
					ResourceVersion: &resourceVersion1,
				},
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &newSpecPort,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Monitor: newMonitor1,
						},
					}},
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())

		By("LoadBalancerServiceGet should return last updated values")
		getResp, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, resourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(getResp.Spec.Listeners[0].Pool.InstanceSelectors).Should(Equal(&map[string]string{"foo": "bar"}))

		Expect(getResp.Spec.Listeners[0].Pool.Monitor).Should(Equal(monitorType1))
		Expect(*getResp.Spec.Listeners[0].Port).Should(Equal(newSpecPort))
	})

	It("ServiceCreate should fail and return 400 BadRequest when creating above the total load balancer quota limit, default quota", func() {
		labels1 := map[string]string{"key1": "value1"}
		port := int32(8080)

		cloudAccountId1 := cloudaccount.MustNewId()

		// Create 1st LB should succeed
		_, err := createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels1)
		Expect(err).Should(Succeed())

		// Create 2nd LB should succeed
		_, err = createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels1)
		Expect(err).Should(Succeed())

		// Create 3rd LB should succeed
		_, err = createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels1)
		Expect(err).Should(Succeed())

		// Create 4th LB should succeed
		_, err = createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels1)
		Expect(err).Should(Succeed())

		// Create 5th LB should succeed
		_, err = createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels1)
		Expect(err).Should(Succeed())

		// Create a 6th should not succeed since the limit is 5.
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId1).Body(
			openapi.LoadBalancerServiceCreateRequest{
				Metadata: &openapi.LoadBalancerServiceCreateRequestMetadata{
					Labels: &labels1,
				},
				Spec: &openapi.ProtoLoadBalancerSpec{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &port,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &poolPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{"1.2.3.4"},
					},
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("ServiceCreate should fail and return 400 BadRequest when creating above the total load balancer quota limit, customer override", func() {
		labels1 := map[string]string{"key1": "value1"}
		port := int32(8080)

		// Creating a single LB should succeed
		_, err := createLoadBalancerSelector(loadBalancerCustomQuotaAccountId, port, []string{"1.2.3.4"}, labels1)
		Expect(err).Should(Succeed())

		// Creating a second should not succeed since the limit is 1
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, loadBalancerCustomQuotaAccountId).Body(
			openapi.LoadBalancerServiceCreateRequest{
				Metadata: &openapi.LoadBalancerServiceCreateRequestMetadata{
					Labels: &labels1,
				},
				Spec: &openapi.ProtoLoadBalancerSpec{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &port,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &poolPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{"1.2.3.4"},
					},
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("ServiceCreate should fail and return 400 BadRequest when creating above the total listener quota limit, customer override", func() {
		labels1 := map[string]string{"key1": "value1"}
		port := int32(8080)
		port2 := int32(8081)
		port3 := int32(8082)

		// Creating a load balancer with more listeners than allowed should not succeed
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, loadBalancerCustomQuotaAccountId).Body(
			openapi.LoadBalancerServiceCreateRequest{
				Metadata: &openapi.LoadBalancerServiceCreateRequestMetadata{
					Labels: &labels1,
				},
				Spec: &openapi.ProtoLoadBalancerSpec{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &port,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &poolPort,
						},
					}, {
						Port: &port2,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &poolPort,
						},
					}, {
						Port: &port3,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &poolPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{"1.2.3.4"},
					},
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("ServiceCreate should fail and return 400 BadRequest when creating sourceIPs with both 'any' and additional", func() {
		labels1 := map[string]string{"key1": "value1"}
		port := int32(8080)

		cloudAccountId1 := cloudaccount.MustNewId()

		// Creating a load balancer with more listeners than allowed should not succeed
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId1).Body(
			openapi.LoadBalancerServiceCreateRequest{
				Metadata: &openapi.LoadBalancerServiceCreateRequestMetadata{
					Labels: &labels1,
				},
				Spec: &openapi.ProtoLoadBalancerSpec{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &port,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &poolPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{"1.2.3.4", "1.2.3.5", "any"},
					},
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("ServiceCreate should fail and return 400 BadRequest when creating above the total sourceIPs quota limit, customer override", func() {
		labels1 := map[string]string{"key1": "value1"}
		port := int32(8080)

		// Creating a load balancer with more listeners than allowed should not succeed
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, loadBalancerCustomQuotaAccountId).Body(
			openapi.LoadBalancerServiceCreateRequest{
				Metadata: &openapi.LoadBalancerServiceCreateRequestMetadata{
					Labels: &labels1,
				},
				Spec: &openapi.ProtoLoadBalancerSpec{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &port,
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Port: &poolPort,
						},
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{"1.2.3.4", "1.2.3.5", "1.2.3.6"},
					},
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceSearch should return load balancers filtered with labels", func() {
		const (
			monitorType = "tcp"
			port        = int32(8080)
		)

		cloudAccountId1 := cloudaccount.MustNewId()
		labels1 := map[string]string{"key1": "value1"}
		labels2 := map[string]string{"key1": "value1", "key2": "value2"}
		labels3 := map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"}
		invalidLabels := map[string]string{"@0919@": "##IDC&rele@$e"}
		labels4 := map[string]string{"key4": "value4"}

		By("LoadBalancerServiceCreate loadbalancer with labels")
		_, err := createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels1)
		Expect(err).Should(Succeed())
		_, err = createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels2)
		Expect(err).Should(Succeed())
		_, err = createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels3)
		Expect(err).Should(Succeed())

		By("LoadBalancerServiceSearch with no labels")
		searchResp, _, err := api.LoadBalancerServiceSearch(ctx, cloudAccountId1).Execute()
		log.Info("Search", "searchResp", searchResp)
		Expect(err).Should(Succeed())
		Expect(len(searchResp.Items)).Should(Equal(3))

		By("InstanceServiceSearch2 with 1 label")
		searchResp2, _, err := api.LoadBalancerServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.LoadBalancerServiceSearch2Request{
				Metadata: &openapi.LoadBalancerServiceSearch2RequestMetadata{
					Labels: &labels1,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).Should(Succeed())
		Expect(len(searchResp2.Items)).Should(Equal(3))

		By("LoadBalancerServiceSearch2 with 2 labels")
		searchResp2, _, err = api.LoadBalancerServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.LoadBalancerServiceSearch2Request{
				Metadata: &openapi.LoadBalancerServiceSearch2RequestMetadata{
					Labels: &labels2,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).Should(Succeed())
		Expect(len(searchResp2.Items)).Should(Equal(2))

		By("LoadBalancerServiceSearch2 with 3 labels")
		searchResp2, _, err = api.LoadBalancerServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.LoadBalancerServiceSearch2Request{
				Metadata: &openapi.LoadBalancerServiceSearch2RequestMetadata{
					Labels: &labels3,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).Should(Succeed())
		Expect(len(searchResp2.Items)).Should(Equal(1))

		By("LoadBalancerServiceSearch2 with labels with no existing instance for those labels")
		searchResp2, _, err = api.LoadBalancerServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.LoadBalancerServiceSearch2Request{
				Metadata: &openapi.LoadBalancerServiceSearch2RequestMetadata{
					Labels: &labels4,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).Should(Succeed())
		Expect(len(searchResp2.Items)).Should(Equal(0))

		By("LoadBalancerServiceSearch2 with invalid labels should return error")
		searchResp2, _, err = api.LoadBalancerServiceSearch2(ctx, cloudAccountId1).Body(
			openapi.LoadBalancerServiceSearch2Request{
				Metadata: &openapi.LoadBalancerServiceSearch2RequestMetadata{
					Labels: &invalidLabels,
				},
			},
		).Execute()
		log.Info("Search", "searchResp", searchResp2)
		Expect(err).ShouldNot(Succeed())
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when loadbalancer body is empty", func() {
		cloudAccountId, _ := baselineInstanceSelector()
		loadbalancerServiceReq := openapi.LoadBalancerServiceCreateRequest{}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when loadbalancer spec is empty", func() {
		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when Port is empty", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
				},
				/* Port: &port, */
			}},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when Port is invalid (low port)", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(0)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when Port is invalid (high port)", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		highPort := int32(95535)
		port := int32(80)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
					Port:    &port,
				},
				Port: &highPort,
			}},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when number of listeners exceeds the max", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		var listeners []openapi.ProtoLoadBalancerListener
		for i := 0; i < DefaultLoadBalancerListenerQuota+10; i++ {

			// Incremenet port to make unique
			tmpPort := port + int32(i)

			listeners = append(listeners, openapi.ProtoLoadBalancerListener{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
					Port:    &tmpPort,
				},
				Port: &tmpPort,
			})
		}

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: listeners,
			Security: &openapi.ProtoLoadBalancerSecurity{
				Sourceips: []string{"1.2.3.4"},
			},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when number of sourceIPs exceeds the max", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		var sourceIPs []string

		for i := 0; i < DefaultLoadBalancerSourceIPQuota+10; i++ {
			sourceIPs = append(sourceIPs, fmt.Sprintf("1.0.0.%d", i))
		}

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
			Security: &openapi.ProtoLoadBalancerSecurity{
				Sourceips: sourceIPs,
			},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when instance resourceid is invalid", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceResourceIds: []string{
						"invalid",
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
			Security: &openapi.ProtoLoadBalancerSecurity{
				Sourceips: []string{"any"},
			},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when instance resourceid is part of a different account", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		cloudAccountIdInvalid := cloudaccount.MustNewId()
		instanceCreateRespInvalid := createInstance(cloudAccountIdInvalid)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()

		instanceCreateResp := createInstance(cloudAccountId)

		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceResourceIds: []string{
						*instanceCreateResp.Metadata.ResourceId,
						*instanceCreateRespInvalid.Metadata.ResourceId,
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
			Security: &openapi.ProtoLoadBalancerSecurity{
				Sourceips: []string{"any"},
			},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when Port listener ports are duplicated", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
					Port:    &port,
				},
				Port: &port,
			}, {
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
					Port:    &port,
				},
				Port: &port,
			}},
			Security: &openapi.ProtoLoadBalancerSecurity{
				Sourceips: []string{
					"any",
				},
			},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceUpdate should fail and return 400 BadRequest when Port listener ports are duplicated", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)
		labels := make(map[string]string)

		// Create a load balancer
		createLoadBalancer, err := createLoadBalancerSelector(cloudAccountId, port, []string{"1.2.3.4"}, labels)
		Expect(err).Should(Succeed())

		// Update with duplicate listener ports
		_, httpResp, err := api.LoadBalancerServiceUpdate(ctx, cloudAccountId, *createLoadBalancer.Metadata.ResourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Monitor: monitorType1,
							Port:    &port,
						},
						Port: &port,
					}, {
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceSelectors: &map[string]string{
								"foo": "bar",
							},
							Monitor: monitorType1,
							Port:    &port,
						},
						Port: &port,
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{
							"any",
						},
					},
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceUpdate should fail and return 400 BadRequest when instance resourceid is invalid", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)
		labels := make(map[string]string)

		// Create a load balancer
		createLoadBalancer, err := createLoadBalancerSelector(cloudAccountId, port, []string{"1.2.3.4"}, labels)
		Expect(err).Should(Succeed())

		// Update with duplicate listener ports
		_, httpResp, err := api.LoadBalancerServiceUpdate(ctx, cloudAccountId, *createLoadBalancer.Metadata.ResourceId).Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Pool: &openapi.ProtoLoadBalancerPool{
							InstanceResourceIds: []string{
								"invalid",
							},
							Monitor: monitorType1,
							Port:    &port,
						},
						Port: &port,
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{
							"any",
						},
					},
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when listeners are missing", func() {
		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when spec.security is not defined", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest when source IPs are not provided", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
			Security: &openapi.ProtoLoadBalancerSecurity{},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest invalid source IP is provided", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
			Security: &openapi.ProtoLoadBalancerSecurity{
				Sourceips: []string{
					"invalid",
				},
			},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest invalid source IP is provided, 10.0.0.1/16", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
			Security: &openapi.ProtoLoadBalancerSecurity{
				Sourceips: []string{
					"10.0.0.1/16",
				},
			},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest invalid subnet IP is provided", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
			Security: &openapi.ProtoLoadBalancerSecurity{
				Sourceips: []string{
					"10.0.0.0/99",
				},
			},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceCreate should fail and return 400 BadRequest with valid IP & invalid subnet IP is provided", func() {
		monitorType1, err := openapi.NewProtoLoadBalancerMonitorTypeFromValue("tcp")
		Expect(err).Should(Succeed())

		port := int32(80)

		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Spec = &openapi.ProtoLoadBalancerSpec{
			Listeners: []openapi.ProtoLoadBalancerListener{{
				Pool: &openapi.ProtoLoadBalancerPool{
					InstanceSelectors: &map[string]string{
						"foo": "bar",
					},
					Monitor: monitorType1,
				},
				Port: &port,
			}},
			Security: &openapi.ProtoLoadBalancerSecurity{
				Sourceips: []string{
					"1.1.1.1",
					"10.0.0.0/99",
				},
			},
		}
		_, httpResp, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
		Expect(httpResp.StatusCode).Should(Equal(http.StatusBadRequest))
	})

	It("LoadBalancerServiceDelete twice with same name should succeed", func() {
		const (
			monitorType = "tcp"
			port        = int32(8080)
		)

		grpcClient := getLoadBalancerGrpcClient()
		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)

		for lbIndex := 0; lbIndex < 2; lbIndex++ {
			desc := fmt.Sprintf(" (loadbalancer %v)", lbIndex)
			By("LoadBalancerServiceCreate" + desc)

			createLoadBalancer, err := createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels)
			Expect(err).Should(Succeed())

			By("ServiceDelete request deletion by name" + desc)
			_, _, err = api.LoadBalancerServiceDelete2(ctx, cloudAccountId1, *createLoadBalancer.Metadata.Name).Execute()
			Expect(err).Should(Succeed())

			By("RemoveFinalizer simulates LoadBalancer Scheduler" + desc)
			_, err = grpcClient.RemoveFinalizer(ctx, &pb.LoadBalancerRemoveFinalizerRequest{
				Metadata: &pb.LoadBalancerIdReference{
					CloudAccountId: cloudAccountId1,
					ResourceId:     *createLoadBalancer.Metadata.ResourceId,
				},
			})
			Expect(err).Should(Succeed())
		}
	})

	It("LoadBalancerService, invalid resourceId returns error", func() {
		const (
			monitorType = "tcp"
			port        = int32(8080)
		)

		cloudAccountId1 := cloudaccount.MustNewId()
		labels := make(map[string]string)

		By("LoadBalancerServiceCreate - InstanceSelector")

		_, err := createLoadBalancerSelector(cloudAccountId1, port, []string{"1.2.3.4"}, labels)
		Expect(err).Should(Succeed())

		By("ServiceDelete request by invalid resourceId")
		_, _, err = api.LoadBalancerServiceDelete(ctx, cloudAccountId1, "1234").Execute()
		Expect(err).ShouldNot(Succeed())
		openAPIError := err.(*openapi.GenericOpenAPIError)
		openAPIErrorModel := openAPIError.Model().(openapi.RpcStatus)
		Expect(*openAPIErrorModel.Message).Should((Equal("invalid resourceId")))

		By("ServiceGet request by invalid resourceId")
		_, _, err = api.LoadBalancerServiceGet(ctx, cloudAccountId1, "1234").Execute()
		Expect(err).ShouldNot(Succeed())
		openAPIError = err.(*openapi.GenericOpenAPIError)
		openAPIErrorModel = openAPIError.Model().(openapi.RpcStatus)
		Expect(*openAPIErrorModel.Message).Should((Equal("invalid resourceId")))

		By("ServiceUpdate request by invalid resourceId")
		newSpecPort := int32(9191)
		_, _, err = api.LoadBalancerServiceUpdate(ctx, cloudAccountId1, "1234").Body(
			openapi.LoadBalancerServiceUpdateRequest{
				Spec: &openapi.ProtoLoadBalancerSpecUpdate{
					Listeners: []openapi.ProtoLoadBalancerListener{{
						Port: &newSpecPort,
					}},
					Security: &openapi.ProtoLoadBalancerSecurity{
						Sourceips: []string{
							"1.1.1.1",
						},
					},
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())
		openAPIError = err.(*openapi.GenericOpenAPIError)
		openAPIErrorModel = openAPIError.Model().(openapi.RpcStatus)
		Expect(*openAPIErrorModel.Message).Should((Equal("invalid resourceId")))
	})

	It("LoadBalancerServiceCreate should fail when loadbalancerName is beginning and ending with -", func() {
		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		lbName := "-my-lb1"
		loadbalancerServiceReq.Metadata.Name = &lbName
		_, _, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())

		lbName = "my-lb-1-"
		loadbalancerServiceReq.Metadata.Name = &lbName
		_, _, err = api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("LoadBalancerServiceCreate should fail when loadbalancerName is invalid", func() {
		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		lbName := "my_lb1"
		loadbalancerServiceReq.Metadata.Name = &lbName
		_, _, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())

		lbName = "my$lb!1"
		loadbalancerServiceReq.Metadata.Name = &lbName
		_, _, err = api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())

		lbName = "My_lb_1"
		loadbalancerServiceReq.Metadata.Name = &lbName
		_, _, err = api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("LoadBalancerServiceCreate should fail when loadbalancerName exceeeds 63 characters", func() {
		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		lbName := "loadbalancer_name_exceeding_63_characters_in_it_s_name_for_testing__"
		loadbalancerServiceReq.Metadata.Name = &lbName
		_, _, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

	It("LoadBalancerServiceCreate should fail when label name has an invalid character", func() {
		cloudAccountId, loadbalancerServiceReq := baselineInstanceSelector()
		loadbalancerServiceReq.Metadata.Labels = &map[string]string{
			"invalid!!!.label.com/name": "value",
		}
		_, _, err := api.LoadBalancerServiceCreate(ctx, cloudAccountId).Body(loadbalancerServiceReq).Execute()
		Expect(err).ShouldNot(Succeed())
	})

})
