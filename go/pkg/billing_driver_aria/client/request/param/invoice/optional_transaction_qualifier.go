// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package invoice

type OptionalTransactionQualifier struct {
	QualifierName  string `json:"qualifier_name,omitempty"`
	QualifierValue string `json:"qualifier_value,omitempty"`
}
