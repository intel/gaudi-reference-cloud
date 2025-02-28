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

var _ = Describe("BMaaS Multi SSH keys test", Label("compute", "compute_business_uc", "bm_multissh"), Ordered, func() {
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

		// Instance creation
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload, bmName,
			instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+bmName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
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

	Context("SSH into the BM instance", func() {
		var getStatusCode int
		var getResponseBody string

		When("SSH into the BM instance with first key", func() {
			It("SSH into the BM instance with first key", func() {
				logInstance.Println("SSH into the BM instance with first key")
				getStatusCode, getResponseBody = service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
				Expect(getStatusCode).To(Equal(200), getResponseBody)

				err := utils.SSHIntoInstance(getResponseBody, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
					"../../../../ansible-files/ssh-and-apt-get-on-bm.yml", "~/.ssh/id_rsa")
				Expect(err).NotTo(HaveOccurred(), err)
			})
		})

		When("Add second key to the authorized_keys and SSH into the instance with second key", func() {
			It("Add second key to the authorized_keys and SSH into the instance with second key", func() {
				logInstance.Println("Add second key to the authorized_keys and SSH into the instance with second key")

				machineIp, proxyIp, proxyUser, machineUser := utils.ExtractInterfaceDetailsFromResponse(getResponseBody)

				// Upload the script to the remote machine
				scpCommand := []string{"scp", "-J", proxyUser + "@" + proxyIp, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
					secondPubKeyPath, machineUser + "@" + machineIp + ":/tmp/testkey.pub"}
				logInstance.Println("scpCommand: ", scpCommand)
				scpOutput, err := utils.RunCommand(scpCommand)
				Expect(err).Should(Succeed(), "Error copying script to remote machine")
				logInstance.Println("Scp Output: ", scpOutput.String())

				// Append the contents of the temporary file to the authorized_keys file on the remote machine
				appendCommand := []string{"ssh", "-J", proxyUser + "@" + proxyIp, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
					machineUser + "@" + machineIp, "cat /tmp/testkey.pub >> ~/.ssh/authorized_keys"}
				copyout, commandErr := utils.RunCommand(appendCommand)
				Expect(commandErr).Should(Succeed(), "Error running ssh command")
				logInstance.Println("SSH Output: ", copyout.String())

				err = utils.SSHIntoInstance(getResponseBody, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
					"../../../../ansible-files/ssh-and-apt-get-on-bm.yml", "~/.ssh/testkey")
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
