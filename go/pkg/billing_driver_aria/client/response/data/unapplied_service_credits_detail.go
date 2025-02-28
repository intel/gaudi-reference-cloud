// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type UnappliedServiceCreditsDetail struct {
	CreateDate             string  `json:"create_date,omitempty"`
	CreateUser             string  `json:"create_user,omitempty"`
	InitialAmount          float32 `json:"initial_amount,omitempty"`
	AmountLeftToApply      float32 `json:"amount_left_to_apply,omitempty"`
	ReasonCd               int64   `json:"reason_cd,omitempty"`
	ReasonText             string  `json:"reason_text,omitempty"`
	Comments               string  `json:"comments,omitempty"`
	CurrencyCd             string  `json:"currency_cd,omitempty"`
	ServiceNoToApply       int64   `json:"service_no_to_apply,omitempty"`
	ServiceNameToApply     string  `json:"service_name_to_apply,omitempty"`
	ClientServiceIdToApply string  `json:"client_service_id_to_apply,omitempty"`
	OutAcctNo              int64   `json:"out_acct_no,omitempty"`
	CreditId2              int64   `json:"credit_id_2,omitempty"`
}
