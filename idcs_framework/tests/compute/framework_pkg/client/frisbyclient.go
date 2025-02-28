package client

/*import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/mozillazg/request"
	"github.com/verdverm/frisby"
)

func Get(url string, token string) *request.Response {
	response := frisby.Create("GET Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetProxy(os.Getenv("https_proxy")).
		Get(url).
		Send().
		Resp

	return response
}

func Post(url string, token string, payload map[string]interface{}) *request.Response {
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

func Put(url string, token string, payload map[string]interface{}) *request.Response {
	response := frisby.Create("GET Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetProxy(os.Getenv("https_proxy")).
		SetJson(payload).
		Put(url).
		Send().
		Resp

	return response
}

func Delete(url string, token string) *request.Response {
	response := frisby.Create("GET Call for REST-API Endpoint").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetProxy(os.Getenv("https_proxy")).
		Delete(url).
		Send().
		Resp

	return response
}

func LogFrisbyInfo(response *request.Response, message string) (int, string) {
	responseBody, err := io.ReadAll(response.Body)
	fmt.Println("Response code is : " + strconv.Itoa(response.StatusCode))
	fmt.Println("Response body is : " + string(responseBody))
	if err != nil {
		//logger.Log.Error("Failed to read " + message + " Response... Err: " + err.Error())
		return 0, "Failed"
	}
	defer response.Body.Close()
	return response.StatusCode, string(responseBody)
}*/
