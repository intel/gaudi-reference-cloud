// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type TaxDetail struct {
	TaxDetailLine             int64   `json:"tax_detail_line,omitempty"`
	SeqNum                    int64   `json:"seq_num,omitempty"`
	TaxedSeqNum               int64   `json:"taxed_seq_num,omitempty"`
	Debit                     float32 `json:"debit,omitempty"`
	TaxAuthorityLevel         int64   `json:"tax_authority_level,omitempty"`
	TaxRate                   float32 `json:"tax_rate,omitempty"`
	OrigWasTaxInclusive       int64   `json:"orig_was_tax_inclusive,omitempty"`
	TaxSrvTaxTypeId           string  `json:"tax_srv_tax_type_id,omitempty"`
	TaxSrvTaxTypeDesc         string  `json:"tax_srv_tax_type_desc,omitempty"`
	TaxSrvCatText             string  `json:"tax_srv_cat_text,omitempty"`
	TaxSrvJurisNm             string  `json:"tax_srv_juris_nm,omitempty"`
	TaxSrvTaxSumText          string  `json:"tax_srv_tax_sum_text,omitempty"`
	UnroundedTaxAmt           float32 `json:"unrounded_tax_amt,omitempty"`
	CarryoverFromPrevAmt      float32 `json:"carryover_from_prev_amt,omitempty"`
	BeforeRoundAdjustedTaxAmt float32 `json:"before_round_adjusted_tax_amt,omitempty"`
	CarryoverFromCurrentAmt   float32 `json:"carryover_from_current_amt,omitempty"`
}
