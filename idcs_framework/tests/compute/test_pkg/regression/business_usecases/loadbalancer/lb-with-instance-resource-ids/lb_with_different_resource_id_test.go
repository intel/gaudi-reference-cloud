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

var _ = Describe("Compute Instance API endpoint(VM positive flow)", Label("compute", "lb", "compute_lb", "vmResourceIdLists"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType        string
		vmInstancePayload   string
		lbInstancePayload   string
		vmResourceIdsList   []string
		vmResourceIdList    string
		isVMInstanceCreated = false
	)

	BeforeAll(func() {
		// load instance details to be created
		vmName := "automation-vm-" + utils.GetRandomString()
		secondVMName := "automation-vm-" + utils.GetRandomString()
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		machineImage := utils.GetJsonValue("machineImageToBeUsed")
		vmInstancePayload = utils.GetJsonValue("vmInstancePayload")
		lbInstancePayload = utils.GetJsonValue("lbInstancePayload")

		// first instance creation
		logger.Log.Info("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createRespBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, vmInstancePayload, vmName,
			instanceType, sshkeyName, vnet, machineImage, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
		vmResourceId := gjson.Get(createRespBody, "metadata.resourceId").String()
		vmResourceIdsList = append(vmResourceIdsList, vmResourceId)

		// Validation
		logger.Log.Info("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, vmResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())

		// second instance creation
		logger.Log.Info("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, vmInstancePayload, secondVMName,
			instanceType, sshkeyName, vnet, machineImage, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+secondVMName+`"`)).To(BeTrue(), createRespBody)
		vmRecourceId2 := gjson.Get(createRespBody, "metadata.resourceId").String()
		vmResourceIdsList = append(vmResourceIdsList, vmRecourceId2)

		// Validation
		logger.Log.Info("Checking whether instance is in ready state")
		instanceValidation2 := utils.CheckInstancePhase(instanceEndpoint, token, vmRecourceId2)
		Eventually(instanceValidation2, 5*time.Minute, 5*time.Second).Should(BeTrue())
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

	When("LB creation by associating more than one instances", func() {
		It("creation should be successful...", func() {
			logger.Log.Info("LB creation via api...")
			lbName := "automation-lb-" + utils.GetRandomString()
			vmResourceIds := strings.Join(vmResourceIdsList, "\", \"")
			createStatusCode, createRespBody := service_apis.CreateLB(lbInstanceEndpoint,
				token, lbInstancePayload, lbName, cloudAccount, "80", "tcp", vmResourceIds, "any")
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+lbName+`"`)).To(BeTrue(), createRespBody)
			vmResourceIdList = gjson.Get(createRespBody, "metadata.resourceId").String()

			// Validation
			logger.Log.Info("Checking whether LB instance is in ready state")
			instanceValidation := utils.CheckLBPhase(lbInstanceEndpoint, token, vmResourceIdList)
			Eventually(instanceValidation, 30*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	AfterAll(func() {
		deleteStatusCode, deleteRespBody := service_apis.DeleteLBById(lbInstanceEndpoint, token, vmResourceIdList)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logger.Log.Info("Validation of Instance Deletion")
		instanceValidation := utils.CheckLBDeletionById(lbInstanceEndpoint, token, vmResourceIdList)
		Eventually(instanceValidation, 30*time.Minute, 5*time.Second).Should(BeTrue())

		// instance deletion using resource id
		logger.Log.Info("Remove the instance via DELETE api using resource id")
		for _, eachInstance := range vmResourceIdsList {
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, eachInstance)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logger.Log.Info("Validation of Instance Deletion using Id")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, eachInstance)
			Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
		}

	})
})
