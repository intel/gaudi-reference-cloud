// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package plans

type Schedule struct {
	// TODO: Remove omitempty for ScheduleName and CurrencyCd when billing handles IntelRates
	ScheduleName               string `json:"schedule_name,omitempty" url:"schedule_name,omitempty"`
	CurrencyCd                 string `json:"currency_cd,omitempty" url:"currency_cd,omitempty"`
	ClientRateScheduleId       string `json:"client_rate_schedule_id,omitempty" url:"client_rate_schedule_id,omitempty"`
	RecurringBillingInterval   int    `json:"recurring_billing_interval,omitempty" url:"recurring_billing_interval,omitempty"`
	RecurringBillingPeriodType int    `json:"recurring_billing_period_type,omitempty" url:"recurring_billing_period_type,omitempty"`
	UsageBillingInterval       int    `json:"usage_billing_interval,omitempty" url:"usage_billing_interval,omitempty"`
	UsageBillingPeriodType     int    `json:"usage_billing_period_type,omitempty" url:"usage_billing_period_type,omitempty"`
	IsDefault                  int    `json:"is_default,omitempty" url:"is_default,omitempty"`
}
