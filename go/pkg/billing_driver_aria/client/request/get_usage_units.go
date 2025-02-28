// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetUsageUnits struct {
	RestCall     string `json:"rest_call"`
	OutputFormat string `json:"output_format"`
	ClientNo     int64  `json:"client_no"`
	AuthKey      string `json:"auth_key"`
	AltCallerId  string `json:"alt_caller_id"`
}
