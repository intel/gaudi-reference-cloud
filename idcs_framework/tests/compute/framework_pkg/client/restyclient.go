package client

import (
	"os"
	"crypto/tls"
	"github.com/go-resty/resty/v2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/compute/framework_pkg/logger"
	"go.uber.org/zap"
)

var logInstance *logger.CustomLogger

func SetLogger(logger *logger.CustomLogger) {
	logInstance = logger
}

func getEnvironment() string {
	environment := os.Getenv("proxy_required")
	return environment
}

func Get(url string, token string) *resty.Response {
	client := resty.New().SetProxy(os.Getenv("https_proxy")).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if getEnvironment() != "true" {
		client.RemoveProxy()
	}

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		Get(url)

	if err != nil {
		logInstance.Println("GET API error",
			zap.Error(err),
		)
	}

	return response
}

func Post(url string, token string, payload map[string]interface{}) *resty.Response {
	client := resty.New().SetProxy(os.Getenv("https_proxy")).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if getEnvironment() != "true" {
		client.RemoveProxy()
	}

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetBody(&payload).
		Post(url)

	if err != nil {
		logInstance.Println("POST API error",
			zap.Error(err),
		)
	}

	return response
}

func Put(url string, token string, payload map[string]interface{}) *resty.Response {
	client := resty.New().SetProxy(os.Getenv("https_proxy")).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if getEnvironment() != "true" {
		client.RemoveProxy()
	}

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetBody(&payload).
		Put(url)

	if err != nil {
		logInstance.Println("PUT API error",
			zap.Error(err),
		)
	}

	return response
}

func Patch(url string, token string, payload map[string]interface{}) *resty.Response {
	client := resty.New().SetProxy(os.Getenv("https_proxy")).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if getEnvironment() != "true" {
		client.RemoveProxy()
	}

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		SetBody(&payload).
		Patch(url)

	if err != nil {
		logInstance.Println("PATCH API error",
			zap.Error(err),
		)
	}

	return response
}

func Delete(url string, token string) *resty.Response {
	client := resty.New().SetProxy(os.Getenv("https_proxy")).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	if getEnvironment() != "true" {
		client.RemoveProxy()
	}

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
		Delete(url)

	if err != nil {
		logInstance.Println("DELETE API error",
			zap.Error(err),
		)
	}

	return response
}

func LogRestyInfo(response *resty.Response, message string) (int, string) {
	/*logger.Log.Info("Response for "+message,
		zap.String("code", response.Status()),
		zap.String("body", string(response.Body())),
	)*/
	logInstance.Println("responseCode:", response.Status(), "body:", string(response.Body()))
	return response.StatusCode(), string(response.Body())
}
