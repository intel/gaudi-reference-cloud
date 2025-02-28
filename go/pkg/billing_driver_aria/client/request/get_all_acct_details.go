// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetAcctDetailsAllMRequest struct {
	AriaRequest
	OutputFormat          string `json:"output_format"`
	ClientNo              int64  `json:"client_no"`
	ClientAcctId          string `json:"client_acct_id,omitempty"`
	AuthKey               string `json:"auth_key"`
	AltCallerId           string `json:"alt_caller_id"`
	ClientReceiptId       string `json:"client_receipt_id,omitempty"`
	AcctNo                int64  `json:"acct_no,omitempty"`
	IncludeMasterPlans    int64  `json:"include_master_plans,omitempty"`
	IncludeSuppPlans      int64  `json:"include_supp_plans,omitempty"`
	IncludeBillingGroups  int64  `json:"include_billing_groups,omitempty"`
	IncludePaymentMethods int64  `json:"include_payment_methods,omitempty"`
}
