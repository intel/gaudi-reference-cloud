package vmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("Compute SSH public key endpoint(VM positive flow)", Label("compute", "vmaas", "vmaas_ssh_key", "vmaas_ssh_key_positive"), Ordered, ContinueOnFailure, func() {
	var (
		sshKeyValue        string
		sshKeyName         string
		sshResourceId      string
		sshKeyPayload      string
		createResponseBody string
		createStatusCode   int
		err                error
	)

	BeforeAll(func() {
		// load ssh key details
		sshKeyPayload = utils.GetJsonValue("sshPublicKeyPayload")
		sshKeyName = "automation-sshkey-" + utils.GetRandomString()

		// Read publickey
		logInstance.Println("Fetching SSH key from the given path...")
		sshKeyValue, err = utils.ReadPublicKey(sshPublicKey)
		Expect(err).Should(Succeed(), "Couldn't read the SSH public key from the specified path: "+sshPublicKey)

		// create ssh key
		logInstance.Println("Starting the SSH-Public-Key Creation flow via API - before all...")
		createStatusCode, createResponseBody = service_apis.CreateSSHKey(sshEndpoint, token, sshKeyPayload, sshKeyName, sshKeyValue)
		Expect(createStatusCode).To(Equal(200), createResponseBody)
		Expect(strings.Contains(createResponseBody, `"name":"`+sshKeyName+`"`)).To(BeTrue(), createResponseBody)
		sshResourceId = gjson.Get(createResponseBody, "metadata.resourceId").String()
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Creating an SSH key without name", func() {
		It("should be successful", func() {
			logInstance.Println("Starting the SSH-Public-Key Creation flow via API...")
			statusCode, responseBody := service_apis.CreateSSHKey(sshEndpoint, token, sshKeyPayload, "", sshKeyValue)
			time.Sleep(10 * time.Second)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `"resourceId"`)).To(BeTrue(), responseBody)

			// Deletion of ssh public key via name
			sshNameCreated := gjson.Get(responseBody, "metadata.name").String()
			deleteStatusCode, deleteResponseBody := service_apis.DeleteSSHKeyByName(sshEndpoint, token, sshNameCreated)
			Expect(deleteStatusCode).To(Equal(200), deleteResponseBody)

			// Validation
			logInstance.Println("Validation of SSH Key Deletion using Name")
			sshKeyValidation := utils.CheckSSHDeletionByName(sshEndpoint, token, sshNameCreated)
			Eventually(sshKeyValidation, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Listing all SSH key's in CA", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve all the ssh-public-key available via GET method")
			statusCode, responseBody := service_apis.GetAllSSHKey(sshEndpoint, token)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Retrieving already created SSH key using id", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the ssh-public-key via GET method using id")
			statusCode, responseBody := service_apis.GetSSHKeyById(sshEndpoint, token, sshResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, sshResourceId)).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving already created SSH key using name", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the ssh-public-key via GET method using name")
			statusCode, responseBody := service_apis.GetSSHKeyByName(sshEndpoint, token, sshKeyName)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, sshKeyName)).To(BeTrue(), responseBody)
		})
	})

	When("Retrieving already created SSH key", func() {
		It("creation timestamp shouldn't be null", func() {
			logInstance.Println("Retrieve the SSH public key via GET method using id")
			statusCode, responseBody := service_apis.GetSSHKeyById(sshEndpoint, token, sshResourceId)
			Expect(statusCode).To(Equal(200), responseBody)

			// validate the ssh-key creation timestamp is not null
			creationTimestampResponse := gjson.Get(createResponseBody, "metadata.creationTimestamp").String()
			Expect(creationTimestampResponse).To(Not(Equal("null")), "Creation time stamp shouldn't be null: %v", creationTimestampResponse)
		})
	})

	When("Retrieving already created SSH key", func() {
		It("validation of creation timestamp should be successful", func() {
			logInstance.Println("Retrieve the SSH public key via GET method using id")
			statusCode, responseBody := service_apis.GetSSHKeyById(sshEndpoint, token, sshResourceId)
			Expect(statusCode).To(Equal(200), responseBody)

			// validate the ssh-key creation timestamp is not null
			creationTimestampResponse := gjson.Get(createResponseBody, "metadata.creationTimestamp").String()
			Expect(creationTimestampResponse).To(Not(Equal("null")), "Creation time stamp shouldn't be null: %v", creationTimestampResponse)

			// validate the ssh-key creation timestamp
			creationTimestampUnix := utils.GetUnixTime(creationTimestampResponse)
			Expect(utils.ValidateTimeStamp(creationTimestampUnix, creationTimestampUnix-6000, creationTimestampUnix+6000)).Should(BeTrue())
		})
	})

	When("Deleting SSH key using valid resource name", func() {
		It("Creation and deletion of SSH key should be successful", func() {
			logInstance.Println("Remove the ssh-public-key via Delete method using name")

			// Creation of ssh public key
			secondSSHName := "automation-sshkey-" + utils.GetRandomString()
			statusCode, responseBody := service_apis.CreateSSHKey(sshEndpoint, token, sshKeyPayload, secondSSHName, sshKeyValue)
			time.Sleep(10 * time.Second)
			Expect(statusCode).To(Equal(200), responseBody)

			// Deletion of ssh public key via name
			deletionStatusCode, deletionRespBody := service_apis.DeleteSSHKeyByName(sshEndpoint, token, secondSSHName)
			Expect(deletionStatusCode).To(Equal(200), deletionRespBody)

			// Retrieve after deletion
			logInstance.Println("Validation of Second SSH Key Deletion using Name")
			keyValidation := utils.CheckSSHDeletionByName(sshEndpoint, token, secondSSHName)
			Eventually(keyValidation, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	AfterAll(func() {
		logInstance.Println("Remove the First SSH public key via Delete method using id")
		statusCode, responseBody := service_apis.DeleteSSHKeyById(sshEndpoint, token, sshResourceId)
		Expect(statusCode).To(Equal(200), responseBody)

		// Retrieve after deletion
		logInstance.Println("Validation of First SSH Key Deletion using Name")
		keyValidation := utils.CheckSSHDeletionById(sshEndpoint, token, sshResourceId)
		Eventually(keyValidation, 1*time.Minute, 5*time.Second).Should(BeTrue())
	})
})
