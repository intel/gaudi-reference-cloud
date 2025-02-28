package metering_monitor

import (
	"encoding/json"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/test_pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

var _ = Describe("BMaas Instance CRUD operation and validating the records in metering service",
	Label("compute", "compute_business_uc", "bm_metering_validation_uc"), Ordered, func() {
		var (
			createRespBody            string
			createStatusCode          int
			instanceResourceId        string
			creationTimestamp         string
			firstReadyTimestamp       string
			bmName                    string
			bmPayload, metPayload     string
			instanceType              string
			instanceCreationTimestamp int64
			adminToken                string
		)

		BeforeAll(func() {
			if skipBMCreation == "true" {
				Skip("Skipping the Entire BM negative flow due to the flag")
			}
			// name and payload creation
			instanceType = utils.GetJsonValue("instanceTypeToBeCreatedMeteringMonitor")
			bmName = "automation-vm-" + utils.GetRandomString()
			bmPayload = utils.GetJsonValue("instancePayload")
			metPayload = utils.GetJsonValue("meteringMonitoringPayload")
			adminToken = utils.FetchAdminToken(testEnv)
		})

		JustBeforeEach(func() {
			logInstance.Println("----------------------------------------------------")
		})

		When("Starting the Instance Creation flow via Instance API", func() {
			It("Starting the Instance Creation flow via Instance API", func() {
				createStatusCode, createRespBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, bmPayload, bmName,
					instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
				Expect(createStatusCode).To(Equal(200), createRespBody)
				Expect(strings.Contains(createRespBody, `"name":"`+bmName+`"`)).To(BeTrue(), createRespBody)
				instanceResourceId = gjson.Get(createRespBody, "metadata.resourceId").String()
				creationTimestamp = gjson.Get(createRespBody, "metadata.creationTimestamp").String()
				logInstance.Println("creationTimestamp: ", creationTimestamp)

				// Validation
				logInstance.Println("Checking whether instance is in ready state")
				instanceValidation := utils.CheckInstancePhase(instanceEndpoint, token, instanceResourceId)
				Eventually(instanceValidation, 20*time.Minute, 30*time.Second).Should(BeTrue())
				instanceCreationTimestamp = time.Now().UnixMilli()
				logInstance.Println(fmt.Sprint(instanceCreationTimestamp))
			})
		})

		When("Record validation in DB after instance creation via metering service", func() {
			It("Record validation in DB after instance creation via metering service instance", func() {
				logInstance.Println("Starting validation flow via metering service after instance creation")
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)
				firstReadyTimestamp = gjson.Get(responseBody, "result.properties.firstReadyTimestamp").String()
				logInstance.Println("firstReadyTimestamp: ", firstReadyTimestamp)
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
				Expect(instanceTypeFromRecord).To(Equal(instanceType))
				Expect(regionFromRecord).To(Equal(region))
				Expect(serviceTypeFromRecord).To(Equal("ComputeAsAService"))

				//TODO: firstReadyTimestamp & runningSeconds (validation needs to be added)

			})
		})

		When("validate first ready timestamp after record insertion by compute metering monitor", func() {
			It("validate first ready timestamp after record insertion by compute metering monitor", func() {
				logInstance.Println("validate first ready timestamp after record insertion by compute metering monitor")
				latestRecord := utils.GetLastMeteringRecord(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)

				//Retrieve the required fields to validate
				firstReadyTimestampFromRecord := gjson.Get(latestRecord, "result.properties.firstReadyTimestamp").String()

				//validate the first ready timestamp
				Expect(utils.ValidateTimeStamp(utils.GetUnixTime(firstReadyTimestampFromRecord), instanceCreationTimestamp-60000, instanceCreationTimestamp+60000)).Should(BeTrue())
			})
		})

		When("when the instance is powered ON", func() {
			It("Attempt to Stop the instance", func() {
				// stop the instance
				logInstance.Println("Updating the run strategy from Always to Halted...")
				statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Halted", sshkeyName, bmName)
				Expect(statusCode).To(Equal(200), responseBody)

				instancePhaseValidation := utils.CheckInstanceState(instanceEndpoint, token, instanceResourceId, "Stopped")
				Eventually(instancePhaseValidation, 10*time.Minute, 5*time.Second).Should(BeTrue())
			})
		})

		When("when the instance is powered OFF", func() {
			It("Start the instance", func() {
				// start the instance
				logInstance.Println("Updating the run strategy from Halted to Always...")
				statusCode, responseBody := utils.UpdateInstanceRunStrategy(instanceEndpoint, token, "Always", sshkeyName, bmName)
				Expect(statusCode).To(Equal(200), responseBody)

				instancePhaseValidation := utils.CheckInstanceState(instanceEndpoint, token, instanceResourceId, "Ready")
				Eventually(instancePhaseValidation, 10*time.Minute, 10*time.Second).Should(BeTrue())
			})
		})

		When("Validate creationTimeStamp in latest instance response", func() {
			It("Validate creationTimeStamp in latest instance response", func() {
				statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)

				latestcreationTimeStamp := gjson.Get(responseBody, "metadata.creationTimestamp").String()
				Expect(latestcreationTimeStamp).To(Equal(creationTimestamp), responseBody)
			})
		})

		When("Validate firstReadyTimestamp after instance runStrategy is updated", func() {
			It("Validate firstReadyTimestamp after instance runStrategy is updated", func() {
				logInstance.Println("Record validation in DB after instance runStrategy is updated")
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)
				response := string(responseBody)
				responses := strings.Split(response, "\n")

				var latestReadyTimeStamp string
				for _, lastRecord := range responses {
					// Skip empty lines
					if lastRecord == "" {
						continue
					}
					// Parse the JSON object
					var data map[string]interface{}
					err := json.Unmarshal([]byte(lastRecord), &data)
					Expect(err).ShouldNot(HaveOccurred())
					latestReadyTimeStamp = gjson.Get(lastRecord, "result.properties.firstReadyTimestamp").String()
				}
				Expect(firstReadyTimestamp).To(Equal(latestReadyTimeStamp), responseBody)

			})
		})

		When("Record Validation in DB after creation via metering service", func() {
			It("Record Validation in DB after creation via metering service", func() {
				// Minimum time limit for record updation as per BMaaS design
				time.Sleep(3 * time.Minute)
			})
		})

		When("Record validation in DB after wait time via API", func() {
			It("Record validation in DB after wait time via API", func() {
				logInstance.Println("Record validation in DB after wait time via API")
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)
				Expect(strings.Contains(responseBody, `"resourceId":"`+instanceResourceId+`"`)).To(BeTrue(), responseBody)
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
			})
		})

		When("Record validation in DB after deletion via metering service", func() {
			It("Record validation in DB after deletion via metering service", func() {
				logInstance.Println("Starting validation flow via metering service after isntance deletion...")
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, adminToken, metPayload, cloudAccount, instanceResourceId)
				Expect(statusCode).To(Equal(200), responseBody)
				Expect(strings.Contains(responseBody, `"deleted":"true"`)).To(BeTrue(), responseBody)
			})
		})

		When("Instance Deletion and Validation via API", func() {
			It("should be successful", func() {
				logInstance.Println("Ensure the created instance is deleted")
			})
		})

		AfterAll(func() {
			// Delete the instance if any testcases above fails
			logInstance.Println("Ensure the created instance is deleted")
			getStatusCode, _ := service_apis.GetInstanceById(instanceEndpoint, token, instanceResourceId)

			if getStatusCode != 404 {
				deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceById(instanceEndpoint, token, instanceResourceId)
				Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

				logInstance.Println("Validation of Instance Deletion using Id")
				instanceValidation := utils.CheckInstanceDeletionById(instanceEndpoint, token, instanceResourceId)
				Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
			} else {
				logInstance.Println("Instance had been deleted")
			}
		})
	})
