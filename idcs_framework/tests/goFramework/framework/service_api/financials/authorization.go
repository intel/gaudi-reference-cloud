package financials

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/frisby_client"
)

func ping(url string, token string) (int, string) {
	frisby_response := frisby_client.Get(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

func getUserRoles(url string, token string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte("{}"), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	frisby_response := frisby_client.Get_With_Json(url, token, jsonMap)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

func assignRole(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func unassignRole(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func lookup(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func check(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func actions(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	fmt.Println("URL...0", url)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func systemRoleExists(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	fmt.Println("URL...0", url)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func createUserRole(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func updateUserRole(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	return responseCode, responseBody
}

func deleteUserRole(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	return responseCode, responseBody
}

func getUserRole(url string, token string) (int, string) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte("{}"), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	frisby_response := frisby_client.Get_With_Json(url, token, jsonMap)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	return responseCode, responseBody
}

func deleteUserRoleResources(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	return responseCode, responseBody
}

func createUserRoleResources(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func updateUserRolePermission(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Put(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "PUT API")
	return responseCode, responseBody
}

func deleteUserRolePermission(url string, token string) (int, string) {
	frisby_response := frisby_client.Delete(url, token)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "DELETE API")
	return responseCode, responseBody
}

func addUserRoleResourcesUser(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func getJwtCognito(url string, client_id string, client_secret string, grant_type string, scope string) (int, string) {
	frisby_response := frisby_client.PostCognito(url, client_id, client_secret, grant_type, scope)
	fmt.Println("Resp: ", frisby_response.Body)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	return responseCode, responseBody
}

func Ping(url string, token string) (int, string) {
	var get_url = url + "/v1/authorization/ping"
	fmt.Println("get_url", get_url)
	get_response_status, get_response_body := ping(get_url, token)
	return get_response_status, get_response_body
}

func GetUserRoles(url string, token string, cloudaccountId string) (int, string) {
	var get_url = url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles"
	fmt.Println("get_url", get_url)
	fmt.Println("token: ", token)
	get_response_status, get_response_body := getUserRoles(get_url, token)
	return get_response_status, get_response_body
}

func AssignRole(base_url string, token string, api_payload string) (int, string) {
	var post_url = base_url + "/v1/authorization/system_role/assign"
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	fmt.Println("URL: ", post_url)
	response_status, response_body := assignRole(post_url, token, jsonMap)
	return response_status, response_body
}

func AssignRoleUser(base_url string, token string, api_payload string, admin_cloudacount string, username_member string) (int, string) {
	var post_url = base_url + "/v1/authorization/cloudaccounts/" + admin_cloudacount + "/users/" + username_member + "/roles"
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	fmt.Println("URL: ", post_url)
	response_status, response_body := assignRole(post_url, token, jsonMap)
	return response_status, response_body
}

func UnAssignRole(base_url string, token string, api_payload string) (int, string) {
	var post_url = base_url + "/v1/authorization/system_role/unassign"
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	fmt.Println("URL: ", post_url)
	response_status, response_body := unassignRole(post_url, token, jsonMap)
	return response_status, response_body
}

func LookUp(base_url string, token string, api_payload string) (int, string) {
	var post_url = base_url + "/proto.AuthzService/Lookup"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	fmt.Println("URL: ", post_url)
	response_status, response_body := lookup(post_url, token, jsonMap)
	return response_status, response_body
}

func Check(base_url string, token string, api_payload string) (int, string) {
	var post_url = base_url + "/proto.AuthzService/Check"
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	fmt.Println("URL: ", post_url)
	response_status, response_body := check(post_url, token, jsonMap)
	return response_status, response_body
}

func Actions(base_url string, token string, api_payload string) (int, string) {
	var post_url = base_url + "/v1/authorization/actions"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	fmt.Println("URL: ", post_url)
	response_status, response_body := actions(post_url, token, jsonMap)
	return response_status, response_body
}

func SystemRoleExists(base_url string, token string, api_payload string) (int, string) {
	var post_url = base_url + "/proto.AuthzService/SystemRoleExists"
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	fmt.Println("URL: ", post_url)
	response_status, response_body := systemRoleExists(post_url, token, jsonMap)
	return response_status, response_body
}

func CreateUserRole(base_url string, token string, api_payload string, cloudaccountId string) (int, string) {
	var post_url = base_url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	fmt.Println("URL: ", post_url)
	response_status, response_body := createUserRole(post_url, token, jsonMap)
	return response_status, response_body
}

func UpdateUserRole(base_url string, token string, api_payload string, cloudaccountId string, roleId string) (int, string) {
	var post_url = base_url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles/" + roleId
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	fmt.Println("URL: ", post_url)
	response_status, response_body := updateUserRole(post_url, token, jsonMap)
	return response_status, response_body
}

func DeleteUserRole(base_url string, token string, cloudaccountId string, roleId string) (int, string) {
	var post_url = base_url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles/" + roleId
	fmt.Println("URL: ", post_url)
	response_status, response_body := deleteUserRole(post_url, token)
	return response_status, response_body
}

func GetUserRole(url string, token string, cloudaccountId string, roleId string) (int, string) {
	var get_url = url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles/" + roleId
	fmt.Println("get_url", get_url)
	get_response_status, get_response_body := getUserRole(get_url, token)
	return get_response_status, get_response_body
}

func DeleteUserRoleResources(base_url string, token string, cloudaccountId string, roleId string, resourceId string) (int, string) {
	var post_url = base_url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles/" + roleId + "/resources?resourceId=" + resourceId
	fmt.Println("URL: ", post_url)
	response_status, response_body := deleteUserRoleResources(post_url, token)
	return response_status, response_body
}

func CreateUserRolePermission(base_url string, token string, api_payload string, cloudaccountId string, roleId string) (int, string) {
	var post_url = base_url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles/" + roleId + "/permissions"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(api_payload), &jsonMap)
	fmt.Println("URL: ", post_url)
	response_status, response_body := createUserRoleResources(post_url, token, jsonMap)
	return response_status, response_body
}

func UpdateUserRolePermission(base_url string, token string, api_payload string, cloudaccountId string, roleId string, permissionId string) (int, string) {
	var put_url = base_url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles/" + roleId + "/permissions/" + permissionId
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	fmt.Println("URL: ", put_url)
	response_status, response_body := updateUserRolePermission(put_url, token, jsonMap)
	return response_status, response_body
}

func DeleteUserRolePermission(base_url string, token string, cloudaccountId string, roleId string, permissionId string) (int, string) {
	var delete_url = base_url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles/" + roleId + "/permissions/" + permissionId
	fmt.Println("URL: ", delete_url)
	response_status, response_body := deleteUserRolePermission(delete_url, token)
	return response_status, response_body
}

func AddUserRoleResourceUser(base_url string, token string, api_payload string, cloudaccountId string, roleId string, userId string) (int, string) {
	var post_url = base_url + "/v1/authorization/cloudaccounts/" + cloudaccountId + "/roles/" + roleId + "/users?userId=" + userId
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(api_payload), &jsonMap)
	if err != nil {
		fmt.Println("Error while unmarshaling json:", err)
	}
	fmt.Println("URL: ", post_url)
	response_status, response_body := addUserRoleResourcesUser(post_url, token, jsonMap)
	return response_status, response_body
}

func GetJwtCognito(base_url string, grant_type string, client_id string, client_secret string, scope string) (int, string) {
	var jwt_cognito_url = base_url + "/oauth2/token"
	fmt.Println("URL: ", jwt_cognito_url)
	response_status, response_body := getJwtCognito(jwt_cognito_url, client_id, client_secret, grant_type, scope)
	return response_status, response_body
}
