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

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	objectStore "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/object_store_operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

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
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
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
		cfg := &cloudv1alpha1.ObjectStoreOperatorConfig{}
		if configFile != "" {
			var err error
			options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(cfg))
			if err != nil {
				return fmt.Errorf("unable to load the config file: %w", err)
			}
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
		// Initialize tracing
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)
		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
		if err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}
		setupLog.Info("Configuration", logkeys.Configuration, cfg)
		// read this from config file
		serverAddr := cfg.StorageControllerServerAddr
		strCntrClient := storagecontroller.StorageControllerClient{}
		err = strCntrClient.Init(context.Background(), serverAddr, cfg.StorageControllerServerUseMtls)
		if err != nil {
			setupLog.Info("unable to initialize storage controller client")
			setupLog.Error(err, "strCntrClient")
		}
		setupLog.Info("Storage controller server address", logkeys.ServerAddr, serverAddr)
		ctx := context.Background()
		_, err = objectStore.NewObjectStoreOperator(ctx, mgr, &strCntrClient)
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
