// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

func (*MockWekaClient) CreateS3BucketWithResponse(ctx context.Context, body v4.CreateS3BucketJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.CreateS3BucketResponse, error) {

	var data interface{}
	resp := v4.CreateS3BucketResponse{
		JSON200: &v4.N200{
			Data: &data,
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.CreateS3BucketWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.CreateS3BucketWithResponse.empty") != nil {
		return &v4.CreateS3BucketResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) DestroyS3BucketWithResponse(ctx context.Context, bucket string, params *v4.DestroyS3BucketParams, reqEditors ...v4.RequestEditorFn) (*v4.DestroyS3BucketResponse, error) {
	resp := v4.DestroyS3BucketResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.DestroyS3BucketWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.DestroyS3BucketWithResponse.empty") != nil {
		return &v4.DestroyS3BucketResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) GetS3BucketPolicyWithResponse(ctx context.Context, bucket string, reqEditors ...v4.RequestEditorFn) (*v4.GetS3BucketPolicyResponse, error) {
	public := "public"
	resp := v4.GetS3BucketPolicyResponse{
		JSON200: &struct {
			Data *struct {
				Policy *string "json:\"policy,omitempty\""
			} "json:\"data,omitempty\""
		}{
			Data: &struct {
				Policy *string "json:\"policy,omitempty\""
			}{
				Policy: &public,
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetS3BucketPolicyWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetS3BucketPolicyWithResponse.empty") != nil {
		return &v4.GetS3BucketPolicyResponse{}, nil
	}

	if ctx.Value("testing.GetS3BucketPolicyWithResponse.incomplete") != nil {
		resp.JSON200.Data.Policy = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) GetS3BucketsWithResponse(ctx context.Context, reqEditors ...v4.RequestEditorFn) (*v4.GetS3BucketsResponse, error) {
	var limit uint64 = 5000000
	var used uint64 = 3000000
	name := "bucketId"
	path := "path"
	resp := v4.GetS3BucketsResponse{
		JSON200: &struct {
			Data *v4.S3Bucket "json:\"data,omitempty\""
		}{
			Data: &v4.S3Bucket{
				Buckets: &[]struct {
					HardLimitBytes *uint64 "json:\"hard_limit_bytes,omitempty\""
					Name           *string "json:\"name,omitempty\""
					Path           *string "json:\"path,omitempty\""
					UsedBytes      *uint64 "json:\"used_bytes,omitempty\""
				}{
					{
						HardLimitBytes: &limit,
						Name:           &name,
						Path:           &path,
						UsedBytes:      &used,
					},
				},
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetS3BucketsWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetS3BucketsWithResponse.empty") != nil {
		return &v4.GetS3BucketsResponse{}, nil
	}

	if ctx.Value("testing.GetS3BucketsWithResponse.noName") != nil {
		buckets := *resp.JSON200.Data.Buckets
		buckets[0].Name = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) SetS3BucketPolicyWithResponse(ctx context.Context, bucket string, body v4.SetS3BucketPolicyJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.SetS3BucketPolicyResponse, error) {
	resp := v4.SetS3BucketPolicyResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.SetS3BucketPolicyWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.SetS3BucketPolicyWithResponse.empty") != nil {
		return &v4.SetS3BucketPolicyResponse{}, nil
	}

	return &resp, nil
}
