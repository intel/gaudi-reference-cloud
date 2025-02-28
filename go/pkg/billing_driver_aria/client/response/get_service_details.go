// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type GetServiceDetails struct {
	AriaResponse
	ServiceNo       string `json:"service_no,omitempty"`
	ClientServiceId string `json:"client_service_id,omitempty"`
	ServiceName     string `json:"service_name,omitempty"`
	ServiceType     string `json:"service_type,omitempty"`
	GlCd            string `json:"gl_cd,omitempty"`
	TaxableInd      int    `json:"taxable_ind,omitempty"`
	TaxGroup        int    `json:"tax_group,omitempty"`
	UsageType       int    `json:"usage_type,omitempty"`
}
