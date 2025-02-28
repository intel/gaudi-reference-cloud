// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

func (*MockWekaClient) S3ListAllLifecycleRulesWithResponse(ctx context.Context, bucket string, reqEditors ...v4.RequestEditorFn) (*v4.S3ListAllLifecycleRulesResponse, error) {
	enabled := true
	expireDays := "60"
	id := "lfId"
	prefix := "pre"
	resp := v4.S3ListAllLifecycleRulesResponse{
		JSON200: &struct {
			Data *v4.S3LifecycleRule "json:\"data,omitempty\""
		}{
			Data: &v4.S3LifecycleRule{
				Bucket: &bucket,
				Rules: &[]struct {
					Enabled    *bool   "json:\"enabled,omitempty\""
					ExpiryDays *string "json:\"expiry_days,omitempty\""
					Id         *string "json:\"id,omitempty\""
					Prefix     *string "json:\"prefix,omitempty\""
				}{
					{
						Enabled:    &enabled,
						ExpiryDays: &expireDays,
						Id:         &id,
						Prefix:     &prefix,
					},
				},
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.S3ListAllLifecycleRulesWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.S3ListAllLifecycleRulesWithResponse.empty") != nil {
		return &v4.S3ListAllLifecycleRulesResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) S3CreateLifecycleRuleWithResponse(ctx context.Context, bucket string, body v4.S3CreateLifecycleRuleJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.S3CreateLifecycleRuleResponse, error) {
	id := "lfId"
	resp := v4.S3CreateLifecycleRuleResponse{
		JSON200: &struct {
			Data *struct {
				Id     *string "json:\"id,omitempty\""
				Target *string "json:\"target,omitempty\""
			} "json:\"data,omitempty\""
		}{
			Data: &struct {
				Id     *string "json:\"id,omitempty\""
				Target *string "json:\"target,omitempty\""
			}{
				Id:     &id,
				Target: &bucket,
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.S3CreateLifecycleRuleWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.S3CreateLifecycleRuleWithResponse.empty") != nil {
		return &v4.S3CreateLifecycleRuleResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) S3DeleteAllLifecycleRulesWithResponse(ctx context.Context, bucket string, reqEditors ...v4.RequestEditorFn) (*v4.S3DeleteAllLifecycleRulesResponse, error) {
	resp := v4.S3DeleteAllLifecycleRulesResponse{
		JSON200: &struct {
			Target *string "json:\"target,omitempty\""
		}{
			Target: &bucket,
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.S3DeleteLifecycleRuleWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.S3DeleteLifecycleRuleWithResponse.empty") != nil {
		return &v4.S3DeleteAllLifecycleRulesResponse{}, nil
	}

	return &resp, nil
}
