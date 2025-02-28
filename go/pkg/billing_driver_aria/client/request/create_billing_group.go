// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type CreateBillingGroupRequest struct {
	AriaRequest
	OutputFormat                      string `json:"output_format"`
	ClientNo                          int64  `json:"client_no"`
	AuthKey                           string `json:"auth_key"`
	AcctNo                            int64  `json:"acct_no,omitempty"`
	ClientAcctId                      string `json:"client_acct_id,omitempty"`
	NotifyMethod                      int64  `json:"notify_method,omitempty"`
	BillingGroupName                  int64  `json:"billing_group_name,omitempty"`
	BillingGroupDescription           int64  `json:"billing_group_description,omitempty"`
	ClientBillingGroupId              string `json:"client_billing_group_id,omitempty"`
	AltCallerId                       string `json:"alt_caller_id,omitempty"`
	ClientNotificationTemplateGroupId string `json:"client_notification_template_group_id,omitempty"`
	StmtFirstName                     string `json:"stmt_first_name,omitempty"`
	StmtLastName                      string `json:"stmt_last_name,omitempty"`
	StmtCompanyName                   string `json:"stmt_company_name,omitempty"`
	StmtAddress1                      string `json:"stmt_address1,omitempty"`
	StmtCity                          string `json:"stmt_city,omitempty"`
	StmtStateProv                     string `json:"stmt_state_prov,omitempty"`
	StmtCountry                       string `json:"stmt_country,omitempty"`
	StmtPostalCd                      string `json:"stmt_postal_cd,omitempty"`
	StmtPhone                         string `json:"stmt_phone,omitempty"`
	StmtEmail                         string `json:"stmt_email,omitempty"`
}
