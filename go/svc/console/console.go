// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/console"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func main() {
	cfg := grpcutil.ListenConfig{}
	ctx := context.Background()
	err := grpcutil.Run[*grpcutil.ListenConfig](ctx, &console.Service{}, &cfg)
	if err != nil {
		logger := log.FromContext(ctx).WithName("main")
		logger.Error(err, "init err")
	}
}
