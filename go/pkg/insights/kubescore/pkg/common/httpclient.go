// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package common

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

// MakePUTAPICall :
func MakePUTAPICall(ctx context.Context, server, uri string, payload []byte, payloadType string) (int, error) {
	logger := log.FromContext(ctx).WithName("common.MakePUTAPICall")
	connURL := fmt.Sprintf("%s%s", server, uri)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("PUT", connURL, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error(err, "failed to create http PUT request")
	}
	req.Header.Set("Content-Type", "application/json")
	retries := 3
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 300 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				logger.Info("PUT", "trial", try, "error making http call: ", err)
				return http.StatusInternalServerError,
					errors.New("error conencting to gitsecure api service")
			}
			logger.Info("PUT", "trial", try, "error making http call: ", err)
			logger.Info("PUT", "trying again after seconds", 5)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		retcode = resp.StatusCode
		break
	}
	return retcode, nil
}

// MakeGetAPICall :
func MakeGetAPICall(ctx context.Context, server, uri string, payload []byte) (int, []byte, error) {
	logger := log.FromContext(ctx).WithName("common.MakeGetAPICall")
	connURL := fmt.Sprintf("%s%s", server, uri)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("GET", connURL, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error(err, "failed to create http GET request")
	}
	req.Header.Set("Content-Type", "application/json")
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 300 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				logger.Info("GET", "trial", try, "error making http call: ", err)
				return http.StatusInternalServerError, nil,
					errors.New("error conencting to gitsecure api service")
			}
			logger.Info("GET", "trial", try, "error making http call: ", err)
			logger.Info("GET", "trying again after seconds", 5)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			logger.Error(err, "failed to read resp body")
		}
		retcode = resp.StatusCode
		break
	}

	return retcode, body, nil
}

// MakePOSTAPICall :
func MakePOSTAPICall(ctx context.Context, server, url string, payload []byte) (int, []byte, error) {
	logger := log.FromContext(ctx).WithName("common.MakeGetAPICall")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error(err, "failed to create http POST request")
	}
	req.Header.Set("Content-Type", "application/json")
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 300 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				logger.Info("PUT", "trial", try, "error making http call: ", err)
				return http.StatusInternalServerError, nil,
					errors.New("error conencting to gitsecure api service")
			}
			logger.Info("PUT", "trial", try, "error making http call: ", err)
			logger.Info("PUT", "trying again after seconds", 5)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			logger.Error(err, "failed to read resp body")
		}
		retcode = resp.StatusCode
		break
	}
	return retcode, body, nil
}

// MakeDeleteAPICall :
func MakeDeleteAPICall(ctx context.Context, server, uri string, payload []byte) (int, []byte, error) {
	logger := log.FromContext(ctx).WithName("common.MakeGetAPICall")
	connURL := fmt.Sprintf("%s%s", server, uri)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, err := http.NewRequest("DELETE", connURL, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error(err, "failed to create http DELETE request")
	}
	req.Header.Set("Content-Type", "application/json")
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 300 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				logger.Info("PUT", "trial", try, "error making http call: ", err)
				return http.StatusInternalServerError, nil,
					errors.New("error conencting to gitsecure api service")
			}
			logger.Info("PUT", "trial", try, "error making http call: ", err)
			logger.Info("PUT", "trying again after seconds", 5)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			logger.Error(err, "failed to read response body")
		}
		retcode = resp.StatusCode
		break
	}
	return retcode, body, nil
}
