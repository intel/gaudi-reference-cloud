package frisby

import (
	"goFramework/framework/frisby_client"
)

func getInstanceType(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Instance type response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetAllInstanceType(instance_type_base_url string, token string) (int, string) {
	get_response_status, get_response_body := getInstanceType(instance_type_base_url, token)
	return get_response_status, get_response_body
}

func GetInstanceTypeByName(instance_type_base_url string, token string, instance_type_name string) (int, string) {
	var get_instance_type_byname_url = instance_type_base_url + "/" + instance_type_name
	get_response_status, get_response_body := getInstanceType(get_instance_type_byname_url, token)
	return get_response_status, get_response_body
}
