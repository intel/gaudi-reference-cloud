// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	meteringTests "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/tests"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var (
	meteringClientConn *grpc.ClientConn
)

type Service struct{}

func (*Service) Init(ctx context.Context, cfg *grpcutil.ListenConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	reflection.Register(grpcServer)
	return nil
}

type TestService struct {
	Service
}

func (testService *TestService) Init(ctx context.Context, cfg *grpcutil.ListenConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	if err := testService.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}
	addr, err := resolver.Resolve(ctx, "metering")
	if err != nil {
		return err
	}
	meteringClientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	return nil
}

func (*TestService) Name() string {
	return "metering-test-service"
}

func (*TestService) Done() {
}

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*grpcutil.ListenConfig](&TestService{}, &grpcutil.ListenConfig{})
	meteringTests.EmbedService(ctx)
}
