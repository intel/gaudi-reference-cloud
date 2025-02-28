package vmaasutils

import (
	"goFramework/testify/test_config"

	"github.com/tidwall/gjson"
)

var instance_payload string
var url string
var cloudaccount string
var instance_put_payload string
var sshpublickey_payload string

func LoadVMaaSAPIConfig(filepath string) {
	configData, _ := test_config.LoadConfig(filepath)
	url = gjson.Get(configData, "baseUrl").String()
	cloudaccount = gjson.Get(configData, "cloudAccount").String()
	instance_payload = gjson.Get(configData, "instance_payload").String()
	instance_put_payload = gjson.Get(configData, "instance_put_payload").String()
	sshpublickey_payload = gjson.Get(configData, "sshpublickey_payload").String()
}

func GetBaseUrl() string {
	return url
}

func GetCloudAccount() string {
	return cloudaccount
}

func GetInstancePayload() string {
	return instance_payload
}

func GetInstancePutPayload() string {
	return instance_put_payload
}

func GetSSHPayload() string {
	return sshpublickey_payload
}
