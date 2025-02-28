// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type AssignAcctPlanMResponse struct {
	AriaResponse
	ProrationResultAmount       float32 `json:"proration_result_amount,omitempty"`
	TotalChargesBeforeTax       float32 `json:"total_charges_before_tax,omitempty"`
	TotalTaxCharges             float32 `json:"total_tax_charges,omitempty"`
	TotalChargesAfterTax        float32 `json:"total_charges_after_tax,omitempty"`
	TotalCredit                 float32 `json:"total_credit,omitempty"`
	TotalTaxCredit              float32 `json:"total_tax_credit,omitempty"`
	TotalCreditBeforeTax        float32 `json:"total_credit_before_tax,omitempty"`
	Total                       float32 `json:"total,omitempty"`
	ProrationTaxAmount          float32 `json:"proration_tax_amount,omitempty"`
	ProrationCreditResultAmount float32 `json:"proration_credit_result_amount,omitempty"`
	ProrationCreditAmount       float32 `json:"proration_credit_amount,omitempty"`
	PlanInstanceNo              int64   `json:"plan_instance_no,omitempty"`
}
