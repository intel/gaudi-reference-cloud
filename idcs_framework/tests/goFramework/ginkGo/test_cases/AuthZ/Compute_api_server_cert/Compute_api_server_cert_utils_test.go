package Compute_api_server_cert_test

import (
	"encoding/json"
	"goFramework/ginkGo/test_cases/testutils"
	"os"

	. "github.com/onsi/ginkgo/v2"
)

type Domain struct {
	Common_Name string `json:"common_name"`
}

func checkRoles() map[string]any {
	v_token := os.Getenv("VAULT_TOKEN")
	url := os.Getenv("VAULT_ADDR_CA") + "/roles"
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

func GenerateCertificate(role string, common_name string) bool {
	v_token := os.Getenv("VAULT_TOKEN")
	url_passed := os.Getenv("VAULT_ADDR_CA") + "/issue/" + role
	method := "POST"

	testutils.GenerateCertificate(v_token, role, common_name, url_passed, method)

	return true
}
