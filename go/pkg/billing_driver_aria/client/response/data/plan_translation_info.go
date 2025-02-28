// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PlanTranslationInfo struct {
	PlanNo      string        `json:"plan_no,omitempty"`
	LocaleName  string        `json:"locale_name,omitempty"`
	PlanName    string        `json:"plan_name,omitempty"`
	PlanDesc    string        `json:"plan_desc,omitempty"`
	LocaleNo    int           `json:"locale_no,omitempty"`
	RateSchedsT []RateSchedsT `json:"rate_scheds_t,omitempty"`
}
