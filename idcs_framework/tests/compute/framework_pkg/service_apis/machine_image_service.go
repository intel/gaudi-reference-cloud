package service_apis

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
)

func getMachineImage(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	//Machine Image response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetAllMachineImage(machine_image_base_url string, token string) (int, string) {
	get_response_status, get_response_body := getMachineImage(machine_image_base_url, token)
	return get_response_status, get_response_body
}

func GetMachineImageByName(machine_image_base_url string, token string, machine_image_name string) (int, string) {
	var get_machine_image_byname_url = machine_image_base_url + "/" + machine_image_name
	get_response_status, get_response_body := getMachineImage(get_machine_image_byname_url, token)
	return get_response_status, get_response_body
}
