// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type BillingError struct {
	InvoicingErrorCode   int64  `json:"invoicing_error_code,omitempty"`
	InvoicingErrorMsg    string `json:"invoicing_error_msg,omitempty"`
	CollectionErrorCode  int64  `json:"collection_error_code,omitempty"`
	CollectionErrorMsg   string `json:"collection_error_msg,omitempty"`
	StatementErrorCode   int64  `json:"statement_error_code,omitempty"`
	StatementErrorMsg    string `json:"statement_error_msg,omitempty"`
	BillingGroupNo       int64  `json:"billing_group_no,omitempty"`
	ClientBillingGroupId string `json:"client_billing_group_id,omitempty"`
}
