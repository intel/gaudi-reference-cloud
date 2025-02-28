// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetInvoiceHistory struct {
	AriaResponse
	InvoiceHist []data.InvoiceHist `json:"invoice_hist,omitempty"`
}
