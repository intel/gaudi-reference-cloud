package alltypes

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

func instanceCreationAndValidation(instance_type string) string {
	bmName := "automation-instance-" + utils.GetRandomString()
	bmPayload := utils.GetJsonValue("instancePayload")
	createStatusCode, createRespBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload, bmName,
		instance_type, sshkeyName, vnet, machineImageMapping, availabilityZone)
	Expect(createStatusCode).To(Equal(200), createRespBody)
	Expect(strings.Contains(createRespBody, `"name":"`+bmName+`"`)).To(BeTrue(), createRespBody)
	instanceResourceId := gjson.Get(createRespBody, "metadata.resourceId").String()

	// Validation
	logInstance.Println("Checking whether instance is in ready state")
	instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
	Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
	return instanceResourceId
}

var _ = Describe("BMaaS provision with all instance types", Label("compute", "bm_all_types", "bm-spr"), func() {
	When("create an spr instance", func() {
		It("create an spr instance", func() {
			logInstance.Println("Starting the spr Instance Creation via API...")
			instanceResourceId := instanceCreationAndValidation("bm-spr")
			resourceIds = append(resourceIds, instanceResourceId)
		})
	})
})

var _ = Describe("BMaaS provision with all instance types", Label("compute", "bm_all_types", "bm-spr-gaudi2"), func() {
	When("create an spr gaudi instance", func() {
		It("create an spr gaudi instance", func() {
			logInstance.Println("Starting the gaudi Instance Creation via API...")
			instanceResourceId := instanceCreationAndValidation("bm-spr-gaudi2")
			resourceIds = append(resourceIds, instanceResourceId)
		})
	})
})

var _ = Describe("BMaaS provision with all instance types", Label("compute", "bm_all_types", "bm-spr-pvc-1550-8"), func() {
	When("create a pvc-1550-8 instance", func() {
		It("create a pvc-1550-8 instance", func() {
			logInstance.Println("Starting the pvc-1550-8 Instance Creation via API...")
			instanceResourceId := instanceCreationAndValidation("bm-spr-pvc-1550-8")
			resourceIds = append(resourceIds, instanceResourceId)
		})
	})
})

var _ = Describe("BMaaS provision with all instance types", Label("compute", "bm_all_types", "bm-spr-pvc-1100-8"), func() {
	When("create a pvc-1100-8 instance", func() {
		It("create a pvc-1100-8 instance", func() {
			logInstance.Println("Starting the pvc-1100-8 Instance Creation via API...")
			instanceResourceId := instanceCreationAndValidation("bm-spr-pvc-1100-8")
			resourceIds = append(resourceIds, instanceResourceId)
		})
	})
})

var _ = Describe("BMaaS provision with all instance types", Label("compute", "bm_all_types", "bm-spr-pvc-1100-4"), func() {
	When("create a pvc-1100-4 instance", func() {
		It("create a pvc-1100-4 instance", func() {
			logInstance.Println("Starting the pvc-1100-4 Instance Creation via API...")
			instanceResourceId := instanceCreationAndValidation("bm-spr-pvc-1100-4")
			resourceIds = append(resourceIds, instanceResourceId)
		})
	})
})
