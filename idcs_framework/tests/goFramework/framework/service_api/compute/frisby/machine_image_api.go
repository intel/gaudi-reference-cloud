package frisby

import (
	"goFramework/framework/frisby_client"
)

func getMachineImage(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
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
