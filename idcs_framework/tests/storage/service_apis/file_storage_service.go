package storage

import (
	"encoding/json"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
)

func getFilesystem(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	//Schema validation yet to be implemented
	return responseCode, responseBody
}

func createFilesystem(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Post(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "POST API")
	// Schema validation yet to be implemented
	return responseCode, responseBody
}

func getUserCreds(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	// Schema validation yet to be implemented
	return responseCode, responseBody
}

func deleteFilesystem(url string, token string) (int, string) {
	response := client.Delete(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "DELETE API")
	// Schema validation yet to be implemented
	return responseCode, responseBody
}

func putFilesystem(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Put(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "PUT API")
	// Schema yet to be implemented
	return responseCode, responseBody
}

func searchFilesystems(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Post(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "POST API")
	// Sechma yet to be implemented
	return responseCode, responseBody
}

func GetAllFilesystems(get_filesystems_url string, token string) (int, string) {
	response_status, response_body := getFilesystem(get_filesystems_url, token)
	return response_status, response_body
}

func CreateFilesystem(create_filesystem_url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := createFilesystem(create_filesystem_url, token, jsonMap)
	return response_status, response_body
}

func DeleteFilesystemById(delete_filesystems_url string, token string, resource_id string) (int, string) {
	var delete_filesystem_id_url = delete_filesystems_url + "/id/" + resource_id
	delete_response_byid_body, delete_response_byid_status := deleteFilesystem(delete_filesystem_id_url, token)

	// Temporary fix to avoid backend corruption issues
	logger.Log.Info("Sleeping for 2 minutes after volume deletion...")
	time.Sleep(2 * time.Minute)

	return delete_response_byid_body, delete_response_byid_status
}

func GetFilesystemById(get_filesystem_url string, token string, resource_id string) (int, string) {
	var get_filesystem_id_url = get_filesystem_url + "/id/" + resource_id
	get_response_byid_status, get_response_byid_body := getFilesystem(get_filesystem_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func PutFilesystemById(filesystem_url string, token string, resource_id string, payload string) (int, string) {
	var put_filesystem_id_url = filesystem_url + "/id/" + resource_id
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byid_body, put_response_byid_status := putFilesystem(put_filesystem_id_url, token, jsonMap)
	return put_response_byid_body, put_response_byid_status
}

func DeleteFilesystemByName(delete_filesystems_url string, token string, name string) (int, string) {
	var delete_filesystem_name_url = delete_filesystems_url + "/name/" + name
	delete_response_byname_body, delete_response_byname_status := deleteFilesystem(delete_filesystem_name_url, token)

	// Temporary fix to avoid backend corruption issues
	logger.Log.Info("Sleeping for 2 minutes after volume deletion...")
	time.Sleep(2 * time.Minute)

	return delete_response_byname_body, delete_response_byname_status
}

func GetFilesystemByName(get_filesystem_url string, token string, name string) (int, string) {
	var get_filesystem_id_url = get_filesystem_url + "/name/" + name
	get_response_byid_status, get_response_byid_body := getFilesystem(get_filesystem_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func PutFilesystemByName(filesystem_url string, token string, name string, payload string) (int, string) {
	var put_filesystem_name_url = filesystem_url + "/name/" + name
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byname_body, put_response_byname_status := putFilesystem(put_filesystem_name_url, token, jsonMap)
	return put_response_byname_body, put_response_byname_status
}

func ListFilesystems(filesystem_url string, token string, payload string) (int, string) {
	var list_filesystems_url = filesystem_url + "/search"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := searchFilesystems(list_filesystems_url, token, jsonMap)
	return response_status, response_body
}

func GetUserCreds(user_url string, token string) (int, string) {
	response_status, response_body := getUserCreds(user_url, token)
	return response_status, response_body
}
