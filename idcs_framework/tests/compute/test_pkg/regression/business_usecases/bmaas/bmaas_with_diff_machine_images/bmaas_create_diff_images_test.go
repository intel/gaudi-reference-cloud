package diffMachineImage

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	"time"
)

var _ = Describe("Instance creation with all available machineImages - BMaaS", Label("compute", "bm_with_all_images"), Ordered, func() {

	var instanceType string

	BeforeAll(func() {
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
	})

	When("Attempt to create instances with different images", func() {
		It("Attempt to create instances with different images", func() {
			defer GinkgoRecover()

			// Define channels to signal when all instances are created
			allInstancesCreated := make(chan string, len(allMachineImages))

			for _, machineImage := range allMachineImages {
				logInstance.Println("Creating the instance with image: " + machineImage)
				go utils.CreateInstancesWithAllImages(instanceEndpoint, token, sshkeyName, vnet, machineImage, availabilityZone, instanceType,
					allInstancesCreated, 20*time.Minute)
			}

			// Wait for all instances to be created
			for range allMachineImages {
				resourceId := <-allInstancesCreated
				resourceIds = append(resourceIds, resourceId)
			}

		})
	})

	When("Attempt to delete all the created instances", func() {
		It("Attempt to delete all the created instances", func() {
			// Define channel to signal when all instances are deleted
			allInstancesDeleted := make(chan struct{}, len(allMachineImages))

			// Launch goroutines to delete instances concurrently
			for _, id := range resourceIds {
				go utils.DeleteAllInstances(instanceEndpoint, token, id, allInstancesDeleted)
			}

			// Wait for all instances to be deleted
			for range allMachineImages {
				<-allInstancesDeleted
			}
		})
	})
})
