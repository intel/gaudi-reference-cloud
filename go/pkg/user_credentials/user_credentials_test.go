// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package user_credentials

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/user_credentials/config"
)

var client pb.UserCredentialsServiceClient

func TestMain(m *testing.M) {
	log.SetDefaultLogger()
	config.Cfg.InitTestConfig()
	ctx := context.Background()
	//EmbedService(ctx)
	grpcutil.StartTestServices(ctx)
	defer grpcutil.StopTestServices()

	// Single client used for testing the APIs
	//client = pb.NewUserCredentialsServiceClient(test.clientConn)
	m.Run()
}
