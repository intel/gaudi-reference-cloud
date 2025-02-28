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

	// "crypto/tls"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	mineralRiver "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/mineral-river"

	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/controllers"
	devicesmanager "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/devices_manager"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/pools"
	statusreporter "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/status_reporter"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(idcnetworkv1alpha1.AddToScheme(scheme))

	utilruntime.Must(baremetalv1alpha1.AddToScheme(scheme))

	utilruntime.Must(corev1.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var configFile string
	var allowedNativeVlanIds []int
	var allowedVlanIds []int
	var provisioningVlanIds []int
	flag.StringVar(&configFile, "config", "tests/config/controller_manager_config.yaml", "The application will load its configuration from this file.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	log.BindFlags()
	flag.Parse()

	ctx := context.Background()
	log.SetDefaultLogger()
	logger := log.FromContext(ctx)

	var err error
	errCh := make(chan error)

	// read configuration from config file
	cfg := &idcnetworkv1alpha1.SDNControllerConfig{}
	options := ctrl.Options{
		Scheme: scheme,
	}
	if configFile == "" {
		logger.Error(fmt.Errorf("unable to read configuration"), "configuration file is not provided")
		os.Exit(1)
	}

	options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(cfg))
	if err != nil {
		logger.Error(err, "unable to load the config file")
		os.Exit(1)
	}
	err = validateConfig(cfg)
	if err != nil {
		logger.Error(err, "unable to validate the config file")
		os.Exit(1)
	}

	logger.Info("configurations", "cfg", cfg)
	if cfg.ControllerConfig.EnableReadOnlyMode {
		logger.Info("!! Note: enableReadOnlyMode is set to true. Switch and Raven client will be enforced to use read-only clients. In read-only mode, a SwitchPort's status.vlanId and status.ravenDBVlanId don't reflect the actual vlan value.")
		// enforce Switch client read-only client.
		cfg.ControllerConfig.SwitchBackendMode = idcnetworkv1alpha1.SwitchBackendModeReadOnly
	}
	logger.Info("", "enableReadOnlyMode", cfg.ControllerConfig.EnableReadOnlyMode)
	logger.Info("", "SwitchBackendMode", cfg.ControllerConfig.SwitchBackendMode)
	logger.Info("", "SwitchImportSource", cfg.ControllerConfig.SwitchImportSource)
	logger.Info("", "SwitchPortImportSource", cfg.ControllerConfig.SwitchPortImportSource)
	logger.Info("", "GroupToPoolMappingSource", cfg.ControllerConfig.NodeGroupToPoolMappingSource)

	logger.V(1).Info("Debug logging enabled")

	// Initialize monitoring
	mr := mineralRiver.New(
		mineralRiver.WithLogLevel("Debug"),
	)
	tracerProvider := mr.InitTracer(context.Background())
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			logger.Error(err, "Error shutting down tracer provider")
		}
	}()

	nwcpClient := utils.NewK8SWatchClientWithScheme(scheme)

	// create the raven client
	var ravenClient raven.RavenPrivateAPI

	// start the Switch Importer
	func() {
		if cfg.ControllerConfig.SwitchImportSource == idcnetworkv1alpha1.SwitchImportSourceNetbox || cfg.ControllerConfig.SwitchPortImportSource == idcnetworkv1alpha1.SwitchPortImportSourceNetbox {
			netboxController, err := controllers.NewNetboxController(*cfg, nwcpClient)
			if err != nil {
				logger.Error(err, "NewNetboxController failed")
				// dont panic if Netbox controller failed to start, the only impact is no new switch will be imported.
				// we should not block SDN controller to work on the existing switches.
				return
			}

			go func() {
				netboxErr := netboxController.Start(ctx)
				logger.Error(netboxErr, "NetboxController failed")
				// errCh <- netboxErr
			}()
		}
		if cfg.ControllerConfig.SwitchImportSource == idcnetworkv1alpha1.SwitchImportSourceRaven {
			ravenClient, err = raven.NewRavenPrivateClient(ctx, *cfg)
			if err != nil {
				logger.Error(err, "unable to create Raven client")
				os.Exit(1)
			}

			go func() {
				// get switches from Raven and create Switch CRs
				err = devicesmanager.StartSwitchImporter(ctx, cfg.ControllerConfig.SwitchImportPeriodInSec, cfg.ControllerConfig.DataCenter, cfg.ControllerConfig.SwitchSecretsPath, ravenClient, nwcpClient)
				if err != nil {
					// if we failed to init switches, it's not necessary to exit the program, log the error.
					logger.Error(err, "Switch Importer failed")
				}
			}()
		}
	}()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		// LeaderElection:         enableLeaderElection,
		LeaderElectionID: "68940a07.intel.com",
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
		NewClient: utils.NewClientWithoutRateLimiter,
	})
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	allowedNativeVlanIds, err = utils.ExpandVlanRanges(cfg.ControllerConfig.AllowedNativeVlanIds)
	if err != nil {
		logger.Error(err, "Error expanding valid VLAN range")
	}

	allowedVlanIds, err = utils.ExpandVlanRanges(cfg.ControllerConfig.AllowedVlanIds)
	if err != nil {
		logger.Error(err, "Error expanding valid VLAN range")
	}

	provisioningVlanIds, err = utils.ExpandVlanRanges(cfg.ControllerConfig.ProvisioningVlanIds)
	if err != nil {
		logger.Error(err, "Error expanding valid provisioningVlans")
	}

	// create the event recorders
	switchPortEventRecorder := mgr.GetEventRecorderFor("switchport-controller")
	switchEventRecorder := mgr.GetEventRecorderFor("switch-controller")
	networkNodeEventRecorder := mgr.GetEventRecorderFor("networknode-controller")
	nodeGroupEventRecorder := mgr.GetEventRecorderFor("nodegroup-controller")
	statusReporterEventRecorder := mgr.GetEventRecorderFor("status-reporter")
	portChannelEventRecorder := mgr.GetEventRecorderFor("portchannel-controller")
	bmhEventRecorder := mgr.GetEventRecorderFor("BMH-controller")
	deviceManagerEventRecorder := mgr.GetEventRecorderFor("device-manager")

	var deviceMgrCfg = devicesmanager.DeviceManagerConfig{
		ControllerConfig: devicesmanager.DeviceManagerControllerConfig{
			SwitchBackendMode:     cfg.ControllerConfig.SwitchBackendMode,
			SwitchSecretsPath:     cfg.ControllerConfig.SwitchSecretsPath,
			AllowedTrunkGroups:    cfg.ControllerConfig.AllowedTrunkGroups,
			Datacenter:            cfg.ControllerConfig.DataCenter,
			DeviceManagerRecorder: deviceManagerEventRecorder,
			AllowedModes:          cfg.ControllerConfig.AllowedModes,
			AllowedVlanIds:        allowedVlanIds,
			AllowedNativeVlanIds:  allowedNativeVlanIds,
			ProvisioningVlanIds:   provisioningVlanIds,
		},
	}
	dam := devicesmanager.NewDeviceAccessManager(nwcpClient, deviceMgrCfg)
	// start watcher for secrets file
	go func() {
		logger.Info("starting watcher for secrets file")
		interval := 10 * time.Second

		err := dam.WatchFileChanges(ctx, cfg.ControllerConfig.SwitchSecretsPath, interval)
		if err != nil {
			logger.Error(err, "Watcher failed")
		}
	}()

	bmhUsageReporter := pools.NewBMHUsageReporter()

	// create the Pool Manager
	pmConf := &pools.PoolResourceManagerConf{
		NwcpK8sClient:     nwcpClient,
		CtrlConf:          cfg,
		NodeUsageReporter: bmhUsageReporter,
		Mgr:               mgr,
	}
	poolMgr, err := pools.NewPoolManager(pmConf)
	if err != nil {
		logger.Error(err, "NewPoolResourceManager failed")
		os.Exit(1)
	}
	// start the Pool Manager.
	go func(errC chan error) {
		logger.Info("starting PoolManager")
		err = poolMgr.Start(ctx)
		errC <- err
	}(errCh)

	// prepare the configurations for the BMH controllers
	if cfg.ControllerConfig.SwitchPortImportSource == idcnetworkv1alpha1.SwitchPortImportSourceBMH {
		bmhConf := controllers.BMHControllerConf{
			Dam:                                dam,
			CtrlConf:                           cfg,
			NwcpK8sClient:                      nwcpClient,
			GroupPoolMappingWatchIntervalInSec: 10,
			PoolManager:                        poolMgr,
			BMHUsageReporter:                   bmhUsageReporter,
			EventRecorder:                      bmhEventRecorder,
			BMHControllerCreationRetryInterval: 120 * time.Second,
		}

		bmhManager := controllers.NewBMHControllerManager(
			bmhConf,
			controllers.NewBMHController,
			utils.GenerateServersKey,
			cfg.ControllerConfig.BMHClusterKubeConfigFilePath,
		)

		// create all the BMH Controllers
		bmhManager.CreateAllControllers(ctx)

		// start all the BMHController
		// BMH Manager will try to start all the BMH Controllers, but if one or even all of them fail, we will NOT shutdown the SDN Controller,
		// as the failure of the BMH Controller will only affect the new provisioned nodes,
		// and shuting down the SDN will block the network configuration requests for all the existing nodes.
		bmhManager.StartAllControllers(ctx)
		// if we want, we can add some logic here to wait for the BMH Controller's startup results and determine if we want to continue or panic.
		// for example, if none of the controllers works, then exit out. But, again, it's not necessary.
		// var runnableControllers map[string]controllers.BMHControllerIF
		// var failedControllers map[string]controllers.BMHControllerIF
		// for {
		// 	time.Sleep(2 * time.Second)
		// 	runnableControllers = bmhManager.GetRunnableControllers()
		// 	failedControllers = bmhManager.GetFailedControllers()
		// 	fmt.Printf("runnableControllers: [%d], failedControllers: [%d] \n", len(runnableControllers), len(failedControllers))
		// 	if len(runnableControllers)+len(failedControllers) == len(metal3KubeconfigFilePaths) {
		// 		break
		// 	}
		// }
		// if len(runnableControllers) == 0 {
		// 	os.Exit(-1)
		// }
	}

	// start the status reporter
	statusReporter := statusreporter.NewStatusReporter(statusreporter.StatusReporterConfig{
		ReportIntervalInSec:            cfg.ControllerConfig.StatusReportPeriodInSec,
		ReportAcceleratedIntervalInSec: cfg.ControllerConfig.StatusReportAcceleratedPeriodInSec,
		DeviceAccessManager:            dam,
		NetworkK8sClient:               nwcpClient,
		Ctx:                            ctx,
		StatusReporterRecorder:         statusReporterEventRecorder,
		BGPCommunityIncomingGroupName:  cfg.ControllerConfig.BGPCommunityIncomingGroupName,
		PortChannelsEnabled:            cfg.ControllerConfig.PortChannelsEnabled,
	})

	if err = (&controllers.SwitchReconciler{
		Client:               mgr.GetClient(),
		Scheme:               mgr.GetScheme(),
		Conf:                 *cfg,
		DevicesAccessManager: dam,
		EventRecorder:        switchEventRecorder,
		StatusReporter:       statusReporter,
	}).SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to create controller", "controller", "Switch")
		os.Exit(1)
	}
	logger.Info("Finished Switch controller setup.")

	// switch port controller
	if err = (&controllers.SwitchPortReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		EventRecorder:  switchPortEventRecorder,
		DevicesManager: dam,
		StatusReporter: statusReporter,
		Conf:           *cfg,
	}).SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to create controller", "controller", "SwitchPort")
		os.Exit(1)
	}
	logger.Info("Finished SwitchPort controller setup.")

	// NetworkNode controller
	if err = (&controllers.NetworkNodeReconciler{
		Client:               mgr.GetClient(),
		Scheme:               mgr.GetScheme(),
		EventRecorder:        networkNodeEventRecorder,
		DevicesManager:       dam,
		Conf:                 *cfg,
		PoolManager:          poolMgr,
		AllowedVlanIds:       allowedVlanIds,
		AllowedNativeVlanIds: allowedNativeVlanIds,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NetworkNode")
		os.Exit(1)
	}
	logger.Info("Finished NetworkNode controller setup.")

	// setup the NodeGroup Controller
	if err = (&controllers.NodeGroupReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		EventRecorder: nodeGroupEventRecorder,
		PoolManager:   poolMgr,
		Conf:          *cfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeGroup")
		os.Exit(1)
	}

	if cfg.ControllerConfig.PortChannelsEnabled {
		if err = (&controllers.PortChannelReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			DevicesManager: dam,
			StatusReporter: statusReporter,
			EventRecorder:  portChannelEventRecorder,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "PortChannel")
			os.Exit(1)
		}
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	go func(errC chan error) {
		// start the controller
		logger.Info("starting controllers")
		err = mgr.Start(ctrl.SetupSignalHandler())
		errC <- err
	}(errCh)

	select {
	case <-ctx.Done():
		logger.Error(ctx.Err(), "context is done")
	case err := <-errCh:
		logger.Error(err, "failed to run SDN-Controller")
	}
	logger.Error(err, "terminating server")
}

func validateConfig(cfg *idcnetworkv1alpha1.SDNControllerConfig) error {
	logger := log.FromContext(context.Background())
	var errMsgs []string
	if len(cfg.ControllerConfig.SwitchSecretsPath) == 0 {
		errMsgs = append(errMsgs, "switchSecretsPath is not provided")
	}
	if cfg.ControllerConfig.MaxConcurrentReconciles < 0 {
		errMsgs = append(errMsgs, "MaxConcurrentReconciles is invalid")
	}
	if len(cfg.ControllerConfig.DataCenter) == 0 {
		errMsgs = append(errMsgs, "DataCenter is not provided")
	}

	if cfg.ControllerConfig.SwitchPortImportSource == idcnetworkv1alpha1.SwitchPortImportSourceBMH && len(cfg.ControllerConfig.BMHClusterKubeConfigFilePath) == 0 {
		errMsgs = append(errMsgs, "SwitchPortImportSourceBMH is configured but BMHClusterKubeConfigFilePath is not provided")
	}

	if len(cfg.ControllerConfig.SwitchBackendMode) == 0 {
		logger.Info("SwitchBackendMode is not provided, defaulting it to mock Switch mode")
		cfg.ControllerConfig.SwitchBackendMode = idcnetworkv1alpha1.SwitchBackendModeMock
	}

	if len(errMsgs) == 0 {
		return nil
	}
	errMsg := "config file validation failed"
	for _, msg := range errMsgs {
		errMsg += ", " + msg
	}
	return fmt.Errorf(errMsg)
}
