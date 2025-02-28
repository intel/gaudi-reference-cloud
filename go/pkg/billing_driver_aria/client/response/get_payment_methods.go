// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetPaymentMethodsMResponse struct {
	AriaResponse
	AccountPaymentMethods []data.AccountPaymentMethods `json:"account_payment_methods,omitempty"`
}
