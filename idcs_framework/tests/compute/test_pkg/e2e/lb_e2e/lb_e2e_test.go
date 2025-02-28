package lb_e2e

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

var _ = Describe("Load Balancer E2E Test", Label("compute_e2e", "lb_e2e"), Ordered, ContinueOnFailure, func() {
	var (
		meteringPayload    string
		instanceType       string
		instanceResourceId string
		lbResourceId       string
		isInstanceCreated  bool
		isLBCreated        bool
		isInstanceDeleted  bool
		isLBDeleted        bool
		listOfImages       []string
	)

	BeforeAll(func() {
		// load instance details
		instanceName := "automation-vm-" + utils.GetRandomString()
		instancePayload := utils.GetJsonValue("instancePayload")
		instanceType = utils.GetJsonValue("instanceTypeToBeCreatedVM")
		getProductsPayload := utils.GetJsonValue("productsPayload")
		meteringPayload = utils.GetJsonValue("meteringMonitoringPayload")
		if adminToken != "" {
			adminToken = "Bearer " + adminToken
		} else {
			adminToken = utils.FetchAdminToken(testEnv)
		}

		// Get the product list
		getStatusCode, getResponseBody := utils.GetAllProducts(productsEndpoint, token, getProductsPayload, cloudAccountId)
		Expect(getStatusCode).To(Equal(200), getResponseBody)
		productsList := utils.GetProductsList(getResponseBody)
		logInstance.Println("products List: ", productsList)
		Expect(productsList).To(ContainElement(instanceType), "instance type is not in the product list")

		// Get image for the instance type
		logInstance.Println("Get list of machine images for instance type: ", instanceType)
		listOfImages = utils.FindImageByInstanceType(imageAndProductsList, instanceType, predefinedMachineImages)
		logInstance.Println("list of images by instance type: ", listOfImages)
		pickRandomMachineImage := utils.PickRandomItemFromList(listOfImages)
		logInstance.Println("Random machine image picked from the list: ", pickRandomMachineImage)

		// instance creation
		logInstance.Println("Creation of instance via API...")
		createStatusCode, createRespBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, instancePayload, instanceName,
			instanceType, sshkeyName, vnet, pickRandomMachineImage, availabilityZone)
		Expect(createStatusCode).To(Equal(200), createRespBody)
		Expect(strings.Contains(createRespBody, `"name":"`+instanceName+`"`)).To(BeTrue(), createRespBody)
		instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in reinstanceEndpointady state")
		instanceValidation1 := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
		Eventually(instanceValidation1, 5*time.Minute, 5*time.Second).Should(BeTrue())
		isInstanceCreated = true
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Instance validation - prerequisite", func() {
		It("Validate the instance creation in before all", func() {
			logInstance.Println("is instance created? " + strconv.FormatBool(isInstanceCreated))
			Expect(isInstanceCreated).Should(BeTrue())
		})
	})

	When("Creation of load balancer - via API...", func() {
		It("creation should be successful...", func() {
			logInstance.Println("LB creation via api...")
			lbName := "automation-lb-" + utils.GetRandomString()
			lbInstancePayload := utils.GetJsonValue("lbInstancePayload")

			createStatusCode, createRespBody := service_apis.CreateLB(lbInstanceEndpoint, token, lbInstancePayload, lbName, cloudAccount, "80", "tcp", instanceResourceId, "any")
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+lbName+`"`)).To(BeTrue(), createRespBody)
			lbResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

			// Validation
			logInstance.Println("Checking whether LB instance is in ready state")
			instanceValidation := utils.CheckLBPhase(lbInstanceEndpoint, token, lbResourceId)
			Eventually(instanceValidation, 30*time.Minute, 5*time.Second).Should(BeTrue())
			isLBCreated = true
			Expect(isLBCreated).Should(BeTrue(), "Instance creation failed with following error "+createRespBody)
		})
	})

	When("LB creation validation - prerequisite", func() {
		It("Validate the lb creation..", func() {
			logInstance.Println("is LB created? " + strconv.FormatBool(isLBCreated))
			Expect(isLBCreated).Should(BeTrue())
		})
	})

	When("Record validation in DB after Instance and LB creation via metering service", func() {
		It("Record validation in DB after instance creation via metering service", func() {
			logInstance.Println("Starting validation flow via metering service")
			getMeteringRecords := utils.GetMeteringRecords(meteringEndpoint, adminToken, meteringPayload, cloudAccount, instanceResourceId)
			Eventually(getMeteringRecords, 5*time.Minute, 5*time.Second).Should(BeTrue())

			// record validation for load balancer resource - once metering is enabled for LB
		})
	})

	When("Record validation for instance and LB in DB after wait time via API", func() {
		It("Record validation in DB after wait time via API", func() {
			logInstance.Println("Record validation in DB after wait time via API")
			responseStatus, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, meteringPayload, cloudAccount, instanceResourceId)
			Expect(responseStatus).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `"resourceId":"`+instanceResourceId+`"`)).To(BeTrue(), "assert the record in metering db using resource id")
			// Validation for new records
			response := string(responseBody)
			responses := strings.Split(response, "\n")
			numofRecords := len(responses)
			logInstance.Println("Number of records: ", numofRecords)
			Expect(numofRecords).To(BeNumerically(">", 1), "Mismatch in number of records found.")
			// Compare the run time of instance
			var allRunningSeconds []float64
			for _, eachResponse := range responses {
				allRunningSeconds = append(allRunningSeconds, gjson.Get(eachResponse, "result.properties.runningSeconds").Float())
			}
			allRunningSeconds = allRunningSeconds[:len(allRunningSeconds)-1]
			Expect(utils.IsArrayInIncreasingOrder(allRunningSeconds)).To(Equal(true))

			// record validation for load balancer resource - once metering is enabled for LB
		})
	})

	When("Instance Deletion and Validation via API", func() {
		It("Instance Deletion and Validation via API", func() {
			// LB deletion
			logInstance.Println("Remove the LB via DELETE api using resource id")
			deleteStatusCode, deleteRespBody := service_apis.DeleteLBById(lbInstanceEndpoint, token, lbResourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of LB Instance Deletion")
			lbValidation := utils.CheckLBDeletionById(lbInstanceEndpoint, token, lbResourceId)
			Eventually(lbValidation, 30*time.Minute, 5*time.Second).Should(BeTrue())
			isLBDeleted = true

			// instance deletion using resource id
			logInstance.Println("Remove the instance via DELETE api using resource id")
			deleteStatusCode, deleteRespBody = service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Deletion using Id")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
			Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
			isInstanceDeleted = true

		})
	})

	When("Record validation of instance and LB in DB after deletion via metering service", func() {
		It("Record validation in DB after deletion via metering service", func() {
			logInstance.Println("Starting validation flow via metering service...")
			statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, meteringPayload, cloudAccount, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `"deleted":"true"`)).To(BeTrue(), responseBody)

			// record validation of LB after deletion - lb metering is yet to be enabled
		})
	})

	AfterAll(func() {
		if !isLBDeleted {
			// LB deletion
			logInstance.Println("Remove the LB via DELETE api using resource id")
			deleteStatusCode, deleteRespBody := service_apis.DeleteLBById(lbInstanceEndpoint, token, lbResourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)
			// Validation
			logInstance.Println("Validation of Instance Deletion")
			lbValidation := utils.CheckLBDeletionById(lbInstanceEndpoint, token, lbResourceId)
			Eventually(lbValidation, 30*time.Minute, 30*time.Second).Should(BeTrue())
		}

		if !isInstanceDeleted {
			// Instance deletion using resource id
			logInstance.Println("Remove the instance via DELETE api using resource id")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Deletion using Id")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
			Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
		}
	})
})
