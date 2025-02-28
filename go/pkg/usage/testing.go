// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	meteringTests "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/tests"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	clientConn       *grpc.ClientConn
	cloudAccountConn *grpc.ClientConn
	meteringConn     *grpc.ClientConn
	test             Test
)

func EmbedService(ctx context.Context) {
	meteringTests.EmbedService(ctx)
	authz.EmbedService(ctx)
	cloudaccount.EmbedService(ctx)
	grpcutil.AddTestService[*grpcutil.ListenConfig](&TestCatalogService{}, &grpcutil.ListenConfig{})
	grpcutil.AddTestService[*Config](&test, NewDefaultConfig())
}

type Test struct {
	Service
	testDb     manageddb.TestDb
	clientConn *grpc.ClientConn
}

func (test *Test) Init(ctx context.Context, cfg *Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	var err error
	test.Mdb, err = test.testDb.Start(ctx)
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}
	if err = test.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}
	addr, err := resolver.Resolve(ctx, "usage")
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
	addr, err = resolver.Resolve(ctx, "metering")
	if err != nil {
		return err
	}
	if meteringConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	return nil
}

func (test *Test) Done() {
	grpcutil.ServiceDone[*Config](&test.Service)
	err := test.testDb.Stop(context.Background())
	if err != nil {
		panic(err)
	}
}

// Define a type for the mocking of the product catalog server.
type MockProductCatalogServer struct {
	pb.UnimplementedProductCatalogServiceServer
}

// Define a type for the mocking of the product vendor server.
type MockProductVendorServer struct {
	pb.UnimplementedProductVendorServiceServer
}

// Define the APIs for the product catalog server.
// This should take care of the filter.
// Because it is not, and the product client from billing common uses all account types - WOW! - we get multiple product entries.
// 5 Account types - 10 entries..
func (srv *MockProductCatalogServer) AdminRead(ctx context.Context, request *pb.ProductFilter) (*pb.ProductResponse, error) {
	log := log.FromContext(ctx).WithName("MockProductCatalogServer.AdminRead")
	log.Info("productCatalogServer.AdminRead", "req", request)
	products, _, _ := BuildStaticProductsAndVendors()
	return &pb.ProductResponse{Products: products}, nil
}

func (srv *MockProductCatalogServer) SetStatus(ctx context.Context, request *pb.SetProductStatusRequest) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("MockProductCatalogServer.Read")
	log.Info("ProductCatalogServer.SetStatus", "req", request)
	return &emptypb.Empty{}, nil
}

// Define the APIs for the product vendor server.
func (srv *MockProductVendorServer) Read(ctx context.Context, request *pb.VendorFilter) (*pb.VendorResponse, error) {
	log := log.FromContext(ctx).WithName("MockProductVendorServer.Read")
	log.Info("ProductVendorServer.Read", "req", request)
	_, vendors, _ := BuildStaticProductsAndVendors()
	return &pb.VendorResponse{Vendors: vendors}, nil
}

// Define the product catalog service.
type TestCatalogService struct {
}

func (ts *TestCatalogService) Init(ctx context.Context, cfg *grpcutil.ListenConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	pb.RegisterProductCatalogServiceServer(grpcServer, &MockProductCatalogServer{})
	pb.RegisterProductVendorServiceServer(grpcServer, &MockProductVendorServer{})
	reflection.Register(grpcServer)
	return nil
}

func (*TestCatalogService) Name() string {
	return "productcatalog"
}
