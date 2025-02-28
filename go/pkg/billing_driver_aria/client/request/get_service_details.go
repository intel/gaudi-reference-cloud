// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetServiceDetails struct {
	AriaRequest
	OutputFormat    string `json:"output_format" url:"output_format"`
	ClientNo        int64  `json:"client_no" url:"client_no"`
	AuthKey         string `json:"auth_key" url:"auth_key"`
	AltCallerId     string `json:"alt_caller_id" url:"alt_caller_id"`
	ServiceNo       int    `json:"service_no" url:"service_no"`
	ClientServiceId string `json:"client_service_id" url:"client_service_id"`
}
