package service_apis

import (
	"encoding/json"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis/utils"
)

func postProductCatalog(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Post(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "POST API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetProducts(get_products_base_url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := postProductCatalog(get_products_base_url, token, jsonMap)
	return response_status, response_body
}

func getUsage(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	//ProductCatalog response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetUsage(get_usage_base_url string, token string, params map[string]string) (int, string) {
	get_usages_url, err := utils.ConstructURL(get_usage_base_url, params)
	if err != nil {
		logInstance.Println("error in constructing url", err)
	}
	logInstance.Println("constructed url: ", get_usages_url)
	// var get_usages_url = get_usage_base_url + "/v1/biling/usages"
	get_response_byid_status, get_response_byid_body := getUsage(get_usages_url, token)
	return get_response_byid_status, get_response_byid_body
}
