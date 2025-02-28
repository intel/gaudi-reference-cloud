// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type AllCredit struct {
	OutAcctNo               int64   `json:"out_acct_no,omitempty"`
	OutMasterPlanInstanceNo int64   `json:"out_master_plan_instance_no,omitempty"`
	OutClientMpInstanceId   string  `json:"out_client_mp_instance_id,omitempty"`
	CreditNo                int64   `json:"credit_no,omitempty"`
	CreatedBy               string  `json:"created_by,omitempty"`
	CreatedDate             string  `json:"created_date,omitempty"`
	Amount                  float32 `json:"amount,omitempty"`
	CreditType              string  `json:"credit_type,omitempty"`
	AppliedAmount           float32 `json:"applied_amount,omitempty"`
	UnappliedAmount         float32 `json:"unapplied_amount,omitempty"`
	ReasonCode              int64   `json:"reason_code,omitempty"`
	ReasonText              string  `json:"reason_text,omitempty"`
	TransactionId           int64   `json:"transaction_id,omitempty"`
	VoidTransactionId       int64   `json:"void_transaction_id,omitempty"`
}
