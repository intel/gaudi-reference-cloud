// Copyright (C) 2024 Intel Corporation
package mocks

import (
	"context"
	"net/http"

	client "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
)

func (*MockVastClient) RolesCreateWithResponse(ctx context.Context, body client.Role, reqEditors ...client.RequestEditorFn) (*client.RolesCreateResponse, error) {
	return &client.RolesCreateResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 201,
		},
		JSON201: &client.Role{
			Id: w(1),
		},
	}, nil
}

func (*MockVastClient) RolesPartialUpdateWithResponse(ctx context.Context, id int, body client.RolesPartialUpdateJSONRequestBody, reqEditors ...client.RequestEditorFn) (*client.RolesPartialUpdateResponse, error) {
	return &client.RolesPartialUpdateResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
		JSON200: &client.Role{
			Id: w(1),
		},
	}, nil
}

func (*MockVastClient) RolesDeleteWithResponse(ctx context.Context, id int, reqEditors ...client.RequestEditorFn) (*client.RolesDeleteResponse, error) {
	return &client.RolesDeleteResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 204,
		},
	}, nil
}

func (*MockVastClient) RolesReadWithResponse(ctx context.Context, id int, reqEditors ...client.RequestEditorFn) (*client.RolesReadResponse, error) {
	return &client.RolesReadResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
		JSON200: &client.Role{
			Id:      w(1),
			Tenants: &[]int{*w(100)},
		},
	}, nil
}

func (*MockVastClient) RolesListWithResponse(ctx context.Context, params *client.RolesListParams, reqEditors ...client.RequestEditorFn) (*client.RolesListResponse, error) {
	return &client.RolesListResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
		JSON200: &[]client.Role{
			client.Role{
				Id:      w(1),
				Tenants: &[]int{*w(100)},
			},
		},
	}, nil
}
