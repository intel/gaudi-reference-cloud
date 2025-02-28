// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetAccountNoFromUserIdMRequest struct {
	AriaRequest
	OutputFormat string `json:"output_format"`
	ClientNo     int64  `json:"client_no"`
	AuthKey      string `json:"auth_key"`
	AltCallerId  string `json:"alt_caller_id"`
	UserId       string `json:"user_id"`
}
