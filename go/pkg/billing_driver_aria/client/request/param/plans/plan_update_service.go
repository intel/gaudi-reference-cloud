// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package plans

type PlanUpdateService struct {
	ServiceNo           int64  `json:"service_no,omitempty"`
	ClientServiceId     string `json:"client_service_id,omitempty"`
	SvcLocationNo       int64  `json:"svc_location_no,omitempty"`
	ClientSvcLocationId string `json:"client_svc_location_id,omitempty"`
	DestContactNo       int64  `json:"dest_contact_no,omitempty"`
}
