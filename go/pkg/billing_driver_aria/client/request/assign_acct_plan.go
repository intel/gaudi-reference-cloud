// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/plans"
)

type AssignAcctPlanMRequest struct {
	AriaRequest
	OutputFormat                    string                    `json:"output_format"`
	ClientNo                        int64                     `json:"client_no"`
	AuthKey                         string                    `json:"auth_key"`
	AltCallerId                     string                    `json:"alt_caller_id"`
	DoWrite                         string                    `json:"do_write,omitempty"`
	AcctNo                          int64                     `json:"acct_no,omitempty"`
	ClientAcctId                    string                    `json:"client_acct_id,omitempty"`
	ClientPlanInstanceId            string                    `json:"client_plan_instance_id,omitempty"`
	ClientAltRateScheduleId         string                    `json:"client_alt_rate_schedule_id,omitempty"`
	NewPlanNo                       int64                     `json:"new_plan_no,omitempty"`
	NewClientPlanId                 string                    `json:"new_client_plan_id,omitempty"`
	PlanUnits                       float32                   `json:"plan_units,omitempty"`
	ExistingBillingGroupNo          int64                     `json:"existing_billing_group_no,omitempty"`
	ExistingClientBillingGroupId    string                    `json:"existing_client_billing_group_id,omitempty"`
	ExistingDunningGroupNo          int64                     `json:"existing_dunning_group_no,omitempty"`
	ExistingClientDefDunningGroupId string                    `json:"existing_client_def_dunning_group_id,omitempty"`
	InvoicingOption                 int64                     `json:"invoicing_option,omitempty"`
	AssignmentDirective             int64                     `json:"assignment_directive,omitempty"`
	BillLagDays                     int64                     `json:"bill_lag_days,omitempty"`
	AltBillDay                      int64                     `json:"alt_bill_day,omitempty"`
	PlanInstanceFields              []plans.PlanInstanceField `json:"plan_instance_fields,omitempty"`
	PlanUpdateServices              []plans.PlanUpdateService `json:"plan_update_services,omitempty"`
	PlanStatus                      int64                     `json:"plan_status,omitempty"`
	StatusUntilAltStart             int64                     `json:"status_until_alt_start,omitempty"`
	RetroactiveStartDate            string                    `json:"retroactive_start_date,omitempty"`
	OverrideBillThruDate            string                    `json:"override_bill_thru_date,omitempty"`
	OverrideDatesMpInstanceNo       int64                     `json:"override_dates_mp_instance_no,omitempty"`
	OverrideDatesClientMpInstanceId string                    `json:"override_dates_client_mp_instance_id,omitempty"`
	RespLevelCd                     int64                     `json:"resp_level_cd,omitempty"`
	RespMasterPlanInstanceNo        int64                     `json:"resp_master_plan_instance_no,omitempty"`
	RespClientMasterPlanInstanceId  string                    `json:"resp_client_master_plan_instance_id,omitempty"`
}
