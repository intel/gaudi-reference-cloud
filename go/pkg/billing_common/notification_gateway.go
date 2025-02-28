// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"
	"errors"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

//Intended for notification of different types - log, SMTP, MMS, etc.
// Todo: Identify how to do structured logging to notify Ops.

// NotifyInvalidProductEntryForSync to notify of a invalid product entry.
func NotifyInvalidProductEntryForSync(ctx context.Context, product *pb.Product, err error) {
	_, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("ProductController.SyncError").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	if product == nil {
		err := errors.New("product cannot be nil")
		logger.Error(err, "cannot notify for a invalid product entry with nil product")
		return
	}
	if err == nil {
		err := errors.New("err cannot be nil")
		logger.Error(err, "cannot notify for a invalid product entry with nil error", "productId", product.GetId())
		return
	}
	logger.Error(err, "product invalid", "productId", product.GetId())
}
