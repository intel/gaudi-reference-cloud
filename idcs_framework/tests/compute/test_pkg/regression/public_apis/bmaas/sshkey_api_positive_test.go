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

var _ = Describe("Compute SSH public key endpoint(BM positive flow)", Label("compute", "bmaas", "compute_ssh_key", "bmaas_ssh_key", "bmaas_ssh_key_positive"), Ordered, ContinueOnFailure, func() {
	var (
		createStatusCode    int
		createRespBody      string
		sshPublicKeyPayload string
		sshKeyValue         string
		sshResourceId       string
		sshKeyName          string
		err                 error
	)

	BeforeAll(func() {
		// Load SSH key details
		sshPublicKeyPayload = utils.GetJsonValue("sshPublicKeyPayload")
		sshKeyName = "automation-sshkey-" + utils.GetRandomString()

		// Read publickey
		logInstance.Println("Fetching SSH key from the given path...")
		sshKeyValue, err = utils.ReadPublicKey(sshPublicKey)
		Expect(err).Should(Succeed(), "Couldn't read the SSH public key from the specified path."+sshPublicKey)

		// Create SSH key
		logInstance.Println("Starting the SSH-Public-Key Creation flow via API - before all...")
		createStatusCode, createRespBody = service_apis.CreateSSHKey(sshEndpoint, token, sshPublicKeyPayload, sshKeyName, sshKeyValue)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+sshKeyName+`"`)).To(BeTrue(), createRespBody)
		sshResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Creating an SSH key without name", func() {
		It("should be successful", func() {
			logInstance.Println("Starting the SSH-Public-Key Creation flow via API...")
			statusCode, responseBody := service_apis.CreateSSHKey(sshEndpoint, token, sshPublicKeyPayload, "", sshKeyValue)
			time.Sleep(10 * time.Second)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `"resourceId"`)).To(BeTrue(), responseBody)
			sshNameCreated := gjson.Get(responseBody, "metadata.name").String()

			// Deletion of ssh public key via name
			deleteStatusCode, deleteRespBody := service_apis.DeleteSSHKeyByName(sshEndpoint, token, sshNameCreated)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of SSH Key Deletion using Name")
			keyValidation := utils.CheckSSHDeletionByName(sshEndpoint, token, responseBody)
			Eventually(keyValidation, 1*time.Minute, 5*time.Second).Should(BeTrue())
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
			creationTimestampResponse := gjson.Get(createRespBody, "metadata.creationTimestamp").String()
			Expect(creationTimestampResponse).To(Not(Equal("null")), "Creation time stamp shouldn't be null: %v", creationTimestampResponse)
		})
	})

	When("Retrieving already created SSH key", func() {
		It("validation of creation timestamp should be successful", func() {
			logInstance.Println("Retrieve the SSH public key via GET method using id")
			statusCode, responseBody := service_apis.GetSSHKeyById(sshEndpoint, token, sshResourceId)
			Expect(statusCode).To(Equal(200), responseBody)

			// validate the ssh-key creation timestamp is not null
			creationTimestampResponse := gjson.Get(createRespBody, "metadata.creationTimestamp").String()
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
			statusCode, responseBody := service_apis.CreateSSHKey(sshEndpoint, token, sshPublicKeyPayload, secondSSHName, sshKeyValue)
			time.Sleep(10 * time.Second)
			Expect(statusCode).To(Equal(200), responseBody)

			// Deletion of ssh public key via name
			deleteStatusCode, deleteRespBody := service_apis.DeleteSSHKeyByName(sshEndpoint, token, secondSSHName)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of SSH Key Deletion using Name")
			keyValidation := utils.CheckSSHDeletionByName(sshEndpoint, token, secondSSHName)
			Eventually(keyValidation, 1*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	AfterAll(func() {
		logInstance.Println("Remove the ssh-public-key via Delete method using id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteSSHKeyById(sshEndpoint, token, sshResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of SSH Key Deletion using Name")
		keyValidation := utils.CheckSSHDeletionByName(sshEndpoint, token, sshKeyName)
		Eventually(keyValidation, 1*time.Minute, 5*time.Second).Should(BeTrue())
	})
})
