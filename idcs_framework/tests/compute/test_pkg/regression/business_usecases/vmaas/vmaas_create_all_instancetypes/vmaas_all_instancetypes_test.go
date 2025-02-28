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
	vmName := "automation-instance-" + utils.GetRandomString()
	vmPayload := utils.GetJsonValue("instancePayload")
	createStatusCode, createRespBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, vmName,
		instance_type, sshkeyName, vnet, machineImageMapping, availabilityZone)
	Expect(createStatusCode).To(Equal(200), createRespBody)
	Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
	instanceResourceId := gjson.Get(createRespBody, "metadata.resourceId").String()

	// Validation
	logInstance.Println("Checking whether instance is in ready state")
	instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
	Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
	return instanceResourceId
}

// Commenting out - Since we don't support tiny for users
/*var _ = Describe("VMaaS provision with all instance types", Label("compute", "vm_all_types", "vm_tiny"), func() {
	When("create tiny instance", func() {
		It("create tiny instance", func() {
			logInstance.Println("Starting the Tiny Instance Creation via API...")
			instanceResourceId := instanceCreationAndValidation("vm-spr-tny")
			resourceIds = append(resourceIds, instanceResourceId)
		})
	})
})*/

// Skipping the small instance creation, since all the instance creation using small instance type
var _ = PDescribe("VMaaS provision with all instance types", Label("compute", "vm_all_types", "vm_small"), func() {
	When("create small instance", func() {
		It("Small instance creation and deletion", func() {
			By("Initializing instance creation")
			logInstance.Println("Starting the Small Instance Creation via API...")
			instanceResourceId := instanceCreationAndValidation("vm-spr-sml")
			resourceIds = append(resourceIds, instanceResourceId)

			By("Cleaning up the instance")
			utils.DeleteInstanceWithId(instanceEndpoint, token, instanceResourceId)
		})
	})
})

var _ = Describe("VMaaS provision with all instance types", Label("compute", "vm_all_types", "vm_medium"), func() {
	When("create medium instance", func() {
		It("Medium instance creation and deletion", func() {
			By("Initializing instance creation")
			logInstance.Println("Starting the Medium Instance Creation via API...")
			instanceResourceId := instanceCreationAndValidation("vm-spr-med")
			resourceIds = append(resourceIds, instanceResourceId)

			By("Cleaning up the instance")
			utils.DeleteInstanceWithId(instanceEndpoint, token, instanceResourceId)
		})
	})
})

var _ = Describe("VMaaS provision with all instance types", Label("compute", "vm_all_types", "vm_large"), func() {
	When("create large instance", func() {
		It("Large instance creation and deletion", func() {
			By("Initializing instance creation")
			logInstance.Println("Starting the Large Instance Creation via API...")
			instanceResourceId := instanceCreationAndValidation("vm-spr-lrg")
			resourceIds = append(resourceIds, instanceResourceId)

			By("Cleaning up the instance")
			utils.DeleteInstanceWithId(instanceEndpoint, token, instanceResourceId)
		})
	})
})

/*var _ = Describe("VMaaS provision with all instance types", Label("compute", "vm_all_types", "vm_gaudi"), func() {
    When("create large instance", func() {
        It("create large instance", func() {
            logInstance.Println("Starting the Gaudi2 Instance Creation via API...")
            instanceCreationAndValidation("vm-icp-gaudi2")
        })
    })
})
var _ = Describe("VMaaS provision with all instance types", Label("compute", "vm_all_types", "vm_pvc"), func() {
    When("create large instance", func() {
        It("create large instance", func() {
            logInstance.Println("Starting the PVC Instance Creation via API...")
            instanceCreationAndValidation("vm-spr-pvc-1100-1")
        })
    })
})*/
