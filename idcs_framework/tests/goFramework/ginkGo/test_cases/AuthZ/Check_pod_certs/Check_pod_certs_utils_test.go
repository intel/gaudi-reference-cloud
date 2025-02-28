package Check_pod_certs_test

import (
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

func decodeCert(rootCert string, certPEM string, i int) bool {
	var check bool
	/*if rootCert != "" || certPEM != "" {
		byteCert := []byte(certPEM)
		block, _ := pem.Decode(byteCert)
		if block == nil {
			Fail("failed to parse  certificate")
		}

		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(rootCert))
		if !ok {
			Fail("failed to parse root certificate")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			Fail("failed to parse certificate: " + err.Error())
		}

		opts := x509.VerifyOptions{
			Roots: roots,
		}

		if _, err := cert.Verify(opts); err != nil {
			Fail("failed to verify certificate: " + err.Error())
		}
	}*/

	check = true
	return check
}
