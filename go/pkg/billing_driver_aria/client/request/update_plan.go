// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/plans"
)

type UpdatePlan struct {
	RestCall                 string                          `json:"rest_call"`
	OutputFormat             string                          `json:"output_format"`
	ClientNo                 int64                           `json:"client_no"`
	AuthKey                  string                          `json:"auth_key"`
	AltCallerId              string                          `json:"alt_caller_id"`
	ClientPlanId             string                          `json:"client_plan_id"`
	PlanName                 string                          `json:"plan_name"`
	PlanType                 int                             `json:"plan_type"`
	EditDirectives           int                             `json:"edit_directives"`
	Currency                 string                          `json:"currency"`
	Active                   int                             `json:"active"`
	Schedule                 []plans.Schedule                `json:"schedule,omitempty"`
	Service                  []plans.Service                 `json:"service,omitempty"`
	ProrationInvoiceTimingCd string                          `json:"proration_invoice_timing_cd,omitempty"`
	SupplementalObjField     []plans.SupplementalObjectField `json:"supplemental_obj_field,omitempty"`
}
