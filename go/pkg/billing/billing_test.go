// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"testing"

	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("BillingService.TestMain")
	logger.Info("\n\nembedding billing services\n\n")
	//EmbedService(ctx)
	//logger.Info("\n\nstarting billing test services\n\n")
	//grpcutil.StartTestServices(ctx)
	//defer grpcutil.StopTestServices()
	logger.Info("\n\nstarting billing test suite\n\n")
	//m.Run()
}
