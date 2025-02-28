// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type ApplyServiceCredit struct {
	AriaRequest
	OutputFormat     string  `json:"output_format"`
	ClientNo         int64   `json:"client_no"`
	AuthKey          string  `json:"auth_key"`
	AcctNo           int64   `json:"acct_no,omitempty"`
	ClientAcctId     string  `json:"client_acct_id,omitempty"`
	CreditReasonCode int64   `json:"credit_reason_code,omitempty"`
	CreditAmount     float64 `json:"credit_amount,omitempty"`
}
