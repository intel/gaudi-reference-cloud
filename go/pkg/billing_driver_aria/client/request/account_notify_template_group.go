// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type AccountNotifyTemplateGroup struct {
	AriaRequest
	OutputFormat                string `json:"output_format" url:"output_format"`
	ClientNo                    int64  `json:"client_no" url:"client_no"`
	AuthKey                     string `json:"auth_key" url:"auth_key"`
	AltCallerId                 string `json:"alt_caller_id" url:"alt_caller_id,omitempty"`
	AcctNo                      int64  `json:"acct_no,omitempty"`
	AcctUserId                  string `json:"acct_user_id,omitempty"`
	ClientAcctId                string `json:"client_acct_id,omitempty"`
	NotificationTemplateGroupId string `json:"notification_template_group_id"`
}
