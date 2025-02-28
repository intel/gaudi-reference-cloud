// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"

	credits "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	ctx := context.Background()

	// Init tracing
	tracerProvider := observability.New(ctx).InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*config.Config](ctx, &credits.Service{}, config.NewDefaultConfig())
	if err != nil {
		logger := log.FromContext(ctx).WithName("main")
		logger.Error(err, "init err")
	}
}
