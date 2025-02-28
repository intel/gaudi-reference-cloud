// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/server"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	test Test
)

func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*server.Config](&test, &server.Config{
		TestMode:          true,
		ValidServiceTypes: []string{"ComputeAsAService", "FileStorageAsAService", "ObjectStorageAsAService"},
	})
}

type Test struct {
	server.Service
	testDb     manageddb.TestDb
	clientConn *grpc.ClientConn
}

func (test *Test) Init(ctx context.Context, cfg *server.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	var err error
	test.Mdb, err = test.testDb.Start(ctx)
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}
	if err = test.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}
	addr, err := resolver.Resolve(ctx, "metering")
	if err != nil {
		return err
	}
	if test.clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	return nil
}

func (test *Test) Done() {
	err := test.testDb.Stop(context.Background())
	if err != nil {
		panic(err)
	}
	grpcutil.ServiceDone[*server.Config](&test.Service)
}
