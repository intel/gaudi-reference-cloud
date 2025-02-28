// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type BillingContactInfo struct {
	PaymentMethodNo  int64 `json:"payment_method_no,omitempty"`
	BillingContactNo int64 `json:"bill_contact_no,omitempty"`
}
