// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"

	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

func main() {
	ctx := context.Background()
	// Initialize logger.
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("Billing Service")
	logger.Info("billing service start")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	if err := grpcutil.Run[*billing.Config](ctx, &billing.Service{}, billing.NewDefaultConfig()); err != nil {
		logger.Error(err, "startup failed")
	}
}
