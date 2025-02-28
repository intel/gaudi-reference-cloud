package vmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Compute instance type endpoint(VM negative flow)", Label("compute", "vmaas", "vmaas_instance_type", "vmaas_instance_type_negative"), Ordered, func() {
	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Retrieving instance type using invalid name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Retrieve an instance type via invalid name...")
			statusCode, responseBody := service_apis.GetInstanceTypeByName(instanceTypeEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"resource not found"`)).To(BeTrue(), responseBody)
		})
	})
})
