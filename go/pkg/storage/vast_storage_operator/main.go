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
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	cloud "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/vast_storage_operator/controllers"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	cloudclient "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/clientset/versioned"
	idcinformerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/informers/externalversions"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

type Config struct {
	ListenPort                     uint16 `koanf:"listenPort"`
	StorageControllerServerAddr    string `koanf:"storageControllerServerAddr"`
	StorageControllerServerUseMtls string `koanf:"storageControllerServerUseMtls"`
	TestMode                       bool
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(cloudv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	ctx := context.Background()
	var configFile string
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	log.BindFlags()
	flag.Parse()

	// Initialize logger.
	log.SetDefaultLogger()
	setupLog := log.FromContext(ctx).WithName("main")
	options := ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e16b32eb.intel.com",
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
	}
	setupLog.Info("Configuration Details")
	setupLog.Info("Configuration file", logkeys.ConfigFile, configFile)

	err := func() error {
		cfg := &privatecloudv1alpha1.VastStorageOperatorConfig{}
		if configFile != "" {
			var err error
			options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(cfg))
			if err != nil {
				return fmt.Errorf("unable to load the config file: %w", err)
			}
		}
		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}
		// Initialize tracing
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		setupLog.Info("Configuration", logkeys.Configuration, cfg)
		// read this from config file
		serverAddr := cfg.StorageControllerServerAddr
		strCntrClient := storagecontroller.StorageControllerClient{}
		err = strCntrClient.Init(context.Background(), serverAddr, cfg.StorageControllerServerUseMtls)
		if err != nil {
			setupLog.Info("unable to initialize storage controller client")
			setupLog.Error(err, "strCntrClient")
		}

		bs, err := os.ReadFile("/vault/secrets/cert.pem")
		block := &pem.Block{}
		if err != nil {
			setupLog.Info("error reading cert.pem")
		} else {
			block, _ = pem.Decode(bs)
			if block == nil {
				setupLog.Info("failed to parse PEM block containing the public key")
			} else {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					setupLog.Info("failed to parse certificate", logkeys.Error, err)
				} else {
					setupLog.Info("Subject: ", logkeys.Subject, cert.Subject)
					setupLog.Info("DNS names:", logkeys.DNSNames, cert.DNSNames)
				}
			}
		}

		setupLog.Info("Storage controller server address", logkeys.ServerAddr, serverAddr)
		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}
		setupLog.Info("Before creating KMS Client", logkeys.ServerAddr, cfg.StorageKmsAddr)
		kmsclientConn, err := grpcutil.NewClient(context.Background(), cfg.StorageKmsAddr, dialOptions...)
		if err != nil {
			setupLog.Error(err, "unable to obtain connection for storage kms", logkeys.ServerAddr, cfg.StorageKmsAddr)
			return fmt.Errorf("storage controller server grpc dial failed")
		}

		storageKmsClient := pb.NewStorageKMSPrivateServiceClient(kmsclientConn)
		// creating kubernetes clientset
		kubeClientSet, err := kubernetes.NewForConfig(mgr.GetConfig())
		if err != nil {
			return fmt.Errorf("error building kubernetes clientset: %w", err)
		}

		// create IDC clientset
		idcClientSet, err := cloudclient.NewForConfig(mgr.GetConfig())
		if err != nil {
			return fmt.Errorf("unable to build idc clientset: %w", err)
		}

		defaultResync := 10 * time.Minute
		// creating SharedInformers
		informerFactory := idcinformerfactory.NewSharedInformerFactory(idcClientSet, defaultResync)

		// Creating sshProxy controller
		ctx := context.Background()
		var vastStorageController *cloud.StorageReconciler
		if vastStorageController, err = cloud.NewStorageOperator(ctx, kubeClientSet, idcClientSet, informerFactory.Private().V1alpha1().VastStorages(), informerFactory, &strCntrClient, storageKmsClient, mgr); err != nil {
			setupLog.Error(err, "unable to set up operator")
			os.Exit(1)
		}

		if err != nil {
			return fmt.Errorf("error creating vastStorageController: %w", err)
		}

		stopCh := make(chan struct{})

		// Intializing all the requested informers
		informerFactory.Start(stopCh)
		// Adding the vast controller to the manager
		err = mgr.Add(manager.RunnableFunc(func(context.Context) error {
			return vastStorageController.Run(ctx, stopCh)
		}))
		if err != nil {
			return fmt.Errorf("unable to add SSH proxy controller to manager: %w", err)
		}

		//+kubebuilder:scaffold:builder

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
		setupLog.Error(err, "unable to setup config", logkeys.Controller, "Storage")
		os.Exit(1)
	}
}
