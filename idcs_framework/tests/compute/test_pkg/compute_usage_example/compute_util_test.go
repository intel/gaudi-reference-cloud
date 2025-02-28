package compute_usage

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/compute_util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/tidwall/gjson"
)

var _ = Describe("Compute util - test", Label("compute_util_test"), Ordered, ContinueOnFailure, func() {
	var computeInstanceUrl string
	var token string
	var createInstanceResponse string
	var instanceId string

	BeforeAll(func() {
		// instance creation example using compute util
		computeInstanceUrl = "<<env>>/v1/cloudaccounts/<<cloud-account>>/instances"                       // Replace <<env>> with actual compute url
		computeInstanceUrl = strings.Replace(computeInstanceUrl, "<<cloud-account>>", "cloud-account", 1) // Replace cloud-account with Id
		token = "Bearer "                                                                                 // Add token
		instancePayload := `{
			"metadata": {
			"name": "<<instance-name>>"
			},
			"spec": {
			"availabilityZone": "us-staging-1a",
			"instanceType": "vm-spr-sml",
			"machineImage": "ubuntu-2204-jammy-v20240308",
			"runStrategy": "RerunOnFailure",
			"sshPublicKeyNames": [
				"ssh-key"
			],
			"interfaces": [
				{
				"name": "eth0",
				"vNet": "us-staging-1a-default"
				}
			]
			}
		}`
		instanceName := "test-instance-vm-" + utils.GetRandomString()
		instancePayload = strings.Replace(instancePayload, "<<instance-name>>", instanceName, 1)
		response := compute_util.ComputeCreateVMInstance(computeInstanceUrl, token, instancePayload)
		createInstanceResponse = string(response.Body())
		instanceId = gjson.Get(createInstanceResponse, "metadata.resourceId").String()
		logger.Log.Info("Response status code " + strconv.Itoa(response.StatusCode()))
		logger.Log.Info("Response body is " + string(response.Body()))
	})

	When("Get the instance", func() {
		It("Get the instance created above", func() {
			response := compute_util.ComputeGetInstanceById(computeInstanceUrl, token, instanceId)
			logger.Log.Info("Response status code " + strconv.Itoa(response.StatusCode()))
			logger.Log.Info("Response body is " + string(response.Body()))
		})
	})

	AfterAll(func() {
		// instance deletion using resource id is covered here
		response := compute_util.ComputeDeleteVMInstanceById(computeInstanceUrl, token, instanceId)
		logger.Log.Info("Response status code " + strconv.Itoa(response.StatusCode()))
		logger.Log.Info("Response body is " + string(response.Body()))
	})
})
