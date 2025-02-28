package bmaas

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("Compute instance api(BM positive flow)", Label("compute", "bmaas", "compute_instance", "bmaas_instance", "bmaas_instance_positive"), Ordered, ContinueOnFailure, func() {
	var (
		createStatusCode   int
		createRespBody     string
		bmName             string
		instanceType       string
		bmPayload          string
		instancePutPayload string
		isInstanceCreated  bool
		instanceResourceId string
	)

	BeforeAll(func() {

		if skipBMCreation == "true" {
			Skip("Skipping the Entire BM positive flow due to the flag")
		}

		// load instance details to be created
		instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
		instancePutPayload = utils.GetJsonValue("instancePutPayload")
		bmName = "automation-bm-" + utils.GetRandomString()
		bmPayload = utils.GetJsonValue("instancePayload")

		// instance creation
		logInstance.Println("Starting Instance Creation flow via Instance API...")
		createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload, bmName,
			instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+bmName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 30*time.Minute, 30*time.Second).Should(BeTrue())
		isInstanceCreated = true
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Instance creation and its validation - prerequisite", func() {
		It("Validate the instance creation in before all", func() {
			logInstance.Println("is instance created? " + strconv.FormatBool(isInstanceCreated))
			Expect(isInstanceCreated).Should(BeTrue(), "Instance creation failed with following error "+createRespBody)
		})
	})

	When("Retrieving an instance created using resource name", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance via GET method using name")
			statusCode, responseBody := service_apis.GetInstanceByName(instanceEndpoint, token, bmName)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Updating already created instance using resource name", func() {
		It("should be successful", func() {
			logInstance.Println("Modify the instance via PUT method using resource name")
			payload := instancePutPayload
			payload = strings.Replace(payload, "<<ssh-public-key>>", sshkeyName, 1)
			statusCode, responseBody := service_apis.PutInstanceByName(instanceEndpoint, token, bmName, payload)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Listing all the instance's created in CA", func() {
		It("should be successful", func() {
			logInstance.Println("List all the instances via GET api")
			statusCode, responseBody := service_apis.GetAllInstance(instanceEndpoint, token, nil)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("Retrieving an instance created using resource id", func() {
		It("should be successful", func() {
			logInstance.Println("Retrieve the instance via GET method using id")
			statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
		})
	})

	When("SSH into the BM instance created", func() {
		It("SSH into the BM instance created", func() {
			logInstance.Println("SSH into the BM instance created")
			statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
			err := utils.SSHIntoInstance(responseBody, "../../../ansible-files", "../../../ansible-files/inventory.ini",
				"../../../ansible-files/ssh-and-apt-get-on-bm.yml", "~/.ssh/id_rsa")
			Expect(err).NotTo(HaveOccurred(), err)
		})
	})

	// powering ON the existing powered on bmaas instance (Run strategy RerunOnFailure -> RerunOnFailure)
	Context("when the instance is powered ON", func() {
		It("Updating the run strategy from RerunOnFailure to RerunOnFailure", func() {
			logInstance.Println("Updating the run strategy from RerunOnFailure to RerunOnFailure...")
			statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "RerunOnFailure", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			// Get the instance and validate the run-strategy and phase
			statusCode, responseBody = service_apis.GetInstanceByName(instanceEndpoint, token, bmName)
			Expect(statusCode).To(Equal(200), responseBody)
			instanceRunStrategy := gjson.Get(responseBody, "spec.runStrategy").String()
			Expect(instanceRunStrategy == "RerunOnFailure").To(Equal(true), "run strategy assertion failed: %v", instanceRunStrategy)
			instancePhase := gjson.Get(responseBody, "status.phase").String()
			Expect(instancePhase == "Ready").To(Equal(true), "instance phase should be ready - assertion failed: %v", instancePhase)
		})

		It("Updating the run strategy from RerunOnFailure to Always & Always to RerunOnFailure..", func() {
			logInstance.Println("Updating the run strategy from RerunOnFailure to Always...")
			// update the payload (RerunOnFailure to Always)
			statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Always", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			// Get the instance and validate the run-strategy and phase
			statusCode, responseBody = service_apis.GetInstanceByName(instanceEndpoint, token, bmName)
			Expect(statusCode).To(Equal(200), responseBody)
			instanceRunStrategy := gjson.Get(responseBody, "spec.runStrategy").String()
			Expect(instanceRunStrategy == "Always").To(Equal(true), "run strategy assertion failed: %v", instanceRunStrategy)
			instancePhase := gjson.Get(responseBody, "status.phase").String()
			Expect(instancePhase == "Ready").To(Equal(true), "instance phase should be ready - assertion failed: %v", instancePhase)

			// update the payload (Always to RerunOnFailure)
			logInstance.Println("Updating the run strategy from Always to RerunOnFailure...")
			statusCode, responseBody = utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "RerunOnFailure", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			// Get the instance and validate the run-strategy and phase
			statusCode, responseBody = service_apis.GetInstanceByName(instanceEndpoint, token, bmName)
			Expect(statusCode).To(Equal(200), responseBody)
			instanceRunStrategy = gjson.Get(responseBody, "spec.runStrategy").String()
			Expect(instanceRunStrategy == "RerunOnFailure").To(Equal(true), "run strategy assertion failed: %v", instanceRunStrategy)
			instancePhase = gjson.Get(responseBody, "status.phase").String()
			Expect(instancePhase == "Ready").To(Equal(true), "instance phase should be ready - assertion failed: %v", instancePhase)
		})

		It("Updating the run strategy from Always to Always...", func() {
			logInstance.Println("Updating the run strategy from Always to Always...")
			// update the payload (RerunOnFailure to Always)
			statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Always", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			//Repeat the updation (always to always)
			statusCode, responseBody = utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Always", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			// Get the instance and validate the run-strategy and phase
			statusCode, responseBody = service_apis.GetInstanceByName(instanceEndpoint, token, bmName)
			Expect(statusCode).To(Equal(200), responseBody)
			instanceRunStrategy := gjson.Get(responseBody, "spec.runStrategy").String()
			Expect(instanceRunStrategy == "Always").To(Equal(true), "run strategy assertion failed: %v", instanceRunStrategy)
			instancePhase := gjson.Get(responseBody, "status.phase").String()
			Expect(instancePhase == "Ready").To(Equal(true), "instance phase should be ready - assertion failed: %v", instancePhase)
		})
	})

	Context("When the instance is in transitioning phase (ON to OFF and viceversa)", func() {
		It("Attempt to Update an instance during transition stage(starting or stopping phase)", func() {
			// stop the instance
			logInstance.Println("Updating the run strategy from Always to Halted...")
			statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Halted", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			instancePhaseValidation := utils.CheckInstanceState(instanceEndpoint, token, instanceResourceId, "Stopping")
			Eventually(instancePhaseValidation, 20*time.Minute, 5*time.Second).Should(BeTrue())

			// start the instance immediately during stopping phase
			logInstance.Println("Updating the run strategy to Always when the instance is in stopping phase...")
			statusCode, responseBody = utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Always", sshkeyName, bmName)
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `unable to update the instance Run Strategy to Always during Stopping`)).To(BeTrue(), responseBody)

			// attempt to stop the instance again during stopping phase
			logInstance.Println("Updating the run strategy to Halted when the instance is in stopping phase...")
			statusCode, responseBody = utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Halted", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			// validate whether the instance phase is moved to stopped state
			instancePhaseValidation = utils.CheckInstanceState(instanceEndpoint, token, instanceResourceId, "Stopped")
			Eventually(instancePhaseValidation, 20*time.Minute, 10*time.Second).Should(BeTrue())

			// start the instance
			logInstance.Println("Updating the run strategy from Halted to Always...")
			statusCode, responseBody = utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Always", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			instancePhaseValidation = utils.CheckInstanceState(instanceEndpoint, token, instanceResourceId, "Starting")
			Eventually(instancePhaseValidation, 20*time.Minute, 5*time.Second).Should(BeTrue())

			// stop the instance immediately during starting phase
			logInstance.Println("Updating the run strategy to Halted when the instance is in starting phase...")
			statusCode, responseBody = utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Halted", sshkeyName, bmName)
			Expect(statusCode).To(Equal(400), responseBody)
			Expect(strings.Contains(responseBody, `unable to update the Instance RunStrategy to Halted during Starting`)).To(BeTrue(), responseBody)

			// attempt to start the instance again during starting phase
			logInstance.Println("Updating the run strategy to Always when the instance is in starting phase...")
			statusCode, responseBody = utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Always", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			instancePhaseValidation = utils.CheckInstanceState(instanceEndpoint, token, instanceResourceId, "Ready")
			Eventually(instancePhaseValidation, 20*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Updating the run strategy from halted to halted", func() {
		It("should be successful", func() {
			logInstance.Println("Updating the run strategy from halted to halted...")
			statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Halted", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			// validate whether the instance phase is moved to stopped state
			instancePhaseValidation := utils.CheckInstanceState(instanceEndpoint, token, instanceResourceId, "Stopped")
			Eventually(instancePhaseValidation, 30*time.Minute, 10*time.Second).Should(BeTrue())

			// Get the instance and validate the run-strategy and phase
			statusCode, responseBody = service_apis.GetInstanceByName(instanceEndpoint, token, bmName)
			Expect(statusCode).To(Equal(200), responseBody)
			instanceRunStrategy := gjson.Get(responseBody, "spec.runStrategy").String()
			Expect(instanceRunStrategy == "Halted").To(Equal(true), "run strategy assertion failed: %v", instanceRunStrategy)
			instancePhase := gjson.Get(responseBody, "status.phase").String()
			Expect(instancePhase == "Stopped").To(Equal(true), "instance phase should be ready - assertion failed: %v", instancePhase)
		})
	})

	When("Remove the instance via DELETE api using resource id", func() {
		It("Placeholder for instance deletion", func() {
			logInstance.Println("Remove the instance via DELETE api using resource id")
		})
	})

	AfterAll(func() {
		// Instance deletion using resource id is covered here
		logInstance.Println("Remove the instance via DELETE api using resource id")
		deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
		Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
	})
})
