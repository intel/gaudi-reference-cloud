// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type RateSchedule struct {
	ScheduleNo                 int    `json:"schedule_no,omitempty"`
	ScheduleName               string `json:"schedule_name,omitempty"`
	CurrencyCd                 string `json:"currency_cd,omitempty"`
	ClientRateScheduleId       string `json:"client_rate_schedule_id,omitempty"`
	RecurringBillingInterval   int    `json:"recurring_billing_interval,omitempty"`
	RecurringBillingPeriodType int    `json:"recurring_billing_period_type,omitempty"`
	UsageBillingInterval       int    `json:"usage_billing_interval,omitempty"`
	UsageBillingPeriodType     int    `json:"usage_billing_period_type,omitempty"`
	DefaultInd                 int    `json:"default_ind,omitempty"`
	DefaultIndCurr             int    `json:"default_ind_curr,omitempty"`
}
