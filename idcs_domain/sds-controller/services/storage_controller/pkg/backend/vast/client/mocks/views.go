// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	client "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
)

var view = client.View{
	Id:       w(2),
	Name:     w("view"),
	Path:     "/view",
	Protocols: &[]client.ViewProtocols{client.NFS},
	PolicyId: 1,
}

func (*MockVastClient) ViewpoliciesCreateWithResponse(ctx context.Context, body client.ViewPolicy, reqEditors ...client.RequestEditorFn) (*client.ViewpoliciesCreateResponse, error) {
	if ctx.Value("testing.ViewpoliciesCreateWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.ViewpoliciesCreateWithResponse.empty") != nil {
		return &client.ViewpoliciesCreateResponse{}, nil
	}

	return &client.ViewpoliciesCreateResponse{
		HTTPResponse: &http.Response{
			StatusCode: 201,
		},
		JSON201: &client.ViewPolicy{
			Id: w(1),
		},
	}, nil
}

func (*MockVastClient) ViewpoliciesDeleteWithResponse(ctx context.Context, id int, reqEditors ...client.RequestEditorFn) (*client.ViewpoliciesDeleteResponse, error) {
	if ctx.Value("testing.ViewpoliciesCreateWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.ViewpoliciesCreateWithResponse.empty") != nil {
		return &client.ViewpoliciesDeleteResponse{}, nil
	}

	return &client.ViewpoliciesDeleteResponse{
		HTTPResponse: &http.Response{
			StatusCode: 204,
		},
	}, nil
}
func (*MockVastClient) ViewsCreateWithResponse(ctx context.Context, body client.ViewsCreateJSONRequestBody, reqEditors ...client.RequestEditorFn) (*client.ViewsCreateResponse, error) {
	if ctx.Value("testing.ViewsCreateWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.ViewsCreateWithResponse.empty") != nil {
		return &client.ViewsCreateResponse{}, nil
	}

	return &client.ViewsCreateResponse{
		HTTPResponse: &http.Response{
			StatusCode: 201,
		},
		JSON201: &view,
	}, nil
}

func (*MockVastClient) ViewsListWithResponse(ctx context.Context, params *client.ViewsListParams, reqEditors ...client.RequestEditorFn) (*client.ViewsListResponse, error) {
	if ctx.Value("testing.ViewsListWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.ViewsListWithResponse.empty") != nil {
		return &client.ViewsListResponse{}, nil
	}

	return &client.ViewsListResponse{
		HTTPResponse: &http.Response{
			StatusCode: 201,
		},
		JSON200: &[]client.View{view},
	}, nil
}

func (*MockVastClient) ViewsDeleteWithResponse(ctx context.Context, id int, reqEditors ...client.RequestEditorFn) (*client.ViewsDeleteResponse, error) {
	if ctx.Value("testing.QuotasDeleteWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.QuotasDeleteWithResponse.empty") != nil {
		return &client.ViewsDeleteResponse{}, nil
	}

	return &client.ViewsDeleteResponse{
		HTTPResponse: &http.Response{
			StatusCode: 204,
		},
	}, nil
}

func (*MockVastClient) ViewsUpdateWithResponse(ctx context.Context, id int, body client.ViewsUpdateJSONRequestBody, reqEditors ...client.RequestEditorFn) (*client.ViewsUpdateResponse, error) {
	if ctx.Value("testing.ViewsUpdateWithResponse.error") != nil {
		return nil, errors.New("something wrong")
	}

	if ctx.Value("testing.ViewsUpdateWithResponse.empty") != nil {
		return &client.ViewsUpdateResponse{}, nil
	}

	return &client.ViewsUpdateResponse{
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}, nil
}
