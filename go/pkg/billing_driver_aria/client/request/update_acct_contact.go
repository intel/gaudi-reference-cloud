// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type UpdateAccountContact struct {
	AriaRequest
	ReleaseVersion       string `json:"releaseVersion"`
	OutputFormat         string `json:"output_format"`
	ClientNo             int64  `json:"client_no"`
	AuthKey              string `json:"auth_key"`
	AcctNo               int64  `json:"acct_no,omitempty"`
	ContactInd           int64  `json:"contact_ind"`
	BillingGroupNo       int64  `json:"billing_group_no,omitempty"`
	ClientBillingGroupId string `json:"client_billing_group_id,omitempty"`
	FirstName            string `json:"first_name,omitempty"`
	MiddleInitial        string `json:"middle_initial,omitempty"`
	LastName             string `json:"last_name,omitempty"`
	CompanyName          string `json:"company_name,omitempty"`
	Address1             string `json:"address1,omitempty"`
	Address2             string `json:"address2,omitempty"`
	Address3             string `json:"address3,omitempty"`
	City                 string `json:"city,omitempty"`
	Locality             string `json:"locality,omitempty"`
	StateProv            string `json:"state_prov,omitempty"`
	CountryCd            string `json:"country_cd,omitempty"`
	PostalCd             string `json:"postal_cd,omitempty"`
	Phone                string `json:"phone,omitempty"`
	PhoneExt             string `json:"phone_ext,omitempty"`
	CellPhone            string `json:"cell_phone,omitempty"`
	WorkPhone            string `json:"work_phone,omitempty"`
	WorkPhoneExt         string `json:"work_phone_ext,omitempty"`
	Fax                  string `json:"fax,omitempty"`
	Email                string `json:"email,omitempty"`
	Birthdate            string `json:"birthdate,omitempty"`
	ClientAcctId         string `json:"client_acct_id,omitempty"`
	ApplicationId        string `json:"application_id,omitempty"`
	ApplicationDate      string `json:"application_date,omitempty"`
	AltCallerId          string `json:"alt_caller_id,omitempty"`
}
