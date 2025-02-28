// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type AccountPaymentMethods struct {
	BillFirstName         string `json:"bill_first_name,omitempty"`
	BillMiddelInitial     string `json:"bill_middle_initial,omitempty"`
	BillLastName          string `json:"bill_last_name,omitempty"`
	BillComapanyName      string `json:"bill_company_name,omitempty"`
	BillAddress1          string `json:"bill_address1,omitempty"`
	BillAddress2          string `json:"bill_address2,omitempty"`
	BillAddress3          string `json:"bill_address3,omitempty"`
	BillCity              string `json:"bill_city,omitempty"`
	BillLocality          string `json:"bill_locality,omitempty"`
	BillCountry           string `json:"bill_country,omitempty"`
	BillPostalCd          string `json:"bill_postal_cd,omitempty"`
	BillCellPhone         string `json:"bill_cell_phone,omitempty"`
	BillEmail             string `json:"bill_email,omitempty"`
	BillBirthdate         string `json:"bill_birthdate,omitempty"`
	CCExpireMonth         int64  `json:"cc_expire_mm,omitempty"`
	CCExpireYear          int64  `json:"cc_expire_yyyy,omitempty"`
	CCId                  int64  `json:"cc_id,omitempty"`
	CCType                string `json:"cc_type,omitempty"`
	ClientPaymentMethodId string `json:"client_payment_method_id,omitempty"`
	PaymentMethodNo       int64  `json:"payment_method_no,omitempty"`
	Suffix                string `json:"suffix,omitempty"`
	PaymentMethodType     int64  `json:"pay_method_type,omitempty"`
	Status                int64  `json:"status,omitempty"`
}
