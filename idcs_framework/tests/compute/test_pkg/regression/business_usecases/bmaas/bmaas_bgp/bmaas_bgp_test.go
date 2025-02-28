package bu_bgp

import (
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/auth"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	kube "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis/bmaas"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("BMaaS BGP", Label("compute", "bmaas_bgp"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType              string
		instanceGroupName         string
		listOfIds                 []string
		listOfBMH                 []string
		deviceName                string
		isInstanceGroupCreated    bool
		createRespBody            string
		createStatusCode          int
		instanceGroupPayload      string
		instanceGroupPatchPayload string
		instanceSearchPayload     string
	)

	BeforeAll(func() {
		// instances under group creation
		logInstance.Println("Before all...")
		instanceGroupName = "automation-group-bgp-" + utils.GetRandomString()
		instanceGroupPayload = utils.GetJsonValue("instanceGroupPayload")
		instanceSearchPayload = utils.GetJsonValue("instanceSearchPayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreatedBGP")
		instanceGroupPatchPayload = utils.GetJsonValue("instanceGroupPatchPayload")

		logInstance.Println("Starting the BGP instance group via API...")
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

	When("Instance Group creation and its validation - prerequisite", func() {
		It("Validate the instance Group creation in before all", func() {
			logInstance.Println("is instance group created? " + strconv.FormatBool(isInstanceGroupCreated))
			Expect(isInstanceGroupCreated).Should(BeTrue(), "Instance Group creation failed with following error "+createRespBody)
		})
	})

	When("Ensure bgp0 accelerator vnet and storage vnet are created for each instance", func() {
		It("Ensure bgp0 accelerator vnet and storage vnet are created for each instance", func() {
			logInstance.Println("Ensure bgp0 accelerator vnet and storage vnet are created for each instance")
			payload := instanceSearchPayload
			payload = strings.Replace(payload, "<<instance-group-name>>", instanceGroupName, 1)
			statusCode, responseBody := service_apis.SearchInstances(instanceEndpoint, token, payload)
			Expect(statusCode).To(Equal(200), responseBody)

			acceleratorVnets := gjson.Get(responseBody, "items.#.spec.interfaces.1.name").Array()
			logInstance.Println("accelerator vnets:", acceleratorVnets)
			for _, vnet := range acceleratorVnets {
				Expect(vnet.String()).To(Equal("bgp0"), responseBody)
			}

			// Storage interface is only enabled for sc instance types
			/*
				storageVnets := gjson.Get(responseBody, "items.#.spec.interfaces.1.name").Array()
				logInstance.Println("storage vnets:", storageVnets)
				for _, vnet := range storageVnets {
					Expect(vnet.String()).To(Equal("storage0"), responseBody)
				}
			*/
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

	When("test PATCH API to scale-up the instance group", func() {
		It("should be successful", func() {
			logInstance.Println("Add instances to already created instance group using scaleup API")
			statusCode, responseBody := service_apis.InstanceGroupScaleUp(instanceGroupEndpoint, token, instanceGroupPatchPayload, "4", instanceGroupName)
			Expect(statusCode).To(Equal(200), responseBody)

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceGroupValidation := utils.CheckInstanceGroupProvisionState(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
		})
	})

	When("Ensure all the reserved nodes are with network-mode=XBX", func() {
		It("Check system state (kube)", func() {
			logInstance.Println("Ensure all the reserved nodes are with network-mode=XBX")
			for _, bmhname := range listOfBMH {
				By("Ensure all the reserved nodes are with network-mode=XBX")
				Eventually(func(g Gomega) {
					bmh, err := kube.GetBmhByName(bmhname, kubeConfigPath)
					Expect(err).Error().ShouldNot(HaveOccurred(), "error fetching device config: %v", err)
					Expect(bmh.Labels).Should(HaveKeyWithValue("cloud.intel.com/network-mode", "XBX"))
				})
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

	When("Ensure validation operator validates the BMH", func() {
		It("Check system state (kube)", func() {
			logInstance.Println("Wait until device is in available state before validation...")
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
				}, 1*time.Hour, 30*time.Second).Should(Succeed())
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
		// refresh token
		var err error
		token, err = auth.Get_Azure_Bearer_Token(userEmail)
		Expect(err).ShouldNot(HaveOccurred(), err)
		getStatusCode, getResponseBody := service_apis.GetInstancesWithinGroup(instanceEndpoint, token, instanceGroupName)
		Expect(getStatusCode).To(Equal(200), getResponseBody)

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
