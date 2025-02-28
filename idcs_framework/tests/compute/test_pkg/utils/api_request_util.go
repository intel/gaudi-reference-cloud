package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var logInstance *logger.CustomLogger

func SetLogger(logger *logger.CustomLogger) {
	logInstance = logger
}

// cloud account creation util without service api
func CreateCloudAccount(cloudaccountUrl string, token string, username string, accountType string) string {
	tid := GetRandomStringWithLimit(12)
	oid := GetRandomStringWithLimit(12)
	name := "compute-cloudaccount-" + GetRandomString()
	payload := fmt.Sprintf(`{"name":"%s","owner":"%s","tid":"%s","oid":"%s","type":"%s"}`, name, username, tid, oid, accountType)
	var payloadMap map[string]interface{}
	json.Unmarshal([]byte(payload), &payloadMap)
	response := client.Post(cloudaccountUrl, token, payloadMap)
	statusCode, responseBody := client.LogRestyInfo(response, "POST API")
	Expect(statusCode).To(Equal(200), "Failed to create Cloud Account: %v", responseBody)
	cloudAccount := gjson.Get(responseBody, "id").String()
	return cloudAccount
}

// update run strategy of an instance (BM) to power ON or OFF the instance
func UpdateInstanceRunStrategy(instanceEndpoint, token, runStrategy, sshkeyName, instanceName string) (int, string) {
	payload := `{
		"spec": {
			"runStrategy": "<<run-strategy>>",
			"sshPublicKeyNames": [
			"<<ssh-publickey-created>>"
			]
		}
	}`
	payload = strings.ReplaceAll(payload, "<<run-strategy>>", runStrategy)
	payload = strings.ReplaceAll(payload, "<<ssh-publickey-created>>", sshkeyName)
	// update the instance via API
	statusCode, responseBody := service_apis.PutInstanceByName(instanceEndpoint, token, instanceName, payload)
	return statusCode, responseBody
}

func CheckInstancePhase(instanceEndpoint string, token string, resourceId string) func() bool {
	startTime := time.Now()
	var failedAt time.Time

	return func() bool {
		statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, resourceId)
		Expect(statusCode).To(Equal(200), responseBody)
		instancePhase := gjson.Get(responseBody, "status.phase").String()
		logInstance.Println("instancePhase: ", instancePhase)

		if instancePhase == "Failed" {
			if failedAt.IsZero() {
				failedAt = time.Now()
			} else if time.Since(failedAt) >= 4*time.Minute { // 4 minute window exceeded, fail the loop
				Expect(false).To(BeTrue(), "Instance did not recover from 'Failed' state within 4 minutes")
			}
			return false // Still within 4 minute window

		} else if instancePhase != "Ready" {
			logInstance.Println("Instance is not in ready state")
			return false

		} else if instancePhase == "Ready" {
			logInstance.Println("Instance is in ready state")
			elapsedTime := time.Since(startTime)
			logInstance.Println("Time took for instance to get to ready state: ", elapsedTime)
			return true
		}

		// This will ensure that if the instance goes to "Failed" state again -
		// the timer will start again from 0
		failedAt = time.Time{} // Reset failedAt if the instance recovers from "Failed" state
		return false
	}
}

func CheckLBPhase(lbInstanceEndpoint string, token string, lbResourceId string) func() bool {
	startTime := time.Now()
	return func() bool {
		statusCode, responseBody := service_apis.GetLBById(lbInstanceEndpoint, token, lbResourceId)
		Expect(statusCode).To(Equal(200), responseBody)
		instancePhase := gjson.Get(responseBody, "status.state").String()
		logInstance.Println("LBInstancePhase: ", instancePhase)
		Expect(instancePhase).ToNot(Equal("Failed"), "LBInstance is in failed state")

		if instancePhase != "Active" {
			logInstance.Println("LB is not in active state")
			return false
		} else {
			logInstance.Println("LB is in active state")
			elapsedTime := time.Since(startTime)
			logInstance.Println("Time took for instance to get to ready state: ", elapsedTime)
			return true
		}
	}
}

func CheckInstanceDeletionById(instanceEndpoint string, token string, resourceId string) func() bool {
	startTime := time.Now()
	return func() bool {
		statusCode, _ := service_apis.GetInstanceById(instanceEndpoint, token, resourceId)
		if statusCode != 404 {
			logInstance.Println("Instance is not yet deleted.")
			return false
		} else {
			logInstance.Println("Instance has been deleted.")
			elapsedTime := time.Since(startTime)
			logInstance.Println("Time took for instance deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckLBDeletionById(lbInstanceEndpoint string, token string, resourceId string) func() bool {
	startTime := time.Now()
	return func() bool {
		statusCode, _ := service_apis.GetLBById(lbInstanceEndpoint, token, resourceId)
		if statusCode != 404 {
			logInstance.Println("LB is not yet deleted.")
			return false
		} else {
			logInstance.Println("LB has been deleted.")
			elapsedTime := time.Since(startTime)
			logInstance.Println("Time took for LB deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckInstanceDeletionByName(instanceEndpoint string, token string, name string) func() bool {
	startTime := time.Now()
	return func() bool {
		statusCode, _ := service_apis.GetInstanceByName(instanceEndpoint, token, name)
		if statusCode != 404 {
			logInstance.Println("Instance is not yet deleted.")
			return false
		} else {
			logInstance.Println("Instance has been deleted.")
			elapsedTime := time.Since(startTime)
			logInstance.Println("Time took for instance deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckSSHDeletionByName(sshkeyEndpoint string, token string, name string) func() bool {
	startTime := time.Now()
	return func() bool {
		statusCode, _ := service_apis.GetSSHKeyByName(sshkeyEndpoint, token, name)
		if statusCode != 404 {
			logInstance.Println("SSH Key is not yet deleted.")
			return false
		} else {
			logInstance.Println("SSH Key has been deleted.")
			elapsedTime := time.Since(startTime)
			logInstance.Println("Time took for SSH key deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckSSHDeletionById(sshkeyEndpoint string, token string, resourceId string) func() bool {
	startTime := time.Now()
	return func() bool {
		statusCode, _ := service_apis.GetSSHKeyById(sshkeyEndpoint, token, resourceId)
		if statusCode != 404 {
			logInstance.Println("SSH Key is not yet deleted.")
			return false
		} else {
			logInstance.Println("SSH Key has been deleted.")
			elapsedTime := time.Since(startTime)
			logInstance.Println("Time took for SSH key deletion: ", elapsedTime)
			return true
		}
	}
}

func CheckInstanceGroupProvisionState(instanceEndpoint string, token string, instanceGroupName string) func() bool {
	startTime := time.Now()
	failedAt := make(map[int]time.Time)

	return func() bool {
		statusCode, responseBody := service_apis.GetInstancesWithinGroup(instanceEndpoint, token, instanceGroupName)
		Expect(statusCode).To(Equal(200), responseBody)
		parsed := gjson.Parse(responseBody)
		items := parsed.Get("items").Array()

		for i, item := range items {
			instancePhase := gjson.Get(item.String(), "status.phase").String()

			if instancePhase == "Failed" {
				if _, exists := failedAt[i]; !exists {
					failedAt[i] = time.Now() // Record the time when the instance entered the Failed state
				} else if time.Since(failedAt[i]) >= 4*time.Minute {
					// 4-minute grace period exceeded, fail the loop
					Expect(false).To(BeTrue(), fmt.Sprintf("Instance %dfailed and did not recover within 4 minutes", i+1))
				}
				logInstance.Println("Instance %d is in Failed state, waiting for recovery...\n", i+1)
				return false

			} else if instancePhase != "Ready" {
				logInstance.Println("Validating Instance: %v. Instances are not in ready state", i+1)
				return false
			}
			// If the instance is in the Ready state, reset its failure tracking
			delete(failedAt, i)
		}

		// All instances are in the Ready state
		logInstance.Println("All instances are in Ready state")
		elapsedTime := time.Since(startTime)
		logInstance.Println("Time took for instances to get to ready state: ", elapsedTime)
		return true
	}
}

func GetMeteringRecords(meteringEndpoint string, token string, payload string, cloudAccount string, resourceId string) func() bool {
	startTime := time.Now()
	return func() bool {
		var meteringPayload = payload
		meteringPayload = strings.Replace(meteringPayload, "<<cloud-account-id>>", cloudAccount, 1)
		meteringPayload = strings.Replace(meteringPayload, "<<resource-id>>", resourceId, 1)
		statusCode, responseBody := service_apis.SearchAllMeteringRecords(meteringEndpoint, token, meteringPayload)
		Expect(statusCode).To(Equal(200), responseBody)

		if responseBody == "" {
			logInstance.Println("Metering record is not yet created.")
			return false
		} else {
			logInstance.Println("Metering record has been created for the instance.")
			elapsedTime := time.Since(startTime)
			logInstance.Println("Time took for metering record creation: ", elapsedTime)
			return true
		}
	}
}

func GetLastMeteringRecord(meteringEndpoint string, token string, payload string, cloudAccount string, instanceResourceId string) string {
	var meteringPayload = payload
	meteringPayload = strings.Replace(meteringPayload, "<<cloud-account-id>>", cloudAccount, 1)
	meteringPayload = strings.Replace(meteringPayload, "<<resource-id>>", instanceResourceId, 1)
	statusCode, responseBody := service_apis.SearchAllMeteringRecords(meteringEndpoint, token, meteringPayload)
	Expect(statusCode).To(Equal(200), responseBody)
	allRecords := strings.Split(string(responseBody), "\n") // split the records
	numofRecords := len(allRecords)
	lastRecord := allRecords[numofRecords-2] //Fetch the last record
	return lastRecord
}

func CheckInstanceGroupDeletionByName(instanceEndpoint string, token string, instanceGroupName string) func() bool {
	startTime := time.Now()
	return func() bool {
		statusCode, responseBody := service_apis.GetInstancesWithinGroup(instanceEndpoint, token, instanceGroupName)
		if statusCode == 200 {
			if responseBody != `{"items":[]}` {
				logInstance.Println("Instance group is not yet deleted.")
				return false
			}
		}
		logInstance.Println("Instance group has been deleted.")
		elapsedTime := time.Since(startTime)
		logInstance.Println("Time took for instance deletion: ", elapsedTime)
		return true
	}
}

func CheckInstanceState(instanceEndpoint, token, resourceId, expectedState string) func() bool {
	startTime := time.Now()
	return func() bool {
		statusCode, responseBody := service_apis.GetInstanceById(instanceEndpoint, token, resourceId)
		Expect(statusCode).To(Equal(200), responseBody)
		instancePhase := gjson.Get(responseBody, "status.phase").String()
		logInstance.Println("instancePhase: ", instancePhase)

		if instancePhase != expectedState {
			logInstance.Println("Instance is not in " + expectedState + " state")
			return false
		} else {
			logInstance.Println("Instance is in " + expectedState + " state")
			elapsedTime := time.Since(startTime)
			logInstance.Println("Time took for instance to get to "+expectedState+" state: ", elapsedTime)
			return true
		}
	}
}

func DeleteMultiSSHKeys(baseUrl string, token string, cloudAccounts []string, names []string) {
	for i := 0; i < len(cloudAccounts); i++ {
		sshEndpoint := fmt.Sprintf("%s%s/sshpublickeys", baseUrl, cloudAccounts[i])
		logInstance.Println("Deleting the SSH keys with name: " + names[i])
		statusCode, responseBody := service_apis.DeleteSSHKeyByName(sshEndpoint, token, names[i])
		Expect(statusCode).To(Equal(200), "Failed to delete sshkey: %s: %s", names[i], responseBody)

		// Validation
		logInstance.Println("Validation of SSH key Deletion using Id")
		DeleteValidationByName := CheckSSHDeletionByName(sshEndpoint, token, names[i])
		Eventually(DeleteValidationByName, 1*time.Minute, 5*time.Second).Should(BeTrue())
	}
	logInstance.Println("All keys have been deleted.")
}

func DeleteMultiVNets(baseUrl string, token string, cloudAccounts []string, names []string) {
	for i := 0; i < len(cloudAccounts); i++ {
		vnetEndpoint := fmt.Sprintf("%s%s/vnets", baseUrl, cloudAccounts[i])
		logInstance.Println("Deleting the vnet with name: " + names[i])
		statusCode, responseBody := service_apis.DeleteVnetByName(vnetEndpoint, token, names[i])
		Expect(statusCode).To(Equal(200), "Failed to delete vnet: %s: %s", names[i], responseBody)
	}
	logInstance.Println("All vnets have been deleted.")
}

func DeleteCloudAccount(cloudaccountUrl string, cloudAccountId string, token string) {
	url := fmt.Sprintf("%s/id/%s", cloudaccountUrl, cloudAccountId)
	logInstance.Println("Deleting the cloud account with id: ", cloudAccountId)
	response := client.Delete(url, token)
	statusCode, responseBody := client.LogRestyInfo(response, "DELETE API")
	Expect(statusCode).To(Equal(200), "Failed to delete Cloud Account: %s", responseBody)
}

// Quota related creation and deletion util
// TODO: machine needs to be removed from the input, fetch the machine image via instanceType (pattern similar to instance creation)
func CreateInstanceWithinQuota(instanceEndpoint string, token string, userQuota int, instancePayload string, instanceType string, sshkeyName string,
	vnet string, machineImage string, availabilityZone string) []string {
	//iterate through the quota and validate
	var instanceNameList []string
	logInstance.Println("Starting the Instance Creation flow via Instance API...")
	for i := 1; i <= userQuota; i++ {
		name := "automation-" + GetRandomString()
		instanceNameList = append(instanceNameList, name)
		statusCode, responseBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, instancePayload, name,
			instanceType, sshkeyName, vnet, machineImage, availabilityZone)
		Expect(statusCode).To(Equal(200), responseBody)
		Expect(strings.Contains(responseBody, `"name":"`+name+`"`)).To(BeTrue(), responseBody)
		time.Sleep(10 * time.Second)
	}
	return instanceNameList
}

// TODO: machine needs to be removed from the input, fetch the machine image via instanceType (pattern similar to instance creation)
func CreateInstanceAfterQuotaLimit(instanceEndpoint string, token string, instancePayload string, instanceType string, sshkeyName string, vnet string, machineImage string, availabilityZone string) {
	//Try to create instance after quota exhaustion
	instanceName := "automation-" + GetRandomString()
	statusCode, responseBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, instancePayload, instanceName, instanceType, sshkeyName, vnet, machineImage, availabilityZone)
	Expect(statusCode).To(Equal(400), responseBody)
	Expect(strings.Contains(responseBody, `"Your account has reached the maximum allowed limit `)).To(BeTrue(), responseBody)
}

func DeleteInstanceWithinQuota(instanceEndpoint string, token string, instanceNameList []string, eventuallyTime time.Duration) {
	// Delete all instances
	for i := 0; i < len(instanceNameList); i++ {
		logInstance.Println("Remove the instance: %v - via DELETE api using name", instanceNameList[i])
		statusCode, responseBody := service_apis.DeleteInstanceByName(instanceEndpoint, token, instanceNameList[i])
		Expect(statusCode).To(Equal(200), responseBody)
		time.Sleep(10 * time.Second)

		// Validation of deletion
		logInstance.Println("Validation of Instance Deletion using Id: %v", instanceNameList[i])
		instanceValidation := CheckInstanceDeletionByName(instanceEndpoint, token, instanceNameList[i])
		Eventually(instanceValidation, eventuallyTime, 5*time.Second).Should(BeTrue())
	}
}

func CreateInstanceMaximumQuota(instanceEndpoint string, token string, instancePayload string, instanceName string,
	instanceType string, sshkeyName string, currentQuota int, maximumQuota int, vnet string, machineImage map[string]string, availabilityZone string) (int, string) {
	var (
		statusCode   int
		responseBody string
		resourceIds  []string
	)
	for currentQuota <= maximumQuota {
		bmName := "automation-bm-" + GetRandomString()
		statusCode, responseBody = service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, instancePayload, bmName, instanceType, sshkeyName, vnet, machineImage, availabilityZone)

		if statusCode == 200 {
			currentQuota += 1
			resourceId := gjson.Get(responseBody, "metadata.resourceId").String()
			resourceIds = append(resourceIds, resourceId)
		} else {
			break
		}
	}

	if statusCode != 500 {
		removeInstancesCreatedForQuotaTest(resourceIds, token, instanceEndpoint)
	}

	return statusCode, responseBody
}

func removeInstancesCreatedForQuotaTest(resourceIds []string, token string, instanceEndpoint string) {
	for _, element := range resourceIds {
		service_apis.DeleteInstanceById(instanceEndpoint, token, element)
	}
}

func SearchMeteringRecords(instanceEndpoint string, token string, payload string, cloudAccountId string, instanceResourceId string) (int, string) {
	var meteringPayload = payload
	meteringPayload = strings.Replace(meteringPayload, "<<cloud-account-id>>", cloudAccountId, 1)
	meteringPayload = strings.Replace(meteringPayload, "<<resource-id>>", instanceResourceId, 1)
	statusCode, responseBody := service_apis.SearchAllMeteringRecords(instanceEndpoint, token, meteringPayload)
	return statusCode, responseBody
}

func InstanceCreationMultiSSHKey(instanceEndpoint string, token string, instancePayload string, instanceName string,
	instanceType string, sshKeys []string, vnetName string, imageMapping map[string]string, availabilityZone string) (int, string) {
	// Replace with multi-keys list
	replacement := strings.Join(sshKeys, `", "`)
	statusCode, responseBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, instancePayload, instanceName, instanceType,
		replacement, vnetName, imageMapping, availabilityZone)
	return statusCode, responseBody
}

func CreateInstancesWithAllImages(instanceEndpoint string, token string, sshkeyName string, vnet string, machineImage string, availabilityZone string,
	instanceType string, instanceCreated chan<- string, duration time.Duration) {
	go func() {
		logInstance.Println("Starting Instance Creation flow via Instance API..")
		name := "test-automation-" + GetRandomString()
		payload := GetJsonValue("instancePayload")
		machineImage = strings.Trim(machineImage, `"`)
		// Instance creation
		statusCode, responseBody := service_apis.CreateInstanceWithoutMIMap(instanceEndpoint, token, payload,
			name, instanceType, sshkeyName, vnet, machineImage, availabilityZone)
		Expect(statusCode).To(Equal(200), responseBody)
		Expect(strings.Contains(responseBody, `"name":"`+name+`"`)).To(BeTrue(), responseBody)
		resourceId := gjson.Get(responseBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := CheckInstancePhase(instanceEndpoint, token, resourceId)
		Eventually(instanceValidation, duration, 10*time.Second).Should(BeTrue())
		instanceCreated <- resourceId
	}()
}

func CreateInstancesAllTypes(instanceEndpoint string, token string, sshkeyName string, vnet string, machineImageMapping map[string]string, availabilityZone string,
	instanceType string, instanceCreated chan<- string, duration time.Duration) {
	go func() {
		// defer GinkgoRecover()
		logInstance.Println("Starting Instance Creation flow via Instance API..")
		instanceName := "automation-instance-" + GetRandomString()
		instancePayload := GetJsonValue("instancePayload")
		statusCode, responseBody := service_apis.CreateInstanceWithMIMap(instanceEndpoint, token, instancePayload,
			instanceName, instanceType, sshkeyName, vnet, machineImageMapping, availabilityZone)
		Expect(statusCode).To(Equal(200), responseBody)
		Expect(strings.Contains(responseBody, `"name":"`+instanceName+`"`)).To(BeTrue(), responseBody)
		resourceId := gjson.Get(responseBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether instance is in ready state")
		instanceValidation := CheckInstancePhase(instanceEndpoint, token, resourceId)
		Eventually(instanceValidation, duration, 10*time.Second).Should(BeTrue())
		instanceCreated <- resourceId
	}()
}

func DeleteInstanceWithId(instanceEndpoint string, token string, resourceId string) {
	logInstance.Println("Remove the instance via DELETE api using resource id")
	statusCode, responseBody := service_apis.DeleteInstanceById(instanceEndpoint, token, resourceId)
	Expect(statusCode).To(Equal(200), responseBody)

	// Validation
	logInstance.Println("Validation of Instance Deletion using Id")
	instanceValidation := CheckInstanceDeletionById(instanceEndpoint, token, resourceId)
	Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
}

func DeleteAllInstances(instanceEndpoint string, token string, resourceId string, instanceDeleted chan<- struct{}) {
	go func() {
		logInstance.Println("Remove the instance via DELETE api using resource id")
		statusCode, responseBody := service_apis.DeleteInstanceById(instanceEndpoint, token, resourceId)
		Expect(statusCode).To(Equal(200), responseBody)

		// Validation
		logInstance.Println("Validation of Instance Deletion using Id")
		instanceValidation := CheckInstanceDeletionById(instanceEndpoint, token, resourceId)
		Eventually(instanceValidation, 5*time.Minute, 30*time.Second).Should(BeTrue())
		instanceDeleted <- struct{}{}
	}()
}

func CreateLBWithMonitorTypes(lbInstanceEndpoint string, token string, lbInstancePayload string, cloudAccount string, listenerPort string,
	monitorType string, instanceResourceId string, sourceIp string, lbCreated chan<- string, duration time.Duration) {
	go func() {
		logInstance.Println("Starting Instance Creation flow via Instance API..")
		lbName := "automation-lb-" + GetRandomString()
		statusCode, responseBody := service_apis.CreateLB(lbInstanceEndpoint,
			token, lbInstancePayload, lbName, cloudAccount, listenerPort, monitorType, instanceResourceId, sourceIp)
		Expect(statusCode).To(Equal(200), responseBody)
		Expect(strings.Contains(responseBody, `"name":"`+lbName+`"`)).To(BeTrue(), responseBody)
		lbResourceId := gjson.Get(responseBody, "metadata.resourceId").String()

		// Validation
		logInstance.Println("Checking whether lb is in ready state")
		lbValidation := CheckLBPhase(lbInstanceEndpoint, token, lbResourceId)
		Eventually(lbValidation, duration, 30*time.Second).Should(BeTrue())
		lbCreated <- lbResourceId
	}()
}

func GetInstanceIdsFromInstanceGroup(instanceEndpoint string, token string, instanceGroupName string, payload string) []string {
	instanceGroupSearchPayload := payload
	instanceGroupSearchPayload = strings.Replace(instanceGroupSearchPayload, "<<instance-group-name>>", instanceGroupName, 1)
	statusCode, responseBody := service_apis.SearchInstances(instanceEndpoint, token, instanceGroupSearchPayload)
	Expect(statusCode).To(Equal(200), responseBody)

	ids := gjson.Get(responseBody, "items.#.metadata.resourceId").Array()
	idList := []string{}
	for _, id := range ids {
		idList = append(idList, id.String())
	}
	return idList
}

func GetInstanceNamesFromInstanceGroup(instanceEndpoint string, token string, instanceGroupName string, payload string) []string {
	instanceGroupSearchPayload := payload
	instanceGroupSearchPayload = strings.Replace(instanceGroupSearchPayload, "<<instance-group-name>>", instanceGroupName, 1)
	statusCode, responseBody := service_apis.SearchInstances(instanceEndpoint, token, instanceGroupSearchPayload)
	Expect(statusCode).To(Equal(200), responseBody)

	names := gjson.Get(responseBody, "items.#.metadata.name").Array()
	idList := []string{}
	for _, name := range names {
		idList = append(idList, name.String())
	}
	return idList
}

func GetAllProducts(url string, token string, productsPayload string, cloudAccountId string) (int, string) {
	var payload = productsPayload
	payload = strings.Replace(payload, "<<cloud-account-id>>", cloudAccountId, 1)
	statusCode, responseBody := service_apis.GetProducts(url, token, payload)
	return statusCode, responseBody
}

func GetProductsList(responseBody string) []string {
	products := gjson.Get(responseBody, "products.#.name").Array()
	productsList := []string{}
	for _, product := range products {
		productsList = append(productsList, product.String())
	}
	return productsList
}

func GetImagesAndProductsList(machineImagesEndpoint string, token string) map[string][]string {
	statusCode, responseBody := service_apis.GetAllMachineImage(machineImagesEndpoint, token)
	Expect(statusCode).To(Equal(200), responseBody)

	result := gjson.Get(responseBody, "items")
	imageInstanceTypesMap := make(map[string][]string)

	result.ForEach(func(_, item gjson.Result) bool {
		images := item.Get("metadata.name").String()
		instanceTypes := item.Get("spec.instanceTypes").Array()
		var imagesList []string
		for _, instanceType := range instanceTypes {
			imagesList = append(imagesList, instanceType.String())
		}
		imageInstanceTypesMap[images] = imagesList
		return true // keep iterating
	})
	return imageInstanceTypesMap // Example: map[image1:[instanceType1 instanceType2]]
}

func FindImageByInstanceType(imagesAndInstanceTypeMap map[string][]string, instanceType string, predefinedTypes map[string][]string) []string {
	var imageNames []string
	// Check for vmaas instance types
	if predefinedImages, exists := predefinedTypes[instanceType]; exists {
		return predefinedImages
	}

	// Normal search for machine images for instance type
	for imageName, instanceTypes := range imagesAndInstanceTypeMap {
		for _, u := range instanceTypes {
			if u == instanceType {
				imageNames = append(imageNames, imageName)
				break
			}
		}
	}
	return imageNames
}

func SSHIntoInstance(responseBody string, ansibleFilesPath string, inventoryFilePath string, yamlFilePath string, sshPrivateKeyPath string) error {
	// Create a temp directory
	tempDir := fmt.Sprintf("tmp_ansible_config_%d", time.Now().UnixNano())
	err := os.Mkdir(tempDir, 0700) // Restrict permissions to owner-only
	Expect(err).ToNot(HaveOccurred(), err)
	defer os.RemoveAll(tempDir)

	// Create a temp inventory file by copying existing one
	tempInventoryFile := filepath.Join(tempDir, "inventory.ini")
	err = CopyFile(inventoryFilePath, tempInventoryFile)
	if err != nil {
		logInstance.Println("Failed to copy inventory file: %v", err)
		return err
	}

	// Fetch the related IPs and usernames
	machineIP, proxyIP, proxyUser, _ := ExtractInterfaceDetailsFromResponse(responseBody)

	inventoryRawData, err := ConvertFileToString(ansibleFilesPath, "inventory_raw.ini")
	Expect(err).ToNot(HaveOccurred(), err)
	inventoryGenerated := EnrichInventoryData(inventoryRawData, proxyIP, proxyUser, machineIP, sshPrivateKeyPath)
	WriteStringToFile(tempDir, "inventory.ini", inventoryGenerated)

	var output, stderr bytes.Buffer
	runCommand := exec.Command("ansible-playbook", "-vvvv", "-i", tempInventoryFile, yamlFilePath)
	runCommand.Stdout = &output
	runCommand.Stderr = &stderr

	err = runCommand.Run()
	if err != nil {
		logInstance.Println("Execution of ansible playbook is not successful: ", stderr.String())
		return err
	}

	// Log the ansible output
	ansibleOutput := strings.Split(output.String(), "\n")
	logInstance.Println(ansibleOutput)

	return nil
}

func SSHIntoInstanceMultiTenancy(responseBody string, secondResponseBody string, ansibleFilesPath string, inventoryFilePath string, yamlFilePath string, sshPrivateKeyPath string) error {
	// Create a temp directory
	tempDir := fmt.Sprintf("tmp_ansible_config_%d", time.Now().UnixNano())
	err := os.Mkdir(tempDir, 0700) // Restrict permissions to owner-only
	Expect(err).ToNot(HaveOccurred(), err)
	defer os.RemoveAll(tempDir)

	// Create a temp inventory file by copying existing one
	tempInventoryFile := filepath.Join(tempDir, "inventory.ini")
	err = CopyFile(inventoryFilePath, tempInventoryFile)
	if err != nil {
		logInstance.Println("Failed to copy inventory file: %v", err)
		return err
	}

	// Fetch the related IPs and usernames
	firstInstanceIP, proxyIP, proxyUser, _ := ExtractInterfaceDetailsFromResponse(responseBody)
	secondInstanceIP := ExtractMachineIPFromResponse(secondResponseBody)

	inventoryRawData, err := ConvertFileToString(ansibleFilesPath, "inventory_raw.ini")
	Expect(err).ToNot(HaveOccurred(), err)
	inventoryGenerated := EnrichInventoryData(inventoryRawData, proxyIP, proxyUser, firstInstanceIP, sshPrivateKeyPath)
	WriteStringToFile(tempDir, "inventory.ini", inventoryGenerated)

	var output bytes.Buffer
	runCommand := exec.Command("ansible-playbook", "-vvv", "-i", tempInventoryFile, yamlFilePath, "-e", "another_vm="+secondInstanceIP)
	runCommand.Stdout = &output
	err = runCommand.Run()
	if err != nil {
		logInstance.Println("Execution of ansible playbook is not successful: ", err)
		return err // Return nil for []string when there is an error
	}

	// Log the ansible output
	ansibleOutput := strings.Split(output.String(), "\n")
	logInstance.Println(ansibleOutput)

	return nil
}
