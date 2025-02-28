// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/amacneil/dbmate/pkg/driver/postgres"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/client-go/rest"
)

var (
	// AWS Cognito client for auth token generation, and required by global grpc-proxy
	cognitoEnabled, _ = strconv.ParseBool(os.Getenv("IDC_COGNITO_ENABLED"))
	cognitoURL, _     = url.Parse(os.Getenv("IDC_COGNITO_ENDPOINT"))
)

func newClient(ctx context.Context, addr string, clientOptions ...grpc.DialOption) *grpc.ClientConn {

	var conn *grpc.ClientConn
	var err error

	logger := log.FromContext(ctx)

	if cognitoEnabled {
		// create the cognitoClient to access AWS Cognito
		cognitoClient, err := authutil.NewCognitoClient(&authutil.CognitoConfig{
			URL:     cognitoURL,
			Timeout: 1 * time.Minute,
		})
		if err != nil {
			logger.Error(err, "unable to read AWS Cognito credentials", "addr", addr)
			os.Exit(1)
		}

		// prefetch the access token to access global: cloudaccount svc
		token, err := cognitoClient.GetGlobalAuthToken(ctx)
		if err != nil {
			logger.Error(err, "unable to get AWS Cognito token", "addr", addr)
			os.Exit(1)
		}
		logger.V(9).Info("Prefetched Cognito Token", "cognitoToken", token)

		conn, err = grpcutil.NewClient(ctx, addr,
			grpc.WithPerRPCCredentials(authutil.NewCognitoAuth(ctx, cognitoClient)))
		if err != nil {
			logger.Error(err, "Not able to connect to gRPC service using grpcutil.NewClient", "addr", addr)
			os.Exit(1)
		}
	} else {
		conn, err = grpcutil.NewClient(ctx, addr)
		if err != nil {
			logger.Error(err, "Not able to connect to gRPC service using grpcutil.NewClient", "addr", addr)
			os.Exit(1)
		}
	}

	return conn
}

func main() {
	ctx := context.Background()

	// Parse command line.
	var configFile string
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	log.BindFlags()
	flag.Parse()

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	err := func() error {
		// Load configuration from file.
		log.Info("main", "configFile", configFile)
		var cfg config.Config
		if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
			return err
		}
		log.Info("main", "cfg", cfg)

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		// Load Kubernetes client configuration (currently not used).
		var defaultKubeRestConfig *rest.Config
		var err error
		if defaultKubeRestConfig, err = config.GetKubeRestConfig(); err != nil {
			defaultKubeRestConfig = &rest.Config{}
		}
		log.V(9).Info("main", "defaultKubeRestConfig", defaultKubeRestConfig)

		// Load database configuration.
		managedDb, err := manageddb.New(ctx, &cfg.Database)
		if err != nil {
			return err
		}

		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		// Connect to VM Instance Scheduler.
		clientConn, err := grpcutil.NewClient(ctx, cfg.VmInstanceSchedulerAddr, dialOptions...)
		if err != nil {
			return err
		}
		vmInstanceSchedulingService := pb.NewInstanceSchedulingServiceClient(clientConn)

		// Connect to Billing Service
		billingClientConn := newClient(ctx, cfg.BillingServerAddr, dialOptions...)
		defer billingClientConn.Close()
		billingDeactivateInstancesService := pb.NewBillingDeactivateInstancesServiceClient(billingClientConn)

		// Connect to Cloudaccount Service
		cloudaccountClientConn := newClient(ctx, cfg.CloudaccountServerAddr, dialOptions...)
		defer cloudaccountClientConn.Close()
		cloudAccountServiceClient := pb.NewCloudAccountServiceClient(cloudaccountClientConn)
		cloudAccountAppClientServiceClient := pb.NewCloudAccountAppClientServiceClient(cloudaccountClientConn)

		// Connect to Object Storage Service
		var objectStorageServicePrivateClient pb.ObjectStorageServicePrivateClient
		if cfg.ObjectStoragePrivateServerAddr != "" {
			objectStoragePrivateClientConn := newClient(ctx, cfg.ObjectStoragePrivateServerAddr, dialOptions...)
			defer objectStoragePrivateClientConn.Close()
			objectStorageServicePrivateClient = pb.NewObjectStorageServicePrivateClient(objectStoragePrivateClientConn)
		}

		// Connect to Fleet Admin Service
		var fleetAdminServiceClient pb.FleetAdminServiceClient
		if cfg.FleetAdminServerAddr != "" {
			fleetAdminClientConn := newClient(ctx, cfg.FleetAdminServerAddr, dialOptions...)
			defer fleetAdminClientConn.Close()
			fleetAdminServiceClient = pb.NewFleetAdminServiceClient(fleetAdminClientConn)
		}

		// Connect to Quota Management Service
		log.Info("Quota Management Service",
			"EnableQMSForQuotaProcessing", cfg.FeatureFlags.EnableQMSForQuotaProcessing,
			"QuotaManagementServerAddr", cfg.QuotaManagementServerAddr,
		)
		var qmsClient pb.QuotaManagementPrivateServiceClient
		if cfg.FeatureFlags.EnableQMSForQuotaProcessing && cfg.QuotaManagementServerAddr != "" {
			qmsClientConn := newClient(ctx, cfg.QuotaManagementServerAddr, dialOptions...)
			defer qmsClientConn.Close()
			qmsClient = pb.NewQuotaManagementPrivateServiceClient(qmsClientConn)
		}

		// Try to ping the VM Instance Scheduling service.
		// If this fails, the service will continue.
		pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := vmInstanceSchedulingService.Ping(pingCtx, &emptypb.Empty{}); err != nil {
			log.Error(err, "unable to ping VM Instance Scheduling service")
		} else {
			log.Info("Ping VM Instance Scheduling service successful")
		}

		// Try to ping Billing Deactivate Instances service. If this fails, the service will return with error.
		pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := billingDeactivateInstancesService.Ping(pingCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping Billing Deactivate Instances service: %w", err)
		} else {
			log.Info("Ping Billing Deactivate Instances service successful")
		}

		// Try to ping Cloudaccount Service service. If this fails, the service will return with error.
		pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := cloudAccountServiceClient.Ping(pingCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping Cloudaccount service: %w", err)
		} else {
			log.Info("Ping Cloudaccount service successful")
		}

		// Try to ping Object Storage service. If this fails, the service will continue.
		if objectStorageServicePrivateClient != nil {
			pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			if _, err := objectStorageServicePrivateClient.PingPrivate(pingCtx, &emptypb.Empty{}); err != nil {
				log.Error(err, "unable to ping Object Storage service")
			} else {
				log.Info("Ping Object Storage service successful")
			}
		}

		// Try to ping Fleet Admin service. If this fails, the service will continue.
		if fleetAdminServiceClient != nil {
			pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			if _, err := fleetAdminServiceClient.Ping(pingCtx, &emptypb.Empty{}); err != nil {
				log.Error(err, "unable to ping Fleet Admin service")
			} else {
				log.Info("Ping Fleet Admin service successful")
			}
		}

		// Start GRPC server.
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
		if err != nil {
			return fmt.Errorf("unable to listen on port %d: %w", cfg.ListenPort, err)
		}
		grpcService, err := server.New(
			ctx,
			&cfg,
			managedDb,
			vmInstanceSchedulingService,
			billingDeactivateInstancesService,
			cloudAccountServiceClient,
			cloudAccountAppClientServiceClient,
			objectStorageServicePrivateClient,
			fleetAdminServiceClient,
			qmsClient,
			listener,
		)
		if err != nil {
			return err
		}
		return grpcService.Run(ctx)
	}()
	if err != nil {
		log.Error(err, "fatal error")
		os.Exit(1)
	}
}
