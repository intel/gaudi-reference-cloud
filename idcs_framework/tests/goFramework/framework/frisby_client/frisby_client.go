package frisby_client

import (
	"fmt"
	"goFramework/framework/common/logger"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"crypto/tls"
	"os"
	"strconv"

	"github.com/mozillazg/request"
	"github.com/verdverm/frisby"
)

func AriaPost(url string, token string, payload map[string]interface{}, certFilePath string, certFileKey string) *request.Response {
	tlsConfig := tls.Config{}
	tlsConfig.GetClientCertificate = func(cri *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		cert, err := tls.LoadX509KeyPair(certFilePath, certFileKey)
		if err != nil {
			logger.Logf.Infof("failed to load X509 key pair, failed with error :", err)

		}
		return &cert, nil
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tlsConfig
	response := frisby.Create("POST Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetProxy(os.Getenv("https_proxy")).
		SetJson(payload).
		Post(url).
		Send().
		Resp

	return response
}

func formatUrl(url string) string {
	// Generate a unique timestamp
	timestamp := time.Now().UnixNano()

	// Check if the URL already contains query parameters
	if strings.Contains(url, "?") {
		return fmt.Sprintf("%s&t=%d", url, timestamp)
	} else {
		return fmt.Sprintf("%s?t=%d", url, timestamp)
	}
}

func Get(url string, token string) *request.Response {
	//url = formatUrl(url)
	response := frisby.Create("GET Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetHeader("Accept", "application/json").
		SetProxy(os.Getenv("https_proxy")).
		Get(url).
		Send().Resp

	if response != nil {
		fmt.Println("URL: ", url)
		fmt.Println("inner url: ", response.Request.URL)
		fmt.Println("Response code is :", response.StatusCode)
		fmt.Println("Response body is :", response)
		url = formatUrl(url)
	}
	return response
}

func Get_With_Json(url string, token string, payload map[string]interface{}) *request.Response {
	url = formatUrl(url)
	response := frisby.Create("Get Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetProxy(os.Getenv("https_proxy")).
		SetJson(payload).
		Get(url).
		Send().
		Resp

	if response != nil && response.Response != nil {
		fmt.Println("URL: ", url)
		fmt.Println("inner url: ", response.Request.URL)
	}

	return response
}

func Patch(url string, token string, payload map[string]interface{}) *request.Response {
	response := frisby.Create("PATCH Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetJson(payload).
		SetProxy(os.Getenv("https_proxy")).
		Patch(url).
		Send().
		Resp

	return response
}

func Post(url string, token string, payload map[string]interface{}) *request.Response {
	url = formatUrl(url)
	response := frisby.Create("POST Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetProxy(os.Getenv("https_proxy")).
		SetJson(payload).
		Post(url).
		Send().
		Resp

	if response != nil && response.Response != nil {
		fmt.Println("URL: ", url)
		fmt.Println("inner url: ", response.Request.URL)
	}

	return response
}

func Put(url string, token string, payload map[string]interface{}) *request.Response {
	url = formatUrl(url)
	response := frisby.Create("GET Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetProxy(os.Getenv("https_proxy")).
		SetJson(payload).
		Put(url).
		Send().
		Resp

	if response != nil && response.Response != nil {
		fmt.Println("URL: ", url)
		fmt.Println("inner url: ", response.Request.URL)
	}

	return response
}

func Delete(url string, token string) *request.Response {
	url = formatUrl(url)
	response := frisby.Create("GET Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetProxy(os.Getenv("https_proxy")).
		Delete(url).
		Send().
		Resp

	if response != nil && response.Response != nil {
		fmt.Println("URL: ", url)
		fmt.Println("inner url: ", response.Request.URL)
	}

	return response
}

func Delete_With_Json(url string, token string, payload map[string]interface{}) *request.Response {
	response := frisby.Create("Get Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetProxy(os.Getenv("https_proxy")).
		SetJson(payload).
		Delete(url).
		Send().
		Resp

	return response
}

func LogFrisbyInfo(response *request.Response, message string) (int, string) {
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Failed to read %s Response... Err: %s", message, err.Error())
		return 0, "Failed"
	}
	defer response.Body.Close()

	fmt.Println("Response code is : " + strconv.Itoa(response.StatusCode))
	fmt.Println("Response body is : " + string(responseBody))

	return response.StatusCode, string(responseBody)
}

func PostCognito(url string, client_id string, client_secret string, grant_type string, scope string) *request.Response {
	response := frisby.Create("Post Call Cognito endpoint").
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		BasicAuth(client_id, client_secret).
		SetData("grant_type", "client_credentials").
		SetData("client_id", client_id).
		SetData("client_secret", client_secret).
		SetData("scope", scope).
		SetProxy(os.Getenv("https_proxy")).
		Post(url).
		Send().
		Resp

	return response
}
