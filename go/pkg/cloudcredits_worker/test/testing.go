// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	billc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	crditsvc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits/test"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	credits "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker"
	cfg "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

var (
	TestCloudAccountSvcClient     *billc.CloudAccountSvcClient
	TestNotificationGatewayClient *billc.NotificationGatewayClient
)

type TestService struct {
	credits.Worker
}

var Test TestService

func (test *TestService) Init(ctx context.Context, cfg *cfg.Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	var err error
	if err = test.Worker.Init(ctx, cfg, resolver); err != nil {
		return err
	}
	TestCloudAccountSvcClient, err = billc.NewCloudAccountClient(ctx, resolver)
	if err != nil {
		return err
	}
	TestNotificationGatewayClient, err = billc.NewNotificationGatewayClient(ctx, resolver)
	if err != nil {
		return err
	}
	return nil
}

func EmbedService(ctx context.Context) {
	authz.EmbedService(ctx)
	cloudaccount.EmbedService(ctx)
	crditsvc.EmbedService(ctx)
	grpcutil.AddTestService[*cfg.Config](&Test, &cfg.Config{TestProfile: true})
}

func (test *TestService) Done() error {
	// grpcutil.ServiceDone[*config.Config](&tesgt.Service)
	return nil
}
