package Check_pod_keys_test

import (
	"crypto/x509"
	"encoding/pem"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
)

func executeBashScript(path string, host string) []string {
	var result []string
	cmd, err := exec.Command(path, host).Output()
	if err != nil {
		Fail(err.Error())
	}

	output := string(cmd)

	if strings.Contains(path, "Certs") {
		result = strings.SplitAfter(output, "-----END CERTIFICATE-----")
	} else if strings.Contains(path, "Keys") {
		result = strings.SplitAfter(output, "-----END RSA PRIVATE KEY-----")
	}
	return result
}

func decodeKey(keyPEM string) bool {
	var check bool
	if keyPEM != "" {
		byteKey := []byte(keyPEM)
		block, _ := pem.Decode(byteKey)
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			Fail("failed to decode PEM block containing public key")
		}

		pub, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			Fail(err.Error())
		}

		if pub.Validate() != nil {
			Fail("Error validating key")
		}
	}

	check = true

	return check
}
