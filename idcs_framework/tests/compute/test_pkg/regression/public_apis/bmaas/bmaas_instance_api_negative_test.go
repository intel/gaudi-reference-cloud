package bmaas

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

var _ = Describe("Compute instance api(BM negative flow)", Label("compute", "bmaas", "bmaas_instance", "bmaas_instance_negative"), Ordered, ContinueOnFailure, func() {
	var (
		createStatusCode   int
		createRespBody     string
		bmName             string
		instanceType       string
		bmPayload          string
		instancePutPayload string
		isInstanceCreated  bool
		instanceResourceId string
	)

	BeforeAll(func() {
		if skipBMCreation == "true" {
			Skip("Skipping the Entire BM negative flow due to the flag")
		}

		// load instance details to be created
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		instancePutPayload = utils.GetJsonValue("instancePutPayload")
		bmName = "automation-bm-" + utils.GetRandomString()
		bmPayload = utils.GetJsonValue("instancePayload")

		logInstance.Println("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload, bmName,
			instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+bmName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()
		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 30*time.Minute, 30*time.Second).Should(BeTrue())
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
			statusCode, responseBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload,
				"bm-wo-instancetype-invalid", "invalid name", sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(statusCode).To(Equal(403), responseBody)
		})
	})

	When("Creating an instance using invalid machine image", func() {
		It("should fail with client error", func() {
			logInstance.Println("Attempt to create an instance with invalid machine image") // Despite using non existing image api is able to create the instance
			statusCode, responseBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, bmPayload,
				"bm-wo-machineimage-test-negative", "bm-spr", sshkeyName, vnet, "invalid", availabilityZone)
			Expect(statusCode).To(Equal(400), responseBody)
		})
	})

	When("Creating an instance using invalid SSH public key", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an instance with invalid SSH public key")
			bmNameNegative := "automation-bm-invalid-" + utils.GetRandomString()
			statusCode, responseBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload, bmNameNegative,
				instanceType, "invalid-ssh", vnet, machineImageMapping, availabilityZone)
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
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
		})
	})

	When("Updating an instance with invalid SSH key using resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to Update an instance with invalid field in payload via resource id")
			payload := instancePutPayload
			payload = strings.Replace(payload, "<<ssh-public-key>>", "invalid-key", 1)
			statusCode, responseBody := service_apis.PutInstanceById(instanceEndpoint, token, instanceResourceId, payload)
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
		})
	})

	When("Updating an instance with invalid SSH key using resource name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to Update an instance with invalid field in payload via resource name")
			payload := instancePutPayload
			payload = strings.Replace(payload, "<<ssh-public-key>>", "invalid-key", 1)
			statusCode, responseBody := service_apis.PutInstanceByName(instanceEndpoint, token, bmName, payload)
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
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
			statusCode, responseBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload,
				"instance-name-to-validate-the-character-length-for-testing-purpose-attempt", instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
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
		// deletion of instance which is deleted
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
	})
})
