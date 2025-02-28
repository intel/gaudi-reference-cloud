// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

// MakePUTAPICall :
func MakePUTAPICall(ctx context.Context, server, uri string, payload []byte, payloadType string) (int, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("common.MakePUTAPICall").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	connURL := fmt.Sprintf("%s%s", server, uri)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	req, err := http.NewRequest("PUT", connURL, bytes.NewBuffer(payload))
	if err != nil {
		// Handle error
		fmt.Println("Error creating request:", err)
		return http.StatusInternalServerError, err
	}
	req.Header.Set("Content-Type", "application/json")
	retries := 3
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 30 * time.Second}
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
func MakeGetAPICall(ctx context.Context, server, uri, auth string, payload []byte) (int, []byte, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("common.MakeGetAPICall").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	connURL := fmt.Sprintf("%s%s", server, uri)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	req, err := http.NewRequest("GET", connURL, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				// Log the info message
				logger.Info("GET", "trial", try, "error making http call: ", err)

				// Create the error to be returned
				err := errors.New("error connecting to gitsecure api service")

				// Return the error and appropriate HTTP status code
				return http.StatusInternalServerError, nil, err
			}
			logger.Info("GET", "trial", try, "error making http call: ", err)
			logger.Info("GET", "trying again after seconds", 5)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			// Handle the error appropriately
			return http.StatusInternalServerError, nil, fmt.Errorf("failed to read response body: %v", err)
		}
		retcode = resp.StatusCode
		break
	}
	return retcode, body, nil
}

// MakePOSTAPICall :
func MakePOSTAPICall(ctx context.Context, server, uri, auth string, payload []byte) (int, []byte, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("common.MakePOSTAPICall").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	cert, err := tls.LoadX509KeyPair("/vault/secrets/api_crt", "/vault/secrets/api_key")
	if err != nil {
		// handle the error for tls.LoadX509KeyPair here
		logger.Error(err, "Error loading X.509 key pair")
		return 0, nil, err
	}
	// caCert, _ := os.ReadFile("/vault/secrets/ca_crt")
	caCert, err := os.ReadFile("/vault/secrets/ca_crt")
	if err != nil {
		// handle the error here
		logger.Error(err, "Error reading file")
		return 0, nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	// if err != nil {
	// 	logger.Info("debug", "crt", err)
	// }
	connURL := fmt.Sprintf("%s%s", server, uri)
	logger.Info("debug", "connection url", connURL)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: caCertPool, Certificates: []tls.Certificate{cert}}
	req, err := http.NewRequest("POST", connURL, bytes.NewBuffer(payload))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				logger.Info("PUT", "trial", try, "error making http call: ", err)

				// Create the error to be returned
				err := errors.New("error connecting to gitsecure api service")

				// Log the error with the appropriate arguments
				logger.Error(err, "Failed to connect to gitsecure api service")

				return http.StatusInternalServerError, nil, err
			}
			logger.Info("PUT", "trial", try, "error making http call: ", err)
			logger.Info("PUT", "trying again after seconds", 5)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			// Handle the error appropriately
			return http.StatusInternalServerError, nil, fmt.Errorf("failed to read response body: %v", err)
		}
		retcode = resp.StatusCode
		break
	}
	return retcode, body, nil
}

// MakeDeleteAPICall :
func MakeDeleteAPICall(ctx context.Context, server, uri string, auth string, payload []byte) (int, []byte, error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("common.MakeDeleteAPICall").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	connURL := fmt.Sprintf("%s%s", server, uri)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{}
	req, err := http.NewRequest("DELETE", connURL, bytes.NewBuffer(payload))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				// Log the info message
				logger.Info("PUT", "trial", try, "error making http call: ", err)

				// Create the error to be returned
				err := errors.New("error connecting to gitsecure api service")

				// Return the error and appropriate HTTP status code
				return http.StatusInternalServerError, nil, err
			}
			logger.Info("PUT", "trial", try, "error making http call: ", err)
			logger.Info("PUT", "trying again after seconds", 5)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			// Handle the error appropriately
			return http.StatusInternalServerError, nil, fmt.Errorf("failed to read response body: %v", err)
		}
		retcode = resp.StatusCode
		break
	}
	return retcode, body, nil
}
