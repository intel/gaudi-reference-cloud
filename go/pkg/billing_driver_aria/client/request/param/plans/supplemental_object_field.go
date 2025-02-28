// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package plans

type SupplementalObjectField struct {
	FieldName  string   `json:"field_name,omitempty" url:"field_name,omitempty"`
	FieldValue []string `json:"field_value,omitempty" url:"field_value,omitempty"`
}
