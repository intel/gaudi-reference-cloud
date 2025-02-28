package financials

import (
	"encoding/json"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/frisby_client"
	"time"
)

func searchMeteringRecords(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	time.Sleep(1 * time.Second)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	//Metering response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func SearchAllMeteringRecords(metering_api_base_url string, token string, metering_search_payload string) (int, string) {
	metering_api_base_url = metering_api_base_url + "/search"
	logger.Log.Info("Metering Search URL" + metering_api_base_url)
	logger.Logf.Infof("Metering Search payload : %s ", metering_search_payload)
	time.Sleep(1 * time.Second)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(metering_search_payload), &jsonMap)
	response_status, response_body := searchMeteringRecords(metering_api_base_url, token, jsonMap)
	logger.Logf.Infof("Metering Search response", response_body)
	return response_status, response_body
}

func postMeteringRecords(url string, token string, payload map[string]interface{}) (int, string) {
	frisby_response := frisby_client.Post(url, token, payload)
	fmt.Println("Response Response is nil: ", frisby_response.Response == nil, frisby_response)

	if frisby_response == nil || frisby_response.Response == nil || frisby_response.StatusCode == 503 {
		maxRetries := 200
		for retries := 0; retries < maxRetries; retries++ {
			frisby_response = frisby_client.Post(url, token, payload)
			if frisby_response != nil && frisby_response.Response != nil && frisby_response.StatusCode == 503 {
				break
			}
			fmt.Println("Frisby response has failed, retrying...", frisby_response)
			time.Sleep(5 * time.Second)
		}

		if frisby_response == nil || frisby_response.Response == nil {
			fmt.Println("Failed to get a valid response after retries")
			return 0, "Failed to get a valid response"
		}
	}

	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "POST API")
	// Metering response schema and common validation goes here - yet to be implemented
	return responseCode, responseBody
}

func CreateMeteringRecords(metering_api_base_url string, token string, metering_post_payload string) (int, string) {
	logger.Log.Info("Metering POST URL" + metering_api_base_url)
	logger.Logf.Infof("Metering POST payload : %s ", metering_post_payload)
	time.Sleep(1 * time.Second)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(metering_post_payload), &jsonMap)
	responseCode, responseBody := postMeteringRecords(metering_api_base_url, token, jsonMap)
	return responseCode, responseBody
}

func CreateUsageRecords(usage_api_base_url string, token string, usage_post_payload string) (int, string) {
	logger.Log.Info("Usage POST URL" + usage_api_base_url)
	logger.Logf.Infof("Usage POST payload : %s ", usage_api_base_url)
	time.Sleep(1 * time.Second)
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(usage_post_payload), &jsonMap)
	responseCode, responseBody := postMeteringRecords(usage_api_base_url, token, jsonMap)
	return responseCode, responseBody
}
