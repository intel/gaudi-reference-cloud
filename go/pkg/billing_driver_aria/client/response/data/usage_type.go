// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type UsageType struct {
	UsageTypeNo   int    `json:"usage_type_no"`
	UsageTypeDesc string `json:"usage_type_desc"`
	UsageUnitType string `json:"usage_unit_type"`
	UsageTypeName string `json:"usage_type_name"`
	IsEditable    bool   `json:"is_editable"`
	UsageTypeCode string `json:"usage_type_code,omitempty"`
}
