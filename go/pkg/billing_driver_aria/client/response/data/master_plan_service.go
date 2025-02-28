// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type MasterPlanService struct {
	ServiceNo           int64  `json:"service_no,omitempty"`
	ClientServiceId     string `json:"client_service_id,omitempty"`
	SvcLocationNo       int64  `json:"svc_location_no,omitempty"`
	ClientSvcLocationId string `json:"client_svc_location_id,omitempty"`
	DestContactIdx      int64  `json:"dest_contact_idx,omitempty"`
}
