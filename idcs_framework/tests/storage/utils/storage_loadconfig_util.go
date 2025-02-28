package utils

import (
	"strings"

	"github.com/tidwall/gjson"
)

var st_testEnv string
var st_url string
var st_cloudaccount string
var st_vnetName string
var storage_payload string
var storage_put_payload string
var bucket_payload string
var principal_payload string
var principal_put_policy_payload string
var principal_put_policy_negative_payload string
var principal_multi_put_policy_payload string
var rule_payload string
var rule_invalid_payload string
var rule_invalid_payload2 string
var st_instancetypetobecreated string
var st_instance_payload string
var st_instance_payload_fs string
var st_machine_image string
var st_filesystem_storage_payload string
var st_filesystem_storage_payload_fast string
var st_instance_group_payload_fs string
var st_instance_group_payload_os string

func LoadStorageTestConfig(filepath, filename, cloudAccount, testEnv, region, instanceType, machineImage string) {
	configData, _ := ConvertFileToString(filepath, filename)

	avzone := region + "a"
	vnet := region + "a-default"

	configData = strings.Replace(configData, "<<test-env>>", testEnv, -1)
	configData = strings.Replace(configData, "<<cloud-account>>", cloudAccount, -1)
	configData = strings.Replace(configData, "<<availability-zone>>", avzone, -1)
	configData = strings.Replace(configData, "<<vnet>>", vnet, -1)
	configData = strings.Replace(configData, "<<vnet-name>>", vnet, -1)
	configData = strings.Replace(configData, "<<instance-type>>", instanceType, -1)
	configData = strings.Replace(configData, "<<machine-image>>", machineImage, -1)

	st_testEnv = gjson.Get(configData, "testEnv").String()
	st_url = gjson.Get(configData, "baseUrl").String()
	st_cloudaccount = gjson.Get(configData, "cloudAccount").String()
	st_vnetName = gjson.Get(configData, "vnetName").String()
	storage_payload = gjson.Get(configData, "storage_creation_playload").String()
	storage_put_payload = gjson.Get(configData, "storage_put_payload").String()
	bucket_payload = gjson.Get(configData, "bucket_creation_payload").String()
	principal_payload = gjson.Get(configData, "principal_creation_payload").String()
	principal_put_policy_payload = gjson.Get(configData, "principal_update_policy_payload").String()
	principal_put_policy_negative_payload = gjson.Get(configData, "principal_update_policy_negative_payload").String()
	principal_multi_put_policy_payload = gjson.Get(configData, "principal_multi_update_policy_payload").String()
	rule_payload = gjson.Get(configData, "rule_creation_payload").String()
	rule_invalid_payload = gjson.Get(configData, "rule_invalid_payload").String()
	rule_invalid_payload2 = gjson.Get(configData, "rule_invalid_payload2").String()
	st_instancetypetobecreated = gjson.Get(configData, "instanceTypeToBeCreated").String()
	st_instance_payload = gjson.Get(configData, "instance_payload").String()
	st_instance_payload_fs = gjson.Get(configData, "instance_payload_fs").String()
	st_machine_image = gjson.Get(configData, "machine_image").String()
	st_filesystem_storage_payload = gjson.Get(configData, "filesystem_creation_playload").String()
	st_filesystem_storage_payload_fast = gjson.Get(configData, "filesystem_creation_playload_fast").String()
	st_instance_group_payload_os = gjson.Get(configData, "instance_group_payload_os").String()
	st_instance_group_payload_fs = gjson.Get(configData, "instance_group_payload_fs").String()
}

func GetStTestEnv() string {
	return st_testEnv
}

func GetStorageBaseUrl() string {
	return st_url
}

func GetStCloudAccount() string {
	return st_cloudaccount
}

func GetStVnetName() string {
	return st_vnetName
}

func GetStoragePayload() string {
	return storage_payload
}
func GetStoragePutPayload() string {
	return storage_put_payload
}

func GetBucketPayload() string {
	return bucket_payload
}

func GetPrincipalPayload() string {
	return principal_payload
}

func GetPrincipalPutPolicyPayload() string {
	return principal_put_policy_payload
}

func GetPrincipalPutPolicyNegativePayload() string {
	return principal_put_policy_negative_payload
}

func GetPrincipalMultiPutPolicyPayload() string {
	return principal_multi_put_policy_payload
}

func GetRulePayload() string {
	return rule_payload
}

func GetInvalidRulePayload() string {
	return rule_invalid_payload
}

func GetInvalidRulePayload2() string {
	return rule_invalid_payload2
}

func GetStInstanceTypeToBeCreated() string {
	return st_instancetypetobecreated
}

func GetStInstancePayload() string {
	return st_instance_payload
}

func GetStInstancePayloadFS() string {
	return st_instance_payload_fs
}

func GetStMachineImage() string {
	return st_machine_image
}

func GetStFilesystemStoragePayload() string {
	return st_filesystem_storage_payload
}

func GetStFilesystemStoragePayloadFast() string {
	return st_filesystem_storage_payload_fast
}

func GetStInstanceGroupPayloadFS() string {
	return st_instance_group_payload_fs
}

func GetStInstanceGroupPayloadOS() string {
	return st_instance_group_payload_os
}
