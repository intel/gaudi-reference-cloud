package bm_sshkeys

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("BMaaS Multi SSH keys test", Label("compute", "compute_business_uc", "bm_generic_flow"), Ordered, func() {
	var (
		createRespBody     string
		createStatusCode   int
		instanceType       string
		instanceResourceId string
		bmName             string
		bmPayload          string
		secondSSHKeyName   string
	)

	BeforeAll(func() {
		// name and payload creation
		bmName = "automation-bm-" + utils.GetRandomString()
		secondSSHKeyName = "automation-sshkey-" + utils.GetRandomString()
		bmPayload = utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreatedMultiSSH")
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Creation of second SSH key", func() {
		It("Creation of second SSH key", func() {
			logInstance.Println("Starting second SSH key creation")
			sshPublicKeyPayload := utils.GetJsonValue("sshPublicKeyPayload")
			createStatusCode, createRespBody := service_apis.CreateSSHKey(sshEndpoint, token, sshPublicKeyPayload,
				secondSSHKeyName, secondPublicKey)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+secondSSHKeyName+`"`)).To(BeTrue(), createRespBody)
		})
	})

	When("Instance Creation with two SSH keys", func() {
		It("Instance Creation with two SSH keys", func() {
			logInstance.Println("Instance Creation with two SSH keys")
			allSSHKeys := []string{sshkeyName, secondSSHKeyName}
			logInstance.Println("allSSHKeys: ", allSSHKeys)
			createStatusCode, createRespBody = utils.InstanceCreationMultiSSHKey(instanceEndpoint, token, bmPayload, bmName,
				instanceType, allSSHKeys, vnet, machineImageMapping, availabilityZone)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+bmName+`"`)).To(BeTrue(), createRespBody)
			instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
			Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
		})
	})

	Context("SSH into the BM instance", func() {
		var getStatusCode int
		var getResponseBody string

		When("SSH into the BM instance created with first key", func() {
			It("SSH into the BM instance created with first key", func() {
				logInstance.Println("SSH into the BM instance created with first key")
				getStatusCode, getResponseBody = service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
				Expect(getStatusCode).To(Equal(200), createRespBody)

				err := utils.SSHIntoInstance(getResponseBody, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
					"../../../../ansible-files/ssh-and-apt-get-on-bm.yml", "~/.ssh/id_rsa")
				Expect(err).NotTo(HaveOccurred(), err)
			})
		})

		When("SSH into the BM instance created with second key", func() {
			It("SSH into the BM instance created with second key", func() {
				logInstance.Println("SSH into the BM instance created with second key")
				err := utils.SSHIntoInstance(getResponseBody, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
					"../../../../ansible-files/ssh-and-apt-get-on-vm.yml", "~/.ssh/testkey")
				Expect(err).NotTo(HaveOccurred(), err)
			})
		})
	})

	When("Remove the instance via DELETE api using resource id", func() {
		It("should be successful", func() {
			logInstance.Println("Remove the instance via DELETE api using resource id")
		})
	})

	AfterAll(func() {
		// Delete the instance
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())

		// ssh keys deletion
		deleteStatusCode, deleteRespBody = service_apis.DeleteSSHKeyByName(sshEndpoint, token, secondSSHKeyName)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)
	})
})
