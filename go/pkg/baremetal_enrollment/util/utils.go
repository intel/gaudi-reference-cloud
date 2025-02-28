// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type ApiSession struct {
	Ok             int    `json:"ok"`
	Privilege      int    `json:"privilege"`
	UserID         int    `json:"user_id"`
	Extendedpriv   int    `json:"extendedpriv"`
	RacsessionID   int    `json:"racsession_id"`
	RemoteAddr     string `json:"remote_addr"`
	ServerName     string `json:"server_name"`
	ServerAddr     string `json:"server_addr"`
	HTTPSEnabled   int    `json:"HTTPSEnabled"`
	CSRFToken      string `json:"CSRFToken"`
	Channel        int    `json:"channel"`
	PasswordStatus int    `json:"passwordStatus"`
}

const (
	BMCInsecureSkipVerifyEnvVar = "BMC_INSECURE_SKIP_VERIFY"
	ProvisioningTimeoutVar      = "PROVISIONING_TIMEOUT_VAR"
	DeprovisionTimeoutVar       = "INSPECTION_TIMEOUT_VAR"
)

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func GetRoleID(roleIdPath string) (string, error) {
	b, err := os.ReadFile(roleIdPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s", err)
	}
	return string(b), nil
}

func MoveMatchingStringsToStart(strings []string, regex string) ([]string, error) {
	dst := make([]string, len(strings))
	copy(dst, strings)

	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize regex: %v", err)
	}
	for i := 0; i < len(dst); i++ {
		if r.MatchString(dst[i]) {
			dst[0], dst[i] = dst[i], dst[0]
		}
	}

	return dst, nil
}

func NormalizeMACAddress(macAddr string) string {
	// All MACs will be lower case (Linux standard)
	lowerCase := strings.ToLower(macAddr)
	// All MACs will be delineated by colons (Linux standard)
	noHyphens := strings.Replace(lowerCase, "-", ":", -1)

	return noHyphens
}

func HttpLogin(ctx context.Context, loginSession *ApiSession, username string, password string, bmcUrl string) ([]*http.Cookie, error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.httpLogin")
	log.Info("httpLogin")
	insecureSkipVerify, err := strconv.ParseBool(GetEnv(BMCInsecureSkipVerifyEnvVar, "true"))
	if err != nil {
		return nil, fmt.Errorf("failed to read env InsecureSkipVerifyEnvVar")
	}

	// Setup client http
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	// setup payload
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	encodedData := data.Encode()
	client := http.Client{Timeout: 15 * time.Second, Transport: tr}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/session", bmcUrl), strings.NewReader(encodedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send login request: %v", err)
	}
	if res.StatusCode > 200 {
		return nil, fmt.Errorf("failed to request new session %d", res.StatusCode)
	}
	defer res.Body.Close()
	cookies := res.Cookies()
	resBody, err := io.ReadAll(res.Body)
	log.Info(string(resBody))
	if err != nil {
		return nil, fmt.Errorf("failed to read http response: %v", err)
	}
	// unmarshal API session
	err = json.Unmarshal(resBody, loginSession)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshall ApiSession: %v", err)
	}
	return cookies, nil
}

func HttpLogout(ctx context.Context, sessionToken string, cookies []*http.Cookie, bmcUrl string) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.httpLogout")

	log.Info("httpLogout", "cookie", cookies[0].Name)
	_, statusCode, err := HttpRequest(ctx, "session", http.MethodDelete, nil, sessionToken, cookies, bmcUrl)
	if err != nil {
		log.Error(err, "failed to logout")
		return
	}
	if statusCode > 200 {
		log.Info("failed with status code", "statusCode", statusCode)
		return
	}
}

func HttpRequest(ctx context.Context, api string, method string, body []byte, sessionToken string, cookies []*http.Cookie, bmcUrl string) ([]byte, int, error) {
	insecureSkipVerify, err := strconv.ParseBool(GetEnv(BMCInsecureSkipVerifyEnvVar, "true"))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read env InsecureSkipVerifyEnvVar")
	}

	// Setup client http
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, 500, fmt.Errorf("got error while creating cookie jar %s", err.Error())
	}
	client := http.Client{
		Timeout:   5 * time.Second,
		Transport: tr,
		Jar:       jar,
	}
	// create a request
	req, err := http.NewRequest(method, fmt.Sprintf("%s/api/%s", bmcUrl, api), bytes.NewBuffer(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	req.Header.Set("X-Csrftoken", sessionToken)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to create http request: %v", err)
	}
	client.Jar.SetCookies(req.URL, cookies)

	res, err := client.Do(req)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to send http request: %v", err)
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 500, fmt.Errorf("failed to read http response: %v", err)
	}
	return resBody, res.StatusCode, nil
}

func GenerateSSHPrivateKey() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	return string(pem.EncodeToMemory(privateKeyPEM)), nil
}
