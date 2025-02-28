// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"

type GetUsageHistory struct {
	AriaResponse
	AcctLocaleNo       int64                  `json:"acct_locale_no"`
	AcctLocaleName     string                 `json:"acct_locale_name"`
	UsageHistoryRecs   []data.UsageHistoryRec `json:"usage_history_recs,omitempty"`
	FilteredUsageCount int64                  `json:"filtered_usage_count,omitempty"`
}
