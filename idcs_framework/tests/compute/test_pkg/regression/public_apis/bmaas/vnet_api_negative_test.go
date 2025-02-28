package bmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Compute VNET api endpoint(BM negative flow)", Label("compute", "bmaas", "compute_vnet", "bmaas_vnet", "bmaas_vnet_negative"), Ordered, func() {
	var vnetPayload string

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	BeforeAll(func() {
		vnetPayload = utils.GetJsonValue("vnetPayload")
	})

	When("Creating VNET with invalid char length on name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create vnet creation with invalid char length")
			statusCode, responseBody := service_apis.CreateVNet(vnetEndpoint, token, vnetPayload,
				"vnet-name-to-validate-the-character-length-for-testing-purpose-attempt1")
			Expect(statusCode).To(Equal(500), responseBody)
			Expect(strings.Contains(responseBody, "an unknown error occurred")).To(BeTrue(), responseBody)
		})
	})

	When("Creating VNET without name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("VNet creation without name...")
			payload := vnetPayload
			payload = strings.Replace(payload, "<<vnet-name>>", "", 1)
			statusCode, responseBody := service_apis.CreateVNet(vnetEndpoint, token, payload, "")
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `"missing metadata.name"`)).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving VNET using invalid ID", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to get vnet with invalid id")
			statusCode, responseBody := service_apis.GetVnetById(vnetEndpoint, token, "invalid-id")
			Expect(statusCode).To(Equal(500), responseBody)
			Expect(strings.Contains(responseBody, `"message":"an unknown error occurred"`)).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving VNET using invalid name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to get vnet with invalid name")
			statusCode, responseBody := service_apis.GetVnetByName(vnetEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
		})
	})

	When("Deleting VNET using invalid ID", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to delete vnet with invalid id")
			statusCode, responseBody := service_apis.DeleteVnetById(vnetEndpoint, token, "invalid-id")
			Expect(statusCode).To(Equal(500), responseBody)
			Expect(strings.Contains(responseBody, `"message":"an unknown error occurred"`)).To(BeTrue(), responseBody)
		})
	})

	When("Deleting VNET using invalid name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to delete vnet with invalid name")
			statusCode, responseBody := service_apis.DeleteVnetByName(vnetEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
		})
	})
})
