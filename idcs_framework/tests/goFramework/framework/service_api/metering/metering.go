package metering

import (
	"encoding/json"

	"goFramework/framework/frisby_client"
)

func searchMeteringRecords(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Metering response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func SearchAllMeteringRecords(metering_api_base_url string, token string, metering_search_payload string) (int, string) {
	metering_api_base_url = metering_api_base_url + "/search"
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(metering_search_payload), &jsonMap)
	response_status, response_body := searchMeteringRecords(metering_api_base_url, token, jsonMap)
	return response_status, response_body
}
