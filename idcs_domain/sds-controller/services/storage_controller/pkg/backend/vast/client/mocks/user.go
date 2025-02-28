// Copyright (C) 2024 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	client "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
)

func (*MockVastClient) ManagersCreateWithResponse(ctx context.Context, body client.ManagersCreateJSONRequestBody, reqEditors ...client.RequestEditorFn) (*client.ManagersCreateResponse, error) {

	resp := client.ManagersCreateResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 201,
		},
		JSON201: &client.Manager{
			Id:       w(101),
			Username: *w("username"),
		},
	}

	if ctx.Value("testing.CreateUserWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.CreateUserWithResponse.empty") != nil {
		return &client.ManagersCreateResponse{}, nil
	}

	return &resp, nil
}

func (*MockVastClient) ManagersReadWithResponse(ctx context.Context, id int, reqEditors ...client.RequestEditorFn) (*client.ManagersReadResponse, error) {
	// The test id == 101, so return 404 not found for other ids
	if id != 101 {
		return &client.ManagersReadResponse{
			Body: []byte{},
			HTTPResponse: &http.Response{
				StatusCode: 404,
			}}, nil
	}

	resp := client.ManagersReadResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
		JSON200: &client.Manager{
			Id:       w(101),
			Username: *w("username"),
		},
	}

	if ctx.Value("testing.GetUsersWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetUsersWithResponse.empty") != nil {
		return &client.ManagersReadResponse{}, nil
	}

	return &resp, nil
}

func (*MockVastClient) ManagersDeleteWithResponse(ctx context.Context, id int, reqEditors ...client.RequestEditorFn) (*client.ManagersDeleteResponse, error) {
	resp := client.ManagersDeleteResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 204,
		},
	}

	if ctx.Value("testing.DeleteUserWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.DeleteUserWithResponse.empty") != nil {
		return &client.ManagersDeleteResponse{}, nil
	}

	return &resp, nil
}

func (*MockVastClient) ManagersPartialUpdateWithResponse(ctx context.Context, id int, body client.ManagersPartialUpdateJSONRequestBody, reqEditors ...client.RequestEditorFn) (*client.ManagersPartialUpdateResponse, error) {
	resp := client.ManagersPartialUpdateResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
		JSON200: &client.Manager{
			Id:       w(101),
			Username: *w("username"),
		},
	}

	if ctx.Value("testing.SetUserPasswordWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.SetUserPasswordWithResponse.empty") != nil {
		return &client.ManagersPartialUpdateResponse{}, nil
	}

	return &resp, nil
}

func (*MockVastClient) ManagersListWithResponse(ctx context.Context, params *client.ManagersListParams, reqEditors ...client.RequestEditorFn) (*client.ManagersListResponse, error) {

	resp := client.ManagersListResponse{
		Body: []byte{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
		JSON200: &[]client.Manager{
			client.Manager{
				Id:       w(101),
				Username: *w("username"),
				Roles: &[]client.PartialRole{
					client.PartialRole{
						Id: w(1),
					},
				},
			},
		},
	}

	if ctx.Value("testing.GetUsersWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetUsersWithResponse.empty") != nil {
		return &client.ManagersListResponse{}, nil
	}

	if ctx.Value("testing.GetUsersWithResponse.incomplete") != nil {
		return &client.ManagersListResponse{
			Body: []byte{},
			HTTPResponse: &http.Response{
				StatusCode: 200,
			},
			JSON200: &[]client.Manager{},
		}, nil
	}
	return &resp, nil
}
