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

var _ = Describe("VMaas Instance group creation and validation of records in metering service",
	Label("compute", "compute_business_uc", "vm_instance_group_metering_validation_uc"), Ordered, func() {

		var (
			instanceType                   string
			instanceGroupName              string
			instanceGroupPayload           string
			metPayload                     string
			searchPayload                  string
			listOfIds                      []string
			instanceGroupCreationTimestamp int64
			instanceGroupDeletion          = false
			adminToken                     string
		)

		BeforeAll(func() {
			if skipVMClusterCreation == "true" {
				Skip("Skipping the Entire VM negative flow due to the flag")
			}
			// Load details
			instanceType = utils.GetJsonValue("instanceTypeToBeCreated")
			instanceGroupPayload = utils.GetJsonValue("instanceGroupPayload")
			metPayload = utils.GetJsonValue("meteringMonitoringPayload")
			searchPayload = utils.GetJsonValue("instanceSearchPayload")
			instanceGroupName = "automation-ins-group-" + utils.GetRandomString()
			adminToken = utils.FetchAdminToken(testEnv)
		})

		When("Creating instances within cluster or group", func() {
			It("instances should be created within cluster", func() {
				logInstance.Println("Starting the Instance Creation flow via Instance Group API...")
				createStatusCode, createRespBody := service_apis.InstanceGroupCreationWithMIMap(instanceGroupEndpoint, token, instanceGroupPayload,
					instanceGroupName, "2", instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
				Expect(createStatusCode).To(Equal(200), createRespBody)
				Expect(strings.Contains(createRespBody, `"name":"`+instanceGroupName+`"`)).To(BeTrue(), createRespBody)

				// Validation
				logInstance.Println("Checking whether instance is in ready state")
				instanceGroupValidation := utils.CheckInstanceGroupProvisionState(instanceEndpoint, token, instanceGroupName)
				Eventually(instanceGroupValidation, 10*time.Minute, 5*time.Second).Should(BeTrue())

				instanceGroupCreationTimestamp = time.Now().UnixMilli()
				logInstance.Println(fmt.Sprint(instanceGroupCreationTimestamp))
			})
		})

		When("Validate metering records after instance group creation", func() {
			It("Validate metering records after instance group creation", func() {
				logInstance.Println("Starting validation flow via metering service after instance group creation")
				var firstReadyTimestamp string

				// Get list if Ids from the group
				listOfIds = utils.GetInstanceIdsFromInstanceGroup(instanceEndpoint, token, instanceGroupName, searchPayload)
				logInstance.Println("listOfIds ", listOfIds)

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
					Expect(instanceGroupSizeFromRecord).To(Equal("2"))
					Expect(instanceTypeFromRecord).To(Equal(instanceType))
					Expect(regionFromRecord).To(Equal(region))
					Expect(serviceTypeFromRecord).To(Equal("ComputeAsAService"))

					//Retrieve the first ready timestamp and validate
					firstReadyTimestampFromRecord := gjson.Get(latestRecord, "result.properties.firstReadyTimestamp").String()
					Expect(utils.ValidateTimeStamp(utils.GetUnixTime(firstReadyTimestampFromRecord), instanceGroupCreationTimestamp-60000, instanceGroupCreationTimestamp+60000)).Should(BeTrue())
				}

			})
		})

		/*When("Record Validation in DB after creation via metering service", func() {
			It("Record Validation in DB after creation via metering service", func() {
				// Minimum time limit for record updation as per VMaaS design
				time.Sleep(3 * time.Minute)
			})
		})

		When("Record validation in DB after wait time via API", func() {
			It("Record validation in DB after wait time via API", func() {
				logInstance.Println("Record validation in DB after wait time via API")
				statusCode, responseBody := utils.SearchMeteringRecords(meteringEndpoint, token, metPayload, cloudAccount, resource_id)
				Expect(statusCode).To(Equal(200), responseBody)
				Expect(strings.Contains(responseBody, `"resourceId":"`+resource_id+`"`)).To(BeTrue(), responseBody)

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

		When("Instance Group Deletion and Validation via API", func() {
			It("Instance Deletion and Validation via API", func() {
				logInstance.Println("Delete all instances present inside group using name...")
				deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceGroupByName(instanceGroupEndpoint, token, instanceGroupName)
				Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

				// Validation
				logInstance.Println("Validation of Instance Group Deletion using Name")
				instanceGroupValidation := utils.CheckInstanceGroupDeletionByName(instanceEndpoint, token, instanceGroupName)
				Eventually(instanceGroupValidation, 3*time.Minute, 5*time.Second).Should(BeTrue())

				// sleep for deletion to reflect
				time.Sleep(120 * time.Second)

				// validate the deletion report inside metering record.
				for _, eachInstanceResourceId := range listOfIds {
					latestRecord := utils.GetLastMeteringRecord(meteringEndpoint, adminToken, metPayload, cloudAccount, eachInstanceResourceId)
					Expect(strings.Contains(latestRecord, `"deleted":"true"`)).To(BeTrue(), "assertion failed at deletion report inside metering record")
				}
				instanceGroupDeletion = true
			})
		})

		AfterAll(func() {
			if !instanceGroupDeletion {
				// Delete the instance if any testcases above fails
				logInstance.Println("Delete all instances present inside group using name...")
				deleteStatusCode, deleteRespBody := service_apis.DeleteInstanceGroupByName(instanceGroupEndpoint, token, instanceGroupName)
				Expect(deleteStatusCode).To(Equal(200), deleteRespBody)

				// Validation
				logInstance.Println("Validation of Instance Group Deletion using Name")
				instanceGroupValidation := utils.CheckInstanceGroupDeletionByName(instanceEndpoint, token, instanceGroupName)
				Eventually(instanceGroupValidation, 3*time.Minute, 5*time.Second).Should(BeTrue())
			}
		})
	})
