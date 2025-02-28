package loadbalancer

import (
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Compute Instance API endpoint(LB negative flow)", Label("compute", "lb", "compute_instance", "lb_instance", "lb_instance_negative"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType       string
		createStatusCode   int
		createRespBody     string
		vmName             string
		vmPayload          string
		lbPutPayload       string
		lbInstancePayload  string
		isInstanceCreated  bool
		instanceResourceId string
	)

	// TODO: LB creation with duplicate listener port
	/*var lb_creation_status_negative int
	  var lb_creation_body_negative string
	  var lb_name string
	  var lb_payload string
	  var lb_resource_id string*/

	BeforeAll(func() {
		// load instance details to be created
		vmName = "automation-vm-" + utils.GetRandomString()
		vmPayload = utils.GetJsonValue("vmInstancePayload")
		lbInstancePayload = utils.GetJsonValue("lbInstancePayload")
		lbPutPayload = utils.GetJsonValue("lbPutPayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		machineImage := utils.GetJsonValue("machineImageToBeUsed")

		logInstance.Println("Starting the Instance Creation flow via Instance API for negative scenarios")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, vmPayload,
			vmName, instanceType, sshkeyName, vnet, machineImage, availabilityZone)
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

	When("Creating an LB with invalid source id", func() {
		It("should fail with client error", func() {
			logInstance.Println("Attempt to create an LB with invalid source id")
			payload := lbInstancePayload
			payload = strings.Replace(payload, "any", "invalid", 1)
			payload = strings.Replace(payload, "<<lb-name>>", "lb-with-invalid-sourceip", 1)
			payload = strings.Replace(payload, "<<instance-resource-id>>", instanceResourceId, 1)
			payload = strings.Replace(payload, "<<cloud-account>>", cloudAccount, 1)
			statusCode, responseBody := service_apis.CreateLBWithCustomizedPayload(lbInstanceEndpoint, token, payload)
			Expect(statusCode).To(Equal(400), responseBody)
		})
	})

	// TODO: LB creation with duplicate listener port
	/*
	   When("Creating an instance using duplicate listener port", func() {
	       It("should fail with client error", func() {
	           logInstance.Println("Attempt to create an instance with invalid machine image")
	           var payload := vmPayload
	           payload = strings.Replace(payload, "<<lb-name>>", "lb-with-dup-listener-port", 1)
	           payload = strings.Replace(payload, "<<instance-resource-id>>", instanceResourceId, 1)
	           payload = strings.Replace(payload, "<<cloud-account>>", cloudAccount, 1)
	           statusCode, responseBody := service_apis.CreateInstanceWithCustomizedPayload(instanceEndpoint, token, payload)
	           Expect(statusCode).To(Equal(400), responseBody)
	           Expect(strings.Contains(responseBody, `"message":"unable to get machine image`)).To(BeTrue(), responseBody)
	       })
	   })*/

	When("Creating an instance using invalid selector type - invalid resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an LB with invalid selector")
			payload := lbInstancePayload
			payload = strings.Replace(payload, "<<lb-name>>", "lb-with-invalid-selectortype", 1)
			payload = strings.Replace(payload, "<<instance-resource-id>>", "invalid-rsc-id", 1)
			payload = strings.Replace(payload, "<<cloud-account>>", cloudAccount, 1)
			statusCode, responseBody := service_apis.CreateLBWithCustomizedPayload(lbInstanceEndpoint, token, payload)
			Expect(statusCode).To(Equal(400), responseBody)
		})
	})

	When("Retrieving an LB using invalid resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to get an LB with invalid id")
			statusCode, responseBody := service_apis.GetLBById(lbInstanceEndpoint, token, "invalid-id")
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, "invalid resourceId")).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving an LB using invalid resource name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to get an LB with invalid name")
			statusCode, responseBody := service_apis.GetLBByName(lbInstanceEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
		})
	})

	When("Deleting an LB using invalid resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to delete an LB with invalid id")
			statusCode, responseBody := service_apis.DeleteLBById(lbInstanceEndpoint, token, "invalid-id")
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, "invalid resourceId")).To(BeTrue(), responseBody)
		})
	})

	When("Deleting an LB using invalid resource name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to delete an LB with invalid name")
			statusCode, responseBody := service_apis.DeleteLBByName(lbInstanceEndpoint, token, "invalid-name")
			Expect(statusCode).To(Equal(404), responseBody)
			Expect(strings.Contains(responseBody, `"message":"resource not found"`)).To(BeTrue(), responseBody)
		})
	})

	When("Updating an LB with invalid resource id", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to Update an instance with invalid field in payload via resource id")
			payload := lbPutPayload
			payload = strings.Replace(payload, "<<instance-resource-id>>", instanceResourceId, 1)
			payload = strings.Replace(payload, "existingPort", "90", 1)
			payload = strings.Replace(payload, "<<cloud-account>>", cloudAccount, 1)
			statusCode, responseBody := service_apis.PutLBById(lbInstanceEndpoint, token, "lb-invalid-id", payload)
			Expect(statusCode).To(Equal(400), responseBody)
		})
	})

	When("Updating an LB with invalid resource name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to Update an instance with invalid field in payload via resource name")
			payload := lbPutPayload
			payload = strings.Replace(payload, "<<instance-resource-id>>", instanceResourceId, 1)
			payload = strings.Replace(payload, "existingPort", "90", 1)
			payload = strings.Replace(payload, "<<cloud-account>>", cloudAccount, 1)
			statusCode, responseBody := service_apis.PutLBByName(lbInstanceEndpoint, token, "lb-invalid-name", payload)
			Expect(statusCode).To(Equal(400), responseBody)
		})
	})

	When("Creating an LB with too many char's in name", func() {
		It("should fail with valid error...", func() {
			logInstance.Println("Attempt to create an instance with invalid instance name with too many characters")
			payload := lbInstancePayload
			payload = strings.Replace(payload, "<<lb-name>>", "instance-lb-name-to-validate-the-character-length-for-testing-purpose-attempt"+utils.GetRandomString(), 1)
			payload = strings.Replace(payload, "<<instance-resource-id>>", instanceResourceId, 1)
			payload = strings.Replace(payload, "<<cloud-account>>", cloudAccount, 1)
			statusCode, responseBody := service_apis.CreateLBWithCustomizedPayload(lbInstanceEndpoint, token, payload)
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `"message":"invalid load balancer name"`)).To(BeTrue(), responseBody)
		})
	})

	AfterAll(func() {
		// deletion of instance which is deleted
		logInstance.Println("Remove the instance via DELETE api using resource id")
		statusCode, responseBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(statusCode).To(Equal(200), responseBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
	})

})
