// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s_resource_patcher/config"
	patcher "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s_resource_patcher/k8s_resource_patcher"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse command line.
	var configFile string
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	log.BindFlags()
	flag.Parse()

	// Initialize logger
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("main")

	err := func() error {
		// Load configuration from file.
		log.Info("main", logkeys.ConfigFile, configFile)
		var cfg config.Config
		if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
			return err
		}
		log.Info("main", logkeys.Configuration, cfg)

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		creds, err := grpcutil.GetClientCredentials(ctx)
		if err != nil {
			return fmt.Errorf("unable to create gRPC credentials: %w", err)
		}
		var clientConn *grpc.ClientConn
		clientOptions := []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		clientConn, err = grpcutil.NewClient(ctx, cfg.FleetAdminServerAddr, clientOptions...)
		if err != nil {
			return err
		}

		fleetAdminServiceClient := pb.NewFleetAdminServiceClient(clientConn)
		// Try to ping Fleet Admin service. If this fails, the service will return with error.
		pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := fleetAdminServiceClient.Ping(pingCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping Fleet Admin service: %w", err)
		}
		kubernetesResourcePatcher, err := patcher.NewKubernetesResourcePatcher(ctx, fleetAdminServiceClient, cfg)
		if err != nil {
			return err
		}

		kubernetesResourcePatcher.Start(ctx)
		<-ctx.Done()
		return nil

	}()
	if err != nil {
		log.Error(err, logkeys.Error)
		os.Exit(1)
	}
}
