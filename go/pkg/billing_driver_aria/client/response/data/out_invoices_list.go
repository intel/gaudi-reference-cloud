// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type OutInvoicesList struct {
	InvoicingErrorCode      int64             `json:"invoicing_error_code,omitempty"`
	InvoicingErrorMsg       string            `json:"invoicing_error_msg,omitempty"`
	InvoiceNo               int64             `json:"invoice_no,omitempty"`
	BillingGroupNo          int64             `json:"billing_group_no,omitempty"`
	ClientBillingGroupId    string            `json:"client_billing_group_id,omitempty"`
	InvoiceChargesBeforeTax float32           `json:"invoice_charges_before_tax,omitempty"`
	InvoiceTaxCharges       float32           `json:"invoice_tax_charges,omitempty"`
	InvoiceChargesAfterTax  float32           `json:"invoice_charges_after_tax,omitempty"`
	InvoiceCreditAmount     float32           `json:"invoice_credit_amount,omitempty"`
	InvoiceTotalAmount      float32           `json:"invoice_total_amount,omitempty"`
	InvoiceItems            []InvoiceItem     `json:"invoice_items,omitempty"`
	TaxDetails              []TaxDetail       `json:"tax_details,omitempty"`
	ThirdPartyErrors        []ThirdPartyError `json:"third_party_errors,omitempty"`
}
