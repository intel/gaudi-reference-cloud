package testsetup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/utils"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

var Testdata map[string]UserData

var ProductUsageData map[string]UsageData

// VMaaS utilities

// Instance

func Validate_Instance_Creation_Response(responseBody string) error {
	var structResponse InstanceCreateResponseStruct
	flag := utils.CompareJSONToStruct([]byte(responseBody), structResponse)
	if !flag {
		return fmt.Errorf("\n schema Validation Failed for Instance Create response : %s", &responseBody)
	}
	return nil
}

func CreateVmInstance(data string, url1 string, baseUrl string, authToken string) error {
	//baseUrl := gjson.Get(ConfigData, "regionalUrl").String()
	//url1 := gjson.Get(ConfigData, "baseUrl").String()
	userName := gjson.Get(data, "userName").String()
	logger.Log.Info("Create instance for user " + userName)
	sshKeyName := gjson.Get(data, "sshKeyName").String()
	vnetName := gjson.Get(data, "vnetName").String()
	machineImage := gjson.Get(data, "machineImage").String()
	instanceType := gjson.Get(data, "instanceType").String()
	instance_name := gjson.Get(data, "name").String()
	cloud_account_created, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	instance_endpoint := url1 + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
	//instance_name := "AutoVMDev-" + utils.GenerateSSHKeyName(4)
	instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), instance_name, instanceType, machineImage, sshKeyName, vnetName)
	//authToken := Get_OIDC_Admin_Token()
	create_instance_response_status, create_instance_response_body := frisby.CreateInstance(instance_endpoint, authToken, instance_payload)
	if create_instance_response_status != 200 {
		return fmt.Errorf("\n instance Creation Failed for user : %s, Got Response Code : %d, and Response : %s", &userName, create_instance_response_status, create_instance_response_body)
	}
	if instance_name != gjson.Get(create_instance_response_body, "metadata.name").String() {
		return fmt.Errorf("\n instance Creation Failed for user : %s, Got Response Code : %d, and Response : %s", &userName, create_instance_response_status, create_instance_response_body)
	}
	// Validate response
	err = Validate_Instance_Creation_Response(create_instance_response_body)
	if err != nil {
		return err
	}
	//instanceId := gjson.Get(create_instance_response_body, "metadata.resourceId").String()
	//err = ValidateCreatedInstance(userName, cloud_account_created, instanceId)

	//Fetch and store product data

	filter := ProductFilter{
		Name: instanceType,
	}

	product_filter, _ := json.Marshal(filter)
	response_status, response_body := financials.GetProducts(baseUrl, authToken, string(product_filter))
	var structResponse GetProductsResponse
	json.Unmarshal([]byte(response_body), &structResponse)
	logger.Logf.Infof("structResponse", structResponse)
	if response_status != 200 {
		return fmt.Errorf("\n failed to retrieve Product Details : %s", instanceType)
	}
	var productRate string
	var usageExp string
	var unit string
	for i := 0; i < len(structResponse.Products[0].Rates); i++ {
		if structResponse.Products[0].Rates[i].AccountType == Testdata[userName].AccType {
			productRate = structResponse.Products[0].Rates[i].Rate
			usageExp = structResponse.Products[0].Rates[i].UsageExpr
			unit = structResponse.Products[0].Rates[i].Unit
		}
	}
	productData := Product{
		Highlight:    structResponse.Products[0].Metadata.Highlight,
		Disks:        structResponse.Products[0].Metadata.Disks,
		InstanceType: structResponse.Products[0].Metadata.InstanceType,
		Memory:       structResponse.Products[0].Metadata.Memory,
		Region:       structResponse.Products[0].Metadata.Region,
		Rate:         productRate,
		UsageExp:     usageExp,
		Unit:         unit,
	}

	vmData := VmStruct{
		InstanceName:   instance_name,
		SshKeyname:     sshKeyName,
		InstanceId:     gjson.Get(create_instance_response_body, "metadata.resourceId").String(),
		Vnet:           vnetName,
		CloudAccountId: Testdata[userName].CloudAccountId,
		CreationTime:   strconv.FormatInt(time.Now().UnixMilli(), 10),
		Product:        productData,
	}
	if val, ok := Testdata[userName]; ok {
		val.Vms = append(val.Vms, vmData)
		Testdata[userName] = val
	}
	logger.Logf.Infof("\nTestdata", Testdata)
	logger.Log.Info("Create instance completed successfully for user " + userName)
	return nil
}

func ValidateCreatedInstancesAll(userName string, instanceId string, baseUrl string, authToken string) error {
	//baseUrl := gjson.Get(ConfigData, "regionalUrl").String()
	for _, v := range Testdata {
		logger.Log.Info("Validate created instances for user " + userName)
		vmData := v.Vms
		for i := 0; i < len(vmData); i++ {
			cloudAccountId := vmData[i].CloudAccountId
			instanceId := vmData[i].InstanceId
			instance_endpoint := baseUrl + "/v1/cloudaccounts/" + cloudAccountId + "/instances"
			authToken := Get_OIDC_Admin_Token()
			response_status, response_body := frisby.GetInstanceById(instance_endpoint, authToken, instanceId)
			if response_status != 200 {
				return fmt.Errorf("\n failed to retrieve VM instance,  Error: %s", response_body)
			}
			phase := gjson.Get(response_body, "phase").String()
			if phase != "Ready" {
				return fmt.Errorf("\n instance %s is not in Ready state, Expecte: Ready, Actual:%s", vmData[i].InstanceName, phase)
				vmData[i].MachineIp = "VM not in ready state"

			} else {
				vmData[i].MachineIp = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()
			}

		}

	}
	logger.Log.Info(" Successfully Validated created instances for user, All instances are up and running " + userName)
	return nil
}

func ValidateCreatedInstance(userName string, cloudAccountId string, instanceId string, baseUrl string, authToken string) error {
	//baseUrl := gjson.Get(ConfigData, "regionalUrl").String()
	instance_endpoint := baseUrl + "/v1/cloudaccounts/" + cloudAccountId + "/instances"
	//authToken := Get_OIDC_Admin_Token()
	var machineIP string
	response_status, response_body := frisby.GetInstanceById(instance_endpoint, authToken, instanceId)
	if response_status != 200 {
		return fmt.Errorf("\n failed to retrieve VM instance,  Error: %s", response_body)
	}
	phase := gjson.Get(response_body, "phase").String()
	if phase != "Ready" {
		return fmt.Errorf("\n instance %s is not in Ready state, Expecte: Ready, Actual:%s", instanceId, phase)
		machineIP = "VM not in ready state"

	} else {
		machineIP = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()
	}

	if val, ok := Testdata[userName]; ok {
		vmData := val.Vms
		for i := 0; i < len(vmData); i++ {
			if instanceId == vmData[i].InstanceId {
				vmData[i].MachineIp = machineIP
			}
		}
	}
	logger.Log.Info("Successfully Validated created instances for user, All instances are up and running " + userName)
	return nil
}

func SSHintoInstance() {
	for _, v := range Testdata {
		vmData := v.Vms
		for i := 0; i < len(vmData); i++ {
			inventory_raw_data, _ := compute_utils.ConvertFileToString("../../ansible-files", "inventory_raw.ini")
			logger.Logf.Infof("Inventory Raw Data is :" + inventory_raw_data)
			proxyIP := gjson.Get(ConfigData, "proxyIP").String()
			inventory_generated := compute_utils.EnrichInventoryData(inventory_raw_data, proxyIP, "guest", vmData[i].MachineIp, "~/.ssh/id_rsa")
			logger.Logf.Infof("Inventory generated is :" + inventory_generated)
			compute_utils.WriteStringToFile("../../ansible-files", "inventory.ini", inventory_generated)
			// Get the pod details after restart
			var output bytes.Buffer
			get_pod_cmd := exec.Command("ansible-playbook", "-i", "../../ansible-files/inventory.ini", "../../ansible-files/ssh-and-apt-get-on-vm.yml")
			get_pod_cmd.Stdout = &output
			error := get_pod_cmd.Run()
			if error != nil {
				logger.Logf.Infof("Execution of ansible playbook is not successful: ", error)
			}
			// Log the ansible output
			ansible_output := strings.Split(output.String(), "\n")
			logger.Logf.Infof("Ansible OUtput ", ansible_output)
			// keeping the sleep time for billing purpose
			time.Sleep(300 * time.Second)
		}
	}
}

// Vnet Utils

func Validate_Vnet_Creation_Response(responseBody string) error {
	var structResponse VnetCreationResponse
	flag := utils.CompareJSONToStruct([]byte(responseBody), structResponse)
	if !flag {
		return fmt.Errorf("\n schema Validation Failed for Vnet Create response : %s", &responseBody)
	}
	return nil
}

func CreateVnet(data string, computeUrl string, baseUrl string, authToken string) error {
	//baseUrl := gjson.Get(ConfigData, "regionalUrl").String()
	userName := gjson.Get(data, "userName").String()
	logger.Log.Info(" Start VNt creation for user " + userName)
	cloud_account_created, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	//authToken := Get_OIDC_Admin_Token()
	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
	vnet_name := "AutoVnetDev-" + utils.GenerateSSHKeyName(4)
	vnet_payload := gjson.Get(data, "payload").String()
	vnet_payload = compute_utils.EnrichVnetPayload(vnet_payload, vnet_name)
	vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, authToken, vnet_payload)
	if vnet_creation_status != 200 {
		return fmt.Errorf("\n vnet Creation Failed for user : %s, Got Response Code : %d, and Response : %s", &userName, vnet_creation_status, vnet_creation_body)
	}

	// Validate response
	err = Validate_Vnet_Creation_Response(vnet_creation_body)
	if err != nil {
		return err
	}
	logger.Log.Info("VNet creation completed successfully !!!" + userName)
	return nil
}

// ssh Key Creation

func Validate_SShKey_Creation_Response(responseBody string) error {
	var structResponse SSHKeyCreationResponse
	flag := utils.CompareJSONToStruct([]byte(responseBody), structResponse)
	if !flag {
		return fmt.Errorf("\n schema Validation Failed for Coupon Create response : %s", &responseBody)
	}
	return nil
}

func CreateSSHKey(data string, computeUrl string, baseUrl string, authToken string) error {
	//baseUrl := gjson.Get(ConfigData, "regionalUrl").String()
	userName := gjson.Get(data, "userName").String()
	sshPublicKey := gjson.Get(data, "sshKey").String()
	sshkey_name := gjson.Get(data, "sshKeyName").String()
	logger.Log.Info("Start SSH Key creation for user " + userName)
	//cloud_account_created := "263017187833"
	cloud_account_created, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	//authToken := Get_OIDC_Admin_Token()
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + cloud_account_created + "/sshpublickeys"
	//sshkey_name := "autosshdev-" + utils.GenerateSSHKeyName(4)
	logger.Log.Info("ssh Key end point:" + ssh_publickey_endpoint)
	ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
	// hit the api
	logger.Log.Info("ssh Key payload " + ssh_publickey_payload)
	sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, authToken, ssh_publickey_payload)
	logger.Log.Info("ssh Key creation response" + sshkey_creation_body)
	ssh_publickey_name_created := gjson.Get(sshkey_creation_body, "metadata.name").String()
	ssh_cloudAccId := gjson.Get(sshkey_creation_body, "metadata.cloudAccountId").String()
	if ssh_publickey_name_created != sshkey_name {
		return fmt.Errorf("\n ssh Key name validation failed for user : %s , Expected :%s, Actual : %s", userName, sshkey_name, ssh_publickey_name_created)
	}
	if ssh_cloudAccId != cloud_account_created {
		return fmt.Errorf("\n ssh Key cloudAccountId validation failed for user : %s , Expected :%s, Actual : %s", userName, sshkey_name, ssh_publickey_name_created)
	}
	if sshkey_creation_status != 200 {
		return fmt.Errorf("\n ssh Key Creation Failed for user : %s, Got Response Code : %d, and Response : %s", &userName, sshkey_creation_status, sshkey_creation_body)
	}

	// Validate response
	err = Validate_SShKey_Creation_Response(sshkey_creation_body)
	if err != nil {
		return err
	}
	logger.Log.Info("SSH Key creation completed successfully !!!" + userName)
	return nil
}

// func GetInstanceStartTime(userName string, cloudAccountId string, instanceId string, baseUrl string, authToken string) error {
// 	//baseUrl := gjson.Get(ConfigData, "regionalUrl").String()
// 	var runningTime time.Duration
// 	instance_endpoint := baseUrl + "/v1/cloudaccounts/" + cloudAccountId + "/instances"
// 	//authToken := Get_OIDC_Admin_Token()
// 	var machineIP string
// 	response_status, response_body := frisby.GetInstanceById(instance_endpoint, userToken, instanceId)
// 	if response_status != 200 {
// 		return fmt.Errorf("\n failed to retrieve VM instance,  Error: %s", response_body)
// 	}
// 	firstReadyTimestamp := gjson.Get(response_body, "metadata.creationTimestamp").String()
// 	runningTime = timestamppb.Now().AsTime().Sub(firstReadyTimestamp.Time)
// 	return nil
// }

func GetCloudAccountId(userName string, url1 string, authToken string) (string, error) {
	//url1 := gjson.Get(ConfigData, "baseUrl").String()
	//authToken := Get_OIDC_Admin_Token()
	url := url1 + "/v1/cloudaccounts"
	logger.Log.Info("Get cloud Account ID : " + url)
	response_status, response_body := financials.GetCloudAccountByName(url, authToken, userName)
	if response_status != 200 {
		return "", fmt.Errorf("\n get on cloud account to fetch id failed : %s, Got Response Code : %d, and Response : %s", userName, response_status, response_body)
	}
	cloudAccId := gjson.Get(response_body, "id").String()
	return cloudAccId, nil
}

func GetCloudAccountUserName(cloudAccId string, url1 string, authToken string) (string, error) {
	//url1 := gjson.Get(ConfigData, "baseUrl").String()
	//authToken := Get_OIDC_Admin_Token()
	url := url1 + "/v1/cloudaccounts"
	logger.Log.Info("Get cloud Account ID : " + url)
	logger.Log.Info("Get cloud Account ID : " + cloudAccId)
	response_status, response_body := financials.GetCloudAccountById(url, authToken, cloudAccId)
	if response_status != 200 {
		return "", fmt.Errorf("\n get on cloud account to fetch id failed : %s, Got Response Code : %d, and Response : %s", cloudAccId, response_status, response_body)
	}
	userName := gjson.Get(response_body, "name").String()
	return userName, nil
}

// Delete compute data

func DeleteInstance(instanceName string, userName string, computeUrl string, baseUrl string, authToken string) error {
	logger.Log.Info("Delete Instance for user " + userName)
	//baseUrl := gjson.Get(ConfigData, "regionalUrl").String()
	//authToken := Get_OIDC_Admin_Token()
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	instance_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/instances"
	response_status, response_body := frisby.GetInstanceByName(instance_endpoint, authToken, instanceName)
	if response_status != 200 {
		return fmt.Errorf("\n failed to retrieve VM instance,  Error: %s", response_body)
	}

	time.Sleep(10 * time.Second)
	instance_id_created := gjson.Get(response_body, "metadata.resourceId").String()
	// delete the instance created
	delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, authToken, instance_id_created)
	if delete_response_status != 200 {
		logger.Logf.Infof("Response code for delete instance did not match 200, :", delete_response_status)
		return fmt.Errorf("\n instance deletion failed : %s", instanceName)
	}
	time.Sleep(5 * time.Second)

	// validate the deletion
	// Adding a sleep because it seems to take some time to reflect the deletion status
	time.Sleep(1 * time.Minute)
	get_response_status, _ := frisby.GetInstanceById(instance_endpoint, authToken, instance_id_created)
	if get_response_status != 404 {
		logger.Logf.Infof("Response code for delete instance did not match 404 ,  : %d", delete_response_status)
		return fmt.Errorf("\n instance deletion failed : %s", instanceName)
	}
	logger.Log.Info("Instance Deleted. Completed Instance Deletion for user " + userName)
	return nil

}

func DeleteSSHKey(sshKey string, userName string, computeUrl string, baseUrl string, authToken string) error {
	logger.Log.Info("Delete ssh Key for user" + userName)
	//baseUrl := gjson.Get(ConfigData, "regionalUrl").String()
	//authToken := Get_OIDC_Admin_Token()
	logger.Logf.Infof("Delete the ssh-Public-Key Created above...")
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	ssh_publickey_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/" + "sshpublickeys"

	delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, authToken, sshKey)

	if delete_response_byname_status != 200 {
		return fmt.Errorf("\n ssh Key deletion failed : %s", sshKey)
	}
	logger.Log.Info("SSh Key Deleted. Completed SSH Key Deletion for user " + userName)
	return nil

}

func DeleteVnet(Vnet string, userName string, computeUrl string, baseUrl string, authToken string) error {
	logger.Log.Info("Delete VNet for user " + userName)
	//baseUrl := gjson.Get(ConfigData, "regionalUrl").String()
	//authToken := Get_OIDC_Admin_Token()
	logger.Logf.Infof("Delete the ssh-Public-Key Created above...")
	cloudAccId, err := GetCloudAccountId(userName, baseUrl, authToken)
	if err != nil {
		return err
	}
	logger.Logf.Infof("Delete the Vnet Created above...")
	vnet_endpoint := computeUrl + "/v1/cloudaccounts/" + cloudAccId + "/vnets"
	// Deletion of vnet via name
	delete_response_byname_status, _ := frisby.DeleteVnetByName(vnet_endpoint, authToken, Vnet)
	if delete_response_byname_status != 200 {
		return fmt.Errorf("\n VNet deletion failed : %s", Vnet)
	}
	logger.Log.Info("VNet Deleted. Completed VNet Deletion for user " + userName)

	return nil
}
