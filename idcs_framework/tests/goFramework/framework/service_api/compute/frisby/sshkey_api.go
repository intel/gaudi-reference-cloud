package frisby

import (
	"encoding/json"
	"strings"
	"time"

	"goFramework/framework/frisby_client"
)

func getSSHKey(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//SSH Public key schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createSSHKey(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//SSH Public key response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteSSHKey(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	//SSH Public key response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func CreateSSHKey(ssh_api_base_url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := createSSHKey(ssh_api_base_url, token, jsonMap)
	return response_status, response_body
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

func SSHPublicKeyCreation(ssh_publickey_endpoint string, token string, sshkey_payload string, sshkeyname string, sshkeyvalue string) (int, string) {
	sshkey_payload1 := sshkey_payload
	sshkey_payload1 = strings.Replace(sshkey_payload1, "<<ssh-key-name>>", sshkeyname, 1)
	sshkey_payload1 = strings.Replace(sshkey_payload1, "<<ssh-user-public-key>>", sshkeyvalue, 1)
	response_status, response_body := CreateSSHKey(ssh_publickey_endpoint, token, sshkey_payload1)
	time.Sleep(10 * time.Second)
	return response_status, response_body
}
