// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authutil"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	qs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/api_server/pkg/server"
	server "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/api_server/pkg/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	// AWS Cognito client for auth token generation, and required by global grpc-proxy
	cognitoEnabled, _ = strconv.ParseBool(os.Getenv("IDC_COGNITO_ENABLED"))
	cognitoURL, _     = url.Parse(os.Getenv("IDC_COGNITO_ENDPOINT"))
)

func newClient(ctx context.Context, addr string, clientOptions ...grpc.DialOption) *grpc.ClientConn {

	var conn *grpc.ClientConn
	var err error

	logger := log.FromContext(ctx).WithName("newClient")

	if cognitoEnabled {
		// create the cognitoClient to access AWS Cognito
		cognitoClient, err := authutil.NewCognitoClient(&authutil.CognitoConfig{
			URL:     cognitoURL,
			Timeout: 1 * time.Minute,
		})
		if err != nil {
			logger.Error(err, "unable to read AWS Cognito credentials", logkeys.Address, addr)
			os.Exit(1)
		}

		// prefetch the access token to access global: cloudaccount svc
		_, err = cognitoClient.GetGlobalAuthToken(ctx)
		if err != nil {
			logger.Error(err, "unable to get AWS Cognito token", logkeys.Address, addr)
			os.Exit(1)
		}

		conn, err = grpcutil.NewClient(ctx, addr,
			grpc.WithPerRPCCredentials(authutil.NewCognitoAuth(ctx, cognitoClient)))
		if err != nil {
			logger.Error(err, "Not able to connect to gRPC service using grpcutil.NewClient", logkeys.Address, addr)
			os.Exit(1)
		}
	} else {
		conn, err = grpcutil.NewClient(ctx, addr)
		if err != nil {
			logger.Error(err, "Not able to connect to gRPC service using grpcutil.NewClient", logkeys.Address, addr)
			os.Exit(1)
		}
	}
	return conn
}

type Service struct {
	Mdb *manageddb.ManagedDb
}

type Config struct {
	ListenPort                uint16           `koanf:"listenPort"`
	Database                  manageddb.Config `koanf:"database"`
	TestMode                  bool
	CloudaccountServerAddr    string                   `koanf:"cloudaccountServerAddr"`
	ComputeAPIServerAddr      string                   `koanf:"computeAPIServerAddr"`
	SchedulerServerAddr       string                   `koanf:"storageSchedulerServerAddr"`
	CloudAccountQuota         server.CloudAccountQuota `koanf:"cloudAccountQuota"`
	StorageKmsAddr            string                   `koanf:"storageKmsServerAddr"`
	ProductcatalogServerAddr  string                   `koanf:"productcatalogServerAddr"`
	UserServerAddr            string                   `koanf:"storageUserServerAddr"`
	FilesystemServerAddr      string                   `koanf:"storageAPIServerAddr"`
	StorageControllerAddr     string                   `koanf:"storageControllerAddr"`
	StorageControllerUseMtls  bool                     `koanf:"storageControllerServerUseMtls"`
	ObjectStoreEnabled        bool                     `koanf:"objectStoreEnabled"`
	IksStorageEnabled         bool                     `koanf:"iksStorageEnabled"`
	DefaultBucketSizeInGB     string                   `koanf:"defaultBucketSizeInGB"`
	CustomQuotaMaxAllowedInTB int64                    `koanf:"customQuotaMaxAllowedInTB"`
	SelectedRegion            string                   `koanf:"selectedRegion"`
	MaxVolumesAllowed         int64                    `koanf:"maxVolumesAllowed"`
	MaxBucketsAllowed         int64                    `koanf:"maxBucketsAllowed"`
	QuotaManagementServerAddr string                   `koanf:"quotaManagementServerAddr"`
	QuotaManagementEnabled    bool                     `koanf:"quotaManagementEnabled"`
}

type CloudAccountQuota struct {
	CloudAccounts map[string]LaunchQuota `koanf:"cloudAccounts"`
}

type LaunchQuota struct {
	StorageQuota map[string]int `yaml:"storageQuota"`
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (svc *Service) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("Init")

	log.Info("initializing IDC Storage Admin API service...")

	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

	log.Info("config parsed ", logkeys.CloudAccountId, cfg.CloudAccountQuota)
	log.Info("debug info: Scheduler Server Address", logkeys.ServerAddr, cfg.SchedulerServerAddr)
	if svc.Mdb == nil {
		var err error
		svc.Mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			return err
		}
	}

	sqlDB, err := svc.Mdb.Open(ctx)
	if err != nil {
		return err
	}

	var quotaManagementClient v1.QuotaManagementPrivateServiceClient
	if cfg.QuotaManagementEnabled {
		// Connect to quota management service.
		quotaclientConn, err := grpcutil.NewClient(ctx, cfg.QuotaManagementServerAddr, dialOptions...)
		if err != nil {
			log.Error(err, "unable to obtain connection for quota management service", logkeys.ServerAddr, cfg.QuotaManagementServerAddr)
			return fmt.Errorf("quota management server grpc dial failed")
		}
		quotaManagementClient = pb.NewQuotaManagementPrivateServiceClient(quotaclientConn)
	}

	quotaService := qs.QuotaService{}
	err = quotaService.Init(ctx, sqlDB, cfg.CloudAccountQuota, quotaManagementClient)
	if err != nil {
		log.Error(err, "unable to  initialize quota service")
		return fmt.Errorf("unable to  initialize quota service")
	}

	// Connect to Cloudaccount Service
	cloudaccountClientConn := newClient(ctx, cfg.CloudaccountServerAddr, dialOptions...)
	// defer cloudaccountClientConn.Close()
	cloudAccountServiceClient := pb.NewCloudAccountServiceClient(cloudaccountClientConn)

	// connect to Filesystem Service
	filesystemClientConn, err := grpcutil.NewClient(ctx, cfg.FilesystemServerAddr)
	if err != nil {
		log.Error(err, "failed to connect to filesystem service")
		return err
	}
	filesystemServiceClient := pb.NewFilesystemPrivateServiceClient(filesystemClientConn)

	// init storage controller client
	strCntClient := storagecontroller.StorageControllerClient{}
	if err := strCntClient.Init(ctx, cfg.StorageControllerAddr, cfg.StorageControllerUseMtls); err != nil {
		log.Error(err, "error initializing storage controller client, exiting...")
		return fmt.Errorf("pre-conditioned failed")
	}
	log.Info("successfully initialized storage controller client ")

	filesystemSrv, err := NewStorageAdminServiceClient(ctx, sqlDB,
		cloudAccountServiceClient,
		&quotaService,
		filesystemServiceClient,
		&strCntClient, cfg.CustomQuotaMaxAllowedInTB, cfg.SelectedRegion,
		cfg.MaxVolumesAllowed, cfg.MaxBucketsAllowed)
	if err != nil {
		return err
	}
	v1.RegisterStorageAdminServiceServer(grpcServer, filesystemSrv)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	return nil
}

func (*Service) Name() string {
	return "idc-storage-admin-api-service"
}

func StartStorageAdminAPIService() {
	ctx := context.Background()
	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &Service{}, &Config{})
	if err != nil {
		logger := log.FromContext(ctx).WithName("StartStorageAdminAPIService")
		logger.Error(err, "init err")
	}
}
