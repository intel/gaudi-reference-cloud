package compute_api_comms_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
)

func GenerateRequestToGetCertificate(common_name string, role string) []byte {
	var url = os.Getenv("VAULT_ADDR") + "/issue/"
	var token = os.Getenv("VAULT_TOKEN")
	var reqData = fmt.Sprintf(`{"common_name": "%s"}`, common_name)

	client := &http.Client{}
	var data = strings.NewReader(reqData)
	req, err := http.NewRequest("POST", url+role, data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return bodyText
}

func GetByteValue(body []byte) any {
	var result any
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		Fail("Error during Unmarshal(): " + err.Error())
	}

	return result
}

func GetFieldInfo(body []byte) map[string]any {
	var result map[string]any
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		Fail("Error during Unmarshal(): " + err.Error())
	}

	mapped := result["data"].(map[string]any)

	return mapped
}

func VaultCertSetup(issuing_ca string, private_key string, certificate string) (serverTLSConf *tls.Config, clientTLSConf *tls.Config, err error) {
	// Get CA Certificate
	byteCaCert := []byte(issuing_ca)
	blockCaCert, _ := pem.Decode(byteCaCert)
	if blockCaCert == nil {
		Fail("failed to parse CA certificate")
	}

	// Get Certificate
	byteCert := []byte(certificate)
	blockCert, _ := pem.Decode(byteCert)
	if blockCert == nil {
		Fail("failed to parse certificate")
	}

	// Get Key
	byteKey := []byte(private_key)
	blockKey, _ := pem.Decode(byteKey)
	if blockKey == nil {
		Fail("failed to parse certificate")
	}

	serverCert, err := tls.X509KeyPair([]byte(certificate), []byte(private_key))
	if err != nil {
		return nil, nil, err
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM([]byte(issuing_ca))
	clientTLSConf = &tls.Config{
		RootCAs: certpool,
	}

	return
}
