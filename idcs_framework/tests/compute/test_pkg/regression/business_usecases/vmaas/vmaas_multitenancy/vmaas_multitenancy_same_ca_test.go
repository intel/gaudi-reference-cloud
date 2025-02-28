package vmaas_multitenancy

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("VMaaS Multitenancy within same CA - Business use case's", Label("compute", "vm_multi_CA"), Ordered, func() {
	var (
		firstVMName      string
		secondVMName     string
		vmPayload        string
		firstResourceId  string
		secondResourceId string
		instanceType     string
	)

	BeforeAll(func() {
		firstVMName = "first-automation-vm-" + utils.GetRandomString()
		secondVMName = "second-automation-vm-" + utils.GetRandomString()
		vmPayload = utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
	})

	When("First instance - Starting the Small Instance Creation flow via Instance API", func() {
		It("First instance - Starting the Small Instance Creation flow via Instance API", func() {
			createStatusCode, createRespBody := service_apis.CreateInstanceWithMIMap(instanceEndpointCA1, token, vmPayload, firstVMName, instanceType,
				sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+firstVMName+`"`)).To(BeTrue(), createRespBody)
			firstResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceValidation := utils.CheckInstancePhase(instanceEndpointCA1, token, firstResourceId)
			Eventually(instanceValidation, 5*time.Minute, 10*time.Second).Should(BeTrue())
		})
	})

	When("Second instance - Starting the Small Instance Creation flow via Instance API", func() {
		It("Second instance - Starting the Small Instance Creation flow via Instance API", func() {
			createStatusCode, createRespBody := service_apis.CreateInstanceWithMIMap(instanceEndpointCA1, token, vmPayload, secondVMName, instanceType,
				sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+secondVMName+`"`)).To(BeTrue(), createRespBody)
			secondResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceValidation := utils.CheckInstancePhase(instanceEndpointCA1, token, secondResourceId)
			Eventually(instanceValidation, 5*time.Minute, 10*time.Second).Should(BeTrue())
		})
	})

	When("SSH into the instance present inside CA1 and ping the instance present", func() {
		It("SSH into the instance present inside CA1 and ping the instance present", func() {
			_, getRespBody1 := service_apis.GetInstanceById(instanceEndpointCA1, token, firstResourceId)
			_, getRespBody2 := service_apis.GetInstanceById(instanceEndpointCA1, token, secondResourceId)

			err := utils.SSHIntoInstanceMultiTenancy(getRespBody1, getRespBody2, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
				"../../../../ansible-files/ssh-on-vm-to-ping-another-vm.yml", "~/.ssh/id_rsa")
			Expect(err).NotTo(HaveOccurred(), err)
		})
	})

	When("Delete the first instance and second instance", func() {
		It("should be successful", func() {
			logInstance.Println("Delete the first instance and second instance")
		})
	})

	AfterAll(func() {
		logInstance.Println("Delete the first instance")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpointCA1, token, firstResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of first instance deletion")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpointCA1, token, firstResourceId)
		Eventually(instanceValidation, 2*time.Minute, 5*time.Second).Should(BeTrue())

		logInstance.Println("Delete the second instance")
		deleteStatusCode2, deleteRespBody2 := service_apis.DeleteInstanceById(instanceEndpointCA1, token, secondResourceId)
		Expect(deleteStatusCode2).To(Equal(200), deleteRespBody2)

		// Validation
		logInstance.Println("Validation of second instance deletion")
		instanceValidation2 := utils.CheckInstanceDeletionById(instanceEndpointCA1, token, secondResourceId)
		Eventually(instanceValidation2, 2*time.Minute, 5*time.Second).Should(BeTrue())
	})
})
