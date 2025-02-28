// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Almost all of the code for the cloud account service is here in
// pkg/cloudaccount. This facilitates embedding the cloud account
// service into go test programs such as cloudaccount_test.go. Other services
// that depend on the cloud account service can write their own
// go tests by embedding the cloud account service in the same
// way that cloudaccout_test.go does:
//
// import (
//
//		"testing"
//
//	   import "github.com/intel-innersource/frameworks.cloud.devcloud./services.idc/go/pkg/cloudaccount"
//
// )
//
// var test cloudaccount.Test
//
//	func TestMain(m *testing.M) {
//		test = cloudaccount.Test{}
//		test.Init()
//		defer test.Done()
//		m.Run()
//	}
//
// Other _test.go code in the same package can use the test.ClientConn method
// to make gRPC calls to the embedded cloud account service, like this:
//
//	client := pb.NewCloudAccountServiceClient(test.ClientConn())
//	_, err := client.Create(context.Background(), &acctCreate)
//
// Code that depends on cloud account should be written in such a way
// that the grpc.ClientConn for making calls to the cloud account
// service can be injected into the code for testing.
package cloudaccount

import (
	"context"
	"fmt"

	_ "embed"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"

	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Test struct {
	Service
	clientConn            *grpc.ClientConn
	clientConnAuthz       *grpc.ClientConn
	clientConnCredentials *grpc.ClientConn
	testDb                manageddb.TestDb
	cfg                   *config.Config
}

var test Test

func ClientConn() *grpc.ClientConn {
	return test.clientConn
}

func EmbedService(ctx context.Context) {
	events.EmbedService(ctx)
	grpcutil.AddTestService[*config.Config](&test, &config.Config{DisableEmail: true})
	grpcutil.AddTestService[*grpcutil.ListenConfig](&TestNotificationService{}, &grpcutil.ListenConfig{})
	grpcutil.AddTestService[*grpcutil.ListenConfig](&TestUserCredentialsService{}, &grpcutil.ListenConfig{})
}

func (test *Test) Init(ctx context.Context, cfg *config.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	var err error
	test.mdb, err = test.testDb.Start(ctx)
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}
	cfg.InitTestConfig()
	test.cfg = cfg

	addr, err := resolver.Resolve(ctx, "authz")
	if err != nil {
		return err
	}
	if test.clientConnAuthz, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}

	if err := test.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		return err
	}
	addr, err = resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return err
	}
	if test.clientConn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	addr, err = resolver.Resolve(ctx, "user-credentials")
	if err != nil {
		return err
	}
	if test.clientConnCredentials, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		return err
	}
	return nil
}

func (test *Test) Done() {
	grpcutil.ServiceDone[*config.Config](&test.Service)
	err := test.testDb.Stop(context.Background())
	if err != nil {
		panic(err)
	}
}

func (test *Test) ClientConn() *grpc.ClientConn {
	return test.clientConn
}

type MockNotificationGatewayServer struct {
	pb.UnimplementedEmailNotificationServiceServer
}

func (srv *MockNotificationGatewayServer) SendUserEmail(ctx context.Context, emailRequest *pb.EmailRequest) (*pb.EmailResponse, error) {
	log := log.FromContext(ctx).WithName("MockNotificationGatewayServer.SendInvitationEmail")
	log.Info("MockNotificationGatewayServer.SendInvitationEmail", "req", emailRequest)
	return &pb.EmailResponse{}, nil
}

type TestNotificationService struct {
}

func (ts *TestNotificationService) Init(ctx context.Context, cfg *grpcutil.ListenConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	pb.RegisterEmailNotificationServiceServer(grpcServer, &MockNotificationGatewayServer{})
	reflection.Register(grpcServer)
	return nil
}

func (*TestNotificationService) Name() string {
	return "notification-gateway"
}

type MockUserCredentialsServer struct {
	pb.UnimplementedUserCredentialsServiceServer
}

func (srv *MockUserCredentialsServer) RemoveMemberUserCredentials(ctx context.Context, request *pb.RemoveMemberUserCredentialsRequest) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("MockUserCredentialsServer.RemoveMemberUserCredentials")
	log.Info("MockUserCredentialsServer.RemoveMemberUserCredentials", "req", request)
	return &emptypb.Empty{}, nil
}

func (srv *MockUserCredentialsServer) GetUserCredentials(ctx context.Context, request *pb.GetUserCredentialRequest) (*pb.GetUserCredentialResponse, error) {
	log := log.FromContext(ctx).WithName("MockUserCredentialsServer.GetUserCredentials")
	log.Info("MockUserCredentialsServer.GetUserCredentials", "req", request)
	return &pb.GetUserCredentialResponse{}, nil
}
func (srv *MockUserCredentialsServer) CreateUserCredentials(ctx context.Context, request *pb.CreateUserCredentialsRequest) (*pb.ClientCredentials, error) {
	log := log.FromContext(ctx).WithName("MockUserCredentialsServer.CreateUserCredentials")
	log.Info("MockUserCredentialsServer.CreateUserCredentials", "req", request)
	return &pb.ClientCredentials{}, nil
}
func (srv *MockUserCredentialsServer) RemoveUserCredentials(ctx context.Context, request *pb.DeleteUserCredentialsRequest) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("MockUserCredentialsServer.RemoveUserCredentials")
	log.Info("MockUserCredentialsServer.RemoveUserCredentials", "req", request)
	return &emptypb.Empty{}, nil
}
func (srv *MockUserCredentialsServer) RevokeUserCredentials(ctx context.Context, request *pb.RevokeUserCredentialsRequest) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("MockUserCredentialsServer.RevokeUserCredentials")
	log.Info("MockUserCredentialsServer.RevokeUserCredentials", "req", request)
	return &emptypb.Empty{}, nil
}

func (srv *MockUserCredentialsServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type TestUserCredentialsService struct {
}

func (ts *TestUserCredentialsService) Init(ctx context.Context, cfg *grpcutil.ListenConfig, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	pb.RegisterUserCredentialsServiceServer(grpcServer, &MockUserCredentialsServer{})
	reflection.Register(grpcServer)
	return nil
}

func (*TestUserCredentialsService) Name() string {
	return "user-credentials"
}
