// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/accts"
)

type GetAcctPlansAllMRequest struct {
	AriaRequest
	OutputFormat                   string                           `json:"output_format"`
	ClientNo                       int64                            `json:"client_no"`
	ClientAcctId                   string                           `json:"client_acct_id,omitempty"`
	AuthKey                        string                           `json:"auth_key"`
	AcctNo                         int64                            `json:"acct_no,omitempty"`
	IncludeServiceSuppFields       string                           `json:"include_service_supp_fields,omitempty"`
	IncludeProductFields           string                           `json:"include_product_fields,omitempty"`
	IncludePlanServices            string                           `json:"include_plan_services,omitempty"`
	IncludeSurcharges              string                           `json:"include_surcharges,omitempty"`
	IncludeRateSchedule            string                           `json:"include_rate_schedule,omitempty"`
	IncludeContractAndRolloverInfo string                           `json:"include_contract_and_rollover_info,omitempty"`
	IncludeDunningInfo             string                           `json:"include_dunning_info,omitempty"`
	ProductCatalogPlanFilter       []accts.ProductCatalogPlanFilter `json:"product_catalog_plan_filter,omitempty"`
}
