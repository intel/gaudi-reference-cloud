package service_apis

import (
	"encoding/json"
	"strings"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis/utils"
)

var logInstance *logger.CustomLogger

func SetLogger(logger *logger.CustomLogger) {
	logInstance = logger
}

func getInstance(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createInstance(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Post(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "POST API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func patchInstance(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Patch(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "PATCH API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteInstance(url string, token string) (int, string) {
	response := client.Delete(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "DELETE API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func putInstance(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Put(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "PUT API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetInstanceById(instance_api_base_url string, token string, instance_id string) (int, string) {
	var get_byinstance_id_url = instance_api_base_url + "/id/" + instance_id
	get_response_byid_status, get_response_byid_body := getInstance(get_byinstance_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetInstanceByName(instance_api_base_url string, token string, instance_name string) (int, string) {
	var get_byinstance_name_url = instance_api_base_url + "/name/" + instance_name
	get_response_byname_status, get_response_byname_body := getInstance(get_byinstance_name_url, token)
	return get_response_byname_status, get_response_byname_body
}

func GetInstancesWithinGroup(instance_api_base_url string, token string, instance_group_name string) (int, string) {
	var get_byinstance_name_url = instance_api_base_url + "?metadata.instanceGroup=" + instance_group_name
	get_response_byname_status, get_response_byname_body := getInstance(get_byinstance_name_url, token)
	return get_response_byname_status, get_response_byname_body
}

// TODO : implement full functionality
/*func GetInstancesWithLabel(instance_api_base_url string, token string, instance_name string) (int, string) {
	var get_byinstance_name_url = instance_api_base_url + "search" + instance_name
	get_response_byname_status, get_response_byname_body := getInstance(get_byinstance_name_url, token)
	return get_response_byname_status, get_response_byname_body
}*/

func GetAllInstance(instance_api_base_url string, token string, params map[string]string) (int, string) {
	url, err := utils.ConstructURL(instance_api_base_url, params)
	if err != nil {
		logInstance.Println("error in constructing url", err)
	}
	logInstance.Println("constructed url: ", url)
	get_allinstances_response_status, get_allinstances_response_body := getInstance(url, token)
	return get_allinstances_response_status, get_allinstances_response_body
}

func PutInstanceById(instance_api_base_url string, token string, instance_id string, payload string) (int, string) {
	var put_byinstance_id_url = instance_api_base_url + "/id/" + instance_id
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byid_body, put_response_byid_status := putInstance(put_byinstance_id_url, token, jsonMap)
	return put_response_byid_body, put_response_byid_status
}

func PutInstanceByName(instance_api_base_url string, token string, instance_name string, payload string) (int, string) {
	var put_byinstance_name_url = instance_api_base_url + "/name/" + instance_name
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byname_body, put_response_byname_status := putInstance(put_byinstance_name_url, token, jsonMap)
	return put_response_byname_body, put_response_byname_status
}

func DeleteInstanceById(instance_api_base_url string, token string, instance_id string) (int, string) {
	var delete_byinstance_id_url = instance_api_base_url + "/id/" + instance_id
	delete_response_byid_body, delete_response_byid_status := deleteInstance(delete_byinstance_id_url, token)
	return delete_response_byid_body, delete_response_byid_status
}

func DeleteInstanceByName(instance_api_base_url string, token string, instance_name string) (int, string) {
	var delete_byinstance_name_url = instance_api_base_url + "/name/" + instance_name
	delete_response_byname_body, delete_response_byname_status := deleteInstance(delete_byinstance_name_url, token)
	return delete_response_byname_body, delete_response_byname_status
}

func SearchInstances(instance_api_base_url string, token string, instance_api_payload string) (int, string) {
	instance_api_search_url := instance_api_base_url + "/search"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_api_payload), &jsonMap)
	response_status, response_body := createInstance(instance_api_search_url, token, jsonMap)
	return response_status, response_body
}

func CreateInstanceWithMIMap(instance_api_url string, token string, instance_payload string, instance_name string, instance_type string,
	sshkey_name string, vnet_name string, imageMapping map[string]string, availabilityZone string) (int, string) {
	var instance_api_payload string
	// Lookup default image based on creation_type
	defaultImage := imageMapping[instance_type]
	instance_api_payload = enrichInstancePayload(instance_payload, instance_name, instance_type, defaultImage, sshkey_name, vnet_name, availabilityZone)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_api_payload), &jsonMap)
	response_status, response_body := createInstance(instance_api_url, token, jsonMap)
	return response_status, response_body
}

func CreateInstanceWithoutMIMap(instance_api_url string, token string, instance_payload string, instance_name string, instance_type string,
	sshkey_name string, vnet_name string, machine_image string, availabilityZone string) (int, string) {
	instance_api_payload := enrichInstancePayload(instance_payload, instance_name, instance_type, machine_image, sshkey_name, vnet_name, availabilityZone)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_api_payload), &jsonMap)
	response_status, response_body := createInstance(instance_api_url, token, jsonMap)
	return response_status, response_body
}

func CreateInstanceWithCustomizedPayload(instance_api_base_url string, token string, instance_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_api_payload), &jsonMap)
	response_status, response_body := createInstance(instance_api_base_url, token, jsonMap)
	return response_status, response_body
}

func enrichInstancePayload(rawpayload string, instance_name string, instance_type string, machine_image string, sshkey_name string, vnet_name string, availabilityZone string) string {
	var instance_api_payload = rawpayload
	instance_api_payload = strings.Replace(instance_api_payload, "<<instance-name>>", instance_name, 1)
	instance_api_payload = strings.Replace(instance_api_payload, "<<instance-type>>", instance_type, 1)
	instance_api_payload = strings.Replace(instance_api_payload, "<<ssh-key-name>>", sshkey_name, 1)
	instance_api_payload = strings.Replace(instance_api_payload, "<<vnet-name>>", vnet_name, 1)
	instance_api_payload = strings.Replace(instance_api_payload, "<<availability-zone>>", availabilityZone, 1)
	instance_api_payload = strings.Replace(instance_api_payload, "<<machine-image>>", machine_image, 1)
	instance_api_payload = strings.Replace(instance_api_payload, `\u003c\u003cinstance-name\u003e\u003e`, instance_name, 1)
	instance_api_payload = strings.Replace(instance_api_payload, `\u003c\u003cinstance-type\u003e\u003e`, instance_type, 1)
	instance_api_payload = strings.Replace(instance_api_payload, `\u003c\u003cssh-key-name\u003e\u003e`, sshkey_name, 1)
	instance_api_payload = strings.Replace(instance_api_payload, `\u003c\u003cvnet-name\u003e\u003e`, vnet_name, 1)
	instance_api_payload = strings.Replace(instance_api_payload, `\u003c\u003cavailability-zone\u003e\u003e`, availabilityZone, 1)
	instance_api_payload = strings.Replace(instance_api_payload, `\u003c\u003cmachine-image\u003c\u003c`, machine_image, 1)
	return instance_api_payload
}

/*func updateVM(vm_name string, vmta string) {

}*/
