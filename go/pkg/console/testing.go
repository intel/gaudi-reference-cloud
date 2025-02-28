// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package console

import (
	"context"

	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TestService struct {
	Service
}

var (
	invoiceClient pb.ConsoleInvoiceServiceClient
)

func (ts *TestService) Init(ctx context.Context, cfg *grpcutil.ListenConfig,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	if err := ts.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}

	addr, err := resolver.Resolve(ctx, "console")
	if err != nil {
		return err
	}

	clientConn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	invoiceClient = pb.NewConsoleInvoiceServiceClient(clientConn)
	return nil
}

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*grpcutil.ListenConfig](&TestService{}, &grpcutil.ListenConfig{})
	billing.EmbedService(ctx)
}
