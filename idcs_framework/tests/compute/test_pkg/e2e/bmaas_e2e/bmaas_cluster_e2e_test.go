package bm_e2e

import (
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("BM Instance Group E2E Test", Label("compute_e2e", "bm_e2e", "bm_cluster_e2e"), Ordered, func() {
	var (
		getProductsPayload             string
		listOfImages                   []string
		listOfIds                      []string
		instanceGroupName              string
		bmPayload                      string
		metPayload                     string
		searchPayload                  string
		firstReadyTimestamp            string
		deletionTimestamp              string
		instanceGroupDeletion          = false
		instanceGroupCreationTimestamp int64
	)

	BeforeAll(func() {
		if skipBMClusterCreation == "true" {
			Skip("Skipping the BM e2e flow due to the flag")
		}
		getProductsPayload = utils.GetJsonValue("productsPayload")
		instanceGroupName = "automation-bm-group-e2e-" + utils.GetRandomString()
		bmPayload = utils.GetJsonValue("instanceGroupPayload")
		metPayload = utils.GetJsonValue("meteringMonitoringPayload")
		searchPayload = utils.GetJsonValue("instanceSearchPayload")
	})

	JustBeforeEach(func() {
		logInstance.Println("----------------------------------------------------")
	})

	When("Get List of products and ensure instance type is present", func() {
		It("Get list of products and ensure instance type is present", func() {
			allProductsResponseStatus, allProductsResponseBody := utils.GetAllProducts(productsEndpoint, token, getProductsPayload, cloudAccountId)
			Expect(allProductsResponseStatus).To(Equal(200), allProductsResponseBody)

			// product list
			productsList := utils.GetProductsList(allProductsResponseBody)
			logInstance.Println("products List: ", productsList)

			// Check to ensure instance type is present
			Expect(productsList).To(ContainElement(bmClusterInstanceType), "instance type is not in the product list")
		})
	})

	When("Get image for instance type", func() {
		It("Get image for instance type", func() {
			logInstance.Println("Get list of machine images for instance type: ", bmClusterInstanceType)
			listOfImages = utils.FindImageByInstanceType(imageAndProductsList, bmClusterInstanceType, nil)
			logInstance.Println("list of images by instance type: ", listOfImages)

			// Verify if image exists
			logInstance.Println("Ensure machine image exists")
			Expect(listOfImages).To(ContainElement(bmClusterMachineImage), "Machine image does not exist")
		})
	})

	When("BM Instance Group Creation", func() {
		It("Instance Group Creation using the API", func() {
			logInstance.Println("Instance Group Creation using the API")
			createStatusCode, createRespBody := service_apis.InstanceGroupCreationWithoutMIMap(instanceGroupEndpoint, token, bmPayload,
				instanceGroupName, instanceGroupSize, bmClusterInstanceType, sshkeyName, vnet, bmClusterMachineImage, availabilityZone)
			Expect(createStatusCode).To(Equal(200), createRespBody)
			Expect(strings.Contains(createRespBody, `"name":"`+instanceGroupName+`"`)).To(BeTrue(), createRespBody)

			// Validation
			logInstance.Println("Checking whether instances in the group are in ready state")
			instanceGroupValidation := utils.CheckInstanceGroupProvisionState(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())

			instanceGroupCreationTimestamp = time.Now().UnixMilli()
			logInstance.Println(fmt.Sprint(instanceGroupCreationTimestamp))

			// Get list of Ids from the group
			listOfIds = utils.GetInstanceIdsFromInstanceGroup(instanceEndpoint, token, instanceGroupName, searchPayload)
			logInstance.Println("listOfIds ", listOfIds)
			expectedLength, _ := strconv.Atoi(instanceGroupSize)
			Expect(len(listOfIds)).To(Equal(expectedLength), "Number of resource ids are not matching with the number of instances")
		})
	})

	When("Record validation in DB after instance creation via metering service", func() {
		It("Record validation in DB after instance creation via metering service", func() {
			for _, eachInstanceResourceId := range listOfIds {
				logInstance.Println("Starting validation flow via metering service")
				//statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, token, metPayload, cloudAccount, eachInstanceResourceId)
				//Expect(statusCode).To(Equal(200), responseBody)

				getMeteringRecords := utils.GetMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, eachInstanceResourceId)
				Eventually(getMeteringRecords, 5*time.Minute, 5*time.Second).Should(BeTrue())
			}
		})
	})

	When("Validate metering records after instance group creation", func() {
		It("Validate metering records after instance group creation", func() {
			logInstance.Println("Starting validation flow via metering service after instance group creation")

			for _, eachInstanceResourceId := range listOfIds {
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, eachInstanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)
				firstReadyTimestamp = gjson.Get(responseBody, "result.properties.firstReadyTimestamp").String()
				logInstance.Println("firstReadyTimestamp: ", firstReadyTimestamp)

				// validate other metering fields for all the instances present inside group
				latestRecord := utils.GetLastMeteringRecord(meteringEndpoint, adminToken, metPayload, cloudAccount, eachInstanceResourceId)
				logInstance.Println(latestRecord)

				//Retrieve the required fields to validate
				cloudAccountIdFromRecord := gjson.Get(latestRecord, "result.cloudAccountId").String()
				resourceIdFromRecord := gjson.Get(latestRecord, "result.resourceId").String()
				availabilityZoneFromRecord := gjson.Get(latestRecord, "result.properties.availabilityZone").String()
				deletedFromRecord := gjson.Get(latestRecord, "result.properties.deleted").String()
				instanceGroupFromRecord := gjson.Get(latestRecord, "result.properties.instanceGroup").String()
				instanceGroupSizeFromRecord := gjson.Get(latestRecord, "result.properties.instanceGroupSize").String()
				instanceTypeFromRecord := gjson.Get(latestRecord, "result.properties.instanceType").String()
				regionFromRecord := gjson.Get(latestRecord, "result.properties.region").String()
				serviceTypeFromRecord := gjson.Get(latestRecord, "result.properties.serviceType").String()

				//validate the fields
				Expect(cloudAccountIdFromRecord).To(Equal(cloudAccount))
				Expect(resourceIdFromRecord).To(Equal(eachInstanceResourceId))
				Expect(availabilityZoneFromRecord).To(Equal(availabilityZone))
				Expect(deletedFromRecord).To(Equal("false"))
				Expect(instanceGroupFromRecord).To(Equal(instanceGroupName))
				Expect(instanceGroupSizeFromRecord).To(Equal(instanceGroupSize))
				Expect(instanceTypeFromRecord).To(Equal(bmClusterInstanceType))
				Expect(regionFromRecord).To(Equal(region))
				Expect(serviceTypeFromRecord).To(Equal("ComputeAsAService"))

				//Retrieve the first ready timestamp and validate
				firstReadyTimestampFromRecord := gjson.Get(latestRecord, "result.properties.firstReadyTimestamp").String()
				Expect(utils.ValidateTimeStamp(utils.GetUnixTime(firstReadyTimestampFromRecord), instanceGroupCreationTimestamp-100000, instanceGroupCreationTimestamp+100000)).Should(BeTrue())
			}

		})
	})

	When("SSH into the BM instances in the group", func() {
		It("SSH into the BM instance in the group", func() {
			logInstance.Println("SSH into the BM instance in the group")
			for _, eachInstanceResourceId := range listOfIds {
				statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, eachInstanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)

				err := utils.SSHIntoInstance(responseBody, "../../ansible-files", "../../ansible-files/inventory.ini",
					"../../ansible-files/ssh-and-apt-get-on-bm.yml", "~/.ssh/id_rsa")
				Expect(err).ShouldNot(HaveOccurred(), err)
			}
		})
	})

	When("Record Validation in DB after creation via metering service", func() {
		It("Record Validation in DB after creation via metering service", func() {
			time.Sleep(3 * time.Minute)
		})
	})

	When("Record validation in DB after wait time via API", func() {
		It("Record validation in DB after wait time via API", func() {
			for i, eachInstanceResourceId := range listOfIds {
				logInstance.Println("Record validation in DB after wait time via API")
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, eachInstanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)
				Expect(strings.Contains(responseBody, `"resourceId":"`+eachInstanceResourceId+`"`)).To(BeTrue(), "assert the record in metering db using resource id: %v", responseBody)

				// Validation for new records
				response := string(responseBody)
				responses := strings.Split(response, "\n")
				numofRecords := len(responses)
				logInstance.Println("Number of records: ", numofRecords)
				Expect(numofRecords).To(BeNumerically(">", 1), "Mismatch in number of records found for instance: ", fmt.Sprintf("%s-%d", instanceGroupName, i))

				// Compare the run time of instance
				var allRunningSeconds []float64
				for _, eachResponse := range responses {
					allRunningSeconds = append(allRunningSeconds, gjson.Get(eachResponse, "result.properties.runningSeconds").Float())
				}
				allRunningSeconds = allRunningSeconds[:len(allRunningSeconds)-1]
				Expect(utils.IsArrayInIncreasingOrder(allRunningSeconds)).To(Equal(true))
			}
		})
	})

	When("when the instances are powered ON", func() {
		It("Attempt to Stop a single instance in group", func() {
			// Stop a single instance in group
			logInstance.Println("Updating the run strategy from Always to Halted...")
			statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Halted", sshkeyName, instanceGroupName+"-0")
			Expect(statusCode).To(Equal(200), responseBody)

			// Validation
			instancePhaseValidation := utils.CheckInstanceState(instanceEndpoint, token, listOfIds[0], "Stopped")
			Eventually(instancePhaseValidation, 10*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("when an instance is powered OFF", func() {
		It("Start the instance in group", func() {
			// Start the stopped instance in group
			logInstance.Println("Updating the run strategy from Halted to Always...")
			statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Always", sshkeyName, instanceGroupName+"-0")
			Expect(statusCode).To(Equal(200), responseBody)

			// Validation
			instancePhaseValidation := utils.CheckInstanceState(instanceEndpoint, token, listOfIds[0], "Ready")
			Eventually(instancePhaseValidation, 10*time.Minute, 5*time.Second).Should(BeTrue())
		})
	})

	When("Instance Group Deletion and Validation via API", func() {
		It("Instance Deletion and Validation via API", func() {
			logInstance.Println("Delete all instances present inside group using name...")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceGroupByName(instanceGroupEndpoint, token, instanceGroupName)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Group Deletion using Name")
			instanceGroupValidation := utils.CheckInstanceGroupDeletionByName(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())

			// sleep for deletion to reflect
			time.Sleep(60 * time.Second)

			// validate the deletion report inside metering record.
			for _, eachInstanceResourceId := range listOfIds {
				latestRecord := utils.GetLastMeteringRecord(meteringEndpoint, adminToken, metPayload, cloudAccount, eachInstanceResourceId)
				Expect(strings.Contains(latestRecord, `"deleted":"true"`)).To(BeTrue(), "assertion failed at deletion report inside metering record")
			}
			instanceGroupDeletion = true

			deletionTime := time.Now()
			deletionTimestamp = deletionTime.Format(time.RFC3339)
			logInstance.Println(deletionTimestamp)
		})
	})

	When("Record validation in DB after deletion via metering service", func() {
		It("Record validation in DB after deletion via metering service", func() {
			for _, eachInstanceResourceId := range listOfIds {
				logInstance.Println("Starting validation flow via metering service...")
				time.Sleep(2 * time.Minute)
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, eachInstanceResourceId)
				Expect(statusCode).To(Equal(200), "assert metering record search response code: %v", responseBody)
				Expect(strings.Contains(responseBody, `"deleted":"true"`)).To(BeTrue(), responseBody)
			}
		})
	})

	// Commenting usage check due to a dependency on 'reportUsageSchedulerInterval' in billing configmap
	// The usage is reported every 3600 seconds with respect to the 'reportUsageSchedulerInterval' parameter
	/*
		When("Validate usage for the created BM instance Group", func() {
			It("Validate usage for the created BM instance Group", func() {
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
				Expect(productType).To(Equal(instanceTypeToBeCreatedBMCluster), responseBody)

				// Ensure total usage is > 0
				totalUsage := gjson.Get(responseBody, "totalUsage").Int()
				Expect(totalUsage).To(BeNumerically(">", 0), "Total Usage is not greater than 0.")
			})
		})
	*/

	// billing test case : TODO

	AfterAll(func() {
		if !instanceGroupDeletion {
			// Delete the instance if any testcases above fails
			logInstance.Println("Delete all instances present inside group using name...")
			deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceGroupByName(instanceGroupEndpoint, token, instanceGroupName)
			Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

			// Validation
			logInstance.Println("Validation of Instance Group Deletion using Name")
			instanceGroupValidation := utils.CheckInstanceGroupDeletionByName(instanceEndpoint, token, instanceGroupName)
			Eventually(instanceGroupValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
		}
	})
})
