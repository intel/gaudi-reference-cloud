// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	notification_gateway "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

func main() {
	cfg := config.Config{}
	ctx := context.Background()
	// Initialize logger.
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("Notification-Gateway-Service")
	logger.Info("Notification Gateway service start")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*config.Config](ctx, &notification_gateway.Service{}, &cfg)
	logger.Error(err, "startup failed")
}
