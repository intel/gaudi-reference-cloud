// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PaymentMethodsInfo struct {
	BillingGroupNo             int64   `json:"billing_group_no,omitempty"`
	ClientDefBillingGroupId    string  `json:"client_def_billing_group_id,omitempty"`
	PaymentMethodNo            int64   `json:"payment_method_no,omitempty"`
	BillFirstName              string  `json:"bill_first_name,omitempty"`
	BillMiddleInitial          string  `json:"bill_middle_initial,omitempty"`
	BillLastName               string  `json:"bill_last_name,omitempty"`
	BillCompanyName            string  `json:"bill_company_name,omitempty"`
	BillAddress1               string  `json:"bill_address1,omitempty"`
	BillAddress2               string  `json:"bill_address2,omitempty"`
	BillAddress3               string  `json:"bill_address3,omitempty"`
	BillCity                   string  `json:"bill_city,omitempty"`
	BillLocality               string  `json:"bill_locality,omitempty"`
	BillStateProv              string  `json:"bill_state_prov,omitempty"`
	BillCountry                string  `json:"bill_country,omitempty"`
	BillPostalCd               string  `json:"bill_postal_cd,omitempty"`
	BillAddressMatchScore      float32 `json:"bill_address_match_score,omitempty"`
	BillPhone                  string  `json:"bill_phone,omitempty"`
	BillPhoneExt               string  `json:"bill_phone_ext,omitempty"`
	BillCellPhone              string  `json:"bill_cell_phone,omitempty"`
	BillEmail                  string  `json:"bill_email,omitempty"`
	BillBirthdate              string  `json:"bill_birthdate,omitempty"`
	PayMethodName              string  `json:"pay_method_name,omitempty"`
	ClientPaymentMethodId      string  `json:"client_payment_method_id,omitempty"`
	PayMethodDescription       string  `json:"pay_method_description,omitempty"`
	PayMethodType              int64   `json:"pay_method_type,omitempty"`
	Suffix                     string  `json:"suffix,omitempty"`
	CcExpireMm                 int64   `json:"cc_expire_mm,omitempty"`
	CcExpireYyyy               int64   `json:"cc_expire_yyyy,omitempty"`
	CcType                     string  `json:"cc_type,omitempty"`
	BankRoutingNum             string  `json:"bank_routing_num,omitempty"`
	BillAgreementId            string  `json:"bill_agreement_id,omitempty"`
	BankSwiftCd                string  `json:"bank_swift_cd,omitempty"`
	BankCountryCd              string  `json:"bank_country_cd,omitempty"`
	MandateId                  string  `json:"mandate_id,omitempty"`
	BankIdCd                   string  `json:"bank_id_cd,omitempty"`
	BankBranchCd               string  `json:"bank_branch_cd,omitempty"`
	BillContactNo              int64   `json:"bill_contact_no,omitempty"`
	BillAddressVerificationCd2 string  `json:"bill_address_verification_cd_2,omitempty"`
}
