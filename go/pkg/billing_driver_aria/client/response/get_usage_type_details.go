// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type GetUsageTypeDetails struct {
	AriaResponse
	UsageTypeNo            int    `json:"usage_type_no"`
	UsageTypeName          string `json:"usage_type_name"`
	UsageTypeDesc          string `json:"usage_type_desc"`
	UsageTypeDisplayString string `json:"usage_type_display_string"`
	UsageUnitType          string `json:"usage_unit_type"`
	UsageTypeCode          string `json:"usage_type_code"`
	IsEditable             bool   `json:"is_editable"`
}
