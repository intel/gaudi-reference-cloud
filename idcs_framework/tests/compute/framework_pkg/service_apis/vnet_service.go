package service_apis

import (
	"encoding/json"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"strings"
)

func getVnet(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createVnet(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Post(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "POST API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteVnet(url string, token string) (int, string) {
	response := client.Delete(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "DELETE API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetVnetById(vnet_api_endpoint string, token string, vnet_id string) (int, string) {
	var get_vnetbyid_endpoint = vnet_api_endpoint + "/id/" + vnet_id
	responseCode, responseBody := getVnet(get_vnetbyid_endpoint, token)
	return responseCode, responseBody
}

func GetVnetByName(vnet_api_endpoint string, token string, vnet_name string) (int, string) {
	var get_vnetbyname_endpoint = vnet_api_endpoint + "/name/" + vnet_name
	responseCode, responseBody := getVnet(get_vnetbyname_endpoint, token)
	return responseCode, responseBody
}

func GetAllVnet(vnet_api_endpoint string, token string) (int, string) {
	responseCode, responseBody := getVnet(vnet_api_endpoint, token)
	return responseCode, responseBody
}

func DeleteVnetById(vnet_api_endpoint string, token string, vnet_id string) (int, string) {
	var delete_vnetbyid_endpoint = vnet_api_endpoint + "/id/" + vnet_id
	responseCode, responseBody := deleteVnet(delete_vnetbyid_endpoint, token)
	return responseCode, responseBody
}

func DeleteVnetByName(vnet_api_endpoint string, token string, vnet_name string) (int, string) {
	var delete_vnetbyname_endpoint = vnet_api_endpoint + "/name/" + vnet_name
	responseCode, responseBody := deleteVnet(delete_vnetbyname_endpoint, token)
	return responseCode, responseBody
}

func CreateVNet(vnet_api_endpoint string, token string, rawpayload string, vnet_name string) (int, string) {
	vnet_payload := rawpayload
	vnet_payload = strings.Replace(vnet_payload, "<<vnet-name>>", vnet_name, 1)

	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(vnet_payload), &jsonMap)
	response_status, response_body := createVnet(vnet_api_endpoint, token, jsonMap)
	return response_status, response_body
}

func CreateVnetWithCustomizedPayload(vnet_api_endpoint string, token string, vnet_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(vnet_payload), &jsonMap)
	response_status, response_body := createVnet(vnet_api_endpoint, token, jsonMap)
	return response_status, response_body
}

/*func updateVM(vm_name string, vmta string) {

}*/
