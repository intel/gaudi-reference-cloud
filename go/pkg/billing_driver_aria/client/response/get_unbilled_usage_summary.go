// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import "hash"

type GetUnbilledUsageSummaryMResponse struct {
	AriaResponse
	CurrencyCd                 string      `json:"currency_cd"`
	AccNo                      int         `json:"acc_no"`
	ClientAcctId               string      `json:"client_acct_id"`
	ClientMasterPlanInstanceId string      `json:"client_master_plan_instance_id"`
	MasterPlanInstanceId       int         `json:"master_plan_instance_id,omitempty"`
	MpiMtdThresholdAmount      float64     `json:"mpi_mtd_threshold_amount,omitempty"`
	MpiPtdThresholdAmount      float64     `json:"mpi_ptd_threshold_amount,omitempty"`
	ClientMtdThresholdAmount   float64     `json:"client_mtd_threshold_amount,omitempty"`
	ClientPtdThresholdAmount   float64     `json:"client_ptd_threshold_amount,omitempty"`
	MtdBalanceAmount           float64     `json:"mtd_balance_amount,omitempty"`
	PtdBalanceAmount           float64     `json:"ptd_balance_amount,omitempty"`
	MpiMtdDeltaAmount          float64     `json:"mpi_mtd_delta_amount,omitempty"`
	MpiMtdDeltaSign            string      `json:"mpi_mtd_delta_sign,omitempty"`
	MpiPtdDeltaAmount          float64     `json:"mpi_ptd_delta_amount,omitempty"`
	MpiPtdDeltaSign            string      `json:"mpi_ptd_delta_sign,omitempty"`
	ClientMtdDeltaAmount       float64     `json:"client_mtd_delta_amount,omitempty"`
	ClientMtdDeltaSign         string      `json:"client_mtd_delta_sign,omitempty"`
	ClientPtdDeltaAmount       float64     `json:"client_ptd_delta_amount,omitempty"`
	ClientPtdDeltaSign         string      `json:"client_ptd_delta_sign,omitempty"`
	UnappSvcCreditDeltaSign    string      `json:"unapp_svc_credit_delta_sign,omitempty"`
	UnappSvcCreditBalAmount    float64     `json:"unapp_svc_credit_bal_amount,omitempty"`
	UnappSvcCreditDeltaAmount  float64     `json:"unapp_svc_credit_delta_amount,omitempty"`
	UnbilledUsageRec           hash.Hash64 `json:"unbilled_usage_rec,omitempty"`
	UnitThresholdDetails       hash.Hash64 `json:"unit_threshold_details,omitempty"`
	AcctLocaleNo               int64       `json:"acct_locale_no"`
	AcctLocaleName             string      `json:"acct_locale_name"`
}
