// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetInvoiceDetailsMResponse struct {
	AriaResponse
	InvoiceLineDetails []data.InvoiceLineDetail `json:"invoice_line_details,omitempty"`
}
