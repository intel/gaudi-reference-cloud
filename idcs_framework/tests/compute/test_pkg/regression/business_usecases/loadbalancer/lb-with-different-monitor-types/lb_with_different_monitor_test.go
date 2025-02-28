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

var _ = Describe("Compute Instance API endpoint(VM positive flow)", Label("compute", "lb", "compute_lb", "lb_monitor_type"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType        string
		createStatusCode    int
		createRespBody      string
		vmInstancePayload   string
		lbInstancePayload   string
		vmResourceId        string
		lbResourceIds       []string
		vmName              string
		isVMInstanceCreated = false
	)

	BeforeAll(func() {
		// load instance details
		vmName = "automation-vm-" + utils.GetRandomString()
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		machineImage := utils.GetJsonValue("machineImageToBeUsed")
		vmInstancePayload = utils.GetJsonValue("vmInstancePayload")

		// instance creation
		logger.Log.Info("Starting the Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, vmInstancePayload, vmName,
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

	When("Instance validation - prerequisite", func() {
		It("Validate the instance creation in before all", func() {
			logger.Log.Info("is instance created? " + strconv.FormatBool(isVMInstanceCreated))
			Expect(isVMInstanceCreated).Should(BeTrue(), "Instance creation failed with following error "+createRespBody)
		})
	})

	When("Attempt to create LB with different monitor types", func() {
		It("Attempt to create instances with different monitor types", func() {
			defer GinkgoRecover()

			// Define channel to signal when all lb are created
			allLBCreated := make(chan string, len(monitorTypes))
			lbInstancePayload = utils.GetJsonValue("lbInstancePayload")

			for _, eachMonitorType := range monitorTypes {
				logger.Log.Info("Creating the instance with image: " + eachMonitorType.String())
				go utils.CreateLBWithMonitorTypes(lbInstanceEndpoint, token, lbInstancePayload, cloudAccount, "80", eachMonitorType.String(), vmResourceId, "any", allLBCreated, 45*time.Minute)
			}

			// Wait for all instances to be created
			for range monitorTypes {
				resource_id := <-allLBCreated
				lbResourceIds = append(lbResourceIds, resource_id)
			}
		})
	})

	AfterAll(func() {
		// Define channel to signal when all lb are deleted
		allLBsDeleted := make(chan struct{}, len(monitorTypes))

		// Launch goroutines to delete lb concurrently
		for _, id := range lbResourceIds {
			go utils.DeleteAllInstances(instanceEndpoint, token, id, allLBsDeleted)
		}

		// Wait for all lb instances to be deleted
		for range monitorTypes {
			<-allLBsDeleted
		}

		// instance deletion using resource id is covered here
		logger.Log.Info("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, vmResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logger.Log.Info("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, vmResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
	})

})
