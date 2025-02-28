package vmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Compute machine image endpoint(VM negative flow)", Label("compute", "vmaas", "vmaas_machine_image", "vmaas_machine_image_negative"), Ordered, ContinueOnFailure, func() {
	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Retrieving machine image using invalid name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Retrieve an machine image using invalid name...")
			statusCode, responseBody := service_apis.GetInstanceTypeByName(machineImageEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"resource not found"`)).To(BeTrue(), responseBody)
		})

	})
})
