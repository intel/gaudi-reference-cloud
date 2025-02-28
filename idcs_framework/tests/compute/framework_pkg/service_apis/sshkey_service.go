package service_apis

import (
	"encoding/json"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"strings"
)

func getSSHKey(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	//SSH Public key schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createSSHKey(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Post(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "POST API")
	//SSH Public key response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteSSHKey(url string, token string) (int, string) {
	response := client.Delete(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "DELETE API")
	//SSH Public key response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetSSHKeyById(ssh_api_base_url string, token string, ssh_id string) (int, string) {
	var get_byssh_id_url = ssh_api_base_url + "/id/" + ssh_id
	get_response_byid_status, get_response_byid_body := getSSHKey(get_byssh_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetSSHKeyByName(ssh_api_base_url string, token string, ssh_name string) (int, string) {
	var get_byssh_name_url = ssh_api_base_url + "/name/" + ssh_name
	get_response_byname_status, get_response_byname_body := getSSHKey(get_byssh_name_url, token)
	return get_response_byname_status, get_response_byname_body
}

func GetAllSSHKey(ssh_api_base_url string, token string) (int, string) {
	get_allsshs_response_status, get_allsshs_response_body := getSSHKey(ssh_api_base_url, token)
	return get_allsshs_response_status, get_allsshs_response_body
}

func DeleteSSHKeyById(ssh_api_base_url string, token string, ssh_id string) (int, string) {
	var delete_byssh_id_url = ssh_api_base_url + "/id/" + ssh_id
	delete_response_byid_status, delete_response_byid_body := deleteSSHKey(delete_byssh_id_url, token)
	return delete_response_byid_status, delete_response_byid_body
}

func DeleteSSHKeyByName(ssh_api_base_url string, token string, ssh_name string) (int, string) {
	var delete_byssh_name_url = ssh_api_base_url + "/name/" + ssh_name
	delete_response_byname_status, delete_response_byname_body := deleteSSHKey(delete_byssh_name_url, token)
	return delete_response_byname_status, delete_response_byname_body
}

func CreateSSHKey(ssh_publickey_endpoint string, token string, payload string, sshkeyname string, sshkeyvalue string) (int, string) {
	sshkey_payload := enrichSSHKeyPayload(payload, sshkeyname, sshkeyvalue)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(sshkey_payload), &jsonMap)
	response_status, response_body := createSSHKey(ssh_publickey_endpoint, token, jsonMap)
	return response_status, response_body
}

func CreateSSHKeyWithCustomizedPayload(ssh_api_base_url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := createSSHKey(ssh_api_base_url, token, jsonMap)
	return response_status, response_body
}

func enrichSSHKeyPayload(rawpayload string, sshkey_name string, sshkeyvalue string) string {
	enriched_payload := rawpayload
	enriched_payload = strings.Replace(enriched_payload, "<<ssh-key-name>>", sshkey_name, 1)
	enriched_payload = strings.Replace(enriched_payload, "<<ssh-user-public-key>>", sshkeyvalue, 1)
	return enriched_payload
}
