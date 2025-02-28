// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type Resource struct {
	Resources           int `json:"resources,omitempty"`
	ResourceTypeNo      int `json:"resource_type_no,omitempty"`
	ResourceUnits       int `json:"resource_units,omitempty"`
	ExpireOnPaidThrough int `json:"expire_on_paid_through,omitempty"`
	ResetOnUpdate       int `json:"reset_on_update,omitempty"`
	AddDaysToExpiry     int `json:"add_days_to_expiry,omitempty"`
}
