// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package gateway

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/test/mock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"io"
	"net"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	ctx                context.Context
	logger             logr.Logger
	gateway            *Gateway
	usageRecordServer  *mock.UsageRecordServer
	mockConnector      *client.MockGrpcServiceConnector
	gatewayLN          *bufconn.Listener
	gatewayClient      pb.MaasGatewayClient
	predefinedProducts []*pb.Product
}

func (s *TestSuite) SetupSuite() {
	s.T().Setenv("IDC_SERVER_TLS_ENABLED", "false")
	s.T().Setenv("IDC_CLIENT_TLS_ENABLED", "false")

	log.BindFlags()
	log.SetDefaultLogger()
	s.ctx = context.Background()

	s.logger = log.FromContext(s.ctx).WithName(ServiceName + "_tests")
	s.logger.Info("performing SetupSuite")

	//create config
	cfg := config.Config{
		UsageServerTimeout:             10 * time.Millisecond,
		GrpcHealthCheckExecutionPeriod: 30 * time.Second,
		RetryAttempts:                  3,
	}

	mockServer := mock.NewGrpcServer()
	dispatcherServer := mock.NewDispatcherServer()

	s.predefinedProducts = []*pb.Product{
		{
			Name: "test123",
			Id:   "123",
			Metadata: map[string]string{
				"hfModelName": "model_test_123",
			},
		},
		{
			Name: "test456",
			Id:   "456",
			Metadata: map[string]string{
				"hfModelName": "model_test_456",
			},
		},
	}

	productCatalogServer := mock.NewProductCatalogServerWithProducts(s.predefinedProducts)
	s.usageRecordServer = mock.NewUsageRecordServer()
	healthServer := mock.NewHealthServer()

	grpcMockServer := mockServer.GetGrpcServer()
	pb.RegisterDispatcherServer(grpcMockServer, dispatcherServer)
	pb.RegisterUsageRecordServiceServer(grpcMockServer, s.usageRecordServer)
	pb.RegisterProductCatalogServiceServer(grpcMockServer, productCatalogServer)
	grpc_health_v1.RegisterHealthServer(grpcMockServer, healthServer)
	connection, err := mockServer.Run()
	if err != nil {
		s.T().Fatalf("couldn't create mock server: %v", err)
	}

	s.mockConnector = client.NewMockGrpcServiceConnector(connection)

	metricSDK := metrics.NewPromMetrics(s.logger, ServiceName, MetricsPrefix)

	s.gateway = NewGateway(&cfg, s.logger, s.mockConnector, metricSDK)

	s.gatewayLN = bufconn.Listen(1024)
	go func() {
		err = s.gateway.Run(s.ctx, s.gatewayLN)
		if err != nil {
			s.T().Fatalf("couldn't run gateway: %v", err)
		}
	}()

	conn := createClientConnection(s.ctx, s.T(), s.gatewayLN)
	s.gatewayClient = pb.NewMaasGatewayClient(conn)
}

func (s *TestSuite) TearDownSuite() {
	// This will run once after all tests in the suite
	s.logger.Info("performing TearDownSuite")
	if err := s.gatewayLN.Close(); err != nil {
		s.T().Errorf("couldn't close gateway listener: %v", err)
	}
	s.gateway.Shutdown()
}

func (s *TestSuite) Test_Gateway_GetSupportedModels() {
	s.logger.Info("Test_Gateway_GetSupportedModels")

	models, err := s.gatewayClient.GetSupportedModels(s.ctx, &empty.Empty{})
	if err != nil {
		s.T().Errorf("couldn't get supported models: %v", err)
	}

	s.NoError(err)
	s.NotNil(models)

	s.Len(models.Models, 2)

	for index, _ := range []int{0, 1} {
		s.Equal(s.predefinedProducts[index].Metadata["hfModelName"], models.Models[index].ModelName)
		s.Equal(s.predefinedProducts[index].Id, models.Models[index].ProductId)
		s.Equal(s.predefinedProducts[index].Name, models.Models[index].ProductName)
	}

	s.T().Logf("models: %v", models)
}

func (s *TestSuite) Test_Gateway_GenerateStream() {
	s.logger.Info("Test_Gateway_GetSupportedModels")

	maasRequest := &pb.MaasRequest{
		Model: "model_test_123",
		Request: &pb.MaasGenerateStreamRequest{
			Prompt: "Hello World",
		},
	}

	tests := []struct {
		name               string
		request            *pb.MaasRequest
		usageRecordError   error
		usageRecordTimeout time.Duration
	}{
		{
			name:             "basic prompt",
			request:          maasRequest,
			usageRecordError: nil,
		},
		{
			name:               "usage record error",
			request:            maasRequest,
			usageRecordError:   status.Error(codes.Internal, "usage record error"),
			usageRecordTimeout: 40 * time.Millisecond,
		},
	}

	for _, test := range tests {
		if test.usageRecordError != nil {
			s.usageRecordServer.SetUsageRecordError(test.usageRecordError, test.usageRecordTimeout)
		}

		stream, err := s.gatewayClient.GenerateStream(s.ctx, test.request)
		require.NoError(s.T(), err)
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(s.T(), err)
			s.T().Logf("Received response: %v", response)
		}
	}
}

func TestFetcherSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func createClientConnection(ctx context.Context, t *testing.T, ln *bufconn.Listener) *grpc.ClientConn {
	addr := "bufconn"
	clientConn, err := grpcutil.NewClient(ctx, addr, grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return ln.DialContext(ctx)
	}))
	require.NoError(t, err)

	return clientConn
}
