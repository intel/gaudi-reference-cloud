// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetPromotionsMRequest struct {
	AriaRequest
	OutputFormat string `json:"output_format"`
	ClientNo     int64  `json:"client_no"`
	AuthKey      string `json:"auth_key"`
}
