package bmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
	"time"
)

var _ = Describe("Compute instance group endpoint(BM negative flow)", Label("compute", "bmaas", "bmaas_instance_group", "bmaas_instance_group_negative"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType              string
		instanceGroupPayload      string
		instanceGroupPatchPayload string
		instanceSearchPayload     string
	)

	BeforeAll(func() {
		if skipBMClusterCreation == "true" {
			Skip("Skipping the Entire BM negative flow due to the flag")
		}

		// load instance group details to be created
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		instanceGroupPayload = utils.GetJsonValue("instanceGroupPayload")
		instanceSearchPayload = utils.GetJsonValue("instanceSearchPayload")
		instanceGroupPatchPayload = utils.GetJsonValue("instanceGroupPatchPayload")
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Creating an instance group with invalid instance type in payload", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an instance with invalid instance type")
			statusCode, responseBody := service_apis.InstanceGroupCreationWithMIMap(instanceGroupEndpoint, token, instanceGroupPayload,
				"instance-grp-with-invalid-type", "2", "invalid", sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `"message":"unable to get instance type`)).To(BeTrue(), responseBody)
		})
	})

	When("Creating an instance group with invalid name in payload", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an instance with invalid instance group - char length")
			statusCode, responseBody := service_apis.InstanceGroupCreationWithMIMap(instanceGroupEndpoint, token, instanceGroupPayload,
				"instance-group-name-to-validate-the-character-length-for-testing-purpose-attempt", "2", instanceType, sshkeyName,
				vnet, machineImageMapping, availabilityZone)
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `"message":"invalid instance name"`)).To(BeTrue(), responseBody)
		})
	})

	When("Creating an instance group with invalid machine image in payload", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an instance with invalid machine image")
			statusCode, responseBody := service_apis.InstanceGroupCreationWithoutMIMap(instanceGroupEndpoint, token, instanceGroupPayload,
				"instance-grp-with-invalid-machineimage1", "2", instanceType, sshkeyName, vnet, "invalid", availabilityZone)
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `"message":"unable to get machine image`)).To(BeTrue(), responseBody)
		})
	})

	When("Creating an instance group with invalid SSH key in payload", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an instance with invalid SSH key")
			statusCode, responseBody := service_apis.InstanceGroupCreationWithMIMap(instanceGroupEndpoint, token, instanceGroupPayload,
				"instance-grp-with-invalid-ssh", "2", instanceType, "invalid-ssh-key", vnet, machineImageMapping, availabilityZone)
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"unable to get SSH public key`)).To(BeTrue(), responseBody)
		})
	})

	When("Creating an instance group with invalid instance count in payload", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an instance with invalid instance count")
			statusCode, responseBody := service_apis.InstanceGroupCreationWithMIMap(instanceGroupEndpoint, token, instanceGroupPayload,
				"instance-grp-with-invalid-instance-cnt", "12345", instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `"message":"invalid instance count"`)).To(BeTrue(), responseBody)
		})
	})

	When("Add instances to a group with invalid group name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Add instances to a group with invalid group name")
			statusCode, responseBody := service_apis.InstanceGroupScaleUp(instanceGroupEndpoint, token, instanceGroupPatchPayload,
				"2", "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"require at lest one instance in instanceGroup invalid-name to scale",`)).To(BeTrue(), responseBody)
		})
	})

	Context("Tests that do not require an account", func() {
		var instanceGroupName string

		BeforeAll(func() {
			logInstance.Println("Create instances within group or cluster...")
			instanceGroupName = "automation-ins-group-negative-" + utils.GetRandomString()
			// instances under group creation
			logInstance.Println("Starting the Instance Creation flow via Instance Group API...")
			createStatusCode, createRespBody := service_apis.InstanceGroupCreationWithMIMap(instanceGroupEndpoint, token,
				instanceGroupPayload, instanceGroupName, "2", instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+instanceGroupName+`"`)).To(BeTrue(), createRespBody)

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceGroupValidation := utils.CheckInstanceGroupProvisionState(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 20*time.Minute, 5*time.Second).Should(BeTrue())
		})

		When("Ensure decreasing the instanceCount using patch API return error", func() {
			It("should fail with valid error...", func() {
				logInstance.Println("Ensure decreasing the instanceCount using scale-up API return error")
				statusCode, responseBody := service_apis.InstanceGroupScaleUp(instanceGroupEndpoint, token, instanceGroupPatchPayload,
					"1", instanceGroupName)
				Expect(statusCode).To(Equal(400), responseBody)
				Expect(strings.Contains(responseBody, `"message":"scaling down is unsupported. currentCount=2, desiredCount=1"`)).To(BeTrue(), responseBody)
			})
		})

		When("Add instance to a group with invalid count number(invalid payload)", func() {
			It("should fail with valid error...", func() {
				logInstance.Println("Add instance to a group with invalid count number(invalid payload)")
				statusCode, responseBody := service_apis.InstanceGroupScaleUp(instanceGroupEndpoint, token, instanceGroupPatchPayload,
					"invalid", instanceGroupName)
				Expect(statusCode).To(Equal(400), responseBody)
				Expect(strings.Contains(responseBody, `invalid value for int32 type`)).To(BeTrue(), responseBody)
			})
		})

		When("Add instances to group with exceeded count limit", func() {
			It("should fail with valid error...", func() {
				logInstance.Println("Add instances to group with exceeded count limit")
				statusCode, responseBody := service_apis.InstanceGroupScaleUp(instanceGroupEndpoint, token, instanceGroupPatchPayload,
					"3545", instanceGroupName)
				Expect(statusCode).To(Equal(400), responseBody)
				Expect(strings.Contains(responseBody, `"message":"invalid instance count"`)).To(BeTrue(), responseBody)
			})
		})

		When("Add instance to a group with empty count number(empty payload)", func() {
			It("should fail with valid error...", func() {
				logInstance.Println("Add instance to a group with empty count number(empty payload)")
				statusCode, responseBody := service_apis.InstanceGroupScaleUp(instanceGroupEndpoint, token, instanceGroupPatchPayload,
					"", instanceGroupName)
				Expect(statusCode).To(Equal(400), responseBody)
				Expect(strings.Contains(responseBody, `invalid value for int32 type`)).To(BeTrue(), responseBody)
			})
		})

		When("Delete an instance from group using invalid resourceId", func() {
			It("should fail with valid error...", func() {
				logInstance.Println("Delete an instance from group using invalid resourceId")
				statusCode, responseBody := service_apis.DeleteSingleInstanceFromGroupById(instanceGroupEndpoint, token, instanceGroupName,
					"invalid-id")
				Expect(statusCode).To(Equal(404), responseBody)
				Expect(strings.Contains(responseBody, `"message":"instance with resourceId invalid-id not found in instanceGroup `+instanceGroupName+`"`)).To(BeTrue(), responseBody)
			})
		})

		When("Delete an instance from group using invalid instanceName", func() {
			It("should fail with valid error...", func() {
				logInstance.Println("Delete an instance from group using invalid instanceName")
				statusCode, responseBody := service_apis.DeleteSingleInstanceFromGroupByName(instanceGroupEndpoint, token, instanceGroupName,
					"invalid-name")
				Expect(statusCode).To(Equal(404), responseBody)
				Expect(strings.Contains(responseBody, `"message":"instance with name invalid-name not found in instanceGroup `+instanceGroupName+`"`)).To(BeTrue(), responseBody)
			})
		})

		When("Delete the last instance in group and ensure it fails", func() {
			It("should fail with valid error...", func() {
				logInstance.Println("Delete the last instance in group and ensure it fails")
				listOfIds := utils.GetInstanceIdsFromInstanceGroup(instanceEndpoint, token, instanceGroupName, instanceSearchPayload)
				for i := 0; i < len(listOfIds); i++ {
					statusCode, responseBody := service_apis.DeleteSingleInstanceFromGroupById(instanceGroupEndpoint, token, instanceGroupName,
						listOfIds[i])
					if i < len(listOfIds)-1 {
						Expect(statusCode).To(Equal(200), responseBody)
					} else {
						Expect(statusCode).To(Equal(400), responseBody)
						Expect(strings.Contains(responseBody, `"message":"deleting the last remaining instance from instanceGroup is not allowed by this method."`)).To(BeTrue(), responseBody)
					}
				}
			})
		})

		AfterAll(func() {
			logInstance.Println("Delete the instance group using name...")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceGroupByName(instanceGroupEndpoint, token, instanceGroupName)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Group Deletion using Name")
			instanceGroupValidation := utils.CheckInstanceGroupDeletionByName(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 3*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Delete an instance from invalid group using resourceId", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Delete an instance from invalid group using resourceId")
			statusCode, responseBody := service_apis.DeleteSingleInstanceFromGroupById(instanceGroupEndpoint, token, "invalid-group",
				"invalid-id")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"no instances found for instanceGroup invalid-group"`)).To(BeTrue(), responseBody)
		})
	})

	When("Delete an instance from invalid group using name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Delete an instance from invalid group using name")
			statusCode, responseBody := service_apis.DeleteSingleInstanceFromGroupByName(instanceGroupEndpoint, token, "invalid-group",
				"invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"no instances found for instanceGroup invalid-group"`)).To(BeTrue(), responseBody)
		})
	})
})
