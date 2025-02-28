package storage

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	"github.com/tidwall/gjson"
)

func CreateInstance(instance_api_url string, token string, instance_api_payload string, instance_name string, instance_type string,
	sshkey_name string, vnet_name string, machine_image string, availabilityZone string) (int, string) {
	//instance_api_payload := enrichInstancePayload(instance_payload, instance_name, instance_type, machine_image, sshkey_name, vnet_name, availabilityZone)
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
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_api_payload), &jsonMap)
	response := client.Post(instance_api_url, token, jsonMap)
	return response.StatusCode(), string(response.Body())
}

func InstanceGroupCreation(instance_group_base_url string, token string, rawpayload string, instance_group_name string,
	instance_count string, instance_type string, sshkey_name string, vnet_name string, machine_image string, availabilityZone string) (int, string) {
	instance_group_api_payload := rawpayload
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<instance-group-name>>", instance_group_name, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<instance-count>>", instance_count, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<instance-type>>", instance_type, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<ssh-key-name>>", sshkey_name, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<vnet-name>>", vnet_name, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<availability-zone>>", availabilityZone, 1)
	instance_group_api_payload = strings.Replace(instance_group_api_payload, "<<machine-image>>", machine_image, 1)

	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_group_api_payload), &jsonMap)
	response := client.Post(instance_group_base_url, token, jsonMap)
	return response.StatusCode(), string(response.Body())
}

func GetInstanceById(instance_api_base_url string, token string, instance_id string) (int, string) {
	var get_byinstance_id_url = instance_api_base_url + "/id/" + instance_id
	response := client.Get(get_byinstance_id_url, token)
	return response.StatusCode(), string(response.Body())
}

func GetInstancesWithinGroup(instance_api_base_url string, token string, instance_group_name string) (int, string) {
	var get_byinstance_name_url = instance_api_base_url + "?metadata.instanceGroup=" + instance_group_name
	response := client.Get(get_byinstance_name_url, token)
	return response.StatusCode(), string(response.Body())
}

func DeleteInstanceById(instance_api_base_url string, token string, instance_id string) (int, string) {
	var delete_byinstance_id_url = instance_api_base_url + "/id/" + instance_id
	response := client.Delete(delete_byinstance_id_url, token)
	return response.StatusCode(), string(response.Body())
}

func GetAllInstanceGroups(instance_group_base_url string, token string) (int, string) {
	response := client.Get(instance_group_base_url, token)
	return response.StatusCode(), string(response.Body())
}

func DeleteInstanceGroupByName(instance_group_base_url string, token string, instance_group_name string) (int, string) {
	var delete_by_instancegroup_name_url = instance_group_base_url + "/name/" + instance_group_name
	response := client.Delete(delete_by_instancegroup_name_url, token)
	return response.StatusCode(), string(response.Body())
}

func CreateSSHKey(ssh_publickey_endpoint string, token string, rawpayload string, sshkeyname string, sshkeyvalue string) (int, string) {
	sshkey_payload := rawpayload
	sshkey_payload = strings.Replace(sshkey_payload, "<<ssh-key-name>>", sshkeyname, 1)
	sshkey_payload = strings.Replace(sshkey_payload, "<<ssh-user-public-key>>", sshkeyvalue, 1)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(sshkey_payload), &jsonMap)
	response := client.Post(ssh_publickey_endpoint, token, jsonMap)
	return response.StatusCode(), string(response.Body())
}

func DeleteSSHKeyById(ssh_api_base_url string, token string, ssh_id string) (int, string) {
	var delete_byssh_id_url = ssh_api_base_url + "/id/" + ssh_id
	response := client.Delete(delete_byssh_id_url, token)
	return response.StatusCode(), string(response.Body())
}

func CreateVnet(vnet_api_endpoint string, token string, vnet_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(vnet_payload), &jsonMap)
	response := client.Post(vnet_api_endpoint, token, jsonMap)
	return response.StatusCode(), string(response.Body())
}

func DeleteVnetByName(vnet_api_endpoint string, token string, vnet_name string) (int, string) {
	var delete_vnetbyname_endpoint = vnet_api_endpoint + "/name/" + vnet_name
	response := client.Delete(delete_vnetbyname_endpoint, token)
	return response.StatusCode(), string(response.Body())
}

func CheckInstanceState(instance_endpoint, token, instance_id_created, expectedState string) func() bool {
	startTime := time.Now()
	return func() bool {
		_, get_response_byid_body := GetInstanceById(instance_endpoint, token, instance_id_created)
		instancePhase := gjson.Get(get_response_byid_body, "status.phase").String()
		logger.Log.Info("instancePhase: " + instancePhase)
		if instancePhase != expectedState {
			logger.Log.Info("Instance is not in " + expectedState + " state")
			return false
		} else {
			logger.Log.Info("Instance is in " + expectedState + " state")
			elapsedTime := time.Since(startTime)
			logger.Log.Info("Time took for instance to get to " + expectedState + " state: " + string(elapsedTime))
			return true
		}
	}
}

func CheckInstanceDeletionById(instance_endpoint string, token string, instance_id_created string) func() bool {
	startTime := time.Now()
	return func() bool {
		get_instancebyid_after_delete, _ := GetInstanceById(instance_endpoint, token, instance_id_created)
		if get_instancebyid_after_delete != 404 {
			logger.Log.Info("Instance is not yet deleted.")
			return false
		} else {
			logger.Log.Info("Instance has been deleted.")
			elapsedTime := time.Since(startTime)
			logger.Log.Info("Time took for instance deletion: " + string(elapsedTime))
			return true
		}
	}
}
