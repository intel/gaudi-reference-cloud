// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

func main() {
	ctx := context.Background()

	// Initialize logger.
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("billing-driver-intel")
	logger.Info("billing-driver-intel service start")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	if err := grpcutil.Run[*billing_driver_intel.Config](ctx, &billing_driver_intel.Service{}, billing_driver_intel.NewDefaultConfig()); err != nil {
		logger.Error(err, "startup failure")
	}
}
