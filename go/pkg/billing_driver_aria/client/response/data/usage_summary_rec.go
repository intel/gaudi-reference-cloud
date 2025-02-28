// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type UsageSummaryRec struct {
	OutAcctNo              int64   `json:"out_acct_no,omitempty"`
	OutClientAcctId        string  `json:"out_client_acct_id,omitempty"`
	OutPlanInstanceNo      int64   `json:"out_plan_instance_no,omitempty"`
	OutPlanInstanceCdid    string  `json:"out_plan_instance_cdid,omitempty"`
	UsageTypeNo            int64   `json:"usage_type_no,omitempty"`
	UsageTypeLabel         string  `json:"usage_type_label,omitempty"`
	BilledInd              int64   `json:"billed_ind,omitempty"`
	TotalUnits             float32 `json:"total_units,omitempty"`
	TotalValueAmount       float32 `json:"total_value_amount,omitempty"`
	TotalValueCurrencyCode string  `json:"total_value_currency_code,omitempty"`
	LastUsageDate          string  `json:"last_usage_date,omitempty"`
	UsageTypeCd            string  `json:"usage_type_cd,omitempty"`
}
