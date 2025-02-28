// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/amacneil/dbmate/pkg/driver/postgres"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
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

	logger.Info("Cognito flag enabled: ", "Cognito Flag", cognitoEnabled)
	logger.Info("Cognito Url: ", "Cognito Url", cognitoURL)

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

		// Load database configuration.
		managedDbSo, err := manageddb.New(ctx, &cfg.Database)
		if err != nil {
			return err
		}

		cfg.Database.UsernameFile = cfg.UsernameRwFile
		cfg.Database.PasswordFile = cfg.PasswordRwFile
		managedDbRw, err := manageddb.New(ctx, &cfg.Database)
		if err != nil {
			return err
		}
		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		computeSvcGrpcConn, err := grpcutil.NewClient(ctx, cfg.ComputeServerAddr, dialOptions...)
		if err != nil {
			log.Error(err, "error", "Not able to connect to gRPC service using grpcutil.NewClient")
			return err
		}
		defer computeSvcGrpcConn.Close()

		productcatalogClientConn := newClient(ctx, cfg.ProductcatalogServerAddr, dialOptions...)

		// OTHER CLIENTS
		computeInstanceTypeSvcClient := pb.NewInstanceTypeServiceClient(computeSvcGrpcConn)
		computeInstanceSvcClient := pb.NewInstanceServiceClient(computeSvcGrpcConn)
		sshClient := pb.NewSshPublicKeyServiceClient(computeSvcGrpcConn)
		vnetClient := pb.NewVNetServiceClient(computeSvcGrpcConn)

		productcatalogServiceClient := pb.NewProductCatalogServiceClient(productcatalogClientConn)

		// Start GRPC server.
		grpcService, err := server.New(ctx,
			&cfg,
			managedDbRw,
			computeInstanceTypeSvcClient,
			sshClient,
			vnetClient,
			productcatalogServiceClient,
			computeInstanceSvcClient)
		if err != nil {
			return err
		}
		return grpcService.Run(ctx, managedDbSo)
	}()
	if err != nil {
		log.Error(err, "fatal error")
		os.Exit(1)
	}
}
