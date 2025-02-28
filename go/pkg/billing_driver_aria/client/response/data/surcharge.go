// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type Surcharge struct {
	SurchargeNo       int64           `json:"surcharge_no,omitempty"`
	SurchargeName     string          `json:"surcharge_name,omitempty"`
	ClientSurchargeId string          `json:"client_surcharge_id,omitempty"`
	Description       string          `json:"description,omitempty"`
	ExtDescription    string          `json:"ext_description,omitempty"`
	SurchargeType     string          `json:"surcharge_type,omitempty"`
	Currency          string          `json:"currency,omitempty"`
	TaxGroup          string          `json:"tax_group,omitempty"`
	InvoiceAppMethod  string          `json:"invoice_app_method,omitempty"`
	RevGlCode         string          `json:"rev_gl_code,omitempty"`
	ArGlCode          string          `json:"ar_gl_code,omitempty"`
	SurchargePlan     []SurchargePlan `json:"surcharge_plan,omitempty"`
	SurchargeRate     []SurchargeRate `json:"surcharge_rate,omitempty"`
	TaxInclusiveInd   int64           `json:"tax_inclusive_ind,omitempty"`
}
