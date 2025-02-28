// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package plans

type Service struct {
	Name             string `json:"name,omitempty" url:"name,omitempty"`
	ClientServiceId  string `json:"client_service_id,omitempty" url:"client_service_id,omitempty"`
	ServiceType      string `json:"service_type" url:"service_type"`
	GlCd             string `json:"gl_cd,omitempty" url:"gl_cd,omitempty"`
	TaxableInd       int    `json:"taxable_ind,omitempty" url:"taxable_ind,omitempty"`
	TaxGroup         int    `json:"tax_group,omitempty" url:"tax_group,omitempty"`
	ClientTaxGroupId string `json:"client_tax_group_id,omitempty" url:"client_tax_group_id,omitempty"`
	RateType         string `json:"rate_type,omitempty" url:"rate_type,omitempty"`
	PricingRule      string `json:"pricing_rule,omitempty" url:"pricing_rule,omitempty"`
	HighWater        int    `json:"high_water,omitempty" url:"high_water,omitempty"`
	TaxInclusiveInd  int    `json:"tax_inclusive_ind,omitempty" url:"tax_inclusive_ind,omitempty"`
	Tier             []Tier `json:"tier,omitempty" url:"tier,omitempty"`
	UsageType        int    `json:"usage_type,omitempty" url:"usage_type"`
}
