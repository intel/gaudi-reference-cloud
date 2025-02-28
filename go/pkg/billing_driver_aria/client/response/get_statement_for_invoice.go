// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type GetStatementForInvoiceMResponse struct {
	AriaResponse
	OutStatement string `json:"out_statement,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
}
