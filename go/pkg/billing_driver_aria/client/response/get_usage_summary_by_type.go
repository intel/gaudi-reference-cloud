// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"

type GetUsageSummaryByType struct {
	AriaResponse
	AcctLocaleNo     int64                  `json:"acct_locale_no"`
	AcctLocaleName   string                 `json:"acct_locale_name"`
	StartDate        string                 `json:"start_date,omitempty"`
	StartTime        string                 `json:"start_time,omitempty"`
	EndDate          string                 `json:"end_date,omitempty"`
	EndTime          string                 `json:"end_time,omitempty"`
	UsageSummaryRecs []data.UsageSummaryRec `json:"usage_summary_recs,omitempty"`
}
