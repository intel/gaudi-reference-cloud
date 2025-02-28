package vmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Compute instance type endpoint(VM positive flow)", Label("compute", "vmaas", "vmaas_instance_type", "vmaas_instance_type_positive"), Ordered, ContinueOnFailure, func() {
	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Listing all the instance types", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve all the supported instance types...")
			statusCode, responseBody := service_apis.GetAllInstanceType(instanceTypeEndpoint, token)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Retrieve VM instance type using name", func() {
		It("should be successful", func() {
			instanceTypes := []string{"vm-spr-tny", "vm-spr-sml", "vm-spr-med", "vm-spr-lrg"}
			for _, instanceType := range instanceTypes {
				logInstance.Println("Retrieve an instance type via predefined instance type - name : " + instanceType)
				statusCode, responseBody := service_apis.GetInstanceTypeByName(instanceTypeEndpoint, token, instanceType)
				Expect(statusCode).To(Equal(200), responseBody)
				Expect(strings.Contains(responseBody, `"name":"`+instanceType+`"`)).To(BeTrue(), responseBody)
			}
		})
	})
})
