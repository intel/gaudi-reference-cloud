package Check_ttl_test

import (
	"encoding/json"
	"goFramework/ginkGo/test_cases/testutils"
	"os"

	. "github.com/onsi/ginkgo/v2"
)

func checkRoles() map[string]any {
	v_token := os.Getenv("VAULT_TOKEN")
	url := os.Getenv("VAULT_ADDR") + "/roles"
	method := "LIST"

	results := testutils.VaultRequest(v_token, url, method)

	var result map[string]any
	err := json.Unmarshal([]byte(results), &result)
	if err != nil {
		Fail("Error during Unmarshal(): " + err.Error())
	}

	roles := result["data"].(map[string]any)

	return roles
}

func iterateRolesToGetValue(roles map[string]any) bool {
	method := "GET"
	url := os.Getenv("VAULT_ADDR") + "/roles/"
	token := os.Getenv("VAULT_TOKEN")

	var value float64
	var check = true

	for _, service := range roles["keys"].([]interface{}) {
		value = getInformationRequest(service.(string), method, url, token, "max_ttl")

		if value < 3000 {
			check = false
			break
		}
	}

	return check
}

func getInformationRequest(service string, method string, url string, token string, field_to_check string) float64 {
	url = url + service
	resp := testutils.VaultRequest(token, url, method)

	value := getTtlInfo(resp, service, field_to_check)

	return value
}

func getTtlInfo(body []byte, service string, field string) float64 {
	var result map[string]any
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		Fail("Error during Unmarshal(): " + err.Error())
	}

	mapped := result["data"].(map[string]any)

	return mapped["ttl"].(float64)
}
