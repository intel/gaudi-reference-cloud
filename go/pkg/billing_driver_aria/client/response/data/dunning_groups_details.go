// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type DunningGroupsDetails struct {
	DunningGroupNo          int64               `json:"dunning_group_no,omitempty"`
	DunningGroupName        string              `json:"dunning_group_name,omitempty"`
	DunningGroupDescription string              `json:"dunning_group_description,omitempty"`
	ClientDunningGroupId    string              `json:"client_dunning_group_id,omitempty"`
	DunningProcessNo        int64               `json:"dunning_process_no,omitempty"`
	ClientDunningProcessId  string              `json:"client_dunning_process_id,omitempty"`
	Status                  int64               `json:"status,omitempty"`
	MasterPlanSummary       []MasterPlanSummary `json:"master_plans_summary"`
}
