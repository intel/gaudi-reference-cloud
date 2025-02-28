// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type InvoiceInfo struct {
	InvoiceNo               int64         `json:"invoice_no,omitempty"`
	BillingGroupNo          int64         `json:"billing_group_no,omitempty"`
	ClientBillingGroupId    string        `json:"client_billing_group_id,omitempty"`
	InvoiceChargesBeforeTax float32       `json:"invoice_charges_before_tax,omitempty"`
	InvoiceTaxCharges       float32       `json:"invoice_tax_charges,omitempty"`
	InvoiceChargesAfterTax  float32       `json:"invoice_charges_after_tax,omitempty"`
	InvoiceCreditAmount     float32       `json:"invoice_credit_amount,omitempty"`
	InvoiceTotalAmount      float32       `json:"invoice_total_amount,omitempty"`
	TotalCredit             float32       `json:"total_credit,omitempty"`
	TotalTaxCredit          float32       `json:"total_tax_credit,omitempty"`
	TotalCreditBeforeTax    float32       `json:"total_credit_before_tax,omitempty"`
	Total                   float32       `json:"total,omitempty"`
	ProrationResultAmount   float32       `json:"proration_result_amount,omitempty"`
	ExpectedMonthlyRecCost  float32       `json:"expected_monthly_rec_cost,omitempty"`
	ExpectedAnnualRecCost   float32       `json:"expected_annual_rec_cost,omitempty"`
	InvoiceItems            []InvoiceItem `json:"invoice_items,omitempty"`
}
