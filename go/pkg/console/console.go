// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package console

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Service struct{}

var (
	billingInvoiceClient pb.BillingInvoiceServiceClient
)

func (*Service) Init(ctx context.Context, cfg *grpcutil.ListenConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	addr, err := resolver.Resolve(ctx, "billing")
	if err != nil {
		return err
	}

	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return err
	}

	billingInvoiceClient = pb.NewBillingInvoiceServiceClient(conn)

	pb.RegisterConsoleInvoiceServiceServer(grpcServer, &ConsoleInvoiceServer{})
	reflection.Register(grpcServer)
	return nil
}

func (*Service) Name() string {
	return "console"
}
