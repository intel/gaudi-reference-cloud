package service_apis

import (
	"encoding/json"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/service_apis/utils"
)

func getLB(url string, token string) (int, string) {
	response := client.Get(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	//LB response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func createLB(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Post(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "POST API")
	//LB response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func deleteLB(url string, token string) (int, string) {
	response := client.Delete(url, token)
	responseCode, responseBody := client.LogRestyInfo(response, "DELETE API")
	//LB response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func putLB(url string, token string, payload map[string]interface{}) (int, string) {
	response := client.Put(url, token, payload)
	responseCode, responseBody := client.LogRestyInfo(response, "PUT API")
	//LB response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func GetLBById(lb_api_base_url string, token string, lb_id string) (int, string) {
	var get_bylb_id_url = lb_api_base_url + "/id/" + lb_id
	get_response_byid_status, get_response_byid_body := getLB(get_bylb_id_url, token)
	return get_response_byid_status, get_response_byid_body
}

func GetLBByName(lb_api_base_url string, token string, lb_name string) (int, string) {
	var get_bylb_name_url = lb_api_base_url + "/name/" + lb_name
	get_response_byname_status, get_response_byname_body := getLB(get_bylb_name_url, token)
	return get_response_byname_status, get_response_byname_body
}

// TODO : implement full functionality
/*func GetLBWithLabel(lb_api_base_url string, token string, lb_name string) (int, string) {
    var get_bylb_name_url = lb_api_base_url + "search" + lb_name
    get_response_byname_status, get_response_byname_body := getLB(get_bylb_name_url, token)
    return get_response_byname_status, get_response_byname_body
}*/

func GetAllLB(lb_api_base_url string, token string, params map[string]string) (int, string) {
	url, err := utils.ConstructURL(lb_api_base_url, params)
	if err != nil {
		logInstance.Println("error in constructing url", err)
	}
	logInstance.Println("constructed url: ", url)
	get_alllbs_response_status, get_alllbs_response_body := getLB(url, token)
	return get_alllbs_response_status, get_alllbs_response_body
}

func PutLBById(lb_api_base_url string, token string, lb_id string, payload string) (int, string) {
	var put_bylb_id_url = lb_api_base_url + "/id/" + lb_id
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byid_body, put_response_byid_status := putLB(put_bylb_id_url, token, jsonMap)
	return put_response_byid_body, put_response_byid_status
}

func PutLBByName(lb_api_base_url string, token string, lb_name string, payload string) (int, string) {
	var put_bylb_name_url = lb_api_base_url + "/name/" + lb_name
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	put_response_byname_body, put_response_byname_status := putLB(put_bylb_name_url, token, jsonMap)
	return put_response_byname_body, put_response_byname_status
}

func DeleteLBById(lb_api_base_url string, token string, lb_id string) (int, string) {
	var delete_bylb_id_url = lb_api_base_url + "/id/" + lb_id
	delete_response_byid_body, delete_response_byid_status := deleteLB(delete_bylb_id_url, token)
	return delete_response_byid_body, delete_response_byid_status
}

func DeleteLBByName(lb_api_base_url string, token string, lb_name string) (int, string) {
	var delete_bylb_name_url = lb_api_base_url + "/name/" + lb_name
	delete_response_byname_body, delete_response_byname_status := deleteLB(delete_bylb_name_url, token)
	return delete_response_byname_body, delete_response_byname_status
}

func SearchLoadBalancers(lb_api_base_url string, token string, lb_api_payload string) (int, string) {
	lb_api_search_url := lb_api_base_url + "/search"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(lb_api_payload), &jsonMap)
	response_status, response_body := createLB(lb_api_search_url, token, jsonMap)
	return response_status, response_body
}

func CreateLB(lb_api_base_url string, token string, lb_payload string, lb_name string, cloud_account string,
	listener_port string, monitor_type string, instance_resource_id string, source_ip string) (int, string) {
	lb_payload = enrichLBPayload(lb_payload, lb_name, cloud_account, listener_port, monitor_type, instance_resource_id, source_ip)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(lb_payload), &jsonMap)
	response_status, response_body := createLB(lb_api_base_url, token, jsonMap)
	return response_status, response_body
}

func CreateLBWithCustomizedPayload(lb_api_base_url string, token string, lb_api_payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(lb_api_payload), &jsonMap)
	response_status, response_body := createLB(lb_api_base_url, token, jsonMap)
	return response_status, response_body
}

func enrichLBPayload(lb_payload string, lb_name string, cloud_account string, listener_port string, monitor_type string, instance_resource_id string, source_ip string) string {
	lb_payload = strings.Replace(lb_payload, "<<lb-name>>", lb_name, 1)
	lb_payload = strings.Replace(lb_payload, "<<cloud-account>>", cloud_account, 1)
	lb_payload = strings.Replace(lb_payload, "<<listener-port>>", listener_port, 1)
	lb_payload = strings.Replace(lb_payload, "<<monitor-type>>", monitor_type, 1)
	lb_payload = strings.Replace(lb_payload, "<<instance-resource-id>>", instance_resource_id, 1)
	lb_payload = strings.Replace(lb_payload, "<<source-ips>>", source_ip, 1)
	return lb_payload
}

/*func updateVM(vm_name string, vmta string) {
}*/
