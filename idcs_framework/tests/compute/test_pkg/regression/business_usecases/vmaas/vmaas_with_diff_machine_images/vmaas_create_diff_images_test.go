package diffMachineImage

import (
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

// skipping this, since all the other test cases are creating instance with this machine image.
var _ = PDescribe("IDC - VMaaS Creation with all machine images", Label("compute", "vm_all_machine_images", "mi_ubuntu_2204_jammyv20240308"), Ordered, func() {
	When("Create instance with the machine image ubuntu-2204-jammy-v20240308 ", func() {
		It("Instance creation and deletion", func() {
			By("Initializing instance creation")
			var instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
			logInstance.Println("Instance creation with machine image ubuntu-2204-jammy-v20240308...")
			instanceResourceId := instanceCreationWithAllMachineImages("ubuntu-2204-jammy-v20240308", instanceType)
			resourceIds = append(resourceIds, instanceResourceId)

			By("Cleaning up the instance")
			utils.DeleteInstanceWithId(instanceEndpoint, token, instanceResourceId)
		})
	})
})

var _ = Describe("IDC - VMaaS Creation with all machine images", Label("compute", "vm_all_machine_images", "mi-ubuntu-2204-jammy-v20230122"), Ordered, func() {
	When("Create instance with the machine image ubuntu-2204-jammy-v20230122 ", func() {
		It("Instance creation and deletion", func() {
			By("Initializing instance creation")
			var instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
			logInstance.Println("Instance creation with machine image ubuntu-2204-jammy-v20230122...")
			instanceResourceId := instanceCreationWithAllMachineImages("ubuntu-2204-jammy-v20230122", instanceType)
			resourceIds = append(resourceIds, instanceResourceId)

			By("Cleaning up the instance")
			utils.DeleteInstanceWithId(instanceEndpoint, token, instanceResourceId)
		})
	})
})

var _ = Describe("IDC - VMaaS Creation with all machine images", Label("compute", "vm_all_machine_images", "mi-ubuntu-2204-jammy-v20250123"), Ordered, func() {
	When("Create instance with the machine image ubuntu-2204-jammy-v20250123", func() {
		It("Instance creation and deletion", func() {
			By("Initializing instance creation")
			var instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
			logInstance.Println("Instance creation with machine image ubuntu-2204-jammy-v20250123...")
			instanceResourceId := instanceCreationWithAllMachineImages("ubuntu-2204-jammy-v20250123", instanceType)
			resourceIds = append(resourceIds, instanceResourceId)

			By("Cleaning up the instance")
			utils.DeleteInstanceWithId(instanceEndpoint, token, instanceResourceId)
		})
	})
})

func instanceCreationWithAllMachineImages(machineImage, instanceType string) string {
	logInstance.Println("Creating the instance with image: " + machineImage)
	name := "test-automation-" + utils.GetRandomString()
	payload := utils.GetJsonValue("instancePayload")
	machineImage = strings.Trim(machineImage, `"`)
	statusCode, responseBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, payload,
		name, instanceType, sshkeyName, vnet, machineImage, availabilityZone)

	Expect(statusCode).To(Equal(200), responseBody)
	Expect(strings.Contains(responseBody, `"name":"`+name+`"`)).To(BeTrue(), responseBody)
	instanceResourceId := gjson.Get(responseBody, "metadata.resourceId").String()

	// Validation
	logInstance.Println("Checking whether instance is in ready state")
	instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
	Eventually(instanceValidation, 5*time.Minute, 10*time.Second).Should(BeTrue())
	return instanceResourceId
}
