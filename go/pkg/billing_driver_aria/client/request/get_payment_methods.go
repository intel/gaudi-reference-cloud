// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetPaymentMethodsRequest struct {
	AriaRequest
	OutputFormat     string `json:"output_format"`
	ClientNo         int64  `json:"client_no"`
	AuthKey          string `json:"auth_key"`
	AltCallerId      string `json:"alt_caller_id"`
	ClientAccountId  string `json:"client_acct_id"`
	PaymentsReturned int64  `json:"payments_returned"`
	FilterStatus     int64  `json:"filter_status"`
}
