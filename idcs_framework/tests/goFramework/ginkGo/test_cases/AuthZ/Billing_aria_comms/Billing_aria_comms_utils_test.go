package billing_aria_comms_test

import (
	"context"
	"encoding/json"

	aria "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/tests/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	meteringTests "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/tests"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ariaDriverClientConn *grpc.ClientConn
)

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
	const ROLE = "billing-aria"
	const DOMAIN = "billing-aria.idcs-system.svc.cluster.local"
	pb.RegisterProductCatalogServiceServer(grpcServer, &MockProductCatalogServer{})
	pb.RegisterProductVendorServiceServer(grpcServer, &MockProductVendorServer{})
	reflection.Register(grpcServer)
	var result any

	body := GenerateRequestToGetCertificate(DOMAIN, ROLE)
	jsonerr := json.Unmarshal([]byte(body), &result)
	if jsonerr != nil {
		Fail("Error during Unmarshal(): " + jsonerr.Error())
	}
	json := GetFieldInfo(body)

	_, clientTLSConf, err := VaultCertSetup(json["issuing_ca"].(string), json["private_key"].(string), json["certificate"].(string))
	test_config := credentials.NewTLS(clientTLSConf)
	if err != nil {
		Fail(err.Error())
	}

	addr, err := resolver.Resolve(ctx, "billing-aria")
	if err != nil {
		return err
	}

	if ariaDriverClientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(test_config)); err != nil {
		return err
	}
	return nil
}

func (*TestCatalogService) Name() string {
	return "productcatalog"
}

// Todo: once we have an Aria mock, we can get rid of skipAriaTests
// and SkipAriaTests(), because at that point the Aria billing driver
// should be able to function using the mock for testing.
var skipAriaTests = true

func SkipAriaTests() bool {
	return skipAriaTests
}

func EmbedService(ctx context.Context) {
	meteringTests.EmbedService(ctx)
	cloudaccount.EmbedService(ctx)
	log.SetDefaultLogger()
	logger := log.FromContext(context.Background())
	err := common.Init()
	if err != nil {
		logger.Info("init failed. skipping tests")
		skipAriaTests = true
	} else if config.Cfg.AriaSystem.AuthKey == "" {
		logger.Info("no auth key configured. skipping tests\n")
		skipAriaTests = true
	}
	cfg := config.NewDefaultConfig()
	cfg.ClientIdPrefix = config.GetTestPrefix()
	grpcutil.AddTestService[*config.Config](&aria.AriaService, cfg)
	grpcutil.AddTestService[*grpcutil.ListenConfig](&TestCatalogService{}, &grpcutil.ListenConfig{})
}
