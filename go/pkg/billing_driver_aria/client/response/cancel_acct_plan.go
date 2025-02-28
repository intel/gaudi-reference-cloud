// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type CancelAcctPlanMResponse struct {
	ErrorCode                   int64   `json:"error_code,omitempty"`
	ErrorMsg                    string  `json:"error_msg,omitempty"`
	ProrationResultAmount       float32 `json:"proration_result_amount,omitempty"`
	TotalChargesBeforeTax       float32 `json:"total_charges_before_tax,omitempty"`
	TotalTaxCharges             float32 `json:"total_tax_charges,omitempty"`
	TotalChargesAfterTax        float32 `json:"total_charges_after_tax,omitempty"`
	TotalCredit                 float32 `json:"total_credit,omitempty"`
	TotalTaxCredit              float32 `json:"total_tax_credit,omitempty"`
	TotalCreditBeforeTax        float32 `json:"total_credit_before_tax,omitempty"`
	Total                       float32 `json:"total,omitempty"`
	ProrationCreditResultAmount float32 `json:"proration_credit_result_amount,omitempty"`
	ProrationCreditAmount       float32 `json:"proration_credit_amount,omitempty"`
	ProrationTaxAmount          float32 `json:"proration_tax_amount,omitempty"`
}
