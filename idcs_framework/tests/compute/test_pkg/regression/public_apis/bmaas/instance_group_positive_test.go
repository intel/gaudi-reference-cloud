package bmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
	"time"
)

var _ = Describe("Compute instance group endpoint(BM positive flow)", Label("compute", "bmaas", "bmaas_instance_group", "bmaas_instance_group_positive"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType          string
		instanceGroupName     string
		instanceGroupPayload  string
		instanceSearchPayload string
	)

	BeforeAll(func() {
		if skipBMClusterCreation == "true" {
			Skip("Skipping the Entire BM negative flow due to the flag")
		}
		// load instance group details
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		instanceSearchPayload = utils.GetJsonValue("instanceSearchPayload")
		instanceGroupPayload = utils.GetJsonValue("instanceGroupPayload")
		instanceGroupName = "automation-ins-group-" + utils.GetRandomString()
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Creating instances within cluster or group", func() {
		It("instances should be created within cluster", func() {
			logInstance.Println("Starting the Instance Creation flow via Instance Group API...")
			createStatusCode, createRespBody := service_apis.InstanceGroupCreationWithMIMap(instanceGroupEndpoint, token, instanceGroupPayload,
				instanceGroupName, "2", instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+instanceGroupName+`"`)).To(BeTrue(), createRespBody)

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceGroupValidation := utils.CheckInstanceGroupProvisionState(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 20*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Listing all the clusters or instance groups", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve all the instances present in any group...")
			statusCode, responseBody := service_apis.GetAllInstanceGroups(instanceGroupEndpoint, token)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test GET API to retrieve the instance group using name", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance group using name")
			statusCode, responseBody := service_apis.GetInstancesWithinGroup(instanceEndpoint, token, instanceGroupName)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test GET API to retrieve the instance group using name and 'Default' as filter", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance group using name and 'Default' as filter")
			params := map[string]string{
				"metadata.instanceGroup":       instanceGroupName,
				"metadata.instanceGroupFilter": "Default",
			}
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, params)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test GET API to retrieve the instance group using name and 'Empty' as filter", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance group using name and 'Empty' as filter")
			params := map[string]string{
				"metadata.instanceGroup":       instanceGroupName,
				"metadata.instanceGroupFilter": "Empty",
			}
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, params)
			Expect(statusCode).To(Equal(200), responseBody)
			//Expect(strings.Contains(responseBody, `"items":[]`)).To(BeTrue(), responseBody)
		})
	})

	When("test GET API to retrieve the instance group using name and 'NonEmpty' as filter", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance group using name and 'NonEmpty' as filter")
			params := map[string]string{
				"metadata.instanceGroup":       instanceGroupName,
				"metadata.instanceGroupFilter": "NonEmpty",
			}
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, params)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test GET API to retrieve the instance group using name and 'Any' as filter", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance group using name and 'Any' as filter")
			params := map[string]string{
				"metadata.instanceGroup":       instanceGroupName,
				"metadata.instanceGroupFilter": "Any",
			}
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, params)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test GET API to retrieve the instance group using name and 'ExactValue' as filter", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance group using name and 'ExactValue' as filter")
			params := map[string]string{
				"metadata.instanceGroup":       instanceGroupName,
				"metadata.instanceGroupFilter": "Any",
			}
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, params)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test GET API to retrieve the instance group using only 'Any' as filter", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance group using only 'Any' as filter")
			params := map[string]string{
				"metadata.instanceGroupFilter": "Any",
			}
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, params)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test GET API to retrieve the instance group using only 'NonEmpty' as filter", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance group using only 'NonEmpty' as filter")
			params := map[string]string{
				"metadata.instanceGroupFilter": "NonEmpty",
			}
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, params)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test GET API to retrieve the instance group using only 'Empty' as filter", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance group using only 'Empty' as filter")
			params := map[string]string{
				"metadata.instanceGroupFilter": "Empty",
			}
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, params)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test PATCH API to scale-up the instance group", func() {
		It("should be successful", func() {
			logInstance.Println("Add instances to already created instance group using scaleup API")
			instanceGroupPatchPayload := utils.GetJsonValue("instanceGroupPatchPayload")
			statusCode, responseBody := service_apis.InstanceGroupScaleUp(instanceGroupEndpoint, token, instanceGroupPatchPayload,
				"3", instanceGroupName)
			Expect(statusCode).To(Equal(200), responseBody)

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceGroupValidation := utils.CheckInstanceGroupProvisionState(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 20*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("test DELETE API to delete an instance from the instance group using Id", func() {
		It("should be successful", func() {
			logInstance.Println("Delete an instance from instance group using instanceResourceId")
			listOfIds := utils.GetInstanceIdsFromInstanceGroup(instanceEndpoint, token, instanceGroupName, instanceSearchPayload)
			logInstance.Println("listOfIds ", listOfIds)

			statusCode, responseBody := service_apis.DeleteSingleInstanceFromGroupById(instanceGroupEndpoint, token, instanceGroupName, listOfIds[0])
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("test DELETE API to delete an instance from the instance group using name", func() {
		It("should be successful", func() {
			logInstance.Println("Delete an instance from instance group using instanceName")
			listofNames := utils.GetInstanceNamesFromInstanceGroup(instanceEndpoint, token, instanceGroupName, instanceSearchPayload)
			logger.Logf.Info("list of instance names ", listofNames)

			statusCode, responseBody := service_apis.DeleteSingleInstanceFromGroupByName(instanceGroupEndpoint, token, instanceGroupName, listofNames[0])
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Deleting cluster or instance group using name", func() {
		It("should be successful", func() {
			logInstance.Println("Delete all instances present inside group using name...")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceGroupByName(instanceGroupEndpoint, token, instanceGroupName)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Group Deletion using Name")
			instanceGroupValidation := utils.CheckInstanceGroupDeletionByName(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 3*time.Minute, 5*time.Second).Should(BeTrue())

		})
	})
})
