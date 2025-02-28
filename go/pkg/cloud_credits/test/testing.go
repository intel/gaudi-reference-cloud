// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	aria "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	intel "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	standard "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_standard"
	credits "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	meteringTests "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/tests"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/usage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Used for testing
var SyncWait = atomic.Pointer[sync.WaitGroup]{}

var (
	testService                    *TestService
	clientConn                     *grpc.ClientConn
	TestSchedulerCloudAccountState *credits.SchedulerCloudAccountState
	TestCloudAccountSvcClient      *billingCommon.CloudAccountSvcClient
	TestUsageSvcClient             *billingCommon.UsageSvcClient
	meteringClientConn             *grpc.ClientConn
	usageConn                      *grpc.ClientConn
	cloudAccountConn               *grpc.ClientConn
	productClient                  billingCommon.ProductClientInterface
)

type TestService struct {
	credits.Service
	testDB     manageddb.TestDb
	clientConn *grpc.ClientConn
}

func (ts *TestService) Init(ctx context.Context, cfg *config.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	logger := log.FromContext(ctx).WithName("TestService.Init")
	cfg.InitTestConfig()
	var err error
	ts.Mdb, err = ts.testDB.Start(ctx)
	testService = ts
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}

	if err := ts.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}

	addr, err := resolver.Resolve(ctx, "cloudcredits")
	if err != nil {
		return err
	}
	if ts.clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	clientConn = ts.clientConn
	addr, err = resolver.Resolve(ctx, "metering")
	if err != nil {
		return err
	}
	meteringClientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	// product catalog client
	productClient, err = billingCommon.NewProductClient(ctx, resolver)
	if err != nil {
		return err
	}

	addr, err = resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return err
	}
	if cloudAccountConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}

	addr, err = resolver.Resolve(ctx, "usage")
	if err != nil {
		return err
	}
	if usageConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	TestUsageSvcClient, err = billingCommon.NewUsageServiceClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize usage service client")
		return err
	}

	TestCloudAccountSvcClient, err = billingCommon.NewCloudAccountClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize cloud account client")
		return err
	}
	TestSchedulerCloudAccountState = &credits.SchedulerCloudAccountState{
		AccessTimestamp: "",
	}
	return nil
}

type MockNotificationGatewayServiceServer struct {
	pb.UnimplementedNotificationGatewayServiceServer
}

func (MockNotificationGatewayServiceServer) PublishEvent(ctx context.Context, request *pb.PublishEventRequest) (*pb.PublishEventResponse, error) {
	log := log.FromContext(ctx).WithName("MockNotificationGatewayServiceServer.PublishEvent")
	log.Info("publish event", "req", request)
	return &pb.PublishEventResponse{}, nil
}
func (MockNotificationGatewayServiceServer) SubscribeEvents(ctx context.Context, request *pb.SubscribeEventRequest) (*pb.SubscribeEventResponse, error) {
	log := log.FromContext(ctx).WithName("MockNotificationGatewayServiceServer.SubscribeEvents")
	log.Info("subscribe event", "req", request)
	return &pb.SubscribeEventResponse{}, nil
}
func (MockNotificationGatewayServiceServer) ReceiveEvents(ctx context.Context, request *pb.ReceiveEventRequest) (*pb.ReceiveEventResponse, error) {
	log := log.FromContext(ctx).WithName("MockNotificationGatewayServiceServer.ReceiveEvents")
	log.Info("receive Event", "req", request)
	return &pb.ReceiveEventResponse{}, nil
}

type MockBillingCreditServiceServer struct {
}

func (MockBillingCreditServiceServer) Create(ctx context.Context, request *pb.BillingCredit) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("MockBillingCreditServiceServer.Create")
	log.Info("create", "req", request)
	return &emptypb.Empty{}, nil
}
func (MockBillingCreditServiceServer) ReadInternal(*pb.BillingAccount, pb.BillingCreditService_ReadInternalServer) error {
	return nil
}
func (MockBillingCreditServiceServer) Read(ctx context.Context, request *pb.BillingCreditFilter) (*pb.BillingCreditResponse, error) {
	log := log.FromContext(ctx).WithName("MockBillingCreditServiceServer.Read")
	log.Info("read", "req", request)
	return &pb.BillingCreditResponse{}, nil
}
func (MockBillingCreditServiceServer) ReadUnappliedCreditBalance(ctx context.Context, request *pb.BillingAccount) (*pb.BillingUnappliedCreditBalance, error) {
	log := log.FromContext(ctx).WithName("MockBillingCreditServiceServer.ReadUnappliedCreditBalance")
	log.Info("ReadUnappliedCreditBalance", "req", request)
	return &pb.BillingUnappliedCreditBalance{}, nil
}

type TestNotificationGatewayServiceServer struct {
}

func (ts *TestNotificationGatewayServiceServer) Init(ctx context.Context, cfg *grpcutil.ListenConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	pb.RegisterNotificationGatewayServiceServer(grpcServer, &MockNotificationGatewayServiceServer{})
	reflection.Register(grpcServer)
	return nil
}

func (*TestNotificationGatewayServiceServer) Name() string {
	return "notification-gateway"
}

func EmbedService(ctx context.Context) {
	aria.EmbedService(ctx)
	standard.EmbedService(ctx)
	intel.EmbedService(ctx)
	cloudaccount.EmbedService(ctx)
	billing_driver_intel.EmbedService(ctx)
	cloudaccount.EmbedService(ctx)
	meteringTests.EmbedService(ctx)
	usage.EmbedService(ctx)
	grpcutil.AddTestService[*grpcutil.ListenConfig](&TestNotificationGatewayServiceServer{}, &grpcutil.ListenConfig{})
	grpcutil.AddTestService[*config.Config](&TestService{}, &config.Config{TestProfile: true})
}

func (ts *TestService) Done() {
	grpcutil.ServiceDone[*config.Config](&ts.Service)
	err := ts.testDB.Stop(context.Background())
	if err != nil {
		panic(err)
	}
}
