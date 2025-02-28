package metering_monitor

import (
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("VMaaS Instance CRUD operation and validating the records in metering service",
	Label("compute", "compute_business_uc", "vm_metering_validation_uc"), Ordered, func() {
		var (
			createRespBody            string
			createStatusCode          int
			instanceResourceId        string
			vmName                    string
			vmPayload                 string
			metPayload                string
			instancType               string
			instanceCreationTimestamp int64
			adminToken                string
			instanceDeletion          bool
		)

		BeforeAll(func() {
			if skipVMCreation == "true" {
				Skip("Skipping the Entire VM flow due to the flag")
			}
			// name and payload creation
			vmName = "automation-vm-" + utils.GetRandomString()
			vmPayload = utils.GetJsonValue("instancePayload")
			metPayload = utils.GetJsonValue("meteringMonitoringPayload")
			instancType = utils.GetJsonValue("instanceTypeToBeCreated")
			adminToken = utils.FetchAdminToken(testEnv)
		})

		JustBeforeEach(func() {
			logInstance.Println("----------------------------------------------------")
		})

		When("Starting the Small Instance Creation flow via Instance API", func() {
			It("Starting the Small Instance Creation flow via Instance API", func() {
				createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, vmPayload, vmName, instancType, sshkeyName, vnet, machineImageMapping, availabilityZone)
				Expect(createStatusCode).To(Equal(200), createRespBody)
				Expect(strings.Contains(createRespBody, `"name":"`+vmName+`"`)).To(BeTrue(), createRespBody)
				instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()

				// Validation
				logInstance.Println("Checking whether instance is in ready state")
				instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
				Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())

				instanceCreationTimestamp = time.Now().UnixMilli()
				logInstance.Println(fmt.Sprint(instanceCreationTimestamp))
			})
		})

		When("Record validation in DB after instance creation via metering service", func() {
			It("Record validation in DB after instance creation via metering service", func() {
				logInstance.Println("Starting validation flow via metering service")
				//statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, token, metPayload, cloudAccount, instanceResourceId)
				//Expect(statusCode).To(Equal(200), responseBody)
				getMeteringRecords := utils.GetMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
				Eventually(getMeteringRecords, 5*time.Minute, 5*time.Second).Should(BeTrue())
				//Expect(strings.Contains(responseBody, `"resourceId":"`+instanceResourceId+`"`)).To(BeTrue(), "assert the record in metering db using resource id")
			})
		})

		When("Additional metering field validation after record insertion by compute metering monitor", func() {
			It("Additional metering field validation after record insertion by compute metering monitor", func() {
				logInstance.Println("Additional metering field validation after record insertion by compute metering monitor")
				latestRecord := utils.GetLastMeteringRecord(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
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
				Expect(resourceIdFromRecord).To(Equal(instanceResourceId))
				Expect(availabilityZoneFromRecord).To(Equal(availabilityZone))
				Expect(deletedFromRecord).To(Equal("false"))
				Expect(instanceGroupFromRecord).To(Equal(""))
				Expect(instanceGroupSizeFromRecord).To(Equal("1"))
				Expect(instanceTypeFromRecord).To(Equal(instancType))
				Expect(regionFromRecord).To(Equal(region))
				Expect(serviceTypeFromRecord).To(Equal("ComputeAsAService"))
			})
		})

		When("validate first ready timestamp after record insertion by compute metering monitor", func() {
			It("validate first ready timestamp after record insertion by compute metering monitor", func() {
				logInstance.Println("validate first ready timestamp after record insertion by compute metering monitor")
				latestRecord := utils.GetLastMeteringRecord(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
				logInstance.Println(latestRecord)

				//Retrieve the required fields to validate
				firstReadyTimestampFromRecord := gjson.Get(latestRecord, "result.properties.firstReadyTimestamp").String()
				Expect(utils.ValidateTimeStamp(utils.GetUnixTime(firstReadyTimestampFromRecord), instanceCreationTimestamp-60000, instanceCreationTimestamp+60000)).Should(BeTrue())
			})
		})

		// Enable when there is test specific deployment happens
		/*When("Record Validation in DB after creation via metering service", func() {
			It("Record Validation in DB after creation via metering service", func() {
				// Minimum time limit for record updation as per VMaaS design
				time.Sleep(4 * time.Minute)
			})
		})

		When("Record validation in DB after wait time via API", func() {
			It("Record validation in DB after wait time via API", func() {
				logInstance.Println("Record validation in DB after wait time via API")
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, token, metPayload, cloudAccount, instanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)
				Expect(strings.Contains(responseBody, `"resourceId":"`+instanceResourceId+`"`)).To(BeTrue(), "assert the record in metering db using resource id")

				// Validation for new records
				response := string(responseBody)
				responses := strings.Split(response, "\n")
				numofRecords := len(responses)
				logInstance.Println("Number of records: ", numofRecords)
				Expect(numofRecords).To(BeNumerically(">", 2), "Mismatch in number of records found.")

				// Compare the run time of instance
				var allRunningSeconds []float64
				for _, eachResponse := range responses {
					allRunningSeconds = append(allRunningSeconds, gjson.Get(eachResponse, "result.properties.runningSeconds").Float())
				}
				allRunningSeconds = allRunningSeconds[:len(allRunningSeconds)-1]
				Expect(utils.IsArrayInIncreasingOrder(allRunningSeconds)).To(Equal(true))
			})
		})*/

		When("Instance Deletion and Validation via API", func() {
			It("Instance Deletion and Validation via API", func() {
				logInstance.Println("Deletion and Validation of Instance via API")
				deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
				Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

				// Validation
				logInstance.Println("Validation of Instance Deletion")
				instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
				Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
				instanceDeletion = true
			})
		})

		When("Record validation in DB after deletion via metering service", func() {
			It("Record validation in DB after deletion via metering service", func() {
				logInstance.Println("Starting deletion validation via metering service...")
				time.Sleep(2 * time.Minute)
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)
				Expect(strings.Contains(responseBody, `"deleted":"true"`)).To(BeTrue(), responseBody)
			})
		})

		AfterAll(func() {
			if !instanceDeletion {
				logInstance.Println("Deletion and Validation of Instance via API")
				deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
				Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

				// Validation
				logInstance.Println("Validation of Instance Deletion")
				instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
				Eventually(instanceValidation, 5*time.Minute, 5*time.Second).Should(BeTrue())
			}
		})
	})
