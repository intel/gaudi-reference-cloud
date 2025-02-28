// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type UsageHistoryRec struct {
	BillableAcctNo             int64        `json:"billable_acct_no,omitempty"`
	IncurringAcctNo            int64        `json:"incurring_acct_no,omitempty"`
	ClientBillableAcctId       string       `json:"client_billable_acct_id,omitempty"`
	ClientIncurringAcctId      string       `json:"client_incurring_acct_id,omitempty"`
	PlanInstanceId             int64        `json:"plan_instance_id,omitempty"`
	ClientPlanInstanceId       string       `json:"client_plan_instance_id,omitempty"`
	UsageTypeNo                int64        `json:"usage_type_no,omitempty"`
	UsageTypeDescription       string       `json:"usage_type_description,omitempty"`
	UsageDate                  string       `json:"usage_date,omitempty"`
	UsageTime                  string       `json:"usage_time,omitempty"`
	Units                      float32      `json:"units,omitempty"`
	UnitsDescription           string       `json:"units_description,omitempty"`
	UsageUnitsDescription      string       `json:"usage_units_description,omitempty"`
	InvoiceNo                  int64        `json:"invoice_no,omitempty"`
	TelcoTo                    string       `json:"telco_to,omitempty"`
	TelcoFrom                  string       `json:"telco_from,omitempty"`
	SpecificRecordChargeAmount float32      `json:"specific_record_charge_amount,omitempty"`
	IsExcluded                 string       `json:"is_excluded,omitempty"`
	ExclusionComments          string       `json:"exclusion_comments,omitempty"`
	Comments                   string       `json:"comments,omitempty"`
	PreRatedRate               float32      `json:"pre_rated_rate,omitempty"`
	Qualifier1                 string       `json:"qualifier_1,omitempty"`
	Qualifier2                 string       `json:"qualifier_2,omitempty"`
	Qualifier3                 string       `json:"qualifier_3,omitempty"`
	Qualifier4                 string       `json:"qualifier_4,omitempty"`
	RecordedUnits              float32      `json:"recorded_units,omitempty"`
	UsageRecNo                 int64        `json:"usage_rec_no,omitempty"`
	UsageParentRecNo           int64        `json:"usage_parent_rec_no,omitempty"`
	UsageTypeCode              string       `json:"usage_type_code,omitempty"`
	ClientRecordId             string       `json:"client_record_id,omitempty"`
	ExcludeReasonCd            int64        `json:"exclude_reason_cd,omitempty"`
	MasterPlanInstanceNo       int64        `json:"master_plan_instance_no,omitempty"`
	ClientMasterPlanInstanceId string       `json:"client_master_plan_instance_id,omitempty"`
	UsageField                 []UsageField `json:"usage_field,omitempty"`
}
