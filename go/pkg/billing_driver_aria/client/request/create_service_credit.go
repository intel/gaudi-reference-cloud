// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type CreateAdvancedServiceCreditMRequest struct {
	AriaRequest
	OutputFormat        string  `json:"output_format"`
	ClientNo            int64   `json:"client_no"`
	AuthKey             string  `json:"auth_key"`
	AltCallerId         string  `json:"alt_caller_id"`
	ClientAcctId        string  `json:"client_acct_id,omitempty"`
	Amount              float64 `json:"amount,omitempty"`
	ReasonCode          int64   `json:"reason_code,omitempty"`
	Comments            string  `json:"comments,omitempty"`
	ServiceCodeOption   int64   `json:"service_code_option,omitempty"`
	CreditExpiryDate    string  `json:"credit_expiry_date,omitempty"`
	CreditExpiryTypeInd string  `json:"credit_expiry_type_ind,omitempty"`
}
