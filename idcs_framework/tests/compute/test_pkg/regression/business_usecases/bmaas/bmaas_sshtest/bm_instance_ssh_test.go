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

var _ = Describe("BMaaS SSH Validation", Label("compute", "compute_business_uc", "bm_ssh_uc"), Ordered, func() {
	var (
		createResponseBody string
		createStatusCode   int
		instanceType       string
		instanceResourceId string
		bmName             string
		bmPayload          string
	)

	BeforeAll(func() {
		// name and payload creation
		bmName = "automation-bm-" + utils.GetRandomString()
		bmPayload = utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreatedSSHTest")

		// Instance creation
		createStatusCode, createResponseBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload, bmName,
			instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createResponseBody)
		Expect(strings.Contains(createResponseBody, `"name":"`+bmName+`"`)).To(BeTrue(), createResponseBody)
		instanceResourceId = gjson.Get(createResponseBody, "metadata.resourceId").String()
		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("SSH into the BM instance created", func() {
		It("SSH into the BM instance created", func() {
			logInstance.Println("SSH into the BM instance created")
			_, getRespBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
			err := utils.SSHIntoInstance(getRespBody, "../../../../ansible-files", "../../../../ansible-files/inventory.ini",
				"../../../../ansible-files/ssh-and-apt-get-on-bm.yml", "~/.ssh/id_rsa")
			Expect(err).NotTo(HaveOccurred(), err)
		})
	})

	AfterAll(func() {
		// delete the instance created
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
	})
})
