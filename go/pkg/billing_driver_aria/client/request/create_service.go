// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type CreateService struct {
	AriaRequest
	OutputFormat        string `json:"output_format" url:"output_format"`
	ClientNo            int64  `json:"client_no" url:"client_no"`
	AuthKey             string `json:"auth_key" url:"auth_key"`
	AltCallerId         string `json:"alt_caller_id" url:"alt_caller_id,omitempty"`
	ServiceName         string `json:"service_name" url:"service_name"`
	ClientServiceId     string `json:"client_service_id" url:"client_service_id,omitempty"`
	ServiceType         string `json:"service_type" url:"service_type"`
	UsageType           int    `json:"usage_type" url:"usage_type"`
	GlCd                string `json:"gl_cd" url:"gl_cd"`
	TaxableInd          int    `json:"taxable_ind" url:"taxable_ind"`
	ClientTaxGroupId    string `json:"client_tax_group_id" url:"client_tax_group_id"`
	AllowServiceCredits string `json:"allow_service_credits" url:"allow_service_credits"`
}
