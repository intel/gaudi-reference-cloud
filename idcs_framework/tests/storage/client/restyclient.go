package client

import (
	"os"

	"crypto/tls"

	"github.com/go-resty/resty/v2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/idcs_framework/tests/storage/logger"
	"go.uber.org/zap"
)

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
		logger.Logf.Info("GET API error",
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
		logger.Logf.Info("POST API error",
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
		logger.Logf.Info("PUT API error",
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
		logger.Logf.Info("PATCH API error",
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
		logger.Logf.Info("DELETE API error",
			zap.Error(err),
		)
	}

	return response
}

func LogRestyInfo(response *resty.Response, message string) (int, string) {
	logger.Logf.Info("responseCode:", response.Status(), "body:", string(response.Body()))
	return response.StatusCode(), string(response.Body())
}
