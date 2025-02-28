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

var _ = Describe("Compute Instance API endpoint(VM positive flow)", Label("compute", "lb", "compute_lb", "lb_patch_uc"), Ordered, ContinueOnFailure, func() {
	var (
		instanceType        string
		createStatusCode    int
		createRespBody      string
		lbCreateStatusCode  int
		lbCreateRespBody    string
		vmInstancePayload   string
		lbInstancePayload   string
		lbPutPayload        string
		vmResourceId        string
		lbResourceId        string
		isVMInstanceCreated = false
		isLBInstanceCreated = false
	)

	BeforeAll(func() {
		// load instance details
		vmName := "automation-vm-" + utils.GetRandomString()
		lbName := "automation-lb-" + utils.GetRandomString()
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		machineImage := utils.GetJsonValue("machineImageToBeUsed")
		vmInstancePayload = utils.GetJsonValue("vmWithLabelsPayload")
		lbInstancePayload = utils.GetJsonValue("lbWithSelectorPayload")
		lbPutPayload = utils.GetJsonValue("lbPutTypePayload")

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
		Expect(isVMInstanceCreated).Should(BeTrue())

		// LB Creation
		logger.Log.Info("LB creation via api...")
		lbCreateStatusCode, lbCreateRespBody = service_apis.CreateLB(lbInstanceEndpoint, token, lbInstancePayload, lbName, cloudAccount, "80", "tcp", vmResourceId, "any")
		Expect(lbCreateStatusCode).To(Equal(200), lbCreateRespBody)
		Expect(strings.Contains(lbCreateRespBody, `"name":"`+lbName+`"`)).To(BeTrue(), lbCreateRespBody)
		lbResourceId = gjson.Get(lbCreateRespBody, "metadata.resourceId").String()

		// Validation
		logger.Log.Info("Checking whether LB instance is in ready state")
		instanceValidation = utils.CheckLBPhase(lbInstanceEndpoint, token, lbResourceId)
		Eventually(instanceValidation, 30*time.Minute, 30*time.Second).Should(BeTrue())
		isLBInstanceCreated = true
		Expect(isLBInstanceCreated).Should(BeTrue())

	})

	JustBeforeEach(func() {
		logger.Log.Info("----------------------------------------------------")
	})

	When("Instance validation - prerequisite", func() {
		It("Validate the instance creation in before all", func() {
			logger.Log.Info("is instance created? " + strconv.FormatBool(isVMInstanceCreated))
			Expect(isVMInstanceCreated).Should(BeTrue())

			logger.Log.Info("is LB created? " + strconv.FormatBool(isLBInstanceCreated))
			Expect(isLBInstanceCreated).Should(BeTrue())
		})
	})

	// update the monitor type in the existing load balancer
	When("Update the monitor type in existing LB", func() {
		It("updation should be successful...", func() {
			logger.Log.Info("Update the monitor type of existing LB")
			payload := lbPutPayload
			payload = enrichLBPutPayload(payload, "80", "http", vmResourceId, "any")
			statusCode, ResponseBody := service_apis.PutLBById(lbInstanceEndpoint, token, lbResourceId, payload)
			Expect(statusCode).To(Equal(200), ResponseBody)
		})
	})

	// update the listener port in the existing load balancer
	When("Update the listener port in existing LB", func() {
		It("updation should be successful...", func() {
			logger.Log.Info("Update the listener port in the existing LB")
			payload := lbPutPayload
			payload = enrichLBPutPayload(payload, "98", "http", vmResourceId, "any")
			statusCode, ResponseBody := service_apis.PutLBById(lbInstanceEndpoint, token, lbResourceId, payload)
			Expect(statusCode).To(Equal(200), ResponseBody)
		})
	})

	// update the source ip in the existing load balancer
	When("Update the source IP in existing LB", func() {
		It("updation should be successful...", func() {
			logger.Log.Info("Update the source-ip type in the existing LB")
			payload := lbPutPayload
			payload = enrichLBPutPayload(payload, "98", "http", vmResourceId, "10.0.0.2")
			statusCode, ResponseBody := service_apis.PutLBById(lbInstanceEndpoint, token, lbResourceId, payload)
			Expect(statusCode).To(Equal(200), ResponseBody)
		})
	})

	AfterAll(func() {
		logger.Log.Info("Remove the LB via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteLBById(lbInstanceEndpoint, token, lbResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logger.Log.Info("Validation of Instance Deletion")
		lbValidation := utils.CheckLBDeletionById(lbInstanceEndpoint, token, lbResourceId)
		Eventually(lbValidation, 30*time.Minute, 30*time.Second).Should(BeTrue())

		// instance deletion using resource id is covered here
		logger.Log.Info("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody = service_apis.DeleteInstanceById(instanceEndpoint, token, vmResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logger.Log.Info("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, vmResourceId)
		Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())

	})
})

func enrichLBPutPayload(payload string, listenerPort string, monitorType string, instanceResourceId string, sourceIp string) string {
	payload = strings.Replace(payload, "<<listener-port>>", listenerPort, 1)
	payload = strings.Replace(payload, "<<monitor-type>>", monitorType, -1)
	payload = strings.Replace(payload, "<<instance-resource-id>>", instanceResourceId, -1)
	payload = strings.Replace(payload, "<<source-ip>>", sourceIp, 1)
	return payload
}
