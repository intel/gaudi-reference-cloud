// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package accts

type DunningGroup struct {
	DunningGroupName        string `json:"dunning_group_name,omitempty"`
	DunningGroupDescription string `json:"dunning_group_description,omitempty"`
	ClientDunningGroupId    string `json:"client_dunning_group_id,omitempty"`
	DunningGroupIdx         int64  `json:"dunning_group_idx,omitempty"`
	ClientDunningProcessId  string `json:"client_dunning_process_id,omitempty"`
}
