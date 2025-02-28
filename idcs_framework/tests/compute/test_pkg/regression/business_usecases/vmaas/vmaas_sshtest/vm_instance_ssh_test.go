package sshtest

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("VMaaS SSH Validation", Label("compute", "compute_business_uc", "vm_ssh_uc"), Ordered, func() {
	var (
		createResponseBody string
		createStatusCode   int
		instanceType       string
		instanceResourceId string
		vmName             string
		vmPayload          string
	)

	BeforeAll(func() {
		// name and payload creation
		vmName = "automation-vm-" + utils.GetRandomString()
		vmPayload = utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")

		// Instance creation
		createStatusCode, createResponseBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, vmName, instanceType,
			sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createResponseBody)
		Expect(strings.Contains(createResponseBody, `"name":"`+vmName+`"`)).To(BeTrue(), "assertion failed on response body")
		instanceResourceId = gjson.Get(createResponseBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 10*time.Second).Should(BeTrue())
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("SSH into the created VM instance", func() {
		It("SSH into the created VM instance", func() {
			_, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)

			err := utils.SSHIntoInstance(responseBody, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
				"../../../../ansible-files/ssh-and-apt-get-on-vm.yml", "~/.ssh/id_rsa")
			Expect(err).NotTo(HaveOccurred(), err)
		})
	})

	AfterAll(func() {
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
	})
})
