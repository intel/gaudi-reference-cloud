// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type GetCreditDetails struct {
	AriaResponse
	CreatedBy                string  `json:"created_by,omitempty"`
	CreatedDate              string  `json:"created_date,omitempty"`
	Amount                   float32 `json:"amount,omitempty"`
	CreditType               string  `json:"credit_type,omitempty"`
	AppliedAmount            float32 `json:"applied_amount,omitempty"`
	UnappliedAmount          float32 `json:"unapplied_amount,omitempty"`
	ReasonCode               int64   `json:"reason_code,omitempty"`
	ReasonText               string  `json:"reason_text,omitempty"`
	Comments                 string  `json:"comments,omitempty"`
	TransactionId            int64   `json:"transaction_id,omitempty"`
	VoidTransactionId        int64   `json:"void_transaction_id,omitempty"`
	CreditExpiryTypeInd      string  `json:"credit_expiry_type_ind,omitempty"`
	CreditExpiryMonths       int64   `json:"credit_expiry_months,omitempty"`
	CreditExpiryDate         string  `json:"credit_expiry_date,omitempty"`
	OutClientMpInstanceId    string  `json:"out_client_mp_instance_id,omitempty"`
	AcctLocaleName           string  `json:"acct_locale_name,omitempty"`
	OutAcctNo2               int64   `json:"out_acct_no_2,omitempty"`
	OutMasterPlanInstanceNo2 int64   `json:"out_master_plan_instance_no_2,omitempty"`
	AcctLocaleNo2            int64   `json:"acct_locale_no_2,omitempty"`
	CreditExpiryPeriod       int64   `json:"credit_expiry_period,omitempty"`
}
