package compute_utils

import (
	"fmt"
	"github.com/tidwall/gjson"
	"os"
	"strings"
)

var url string
var cloudaccount_payload string
var compute_url string
var compute_url2 string
var vnet_payload string
var sshpublickey_payload string
var instance_payload string
var enroll_payload string
var product_id_payload string
var metering_search_payload string
var metering_search_payload1 string
var metering_create_payload string
var vnetName string
var consoleUrl string
var staas_payload string
var invalid_metering_create_payload string
var default_region string

func LoadE2EConfig(filepath string, filename string) {
	configData, _ := ConvertFileToString(filepath, filename)
	//fmt.Println("config data is " + configData)
	url = gjson.Get(configData, "baseUrl").String()
	compute_url = gjson.Get(configData, "computeUrl").String()
	compute_url2 = gjson.Get(configData, "computeUrl2").String()
	cloudaccount_payload = gjson.Get(configData, "cloudaccount_payload").String()
	vnet_payload = gjson.Get(configData, "vnet_payload").String()
	vnetName = gjson.Get(configData, "vnetName").String()
	sshpublickey_payload = gjson.Get(configData, "sshpublickey_payload").String()
	instance_payload = gjson.Get(configData, "instance_payload").String()
	enroll_payload = gjson.Get(configData, "cloudaccountenroll_payload").String()
	product_id_payload = gjson.Get(configData, "product_by_id").String()
	metering_create_payload = gjson.Get(configData, "metering_create_payload").String()
	invalid_metering_create_payload = gjson.Get(configData, "invalid_metering_create_payload").String()
	metering_search_payload = gjson.Get(configData, "metering_search_payload").String()
	metering_search_payload1 = gjson.Get(configData, "metering_search_payload1").String()
	consoleUrl = gjson.Get(configData, "consoleUrl").String()
	staas_payload = gjson.Get(configData, "staas_payload").String()
	default_region = gjson.Get(configData, "defaultRegion").String()
}

func GetBaseUrl() string {
	return url
}

func GetComputeUrl() string {
	return compute_url
}

func GetComputeUrlWithRegion() string {
	temp := strings.Replace(compute_url2, "<<env>>", os.Getenv("IDC_ENV"), 1)
	return strings.Replace(temp, "<<region>>", os.Getenv("REGION"), 1)
}

func GetConsoleUrl() string {
	return consoleUrl
}

func GetCloudAccountPayload() string {
	return cloudaccount_payload
}

func GetVnetPayload() string {
	return vnet_payload
}

func GetSSHPayload() string {
	return sshpublickey_payload
}

func GetInstancePayload() string {
	return instance_payload
}

func GetEnrollPayload() string {
	return enroll_payload
}

func GetProductIdPayload() string {
	return product_id_payload
}

func GetMeteringSearchPayload() string {
	return metering_search_payload
}

func GetMeteringSearchPayloadCloudAcc() string {
	return metering_search_payload1
}

func GetVnetName() string {
	return vnetName
}

func GetMeteringCreatePayload() string {
	return metering_create_payload
}

func GetInvalidMeteringCreatePayload() string {
	return invalid_metering_create_payload
}

func GetStaaSPayload(storageSize string, region string) string {
	fmt.Println("Region: ", default_region)
	if region == "" {
		region = default_region
		fmt.Println("default region: ", default_region)
	}
	temp := strings.Replace(staas_payload, "<<volume-size>>", storageSize, 1)
	return strings.Replace(temp, "<<region>>", region, 1)
}
