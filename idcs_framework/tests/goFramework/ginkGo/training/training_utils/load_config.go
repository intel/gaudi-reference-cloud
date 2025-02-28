package training_utils

import (
	"goFramework/ginkGo/compute/compute_utils"

	"github.com/tidwall/gjson"
)

var baseUrl string
var baseApiUrl string
var slurmaasManagementUsername string

func LoadConfig(filepath string, filename string) {
	configData, _ := compute_utils.ConvertFileToString(filepath, filename)
	baseUrl = gjson.Get(configData, "baseUrl").String()
	baseApiUrl = gjson.Get(configData, "baseApiUrl").String()
	slurmaasManagementUsername = gjson.Get(configData, "slurmaasManagementUsername").String()
}

func GetBaseUrl() string {
	return baseUrl
}

func GetBaseApiUrl() string {
	return baseApiUrl
}

func GetSlurmaasManagementUsername() string {
	return slurmaasManagementUsername
}
