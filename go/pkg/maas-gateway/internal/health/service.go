// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package health

import (
	"context"
	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/go-logr/logr"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/maas-gateway/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	health gosundheit.Health
	config *config.Config
	logger logr.Logger
}

func NewHealthService(config *config.Config, logger logr.Logger) *Service {
	return &Service{
		logger: logger.WithName("HealthService"),
		config: config,
	}
}

func (s *Service) SetupGrpcHealthCheck(
	healthServer *health.Server,
	usageClient pb.UsageRecordServiceClient,
	productClient pb.ProductCatalogServiceClient,
	dispatcherClient *grpc.ClientConn,
) error {
	s.logger.Info("setting up health checks")
	dispatcherHealthClient := grpc_health_v1.NewHealthClient(dispatcherClient)

	healthListener := NewGrpcHealthListener(healthServer)
	checkEventsLogger := NewCheckEventsLogger(s.logger)

	s.health = gosundheit.New(gosundheit.WithCheckListeners(checkEventsLogger), gosundheit.WithHealthListeners(healthListener))
	err := s.health.RegisterCheck(
		&checks.CustomCheck{
			CheckName: "grpc-health-check",
			CheckFunc: func(ctx context.Context) (details interface{}, err error) {
				if _, err := usageClient.Ping(ctx, &emptypb.Empty{}); err != nil {
					return nil, errors.Wrap(err, "unable to ping usage service")
				}

				if _, err := productClient.Ping(ctx, &emptypb.Empty{}); err != nil {
					return nil, errors.Wrap(err, "productCatalog service service not connected")
				}

				dispHealthResp, err := dispatcherHealthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
				if err != nil {
					return nil, errors.Wrap(err, "dispatcher not responsive during grpc health check request")
				}

				if dispHealthResp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
					return nil, errors.New("dispatcher service not in serving mode")
				}

				return "all the services are responsive", nil
			},
		},
		gosundheit.ExecutionPeriod(s.config.GrpcHealthCheckExecutionPeriod),
		gosundheit.ExecutionTimeout(s.config.GrpcHealthCheckExecutionTimeout),
		gosundheit.InitiallyPassing(true),
	)

	return errors.Wrap(err, "failed to register grpc health checks")
}
