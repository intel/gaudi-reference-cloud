package bmaas_multitenancy

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("BMaaS Regression - Business use case's", Label("compute", "bm_multitenancy", "bm_same_CA"), Ordered, func() {
	var (
		firstBMName, secondBMName, bmPayload            string
		firstResourceId, secondResourceId, instanceType string
	)

	BeforeAll(func() {
		firstBMName = "first-automation-bm-" + utils.GetRandomString()
		secondBMName = "second-automation-bm-" + utils.GetRandomString()
		bmPayload = utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
	})

	When("First instance - Starting Instance Creation flow via Instance API", func() {
		It("First instance - Starting Instance Creation flow via Instance API", func() {
			createStatusCode, createRespBody := service_apis.CreateInstanceWithMIMap(instanceEndpointCA1, token, bmPayload, firstBMName, instanceType,
				sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+firstBMName+`"`)).To(BeTrue(), createRespBody)
			firstResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceValidation := utils.CheckInstancePhase(instanceEndpointCA1, token, firstResourceId)
			Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
		})
	})

	When("Second instance - Starting Instance Creation flow via Instance API", func() {
		It("Second instance - Starting Instance Creation flow via Instance API", func() {
			createStatusCode, createRespBody := service_apis.CreateInstanceWithMIMap(instanceEndpointCA2, token, bmPayload, secondBMName, instanceType,
				secondSSHKeyName, secondVnet, machineImageMapping, availabilityZone)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+secondBMName+`"`)).To(BeTrue(), createRespBody)
			secondResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceValidation := utils.CheckInstancePhase(instanceEndpointCA2, token, secondResourceId)
			Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
		})
	})

	When("SSH into the instance present inside CA1 and ping the instance present", func() {
		It("SSH into the instance present inside CA1 and ping the instance present", func() {
			_, createRespBody1 := service_apis.GetInstanceById(instanceEndpointCA1, token, firstResourceId)
			_, createRespBody2 := service_apis.GetInstanceById(instanceEndpointCA2, token, secondResourceId)

			err := utils.SSHIntoInstanceMultiTenancy(createRespBody1, createRespBody2, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
				"../../../../ansible-files/ssh-ping-between-2-bm-diff-ca.yml", "~/.ssh/id_rsa")
			Expect(err).NotTo(HaveOccurred(), err)
		})
	})

	AfterAll(func() {
		logInstance.Println("Delete the first instance")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpointCA1, token, firstResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion for first cloud account")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpointCA1, token, firstResourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())

		logInstance.Println("Delete the second instance")
		deleteStatusCode, deleteRespBody = service_apis.DeleteInstanceById(instanceEndpointCA2, token, secondResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion for second cloud account")
		instanceValidation = utils.CheckInstanceDeletionById(instanceEndpointCA2, token, secondResourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
	})
})
