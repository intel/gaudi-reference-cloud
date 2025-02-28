// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/accts"
)

type UpdateAcctCompleteMRequest struct {
	AriaRequest
	OutputFormat               string                             `json:"output_format"`
	ClientNo                   int64                              `json:"client_no"`
	AuthKey                    string                             `json:"auth_key"`
	AltCallerId                string                             `json:"alt_caller_id"`
	ClientReceiptId            string                             `json:"client_receipt_id,omitempty"`
	ApplicationId              string                             `json:"application_id,omitempty"`
	ApplicationDate            string                             `json:"application_date,omitempty"`
	AcctNo                     int64                              `json:"acct_no,omitempty"`
	ClientAcctId               string                             `json:"client_acct_id,omitempty"`
	SeniorAcctNo               int64                              `json:"senior_acct_no,omitempty"`
	SeniorAcctUserid           string                             `json:"senior_acct_userid,omitempty"`
	SeniorClientAcctId         string                             `json:"senior_client_acct_id,omitempty"`
	TestAcctInd                int64                              `json:"test_acct_ind,omitempty"`
	AltClientAcctGroupId       string                             `json:"alt_client_acct_group_id,omitempty"`
	FunctionalAcctGroupsUpdate []accts.FunctionalAcctGroupsUpdate `json:"functional_acct_groups_update,omitempty"`
	CollectionAcctGroupsUpdate []accts.CollectionAcctGroupsUpdate `json:"collection_acct_groups_update,omitempty"`
}
