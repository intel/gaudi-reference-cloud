// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetStatementForInvoiceMRequest struct {
	AriaRequest
	OutputFormat string `json:"output_format"`
	ClientNo     int64  `json:"client_no"`
	AuthKey      string `json:"auth_key"`
	AltCallerId  string `json:"alt_caller_id"`
	AcctNo       int64  `json:"acct_no,omitempty"`
	ClientAcctId string `json:"client_acct_id,omitempty"`
	InvoiceNo    int64  `json:"invoice_no"`
	DoEncoding   string `json:"do_encoding,omitempty"`
}
