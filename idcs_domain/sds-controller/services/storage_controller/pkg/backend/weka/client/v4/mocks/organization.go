// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

func (*MockWekaClient) GetOrganizationsWithResponse(ctx context.Context, reqEditors ...v4.RequestEditorFn) (*v4.GetOrganizationsResponse, error) {
	resp := v4.GetOrganizationsResponse{
		JSON200: &struct {
			Data *[]v4.Organization "json:\"data,omitempty\""
		}{
			Data: &[]v4.Organization{
				{
					Id:             w(0),
					Name:           w("ns"),
					SsdAllocated:   w(uint64(0)),
					SsdQuota:       w(uint64(0)),
					TotalAllocated: w(uint64(500000)),
					TotalQuota:     w(uint64(100000000)),
					Uid:            w("nsId"),
				},
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetOrganizationsWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetOrganizationsWithResponse.empty") != nil {
		return &v4.GetOrganizationsResponse{}, nil
	}

	if ctx.Value("testing.GetOrganizationsWithResponse.incomplete") != nil {
		orgs := *resp.JSON200.Data
		orgs[0].Uid = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) GetOrganizationWithResponse(ctx context.Context, uid string, reqEditors ...v4.RequestEditorFn) (*v4.GetOrganizationResponse, error) {
	resp := v4.GetOrganizationResponse{
		JSON200: &struct {
			Data *v4.Organization "json:\"data,omitempty\""
		}{
			Data: &v4.Organization{
				Id:             w(0),
				Name:           w("ns"),
				SsdAllocated:   w(uint64(0)),
				SsdQuota:       w(uint64(0)),
				TotalAllocated: w(uint64(500000)),
				TotalQuota:     w(uint64(100000000)),
				Uid:            &uid,
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.GetOrganizationWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetOrganizationWithResponse.empty") != nil {
		return &v4.GetOrganizationResponse{}, nil
	}

	if ctx.Value("testing.GetOrganizationWithResponse.incomplete") != nil {
		resp.JSON200.Data.Uid = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) CreateOrganizationWithResponse(ctx context.Context, body v4.CreateOrganizationJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.CreateOrganizationResponse, error) {
	resp := v4.CreateOrganizationResponse{
		JSON200: &struct {
			Data *v4.Organization "json:\"data,omitempty\""
		}{
			Data: &v4.Organization{
				Id:             w(0),
				Name:           w("ns"),
				SsdAllocated:   w(uint64(0)),
				SsdQuota:       w(uint64(0)),
				TotalAllocated: w(uint64(500000)),
				TotalQuota:     w(uint64(100000000)),
				Uid:            w("nsId"),
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.CreateOrganizationWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.CreateOrganizationWithResponse.empty") != nil {
		return &v4.CreateOrganizationResponse{}, nil
	}

	if ctx.Value("testing.CreateOrganizationWithResponse.incomplete") != nil {
		resp.JSON200.Data.Uid = nil
		return &resp, nil
	}

	return &resp, nil
}

func (*MockWekaClient) DeleteOrganizationWithResponse(ctx context.Context, uid string, reqEditors ...v4.RequestEditorFn) (*v4.DeleteOrganizationResponse, error) {
	resp := v4.DeleteOrganizationResponse{
		JSON200: &v4.N200{},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.DeleteOrganizationWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.DeleteOrganizationWithResponse.empty") != nil {
		return &v4.DeleteOrganizationResponse{}, nil
	}

	return &resp, nil
}

func (*MockWekaClient) SetOrganizationLimitWithResponse(ctx context.Context, uid string, body v4.SetOrganizationLimitJSONRequestBody, reqEditors ...v4.RequestEditorFn) (*v4.SetOrganizationLimitResponse, error) {
	resp := v4.SetOrganizationLimitResponse{
		JSON200: &struct {
			Data *v4.Organization "json:\"data,omitempty\""
		}{
			Data: &v4.Organization{
				Id:             w(0),
				Name:           w("ns"),
				SsdAllocated:   w(uint64(0)),
				SsdQuota:       w(uint64(0)),
				TotalAllocated: w(uint64(500000)),
				TotalQuota:     w(uint64(5000000)),
				Uid:            w("nsId"),
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}

	if ctx.Value("testing.SetOrganizationLimitResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.SetOrganizationLimitResponse.empty") != nil {
		return &v4.SetOrganizationLimitResponse{}, nil
	}

	if ctx.Value("testing.SetOrganizationLimitResponse.incomplete") != nil {
		resp.JSON200.Data.Uid = nil
		return &resp, nil
	}

	return &resp, nil
}
