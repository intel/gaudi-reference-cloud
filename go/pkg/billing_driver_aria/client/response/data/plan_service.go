// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PlanService struct {
	ServiceNo            int64             `json:"service_no,omitempty"`
	ServiceDesc          string            `json:"service_desc,omitempty"`
	IsRecurringInd       int64             `json:"is_recurring_ind,omitempty"`
	IsUsageBasedInd      int64             `json:"is_usage_based_ind,omitempty"`
	UsageType            int64             `json:"usage_type,omitempty"`
	TaxableInd           int64             `json:"taxable_ind,omitempty"`
	IsTaxInd             int64             `json:"is_tax_ind,omitempty"`
	IsArrearsInd         int64             `json:"is_arrears_ind,omitempty"`
	IsSetupInd           int64             `json:"is_setup_ind,omitempty"`
	IsMiscInd            int64             `json:"is_misc_ind,omitempty"`
	IsDonationInd        int64             `json:"is_donation_ind,omitempty"`
	IsOrderBasedInd      int64             `json:"is_order_based_ind,omitempty"`
	IsCancellationInd    int64             `json:"is_cancellation_ind,omitempty"`
	CoaId                string            `json:"coa_id,omitempty"`
	LedgerCode           string            `json:"ledger_code,omitempty"`
	ClientCoaCode        string            `json:"client_coa_code,omitempty"`
	DisplayInd           int64             `json:"display_ind,omitempty"`
	TieredPricingRule    int64             `json:"tiered_pricing_rule,omitempty"`
	IsMinFeeInd          int64             `json:"is_min_fee_ind,omitempty"`
	FulfillmentBasedInd  int64             `json:"fulfillment_based_ind,omitempty"`
	PlanServiceRates     []PlanServiceRate `json:"plan_service_rates,omitempty"`
	ApplyUsageRatesDaily int64             `json:"apply_usage_rates_daily,omitempty"`
	TaxInclusiveInd      int64             `json:"tax_inclusive_ind,omitempty"`
	ClientServiceId      string            `json:"client_service_id,omitempty"`
	UsageTypeCd          string            `json:"usage_type_cd,omitempty"`
	UsageTypeName        string            `json:"usage_type_name,omitempty"`
	UsageTypeDesc        string            `json:"usage_type_desc,omitempty"`
	UsageTypeCode        string            `json:"usage_type_code,omitempty"`
	UsageUnitLabel       string            `json:"usage_unit_label,omitempty"`
}
