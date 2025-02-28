// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

func (*MockWekaClient) GetS3PolicyWithResponse(ctx context.Context, policy string, reqEditors ...v4.RequestEditorFn) (*v4.GetS3PolicyResponse, error) {
	version := "2012-10-17"
	allow := "Allow"
	sidRead := "read"
	sidWrite := "write"
	sidDelete := "delete"

	resp := v4.GetS3PolicyResponse{
		JSON200: &struct {
			Data *v4.S3Policy "json:\"data,omitempty\""
		}{
			&v4.S3Policy{
				Policy: &struct {
					Content *v4.S3IAMPolicy "json:\"content,omitempty\""
					Name    *string         "json:\"name,omitempty\""
				}{
					Content: &v4.S3IAMPolicy{
						Version: &version,
						Statement: &[]v4.S3IAMStatement{
							{
								Effect: &allow,
								Sid:    &sidRead,
								Resource: &[]string{
									"arn:aws:s3:::bucket/pre*",
								},
								Action: &[]string{
									"s3:GetObject",
								},
							},
							{
								Effect: &allow,
								Sid:    &sidWrite,
								Resource: &[]string{
									"arn:aws:s3:::bucket/pre*",
								},
								Action: &[]string{
									"s3:WriteObject",
								},
							},
							{
								Effect: &allow,
								Sid:    &sidDelete,
								Resource: &[]string{
									"arn:aws:s3:::bucket/pre*",
								},
								Action: &[]string{
									"s3:DeleteObject",
								},
							},
						},
					},
				},
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetS3PolicyWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetS3PolicyWithResponse.empty") != nil {
		return &v4.GetS3PolicyResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) DeleteS3PolicyWithResponse(ctx context.Context, policy string, reqEditors ...v4.RequestEditorFn) (*v4.DeleteS3PolicyResponse, error) {
	resp := v4.DeleteS3PolicyResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.DeleteS3PolicyWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.DeleteS3PolicyWithResponse.empty") != nil {
		return &v4.DeleteS3PolicyResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) AttachS3PolicyWithResponse(ctx context.Context, body v4.AttachS3PolicyJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.AttachS3PolicyResponse, error) {
	resp := v4.AttachS3PolicyResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.AttachS3PolicyWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.AttachS3PolicyWithResponse.empty") != nil {
		return &v4.AttachS3PolicyResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) CreateS3PolicyWithResponse(ctx context.Context, body v4.CreateS3PolicyJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.CreateS3PolicyResponse, error) {
	resp := v4.CreateS3PolicyResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.CreateS3PolicyWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.CreateS3PolicyWithResponse.empty") != nil {
		return &v4.CreateS3PolicyResponse{}, nil
	}

	return &resp, nil
}
