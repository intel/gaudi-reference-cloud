// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usages

type UsageRec struct {
	AcctNo                     int64   `json:"acct_no,omitempty"`
	PlanInstanceNo             int64   `json:"plan_instance_no,omitempty"`
	UsageTypeCode              string  `json:"usage_type_code,omitempty"`
	UsageUnits                 float64 `json:"usage_units"`
	UsageDate                  string  `json:"usage_date,omitempty"`
	TelcoFrom                  string  `json:"telco_from,omitempty"`
	ClientMasterPlanInstanceId string  `json:"client_master_plan_instance_id,omitempty"`
	ClientRecordId             string  `json:"client_record_id,omitempty"`
	ClientAcctId               string  `json:"client_acct_id,omitempty"`
	Qualifier1                 string  `json:"qualifier_1,omitempty"`
}
