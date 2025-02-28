// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
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
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quota_management/database"
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
	ListenPort             uint16           `koanf:"listenPort"`
	Database               manageddb.Config `koanf:"database"`
	CloudaccountServerAddr string           `koanf:"cloudaccountServerAddr"`
	TestMode               bool
	SelectedRegion         string                 `koanf:"selectedRegion"`
	BootstrappedServices   []*BootstrappedService `koanf:"bootstrappedServices"`
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

	log.Info("initializing IDC Quota Management API service...")
	log.Info("debug info: Quota Management Service Region", logkeys.ServerAddr, cfg.SelectedRegion)
	log.Info("debug info: Quota Management Service Bootstrap", logkeys.ServerAddr, cfg.BootstrappedServices)

	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

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

	log.Info("successfully connected to the database")

	if err := db.AutoMigrateDB(ctx, svc.Mdb); err != nil {
		log.Error(err, "error migrating database")
		return err
	}

	log.Info("successfully migrated database model")

	// Connect to Cloudaccount Service
	cloudaccountClientConn := newClient(ctx, cfg.CloudaccountServerAddr, dialOptions...)
	// defer cloudaccountClientConn.Close()
	cloudAccountServiceClient := v1.NewCloudAccountServiceClient(cloudaccountClientConn)

	quotaManagementSrv, err := NewQuotaManagementServiceClient(ctx, sqlDB, cloudAccountServiceClient, cfg.SelectedRegion)
	if err != nil {
		return err
	}
	v1.RegisterQuotaManagementServiceServer(grpcServer, quotaManagementSrv)
	v1.RegisterQuotaManagementPrivateServiceServer(grpcServer, quotaManagementSrv)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	log.Info("quota management service started on port", logkeys.ListenPort, cfg.ListenPort)

	// bootstrap i.e register all services, with their resource types, quota units, max limits and default quotas (if present)
	err = quotaManagementSrv.BootstrapQuotaManagementService(ctx, cfg.BootstrappedServices, cfg.SelectedRegion)
	if err != nil {
		return err
	}

	return nil
}

func (*Service) Name() string {
	return "idc-quota-management-service"
}

func StartQuotaManagementService() {
	ctx := context.Background()
	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &Service{}, &Config{})
	if err != nil {
		logger := log.FromContext(ctx).WithName("StartQuotaManagementService")
		logger.Error(err, "init err")
	}
}
