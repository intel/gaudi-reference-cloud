// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	"net/http"
)

type HttpClientMock struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *HttpClientMock) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{}, nil
}
