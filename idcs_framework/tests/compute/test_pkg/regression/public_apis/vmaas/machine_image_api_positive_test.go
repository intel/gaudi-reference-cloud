package vmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Compute machine image endpoint(VM positive flow)", Label("compute", "vmaas", "vmaas_machine_image", "vmaas_machine_image_positive"), Ordered, ContinueOnFailure, func() {
	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Listing all the machine images", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve all the supported machineimages...")
			statusCode, responseBody := service_apis.GetAllMachineImage(machineImageEndpoint, token)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Retrieve VM machine image using name", func() {
		It("should be successful", func() {
			machineImages := []string{"ubuntu-2204-jammy-v20230122"}
			for _, machineImage := range machineImages {
				logInstance.Println("Retrieve an machine image via predefined name : " + machineImage)
				statusCode, responseBody := service_apis.GetInstanceTypeByName(machineImageEndpoint, token, machineImage)
				Expect(statusCode).To(Equal(200), responseBody)
				Expect(strings.Contains(responseBody, `"name":"`+machineImage+`"`)).To(BeTrue(), responseBody)
			}
		})
	})
})
