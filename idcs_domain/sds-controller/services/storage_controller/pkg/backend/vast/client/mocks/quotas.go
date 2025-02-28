// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	client "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
)

var quota = client.Quota{
	HardLimit: w(uint64(100000000)),
	Name:      "ns",
	Id:        w(1),
}

func (*MockVastClient) QuotasCreateWithResponse(ctx context.Context, body client.QuotasCreateJSONRequestBody, reqEditors ...client.RequestEditorFn) (*client.QuotasCreateResponse, error) {
	if ctx.Value("testing.QuotasCreateWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.QuotasCreateWithResponse.empty") != nil {
		return &client.QuotasCreateResponse{}, nil
	}

	return &client.QuotasCreateResponse{
		HTTPResponse: &http.Response{
			StatusCode: 201,
		},
		JSON201: &quota,
	}, nil
}

func (*MockVastClient) QuotasListWithResponse(ctx context.Context, params *client.QuotasListParams, reqEditors ...client.RequestEditorFn) (*client.QuotasListResponse, error) {
	if ctx.Value("testing.QuotasListWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.QuotasListWithResponse.empty") != nil {
		return &client.QuotasListResponse{}, nil
	}

	return &client.QuotasListResponse{
		HTTPResponse: &http.Response{
			StatusCode: 201,
		},
		JSON200: &[]client.Quota{quota},
	}, nil
}

func (*MockVastClient) QuotasDeleteWithResponse(ctx context.Context, id int, reqEditors ...client.RequestEditorFn) (*client.QuotasDeleteResponse, error) {
	if ctx.Value("testing.QuotasDeleteWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.QuotasDeleteWithResponse.empty") != nil {
		return &client.QuotasDeleteResponse{}, nil
	}

	return &client.QuotasDeleteResponse{
		HTTPResponse: &http.Response{
			StatusCode: 204,
		},
	}, nil
}

func (*MockVastClient) QuotasPartialUpdateWithResponse(ctx context.Context, id int, body client.QuotasPartialUpdateJSONRequestBody, reqEditors ...client.RequestEditorFn) (*client.QuotasPartialUpdateResponse, error) {
	if ctx.Value("testing.QuotasPartialUpdateWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.QuotasPartialUpdateWithResponse.empty") != nil {
		return &client.QuotasPartialUpdateResponse{}, nil
	}

	return &client.QuotasPartialUpdateResponse{
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}, nil
}
