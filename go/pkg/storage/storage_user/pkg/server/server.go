// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"

	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type UserService struct {
	bucketAPIClient pb.ObjectStorageServicePrivateClient
	strCntClient    storagecontroller.StorageControllerClient
}

type Config struct {
	ListenPort                    uint16 `koanf:"listenPort"`
	StorageControllerAddr         string `koanf:"storageControllerAddr"`
	StorageControllerUseMtls      bool   `koanf:"storageControllerServerUseMtls"`
	StorageKmsAddr                string `koanf:"storageKmsServerAddr"`
	ObjectStoreEnabled            bool   `koanf:"objectStoreEnabled"`
	StorageAPIServerAddr          string `koanf:"storageAPIServerAddr"`
	ScheduleSubnetMonitorInterval uint16 `koanf:"scheduleSubnetMonitorInterval"`
	TestMode                      bool
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (*UserService) Name() string {
	return "idc-storage-user-service"
}

func (svc *UserService) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("UserService.Init")

	log.Info("initializing IDC Storage User service...")

	strCntCli := storagecontroller.StorageControllerClient{}
	if err := strCntCli.Init(ctx, cfg.StorageControllerAddr, cfg.StorageControllerUseMtls); err != nil {
		log.Error(err, "error initializing storage controller client, exiting...")
		return fmt.Errorf("pre-conditioned failed")
	}
	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

	log.Info("debug:storageKmsAddr", logkeys.ServerAddr, cfg.StorageKmsAddr)
	// Connect to storage kms service.
	kmsclientConn, err := grpcutil.NewClient(ctx, cfg.StorageKmsAddr, dialOptions...)
	if err != nil {
		log.Error(err, "unable to obtain connection for storage kms", logkeys.ServerAddr, cfg.StorageKmsAddr)
		return fmt.Errorf("storage controller server grpc dial failed")
	}

	stoargeKmsClient := pb.NewStorageKMSPrivateServiceClient(kmsclientConn)

	storageAPIConn, err := grpcutil.NewClient(ctx, cfg.StorageAPIServerAddr)
	if err != nil {
		log.Error(err, "error creating storageServiceClient")
		return fmt.Errorf("error initializing storagr replicator service")
	}
	svc.bucketAPIClient = pb.NewObjectStorageServicePrivateClient(storageAPIConn)
	svc.strCntClient = strCntCli

	userSrv, err := NewStorageUserServiceServer(stoargeKmsClient, &strCntCli)
	if err != nil {
		return err
	}

	lifecycleSrv, err := NewBucketLifecycleServiceServer(&strCntCli)
	if err != nil {
		return err
	}

	pb.RegisterFilesystemUserPrivateServiceServer(grpcServer, userSrv)
	if cfg.ObjectStoreEnabled {
		pb.RegisterBucketUserPrivateServiceServer(grpcServer, userSrv)
		pb.RegisterBucketLifecyclePrivateServiceServer(grpcServer, lifecycleSrv)
	}

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	go svc.bucketSubnetUpdateHandler(ctx, cfg.ScheduleSubnetMonitorInterval)

	log.Info("storage user service started", logkeys.ListenPort, cfg.ListenPort)

	return nil
}

func (svc *UserService) bucketSubnetUpdateHandler(ctx context.Context, interval uint16) {
	log := log.FromContext(ctx).WithName("UserService.bucketSubnetUpdateHandler")

	// Loop through periodically for every 'handleSubnetUpdate' of time to obtain
	// subnets that have been updated and handle principals for those
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := svc.HandlePrincipalSecurityGroupUpdate(ctx)
			if err != nil {
				log.Error(err, "Error occured when handling security group for buckets")
				continue
			}
		}
	}
}

func StartStorageUserService() {
	ctx := context.Background()
	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &UserService{}, &Config{})
	if err != nil {
		log.FromContext(ctx).WithName("StartStorageUserService").Error(err, "init err")
	}
}
