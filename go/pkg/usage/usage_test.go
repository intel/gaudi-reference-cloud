// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var client pb.UsageRecordServiceClient

func TestMain(m *testing.M) {
	log.SetDefaultLogger()
	Cfg.InitTestConfig()
	ctx := context.Background()
	EmbedService(ctx)
	grpcutil.StartTestServices(ctx)
	defer grpcutil.StopTestServices()

	// Single client used for testing the APIs
	client = pb.NewUsageRecordServiceClient(test.clientConn)
	m.Run()
}
