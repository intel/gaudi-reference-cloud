// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package devicesmanager

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	idcnetworkv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/metrics"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/switch-clients"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-controller/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

type SetOption struct {
}

type GetOption struct {
	SwitchFQDN string
	IP         string
}

type DevicesAccessManager interface {
	GetSwitchClient(getOption GetOption) (sc.SwitchClient, error)
	GetOrCreateSwitchClient(ctx context.Context, option GetOption) (sc.SwitchClient, error)
	AddOrUpdateSwitch(sw *idcnetworkv1alpha1.Switch, forceRecreateSwitchClientChanged bool) error
	DeleteSwitch(switchFQDN string) error
	InsecurelyDisableFQDNValidation()
	WatchFileChanges(ctx context.Context, filePath string, interval time.Duration) error
}

type DeviceManagerControllerConfig struct { // A cut down version of idcnetworkv1alpha1.ControllerConfig.
	SwitchBackendMode     string
	SwitchSecretsPath     string
	AllowedTrunkGroups    []string
	Datacenter            string
	AllowedVlanIds        []int
	AllowedNativeVlanIds  []int
	AllowedModes          []string
	ProvisioningVlanIds   []int
	DeviceManagerRecorder record.EventRecorder
}

type DeviceManagerConfig struct {
	ControllerConfig DeviceManagerControllerConfig
}

// IDCNetworkDeviceManager
type IDCNetworkDeviceManager struct {
	sync.Mutex
	cfg                   DeviceManagerConfig
	switchClientsMap      map[string]sc.SwitchClient
	k8sClient             k8sclient.Client
	disableFQDNValidation bool
	deviceManagerRecorder record.EventRecorder
}

func NewDeviceAccessManager(k8sClient k8sclient.Client, cfg DeviceManagerConfig) DevicesAccessManager {
	return &IDCNetworkDeviceManager{
		cfg:                   cfg,
		switchClientsMap:      make(map[string]sc.SwitchClient),
		k8sClient:             k8sClient,
		disableFQDNValidation: false,
		deviceManagerRecorder: cfg.ControllerConfig.DeviceManagerRecorder,
	}
}

// InsecurelyDisableFQDNValidation Useful for testing. Allows GetOrCreateSwitchClient to be passed ANY FQDN & IP.
// In production, these should be validated against Netbox, not connected to blindly.
func (d *IDCNetworkDeviceManager) InsecurelyDisableFQDNValidation() {
	d.disableFQDNValidation = true
}

func (d *IDCNetworkDeviceManager) GetSwitchClient(getOption GetOption) (sc.SwitchClient, error) {
	d.Lock()
	client, found := d.switchClientsMap[getOption.SwitchFQDN]
	d.Unlock()
	if !found {
		return nil, fmt.Errorf("switch client for [%v] not found", getOption.SwitchFQDN)
	}
	return client, nil
}

func (d *IDCNetworkDeviceManager) GetOrCreateSwitchClient(ctx context.Context, getOption GetOption) (sc.SwitchClient, error) {
	d.Lock()
	client, found := d.switchClientsMap[getOption.SwitchFQDN]
	d.Unlock()
	if found {
		return client, nil
	}

	// See if the Switch CR exists in k8s. If it does, add it to the devices_manager.
	if d.k8sClient != nil {
		existingSwitch := &idcnetworkv1alpha1.Switch{}
		key := types.NamespacedName{Name: getOption.SwitchFQDN, Namespace: idcnetworkv1alpha1.SDNControllerNamespace}
		err := d.k8sClient.Get(ctx, key, existingSwitch)
		if err != nil {
			return nil, err
		}
		err = d.AddOrUpdateSwitch(existingSwitch, false)
		if err != nil {
			return nil, err
		}
	} else if d.disableFQDNValidation {
		var ip string
		if getOption.IP != "" {
			ip = getOption.IP
		} else {
			ip = getOption.SwitchFQDN
		}
		switchDetails := &idcnetworkv1alpha1.Switch{
			Spec: idcnetworkv1alpha1.SwitchSpec{
				FQDN: getOption.SwitchFQDN,
				Ip:   ip,
			},
		}

		err := d.AddOrUpdateSwitch(switchDetails, false)
		if err != nil {
			return nil, err
		}
	}
	d.Lock()
	client, found = d.switchClientsMap[getOption.SwitchFQDN]
	d.Unlock()
	if found {
		return client, nil
	} else {
		return nil, fmt.Errorf("could not find switchClient even after adding it to the deviceManager.switchClientsMap")
	}
}

func (d *IDCNetworkDeviceManager) AddOrUpdateSwitch(sw *idcnetworkv1alpha1.Switch, forceRecreateSwitchClient bool) error {
	if sw == nil {
		return fmt.Errorf("switch CR is nil")
	}

	ipToUse, err := utils.GetIp(sw, d.cfg.ControllerConfig.Datacenter)
	if err != nil {
		d.Lock()
		delete(d.switchClientsMap, sw.Spec.FQDN)
		d.Unlock()
		if d.deviceManagerRecorder != nil {
			d.deviceManagerRecorder.Event(sw, corev1.EventTypeWarning, "error while getting ip for switch", err.Error())
		}
		return fmt.Errorf("error while getting ip for switch: %v", err)
	}

	logger := log.FromContext(context.Background()).WithName("IDCNetworkDeviceManager.AddOrUpdateSwitch").
		WithValues(utils.LogFieldSwitchIpToUse, ipToUse).
		WithValues(utils.LogFieldSwitchBackendMode, d.cfg.ControllerConfig.SwitchBackendMode)

	d.Lock()
	client, found := d.switchClientsMap[sw.Spec.FQDN]
	d.Unlock()

	switchBackendMode := d.cfg.ControllerConfig.SwitchBackendMode

	if !found || client == nil {
		logger.V(1).Info("adding/updating switch client to DeviceManager")
		client, err = d.createSwitchClient(ipToUse, switchBackendMode)
		if err != nil {
			if d.deviceManagerRecorder != nil {
				// set the error flag metric for failure to create switch client.
				metrics.DeviceManagerErrors.With(prometheus.Labels{
					metrics.MetricsLabelErrorType:  metrics.ErrorTypeFailedToCreateSwitchClient,
					metrics.MetricsLabelSwitchFQDN: ipToUse,
				}).Set(1)

				d.deviceManagerRecorder.Event(sw, corev1.EventTypeWarning, "create switch client failed", err.Error())
			}
			return fmt.Errorf("create switch client failed, %v", err)
		} else {
			// reset the value to 0 when the issue is gone or when it's normal
			metrics.DeviceManagerErrors.With(prometheus.Labels{
				metrics.MetricsLabelErrorType:  metrics.ErrorTypeFailedToCreateSwitchClient,
				metrics.MetricsLabelSwitchFQDN: ipToUse,
			}).Set(0)
		}
	} else {
		existingHost, err := client.GetHost()
		if err != nil {
			return fmt.Errorf("could not get client host, %v", err)
		}
		if (existingHost != ipToUse) || forceRecreateSwitchClient {
			d.Lock()
			delete(d.switchClientsMap, sw.Spec.FQDN)
			d.Unlock()
			logger.V(1).Info("IpToUse changed or forceRecreateSwitchClient, creating switch client to DeviceManager")
			client, err = d.createSwitchClient(ipToUse, switchBackendMode)
			if err != nil {
				if d.deviceManagerRecorder != nil {
					// set the error flag metric for failure to create switch client.
					metrics.DeviceManagerErrors.With(prometheus.Labels{
						metrics.MetricsLabelErrorType:  metrics.ErrorTypeFailedToCreateSwitchClient,
						metrics.MetricsLabelSwitchFQDN: ipToUse,
					}).Set(1)

					d.deviceManagerRecorder.Event(sw, corev1.EventTypeWarning, "create switch client failed", err.Error())
				}
				return fmt.Errorf("create switch client failed, %v", err)
			} else {
				// reset the value to 0 when the issue is gone or when it's normal
				metrics.DeviceManagerErrors.With(prometheus.Labels{
					metrics.MetricsLabelErrorType:  metrics.ErrorTypeFailedToCreateSwitchClient,
					metrics.MetricsLabelSwitchFQDN: ipToUse,
				}).Set(0)
			}
		}

	}

	if client == nil {
		return fmt.Errorf("create switch client failed")
	}

	d.Lock()
	// coverity[ATOMICITY:FALSE]
	d.switchClientsMap[sw.Spec.FQDN] = client
	d.Unlock()
	logger.V(1).Info("successfully added/updated switch client to DeviceManager")

	return nil
}

func (d *IDCNetworkDeviceManager) createSwitchClient(ipToUse string, switchBackendMode string) (sc.SwitchClient, error) {
	//func (d *IDCNetworkDeviceManager) setBackendMode(sw *idcnetworkv1alpha1.Switch, client sc.SwitchClient, ipToUse string) (sc.SwitchClient, error) {
	// here we determine what type of switchBackend will be used to talk to a network switch.
	// The options are:
	// SwitchBackendModeMock:       This will talk to the mock switch backend, it's safe to use this for local development.
	// SwitchBackendModeEAPI:       This mode will call the Arista network switches' eAPI to make configuration changes. this should be used in production.
	// SwitchBackendModeReadOnly: Read only is a special mode that used only for the phase that transitioning from Raven to SDN. It's basically the "SwitchBackendModeEAPI" mode
	//                                                              but without any actual "write" operation.
	// Note: when the flag "ControllerConfig.EnableReadOnlyMode" is set to enable, the network switch backend will be enforce to use the SwitchBackendModeReadOnly. For example, even
	//	logger := log.FromContext(context.Background()).WithName("IDCNetworkDeviceManager.AddOrUpdateSwitch").
	//		WithValues(utils.LogFieldSwitchFQDN, sw.Spec.FQDN).
	//		WithValues(utils.LogFieldSwitchBackendMode, d.cfg.ControllerConfig.SwitchBackendMode)

	var client sc.SwitchClient
	var err error
	eapiConnTimeout := 30 * time.Second

	if switchBackendMode == idcnetworkv1alpha1.SwitchBackendModeMock {
		client, err = sc.NewMockSwitchClient(d.switchClientsMap, ipToUse, d.cfg.ControllerConfig.AllowedVlanIds, d.cfg.ControllerConfig.AllowedNativeVlanIds)
		if err != nil {
			return nil, err
		}
	} else if switchBackendMode == idcnetworkv1alpha1.SwitchBackendModeEAPI {
		client, err = sc.NewAristaClient(ipToUse, d.cfg.ControllerConfig.SwitchSecretsPath, 443, "https", eapiConnTimeout, false, d.cfg.ControllerConfig.AllowedVlanIds, d.cfg.ControllerConfig.AllowedNativeVlanIds, d.cfg.ControllerConfig.AllowedModes, d.cfg.ControllerConfig.AllowedTrunkGroups, d.cfg.ControllerConfig.ProvisioningVlanIds)
		if err != nil {
			return nil, err
		}
	} else if switchBackendMode == idcnetworkv1alpha1.SwitchBackendModeReadOnly {
		client, err = sc.NewAristaClient(ipToUse, d.cfg.ControllerConfig.SwitchSecretsPath, 443, "https", eapiConnTimeout, true, d.cfg.ControllerConfig.AllowedVlanIds, d.cfg.ControllerConfig.AllowedNativeVlanIds, d.cfg.ControllerConfig.AllowedModes, d.cfg.ControllerConfig.AllowedTrunkGroups, d.cfg.ControllerConfig.ProvisioningVlanIds)
		if err != nil {
			return nil, err
		}
	} else {
		err = fmt.Errorf("unrecognized SwitchBackendMode [%v]", d.cfg.ControllerConfig.SwitchBackendMode)
		return nil, err
	}
	return client, nil
}

func (d *IDCNetworkDeviceManager) WatchFileChanges(ctx context.Context, filePath string, interval time.Duration) error {
	var lastModTime time.Time
	var lastContents []byte
	var lastError error

	logger := log.FromContext(ctx).WithName("DeviceManager.WatchFileChanges")
	logger.Info("Starting Watcher")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping Watcher due to context cancellation")
			return ctx.Err()
		case <-ticker.C:
			// Get the file info
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				if lastError == nil || lastError.Error() != err.Error() {
					lastError = err
					logger.Error(err, "Error stating file", "filePath", filePath)
				}
				continue
			}

			// Reset lastError if no error occurred
			lastError = nil

			// Check if the modification time has changed
			if fileInfo.ModTime().After(lastModTime) {
				// Read the file contents
				contents, err := ioutil.ReadFile(filePath)
				if err != nil {
					logger.Error(err, "Error reading file", "filePath", filePath)
					continue
				}

				// Check if the contents have changed
				if !bytes.Equal(contents, lastContents) {
					logger.Info(fmt.Sprintf("File modified : %v", filePath))

					// Getting list of the switches
					allSwitchCRs := &idcnetworkv1alpha1.SwitchList{}
					err = d.k8sClient.List(ctx, allSwitchCRs)
					if err != nil {
						logger.Error(err, "Error while getting networkK8sClient.List")
						continue
					}

					// Iterating through all switches to create client again if change observed in the secrets file
					for _, swCR := range allSwitchCRs.Items {
						err := d.AddOrUpdateSwitch(&swCR, true)
						if err != nil {
							logger.Error(err, "Error while creating switch client")
						}
					}
					lastModTime = fileInfo.ModTime()
					lastContents = contents
				}
			}
		}
	}
}

func (d *IDCNetworkDeviceManager) DeleteSwitch(switchFQDN string) error {
	d.Lock()
	delete(d.switchClientsMap, switchFQDN)
	d.Unlock()
	return nil
}
