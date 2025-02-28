package service_apis

/*import (
	"crypto/tls"
	"encoding/json"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"net/http"
	"strings"

	"github.com/verdverm/frisby"
)

func harvesterLoginAndGetCookie(url string, payload map[string]interface{}) string {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	response := frisby.Create("POST Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetJson(payload).
		Post(url).
		Send().
		Resp

	header := response.Header.Get("Set-Cookie")
	cookie := strings.Split(header, ";")[0]
	return cookie
}

func harvesterPostWithCookie(url string, session_cookie string) (int, string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	response := frisby.Create("POST Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Cookie", session_cookie).
		Delete(url).
		Send().
		Resp

	response_code, response_body := frisby_client.LogFrisbyInfo(response, "Delete API")
	return response_code, response_body
}

func DeleteInstanceViaHarvesterApi(login_url string, delete_url string, payload string) (int, string) {
	var jsonMap map[string]interface{}
	json.Unmarshal([]byte(payload), &jsonMap)
	cookie := harvesterLoginAndGetCookie(login_url, jsonMap)
	response_code, response_body := harvesterPostWithCookie(delete_url, cookie)
	return response_code, response_body
}*/
