// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package event

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway/config"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	clientConn *grpc.ClientConn
	test       Test
)

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*config.Config](&test, config.NewDefaultConfig())
}

type Test struct {
	Service
	testDb     manageddb.TestDb
	clientConn *grpc.ClientConn
}

func (test *Test) Init(ctx context.Context, cfg *config.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	var err error
	test.mdb, err = test.testDb.Start(ctx)
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}
	if err = test.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}
	addr, err := resolver.Resolve(ctx, "notification-gateway")
	if err != nil {
		return err
	}
	if test.clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	clientConn = test.clientConn
	return nil
}

func (test *Test) Done() {
	grpcutil.ServiceDone[*config.Config](&test.Service)
	err := test.testDb.Stop(context.Background())
	if err != nil {
		panic(err)
	}
}
