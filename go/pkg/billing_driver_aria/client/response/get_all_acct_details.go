// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetAcctDetailsAllMResponse struct {
	AriaResponse
	ClientAcctId            string                     `json:"client_acct_id,omitempty"`
	Userid                  string                     `json:"userid,omitempty"`
	FirstName               string                     `json:"first_name,omitempty"`
	MiddleInitial           string                     `json:"middle_initial,omitempty"`
	LastName                string                     `json:"last_name,omitempty"`
	CompanyName             string                     `json:"company_name,omitempty"`
	Address1                string                     `json:"address1,omitempty"`
	Address2                string                     `json:"address2,omitempty"`
	Address3                string                     `json:"address3,omitempty"`
	City                    string                     `json:"city,omitempty"`
	Locality                string                     `json:"locality,omitempty"`
	StateProv               string                     `json:"state_prov,omitempty"`
	CountryCd               string                     `json:"country_cd,omitempty"`
	PostalCd                string                     `json:"postal_cd,omitempty"`
	Phone                   string                     `json:"phone,omitempty"`
	PhoneExt                string                     `json:"phone_ext,omitempty"`
	CellPhone               string                     `json:"cell_phone,omitempty"`
	Email                   string                     `json:"email,omitempty"`
	Birthdate               string                     `json:"birthdate,omitempty"`
	StatusCd                int64                      `json:"status_cd,omitempty"`
	NotifyMethod            int64                      `json:"notify_method,omitempty"`
	TestAcctInd             int64                      `json:"test_acct_ind,omitempty"`
	AcctStartDate           string                     `json:"acct_start_date,omitempty"`
	AltMsgTemplateNo        int64                      `json:"alt_msg_template_no,omitempty"`
	SeqFuncGroupNo          int64                      `json:"seq_func_group_no,omitempty"`
	InvoiceApprovalRequired int64                      `json:"invoice_approval_required,omitempty"`
	FunctionalAcctGroup     []data.FunctionalAcctGroup `json:"functional_acct_group,omitempty"`
	AcctCurrency            string                     `json:"acct_currency,omitempty"`
	BillingGroupsInfo       []data.BillingGroupsInfo   `json:"billing_groups_info,omitempty"`
	PaymentMethodsInfo      []data.PaymentMethodsInfo  `json:"payment_methods_info,omitempty"`
	MasterPlanCount         int64                      `json:"master_plan_count,omitempty"`
	SuppPlanCount           int64                      `json:"supp_plan_count,omitempty"`
	MasterPlansInfo         []data.MasterPlansInfo     `json:"master_plans_info,omitempty"`
	ConsumerAcctInd         int64                      `json:"consumer_acct_ind,omitempty"`
	AcctLocaleName          string                     `json:"acct_locale_name,omitempty"`
	AcctNo2                 int64                      `json:"acct_no_2,omitempty"`
	AcctLocaleNo2           int64                      `json:"acct_locale_no_2,omitempty"`
	AcctContactNo2          int64                      `json:"acct_contact_no_2,omitempty"`
	ChiefAcctInfo           []data.ChiefAcctInfo       `json:"chief_acct_info,omitempty"`
	SuppField               []data.SuppField           `json:"supp_field,omitempty"`
	CollectionAcctGroup     []data.CollectionAcctGroup `json:"collection_acct_group,omitempty"`
}
