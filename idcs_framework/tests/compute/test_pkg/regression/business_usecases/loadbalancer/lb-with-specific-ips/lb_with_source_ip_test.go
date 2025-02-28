package loadbalancer

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("Compute Instance API endpoint(VM positive flow)", Label("compute", "lb", "compute_lb", "lb_specific_ips"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType        string
		vmInstancePayload   string
		lbInstancePayload   string
		vmResourceId        string
		lbResourceId        string
		isVMInstanceCreated = false
	)

	BeforeAll(func() {
		// load instance details to be created
		vmName := "automation-vm-" + utils.GetRandomString()
		vmInstancePayload = utils.GetJsonValue("vmInstancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		machineImage := utils.GetJsonValue("machineImageToBeUsed")
		lbInstancePayload = utils.GetJsonValue("lbInstancePayload")

		// instance1 creation
		logger.Log.Info("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createRespBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, vmInstancePayload, vmName,
			instanceType, sshkeyName, vnet, machineImage, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
		vmResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Validation
		logger.Log.Info("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, vmResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
		isVMInstanceCreated = true
	})

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Instance validation - prerequisite", func() {
		It("Validate the instance creation in before all", func() {
			logger.Log.Info("is instance created? " + strconv.FormatBool(isVMInstanceCreated))
			Expect(isVMInstanceCreated).Should(BeTrue())
		})
	})

	When("LB creation with specific source IP", func() {
		It("creation should be successful...", func() {
			logger.Log.Info("LB creation via api...")
			lbName := "automation-lb-" + utils.GetRandomString()
			createStatusCode, createRespBody := service_apis.CreateLB(lbInstanceEndpoint, token, lbInstancePayload, lbName, cloudAccount, "80", "tcp", vmResourceId, "10.0.0.1")
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+lbName+`"`)).To(BeTrue(), createRespBody)
			lbResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

			// Validation
			logger.Log.Info("Checking whether LB instance is in ready state")
			instanceValidation := utils.CheckLBPhase(lbInstanceEndpoint, token, lbResourceId)
			Eventually(instanceValidation, 30*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	AfterAll(func() {
		// LB Deletion
		deleteStatusCode, deleteRespBody := service_apis.DeleteLBById(lbInstanceEndpoint, token, lbResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logger.Log.Info("Validation of Instance Deletion")
		lbValidation := utils.CheckLBDeletionById(lbInstanceEndpoint, token, lbResourceId)
		Eventually(lbValidation, 30*time.Minute, 5*time.Second).Should(BeTrue())

		// instance deletion using resource id
		logger.Log.Info("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody = service_apis.DeleteInstanceById(instanceEndpoint, token, vmResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		logger.Log.Info("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, vmResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())

	})
})
