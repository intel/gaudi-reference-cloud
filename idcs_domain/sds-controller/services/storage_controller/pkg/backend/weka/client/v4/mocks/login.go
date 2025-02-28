// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

func (*MockWekaClient) LoginWithResponse(ctx context.Context, body v4.LoginJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.LoginResponse, error) {
	if body.Password == "error" {
		return nil, errors.New("auth error")
	}

	if body.Password != "password" {
		return &v4.LoginResponse{
			JSON401: &v4.N401{
				Data: w("auth error"),
			},
			HTTPResponse: &http.Response{
				StatusCode: 401,
			},
		}, nil
	}

	return &v4.LoginResponse{
		JSON200: &struct {
			Data *v4.Tokens "json:\"data,omitempty\""
		}{
			&v4.Tokens{
				AccessToken:            w("token"),
				ExpiresIn:              w(5000),
				PasswordChangeRequired: w(false),
				RefreshToken:           w("refresh"),
				TokenType:              w("Bearer"),
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}, nil
}
