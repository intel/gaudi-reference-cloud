// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package accts

import (
	"encoding/json"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

var defaults []byte = []byte(`
rest_call: create_acct_complete_m
output_format: json
acct:
  notify_method: 10
  acct_currency: USD
  invoicing_option: 1
  client_seq_func_group_id: LE036
  functional_acct_group:
    - client_functional_acct_group_id: LE036
  supp_field:
    - supp_field_name: Auto approval by workflow
      supp_field_value: 'True'
    - supp_field_name: Company Code
      supp_field_value: '036'
  billing_group:
    - billing_group_idx: 1
      notify_method: 10
      client_notification_template_group_id: US_Statement_Template
  dunning_group:
    - dunning_group_idx: 1
      client_dunning_process_id: Standard_Dunning
  master_plans_detail:
    - client_plan_id: Customer Account Plan
      plan_instance_idx: 1
      plan_instance_units: 1
      billing_group_idx: 1
      dunning_group_idx: 1
      plan_instance_status: 1
`)

type AcctRequest struct {
	RestCall     string `json:"rest_call,omitempty"`
	OutputFormat string `json:"output_format,omitempty"`
	Account      Acct   `json:"acct"`
}

func mapKeysToString(in map[any]any) map[string]any {
	out := map[string]any{}
	for key, val := range in {
		valMap, ok := val.(map[any]any)
		if ok {
			val = mapKeysToString(valMap)
		}
		out[fmt.Sprintf("%v", key)] = val
	}
	return out
}

func TestPopulateAcct(t *testing.T) {
	mm := map[any]any{}

	if err := yaml.Unmarshal(defaults, &mm); err != nil {
		t.Fatalf("yaml unmarshal error: %v", err)
	}
	bytes, err := json.Marshal(mapKeysToString(mm))
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	acctReq := AcctRequest{}
	if err := json.Unmarshal(bytes, &acctReq); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	t.Logf("acct: %v", &acctReq)
}
