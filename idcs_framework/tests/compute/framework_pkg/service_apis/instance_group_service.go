package service_apis

import (
	"encoding/json"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"strings"
)

func getInstanceGroup(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	//Instance group response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createInstancesWithGroup(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Post(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "POST API")
	//Instance group response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteInstanceGroup(url string, token string) (int, string) {
	response := client.Delete(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "DELETE API")
	//Instance group response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func patchInstanceGroup(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Patch(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "PATCH API")
	//Instance group response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetAllInstanceGroups(instance_group_base_url string, token string) (int, string) {
	get_instances_group_response_status, get_instances_group_response_body := getInstanceGroup(instance_group_base_url, token)
	return get_instances_group_response_status, get_instances_group_response_body
}

func DeleteInstanceGroupByName(instance_group_base_url string, token string, instance_group_name string) (int, string) {
	var delete_by_instancegroup_name_url = instance_group_base_url + "/name/" + instance_group_name
	delete_response_byname_body, delete_response_byname_status := deleteInstanceGroup(delete_by_instancegroup_name_url, token)
	return delete_response_byname_body, delete_response_byname_status
}

func DeleteSingleInstanceFromGroupByName(instance_group_base_url string, token string, instance_group_name string, instance_name string) (int, string) {
	var delete_by_instancegroup_name_url = instance_group_base_url + "/name/" + instance_group_name + "/instance/name/" + instance_name
	delete_response_byname_body, delete_response_byname_status := deleteInstanceGroup(delete_by_instancegroup_name_url, token)
	return delete_response_byname_body, delete_response_byname_status
}

func DeleteSingleInstanceFromGroupById(instance_group_base_url string, token string, instance_group_name string, instance_id string) (int, string) {
	var delete_by_instancegroup_name_url = instance_group_base_url + "/name/" + instance_group_name + "/instance/id/" + instance_id
	delete_response_byname_body, delete_response_byname_status := deleteInstanceGroup(delete_by_instancegroup_name_url, token)
	return delete_response_byname_body, delete_response_byname_status
}

// func PatchInstancesGroup(instance_group_base_url string, token string, instance_group_payload string) (int, string) {
// 	var jsonMap map[string]interface{}
// 	json.Unmarshal([]byte(instance_group_payload), &jsonMap)
// 	response_status, response_body := patchInstanceGroup(instance_group_base_url, token, jsonMap)
// 	return response_status, response_body
// }

func CreateInstancesGroupWithCustomizedPayload(instance_group_base_url string, token string, instance_group_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_group_payload), &jsonMap)
	response_status, response_body := createInstancesWithGroup(instance_group_base_url, token, jsonMap)
	return response_status, response_body
}

func InstanceGroupCreationWithMIMap(instance_group_base_url string, token string, instance_group_payload string, instance_group_name string,
	instance_count string, instance_type string, sshkey_name string, vnet_name string, imageMapping map[string]string, availabilityZone string) (int, string) {
	var instance_group_api_payload string
	defaultImage := imageMapping[instance_type]
	instance_group_api_payload = enrichInstanceGroupPayload(instance_group_payload, instance_group_name, instance_count, instance_type, defaultImage, sshkey_name,
		vnet_name, availabilityZone)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_group_api_payload), &jsonMap)
	response_status, response_body := createInstancesWithGroup(instance_group_base_url, token, jsonMap)
	return response_status, response_body
}

func InstanceGroupCreationWithoutMIMap(instance_group_base_url string, token string, instance_group_payload string, instance_group_name string,
	instance_count string, instance_type string, sshkey_name string, vnet_name string, machine_image string, availabilityZone string) (int, string) {
	instance_group_api_payload := enrichInstanceGroupPayload(instance_group_payload, instance_group_name, instance_count, instance_type, machine_image, sshkey_name,
		vnet_name, availabilityZone)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_group_api_payload), &jsonMap)
	response_status, response_body := createInstancesWithGroup(instance_group_base_url, token, jsonMap)
	return response_status, response_body
}

func enrichInstanceGroupPayload(rawpayload string, instance_group_name string, instance_count string, instance_type string, machine_image string, sshkey_name string,
	vnet_name string, availabilityZone string) string {
	var instance_group_api_payload = rawpayload
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<instance-group-name>>", instance_group_name, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<instance-count>>", instance_count, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<instance-type>>", instance_type, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<ssh-key-name>>", sshkey_name, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<vnet-name>>", vnet_name, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<availability-zone>>", availabilityZone, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<machine-image>>", machine_image, 1)
	return instance_group_api_payload
}

func InstanceGroupScaleUp(instance_group_base_url string, token string, patch_payload string, instance_count string, instance_group_name string) (int, string) {
	instance_group_patch_url := instance_group_base_url + "/name/" + instance_group_name + "/scale-up"
	var instance_group_api_payload = patch_payload
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<instance-count>>", instance_count, 1)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_group_api_payload), &jsonMap)
	response_status, response_body := patchInstanceGroup(instance_group_patch_url, token, jsonMap)
	return response_status, response_body
}
