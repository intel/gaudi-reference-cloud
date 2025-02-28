package bm_gaudi3_test

import (
	"errors"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	kube "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis/bmaas"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("Compute BM Validation Operator(Positive flow)", Label("compute", "gaudi3_test"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType       string
		createStatusCode   int
		createRespBody     string
		instanceResourceId string
		deviceName         string
		vmName             string
		vmPayload          string
		isInstanceCreated  = false
		extractedIPs       []string
		ipAddresses        []string
	)

	BeforeAll(func() {
		// load instance details
		vmName = "automation-vm-" + utils.GetRandomString()
		instanceType = utils.GetJsonValue("gaudi3InstanceType")
		vmPayload = utils.GetJsonValue("instancePayload")

		// instance creation
		logInstance.Println("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload,
			vmName, instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
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

	When("Get BMH device by resource ID", func() {
		It("Get BMH device by resource ID", func() {
			logInstance.Println("Starting the BMH Device Retrieval via Kube...")
			bmhResponse, err := kube.GetBmhByConsumer(instanceResourceId, kubeConfigPath)
			Expect(err).Error().ShouldNot(HaveOccurred())
			Expect(bmhResponse.Spec.ConsumerRef.Name).To(Equal(instanceResourceId))
			deviceName = bmhResponse.ObjectMeta.Name
		})
	})

	When("Validate BMH device is provisoned", func() {
		It("Validate BMH device is provisoned", func() {
			logInstance.Println("Starting the BMH Device Validation via Kube...")
			succeded, err := kube.CheckBMHState(deviceName, "provisioned", 1800, kubeConfigPath)
			Expect(err).Error().ShouldNot(HaveOccurred(), "Timeout reached; "+"unable to reach state within expected time")
			Expect(succeded).To(Equal(true))
		})
	})

	When("Fetch Annotations IP addresses from BMH", func() {
		It("Fetch Annotations IP addresses from BMH", func() {
			logInstance.Println("Fetch Annotations IP addresses from BMH")
			bmh, err := kube.GetBmhByName(deviceName, kubeConfigPath)
			Expect(err).Error().ShouldNot(HaveOccurred(), "Failed to reach BMH")
			getAnnotations := bmh.Annotations

			// Prefix to match
			prefix := "gpu.ip.cloud.intel.com"

			for key, value := range getAnnotations {
				// Check if the key starts with the desired prefix
				if strings.HasPrefix(key, prefix) {
					ipAddresses = append(ipAddresses, value)
				}
			}

			logInstance.Printf("IP Addresses matching: %s: %v", prefix, ipAddresses)
		})
	})

	When("SSH into the BM instance created", func() {
		It("SSH into the BM instance created", func() {
			_, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)

			machineIp, proxyIp, proxyUser, machineUser := utils.ExtractInterfaceDetailsFromResponse(responseBody)

			// SSH into machine and run command: ip a
			sshCommand := []string{"ssh", "-J", proxyUser + "@" + proxyIp, "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
				machineUser + "@" + machineIp, "ip a", ">", "/tmp/output.txt"}
			logInstance.Println("sshcommand: ", sshCommand)
			out, commandErr := utils.RunCommand(sshCommand)
			logInstance.Println("SSH Output: ", out.String())

			// Read output and error files from the remote machine
			ipOutput, err := utils.ReadFileFromSSHEndpoint(proxyUser, proxyIp, machineUser, machineIp, "/tmp/output.txt")
			if err != nil {
				logInstance.Println("Error reading output file: ", err)
			}
			logInstance.Println("Validation output: ", ipOutput)
			Expect(commandErr).Should(Succeed(), "Error running ssh command: ", commandErr)

			// Extract Ips for the output
			extractedIPs, err = extractIPs(ipOutput)
			Expect(err).Error().ShouldNot(HaveOccurred(), "Error extracting IPs for the output: ", err)
			Expect(extractedIPs).NotTo(BeEmpty())
			// Print the extracted IPs
			logInstance.Println("Extracted IPs: ", extractedIPs)
		})
	})

	When("Validate Annotations IPs from BMH are part of instance interfaces", func() {
		It("Validate Annotations IPs from BMH are part of instance interfaces", func() {
			logInstance.Println("Validate that each IP in ipAddresses(from BMH) is in extractIPs(from instance)")
			for _, ip := range ipAddresses {
				Expect(extractedIPs).To(ContainElement(ip), "IP address %s not found in extractIPs", ip)
			}
		})
	})

	AfterAll(func() {
		// delete the instance created
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
	})
})

func extractIPs(ipOutput string) ([]string, error) {
	// Regular expression to match "inet xxx.xxx.xxx.xxx/xx"
	re := regexp.MustCompile(`inet\s+(\d+\.\d+\.\d+\.\d+/\d+)`)

	// Find all matches in the output
	matches := re.FindAllStringSubmatch(ipOutput, -1)

	if len(matches) == 0 {
		return nil, errors.New("no IP addresses found in the provided output")
	}

	// Extract only the IP part
	var ipList []string
	for _, match := range matches {
		ipList = append(ipList, match[1])
	}

	return ipList, nil
}
