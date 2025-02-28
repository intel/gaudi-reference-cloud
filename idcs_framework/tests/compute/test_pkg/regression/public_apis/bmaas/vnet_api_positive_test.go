package bmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("Compute VNET api endpoint(BM positive flow)", Label("compute", "bmaas", "compute_vnet", "bmaas_vnet", "bmaas_vnet_positive"), Ordered, ContinueOnFailure, func() {

	var (
		createStatusCode int
		vnetName         string
		createRespBody   string
		vnetPayload      string
		vnetResourceId   string
	)
	BeforeAll(func() {
		// Load vnet details
		vnetPayload = utils.GetJsonValue("vnetPayload")
		vnetName = "automation-vnet-" + utils.GetRandomString()

		// vnet creation
		logInstance.Println("Starting the Vnet Creation flow via API...")
		createStatusCode, createRespBody = service_apis.CreateVNet(vnetEndpoint, token, vnetPayload, vnetName)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vnetName+`"`)).To(BeTrue(), createRespBody)
		vnetResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()
		time.Sleep(10 * time.Second)
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Listing all available VNET", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve all the vnet's available via GET method")
			statusCode, responseBody := service_apis.GetAllVnet(vnetEndpoint, token)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Retrieving VNET using valid resource id", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the vnet via GET method using id")
			statusCode, responseBody := service_apis.GetVnetById(vnetEndpoint, token, vnetResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, vnetResourceId)).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving VNET using valid resource name", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the vnet via GET method using name")
			statusCode, responseBody := service_apis.GetVnetByName(vnetEndpoint, token, vnetName)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, vnetName)).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving already create VNET", func() {
		It("subnet value should be correct", func() {
			logInstance.Println("Retrieve the vnet via GET method using name and validate subnet value")
			statusCode, responseBody := service_apis.GetVnetByName(vnetEndpoint, token, vnetName)
			Expect(statusCode).To(Equal(200), responseBody)
			vnet_prefix := gjson.Get(responseBody, "spec.prefixLength").Int()
			Expect(int(vnet_prefix)).To(Or(Equal(22), Equal(27)), responseBody)
		})
	})

	When("Deleting VNET using valid resource name", func() {
		It("should be successful", func() {
			logInstance.Println("Remove the vnet via Delete method using name")

			// Creation of vnet
			secondvNetName := "automation-vnet-" + utils.GetRandomString()
			statusCode, responseBody := service_apis.CreateVNet(vnetEndpoint, token, vnetPayload, secondvNetName)
			time.Sleep(10 * time.Second)
			Expect(statusCode).To(Equal(200), responseBody)

			// Deletion of vnet via name
			deleteStatusCode, deleteRespBody := service_apis.DeleteVnetByName(vnetEndpoint, token, secondvNetName)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Retrieve after deletion
			statusCode, responseBody = service_apis.GetVnetByName(vnetEndpoint, token, secondvNetName)
			Expect(statusCode).To(Equal(404), responseBody)
		})
	})

	When("Deleting VNET using valid resource name", func() {
		It("should be successful", func() {
			logInstance.Println("Remove the vnet via Delete method using id")
			deleteStatusCode, deleteRespBody := service_apis.DeleteVnetById(vnetEndpoint, token, vnetResourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			statusCode, responseBody := service_apis.GetVnetById(vnetEndpoint, token, vnetResourceId)
			Expect(statusCode).To(Equal(404), responseBody)
		})
	})
})
