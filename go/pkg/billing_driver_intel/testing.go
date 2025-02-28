// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing_driver_intel

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"

	meteringTests "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/tests"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TestService struct {
	Service
	testDB manageddb.TestDb
}

var (
	intelDriverConn *grpc.ClientConn
	Test            TestService
)

func (ts *TestService) Init(ctx context.Context, cfg *Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	var err error
	log := log.FromContext(ctx)
	ts.Mdb, err = ts.testDB.Start(ctx)

	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}
	if err := ts.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}

	if err := ts.Mdb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir); err != nil {
		log.Error(err, "error migrating database")
		return err
	}
	log.Info("successfully migrated database model")

	addr, err := resolver.Resolve(ctx, "billing-intel")
	if err != nil {
		return err
	}
	if intelDriverConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}

	return nil
}
func EmbedService(ctx context.Context) {
	meteringTests.EmbedService(ctx)
	authz.EmbedService(ctx)
	cloudaccount.EmbedService(ctx)
	grpcutil.AddTestService[*Config](&Test, &Config{})
	grpcutil.AddTestService[*grpcutil.ListenConfig](&TestCatalogService{}, &grpcutil.ListenConfig{})
}

type TestCatalogService struct {
}

type MockProductCatalogServer struct {
	pb.UnimplementedProductCatalogServiceServer
}

type MockProductVendorServer struct {
	pb.UnimplementedProductVendorServiceServer
}

func (srv *MockProductCatalogServer) Read(ctx context.Context, request *pb.ProductFilter) (*pb.ProductResponse, error) {
	log := log.FromContext(ctx).WithName("MockProductCatalogServer.Read")
	log.Info("ProductCatalogServer.Read", "req", request)
	return &pb.ProductResponse{}, nil
}

func (srv *MockProductVendorServer) Read(ctx context.Context, request *pb.VendorFilter) (*pb.VendorResponse, error) {
	log := log.FromContext(ctx).WithName("MockProductVendorServer.Read")
	log.Info("ProductVendorServer.Read", "req", request)
	return &pb.VendorResponse{}, nil
}

func (srv *MockProductCatalogServer) SetStatus(ctx context.Context, request *pb.SetProductStatusRequest) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("MockProductCatalogServer.Read")
	log.Info("ProductCatalogServer.SetStatus", "req", request)
	return &emptypb.Empty{}, nil
}

func (ts *TestCatalogService) Init(ctx context.Context, cfg *grpcutil.ListenConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	pb.RegisterProductCatalogServiceServer(grpcServer, &MockProductCatalogServer{})
	pb.RegisterProductVendorServiceServer(grpcServer, &MockProductVendorServer{})
	return nil
}

func (*TestCatalogService) Name() string {
	return "productcatalog"
}

func (ts *TestService) Done() {
	if err := ts.testDB.Stop(context.Background()); err != nil {
		panic(err)
	}
}
