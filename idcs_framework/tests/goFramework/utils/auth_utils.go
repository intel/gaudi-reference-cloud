package utils

import (
	"goFramework/testify/test_config"

	"github.com/tidwall/gjson"
)

var url string
var compute_api string
var metering_api string
var productcatalog_api string
var billing_api string
var cloudaccount_api string


func LoadAuthNConfig(filepath string) {
	configData, _ := test_config.LoadConfig(filepath)
	url = gjson.Get(configData, "baseUrl").String()
	compute_api = gjson.Get(configData, "compute_api").String()
	metering_api = gjson.Get(configData, "metering_api").String()
	productcatalog_api = gjson.Get(configData, "productcatalog_api").String()
	billing_api = gjson.Get(configData, "billing_api").String()
	cloudaccount_api = gjson.Get(configData, "cloudaccount_api").String()
}

func GetBaseUrl() string {
	return url
}

func GetComputeApiData() string {
	return compute_api
}

func GetMeteringApiData() string {
	return metering_api
}

func GetProductCatalogApiData() string {
	return productcatalog_api
}

func GetBillingApiData() string {
	return billing_api
}

func GetCloudAccountApiData() string {
	return cloudaccount_api
}