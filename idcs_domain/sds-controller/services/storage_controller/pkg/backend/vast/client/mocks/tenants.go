// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	client "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
)

var tenant = client.Tenant{
	Id:   w(1),
	Name: "ns",
}

func (*MockVastClient) TenantsCreateWithResponse(ctx context.Context, body client.TenantsCreateJSONRequestBody, reqEditors ...client.RequestEditorFn) (*client.TenantsCreateResponse, error) {
	if ctx.Value("testing.TenantsCreateWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.TenantsCreateWithResponse.empty") != nil {
		return &client.TenantsCreateResponse{}, nil
	}

	if ctx.Value("testing.TenantsCreateWithResponse.incomplete") != nil {
		incomplete := client.Tenant{
			Name: "ns",
		}
		return &client.TenantsCreateResponse{
			HTTPResponse: &http.Response{StatusCode: 201},
			JSON201:      &incomplete,
		}, nil
	}

	return &client.TenantsCreateResponse{
		HTTPResponse: &http.Response{StatusCode: 201},
		JSON201:      &tenant,
	}, nil
}

func (*MockVastClient) TenantsReadWithResponse(ctx context.Context, id int, reqEditors ...client.RequestEditorFn) (*client.TenantsReadResponse, error) {
	if ctx.Value("testing.TenantsReadWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.TenantsReadWithResponse.empty") != nil {
		return &client.TenantsReadResponse{}, nil
	}

	return &client.TenantsReadResponse{
		HTTPResponse: &http.Response{StatusCode: 200},
		JSON200:      &tenant,
	}, nil
}

func (*MockVastClient) TenantsDeleteWithResponse(ctx context.Context, id int, reqEditors ...client.RequestEditorFn) (*client.TenantsDeleteResponse, error) {
	if ctx.Value("testing.TenantsDeleteWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.TenantsDeleteWithResponse.empty") != nil {
		return &client.TenantsDeleteResponse{}, nil
	}

	return &client.TenantsDeleteResponse{
		HTTPResponse: &http.Response{
			StatusCode: 204,
		},
	}, nil
}

func (*MockVastClient) TenantsListWithResponse(ctx context.Context, params *client.TenantsListParams, reqEditors ...client.RequestEditorFn) (*client.TenantsListResponse, error) {
	if ctx.Value("testing.TenantsListWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.TenantsListWithResponse.empty") != nil {
		return &client.TenantsListResponse{}, nil
	}

	return &client.TenantsListResponse{
		JSON200:      &[]client.Tenant{tenant},
		HTTPResponse: &http.Response{StatusCode: 200},
	}, nil
}
