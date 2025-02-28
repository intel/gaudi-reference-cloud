// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	"context"
	"errors"
	"net/http"

	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

func (*MockWekaClient) GetClusterStatusWithResponse(ctx context.Context, reqEditors ...v4.RequestEditorFn) (*v4.GetClusterStatusResponse, error) {
	resp := v4.GetClusterStatusResponse{
		JSON200: &struct {
			Data *struct {
				Capacity *struct {
					HotSpareBytes      *uint64 "json:\"hot_spare_bytes,omitempty\""
					TotalBytes         *uint64 "json:\"total_bytes,omitempty\""
					UnprovisionedBytes *uint64 "json:\"unprovisioned_bytes,omitempty\""
				} "json:\"capacity,omitempty\""
				Guid   *string "json:\"guid,omitempty\""
				Name   *string "json:\"name,omitempty\""
				Status *string "json:\"status,omitempty\""
			} "json:\"data,omitempty\""
		}{
			Data: &struct {
				Capacity *struct {
					HotSpareBytes      *uint64 "json:\"hot_spare_bytes,omitempty\""
					TotalBytes         *uint64 "json:\"total_bytes,omitempty\""
					UnprovisionedBytes *uint64 "json:\"unprovisioned_bytes,omitempty\""
				} "json:\"capacity,omitempty\""
				Guid   *string "json:\"guid,omitempty\""
				Name   *string "json:\"name,omitempty\""
				Status *string "json:\"status,omitempty\""
			}{
				Status: w("OK"),
				Capacity: &struct {
					HotSpareBytes      *uint64 "json:\"hot_spare_bytes,omitempty\""
					TotalBytes         *uint64 "json:\"total_bytes,omitempty\""
					UnprovisionedBytes *uint64 "json:\"unprovisioned_bytes,omitempty\""
				}{
					TotalBytes:         w(uint64(50000000)),
					UnprovisionedBytes: w(uint64(10000000)),
				},
				Guid: w("efed877e-0fed-4a42-a3b4-864040f19686"),
				Name: w("test"),
			},
		},
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}
	if ctx.Value("testing.GetClusterStatusWithResponse.error") != nil {
		return nil, errors.New("Something wrong")
	}

	if ctx.Value("testing.GetClusterStatusWithResponse.empty") != nil {
		return &v4.GetClusterStatusResponse{}, nil
	}

	if ctx.Value("testing.GetClusterStatusWithResponse.status") != nil {
		resp.JSON200.Data.Status = nil
		return &resp, nil
	}

	return &resp, nil
}
