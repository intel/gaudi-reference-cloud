package auth

import (
	"crypto/tls"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/client"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

func GetBearerTokenViaResty(base_url string) string {
	//base_url := test_url + "/token?email=admin@intel.com&groups=IDC.Admin"
	fmt.Println("base_url", base_url)

	restyClient := resty.New().SetProxy(os.Getenv("https_proxy")).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	restyClient.RemoveProxy()

	response, err := restyClient.R().
		SetHeader("Content-Type", "application/json").
		Post(base_url)

	if err != nil {
		fmt.Println(err)
	}

	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	// fmt.Println("responseBody", responseBody)
	fmt.Println("responseCode", responseCode)
	if responseCode == 200 {
		return "Bearer " + string(responseBody)
	}
	fmt.Println("Unable to get the authentication token")
	return "null"
}

func GetUserTokenViaResty(oidc_url string, user_email string) string {
	base_url := strings.Replace(oidc_url, "admin@intel.com", user_email, 1)
	base_url = strings.Replace(base_url, "IDC.Admin", "IDC.User", 1)
	fmt.Println("base_url", base_url)

	restyClient := resty.New().SetProxy(os.Getenv("https_proxy")).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	restyClient.RemoveProxy()

	response, err := restyClient.R().
		SetHeader("Content-Type", "application/json").
		Post(base_url)

	if err != nil {
		fmt.Println(err)
	}

	responseCode, responseBody := client.LogRestyInfo(response, "GET API")
	// fmt.Println("responseBody", responseBody)
	fmt.Println("responseCode", responseCode)
	if responseCode == 200 {
		return "Bearer " + string(responseBody)
	}
	fmt.Println("Unable to get the authentication token")
	return "null"
}
