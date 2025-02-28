// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type UpdateAccountBillingGroupRequest struct {
	AriaRequest
	OutputFormat                 string `json:"output_format"`
	ClientNo                     int64  `json:"client_no"`
	AuthKey                      string `json:"auth_key"`
	ClientAccountId              string `json:"client_acct_id"`
	ClientBillingGroupId         string `json:"client_billing_group_id,omitempty"`
	ClientPrimaryPaymentMethodId string `json:"client_primary_payment_method_id,omitempty"`
	ClientPaymentMethodId        string `json:"client_payment_method_id,omitempty"`
	PayMethodType                int    `json:"pay_method_type,omitempty"`
	CCNumber                     int64  `json:"cc_num,omitempty"`
	CCExpireMonth                int    `json:"cc_expire_mm,omitempty"`
	CCExpireYear                 int    `json:"cc_expire_yyyy,omitempty"`
	CCV                          int    `json:"ccv,omitempty"`
	PrimaryPaymentMethodNo       int64  `json:"primary_payment_method_no,omitempty"`
}
