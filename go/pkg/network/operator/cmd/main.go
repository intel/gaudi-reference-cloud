// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	vpcv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/api/v1alpha1"
	inputconfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/internal/controller/iprm"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/internal/controller/subnet"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/operator/internal/controller/vpc"
	sdnv1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/sdn"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(vpcv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	ctx := context.Background()

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var operatorConfig string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8078", "The address the metric endpoint binds.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8079", "The address the probe endpoint binds.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&operatorConfig, "operator-config", "/config", "This is the path to the operator config.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("main")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	// Read operator config.
	viper.AddConfigPath(operatorConfig)
	if err := viper.ReadInConfig(); err != nil {
		setupLog.Error(err, "read operator config")
		os.Exit(1)
	}

	var inputConfig inputconfig.NetworkProviderConfig
	if err := viper.Unmarshal(&inputConfig); err != nil {
		setupLog.Error(err, "unmarshal operator config")
		os.Exit(1)
	}

	err := func() error {

		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			Metrics:                metricsserver.Options{BindAddress: metricsAddr},
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         false, // enableLeaderElection,
			LeaderElectionID:       "95f548c04.intel.com",
			// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
			// when the Manager ends. This requires the binary to immediately end when the
			// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
			// speeds up voluntary leader transitions as the new leader don't have to wait
			// LeaseDuration time first.
			//
			// In the default scaffold provided, the program ends immediately after
			// the manager stops, so would be fine to enable this option. However,
			// if you are doing or is intended to do any operation such as perform cleanups
			// after the manager stops then its usage might be unsafe.
			// LeaderElectionReleaseOnCancel: true,

		})
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}

		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		// Create connection to Network API
		networkClientConn, err := grpcutil.NewClient(ctx, inputConfig.NetworkAPIServerAddr, dialOptions...)
		if err != nil {
			return err
		}

		defer networkClientConn.Close()
		vpcServiceClient := pb.NewVPCPrivateServiceClient(networkClientConn)
		subnetServiceClient := pb.NewSubnetPrivateServiceClient(networkClientConn)
		iprmServiceClient := pb.NewIPRMPrivateServiceClient(networkClientConn)

		// Ensure that we can ping the network service before starting the manager.
		pingNetworkCtx, cancelNetwork := context.WithTimeout(ctx, time.Second*10)
		defer cancelNetwork()
		if _, err := vpcServiceClient.PingPrivate(pingNetworkCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping vpc service: %w", err)
		}

		if _, err := subnetServiceClient.PingPrivate(pingNetworkCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping subnet service: %w", err)
		}

		if _, err := iprmServiceClient.PingPrivate(pingNetworkCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping iprm service: %w", err)
		}

		log.Info("sdnserver", "address", inputConfig.SDNServerAddr)

		// Create connection to SDN Controller
		sdnClientConn, err := grpc.Dial(inputConfig.SDNServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials())) // TODO: Only for local dev!
		if err != nil {
			log.Error(err, "did not connect")
			os.Exit(1)
		}
		defer sdnClientConn.Close()
		sdnClient := sdnv1.NewOvnnetClient(sdnClientConn)

		_, err = vpc.NewReconciler(ctx, mgr, vpcServiceClient, sdnClient, inputConfig.Region)
		if err != nil {
			log.Error(err, "could not init vpc reconciler")
			os.Exit(1)
		}

		_, err = subnet.NewReconciler(ctx, mgr, subnetServiceClient, sdnClient)
		if err != nil {
			log.Error(err, "could not init subnet reconciler")
			os.Exit(1)
		}

		_, err = iprm.NewReconciler(ctx, mgr, iprmServiceClient, sdnClient)
		if err != nil {
			log.Error(err, "could not init iprm reconciler")
			os.Exit(1)
		}

		setupLog.Info("initializing controller", "MaxConcurrentReconciles", inputConfig.MaxConcurrentReconciles)

		if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			setupLog.Error(err, "unable to set up health check")
			os.Exit(1)
		}
		if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			setupLog.Error(err, "unable to set up ready check")
			os.Exit(1)
		}

		setupLog.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
		return nil
	}()

	if err != nil {
		log.Error(err, logkeys.Error)
		os.Exit(1)
	}
}
