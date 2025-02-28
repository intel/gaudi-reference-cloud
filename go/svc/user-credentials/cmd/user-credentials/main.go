// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	user_credentials "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/user_credentials"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/user_credentials/config"
)

func main() {
	cfg := config.Config{}
	ctx := context.Background()
	// Initialize logger.
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("User-Credentials-Service")
	logger.Info("User Credentials service start")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*config.Config](ctx, &user_credentials.Service{}, &cfg)
	logger.Error(err, "startup failed")
}
