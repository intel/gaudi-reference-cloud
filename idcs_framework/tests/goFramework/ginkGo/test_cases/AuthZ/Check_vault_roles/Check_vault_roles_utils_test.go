package Check_vault_roles_test

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
