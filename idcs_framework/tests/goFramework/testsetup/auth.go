package testsetup

import (
	"crypto/tls"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/verdverm/frisby"
	"goFramework/framework/common/logger"
	"goFramework/framework/frisby_client"
	"io/ioutil"
	"net/http"
)

func Get_OIDC_Admin_Token() string {
	oidcUrl := gjson.Get(ConfigData, "oidcUrl").String()
	authUrl := oidcUrl + "/token?email=admin@intel.com&groups=IDC.Admin"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	fmt.Println("base_url", authUrl)
	frisby_response := frisby.Create("GET Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		Get(authUrl).
		Send().
		Resp
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	fmt.Println("responseBody", responseBody)
	fmt.Println("responseCode", responseCode)
	if responseCode == 200 {
		return "Bearer " + string(responseBody)
	}
	logger.Log.Error("Unable to get the authentication token")
	return "null"

}

func Get_OIDC_Enroll_Token(oidcUrl string, tid string, username string, enterpriseId string, idp string) string {
	token_base_url := oidcUrl + "/token?"
	var groups = "DevCloud%20Console%20Standard"
	url := token_base_url + "tid=" + tid + "&enterpriseId=" + enterpriseId + "&email=" + username + "&groups=" + groups + "&idp=" + idp
	fmt.Println("CloudAccountTokenUrl : ", url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	// An error is returned if something goes wrong
	if err != nil {
		fmt.Errorf("JWT Genreation Failed with Error " + err.Error())
		panic("Failed to obtain JWT token, Hence Stopping test execution")
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Errorf("JWT Generation Failed with Error " + err.Error())
		panic("Failed to obtain JWT token, Hence Stopping test execution")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		fmt.Errorf("JWT Generation Failed with Error " + err.Error())
		panic("Failed to obtain JWT token, Hence Stopping test execution")
	}
	// Log the request body
	jsonStr := string(body)
	fmt.Println("Token : ", jsonStr)
	return jsonStr
}
