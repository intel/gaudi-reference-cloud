// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type Service struct {
	ErrorCode       int64  `json:"error_code"`
	ErrorMsg        string `json:"error_msg"`
	ServiceNo       int    `json:"service_no"`
	ClientServiceId string `json:"client_service_id"`
}
