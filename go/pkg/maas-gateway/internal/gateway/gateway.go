// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package gateway

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/adminserver"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/health"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/interceptors"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/metering"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/metrics"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type Gateway struct {
	config        *config.Config
	logger        logr.Logger
	grpcConnector client.Connector
	adminServer   *adminserver.HTTPServer
	grpcServer    *grpc.Server
	metrics       *metrics.PromMetrics
}

func NewGateway(config *config.Config, logger logr.Logger, connector client.Connector, metrics *metrics.PromMetrics) *Gateway {
	return &Gateway{
		config:        config,
		logger:        logger,
		grpcConnector: connector,
		metrics:       metrics,
	}
}

func (g *Gateway) Run(ctx context.Context, listener net.Listener) error {

	serviceClients := client.SetupServiceClients(g.config, g.logger, g.grpcConnector)
	defer serviceClients.Close()

	err := serviceClients.Connect(ctx)
	if err != nil {
		return errors.Wrap(err, "couldn't connect service clients")
	}

	dispatcherClient := serviceClients.DispatcherClient()
	productCatalogClient := serviceClients.ProductCatalogClient()
	usageRecordClient := serviceClients.UsageRecordClient()
	usageRecordGenerator := metering.NewRecordGenerator(g.logger, g.config, usageRecordClient.Client())

	grpcServer, err := g.setupServer(ctx, g.logger, productCatalogClient.Client())
	if err != nil {
		return errors.Wrap(err, "couldn't setup grpcServer")
	}
	g.grpcServer = grpcServer
	g.logger.Info("created grpc server")

	gatewayServer, err := NewServer(serviceClients.DispatcherClient().Client(), usageRecordGenerator, productCatalogClient.Client(), g.logger, g.config, g.metrics)
	pb.RegisterMaasGatewayServer(grpcServer, gatewayServer)
	g.logger.Info("registered Gateway server")

	healthServer := grpchealth.NewServer()
	healthService := health.NewHealthService(g.config, g.logger)
	healthgrpc.RegisterHealthServer(grpcServer, healthServer)
	g.logger.Info("registered Health server")

	err = healthService.SetupGrpcHealthCheck(
		healthServer,
		usageRecordClient.Client(),
		productCatalogClient.Client(),
		dispatcherClient.GrpcConnection(),
	)
	if err != nil {
		return errors.Wrap(err, "couldn't setup grpc health check")
	}

	reflection.Register(grpcServer)
	g.logger.Info("registered reflection")

	if err := grpcServer.Serve(listener); err != nil {
		return errors.Wrap(err, "couldn't start gRPC server")
	}

	return nil
}

func skipInterceptor(fullMethod string) bool {
	skippableMethods := []string{
		"/grpc.health.v1.Health/Check",
		"/proto.MaasGateway/GetSupportedModels",
	}

	for _, method := range skippableMethods {
		if fullMethod == method {
			return true
		}
	}
	return false
}

func (g *Gateway) setupServer(ctx context.Context, log logr.Logger, productCatalogClient pb.ProductCatalogServiceClient) (*grpc.Server, error) {

	metaDataInterceptors := interceptors.NewMetaDataInterceptors(log)
	validateInterceptors := interceptors.NewValidateInterceptors(log, productCatalogClient)

	streamInterceptors := grpc.ChainStreamInterceptor(
		interceptors.FilterStreamServerInterceptor(metaDataInterceptors.InjectRequestIdStreamServerInterceptor(), skipInterceptor),
		interceptors.FilterStreamServerInterceptor(validateInterceptors.ValidateStreamServerInterceptor(), skipInterceptor),
	)

	unaryInterceptors := grpc.ChainUnaryInterceptor(
		interceptors.FilterUnaryServerInterceptor(metaDataInterceptors.InjectRequestIdUnaryServerInterceptor(), skipInterceptor),
		interceptors.FilterUnaryServerInterceptor(validateInterceptors.ValidateUnaryServerInterceptor(), skipInterceptor),
	)

	serverOptions := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		streamInterceptors,
		unaryInterceptors,
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    120 * time.Second, // The time a connection is kept alive without any activity.
			Timeout: 20 * time.Second,  // Maximum time the server waits for activity before closing the connection.
		}),
	}

	grpcServer, err := grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't create new grpc server")
	}
	return grpcServer, nil
}

func (g *Gateway) Shutdown() {
	if g.grpcServer != nil {
		g.grpcServer.GracefulStop()
	}
}
