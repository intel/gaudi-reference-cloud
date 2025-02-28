// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	client "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
)

func (*MockVastClient) TokenCreateWithResponse(ctx context.Context, body client.TokenCreateJSONRequestBody, reqEditors ...client.RequestEditorFn) (*client.TokenCreateResponse, error) {
	if body.Password == "error" {
		return nil, errors.New("auth error")
	}

	if body.Password != "password" {
		return &client.TokenCreateResponse{
			HTTPResponse: &http.Response{
				StatusCode: 401,
			},
		}, nil
	}

	return &client.TokenCreateResponse{
		JSON200: &client.VTokenRefresh{
			Access:  w("token"),
			Refresh: "token",
		},
	}, nil
}
