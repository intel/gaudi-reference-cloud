package validation_operator

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	kube "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis/bmaas"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("Compute BM Validation Operator(Positive flow)", Label("compute", "validation_operator", "instance_validation"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType       string
		createStatusCode   int
		createRespBody     string
		instanceResourceId string
		deviceName         string
		vmName             string
		vmPayload          string
		isInstanceCreated  = false
	)

	BeforeAll(func() {
		if skipBMCreation == "true" {
			Skip("Skipping the Entire BM negative flow due to the flag")
		}
		vmName = "automation-vm-" + utils.GetRandomString()
		vmPayload = utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreatedValidationOperator")

		// Instance creation
		logInstance.Println("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload,
			vmName, instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
		isInstanceCreated = true
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Instance creation and its validation - prerequisite", func() {
		It("Validate the instance creation in before all", func() {
			logInstance.Println("is instance created? " + strconv.FormatBool(isInstanceCreated))
			Expect(isInstanceCreated).Should(BeTrue(), "Instance creation failed with following error "+createRespBody)
		})
	})

	When("Get BMH device by resource ID", func() {
		It("Get BMH device by resource ID", func() {
			logInstance.Println("Starting the BMH Device Retrieval via Kube...")
			time.Sleep(5 * time.Second)
			bmhResponse, err := kube.GetBmhByConsumer(instanceResourceId, kubeConfigPath)
			Expect(err).Error().ShouldNot(HaveOccurred())
			Expect(bmhResponse.Spec.ConsumerRef.Name).To(Equal(instanceResourceId))
			deviceName = bmhResponse.ObjectMeta.Name
		})

	})

	When("Validate BMH device is provisoned", func() {
		It("Validate BMH device is provisoned", func() {
			logInstance.Println("Starting the BMH Device Validation via Kube...")
			succeded, err := kube.CheckBMHState(deviceName, "provisioned", 1800, kubeConfigPath)
			Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+"unable to reach state within expected time")
			Expect(succeded).To(Equal(true))
		})
	})

	When("Instance deletion via DELETE api using resource id", func() {
		It("Instance deletion via DELETE api using resource id", func() {
			logInstance.Println("Instance deletion via DELETE api using resource id")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Deletion using Id")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
			Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
		})
	})

	When("Validate instance deprovision", func() {
		It("Validate instance deprovision", func() {
			logInstance.Println("Starting the Instance Deprovision Validation via Kube...")
			succeded, err := kube.CheckBMHState(deviceName, "available", 1800, kubeConfigPath)
			Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+"unable to reach state within expected time")
			Expect(succeded).To(Equal(true))
		})
	})

	When("Ensure validation operator validates the BMH", func() {
		It("Check system state (kube)", func() {
			By("Wait until device is in available state before validation")
			succeded, err := kube.CheckBMHState(deviceName, "available", 1800, kubeConfigPath)
			Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+"unable to reach state within expected time")
			Expect(succeded).To(Equal(true))
			By("Wait until device has been verified")
			Eventually(func(g Gomega) {
				bmh, err := kube.GetBmhByName(deviceName, kubeConfigPath)
				Expect(err).Error().ShouldNot(HaveOccurred(), "error Setting Role")
				g.Expect(bmh.Labels).Should(HaveKeyWithValue("cloud.intel.com/verified", "true"))
			}, 45*time.Minute, 5*time.Second).Should(Succeed())
			By("Wait until device is in available state")
			succeded, err = kube.CheckBMHState(deviceName, "available", 1800, kubeConfigPath)
			Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+"unable to reach state within expected time")
			Expect(succeded).To(Equal(true))
		})
	})

	AfterAll(func() {
		// Delete the instance if any testcases above fails
		logInstance.Println("Ensure the created instance is deleted")
		getStatusCode, _ := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)

		if getStatusCode != 404 {
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			logInstance.Println("Validation of Instance Deletion using Id")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
			Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
		} else {
			logInstance.Println("Instance had been deleted")
		}
	})
})
