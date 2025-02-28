// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type AuthorizeElectronicPaymentRequest struct {
	AriaRequest
	OutputFormat    string  `json:"output_format"`
	ClientNo        int64   `json:"client_no"`
	AuthKey         string  `json:"auth_key"`
	AltCallerId     string  `json:"alt_caller_id"`
	ClientAccountId string  `json:"client_acct_id"`
	BillingGroupNo  int64   `json:"billing_group_no"`
	Amount          float64 `json:"amount"`
}
