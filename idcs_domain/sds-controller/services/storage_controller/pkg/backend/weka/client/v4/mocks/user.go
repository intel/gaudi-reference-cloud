// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

func (*MockWekaClient) CreateUserWithResponse(ctx context.Context, body v4.CreateUserJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.CreateUserResponse, error) {

	resp := v4.CreateUserResponse{
		JSON200: &struct {
			Data *v4.User "json:\"data,omitempty\""
		}{
			Data: &v4.User{
				OrgId:    w(0),
				Role:     w("OrgAdmin"),
				Username: w("username"),
				Uid:      w("userId"),
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.CreateUserWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.CreateUserWithResponse.empty") != nil {
		return &v4.CreateUserResponse{}, nil
	}

	if ctx.Value("testing.CreateUserWithResponse.regular") != nil {
		resp.JSON200.Data.Role = w("Regular")
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) DeleteUserWithResponse(ctx context.Context, uid string, reqEditors ...v4.RequestEditorFn) (*v4.DeleteUserResponse, error) {

	resp := v4.DeleteUserResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.DeleteUserWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.DeleteUserWithResponse.empty") != nil {
		return &v4.DeleteUserResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) GetUsersWithResponse(ctx context.Context, reqEditors ...v4.RequestEditorFn) (*v4.GetUsersResponse, error) {

	resp := v4.GetUsersResponse{
		JSON200: &struct {
			Data *[]v4.User "json:\"data,omitempty\""
		}{
			Data: &[]v4.User{
				{
					Uid:      w("userId"),
					OrgId:    w(0),
					Role:     w("OrgAdmin"),
					Username: w("username"),
				},
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetUsersWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetUsersWithResponse.empty") != nil {
		return &v4.GetUsersResponse{}, nil
	}

	if ctx.Value("testing.GetUsersWithResponse.incomplete") != nil {
		users := *resp.JSON200.Data
		users[0].Uid = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) UpdateUserWithResponse(ctx context.Context, uid string, body v4.UpdateUserJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.UpdateUserResponse, error) {
	resp := v4.UpdateUserResponse{
		JSON200: &struct {
			Data *v4.User "json:\"data,omitempty\""
		}{
			Data: &v4.User{
				Uid:      w("userId"),
				OrgId:    w(0),
				Role:     w("OrgAdmin"),
				Username: w("username"),
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.UpdateUserWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.UpdateUserWithResponse.empty") != nil {
		return &v4.UpdateUserResponse{}, nil
	}

	if ctx.Value("testing.UpdateUserWithResponse.incomplete") != nil {
		resp.JSON200.Data.Uid = nil
		return &resp, nil
	}

	if ctx.Value("testing.UpdateUserWithResponse.regular") != nil {
		resp.JSON200.Data.Role = w("Regular")
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) SetUserPasswordWithResponse(ctx context.Context, uid string, body v4.SetUserPasswordJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.SetUserPasswordResponse, error) {
	resp := v4.SetUserPasswordResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.SetUserPasswordWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.SetUserPasswordWithResponse.empty") != nil {
		return &v4.SetUserPasswordResponse{}, nil
	}

	return &resp, nil
}
