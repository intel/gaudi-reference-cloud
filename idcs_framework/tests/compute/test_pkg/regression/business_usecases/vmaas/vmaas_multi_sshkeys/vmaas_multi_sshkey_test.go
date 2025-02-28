package vm_sshkeys

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("VMaaS Multi SSH keys test", Label("compute", "compute_business_uc", "vm_multissh"), Ordered, func() {
	var (
		createRespBody     string
		createStatusCode   int
		instanceType       string
		instanceResourceId string
		vmName             string
		vmPayload          string
		secondSSHKeyName   string
	)

	BeforeAll(func() {
		// name and payload creation
		vmName = "automation-vm-" + utils.GetRandomString()
		vmPayload = utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")

		// Instance creation
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, vmName,
			instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 10*time.Second).Should(BeTrue())
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Creation of second SSH key", func() {
		It("Creation of second SSH key", func() {
			// second ssh key for instance updation
			second_sshkey_payload := utils.GetJsonValue("sshPublicKeyPayload")
			secondSSHKeyName = "automation-sshkey-" + utils.GetRandomString()
			statusCode, responseBody := service_apis.CreateSSHKey(sshEndpoint, token, second_sshkey_payload,
				secondSSHKeyName, secondPublicKey)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `"name":"`+secondSSHKeyName+`"`)).To(BeTrue(), responseBody)
		})
	})

	Context("SSH into the VM instance created", func() {
		var getStatusCode int
		var getResponseBody string

		When("SSH into the VM instance created with first key", func() {
			It("SSH into the VM instance created with first key", func() {
				getStatusCode, getResponseBody = service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
				Expect(getStatusCode).To(Equal(200), getResponseBody)

				err := utils.SSHIntoInstance(getResponseBody, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
					"../../../../ansible-files/ssh-and-apt-get-on-vm.yml", "~/.ssh/id_rsa")
				Expect(err).NotTo(HaveOccurred(), err)
			})
		})

		When("Add second key to the authorized_keys and SSH into the instance with second key", func() {
			It("Add second key to the authorized_keys and SSH into the instance with second key", func() {

				machineIp, proxyIp, proxyUser, machineUser := utils.ExtractInterfaceDetailsFromResponse(getResponseBody)

				// Upload the script to the remote machine
				scpCommand := []string{"scp", "-J", proxyUser + "@" + proxyIp, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
					secondPubKeyPath, machineUser + "@" + machineIp + ":/tmp/testkey.pub"}
				logInstance.Println("scpCommand: ", scpCommand)
				scpOutput, err := utils.RunCommand(scpCommand)
				Expect(err).Should(Succeed(), "Error copying script to remote machine: %v", err)
				logInstance.Println("Scp Output: ", scpOutput.String())

				// Append the contents of the temporary file to the authorized_keys file on the remote machine
				appendCommand := []string{"ssh", "-J", proxyUser + "@" + proxyIp, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
					machineUser + "@" + machineIp, "cat /tmp/testkey.pub >> ~/.ssh/authorized_keys"}
				copyout, commandErr := utils.RunCommand(appendCommand)
				Expect(commandErr).Should(Succeed(), "Error running ssh command %v", commandErr)
				logInstance.Println("SSH Output: ", copyout.String())

				// SSH into the instance
				err = utils.SSHIntoInstance(getResponseBody, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
					"../../../../ansible-files/ssh-and-apt-get-on-vm.yml", "~/.ssh/testkey")
				Expect(err).NotTo(HaveOccurred(), err)
			})
		})
	})

	When("Remove the instance via DELETE api", func() {
		It("should be successful", func() {
			logInstance.Println("Remove the instance via DELETE api using resource id")
		})
	})

	AfterAll(func() {
		// Delete the instance created
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())

		// ssh keys deletion used in test case (name)
		statusCode, responseBody := service_apis.DeleteSSHKeyByName(sshEndpoint, token, secondSSHKeyName)
		Expect(statusCode).To(Equal(200), responseBody)
	})
})
