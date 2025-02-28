package storage

import (
	"encoding/json"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/client"
)

func getCall(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	//Schema validation yet to be implemented
	return responseCode, responseBody
}

func createCall(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Post(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "POST API")
	// Schema validation yet to be implemented
	return responseCode, responseBody
}

func deleteCall(url string, token string) (int, string) {
	response := client.Delete(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "DELETE API")
	// Schema validation yet to be implemented
	return responseCode, responseBody
}

func putCall(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Put(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "PUT API")
	// Schema yet to be implemented
	return responseCode, responseBody
}

// Bucket functions
func GetAllBuckets(get_buckets_url string, token string) (int, string) {
	response_status, response_body := getCall(get_buckets_url, token)
	return response_status, response_body
}

func CreateBucket(create_bucket_url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := createCall(create_bucket_url, token, jsonMap)
	return response_status, response_body
}

func DeleteBucketById(delete_buckets_url string, token string, resource_id string) (int, string) {
	var delete_bucket_id_url = delete_buckets_url + "/id/" + resource_id
	delete_response_byid_status, delete_response_byid_body := deleteCall(delete_bucket_id_url, token)
	return delete_response_byid_status, delete_response_byid_body
}

func GetBucketById(get_bucket_url string, token string, resource_id string) (int, string) {
	var get_bucket_id_url = get_bucket_url + "/id/" + resource_id
	get_response_byid_status, get_response_byid_body := getCall(get_bucket_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func DeleteBucketByName(delete_buckets_url string, token string, name string) (int, string) {
	var delete_bucket_name_url = delete_buckets_url + "/name/" + name
	delete_response_byname_status, delete_response_byname_body := deleteCall(delete_bucket_name_url, token)
	return delete_response_byname_status, delete_response_byname_body
}

func GetBucketByName(get_bucket_url string, token string, name string) (int, string) {
	var get_bucket_id_url = get_bucket_url + "/name/" + name
	get_response_byid_status, get_response_byid_body := getCall(get_bucket_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

// User functions
func GetAllPrincipals(get_users_url string, token string) (int, string) {
	response_status, response_body := getCall(get_users_url, token)
	return response_status, response_body
}

func CreatePrincipal(create_user_url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := createCall(create_user_url, token, jsonMap)
	return response_status, response_body
}

func DeletePrincipalById(delete_users_url string, token string, resource_id string) (int, string) {
	var delete_user_id_url = delete_users_url + "/id/" + resource_id
	delete_response_byid_status, delete_response_byid_body := deleteCall(delete_user_id_url, token)
	return delete_response_byid_status, delete_response_byid_body
}

func GetPrincipalById(get_user_url string, token string, resource_id string) (int, string) {
	var get_user_id_url = get_user_url + "/id/" + resource_id
	get_response_byid_status, get_response_byid_body := getCall(get_user_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func DeletePrincipalByName(delete_users_url string, token string, name string) (int, string) {
	var delete_user_name_url = delete_users_url + "/name/" + name
	delete_response_byname_status, delete_response_byname_body := deleteCall(delete_user_name_url, token)
	return delete_response_byname_status, delete_response_byname_body
}

func GetPrincipalByName(get_user_url string, token string, name string) (int, string) {
	var get_user_id_url = get_user_url + "/name/" + name
	get_response_byid_status, get_response_byid_body := getCall(get_user_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func PutCredentialPrincipalById(put_credentials_user_url string, token string, resource_id string, payload string) (int, string) {
	var put_credentials_id_url = put_credentials_user_url + "/id/" + resource_id + "/credentials"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byid_status, put_response_byid_body := putCall(put_credentials_id_url, token, jsonMap)
	return put_response_byid_status, put_response_byid_body
}

func PutCredentialPrincipalByName(put_credentials_user_url string, token string, name string, payload string) (int, string) {
	var put_credentials_name_url = put_credentials_user_url + "/name/" + name + "/credentials"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byid_status, put_response_byid_body := putCall(put_credentials_name_url, token, jsonMap)
	return put_response_byid_status, put_response_byid_body
}

func PutPolicyPrincipalById(put_policy_user_url string, token string, resource_id string, payload string) (int, string) {
	var put_policy_id_url = put_policy_user_url + "/id/" + resource_id + "/policy"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byid_status, put_response_byid_body := putCall(put_policy_id_url, token, jsonMap)
	return put_response_byid_status, put_response_byid_body
}

func PutPolicyPrincipalByName(put_policy_user_url string, token string, name string, payload string) (int, string) {
	var put_policy_name_url = put_policy_user_url + "/name/" + name + "/policy"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byid_status, put_response_byid_body := putCall(put_policy_name_url, token, jsonMap)
	return put_response_byid_status, put_response_byid_body
}

func CreateRule(create_rule_url string, token string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	response_status, response_body := createCall(create_rule_url, token, jsonMap)
	return response_status, response_body
}

func DeleteRuleById(delete_rule_url string, token string, resource_id string) (int, string) {
	var delete_rule_id_url = delete_rule_url + resource_id
	delete_response_byid_status, delete_response_byid_body := deleteCall(delete_rule_id_url, token)
	return delete_response_byid_status, delete_response_byid_body
}

// Rule functions
func GetRuleById(get_rule_url string, bucket_id string, token string, resource_id string) (int, string) {
	var get_rule_id_url = get_rule_url + bucket_id + "/lifecyclerule/id/" + resource_id
	get_response_byid_status, get_response_byid_body := getCall(get_rule_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetAllRules(get_rule_url string, token string, bucket_id string) (int, string) {
	response_status, response_body := getCall(get_rule_url+bucket_id+"/lifecyclerule", token)
	return response_status, response_body
}

func PutRuleById(put_rule_url string, bucket_id string, token string, resource_id string, payload string) (int, string) {
	var put_rule_id_url = put_rule_url + bucket_id + "/lifecyclerule/id/" + resource_id
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byid_status, put_response_byid_body := putCall(put_rule_id_url, token, jsonMap)
	return put_response_byid_status, put_response_byid_body
}
