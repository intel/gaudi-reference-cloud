// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

func (*MockWekaClient) CreateFileSystemWithResponse(ctx context.Context, body v4.CreateFileSystemJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.CreateFileSystemResponse, error) {

	resp := v4.CreateFileSystemResponse{
		JSON200: &struct {
			Data *v4.FileSystem "json:\"data,omitempty\""
		}{
			Data: &v4.FileSystem{
				AuthRequired:   w(true),
				IsEncrypted:    w(true),
				Id:             w("id"),
				Uid:            w("fsId"),
				Name:           w("fsName"),
				TotalBudget:    w(uint64(1000000)),
				AvailableTotal: w(uint64(500000)),
				Status:         w("READY"),
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.CreateFileSystemResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.CreateFileSystemResponse.empty") != nil {
		return &v4.CreateFileSystemResponse{}, nil
	}

	if ctx.Value("testing.CreateFileSystemResponse.incomplete") != nil {
		resp.JSON200.Data.Uid = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) DeleteFileSystemWithResponse(ctx context.Context, uid string, params *v4.DeleteFileSystemParams, reqEditors ...v4.RequestEditorFn) (*v4.DeleteFileSystemResponse, error) {
	resp := v4.DeleteFileSystemResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.DeleteFileSystemWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.DeleteFileSystemWithResponse.empty") != nil {
		return &v4.DeleteFileSystemResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) GetFileSystemWithResponse(ctx context.Context, uid string, params *v4.GetFileSystemParams, reqEditors ...v4.RequestEditorFn) (*v4.GetFileSystemResponse, error) {
	resp := v4.GetFileSystemResponse{
		JSON200: &struct {
			Data *v4.FileSystem "json:\"data,omitempty\""
		}{
			Data: &v4.FileSystem{
				AuthRequired:   w(true),
				IsEncrypted:    w(true),
				Id:             w("id"),
				Uid:            w("fsId"),
				Name:           w("fsName"),
				TotalBudget:    w(uint64(1000000)),
				AvailableTotal: w(uint64(500000)),
				Status:         w("READY"),
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetFileSystemWithResponse.404") != nil {
		return &v4.GetFileSystemResponse{
			JSON404: &v4.N404{},
			HTTPResponse: &http.Response{
				StatusCode: 404,
			},
		}, nil
	}

	if ctx.Value("testing.GetFileSystemWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetFileSystemWithResponse.empty") != nil {
		return &v4.GetFileSystemResponse{}, nil
	}

	if ctx.Value("testing.GetFileSystemWithResponse.incomplete") != nil {
		resp.JSON200.Data.Uid = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) GetFileSystemsWithResponse(ctx context.Context, params *v4.GetFileSystemsParams, reqEditors ...v4.RequestEditorFn) (*v4.GetFileSystemsResponse, error) {
	resp := v4.GetFileSystemsResponse{
		JSON200: &struct {
			Data *[]v4.FileSystem "json:\"data,omitempty\""
		}{
			Data: &[]v4.FileSystem{
				{
					AuthRequired:   w(true),
					IsEncrypted:    w(true),
					Id:             w("id"),
					Uid:            w("fsId"),
					Name:           w("fsName"),
					TotalBudget:    w(uint64(1000000)),
					AvailableTotal: w(uint64(500000)),
					Status:         w("READY"),
				},
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetFileSystemsWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetFileSystemsWithResponse.empty") != nil {
		return &v4.GetFileSystemsResponse{}, nil
	}

	if ctx.Value("testing.GetFileSystemsWithResponse.incomplete") != nil {
		fs := *resp.JSON200.Data
		fs[0].Uid = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) UpdateFileSystemWithResponse(ctx context.Context, uid string, body v4.UpdateFileSystemJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.UpdateFileSystemResponse, error) {
	resp := v4.UpdateFileSystemResponse{
		JSON200: &struct {
			Data *v4.FileSystem "json:\"data,omitempty\""
		}{
			Data: &v4.FileSystem{
				AuthRequired:   w(true),
				IsEncrypted:    w(true),
				Id:             w("id"),
				Uid:            w("fsId"),
				Name:           w("fsName"),
				TotalBudget:    w(uint64(1000000)),
				AvailableTotal: w(uint64(500000)),
				Status:         w("READY"),
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.UpdateFileSystemWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.UpdateFileSystemWithResponse.empty") != nil {
		return &v4.UpdateFileSystemResponse{}, nil
	}

	if ctx.Value("testing.UpdateFileSystemWithResponse.incomplete") != nil {
		resp.JSON200.Data.Uid = nil
		return &resp, nil
	}

	return &resp, nil
}
