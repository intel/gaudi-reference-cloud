package Check_create_certificate_test

import (
	"bytes"
	"encoding/json"
	"goFramework/ginkGo/test_cases/testutils"
	"net/http"
	"net/url"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
)

type Domain struct {
	Common_Name string `json:"common_name"`
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

func generatecertificates(domain []any, service string) bool {
	isSuccess := true
	for _, value := range domain {
		f := Domain{
			Common_Name: value.(string),
		}
		data, err := json.Marshal(f)
		if err != nil {
			Fail(err.Error())
		}
		use_proxy, err := strconv.ParseBool(os.Getenv("USE_PROXY"))
		if err != nil {
			Fail("Error getting use proxy variable" + err.Error())
		}
		if use_proxy {
			proxyURL, _ := url.Parse(os.Getenv("https_proxy"))
			proxy := http.ProxyURL(proxyURL)
			transport := &http.Transport{Proxy: proxy}
			client := &http.Client{Transport: transport}
			reader := bytes.NewReader(data)

			req, err := http.NewRequest("POST", os.Getenv("VAULT_ADDR")+"/"+service, reader)
			if err != nil {
				Fail(err.Error())
			}
			req.Header.Set("X-Vault-Token", os.Getenv("VAULT_TOKEN"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp, err := client.Do(req)
			if err != nil {
				Fail(err.Error())
			}
			defer resp.Body.Close()
		} else {
			client := &http.Client{}
			reader := bytes.NewReader(data)

			req, err := http.NewRequest("POST", os.Getenv("VAULT_ADDR")+"/"+service, reader)
			if err != nil {
				Fail(err.Error())
			}
			req.Header.Set("X-Vault-Token", os.Getenv("VAULT_TOKEN"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp, err := client.Do(req)
			if err != nil {
				Fail(err.Error())
			}
			defer resp.Body.Close()
		}
	}

	return isSuccess
}

func iterateRolesToIssueCertificates(roles map[string]any) bool {
	method := "GET"
	url := os.Getenv("VAULT_ADDR") + "/roles/"
	token := os.Getenv("VAULT_TOKEN")

	var check bool

	for _, service := range roles["keys"].([]interface{}) {
		values := getInformationRequestDomains(service.(string), method, url, token, "allowed_domains")
		check = generatecertificates(values, service.(string))

		if !check {
			break
		}
	}

	return check
}

func Flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
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
