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
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/server"
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
			logger.Error(err, "Not able to connect to gRPC service using grpcutil.NewClient with Cognito", "addr", addr)
			os.Exit(1)
		}
	} else {
		conn, err = grpcutil.NewClient(ctx, addr, clientOptions...)
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
		log.Info("main", logkeys.ConfigFile, configFile)
		var cfg config.Config
		if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
			return err
		}
		log.Info("main", logkeys.Configuration, cfg)

		if len(cfg.AvailabilityZones) == 0 {
			return fmt.Errorf("missing availability zones configuration")
		}

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		// Load database configuration.
		managedDb, err := manageddb.New(ctx, &cfg.Database)
		if err != nil {
			return err
		}

		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		// Connect to Cloudaccount Service
		cloudaccountClientConn := newClient(ctx, cfg.CloudaccountServerAddr, dialOptions...)
		defer cloudaccountClientConn.Close()
		cloudAccountServiceClient := pb.NewCloudAccountServiceClient(cloudaccountClientConn)

		// Start GRPC server.
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
		if err != nil {
			return fmt.Errorf("unable to listen on port %d: %w", cfg.ListenPort, err)
		}
		grpcService, err := server.New(ctx, &cfg, managedDb, listener, cloudAccountServiceClient, cfg.AvailabilityZones)
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
