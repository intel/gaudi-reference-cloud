// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_metering_monitor/metering_monitor"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

var (
	// AWS Cognito client for auth token generation, and required by global grpc-proxy
	cognitoEnabled, _ = strconv.ParseBool(os.Getenv("IDC_COGNITO_ENABLED"))
	cognitoURL, _     = url.Parse(os.Getenv("IDC_COGNITO_ENDPOINT"))
)

func main() {
	ctx := context.Background()

	var configFile string
	var metricsAddr string
	var probeAddr string
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8082", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	log.BindFlags()
	flag.Parse()

	//Initialize Logger
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("main")

	err := func() error {
		log.Info("Configuration file", logkeys.ConfigFile, configFile)

		scheme := runtime.NewScheme()
		if err := clientgoscheme.AddToScheme(scheme); err != nil {
			return err
		}
		if err := privatecloudv1alpha1.AddToScheme(scheme); err != nil {
			return err
		}

		cfg := &privatecloudv1alpha1.StorageMeteringMonitorConfig{}
		options := ctrl.Options{
			Scheme: scheme,
		}
		if configFile != "" {
			var err error
			options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(cfg))
			if err != nil {
				return fmt.Errorf("unable to load the config file: %w", err)
			}
		}

		log.Info("Configuration", logkeys.Configuration, cfg)

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		k8sManager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
		if err != nil {
			return fmt.Errorf("unable to start manager: %w", err)
		}

		if err := k8sManager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up health check: %w", err)
		}
		if err := k8sManager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up ready check: %w", err)
		}

		creds, err := grpcutil.GetClientCredentials(ctx)
		if err != nil {
			log.Error(err, "unable to create gRPC credentials")
			return fmt.Errorf("unable to create gRPC credentials: %w", err)
		}

		var clientConn *grpc.ClientConn
		clientOptions := []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		if cognitoEnabled {
			// create the cognitoClient to access AWS Cognito
			cognitoClient, err := authutil.NewCognitoClient(&authutil.CognitoConfig{
				URL:     cognitoURL,
				Timeout: 1 * time.Minute,
			})
			if err != nil {
				return fmt.Errorf("unable to create NewCognitoClient: %w", err)
			}

			// prefetch the access token to access global: metering svc
			_, err = cognitoClient.GetGlobalAuthToken(ctx)
			if err != nil {
				return fmt.Errorf("unable to GetGlobalAuthToken: %w", err)
			}

			clientOptions = append(clientOptions,
				grpc.WithPerRPCCredentials(authutil.NewCognitoAuth(ctx, cognitoClient)))
		}

		clientConn, err = grpcutil.NewClient(ctx, cfg.MeteringServerAddr, clientOptions...)
		if err != nil {
			return err
		}

		meteringClient := pb.NewMeteringServiceClient(clientConn)

		// Ensure that we can ping the metering service before starting the manager.
		pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := meteringClient.Ping(pingCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping metering service: %w", err)
		}

		//_, err = metering_monitor.NewMeteringMonitor(ctx, k8sManager, meteringClient, cfg.MaxUsageRecordSendInterval.Duration, cfg.Region)
		_, err = metering_monitor.NewMeteringMonitor(ctx, k8sManager, meteringClient, cfg)
		if err != nil {
			return fmt.Errorf("error creating metering monitor: %w", err)
		}

		log.Info("Starting Manager")
		if err := k8sManager.Start(ctrl.SetupSignalHandler()); err != nil {
			return fmt.Errorf("problem running manager: %w", err)
		}

		return nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error)
		os.Exit(1)
	}
}
