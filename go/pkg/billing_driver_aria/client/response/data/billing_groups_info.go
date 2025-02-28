// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type BillingGroupsInfo struct {
	BillingGroupNo           int64               `json:"billing_group_no,omitempty"`
	BillingGroupName         string              `json:"billing_group_name,omitempty"`
	BillingGroupDescription  string              `json:"billing_group_description,omitempty"`
	ClientBillingGroupId     string              `json:"client_billing_group_id,omitempty"`
	Status                   int64               `json:"status,omitempty"`
	NotifyMethod             int64               `json:"notify_method,omitempty"`
	NotifyTemplateGroup      int64               `json:"notify_template_group,omitempty"`
	StatementTemplate        float32             `json:"statement_template,omitempty"`
	CreditNoteTemplate       int64               `json:"credit_note_template,omitempty"`
	PaymentOption            string              `json:"payment_option,omitempty"`
	PrimaryPaymentMethodName string              `json:"primary_payment_method_name,omitempty"`
	PrimaryPaymentMethodId   int64               `json:"primary_payment_method_id,omitempty"`
	PrimaryPaymentMethodNo   int64               `json:"primary_payment_method_no,omitempty"`
	BackupPaymentMethodName  string              `json:"backup_payment_method_name,omitempty"`
	BackupPaymentMethodId    int64               `json:"backup_payment_method_id,omitempty"`
	BackupPaymentMethodNo    int64               `json:"backup_payment_method_no,omitempty"`
	ClientPaymentTermId      string              `json:"client_payment_term_id,omitempty"`
	PaymentTermsName         string              `json:"payment_terms_name,omitempty"`
	PaymentTermsNo           int64               `json:"payment_terms_no,omitempty"`
	PaymentTermsType         string              `json:"payment_terms_type,omitempty"`
	StmtFirstName            string              `json:"stmt_first_name,omitempty"`
	StmtMi                   string              `json:"stmt_mi,omitempty"`
	StmtLastName             string              `json:"stmt_last_name,omitempty"`
	StmtCompanyName          string              `json:"stmt_company_name,omitempty"`
	StmtAddress1             string              `json:"stmt_address1,omitempty"`
	StmtAddress2             string              `json:"stmt_address2,omitempty"`
	StmtAddress3             string              `json:"stmt_address3,omitempty"`
	StmtCity                 string              `json:"stmt_city,omitempty"`
	StmtLocality             string              `json:"stmt_locality,omitempty"`
	StmtStateProv            string              `json:"stmt_state_prov,omitempty"`
	StmtCountry              string              `json:"stmt_country,omitempty"`
	StmtPostalCd             string              `json:"stmt_postal_cd,omitempty"`
	StmtPhone                string              `json:"stmt_phone,omitempty"`
	StmtPhoneExt             string              `json:"stmt_phone_ext,omitempty"`
	StmtCellPhone            string              `json:"stmt_cell_phone,omitempty"`
	StmtEmail                string              `json:"stmt_email,omitempty"`
	StmtBirthdate            string              `json:"stmt_birthdate,omitempty"`
	MasterPlanSummary        []MasterPlanSummary `json:"master_plans_summary"`
}
