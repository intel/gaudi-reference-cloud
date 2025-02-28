// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type CreateUsageType struct {
	AriaRequest
	OutputFormat      string `json:"output_format" url:"output_format"`
	ClientNo          int64  `json:"client_no" url:"client_no"`
	AuthKey           string `json:"auth_key" url:"auth_key"`
	UsageTypeName     string `json:"usage_type_name" url:"usage_type_name"`
	UsageTypeDesc     string `json:"usage_type_desc" url:"usage_type_desc"`
	UsageUnitTypeNo   int    `json:"usage_unit_type_no" url:"usage_unit_type_no"`
	UsageTypeCode     string `json:"usage_type_code" url:"usage_type_code"`
	UsageRatingTiming int    `json:"usage_rating_timing" url:"usage_rating_timing"`
}
