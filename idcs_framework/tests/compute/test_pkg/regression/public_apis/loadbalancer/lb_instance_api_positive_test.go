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

var _ = Describe("Compute Instance API endpoint(LB positive flow)", Label("compute", "lb", "compute_lb", "lb_instance_positive"), Ordered, ContinueOnFailure, func() {
	var (
		vmName              string
		lbName              string
		instanceType        string
		createResponseBody  string
		createStatusCode    int
		lbCreateStatusCode  int
		lbCreateRespBody    string
		instanceResourceId  string
		lbResourceId        string
		lbInstancePayload   string
		vmInstancePayload   string
		lbPutPayload        string
		isVMInstanceCreated = false
	)

	BeforeAll(func() {
		// load instance details to be created
		vmName = "automation-vm-" + utils.GetRandomString()
		vmInstancePayload = utils.GetJsonValue("vmInstancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		machineImage := utils.GetJsonValue("machineImageToBeUsed")
		lbInstancePayload = utils.GetJsonValue("lbInstancePayload")
		lbPutPayload = utils.GetJsonValue("lbPutPayload")

		// instance creation
		logInstance.Println("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createResponseBody = service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, vmInstancePayload, vmName,
			instanceType, sshkeyName, vnet, machineImage, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createResponseBody)
		Expect(strings.Contains(createResponseBody, `"name":"`+vmName+`"`)).To(BeTrue(), createResponseBody)
		instanceResourceId = gjson.Get(createResponseBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
		isVMInstanceCreated = true
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Instance validation - prerequisite", func() {
		It("Validate the instance creation in before all", func() {
			logInstance.Println("is instance created? " + strconv.FormatBool(isVMInstanceCreated))
			Expect(isVMInstanceCreated).Should(BeTrue(), "Instance creation failed with following error "+createResponseBody)
		})
	})

	When("LB creation and its validation - prerequisite", func() {
		It("Validate the instance creation in before all", func() {
			logInstance.Println("LB creation via api...")
			lbName = "automation-lb-" + utils.GetRandomString()
			lbCreateStatusCode, lbCreateRespBody = service_apis.CreateLB(lbInstanceEndpoint, token, lbInstancePayload, lbName, cloudAccount, "80", "TCP", instanceResourceId, "any")
			Expect(lbCreateStatusCode).To(Equal(200), lbCreateRespBody)
			Expect(strings.Contains(lbCreateRespBody, `"name":"`+lbName+`"`)).To(BeTrue(), lbCreateRespBody)
			lbResourceId = gjson.Get(lbCreateRespBody, "metadata.resourceId").String()

			// Validation
			logInstance.Println("Checking whether LB instance is in ready state")
			instanceValidation := utils.CheckLBPhase(lbInstanceEndpoint, token, lbResourceId)
			Eventually(instanceValidation, 30*time.Minute, 5*time.Second).Should(BeTrue())
			isLBInstanceCreated := true
			Expect(isLBInstanceCreated).Should(BeTrue(), "Instance creation failed with following error "+lbCreateRespBody)
		})
	})

	When("Retrieving an LB created using resource id", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the LB via GET method using id")
			statusCode, responseBody := service_apis.GetLBById(lbInstanceEndpoint, token, lbResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Retrieving an instance created using resource name", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the LB via GET method using name")
			statusCode, responseBody := service_apis.GetLBByName(lbInstanceEndpoint, token, lbName)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Updating already created instance using resource id", func() {
		It("should be successful", func() {
			payload := lbPutPayload
			payload = strings.Replace(payload, "<<instance-resource-id>>", instanceResourceId, 1)
			payload = strings.Replace(payload, "<<existing-port>>", "90", 1)
			payload = strings.Replace(payload, "<<cloud-account>>", cloudAccount, 1)
			statusCode, responseBody := service_apis.PutLBById(lbInstanceEndpoint, token, lbResourceId, payload)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Updating already created instance using resource name", func() {
		It("should be successful", func() {
			payload := lbPutPayload
			payload = strings.Replace(payload, "<<instance-resource-id>>", instanceResourceId, 1)
			payload = strings.Replace(payload, "<<existing-port>>", "94", 1)
			payload = strings.Replace(payload, "<<cloud-account>>", cloudAccount, 1)
			statusCode, responseBody := service_apis.PutLBByName(lbInstanceEndpoint, token, lbName, payload)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Listing all the LB instance's created in CA", func() {
		It("should be successful", func() {
			logInstance.Println("List all the LB instances via GET api")
			statusCode, responseBody := service_apis.GetAllLB(lbInstanceEndpoint, token, nil)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	// TODO: validation of LB creation time (metering)
	/*When("Retrieving already created instance", func() {
	    It("validation of creation timestamp should be successful", func() {
	        logInstance.Println("Retrieve the instance via GET method using id")
	        statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
	        Expect(statusCode).To(Equal(200), responseBody)

	        // validate the instance creation timestamp is not null
	        creationTimeStampFromResponse := gjson.Get(createResponseBody, "metadata.creationTimestamp").String()
	        Expect(creationTimeStampFromResponse).To(Not(Equal("null")), "Creation time stamp shouldn't be null: %v", creationTimeStampFromResponse)

			// validate the instance creation timestamp
	        creationTimeStampFromResponseUnix := utils.GetUnixTime(creationTimeStampFromResponse)
	        Expect(utils.ValidateTimeStamp(creationTimeStampFromResponseUnix, instance_creation_timestamp-300000, creationTimeStampFromResponseUnix+300000)).Should(BeTrue())
	    })
	})*/

	AfterAll(func() {
		// LB deletion and Validation
		logInstance.Println("Remove the LB via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteLBById(lbInstanceEndpoint, token, lbResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		logInstance.Println("Validation of LB Deletion")
		lbValidation := utils.CheckLBDeletionById(lbInstanceEndpoint, token, lbResourceId)
		Eventually(lbValidation, 30*time.Minute, 5*time.Second).Should(BeTrue())

		// Instance deletion and Validation
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteInstanceStatusCode, deleteInstanceRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteInstanceStatusCode).To(Equal(200), deleteInstanceRespBody)

		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())

	})
})
