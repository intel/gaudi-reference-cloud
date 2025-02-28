package Check_certificate_rotation_test

import (
	"goFramework/ginkGo/test_cases/testutils"
	"os"
)

func RotateCertificates() bool {
	token := os.Getenv("VAULT_TOKEN")
	url_passed := os.Getenv("VAULT_ADDR") + "/crl/rotate"
	method := "POST"
	var check bool

	bodyText := testutils.VaultRequest(token, url_passed, method)

	if bodyText != nil {
		check = true
	}

	return check
}
