// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usages

type ErrorRecords struct {
	OutClientRecordId string `json:"out_client_record_id,omitempty"`
	AcctNo            int64  `json:" acct_no,omitempty"`
	RecordErrorCode   int64  `json:" record_error_code,omitempty"`
	RecordErrorMsg    string `json:"record_error_msg,omitempty"`
}
