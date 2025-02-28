package cloud_init

import (
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("VMaaS cloud init validation", Label("compute", "vmaas_cloud_init"), Ordered, ContinueOnFailure, func() {
	var (
		vmPayload          string
		vmName             string
		instanceResourceId string
		isInstanceCreated  bool
		instanceType       string
		createStatusCode   int
		createRespBody     string
	)

	// Positive Scenarios - Done
	// Negative Scenarios - TODO
	BeforeAll(func() {
		// load instance details to be created
		vmName = "test-cloudinit-vm-" + utils.GetRandomString()
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		vmPayload = utils.GetJsonValue("cloudInitPayload")

		// instance creation
		logInstance.Println("Starting Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload,
			vmName, instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), "assertion failed on response body")
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

	When("SSH into the instance", func() {
		It("SSH into the instance", func() {
			logInstance.Println("SSH into the instance")
			// Get call to retrieve the machine ip and proxies
			_, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)

			machineIp, proxyIp, proxyUser, machineUser := utils.ExtractInterfaceDetailsFromResponse(responseBody)

			scriptPath := "./validation_script.sh"
			// Upload the script to the remote machine
			scpCommand := []string{"scp", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-r", "-oProxyJump=" + proxyUser + "@" + proxyIp, scriptPath,
				machineUser + "@" + machineIp + ":/tmp/validation_script.sh"}
			logInstance.Println("scpCommand: ", scpCommand)
			scpout, err := utils.RunCommand(scpCommand)
			Expect(err).Should(Succeed(), "Error copying script to remote machine")
			logInstance.Println("Scp Output: ", scpout.String())

			// SSH into machine and run validation_script.sh
			sshCommand := []string{"ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-J", proxyUser + "@" + proxyIp,
				machineUser + "@" + machineIp, "sudo", "-i", "sh", "/tmp/validation_script.sh", ">", "/tmp/output.txt"}
			logInstance.Println("sshcommand: ", sshCommand)
			out, commandErr := utils.RunCommand(sshCommand)
			logInstance.Println("SSH Output: ", out.String())

			// Read output and error files from the remote machine
			valOutput, err := utils.ReadFileFromSSHEndpoint(proxyUser, proxyIp, machineUser, machineIp, "/tmp/output.txt")
			if err != nil {
				fmt.Printf("Error reading output file: %v", err)
			}
			logInstance.Println("Validation output: ", valOutput)
			Expect(commandErr).Should(Succeed(), "Error running ssh command: %v", commandErr)
		})
	})

	AfterAll(func() {
		// Instance deletion
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
	})
})
