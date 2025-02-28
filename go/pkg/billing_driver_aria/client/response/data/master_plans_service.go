// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type MasterPlansService struct {
	ServiceNo       int64  `json:"service_no,omitempty"`
	ClientServiceId string `json:"client_service_id,omitempty"`
	TaxInclusiveInd int64  `json:"tax_inclusive_ind,omitempty"`
}
