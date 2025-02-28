package nodepool

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("VMaaS Node pool Negative Scenario Validation", Label("compute", "compute_business_uc", "vm_nodepool_negative_uc"), Ordered, func() {
	var (
		createResponseBody string
		createStatusCode   int
		instanceType       string
		vmName             string
		vmPayload          string
		isInstanceCreated  bool
	)

	BeforeAll(func() {
		// name and payload creation
		vmName = "automation-vm-" + utils.GetRandomString()
		instanceType = utils.GetJsonValue("instanceType")
		vmPayload = utils.GetJsonValue("instancePayload")
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Attempt to create instance outside node pool", func() {
		It("Shouldn't succeed", func() {
			createStatusCode, createResponseBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, vmName, instanceType,
				sshkeyName, vnet, machineImageMapping, availabilityZone)
			Expect(createStatusCode).To(Equal(429), createResponseBody)
			Expect(strings.Contains(createResponseBody, `we are currently experiencing high demand`)).To(BeTrue(), createResponseBody)
			if createStatusCode == 200 {
				isInstanceCreated = true
			}
		})
	})

	When("Remove the instance via DELETE API", func() {
		It("should be successful", func() {
			logInstance.Println("Remove the instance via DELETE api using resource id")
		})
	})

	AfterAll(func() {
		// delete the instance created
		if isInstanceCreated {
			logInstance.Println("Remove the instance via DELETE api using resource id")
			instanceIdCreated := gjson.Get(createResponseBody, "metadata.resourceId").String()
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceIdCreated)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			logInstance.Println("Validation of Instance Deletion using Id")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceIdCreated)
			Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
		}
	})
})
