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
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
}

type Config struct {
	ListenPort             uint16 `koanf:"listenPort"`
	Interval               uint   `koanf:"serviceInterval"`
	Threshold              uint   `koanf:"cleanUpThreshold"`
	StorageAPIServerAddr   string `koanf:"storageAPIServerAddr"`
	NotiicationGatewayAddr string `koanf:"notificationAddr"`
	BillingServerAddr      string `koanf:"billingServerAddr"`
	ServiceEnabled         bool   `koanf:"serviceEnabled"`
	DryRun                 bool   `koanf:"dryRun"`
	SenderEmail            string `koanf:"senderEmail"`
	TemplateName           string `koanf:"templateName"`
	ConsoleUrl             string `koanf:"consoleUrl"`
	PaymentUrl             string `koanf:"paymentUrl"`

	TestMode bool
}

var (
	// AWS Cognito client for auth token generation, and required by global grpc-proxy
	cognitoEnabled, _ = strconv.ParseBool(os.Getenv("IDC_COGNITO_ENABLED"))
	cognitoURL, _     = url.Parse(os.Getenv("IDC_COGNITO_ENDPOINT"))
)

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

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

func (svc *Service) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx)
	log.Info("initializing IDC Storage Resource Cleaner...")

	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100 * 1024 * 1024)),
	}

	// Connect to Email Notification Service
	log.Info("notification gateway address", logkeys.ServerAddr, cfg.NotiicationGatewayAddr)
	emailClientConn := newClient(ctx, cfg.NotiicationGatewayAddr, dialOptions...)

	// Connect to billing Service
	log.Info("billing address", logkeys.ServerAddr, cfg.BillingServerAddr)
	billingClientConn := newClient(ctx, cfg.BillingServerAddr, dialOptions...)

	// Connect to Storage Service
	log.Info("api-server address", logkeys.ServerAddr, cfg.StorageAPIServerAddr)
	storageClientConn, err := grpcutil.NewClient(ctx, cfg.StorageAPIServerAddr, dialOptions...)
	if err != nil {
		log.Error(err, "error connecting to storage service")
		return err
	}
	emailClient := pb.NewEmailNotificationServiceClient(emailClientConn)
	billingDeactivateInstancesService := pb.NewBillingDeactivateInstancesServiceClient(billingClientConn)
	storageFSClient := pb.NewFilesystemPrivateServiceClient(storageClientConn)
	storageBKClient := pb.NewObjectStorageServicePrivateClient(storageClientConn)

	// Try to ping Storage Service
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := storageFSClient.PingPrivate(pingCtx, &emptypb.Empty{}); err != nil {
		log.Error(err, "unable to ping storage service")
		return err
	} else {
		log.Info("Ping FileStorage Service successful")
	}

	// Try to ping ObjectStorage Service
	pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := storageBKClient.PingPrivate(pingCtx, &emptypb.Empty{}); err != nil {
		log.Error(err, "unable to ping object storage service")
		return err
	} else {
		log.Info("Ping ObjectStorage Service successful")
	}

	// Try to ping billing Service
	pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := billingDeactivateInstancesService.Ping(pingCtx, &emptypb.Empty{}); err != nil {
		log.Error(err, "unable to ping billing service")
		return err
	} else {
		log.Info("Ping Billing Service successful")
	}

	// initialize service struct
	storageMonitor, err := NewStorageResourceCleaner(storageFSClient, storageBKClient, billingDeactivateInstancesService, emailClient, cfg)
	if err != nil {
		log.Error(err, "error starting storage resource cleaner")
		return err
	}
	log.Info("storage resource cleaner initialized")

	// Start a goroutine to trigger every interval (minutes)
	duration := time.Duration(cfg.Interval) * time.Minute
	go storageMonitor.StartAccountWatcher(ctx, duration)

	return nil
}

func (*Service) Name() string {
	return "idc-storage-resource-cleaner-service"
}

func StartStorageResourceCleaner() {
	ctx := context.Background()

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &Service{}, &Config{})
	if err != nil {
		logger := log.FromContext(ctx).WithName("StartStorageResourceCleaner")
		logger.Error(err, "init err")
	}
}
