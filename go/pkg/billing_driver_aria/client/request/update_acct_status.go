// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type UpdateAcctStatusMRequest struct {
	AriaRequest
	OutputFormat string `json:"output_format"`
	ClientNo     int64  `json:"client_no"`
	AuthKey      string `json:"auth_key"`
	AltCallerId  string `json:"alt_caller_id"`
	AccountNo    int64  `json:"account_no,omitempty"`
	StatusCd     int64  `json:"status_cd"`
}
