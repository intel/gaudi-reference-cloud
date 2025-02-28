package vmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("Compute Instance API endpoint(VM negative flow)", Label("compute", "vmaas", "compute_instance", "vmaas_instance", "vmaas_instance_negative"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType       string
		createRespBody     string
		vmName             string
		vmPayload          string
		instanceResourceId string
		createStatusCode   int
		isInstanceCreated  bool
	)

	BeforeAll(func() {
		// load prerequisites for instance creation
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		vmName = "automation-vm-" + utils.GetRandomString()
		vmPayload = utils.GetJsonValue("instancePayload")

		logInstance.Println("Starting the Tiny Instance Creation flow via Instance API for negative scenarios")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload,
			vmName, instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()
		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
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

	When("Creating an instance using invalid instance type", func() {
		It("should fail with client error", func() {
			logInstance.Println("Attempt to create an instance with invalid instance type")
			statusCode, responseBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, "vm-wo-instancetype", "invalid-type",
				sshkeyName, vnetName, machineImageMapping, availabilityZone)
			Expect(statusCode).To(Equal(403), responseBody)
		})
	})

	When("Creating an instance using invalid machine image", func() {
		It("should fail with client error", func() {
			logInstance.Println("Attempt to create an instance with invalid machine image")
			statusCode, responseBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, vmPayload, "vm-wo-machineimage",
				"vm-spr-tny", sshkeyName, vnetName, "invalid", availabilityZone)
			Expect(statusCode).To(Equal(403), responseBody)
		})
	})

	When("Creating an instance using invalid SSH public key", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an instance with invalid SSH public key")
			vnNameNegative := "automation-vm-invalid" + utils.GetRandomString()
			statusCode, responseBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, vnNameNegative, instanceType,
				"invalid-ssh", vnet, machineImageMapping, availabilityZone)
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, "resource not found")).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving an instance using invalid resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to get an instance with invalid id")
			statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, "invalid-id")
			Expect(statusCode).To(Equal(500), responseBody)
			Expect(strings.Contains(responseBody, "internal server error")).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving an instance using invalid resource name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to get an instance with invalid name")
			statusCode, responseBody := service_apis.GetInstanceByName(instanceEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, "resource not found")).To(BeTrue(), responseBody)
		})
	})

	When("Updating an instance with invalid SSH key using resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to Update an instance with invalid field in payload via resource id")
			instancePutPayload := utils.GetJsonValue("instancePutPayload")
			instancePutPayload = strings.Replace(instancePutPayload, "<<ssh-public-key>>", "invalid-key", 1)
			statusCode, responseBody := service_apis.PutInstanceById(instanceEndpoint, token, instanceResourceId, instancePutPayload)
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, "resource not found")).To(BeTrue(), responseBody)
		})
	})

	When("Updating an instance with invalid SSH key using resource name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to Update an instance with invalid field in payload via resource name")
			instanceName := gjson.Get(createRespBody, "metadata.name").String()
			instancePutPayload := utils.GetJsonValue("instancePutPayload")
			instancePutPayload = strings.Replace(instancePutPayload, "<<ssh-public-key>>", "invalid-key", 1)
			statusCode, responseBody := service_apis.PutInstanceByName(instanceEndpoint, token, instanceName, instancePutPayload)
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, "resource not found")).To(BeTrue(), responseBody)
		})
	})

	When("Deleting an instance using invalid resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to delete an instance with invalid id")
			statusCode, responseBody := service_apis.DeleteInstanceById(instanceEndpoint, token, "invalid-id")
			Expect(statusCode).To(Equal(500), responseBody)
			Expect(strings.Contains(responseBody, `"message":"internal server error"`)).To(BeTrue(), responseBody)
		})
	})

	When("Deleting an instance using invalid resource name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to delete an instance with invalid name")
			statusCode, responseBody := service_apis.DeleteInstanceByName(instanceEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
		})
	})

	When("Creating an instance with too many char's in name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an instance with invalid instance name with too many characters")
			instanceName := "instance-name-to-validate-the-character-length-for-testing-purpose-attempt" + utils.GetRandomString()
			statusCode, responseBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, instanceName, "vm-spr-sml",
				sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `"message":"invalid instance name"`)).To(BeTrue(), responseBody)
		})
	})

	When("Remove the instance via DELETE api using resource id", func() {
		It("Placeholder for instance deletion", func() {
			logInstance.Println("Remove the instance via DELETE api using resource id")
		})
	})

	AfterAll(func() {
		// Instance deletion
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
	})
})
