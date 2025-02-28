// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package accts

type Acct struct {
	Acct                              []Acct                `json:"accts,omitempty"`
	ClientAcctId                      string                `json:"client_acct_id,omitempty"`
	Userid                            string                `json:"userid,omitempty"`
	NotifyMethod                      int64                 `json:"notify_method,omitempty"`
	FirstName                         string                `json:"first_name,omitempty"`
	Mi                                string                `json:"mi,omitempty"`
	LastName                          string                `json:"last_name,omitempty"`
	Address1                          string                `json:"address1,omitempty"`
	Address2                          string                `json:"address2,omitempty"`
	Address3                          string                `json:"address3,omitempty"`
	City                              string                `json:"city,omitempty"`
	Locality                          string                `json:"locality,omitempty"`
	StateProv                         string                `json:"state_prov,omitempty"`
	Country                           string                `json:"country,omitempty"`
	PostalCd                          string                `json:"postal_cd,omitempty"`
	Phone                             string                `json:"phone,omitempty"`
	PhoneExt                          string                `json:"phone_ext,omitempty"`
	CellPhone                         string                `json:"cell_phone,omitempty"`
	Email                             string                `json:"email,omitempty"`
	Birthdate                         string                `json:"birthdate,omitempty"`
	InvoicingOption                   int64                 `json:"invoicing_option,omitempty"`
	FunctionalAcctGroup               []FunctionalAcctGroup `json:"functional_acct_group,omitempty"`
	SuppField                         []SuppField           `json:"supp_field,omitempty"`
	TestAcctInd                       int64                 `json:"test_acct_ind,omitempty"`
	AcctCurrency                      string                `json:"acct_currency,omitempty"`
	BillingGroup                      []BillingGroup        `json:"billing_group,omitempty"`
	DunningGroup                      []DunningGroup        `json:"dunning_group"`
	MasterPlansDetail                 []MasterPlansDetail   `json:"master_plans_detail"`
	ClientSeqFuncGroupId              string                `json:"client_seq_func_group_id,omitempty"`
	InvoiceApprovalRequired           string                `json:"invoice_approval_required,omitempty"`
	RetroactiveStartDate              string                `json:"retroactive_start_date,omitempty"`
	NotificationTemplateGroupNo       int64                 `json:"notification_template_group_no,omitempty"`
	ClientNotificationTemplateGroupId string                `json:"client_notification_template_group_id,omitempty"`
}
