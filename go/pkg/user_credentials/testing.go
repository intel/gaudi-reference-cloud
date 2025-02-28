// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package user_credentials

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/user_credentials/config"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	clientConn       *grpc.ClientConn
	cloudAccountConn *grpc.ClientConn
	test             Test
)

func EmbedService(ctx context.Context) {
	cloudaccount.EmbedService(ctx)
	grpcutil.AddTestService[*config.Config](&test, config.NewDefaultConfig())
}

type Test struct {
	Service
	clientConn *grpc.ClientConn
}

func (test *Test) Init(ctx context.Context, cfg *config.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	var err error
	if err = test.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}
	addr, err := resolver.Resolve(ctx, "user-credentials")
	if err != nil {
		return err
	}
	if test.clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	clientConn = test.clientConn
	addr, err = resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return err
	}
	if cloudAccountConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}

	return nil
}

func (test *Test) Done() {
	grpcutil.ServiceDone[*config.Config](&test.Service)

}
