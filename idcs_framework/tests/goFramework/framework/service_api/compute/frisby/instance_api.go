package frisby

import (
	"encoding/json"

	"goFramework/framework/frisby_client"
)

func getInstance(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetInstance(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createInstance(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteInstance(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func putInstance(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	//Instance response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func CreateInstance(instance_api_base_url string, token string, instance_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(instance_api_payload), &jsonMap)
	response_status, response_body := createInstance(instance_api_base_url, token, jsonMap)
	return response_status, response_body
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

func GetAllInstance(instance_api_base_url string, token string) (int, string) {
	get_allinstances_response_status, get_allinstances_response_body := getInstance(instance_api_base_url, token)
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

func DeleteInstance2(instance_api_base_url string, token string) (int, string) {
	delete_response_byid_body, delete_response_byid_status := deleteInstance(instance_api_base_url, token)
	return delete_response_byid_body, delete_response_byid_status
}

func PatchInstance(url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(payload), &jsonMap)
	if err != nil {
		return 0, err.Error()
	}
	frisby_response := frisby_client.Patch(url, token, jsonMap)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PATCH API")

	return responseCode, responseBody
}

/*func updateVM(vm_name string, vmta string) {

}*/
