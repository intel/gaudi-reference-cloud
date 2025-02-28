package authentication

import (
	"crypto/tls"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/frisby_client"
	"net/http"

	"github.com/verdverm/frisby"
)

func GetBearerTokenViaFrisby(cluster_name string) string {
	//https: //internal-placeholder.com/
	base_url := cluster_name + "/token?email=admin@intel.com&groups=IDC.Admin"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	fmt.Println("base_url", base_url)
	frisby_response := frisby.Create("GET Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		Get(base_url).
		Send().
		Resp
	fmt.Println("frisby_response", frisby_response)
	responseCode, responseBody := frisby_client.LogFrisbyInfo(frisby_response, "GET API")
	fmt.Println("responseBody", responseBody)
	fmt.Println("responseCode", responseCode)
	if responseCode == 200 {
		return "Bearer " + string(responseBody)
	}
	logger.Log.Error("Unable to get the authentication token")
	return "null"
}

/*func GetBearerTokenViaHttpClient(cluster_name string) string {
	//https: //internal-placeholder.com/token?preferred_username=admin@intel.com&roles=IDC.Admin
	base_url := cluster_name + "/token?preferred_username=admin@intel.com&roles=IDC.Admin"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, _ := http.NewRequest(http.MethodGet, base_url, nil)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Error("GET Request Creation Failed , With Error " + err.Error())
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		return "Bearer " + string(body)
	}
	defer resp.Body.Close()
	logger.Log.Error("Unable to get the authentication token")
	return "null"
}*/
