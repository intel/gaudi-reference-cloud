// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetUsageHistory struct {
	AriaRequest
	OutputFormat               string `json:"output_format"`
	ClientNo                   int64  `json:"client_no,omitempty"`
	AuthKey                    string `json:"auth_key,omitempty"`
	ClientAcctId               string `json:"client_acct_id,omitempty"`
	ClientMasterPlanInstanceId string `json:"client_master_plan_instance_id,omitempty"`
	ReleaseVersion             string `json:"releaseVersion"`
	AcctNo                     int64  `json:"acct_no,omitempty"`
	MasterPlanInstanceNo       int64  `json:"master_plan_instance_no,omitempty"`
	BilledFilter               int64  `json:"billed_filter,omitempty"`
	SpecifiedUsageTypeNo       int64  `json:"specified_usage_type_no,omitempty"`
	DateRangeStart             string `json:"date_range_start,omitempty"`
	DateRangeEnd               string `json:"date_range_end,omitempty"`
	SpecifiedUsageTypeCode     string `json:"specified_usage_type_code,omitempty"`
	Limit                      int64  `json:"limit,omitempty"`
	Offset                     int64  `json:"offset,omitempty"`
	InvoiceNo                  int64  `json:"invoice_no,omitempty"`
	InvoiceLineItem            int64  `json:"invoice_line_item,omitempty"`
	RetrieveExcludedUsage      string `json:"retrieve_excluded_usage,omitempty"`
	LocaleNo                   int64  `json:"locale_no,omitempty"`
	LocaleName                 string `json:"locale_name,omitempty"`
	ClientRecordId             string `json:"client_record_id,omitempty"`
	IncludeUsageFields         string `json:"include_usage_fields,omitempty"`
	AltCallerId                string `json:"alt_caller_id,omitempty"`
}
