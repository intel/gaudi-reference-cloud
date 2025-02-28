// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type RecordUsage struct {
	AriaRequest
	RestCall                   string  `json:"rest_call"`
	OutputFormat               string  `json:"output_format"`
	ClientNo                   int64   `json:"client_no"`
	AuthKey                    string  `json:"auth_key"`
	AltCallerId                string  `json:"alt_caller_id"`
	AcctNo                     int64   `json:"acct_no,omitempty"`
	Userid                     string  `json:"userid,omitempty"`
	MasterPlanInstanceNo       int64   `json:"master_plan_instance_no,omitempty"`
	ClientMasterPlanInstanceId string  `json:"client_master_plan_instance_id,omitempty"`
	PlanInstanceNo             int64   `json:"plan_instance_no,omitempty"`
	ClientPlanInstanceId       string  `json:"client_plan_instance_id,omitempty"`
	UsageType                  int64   `json:"usage_type,omitempty"`
	UsageUnits                 float64 `json:"usage_units"`
	UsageDate                  string  `json:"usage_date,omitempty"`
	BillableUnits              float32 `json:"billable_units,omitempty"`
	Amt                        float32 `json:"amt,omitempty"`
	Rate                       float32 `json:"rate,omitempty"`
	TelcoFrom                  string  `json:"telco_from,omitempty"`
	TelcoTo                    string  `json:"telco_to,omitempty"`
	Comments                   string  `json:"comments,omitempty"`
	ExcludeFromBilling         string  `json:"exclude_from_billing,omitempty"`
	ExclusionComments          string  `json:"exclusion_comments,omitempty"`
	Qualifier1                 string  `json:"qualifier_1,omitempty"`
	Qualifier2                 string  `json:"qualifier_2,omitempty"`
	Qualifier3                 string  `json:"qualifier_3,omitempty"`
	Qualifier4                 string  `json:"qualifier_4,omitempty"`
	ParentUsageRecNo           int64   `json:"parent_usage_rec_no,omitempty"`
	UsageTypeCode              string  `json:"usage_type_code,omitempty"`
	ClientRecordId             string  `json:"client_record_id,omitempty"`
	CallerId                   string  `json:"caller_id,omitempty"`
	ClientReceiptId            string  `json:"client_receipt_id,omitempty"`
	ClientAcctId               string  `json:"client_acct_id,omitempty"`
}
