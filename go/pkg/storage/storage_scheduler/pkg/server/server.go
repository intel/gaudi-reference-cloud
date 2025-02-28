// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type SchedulerService struct {
}

type Config struct {
	ListenPort                uint16 `koanf:"listenPort"`
	StorageControllerAddr     string `koanf:"storageControllerAddr"`
	StorageControllerUseMtls  bool   `koanf:"storageControllerServerUseMtls"`
	StorageKmsAddr            string `koanf:"storageKmsServerAddr"`
	GeneralPurposeVASTEnabled bool   `koanf:"generalPurposeVASTEnabled"`

	TestMode bool
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (*SchedulerService) Name() string {
	return "idc-storage-scheduler-service"
}

func (svc *SchedulerService) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("SchedulerService.Init")
	log.Info("initializing IDC Storage Scheduler service...")

	strCntCli := storagecontroller.StorageControllerClient{}
	if err := strCntCli.Init(ctx, cfg.StorageControllerAddr, cfg.StorageControllerUseMtls); err != nil {
		log.Error(err, "error initializing storage controller client, exiting...")
		return fmt.Errorf("pre-conditioned failed")
	}
	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}
	kmsclientConn, err := grpcutil.NewClient(ctx, cfg.StorageKmsAddr, dialOptions...)
	if err != nil {
		log.Error(err, "unable to obtain connection for storage kms", logkeys.ServerAddr, cfg.StorageKmsAddr)
		return fmt.Errorf("storage controller server grpc dial failed")
	}

	storageKmsClient := pb.NewStorageKMSPrivateServiceClient(kmsclientConn)

	schedSrv, err := NewStorageSchedulerService(&strCntCli, storageKmsClient, cfg.GeneralPurposeVASTEnabled)
	if err != nil {
		return err
	}

	pb.RegisterFilesystemSchedulerPrivateServiceServer(grpcServer, schedSrv)
	pb.RegisterWekaStatefulAgentPrivateServiceServer(grpcServer, schedSrv)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	log.Info("storage scheduler service started", logkeys.ListenPort, cfg.ListenPort)

	return nil
}

func StartStorageSchedulerService() {
	ctx := context.Background()
	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &SchedulerService{}, &Config{})
	if err != nil {
		logger := log.FromContext(ctx).WithName("StartStorageSchedulerService")
		logger.Error(err, "init err")
	}
}
