package validation_operator

import ()

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	kube "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis/bmaas"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("Compute BM Validation Operator(Positive flow)", Label("compute", "validation_operator", "group_validation"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType           string
		createStatusCode       int
		createRespBody         string
		deviceName             string
		instanceGroupName      string
		instanceGroupPayload   string
		instanceSearchPayload  string
		isInstanceGroupCreated = false
		listOfIds              []string
		listOfBMH              []string
	)

	BeforeAll(func() {
		if skipBMClusterCreation == "true" {
			Skip("Skipping the Entire BM negative flow due to the flag")
		}
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		instanceSearchPayload = utils.GetJsonValue("instanceSearchPayload")
		instanceGroupName = "automation-ins-group-" + utils.GetRandomString()
		instanceGroupPayload = utils.GetJsonValue("instanceGroupPayload")

		logInstance.Println("Starting the Instance Gropup Creation flow via API...")
		createStatusCode, createRespBody = service_apis.InstanceGroupCreationWithMIMap(instanceGroupEndpoint, token, instanceGroupPayload,
			instanceGroupName, "2", instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+instanceGroupName+`"`)).To(BeTrue(), createRespBody)

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceGroupValidation := utils.CheckInstanceGroupProvisionState(instanceEndpoint, token, instanceGroupName)
		Eventually(instanceGroupValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
		isInstanceGroupCreated = true
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Instance group creation and its validation - prerequisite", func() {
		It("Validate the instance group creation in before all", func() {
			logInstance.Println("is instance group created? " + strconv.FormatBool(isInstanceGroupCreated))
			Expect(isInstanceGroupCreated).Should(BeTrue(), "Instance group creation failed with following error "+createRespBody)
		})
	})

	When("Get the list of resource Ids in the group", func() {
		It("Get the list of resource Ids in the group", func() {
			listOfIds = utils.GetInstanceIdsFromInstanceGroup(instanceEndpoint, token, instanceGroupName, instanceSearchPayload)
			logInstance.Println("listOfIds ", listOfIds)
		})
	})

	When("Get BMH devices by resource IDs", func() {
		It("Get BMH devices by resource IDs", func() {
			logInstance.Println("Starting the BMH Device Retrieval via Kube...")
			time.Sleep(5 * time.Second)

			listOfBMH = []string{}
			for _, id := range listOfIds {
				bmhResponse, err := kube.GetBmhByConsumer(id, kubeConfigPath)
				Expect(err).Error().ShouldNot(HaveOccurred())
				Expect(bmhResponse.Spec.ConsumerRef.Name).To(Equal(id))
				deviceName = bmhResponse.ObjectMeta.Name
				listOfBMH = append(listOfBMH, deviceName)
			}
		})

	})

	When("Validate BMH devices are provisoned", func() {
		It("Validate BMH devices are provisoned", func() {
			logInstance.Println("Starting the BMH Devices Validation via Kube...")
			for _, bmhName := range listOfBMH {
				succeded, err := kube.CheckBMHState(bmhName, "provisioned", 1800, kubeConfigPath)
				Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+"unable to reach state within expected time")
				Expect(succeded).To(Equal(true))
			}
		})
	})

	When("Instance group deletion via DELETE api using name", func() {
		It("Instance group deletion via DELETE api using name", func() {
			logInstance.Println("Delete instance group using name...")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceGroupByName(instanceGroupEndpoint, token, instanceGroupName)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Group Deletion using Name")
			instanceGroupValidation := utils.CheckInstanceGroupDeletionByName(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
		})
	})

	When("Validate instance group deprovision", func() {
		It("Validate instance group deprovision", func() {
			logInstance.Println("Starting the Instance group Deprovision Validation via Kube...")
			for _, bmhName := range listOfBMH {
				succeded, err := kube.CheckBMHState(bmhName, "available", 1800, kubeConfigPath)
				Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+"unable to reach state within expected time")
				Expect(succeded).To(Equal(true))
			}
		})
	})

	When("Ensure validation operator validates the BMH", func() {
		It("Check system state (kube)", func() {
			for _, bmhname := range listOfBMH {
				By("Wait until device is in available state before validation")
				succeded, err := kube.CheckBMHState(bmhname, "available", 3600, kubeConfigPath)
				Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+"unable to reach state within expected time")
				Expect(succeded).To(Equal(true))
				By("Wait until device has been verified")
				Eventually(func(g Gomega) {
					bmh, err := kube.GetBmhByName(bmhname, kubeConfigPath)
					Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting Role")
					g.Expect(bmh.Labels).Should(HaveKeyWithValue("cloud.intel.com/verified", "true"))
				}, 45*time.Minute, 5*time.Second).Should(Succeed())
				By("Wait until device is in available state")
				succeded, err = kube.CheckBMHState(bmhname, "available", 3600, kubeConfigPath)
				Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+"unable to reach state within expected time")
				Expect(succeded).To(Equal(true))
			}
		})
	})

	AfterAll(func() {
		// Delete the instance if any testcases above fails
		logInstance.Println("Ensure the created instance group is deleted")
		getStatusCode, _ := service_apis.GetInstancesWithinGroup(instanceEndpoint, token, instanceGroupName)

		if getStatusCode != 404 {
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceGroupByName(instanceGroupEndpoint, token, instanceGroupName)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Group Deletion using Name")
			instanceGroupValidation := utils.CheckInstanceGroupDeletionByName(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
		} else {
			logInstance.Println("Instance Group had been deleted")
		}
	})
})
