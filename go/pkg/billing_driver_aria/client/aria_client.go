// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

// Todo: This is a very initial implementation

type AriaClient struct {
	serverUrl string
	apiPrefix string
	// Todo: Fix for production.
	insecureSsl bool
}

func NewAriaClient(serverUrl string, apiPrefix string, insecureSsl bool) *AriaClient {
	return &AriaClient{
		serverUrl:   serverUrl,
		apiPrefix:   apiPrefix,
		insecureSsl: insecureSsl,
	}
}

type AriaAdminClient struct {
	AriaClient
}

func NewAriaAdminClient(serverUrl string, insecureSsl bool) *AriaAdminClient {
	return &AriaAdminClient{
		AriaClient{
			serverUrl:   serverUrl,
			insecureSsl: insecureSsl,
		},
	}
}

func (ariaClient *AriaClient) post(ctx context.Context, reqBytes []byte, contentType string) (*http.Response, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaClient.post").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	reqBody := bytes.NewReader(reqBytes)
	server := ariaClient.serverUrl
	connURL := fmt.Sprintf("%s%s", server, ariaClient.apiPrefix)
	logger.Info("POST using", "connection URL", connURL)

	tlsConfig := tls.Config{}
	if !ariaClient.insecureSsl {
		tlsConfig.GetClientCertificate = func(cri *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := tls.LoadX509KeyPair(config.Cfg.GetAriaSystemApiCrtFile(), config.Cfg.GetAriaSystemApiKeyFile())
			if err != nil {
				logger.Error(err, "failed to load X509 key pair")
				logger.Info("aria failed to load X509 key pair, certificates are required")
			}
			return &cert, nil
		}
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tlsConfig

	req, err := http.NewRequest("POST", connURL, reqBody)
	if err != nil {
		logger.Error(err, "Failed creating new http request")
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Timeout: 200 * time.Second}
	return client.Do(req)
}

func CallAria[R response.Response](ctx context.Context, ariaClient *AriaClient, arg request.Request, idcError string) (*R, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaClient.CallAria").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	res, err := callAriaIgnoreAriaFailure[R](ctx, ariaClient, arg, idcError, convertPassThrough, "application/json")
	if err != nil {
		return nil, err
	}
	if (*res).GetErrorCode() != 0 {
		logger.Info("aria returned error code", "restCall", arg.GetRestCall(), "code", (*res).GetErrorCode(), "msg", (*res).GetErrorMsg())
		return res, GetErrorForAriaErrorCode(idcError, arg.GetRestCall(),
			(*res).GetErrorCode(), (*res).GetErrorMsg())
	}
	return res, nil
}

func CallAriaIgnoreAriaFailure[R response.Response](ctx context.Context, ariaClient *AriaClient, arg request.Request, idcError string) (*R, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaClient.CallAriaIgnoreAriaFailure").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	return callAriaIgnoreAriaFailure[R](ctx, ariaClient, arg, idcError, convertPassThrough, "application/json")
}

func CallAriaAdmin[R response.Response](ctx context.Context, ariaClient *AriaAdminClient, arg request.Request, idcError string) (*R, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaClient.CallAriaAdmin").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	res, err := callAriaIgnoreAriaFailure[R](ctx, &ariaClient.AriaClient, arg, idcError, convertJsonToQuery, "application/x-www-form-urlencoded")
	if err != nil {
		return nil, err
	}
	if (*res).GetErrorCode() != 0 {
		logger.Info("aria returned error code", "restCall", arg.GetRestCall(), "code", (*res).GetErrorCode(), "msg", (*res).GetErrorMsg())
		return res, GetErrorForAriaErrorCode(idcError, arg.GetRestCall(),
			(*res).GetErrorCode(), (*res).GetErrorMsg())
	}
	return res, nil
}

type tConvertFunc func(context.Context, string, []byte) ([]byte, error)

func convertPassThrough(ctx context.Context, restCall string, in []byte) ([]byte, error) {
	DebugLogJsonPayload(ctx, restCall+" query", in)
	return in, nil
}

func convertJsonToQuery(ctx context.Context, restCall string, in []byte) ([]byte, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaClient.convertJsonToQuery").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	out, err := GenerateQuery(in)
	if err != nil {
		logger := log.FromContext(ctx)
		logger.Error(err, "failed to encode json as query string", "restCall", restCall)
		return nil, err
	}
	DebugLogQueryPayload(ctx, restCall+" query", out)
	return []byte(out), nil
}

func callAriaIgnoreAriaFailure[R response.Response](ctx context.Context, ariaClient *AriaClient, arg request.Request, idcError string, convert tConvertFunc, contentType string) (*R, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaClient.callAriaIgnoreAriaFailure").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	reqJson, err := json.Marshal(arg)
	if err != nil {
		logger.Error(err, "failed to marshal request to json", "restCall", arg.GetRestCall())
		return nil, err
	}

	postBody, err := convert(ctx, arg.GetRestCall(), reqJson)
	if err != nil {
		return nil, err
	}

	resp, err := ariaClient.post(ctx, postBody, contentType)
	if err != nil {
		logger.Error(err, "aria call failed", "restCall", arg.GetRestCall())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Info("status from aria request", "status", resp.Status, "restCall", arg.GetRestCall())
		return nil, GetErrorForStatusCode(arg.GetRestCall(), resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Info("failed to read response from aria", "restCall", arg.GetRestCall())
	}
	logger.Info("response body", "body", string(respBody))

	var res R
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		logger.Error(err, "json unmarshal response", "restCall", arg.GetRestCall())
		return nil, err
	}

	return &res, nil
}
