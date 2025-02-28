// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/usage"
)

func main() {
	ctx := context.Background()

	// Initialize logger.
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("usage")
	logger.Info("usage service start")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	if err := grpcutil.Run[*usage.Config](ctx, &usage.Service{}, usage.NewDefaultConfig()); err != nil {
		logger.Error(err, "startup failure")
	}
}
