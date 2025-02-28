// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package restyclient

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

// HTTPClient is a simple utility for making HTTP requests using resty.
type RestyClient struct {
}

// Request makes an HTTP request with the specified method, URL, headers, and payload using resty.
func (c *RestyClient) Request(ctx context.Context, method, url string, payload []byte) (*resty.Response, error) {
	log := log.FromContext(ctx).WithName("TestComputeService.PublicAPI")
	request := resty.New().SetProxy("http://internal-placeholder.com:912").SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).RemoveProxy().R().
		SetBody(string(payload)).
		SetHeader("Content-Type", "application/json")

	log.Info("Request details : ", "Method", method, "Endpoint", url, "payload", string(payload))
	var response *resty.Response
	var err error

	switch method {
	case "GET":
		response, err = request.Get(url)
	case "POST":
		response, err = request.Post(url)
	case "PUT":
		response, err = request.Put(url)
	case "DELETE":
		response, err = request.Delete(url)
	case "PATCH":
		response, err = request.Patch(url)
	case "HEAD":
		response, err = request.Head(url)
	case "OPTIONS":
		response, err = request.Options(url)
	// Add more cases for other HTTP methods as needed
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}
	if err != nil {
		log.Error(err, "Error while making http method call...")
		return nil, err
	} else {
		log.Info("Response details : ", "status-code", response.StatusCode(), "response-body", string(response.Body()))
	}
	return response, nil
}

func (c *RestyClient) GetBearerToken(ctx context.Context, oidc_url string) string {
	log := log.FromContext(ctx).WithName("TestComputeService.BearerToken")
	restyClient := resty.New().SetProxy("http://internal-placeholder.com:912").SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).RemoveProxy()
	response, err := restyClient.R().Post(oidc_url)
	if err != nil {
		log.Error(err, "Error while fetching bearer token...")
	}
	if response.StatusCode() == 200 {
		return "Bearer " + string(response.Body())
	}
	return "null"
}
