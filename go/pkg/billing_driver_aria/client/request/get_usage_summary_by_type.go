// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetUsageSummaryByType struct {
	AriaRequest
	OutputFormat               string `json:"output_format"`
	ClientNo                   int64  `json:"client_no,omitempty"`
	AuthKey                    string `json:"auth_key,omitempty"`
	ClientAcctId               string `json:"client_acct_id,omitempty"`
	ClientMasterPlanInstanceId string `json:"client_master_plan_instance_id,omitempty"`
	ReleaseVersion             string `json:"releaseVersion"`
	AcctNo                     int64  `json:"acct_no,omitempty"`
	UserId                     string `json:"user_id,omitempty"`
	MasterPlanInstanceNo       int64  `json:"master_plan_instance_no,omitempty"`
	UsageTypeFilter            int64  `json:"usage_type_filter,omitempty"`
	DateFilterStartDate        string `json:"date_filter_start_date,omitempty"`
	DateFilterStartTime        string `json:"date_filter_start_time,omitempty"`
	DateFilterEndDate          string `json:"date_filter_end_date,omitempty"`
	DateFilterEndTime          string `json:"date_filter_end_time,omitempty"`
	BilledFilter               int64  `json:"billed_filter,omitempty"`
	BillingPeriodFlag          int64  `json:"billing_period_flag,omitempty"`
}
