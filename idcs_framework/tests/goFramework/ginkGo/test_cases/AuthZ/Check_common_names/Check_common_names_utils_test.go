package Check_common_names_test

import (
	"encoding/json"
	"fmt"
	"goFramework/ginkGo/test_cases/testutils"
	"os"

	. "github.com/onsi/ginkgo/v2"
)

func Flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

func getFieldInfo(body []byte, service string, field string) []any {
	var test []any

	var result map[string]any
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		Fail("Error during Unmarshal(): " + err.Error())
	}

	mapped := result["data"].(map[string]any)
	mapped_domains := mapped[field]

	for _, value := range mapped_domains.([]interface{}) {
		test = append(test, value.(string))
	}
	return test
}

func getInformationRequestDomains(service string, method string, url string, token string, field_to_check string) []any {
	url = url + service

	resp := testutils.VaultRequest(token, url, method)

	domain := getFieldInfo(resp, service, field_to_check)

	return domain
}

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

func iterateRolesAndCheckValue(roles map[string]any, value string) []any {
	method := "GET"
	url := os.Getenv("VAULT_ADDR") + "/roles/"
	token := os.Getenv("VAULT_TOKEN")

	var values [][]any
	var test []any
	for _, service := range roles["keys"].([]interface{}) {
		values = append(values, getInformationRequestDomains(service.(string), method, url, token, value))
	}

	test = Flatten(values)

	return test
}

func retrieveValuesFromJson(value string, path string) []string {
	jsonFile, err := os.ReadFile(path)

	if err != nil {
		Fail(err.Error())
	}

	var data map[string][]string
	json.Unmarshal(jsonFile, &data)

	result := data[value]

	return result
}

func checkForValue(userValue string, domains []string) bool {
	for value := range domains {
		if domains[value] == userValue {
			return true
		}
	}
	fmt.Print("NOT FOUND: ", userValue)
	return false
}
