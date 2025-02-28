// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/invoice"
)

type ManageInvoice struct {
	AriaRequest
	OutputFormat                  string                                 `json:"output_format"`
	ClientNo                      int64                                  `json:"client_no"`
	AuthKey                       string                                 `json:"auth_key"`
	AltCallerId                   string                                 `json:"alt_caller_id"`
	InvoiceNo                     int64                                  `json:"invoice_no,omitempty"`
	AcctNo                        int64                                  `json:"acct_no,omitempty"`
	ClientAcctId                  string                                 `json:"client_acct_id,omitempty"`
	MasterPlanInstanceNo          int64                                  `json:"master_plan_instance_no,omitempty"`
	ClientMasterPlanInstanceId    string                                 `json:"client_master_plan_instance_id,omitempty"`
	ClientReceiptId               string                                 `json:"client_receipt_id,omitempty"`
	AltBillDay                    int64                                  `json:"alt_bill_day,omitempty"`
	ActionDirective               int64                                  `json:"action_directive,omitempty"`
	BillSeq                       int64                                  `json:"bill_seq,omitempty"`
	AltPayMethod                  int64                                  `json:"alt_pay_method,omitempty"`
	CcNumber                      string                                 `json:"cc_number,omitempty"`
	CcExpireMm                    int64                                  `json:"cc_expire_mm,omitempty"`
	CcExpireYyyy                  int64                                  `json:"cc_expire_yyyy,omitempty"`
	BankRoutingNum                string                                 `json:"bank_routing_num,omitempty"`
	BankAcctNum                   string                                 `json:"bank_acct_num,omitempty"`
	BillCompanyName               string                                 `json:"bill_company_name,omitempty"`
	BillFirstName                 string                                 `json:"bill_first_name,omitempty"`
	BillMiddleInitial             string                                 `json:"bill_middle_initial,omitempty"`
	BillLastName                  string                                 `json:"bill_last_name,omitempty"`
	BillAddress1                  string                                 `json:"bill_address1,omitempty"`
	BillAddress2                  string                                 `json:"bill_address2,omitempty"`
	BillCity                      string                                 `json:"bill_city,omitempty"`
	BillLocality                  string                                 `json:"bill_locality,omitempty"`
	BillStateProv                 string                                 `json:"bill_state_prov,omitempty"`
	BillZip                       string                                 `json:"bill_zip,omitempty"`
	BillCountry                   string                                 `json:"bill_country,omitempty"`
	BillEmail                     string                                 `json:"bill_email,omitempty"`
	BillPhone                     string                                 `json:"bill_phone,omitempty"`
	BillPhoneExtension            string                                 `json:"bill_phone_extension,omitempty"`
	BillCellPhone                 string                                 `json:"bill_cell_phone,omitempty"`
	BillWorkPhone                 string                                 `json:"bill_work_phone,omitempty"`
	BillWorkPhoneExtension        string                                 `json:"bill_work_phone_extension,omitempty"`
	Cvv                           string                                 `json:"cvv,omitempty"`
	AltCollectOnApprove           string                                 `json:"alt_collect_on_approve,omitempty"`
	AltSendStatementOnApprove     string                                 `json:"alt_send_statement_on_approve,omitempty"`
	CancelOrdersOnDiscard         string                                 `json:"cancel_orders_on_discard,omitempty"`
	BankAcctType                  string                                 `json:"bank_acct_type,omitempty"`
	BillDriversLicenseNo          string                                 `json:"bill_drivers_license_no,omitempty"`
	BillDriversLicenseState       string                                 `json:"bill_drivers_license_state,omitempty"`
	BillTaxpayerId                string                                 `json:"bill_taxpayer_id,omitempty"`
	BillAddress3                  string                                 `json:"bill_address3,omitempty"`
	TrackData1                    string                                 `json:"track_data1,omitempty"`
	TrackData2                    string                                 `json:"track_data2,omitempty"`
	Iban                          string                                 `json:"iban,omitempty"`
	BankCheckDigit                int64                                  `json:"bank_check_digit,omitempty"`
	BankSwiftCd                   string                                 `json:"bank_swift_cd,omitempty"`
	BankCountryCd                 string                                 `json:"bank_country_cd,omitempty"`
	MandateId                     string                                 `json:"mandate_id,omitempty"`
	BankIdCd                      string                                 `json:"bank_id_cd,omitempty"`
	BankBranchCd                  string                                 `json:"bank_branch_cd,omitempty"`
	CustomStatusLabel             string                                 `json:"custom_status_label,omitempty"`
	ClientNotes                   string                                 `json:"client_notes,omitempty"`
	RecurringProcessingModelInd   int64                                  `json:"recurring_processing_model_ind,omitempty"`
	BillContactNo                 int64                                  `json:"bill_contact_no,omitempty"`
	ProcFieldOverride             []invoice.ProcFieldOverride            `json:"proc_field_override,omitempty"`
	BankName                      string                                 `json:"bank_name,omitempty"`
	BankCity                      string                                 `json:"bank_city,omitempty"`
	OptionalTransactionQualifiers []invoice.OptionalTransactionQualifier `json:"optional_transaction_qualifiers,omitempty"`
}
