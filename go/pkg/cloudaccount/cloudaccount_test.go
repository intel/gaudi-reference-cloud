// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func TestMain(m *testing.M) {
	log.SetDefaultLogger()
	ctx := context.Background()
	EmbedService(ctx)
	authz.EmbedService(ctx)
	grpcutil.StartTestServices(ctx)
	defer grpcutil.StopTestServices()
	m.Run()
}
