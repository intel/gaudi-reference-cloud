// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/plans"
)

type PlanRequest struct {
	AriaRequest
	OutputFormat             string                          `json:"output_format,omitempty" url:"output_format,omitempty"`
	ClientNo                 int64                           `json:"client_no" url:"client_no"`
	AuthKey                  string                          `json:"auth_key" url:"auth_key"`
	AltCallerId              string                          `json:"alt_caller_id" url:"alt_caller_id"`
	ClientPlanId             string                          `json:"client_plan_id,omitempty" url:"client_plan_id,omitempty"`
	PlanName                 string                          `json:"plan_name" url:"plan_name"`
	PlanDescription          string                          `json:"plan_description,omitempty" url:"plan_description,omitempty"`
	PlanType                 string                          `json:"plan_type" url:"plan_type"`
	Currency                 string                          `json:"currency" url:"currency"`
	Active                   int                             `json:"active" url:"active"`
	Schedule                 []plans.Schedule                `json:"schedule,omitempty" url:"schedule,omitempty"`
	Service                  []plans.Service                 `json:"service,omitempty" url:"service,omitempty"`
	TemplateInd              int                             `json:"template_ind,omitempty" url:"template_ind,omitempty"`
	ProrationInvoiceTimingCd string                          `json:"proration_invoice_timing_cd,omitempty" url:"proration_invoice_timing_cd,omitempty"`
	SupplementalObjField     []plans.SupplementalObjectField `json:"supplemental_obj_field,omitempty" url:"supplemental_obj_field,omitempty"`
	EditDirectives           int                             `json:"edit_directives" url:"edit_directives"`
	// EditDirectives is used only in the 'Edit and Deactivate Plan' API, not for CreatePlan
}
