// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	controllers "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/vm/controllers"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
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

	// Initialize logger.
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
		if err := kubevirtv1.AddToScheme(scheme); err != nil {
			return err
		}

		cfg := &privatecloudv1alpha1.VmInstanceOperatorConfig{}
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
		if cfg.InstanceOperator.OperatorFeatureFlags.EnableQuickConnectClientCA && cfg.InstanceOperator.QuickConnectHost == "" {
			return fmt.Errorf("invalid config file: quickConnectHost must be specified when enableQuickConnectClientCA is enabled")
		}

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		if err := util.SetManagerOptions(ctx, &options, &cfg.InstanceOperator); err != nil {
			return err
		}
		k8sManager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
		if err != nil {
			return fmt.Errorf("unable to start manager: %w", err)
		}

		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		computeApiServerClientConn, err := grpcutil.NewClient(ctx, cfg.InstanceOperator.ComputeApiServerAddr, dialOptions...)
		if err != nil {
			return err
		}
		// Connect to VNet private Service client to Request/Release Subnets and addresses.
		vNetPrivateClient := pb.NewVNetPrivateServiceClient(computeApiServerClientConn)

		// Connect to VNet Service client to create/list/delete vNets.
		vNetClient := pb.NewVNetServiceClient(computeApiServerClientConn)

		// Ensure that we can ping the VNet service before starting the manager.
		// TODO: Allow the service to start without this ping. Use a health check to monitor this.
		pingPrivateCtx, cancelPrivate := context.WithTimeout(ctx, time.Second*10)
		defer cancelPrivate()
		if _, err := vNetPrivateClient.PingPrivate(pingPrivateCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping VNet Private service: %w", err)
		}
		pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := vNetClient.Ping(pingCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping VNet service: %w", err)
		}

		_, err = controllers.NewVmInstanceReconciler(ctx, k8sManager, vNetPrivateClient, vNetClient, cfg)
		if err != nil {
			return fmt.Errorf("error creating instance reconciler: %w", err)
		}

		if err := k8sManager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up health check: %w", err)
		}
		if err := k8sManager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up ready check: %w", err)
		}

		log.Info("Starting manager")
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
