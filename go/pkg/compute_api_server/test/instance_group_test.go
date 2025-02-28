// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package test

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var _ = Describe("InstanceGroup API tests using OpenAPI client", func() {
	const (
		pubKey1 = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC38LFb8lQmcT6KiDuvPu3N9XPvhE5ShbDtxcNtc1AqdsV7MRH7uxYIXDVd0tM80dgEKwyi3IzbNNILGWUkxhV9A3bEnVqPNG7Up6rHdo72uwK1koY3KIlu6BzBBB8QpDvGWUwP4DU84zwC4UJxHEtFL7qnUHKzfNAq8a1WCvABtUW3eaB1SjKzGfNWYR7X8/JwZUCUtTCAFy7gaFrNL7XXDNUzfBzXvlKiyOSbSxWXa51f66sPmPdPtLyveo+3PeruTTvWpCPXJW5zOmeDwbBx899pbc72f1U/KJt6fdwxMSDXdSbARC7ONhD2MRoHjdbbl5QkZbLxdtm/jq393vNCxSP6s8/RDg3+Xp/u0LMjW78JqjKMkKnwWIrSlABQEihM7AlsEKLHDXMredUhT9uXgdu/XOn5Q7zNAkHhOr10p8DpD2HqNkVfcmBf3HhWx6HWS6i7teJiinuAtlVvRD7Xaw8xZM1wue0V/lNF74dE9NKLxSBBhOAvsojr7kZorQLRoULWz8nMBSNoomrUdVVB+UGFdiYS07exYvHemCEXMDxRXl/72Cfkv+DCtkgACjbI2qk8h+kgzsAe9DUmOfxFO1JSDdOevU7eDGMEIDn5REbdsXztgjjdiokANwDTWyTWyDRJMoUhtK0TKM21PHxoEnW32AcHH7LeeHTJiI9sFQ== user1@example.org"
	)

	var (
		err error

		instanceGroupServiceApi *openapi.InstanceGroupServiceApiService
		instanceServiceApi      *openapi.InstanceServiceApiService

		cloudAccountId    string
		sshPublicKeyName  string
		sshPublicKeyNames []string
		availabilityZone  string
		instanceType      string
		machineImage      string
		vNet              string
		instanceCount     int

		instanceGroupName1 string
		instanceGroup1     []openapi.ProtoInstance
	)

	BeforeEach(func(ctx context.Context) {
		By("Clearing Database")
		clearDatabase(ctx)

		instanceGroupServiceApi = openApiClient.InstanceGroupServiceApi
		instanceServiceApi = openApiClient.InstanceServiceApi

		cloudAccountId = cloudaccount.MustNewId()
		sshPublicKeyName = "key1-" + uuid.New().String()
		sshPublicKeyNames = []string{sshPublicKeyName}
		instanceType = createBmInstanceType(ctx, "bm-spr-gaudi2")
		availabilityZone = "us-dev-1a"
		machineImage, err = createBmMachineImage(ctx, []pb.InstanceCategory{pb.InstanceCategory_BareMetalHost}, []string{})
		Expect(err).NotTo(HaveOccurred())
		vNet = CreateVNet(ctx, cloudAccountId)
		instanceGroupName1 = "test-instance-group-1"

		By("Creating SSH public key")
		createSshPublicKey(cloudAccountId, sshPublicKeyName, pubKey1)
	})

	Describe("API: Delete instance group member", func() {
		var expectedDeletionCount int32

		BeforeEach(func(ctx context.Context) {
			By("Creating a group of 4")
			instanceCount = 4
			_, err := createInstanceGroup(ctx, cloudAccountId, sshPublicKeyNames, availabilityZone, instanceType, machineImage, vNet, instanceGroupName1, instanceCount)
			Expect(err).NotTo(HaveOccurred())

			By("Searching instance group")
			instances, err := searchInstancesFromGroup(ctx, instanceServiceApi, cloudAccountId, instanceGroupName1)
			Expect(err).NotTo(HaveOccurred())
			Expect(instances.Items).To(HaveLen(instanceCount))
			instanceGroup1 = instances.Items

			expectedDeletionCount = 0
		})

		Context("With valid request", func() {

			It("should delete the instance from the group by resouceId", func(ctx context.Context) {
				By("Deleting the instance from group")
				Eventually(func(g Gomega) {
					deletedResourceId := instanceGroup1[0].Metadata.ResourceId
					_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember(ctx, cloudAccountId, instanceGroupName1, *deletedResourceId).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
					expectedDeletionCount = 1
				}).Should(Succeed())
			})

			It("should delete the instance from the group by name", func(ctx context.Context) {
				By("Deleting instance from group")
				Eventually(func(g Gomega) {
					instanceName := instanceGroup1[0].Metadata.Name
					_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember2(ctx, cloudAccountId, instanceGroupName1, *instanceName).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
					expectedDeletionCount = 1
				}).Should(Succeed())
			})

			It("should do nothing if the instance is already being deleted", func(ctx context.Context) {
				instanceName := instanceGroup1[0].Metadata.Name
				By("Deleting instance from group repetedly")
				Eventually(func(g Gomega) {
					_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember2(ctx, cloudAccountId, instanceGroupName1, *instanceName).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
				}).Should((Succeed()))

				By("Deleting the same instance while it is being deleted")
				Eventually(func(g Gomega) {
					_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember2(ctx, cloudAccountId, instanceGroupName1, *instanceName).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
				}).Should(Succeed())
				expectedDeletionCount = 1
			})

			It("should handle concurrent requests", func(ctx context.Context) {
				By("Deleting all instances from group except one")
				instancesToDelete := instanceGroup1[1:]
				for _, in := range instancesToDelete {
					go func(instanceName *string) {
						defer GinkgoRecover()
						By("Deleting instance")
						_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember2(ctx, cloudAccountId, instanceGroupName1, *instanceName).Execute()
						Expect(err).NotTo(HaveOccurred())
						Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
						atomic.AddInt32(&expectedDeletionCount, 1)
					}(in.Metadata.Name)
				}
				Eventually(func() bool {
					return expectedDeletionCount == int32(len(instancesToDelete))
				}).Within(2 * time.Second).Should(BeTrue())
			})

			AfterEach(func(ctx context.Context) {
				By("Searching instance group")
				instances, err := searchPrivateInstancesFromGroup(ctx, cloudAccountId, instanceGroupName1)
				Expect(err).NotTo(HaveOccurred())

				By("Checking deletion timestamps")
				remainingCount := 0
				for _, in := range instances.Items {
					if in.Metadata.DeletionTimestamp == nil {
						remainingCount++
					}
				}
				Expect(remainingCount).To(Equal(instanceCount - int(expectedDeletionCount)))

				By("Checking the group size")
				for _, in := range instances.Items {
					for _, other := range instances.Items {
						Expect(in.Spec.InstanceGroupSize).To(Equal(other.Spec.InstanceGroupSize))
					}
				}
			})
		})

		Context("With invalid request", func() {

			It("should return an error if the instance is not found", func(ctx context.Context) {
				instanceName := "non-existing-instance"
				By("Ensuring the instance is gone")
				Eventually(func(g Gomega) {
					_, httpResp, err := instanceServiceApi.InstanceServiceGet2(ctx, cloudAccountId, instanceName).Execute()
					g.Expect(err).To(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusNotFound))
				}, "10s").Should((Succeed()))

				By("Deleting the same instance when it is gone")
				Eventually(func(g Gomega) {
					_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember2(ctx, cloudAccountId, instanceGroupName1, instanceName).Execute()
					g.Expect(err).To(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusNotFound))
				}, "3s").Should(Succeed())
			})

			It("should not delete the last instance in the group", func(ctx context.Context) {
				By("Deleting all but one instance from group")
				numInstancesToDelete := len(instanceGroup1) - 1
				for i := 0; i < numInstancesToDelete; i++ {
					Eventually(func(g Gomega) {
						instanceName := instanceGroup1[i].Metadata.Name
						_, _, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember2(ctx, cloudAccountId, instanceGroupName1, *instanceName).Execute()
						g.Expect(err).NotTo(HaveOccurred())
					}).Should(Succeed())
				}

				By("Searching instance group")
				instances, err := searchInstancesFromGroup(ctx, instanceServiceApi, cloudAccountId, instanceGroupName1)
				Expect(err).NotTo(HaveOccurred())

				By("Having one instance remaining in the group")
				remainingInstances := []openapi.ProtoInstance{}
				for _, in := range instances.Items {
					if in.Metadata.DeletionTimestamp == nil {
						remainingInstances = append(remainingInstances, in)
					}
				}
				Expect(remainingInstances).To(HaveLen(1))

				By("Deleting the last instance from group")
				Eventually(func(g Gomega) {
					instanceName := remainingInstances[0].Metadata.Name
					_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember2(ctx, cloudAccountId, instanceGroupName1, *instanceName).Execute()
					g.Expect(err).To(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusBadRequest))
				}).Should(Succeed())

			})

			It("should not delete an instance from a different group", func(ctx context.Context) {
				By("Creating another group")
				instanceCount = 2
				instanceGroupName2 := "another-group"
				_, err := createInstanceGroup(ctx, cloudAccountId, sshPublicKeyNames, availabilityZone, instanceType, machineImage, vNet, instanceGroupName2, instanceCount)
				Expect(err).NotTo(HaveOccurred())

				By("Deleting an instance from another group")
				Eventually(func(g Gomega) {
					instanceName := instanceGroup1[0].Metadata.Name
					_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember2(ctx, cloudAccountId, instanceGroupName2, *instanceName).Execute()
					g.Expect(err).To(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusNotFound))
				}).Should(Succeed())
			})
		})
	})

	Describe("API: Scale up instance group", func() {
		var instanceGroup *pb.InstanceGroup

		BeforeEach(func(ctx context.Context) {
			By("Creating a group of 2")
			instanceCount = 2
			instanceGroup, err = createInstanceGroup(ctx, cloudAccountId, sshPublicKeyNames, availabilityZone, instanceType, machineImage, vNet, instanceGroupName1, instanceCount)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("With valid request", func() {
			var desiredCount int32

			BeforeEach(func(ctx context.Context) {
				By("Scaling up to 4")
				desiredCount = int32(4)
				Eventually(func(g Gomega) {
					resp, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceScaleUp(ctx, instanceGroup.Metadata.CloudAccountId, instanceGroup.Metadata.Name).Body(
						openapi.InstanceGroupServiceScaleUpRequest{
							Spec: &openapi.ProtoInstanceGroupSpec{
								InstanceCount: &desiredCount,
							},
						}).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
					g.Expect(*resp.Status.CurrentCount).To(And(
						Equal(*resp.Status.DesiredCount),
						Equal(desiredCount)))
					g.Expect(resp.Status.CurrentMembers).To(HaveLen(int(desiredCount)))
					g.Expect(resp.Status.NewMembers).To(HaveLen(2))
				}).Should(Succeed())
			})

			It("should not create new instances if the group is at the desired count", func(ctx context.Context) {
				By("Scaling up to same count (4)")
				Eventually(func(g Gomega) {
					resp, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceScaleUp(ctx, instanceGroup.Metadata.CloudAccountId, instanceGroup.Metadata.Name).Body(
						openapi.InstanceGroupServiceScaleUpRequest{
							Spec: &openapi.ProtoInstanceGroupSpec{
								InstanceCount: &desiredCount,
							},
						}).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
					g.Expect(*resp.Status.CurrentCount).To(And(
						Equal(*resp.Status.DesiredCount),
						Equal(desiredCount)))
					g.Expect(resp.Status.CurrentMembers).To(HaveLen(int(desiredCount)))
					g.Expect(resp.Status.NewMembers).To(HaveLen(0))
				}).Should(Succeed())

				By("Searching instance group")
				instances, err := searchInstancesFromGroup(ctx, instanceServiceApi, cloudAccountId, instanceGroupName1)
				Expect(err).NotTo(HaveOccurred())

				By("Checking instances in the group")
				instanceGroupSize := int32(len(instances.Items))
				Expect(instanceGroupSize).To(Equal(desiredCount))
			})

			It("should not create a new instance to replace the one that is still being deleted", func(ctx context.Context) {
				By("Deleting an instance from the group")
				deletingInstanceName := instanceGroup.Metadata.Name + "-0"
				Eventually(func(g Gomega) {
					_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDeleteMember2(ctx, cloudAccountId, instanceGroup.Metadata.Name, deletingInstanceName).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
				}).Should(Succeed())

				By("Scaling up to 5")
				desiredCount := int32(5)
				newInstanceName := instanceGroup.Metadata.Name + "-4"
				Eventually(func(g Gomega) {
					resp, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceScaleUp(ctx, instanceGroup.Metadata.CloudAccountId, instanceGroup.Metadata.Name).Body(
						openapi.InstanceGroupServiceScaleUpRequest{
							Spec: &openapi.ProtoInstanceGroupSpec{
								InstanceCount: &desiredCount,
							},
						}).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
					g.Expect(*resp.Status.DesiredCount).To(Equal(desiredCount))
					g.Expect(*resp.Status.CurrentCount).To(Equal(desiredCount - 1))
					g.Expect(resp.Status.CurrentMembers).To(HaveLen(int(desiredCount - 1)))
					g.Expect(resp.Status.CurrentMembers).To(ContainElement(newInstanceName))
					g.Expect(resp.Status.CurrentMembers).NotTo(ConsistOf(deletingInstanceName))
					g.Expect(resp.Status.NewMembers).NotTo(ConsistOf(deletingInstanceName))
					g.Expect(resp.Status.NewMembers).To(HaveLen(1))
					g.Expect(resp.Status.NewMembers).To(ConsistOf(newInstanceName))
				}).Should(Succeed())

				By("Searching instance group")
				instances, err := searchInstancesFromGroup(ctx, instanceServiceApi, cloudAccountId, instanceGroupName1)
				Expect(err).NotTo(HaveOccurred())

				By("Checking instances in the group")
				instanceGroupSize := int32(len(instances.Items))
				Expect(instanceGroupSize).To(Equal(desiredCount))
			})

			It("should create new instances if the group has not reached the desired count", func(ctx context.Context) {
				By("Searching instance group")
				instances, err := searchInstancesFromGroup(ctx, instanceServiceApi, cloudAccountId, instanceGroupName1)
				Expect(err).NotTo(HaveOccurred())

				By("Checking instances in the group")
				instanceGroupSize := int32(len(instances.Items))
				Expect(instanceGroupSize).To(Equal(desiredCount))
			})

			It("should update the instances in the group", func(ctx context.Context) {
				By("Searching instance group")
				resp, err := searchPrivateInstancesFromGroup(ctx, cloudAccountId, instanceGroupName1)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.Items).To(HaveLen(int(desiredCount)))

				By("Checking instances in the group")
				for _, instance := range resp.Items {
					Expect(instance.Spec.InstanceGroupSize).To(Equal(desiredCount))
					for _, other := range resp.Items {
						Expect(instance.Spec.Region).To(Equal(other.Spec.Region))
						Expect(instance.Spec.AvailabilityZone).To(Equal(other.Spec.AvailabilityZone))
						Expect(instance.Spec.InstanceType).To(Equal(other.Spec.InstanceType))
						Expect(instance.Spec.MachineImage).To(Equal(other.Spec.MachineImage))
						Expect(instance.Spec.UserData).To(Equal(other.Spec.UserData))
						Expect(instance.Spec.InstanceGroup).To(Equal(other.Spec.InstanceGroup))
						Expect(instance.Spec.ClusterGroupId).To(Equal(other.Spec.ClusterGroupId))
						Expect(instance.Spec.SuperComputeGroupId).To(Equal(other.Spec.SuperComputeGroupId))
						Expect(instance.Spec.InstanceGroupSize).To(Equal(other.Spec.InstanceGroupSize))
						Expect(len(instance.Spec.Interfaces)).To(BeNumerically(">", 1))
						Expect(len(other.Spec.Interfaces)).To(BeNumerically(">", 1))
						Expect(instance.Spec.Interfaces[0].VNet).To(Equal(other.Spec.Interfaces[0].VNet))
						Expect(instance.Spec.Interfaces[1].VNet).To(Equal(other.Spec.Interfaces[1].VNet))
						Expect(instance.Spec.SshPublicKeyNames).To(BeEquivalentTo(other.Spec.SshPublicKeyNames))
						Expect(instance.Spec.NetworkMode).To(Equal(other.Spec.NetworkMode))
					}
				}
			})

			It("should allow overwriting the userdata on the new instances", func(ctx context.Context) {
				By("Scaling up to 6")
				desiredCount := int32(6)
				newUserdata := "new-userdata"
				Eventually(func(g Gomega) {
					resp, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceScaleUp(ctx, instanceGroup.Metadata.CloudAccountId, instanceGroup.Metadata.Name).Body(
						openapi.InstanceGroupServiceScaleUpRequest{
							Spec: &openapi.ProtoInstanceGroupSpec{
								InstanceCount: &desiredCount,
								InstanceSpec: &openapi.ProtoInstanceSpec{
									UserData: &newUserdata,
								},
							},
						}).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
					g.Expect(*resp.Status.CurrentCount).To(And(
						Equal(*resp.Status.DesiredCount),
						Equal(desiredCount)))
				}).Should(Succeed())

				By("Searching instance group")
				result, err := searchPrivateInstancesFromGroup(ctx, cloudAccountId, instanceGroupName1)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Items).To(HaveLen(int(desiredCount)))

				By("Checking instances in the group")
				newInstance := result.Items[len(result.Items)-1]
				// userdata is omitted by search API if exists
				Expect(newInstance.Spec.UserData).To(Equal("OMITTED"))
			})
		})

		Context("With invalid request", func() {

			It("should not allow scaling down", func(ctx context.Context) {
				By("Scaling down by 1")
				Eventually(func(g Gomega) {
					desiredCount := int32(instanceCount - 1)
					resp, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceScaleUp(ctx, instanceGroup.Metadata.CloudAccountId, instanceGroup.Metadata.Name).Body(
						openapi.InstanceGroupServiceScaleUpRequest{
							Spec: &openapi.ProtoInstanceGroupSpec{
								InstanceCount: &desiredCount,
							},
						}).Execute()
					g.Expect(err).To(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusBadRequest))
					g.Expect(resp).To(BeNil())
				}).Should(Succeed())
			})

			It("should not allow scaling to zero", func(ctx context.Context) {
				By("Scaling to 0")
				Eventually(func(g Gomega) {
					desiredCount := int32(0)
					resp, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceScaleUp(ctx, instanceGroup.Metadata.CloudAccountId, instanceGroup.Metadata.Name).Body(
						openapi.InstanceGroupServiceScaleUpRequest{
							Spec: &openapi.ProtoInstanceGroupSpec{
								InstanceCount: &desiredCount,
							},
						}).Execute()
					g.Expect(err).To(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusBadRequest))
					g.Expect(resp).To(BeNil())
				}).Should(Succeed())
			})

			It("should not allow scaling up when the whole instance group is already being deleted", func(ctx context.Context) {
				By("Deleting the instance group")
				Eventually(func(g Gomega) {
					_, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceDelete(ctx, instanceGroup.Metadata.CloudAccountId, instanceGroup.Metadata.Name).Execute()
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusOK))
				}).Should(Succeed())

				By("Scaling up")
				Eventually(func(g Gomega) {
					desiredCount := int32(4)
					resp, httpResp, err := instanceGroupServiceApi.InstanceGroupServiceScaleUp(ctx, instanceGroup.Metadata.CloudAccountId, instanceGroup.Metadata.Name).Body(
						openapi.InstanceGroupServiceScaleUpRequest{
							Spec: &openapi.ProtoInstanceGroupSpec{
								InstanceCount: &desiredCount,
							},
						}).Execute()
					g.Expect(err).To(HaveOccurred())
					g.Expect(httpResp).To(HaveHTTPStatus(http.StatusNotFound))
					g.Expect(resp).To(BeNil())
				}).Should(Succeed())
			})
		})
	})
})

func createBmInstanceType(ctx context.Context, name string) string {
	By("Creating instance type client")
	var instanceTypeName string
	computeApiServerAddress := fmt.Sprintf("localhost:%d", grpcListenPort)
	clientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).Should(Succeed())
	instanceTypeClient := pb.NewInstanceTypeServiceClient(clientConn)

	By("Creating InstanceType")
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
			InstanceCategory: pb.InstanceCategory_BareMetalHost,
			Cpu: &pb.CpuSpec{
				Cores:     40,
				Sockets:   2,
				Threads:   2,
				Id:        "0x80380",
				ModelName: "3rd Generation Intel® Xeon® Platinum 8380 Processors (Ice Lake)",
			},
			Gpu: &pb.GpuSpec{
				Count:     8,
				ModelName: "HL-225",
			},
			Description: "3rd Generation Intel® Xeon® Platinum 8380 Processors (Ice Lake)",
			Disks: []*pb.DiskSpec{
				{Size: "2000Gi"},
			},
			DisplayName: "8 Gaudi2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk",
			Memory: &pb.MemorySpec{
				DimmCount: 32,
				Speed:     4000,
				DimmSize:  "32Gi",
				Size:      "1024Gi",
			},
		},
	}
	_, err = instanceTypeClient.Put(ctx, instanceType)
	Expect(err).Should(Succeed())
	return instanceTypeName
}

func createBmMachineImage(ctx context.Context, instanceCategories []pb.InstanceCategory, instanceTypes []string) (string, error) {
	By("Creating machine image client")
	computeApiServerAddress := fmt.Sprintf("localhost:%d", grpcListenPort)
	clientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	machineImageClient := pb.NewMachineImageServiceClient(clientConn)

	By("Creating MachineImage")
	machineImageName := uuid.NewString()
	machineImage := &pb.MachineImage{
		Metadata: &pb.MachineImage_Metadata{
			Name: machineImageName,
		},
		Spec: &pb.MachineImageSpec{
			DisplayName:        "Ubuntu 22.04 LTS (Jammy Jellyfish) v20240212",
			Description:        "Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.",
			Icon:               "https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png",
			InstanceCategories: instanceCategories,
			InstanceTypes:      instanceTypes,
			Labels: map[string]string{
				"architecture": "X86_64 (Baremetal only)",
				"family":       "ubuntu-2204-lts",
			},
			ImageCategories: []string{
				"App Testing",
			},
			Components: []*pb.MachineImageComponent{
				{
					Name:        "Ubuntu 22.04 LTS",
					Type:        "OS",
					Version:     "22.04",
					Description: "Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.",
					InfoUrl:     "https://releases.ubuntu.com/jammy",
					ImageUrl:    "https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png",
				},
				{
					Name:        "SynapseAI SW",
					Type:        "Software Kit",
					Version:     "1.14.0",
					Description: "Designed to facilitate high-performance DL training on Habana Gaudi accelerators.",
					InfoUrl:     "https://docs.habana.ai/en/latest/SW_Stack_Packages_Installation/Synapse_SW_Stack_Installation.html#sw-stack-packages-installation",
				},
			},
			UserName: "sdp",
		},
	}
	_, err = machineImageClient.Put(ctx, machineImage)
	if err != nil {
		return "", err
	}
	return machineImageName, nil
}

func createInstanceGroup(ctx context.Context, cloudAccountId string, sshPublicKeyNames []string, availabilityZone string, instanceType string,
	machineImage string, vNet string, instanceGroupName string, instanceCount int) (*pb.InstanceGroup, error) {
	grpcClient := getInstanceGroupGrpcClient()
	createResp, err := grpcClient.Create(ctx, &pb.InstanceGroupCreateRequest{
		Metadata: &pb.InstanceGroupMetadataCreate{
			CloudAccountId: cloudAccountId,
			Name:           instanceGroupName,
		},
		Spec: &pb.InstanceGroupSpec{
			InstanceCount: int32(instanceCount),
			InstanceSpec: &pb.InstanceSpec{
				AvailabilityZone:  availabilityZone,
				InstanceType:      instanceType,
				MachineImage:      machineImage,
				RunStrategy:       pb.RunStrategy_RerunOnFailure,
				SshPublicKeyNames: sshPublicKeyNames,
				Interfaces: []*pb.NetworkInterface{
					{
						Name: "Interface1",
						VNet: vNet,
					},
				},
			},
		},
	})

	return createResp, err
}

func searchInstancesFromGroup(ctx context.Context, instanceServiceApi *openapi.InstanceServiceApiService, cloudAccountId, instanceGroupName string) (*openapi.ProtoInstanceSearchResponse, error) {
	resp, _, err := instanceServiceApi.InstanceServiceSearch2(ctx, cloudAccountId).Body(
		openapi.InstanceServiceSearch2Request{
			Metadata: &openapi.InstanceServiceSearch2RequestMetadata{
				InstanceGroup:       &instanceGroupName,
				InstanceGroupFilter: openapi.NewProtoInstanceMetadataSearch().InstanceGroupFilter,
			},
		},
	).Execute()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func searchPrivateInstancesFromGroup(ctx context.Context, cloudAccountId, instanceGroupName string) (*pb.InstanceSearchPrivateResponse, error) {
	resp, err := getInstanceGrpcClient().SearchPrivate(ctx, &pb.InstanceSearchPrivateRequest{
		Metadata: &pb.InstanceMetadataSearch{
			CloudAccountId:      cloudAccountId,
			InstanceGroup:       instanceGroupName,
			InstanceGroupFilter: pb.SearchFilterCriteria_ExactValue,
		},
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
