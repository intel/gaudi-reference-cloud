package http_client

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/utils"
	"io/ioutil"
	"net/http"
	"os"
)

func Get_Token(expected_status_code int) string {
	var token string
	logger.Log.Info("Fetching JWT Token")
	if expected_status_code == 401 {
		token = "Bearer abcd.token.com"
	} else if expected_status_code == 503 {
		token = "Bearer " + utils.Get_Expired_Bearer_Token()
	} else if _, ok := os.LookupEnv("fetch_admin_role_token"); ok {
		token = "Bearer " + utils.Get_Admin_Role_Token()
	} else if _, ok := os.LookupEnv("cloudAccTest"); ok {
		token = "Bearer " + os.Getenv("cloudAccToken")
	} else if _, ok := os.LookupEnv("UseUserToken"); ok {
		token = "Bearer " + os.Getenv("UserTokenVal")
	} else {
		token = "Bearer " + utils.Get_Bearer_Token()
	}
	return token
}

func Post_Azure(url string, data *bytes.Buffer, bearer_token string, expected_status_code int) (string, int) {
	var jsonStr string
	logger.Log.Info("POST URL :" + url)
	logger.Log.Info("Request Body  " + data.String())
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodPost, url, data)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Info("POST Request Failed with Error" + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err1 := http.DefaultClient.Do(request)
	if err1 != nil {
		logger.Log.Info("POST Request Failed with Error" + err1.Error())
		return jsonStr, resp.StatusCode
	}
	//Need to close the response stream, once response is read.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Info("Failed to read POST response, Err: " + err.Error())
		return jsonStr, resp.StatusCode
	}

	//Convert bytes to String and print
	jsonStr = string(body)
	// All is OK, server reported success.
	//Check response code, if New user is created then read response.
	if expected_status_code == http.StatusCreated || expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusCreated || expected_status_code == http.StatusOK {
			logger.Log.Info("Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Info("Response: " + jsonStr)
			logger.Log.Info("POST Failed with Error" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode

		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Info("POST Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			logger.Log.Info("POST Request Failed for negative test, Got response" + jsonStr)
			return jsonStr, resp.StatusCode
		}
		logger.Log.Info("POST Request Response for test, Got response" + jsonStr)
		logger.Logf.Info("POST Request Response Code for test, Got response Code", resp.StatusCode)
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode
}

func Get_Response(url string, userEmail string) int {
	bearer_token := "Bearer " + utils.Get_Bearer_Token()
	logger.Logf.Info("Bearer Token :", bearer_token)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Error("GET Request Failed , With Error " + err.Error())
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Error("GET Request Creation Failed , With Error " + err.Error())
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Error("Failed to read GET response, Err: " + err.Error())
		return 0
	}
	// Log the request body
	jsonStr := string(body)
	logger.Log.Info("Get Response: " + jsonStr)
	defer resp.Body.Close()
	return resp.StatusCode
}

func Get(url string, expected_status_code int) (string, int) {
	bearer_token := Get_Token(expected_status_code)
	var jsonStr string
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Error("GET Request Failed , With Error " + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Error("GET Request Failed , With Error " + err.Error())
		return jsonStr, resp.StatusCode
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Error("Failed to read GET response, Err: " + err.Error())
		return jsonStr, resp.StatusCode
	}
	// Log the request body
	jsonStr = string(body)
	if expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusOK {

			logger.Log.Info("Get Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Error("GET Failed with Response Code " + fmt.Sprint(resp.StatusCode))
			logger.Log.Error("GET Failed with Response Code " + fmt.Sprint(resp.Body))
			return jsonStr, resp.StatusCode
		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Error("GET Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode
		}
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode
}

func GetOIDC(url string, bearer_token string, expected_status_code int) (string, int) {
	var jsonStr string
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	logger.Log.Info("Token  " + bearer_token)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Error("GET Request Failed , With Error " + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Error("GET Request Failed , With Error " + err.Error())
		return jsonStr, resp.StatusCode
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Error("Failed to read GET response, Err: " + err.Error())
		return jsonStr, resp.StatusCode
	}
	// Log the request body
	jsonStr = string(body)
	if expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusOK {

			logger.Log.Info("Get Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Error("GET Failed with Response Code " + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode
		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Error("GET Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode
		}
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode
}

func Get_With_Payload_OIDC(url string, data *bytes.Buffer, expected_status_code int, bearer_token string) (string, int) {
	var jsonStr string
	logger.Log.Error("GET Request URL " + url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodGet, url, data)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Error("GET Request Failed , With Error " + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Error("GET Request Failed , With Error " + err.Error())
		return jsonStr, resp.StatusCode
	}
	defer resp.Body.Close()

	if expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				//Failed to read response.
				logger.Log.Error("Failed to read GET response, Err: " + err.Error())
				return jsonStr, resp.StatusCode
			}
			// Log the request body
			jsonStr = string(body)
			logger.Log.Info("Get Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Error("GET Request Failed , With Error " + err.Error())
			logger.Log.Error("GET Failed with Response Code " + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode
		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Error("GET Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode
		}
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode
}

func Get_With_Payload(url string, data *bytes.Buffer, expected_status_code int) (string, int) {
	bearer_token := Get_Token(expected_status_code)
	var jsonStr string
	logger.Log.Error("GET Request URL " + url)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodGet, url, data)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Error("GET Request Failed , With Error " + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Error("GET Request Failed , With Error " + err.Error())
		return jsonStr, resp.StatusCode
	}
	defer resp.Body.Close()

	if expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				//Failed to read response.
				logger.Log.Error("Failed to read GET response, Err: " + err.Error())
				return jsonStr, resp.StatusCode
			}
			// Log the request body
			jsonStr = string(body)
			logger.Log.Info("Get Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Error("GET Request Failed , With Error " + err.Error())
			logger.Log.Error("GET Failed with Response Code " + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode
		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Error("GET Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode
		}
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode
}

func Post(url string, data *bytes.Buffer, expected_status_code int) (string, int) {
	bearer_token := Get_Token(expected_status_code)
	var jsonStr string
	logger.Log.Info("POST URL :" + url)
	logger.Log.Info("Request Body  " + data.String())
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodPost, url, data)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Info("POST Request Failed with Error" + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err1 := http.DefaultClient.Do(request)
	if err1 != nil {
		logger.Log.Info("POST Request Failed with Error" + err1.Error())
		return jsonStr, resp.StatusCode
	}
	//Need to close the response stream, once response is read.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Info("Failed to read POST response, Err: " + err.Error())
		return jsonStr, resp.StatusCode
	}

	//Convert bytes to String and print
	jsonStr = string(body)
	// All is OK, server reported success.
	//Check response code, if New user is created then read response.
	if expected_status_code == http.StatusCreated || expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusCreated || expected_status_code == http.StatusOK {
			logger.Log.Info("Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Info("Response: " + jsonStr)
			logger.Log.Info("POST Failed with Error" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode

		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Info("POST Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			logger.Log.Info("POST Request Failed for negative test, Got response" + jsonStr)
			return jsonStr, resp.StatusCode
		}
		logger.Log.Info("POST Request Response for test, Got response" + jsonStr)
		logger.Logf.Info("POST Request Response Code for test, Got response Code", resp.StatusCode)
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode
}

func PostOIDC(url string, data *bytes.Buffer, expected_status_code int, token string) (string, int) {
	bearer_token := token
	var jsonStr string
	logger.Log.Info("POST URL :" + url)
	logger.Log.Info("Request Body  " + data.String())
	logger.Log.Info("Token  " + token)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodPost, url, data)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Info("POST Request Failed with Error" + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err1 := http.DefaultClient.Do(request)
	if err1 != nil {
		logger.Log.Info("POST Request Failed with Error" + err1.Error())
		return jsonStr, resp.StatusCode
	}
	//Need to close the response stream, once response is read.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Info("Failed to read POST response, Err: " + err.Error())
		return jsonStr, resp.StatusCode
	}

	//Convert bytes to String and print
	jsonStr = string(body)
	// All is OK, server reported success.
	//Check response code, if New user is created then read response.
	if expected_status_code == http.StatusCreated || expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusCreated || expected_status_code == http.StatusOK {
			logger.Log.Info("Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Info("Response: " + jsonStr)
			logger.Log.Info("POST Failed with Error" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode

		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Info("POST Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			logger.Log.Info("POST Request Failed for negative test, Got response" + jsonStr)
			return jsonStr, resp.StatusCode
		}
		logger.Log.Info("POST Request Response for test, Got response" + jsonStr)
		logger.Logf.Info("POST Request Response Code for test, Got response Code", resp.StatusCode)
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode
}

// function to make a post for product sync in PC (Product catalog e2e tests)
func PostPC(url string, data *bytes.Buffer, expected_status_code int) (string, int) {
	bearer_token := auth.Get_Azure_Admin_Bearer_Token()
	var jsonStr string
	logger.Log.Info("POST URL :" + url)
	logger.Log.Info("Request Body  " + data.String())
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodPost, url, data)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Info("POST Request Failed with Error" + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err1 := http.DefaultClient.Do(request)
	if err1 != nil {
		logger.Log.Info("POST Request Failed with Error" + err1.Error())
		return jsonStr, resp.StatusCode
	}
	//Need to close the response stream, once response is read.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Info("Failed to read POST response, Err: " + err.Error())
		return jsonStr, resp.StatusCode
	}

	//Convert bytes to String and print
	jsonStr = string(body)
	// All is OK, server reported success.
	//Check response code, if New user is created then read response.
	if expected_status_code == http.StatusCreated || expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusCreated || expected_status_code == http.StatusOK {
			logger.Log.Info("Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Info("Response: " + jsonStr)
			logger.Log.Info("POST Failed with Error" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode

		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Info("POST Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			logger.Log.Info("POST Request Failed for negative test, Got response" + jsonStr)
			return jsonStr, resp.StatusCode
		}
		logger.Log.Info("POST Request Response for test, Got response" + jsonStr)
		logger.Logf.Info("POST Request Response Code for test, Got response Code", resp.StatusCode)
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode
}

func Put(url string, data *bytes.Buffer, expected_status_code int) (string, int) {
	bearer_token := Get_Token(expected_status_code)
	var jsonStr string
	logger.Log.Info("Bearer Token" + bearer_token)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodPut, url, data)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Info("PUT Request Failed with Error" + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Info("PUT Request Failed with Error" + err.Error())
		return jsonStr, resp.StatusCode
	}
	//Need to close the response stream, once response is read.
	defer resp.Body.Close()

	// All is OK, server reported success.
	//Check response code, if New user is created then read response.
	if expected_status_code == http.StatusOK {
		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				//Failed to read response.
				logger.Log.Info("Failed to read PUT response, Err: " + err.Error())
				return jsonStr, resp.StatusCode
			}

			//Convert bytes to String and print
			jsonStr = string(body)
			logger.Log.Info("Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Info("Response: " + jsonStr)
			logger.Log.Info("PUT Failed with Error" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode

		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Info("PUT Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode
		}
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode

}

func Patch(url string, data *bytes.Buffer, expected_status_code int) (string, int) {
	bearer_token := Get_Token(expected_status_code)
	var jsonStr string
	logger.Log.Info("POST URL :" + url)
	logger.Log.Info("Request Body  " + data.String())
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodPatch, url, data)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Info("PATCH Request Failed with Error" + err.Error())
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Info("PATCH Request Failed with Error" + err.Error())
	}
	//Need to close the response stream, once response is read.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//Failed to read response.
		logger.Log.Info("Failed to read PATCH response, Err: " + err.Error())
	}

	//Convert bytes to String and print
	jsonStr = string(body)
	// All is OK, server reported success.
	//Check response code, if New user is created then read response.
	if expected_status_code == http.StatusCreated {
		if resp.StatusCode == http.StatusCreated {

			logger.Log.Info("Response: " + jsonStr)

		} else {
			//The status is not Created. print the error.
			logger.Log.Info("PATCH Failed with Error" + fmt.Sprint(resp.StatusCode))
		}
	} else {
		//Check response code in case of negative test
		if resp.StatusCode != expected_status_code {
			logger.Log.Info("PATCH Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
			return jsonStr, resp.StatusCode
		}
		return jsonStr, resp.StatusCode
	}
	return jsonStr, resp.StatusCode

}

func Delete(url string, expected_status_code int) (string, int) {
	bearer_token := Get_Token(expected_status_code)
	var jsonStr string
	logger.Log.Info("Bearer Token" + bearer_token)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	// An error is returned if something goes wrong
	if err != nil {
		logger.Log.Info("DELETE Request Failed with Error" + err.Error())
		return "Failed", 0
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Add("Authorization", bearer_token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Log.Info("POST Request Failed with Error" + err.Error())
		return jsonStr, resp.StatusCode
	}

	//Need to close the response stream, once response is read.

	defer resp.Body.Close()

	// All is OK, server reported success.
	//Check response code, if New user is created then read response.
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			//Failed to read response.
			logger.Log.Info("Failed to read POST response, Err: " + err.Error())
			return jsonStr, resp.StatusCode
		}

		//Convert bytes to String and print
		jsonStr = string(body)
		logger.Log.Info("Response: " + jsonStr)

	} else {
		if expected_status_code != http.StatusOK {
			//Check response code in case of negative test
			if resp.StatusCode != expected_status_code {
				logger.Log.Info("Request Failed for negative test, Got response" + fmt.Sprint(resp.StatusCode))
				return jsonStr, resp.StatusCode
			}
			return jsonStr, resp.StatusCode
		}
		//The status is not Created. print the error.
		logger.Log.Info("Delete failed with error: " + resp.Status)
		return jsonStr, resp.StatusCode

	}
	return jsonStr, resp.StatusCode
}
