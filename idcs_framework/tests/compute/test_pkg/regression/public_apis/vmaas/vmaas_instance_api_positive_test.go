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

var _ = Describe("Compute Instance API endpoint(VM positive flow)", Label("compute", "vmaas", "vmaas_instance", "vmaas_instance_positive"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType          string
		createRespBody        string
		createStatusCode      int
		vmName                string
		vmPayload             string
		instancePutPayload    string
		secondSSHKeyName      string
		instanceResourceId    string
		secondResourceId      string
		isInstanceCreated     = false
		secondInstanceDeleted = false
	)

	BeforeAll(func() {
		// load instance details to be created
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		instancePutPayload = utils.GetJsonValue("instancePutPayload")
		vmPayload = utils.GetJsonValue("instancePayload")
		vmName = "automation-vm-" + utils.GetRandomString()
		sshPublicKeyPayload := utils.GetJsonValue("sshPublicKeyPayload")
		secondSSHKeyName = "automation-sshkey-" + utils.GetRandomString()

		// Load the public key
		sshKeyValue, err := utils.ReadPublicKey(sshPublicKey)
		Expect(err).Should(Succeed(), "Couldn't read the SSH public key from the specified path."+sshPublicKey)

		// second ssh key for instance updation
		logInstance.Println("Fetching SSH key from the given path...")
		sshStatusCode, sshResponseBody := service_apis.CreateSSHKey(sshEndpoint, token, sshPublicKeyPayload, secondSSHKeyName, sshKeyValue)
		Expect(sshStatusCode).To(Equal(200), sshResponseBody)
		Expect(strings.Contains(sshResponseBody, `"name":"`+secondSSHKeyName+`"`)).To(BeTrue(), sshResponseBody)

		// instance creation
		logInstance.Println("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, vmName,
			instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Phase validation
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

	When("Retrieving an instance created using resource id", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance via GET method using id")
			statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Retrieving an instance created using resource name", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance via GET method using name")
			statusCode, responseBody := service_apis.GetInstanceByName(instanceEndpoint, token, vmName)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Updating already created instance using resource id", func() {
		It("should be successful", func() {
			logInstance.Println("Modify the instance via PUT method using resource id")
			putPayload := instancePutPayload
			putPayload = strings.Replace(putPayload, "<<ssh-public-key>>", secondSSHKeyName, 1)
			statusCode, responseBody := service_apis.PutInstanceById(instanceEndpoint, token, instanceResourceId, putPayload)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `{}`)).To(BeTrue(), responseBody)
		})
	})

	When("Updating already created instance using resource name", func() {
		It("should be successful", func() {
			logInstance.Println("Modify the instance via PUT method using resource name")
			putPayload := instancePutPayload
			putPayload = strings.Replace(putPayload, "<<ssh-public-key>>", sshkeyName, 1)
			statusCode, responseBody := service_apis.PutInstanceByName(instanceEndpoint, token, vmName, putPayload)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `{}`)).To(BeTrue(), responseBody)
		})
	})

	When("Listing all the instance's created in CloudAccount", func() {
		It("should be successful", func() {
			logInstance.Println("List all the instances via GET api")
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, nil)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Retrieving already created instance", func() {
		It("validation of creation timestamp should be successful", func() {
			logInstance.Println("Retrieve the instance via GET method using id")
			statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)

			// validate the instance creation timestamp is not null
			creationTimestampResponse := gjson.Get(createRespBody, "metadata.creationTimestamp").String()
			Expect(creationTimestampResponse).To(Not(Equal("null")), "Creation time stamp shouldn't be null")

			// validate the instance creation timestamp
			creationTimestampUnix := utils.GetUnixTime(creationTimestampResponse)
			Expect(utils.ValidateTimeStamp(creationTimestampUnix, creationTimestampUnix-300000, creationTimestampUnix+300000)).Should(BeTrue())
		})
	})

	When("Retrieving already created instance", func() {
		It("validation of username field should be successful", func() {
			logInstance.Println("Retrieve the instance via GET method using id")
			statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)

			// validate the instance's username is not null
			instanceUserName := gjson.Get(responseBody, "status.userName").String()
			Expect(instanceUserName).To(Not(Equal("")), responseBody)
		})
	})

	When("Retrieving already created instance", func() {
		It("validation of DNS field should be successful", func() {
			logInstance.Println("Retrieve the instance via GET method using id")
			statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)

			// validate the instance's DNS name contains idcservice.net
			dnsName := gjson.Get(responseBody, "status.interfaces.0.dnsName").String()
			Expect(strings.Contains(dnsName, "idcservice.net")).To(BeTrue(), "dns name is not in the format of instanceName.cloudAccount.region.idcservice.net")
		})
	})

	When("Remove an instance using resource name", func() {
		It("Creation and deletion of an instance should be successful", func() {
			logInstance.Println("Remove the instance via DELETE api using resource name")
			secondVMName := "automation-vm-" + utils.GetRandomString()
			statusCode, responseBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, secondVMName, "vm-spr-sml", secondSSHKeyName,
				vnet, machineImageMapping, availabilityZone)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `"name":"`+secondVMName+`"`)).To(BeTrue(), responseBody)
			secondResourceId = gjson.Get(responseBody, "metadata.resourceId").String()

			// Phase validation
			logInstance.Println("Checking whether instance is in ready state")
			phaseValidation := utils.CheckInstancePhase(instanceEndpoint, token, secondResourceId)
			Eventually(phaseValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())

			// delete the instance
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceByName(instanceEndpoint, token, secondVMName)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			logInstance.Println("Validation of Instance Deletion using Name")
			instanceValidation := utils.CheckInstanceDeletionByName(instanceEndpoint, token, secondVMName)
			Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
			secondInstanceDeleted = true
		})
	})

	When("Remove the instance via DELETE api using resource id", func() {
		It("Instance and SSH deletion should be successful", func() {
			logInstance.Println("Remove the instance and SSH key via DELETE api using resource id")
		})
	})

	AfterAll(func() {
		// instance deletion using resource id is covered here
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())

		// ssh keys deletion used in test case (name)
		statusCode, responseBody := service_apis.DeleteSSHKeyByName(sshEndpoint, token, secondSSHKeyName)
		Expect(statusCode).To(Equal(200), responseBody)

		if !secondInstanceDeleted {
			logInstance.Println("Remove the second instance via DELETE api using resource id")
			statusCode, responseBody := service_apis.DeleteInstanceById(instanceEndpoint, token, secondResourceId)
			Expect(statusCode).To(Equal(200), responseBody)

			logInstance.Println("Validation of Instance Deletion using Id")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, secondResourceId)
			Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
		}
	})
})
