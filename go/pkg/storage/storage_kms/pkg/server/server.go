// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/secrets"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type StorageKMSService struct {
}

type Config struct {
	ListenPort            uint16 `koanf:"listenPort"`
	StorageControllerAddr string `koanf:"storageControllerAddr"`
	TestMode              bool
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (*StorageKMSService) Name() string {
	return "idc-storage-kms-service"
}

func (svc *StorageKMSService) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("StorageKMSService.Init")

	log.Info("initializing IDC Storage KMS service...")

	vaultClient := secrets.NewVaultClient(ctx)
	err := vaultClient.ValidateVaultClient(ctx)
	if err != nil {
		log.Info("unable to initialize Vault client")
		log.Error(err, "ValidateVaultClient")
	}
	kmsSrv, err := NewStorageKMSService(vaultClient)
	if err != nil {
		return err
	}

	pb.RegisterStorageKMSPrivateServiceServer(grpcServer, kmsSrv)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	log.Info("storage kms service started", logkeys.ListenPort, cfg.ListenPort)

	return nil
}

func StartStorageKMSService() {
	ctx := context.Background()
	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &StorageKMSService{}, &Config{})
	if err != nil {
		log.FromContext(ctx).WithName("StartStorageKMSService").Error(err, "init err")
	}
}
