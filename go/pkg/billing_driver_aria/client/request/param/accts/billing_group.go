// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package accts

type BillingGroup struct {
	BillingGroupName                  string `json:"billing_group_name,omitempty"`
	BillingGroupDescription           string `json:"billing_group_description,omitempty"`
	ClientBillingGroupId              string `json:"client_billing_group_id,omitempty"`
	BillingGroupIdx                   int64  `json:"billing_group_idx,omitempty"`
	NotifyMethod                      int64  `json:"notify_method,omitempty"`
	PaymentOption                     string `json:"payment_option,omitempty"`
	PrimaryClientPaymentMethodId      string `json:"primary_client_payment_method_id,omitempty"`
	PrimaryPaymentMethodIdx           int64  `json:"primary_payment_method_idx,omitempty"`
	BackupClientPaymentMethodId       string `json:"backup_client_payment_method_id,omitempty"`
	BackupPaymentMethodIdx            int64  `json:"backup_payment_method_idx,omitempty"`
	FirstName                         string `json:"first_name,omitempty"`
	Mi                                string `json:"mi,omitempty"`
	LastName                          string `json:"last_name,omitempty"`
	CompanyName                       string `json:"company_name,omitempty"`
	Address1                          string `json:"address1,omitempty"`
	Address2                          string `json:"address2,omitempty"`
	Address3                          string `json:"address3,omitempty"`
	City                              string `json:"city,omitempty"`
	Locality                          string `json:"locality,omitempty"`
	StateProv                         string `json:"state_prov,omitempty"`
	Country                           string `json:"country,omitempty"`
	PostalCd                          string `json:"postal_cd,omitempty"`
	Phone                             string `json:"phone,omitempty"`
	PhoneExt                          string `json:"phone_ext,omitempty"`
	CellPhone                         string `json:"cell_phone,omitempty"`
	WorkPhone                         string `json:"work_phone,omitempty"`
	WorkPhoneExt                      string `json:"work_phone_ext,omitempty"`
	Fax                               string `json:"fax,omitempty"`
	Email                             string `json:"email,omitempty"`
	Birthdate                         string `json:"birthdate,omitempty"`
	ClientNotificationTemplateGroupId string `json:"client_notification_template_group_id,omitempty"`
}
