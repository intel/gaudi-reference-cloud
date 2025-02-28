package anti_affinity

import (
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("VMaaS Anti Affinity Validation", Label("compute", "compute_business_uc", "vm_anti_affinity"), Ordered, func() {

	var (
		podCountInEachNodeMap map[string]int
		instancesCreated      []string
	)

	BeforeAll(func() {
		podCountInEachNodeMap = make(map[string]int)
		nodeArray, _ := client.GetNodes()
		for _, eachNode := range nodeArray {
			podCountInEachNodeMap[eachNode] = 0
			logInstance.Println("Node: ", zap.String("name", eachNode))
		}

	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Create instance with topology constraint", func() {
		It("validate the instance created by scheduler", func() {
			// create and validate the first instance
			podCountInEachNodeMap, instancesCreated = createInstanceWithAntiAffinityRule(podCountInEachNodeMap, instancesCreated)

			// create and validate the second instance
			podCountInEachNodeMap, instancesCreated = createInstanceWithAntiAffinityRule(podCountInEachNodeMap, instancesCreated)
		})
	})

	AfterAll(func() {
		// Delete all the instances created
		// Define channel to signal when all instances are deleted
		allInstancesDeleted := make(chan struct{}, len(instancesCreated))

		// Launch goroutines to delete instances concurrently
		for _, id := range instancesCreated {
			go utils.DeleteAllInstances(instanceEndpoint, token, id, allInstancesDeleted)
		}

		// Wait for all instances to be deleted
		for range instancesCreated {
			<-allInstancesDeleted
		}
	})
})

func createInstanceWithAntiAffinityRule(podCountInEachNodeMap map[string]int, instancesCreated []string) (map[string]int, []string) {
	// Get all pods inside each node
	for eachNode := range podCountInEachNodeMap {
		podArray, _ := client.GetPodsInNode(eachNode)
		podCountInEachNodeMap[eachNode] = len(podArray)
		logInstance.Println("Number of pods in the node: " + eachNode + "is " + strconv.Itoa(len(podArray)))
	}

	// Find the number of pods inside each node
	minLoadNodeNames := findNodeWithLeastNumberOfPods(podCountInEachNodeMap)

	vmName := "automation-vm-" + utils.GetRandomString()
	vmPayload := utils.GetJsonValue("instanceAntiAffinityPayload")
	instanceType := utils.GetJsonValue("instanceTypeToBeCreated")

	//Instance creation
	createStatusCode, createRespBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, vmName, instanceType,
		sshkeyName, vnet, machineImageMapping, availabilityZone)
	Expect(createStatusCode).To(Equal(200), createRespBody)
	Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
	instanceResourceId := gjson.Get(createRespBody, "metadata.resourceId").String()
	instancesCreated = append(instancesCreated, instanceResourceId)

	// Validation
	logInstance.Println("Checking whether instance is in ready state")
	instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
	Eventually(instanceValidation, 5*time.Minute, 10*time.Second).Should(BeTrue())

	// validation of first instance
	nodeName, _ := client.GetNodeNameFromPod("virt-launcher-"+instanceResourceId, cloudAccount)
	fmt.Println("Instance created in the node: " + nodeName)
	Expect(stringInSlice(nodeName, minLoadNodeNames)).To(BeTrue())

	return podCountInEachNodeMap, instancesCreated
}

func findNodeWithLeastNumberOfPods(podCountInEachNodeMap map[string]int) []string {

	// Variables to track the smallest value and associated keys
	var minPods int
	var minLoadNodeNames []string

	// Assume the first element has the smallest value
	for _, value := range podCountInEachNodeMap {
		minPods = value
		break
	}

	// Iterate over the map to find the smallest value and all keys with that value
	for key, value := range podCountInEachNodeMap {
		if value < minPods {
			minPods = value
			minLoadNodeNames = []string{key}
		} else if value == minPods {
			minLoadNodeNames = append(minLoadNodeNames, key)
		}
	}

	// Print the smallest value and the corresponding keys
	fmt.Printf("The smallest value is %d for keys: %v\n", minPods, minLoadNodeNames)
	return minLoadNodeNames

}

func stringInSlice(target string, list []string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}
