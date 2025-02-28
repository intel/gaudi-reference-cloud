package financials

import (
	
	"goFramework/framework/common/http_client"
)

// Util Functions

func getBillingResponse(url string, expected_status_code int) (string, int) {
	responseBody, responseCode := http_client.Get(url, expected_status_code)
	if responseBody == "Failed" {
		return "null", responseCode
	} else {
		return responseBody, responseCode
	}
}

func GetBillingWithParams(url string, params string, expected_status_code int) (string, int) {
	billing_endpoing_url := url + "?" + "code=" + params
	get_response_body, get_response_status := getBillingResponse(billing_endpoing_url, expected_status_code)
	return get_response_body, get_response_status
}