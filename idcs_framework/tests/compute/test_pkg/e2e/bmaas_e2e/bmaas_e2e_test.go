package bm_e2e

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("BMaaS E2E Test", Label("compute_e2e", "bm_e2e", "bm_e2e_single"), Ordered, func() {
	var (
		productsPayload     string
		listOfImages        []string
		bmName              string
		bmPayload           string
		metPayload          string
		instanceResourceId  string
		firstReadyTimestamp string
		isInstanceDeleted   bool
		deletionTimestamp   string
	)

	BeforeAll(func() {
		if skipBMCreation == "true" {
			Skip("Skipping the BM e2e flow due to the flag")
		}
		productsPayload = utils.GetJsonValue("productsPayload")
		bmName = "automation-bm-e2e-" + utils.GetRandomString()
		bmPayload = utils.GetJsonValue("instancePayload")
		metPayload = utils.GetJsonValue("meteringMonitoringPayload")
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Get List of products and ensure instance type is present", func() {
		It("Get list of products and ensure instance type is present", func() {
			statusCode, responseBody := utils.GetAllProducts(productsEndpoint, token, productsPayload, cloudAccountId)
			Expect(statusCode).To(Equal(200), responseBody)

			// product list
			productsList := utils.GetProductsList(responseBody)
			logInstance.Println("products List: ", productsList)

			// Check to ensure instance type is present
			Expect(productsList).To(ContainElement(bmInstanceType), "instance type is not in the product list")
		})
	})

	When("Ensure machine image exists", func() {
		It("Ensure machine image exists", func() {
			logInstance.Println("Get list of machine images for instance type: ", bmInstanceType)
			listOfImages = utils.FindImageByInstanceType(imageAndProductsList, bmInstanceType, nil)
			logInstance.Println("list of images by instance type: ", listOfImages)

			// Verify if image exists
			logInstance.Println("Ensure machine image exists")
			Expect(listOfImages).To(ContainElement(bmMachineImage), "Machine image does not exist")
		})
	})

	When("BM Instance Creation", func() {
		It("Instance Creation using the API", func() {
			createStatusCode, createRespBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, bmPayload, bmName,
				bmInstanceType, sshkeyName, vnet, bmMachineImage, availabilityZone)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+bmName+`"`)).To(BeTrue(), createRespBody)
			instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

			// Validation
			logInstance.Println("Checking whether instance is in ready state")
			instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
			Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
		})
	})

	When("Record validation in DB after instance creation via metering service", func() {
		It("Record validation in DB after instance creation via metering service", func() {
			logInstance.Println("Starting validation flow via metering service")
			//statusCode, _ := utils.SearchMeteringRecords(meteringEndpoint, token, metPayload, cloudAccount, instanceResourceId)
			//Expect(statusCode).To(Equal(200), "assertion failed on response code")

			getMeteringRecords := utils.GetMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
			Eventually(getMeteringRecords, 5*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("SSH into the BM instance created", func() {
		It("SSH into the BM instance created", func() {
			statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
			err := utils.SSHIntoInstance(responseBody, "../../ansible-files", "../../ansible-files/inventory.ini",
				"../../ansible-files/ssh-and-apt-get-on-bm.yml", "~/.ssh/id_rsa")
			Expect(err).ShouldNot(HaveOccurred(), err)
		})
	})

	When("Record Validation in DB after creation via metering service", func() {
		It("Record Validation in DB after creation via metering service", func() {
			time.Sleep(3 * time.Minute)
		})
	})

	When("Record validation in DB after wait time via API", func() {
		It("Record validation in DB after wait time via API", func() {
			logInstance.Println("Record validation in DB after wait time via API")
			statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `"resourceId":"`+instanceResourceId+`"`)).To(BeTrue(), "assert the record in metering db using resource id: %v", responseBody)

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
		})
	})

	When("when the instance is powered ON", func() {
		It("Attempt to Stop the instance", func() {
			// Stop the instance
			logInstance.Println("Updating the run strategy from Always to Halted...")
			statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Halted", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			// Validation
			instancePhaseValidation := utils.CheckInstanceState(instanceEndpoint, token, instanceResourceId, "Stopped")
			Eventually(instancePhaseValidation, 10*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("when the instance is powered OFF", func() {
		It("Start the instance", func() {
			// Start the instance
			logInstance.Println("Updating the run strategy from Halted to Always...")
			statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Always", sshkeyName, bmName)
			Expect(statusCode).To(Equal(200), responseBody)

			// Validation
			instancePhaseValidation := utils.CheckInstanceState(instanceEndpoint, token, instanceResourceId, "Ready")
			Eventually(instancePhaseValidation, 10*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Instance Deletion and Validation via API", func() {
		It("Instance Deletion and Validation via API", func() {
			logInstance.Println("Deletion and Validation of Instance via API")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Deletion")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
			Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
			isInstanceDeleted = true

			deletionTime := time.Now()
			deletionTimestamp = deletionTime.Format(time.RFC3339)
			logInstance.Println(deletionTimestamp)
		})
	})

	When("Record validation in DB after deletion via metering service", func() {
		It("Record validation in DB after deletion via metering service", func() {
			logInstance.Println("Starting validation flow via metering service...")
			time.Sleep(2 * time.Minute)
			statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
			Expect(statusCode).To(Equal(200), responseBody)
			Expect(strings.Contains(responseBody, `"deleted":"true"`)).To(BeTrue(), responseBody)

			// firstReadyTimestamp from the record
			firstReadyTimestamp = gjson.Get(responseBody, "result.properties.firstReadyTimestamp").String()
			logInstance.Println("firstReadyTimestamp: ", firstReadyTimestamp)
		})
	})

	// Commenting usage check due to a dependency on 'reportUsageSchedulerInterval' in billing configmap
	// The usage is reported every 3600 seconds with respect to the 'reportUsageSchedulerInterval' parameter
	/*
		When("Validate usage for the created BM instance", func() {
			It("Validate usage for the created BM instance", func() {
				usageParams := map[string]string{
					"cloudAccountId":    cloudAccount,
					"searchStart"   :    firstReadyTimestamp,
					"searchEnd"     :    deletionTimestamp,
					"regionName"    :    region,

				}
				statusCode, responseBody := service_apis.GetUsage(usagesEndpoint, token, usageParams)
				Expect(statusCode).To(Equal(200), responseBody)
				// Fetch product type from the body and validate
				productType := gjson.Get(responseBody, "usages.productType")
				Expect(productType).To(Equal(instanceTypeToBeCreatedBM), responseBody)

				// Ensure total usage is > 0
				totalUsage := gjson.Get(responseBody, "totalUsage").Int()
				Expect(totalUsage).To(BeNumerically(">", 0), "Total Usage is not greater than 0.")
			})
		})
	*/

	// billing test case : TODO

	AfterAll(func() {
		// Delete the instance if the deletion failed above
		if !isInstanceDeleted {
			logInstance.Println("Remove the instance via DELETE api using resource id")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Deletion using Id")
			instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
			Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
		}
	})
})
