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
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database"
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
	CloudaccountServerAddr    string            `koanf:"cloudaccountServerAddr"`
	ComputeAPIServerAddr      string            `koanf:"computeAPIServerAddr"`
	SchedulerServerAddr       string            `koanf:"storageSchedulerServerAddr"`
	CloudAccountQuota         CloudAccountQuota `koanf:"cloudAccountQuota"`
	StorageKmsAddr            string            `koanf:"storageKmsServerAddr"`
	ProductcatalogServerAddr  string            `koanf:"productcatalogServerAddr"`
	UserServerAddr            string            `koanf:"storageUserServerAddr"`
	QuotaManagementServerAddr string            `koanf:"quotaManagementServerAddr"`
	ObjectStoreEnabled        bool              `koanf:"objectStoreEnabled"`
	IksStorageEnabled         bool              `koanf:"iksStorageEnabled"`
	DefaultBucketSizeInGB     string            `koanf:"defaultBucketSizeInGB"`
	GeneralPurposeVASTEnabled bool              `koanf:"generalPurposeVASTEnabled"`
	QuotaManagementEnabled    bool              `koanf:"quotaManagementEnabled"`
	AuthzServerAddr           string            `koanf:"authzServerAddr"`
	AuthzEnabled              bool              `koanf:"authzEnabled"`
	CustomQuotaMaxAllowedInTB int64             `koanf:"customQuotaMaxAllowedInTB"`
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

	log.Info("initializing IDC Storage service...")

	if svc.Mdb == nil {
		var err error
		svc.Mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			return err
		}
	}

	log.Info("successfully connected to the database")

	if err := db.AutoMigrateDB(ctx, svc.Mdb); err != nil {
		log.Error(err, "error migrating database")
		return err
	}

	log.Info("successfully migrated database model")

	sqlDB, err := svc.Mdb.Open(ctx)
	if err != nil {
		return err
	}

	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

	log.Info("config parsed ", logkeys.CloudAccountId, cfg.CloudAccountQuota)
	log.Info("debug info: Scheduler Server Address", logkeys.ServerAddr, cfg.SchedulerServerAddr)

	// Connect to storage fiesystem Scheduler.
	clientConn, err := grpcutil.NewClient(ctx, cfg.SchedulerServerAddr, dialOptions...)
	if err != nil {
		log.Error(err, "unable to obtain connection for storage controller", logkeys.ServerAddr, cfg.SchedulerServerAddr)
		return fmt.Errorf("storage scheduler server grpc dial failed")
	}
	storageFileSchedClient := pb.NewFilesystemSchedulerPrivateServiceClient(clientConn)
	wekaAgentClient := pb.NewWekaStatefulAgentPrivateServiceClient(clientConn)

	// Connect to storage kms service.
	kmsclientConn, err := grpcutil.NewClient(ctx, cfg.StorageKmsAddr, dialOptions...)
	if err != nil {
		log.Error(err, "unable to obtain connection for storage kms", logkeys.ServerAddr, cfg.StorageKmsAddr)
		return fmt.Errorf("storage kms server grpc dial failed")
	}

	storageKmsClient := pb.NewStorageKMSPrivateServiceClient(kmsclientConn)

	// Connect to storage user service.
	userclientConn, err := grpcutil.NewClient(ctx, cfg.UserServerAddr, dialOptions...)
	if err != nil {
		log.Error(err, "unable to obtain connection for storage user service", logkeys.ServerAddr, cfg.UserServerAddr)
		return fmt.Errorf("storage user server grpc dial failed")
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

	storageUserClient := pb.NewFilesystemUserPrivateServiceClient(userclientConn)
	bucketUserClient := pb.NewBucketUserPrivateServiceClient(userclientConn)
	bucketlifecycleClient := pb.NewBucketLifecyclePrivateServiceClient(userclientConn)

	// Connect to Cloudaccount Service
	cloudaccountClientConn := newClient(ctx, cfg.CloudaccountServerAddr, dialOptions...)
	// defer cloudaccountClientConn.Close()
	cloudAccountServiceClient := pb.NewCloudAccountServiceClient(cloudaccountClientConn)

	// Connect to Productcatalog Service
	productcatalogClientConn := newClient(ctx, cfg.ProductcatalogServerAddr, dialOptions...)
	productcatalogServiceClient := pb.NewProductCatalogServiceClient(productcatalogClientConn)

	// Connect to authz Service
	authZClientConn := newClient(ctx, cfg.AuthzServerAddr, dialOptions...)
	authzServiceClient := pb.NewAuthzServiceClient(authZClientConn)

	quotaService := QuotaService{}
	err = quotaService.Init(ctx, sqlDB, cfg.CloudAccountQuota, quotaManagementClient)
	if err != nil {
		log.Error(err, "unable to  initialize quota service")
		return fmt.Errorf("unable to  initialize quota service")
	}

	filesystemSrv, err := NewFilesystemService(ctx, sqlDB,
		cloudAccountServiceClient,
		productcatalogServiceClient,
		authzServiceClient,
		storageFileSchedClient,
		storageKmsClient,
		&quotaService,
		storageUserClient,
		wekaAgentClient, cfg)
	if err != nil {
		return err
	}
	v1.RegisterFileStorageServiceServer(grpcServer, filesystemSrv)
	v1.RegisterFilesystemPrivateServiceServer(grpcServer, filesystemSrv)
	if cfg.IksStorageEnabled {
		v1.RegisterFilesystemOrgPrivateServiceServer(grpcServer, filesystemSrv)
		v1.RegisterFilesystemStorageClusterPrivateServiceServer(grpcServer, filesystemSrv)
	}
	if cfg.ObjectStoreEnabled {
		// Connect to ComputeAPI Service
		computeAPIClientConn := newClient(ctx, cfg.ComputeAPIServerAddr, dialOptions...)
		instanceServiceServiceClient := pb.NewInstancePrivateServiceClient(computeAPIClientConn)

		objectService, err := NewObjectService(ctx, sqlDB, cfg.DefaultBucketSizeInGB,
			cloudAccountServiceClient,
			productcatalogServiceClient,
			authzServiceClient,
			storageFileSchedClient,
			&quotaService,
			storageUserClient, bucketUserClient, bucketlifecycleClient,
			instanceServiceServiceClient, cfg)
		if err != nil {
			return err
		}
		v1.RegisterObjectStorageServiceServer(grpcServer, objectService)
		v1.RegisterObjectStorageServicePrivateServer(grpcServer, objectService)
	}

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)
	log.Info("filesystem service started on port", logkeys.ListenPort, cfg.ListenPort)

	// Start filesystem security group scheduler
	secGroupMgrService := NewSecGroupMgrService(sqlDB)
	go secGroupMgrService.StartSecurityGroupScanner(ctx)

	return nil
}

func (*Service) Name() string {
	return "idc-storage-service"
}

func StartStorageAPIService() {
	ctx := context.Background()
	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &Service{}, &Config{})
	if err != nil {
		logger := log.FromContext(ctx).WithName("StartStorageAPIService")
		logger.Error(err, "init err")
	}
}
