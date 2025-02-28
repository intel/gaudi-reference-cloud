// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubevirt v0.54.0 virt-handler (https://github.com/kubevirt/kubevirt/blob/v0.54.0/pkg/virt-handler/device-manager/pci_device.go)
package pci

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	pluginapi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_device_plugins/deviceplugin/v1beta1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"google.golang.org/grpc"
)

const (
	vfioDevicePath      = "/dev/vfio/"
	vfioMount           = "/dev/vfio/vfio"
	pciBasePath         = "/sys/bus/pci/devices"
	PCI_RESOURCE_PREFIX = "PCI_RESOURCE"
	connectionTimeout   = 5 * time.Second
	HostRootMount       = "/proc/1/root/"
)

type deviceHealth struct {
	DevId  string
	Health string
}

type PCIDevice struct {
	pciID      string
	driver     string
	pciAddress string
	iommuGroup string
	numaNode   int
}

type PCIDevicePlugin struct {
	devs              []*pluginapi.Device
	server            *grpc.Server
	socketPath        string
	stop              <-chan struct{}
	health            chan deviceHealth
	devicePath        string
	resourceName      string
	done              chan struct{}
	deviceRoot        string
	iommuToPCIMap     map[string]string
	initialized       bool
	lock              *sync.Mutex
	deregistered      chan struct{}
	pciAddresses      []string
	pciVendorSelector string
}

func NewPCIDevicePlugin(pciDevices []*PCIDevice, resourceName string, pciVendorSelector string) *PCIDevicePlugin {
	serverSock := SocketPath(strings.Replace(resourceName, "/", "-", -1))
	iommuToPCIMap := make(map[string]string)
	pciAddresses := make([]string, 0)
	for _, pciDevice := range pciDevices {
		pciAddresses = append(pciAddresses, pciDevice.pciAddress)
	}
	devs := ConstructDevices(pciDevices, iommuToPCIMap)
	dpi := &PCIDevicePlugin{
		devs:              devs,
		socketPath:        serverSock,
		resourceName:      resourceName,
		devicePath:        vfioDevicePath,
		deviceRoot:        HostRootMount,
		iommuToPCIMap:     iommuToPCIMap,
		health:            make(chan deviceHealth),
		initialized:       false,
		lock:              &sync.Mutex{},
		pciAddresses:      pciAddresses,
		pciVendorSelector: pciVendorSelector,
	}
	return dpi
}

// Start starts the device plugin
func (dpi *PCIDevicePlugin) Start(ctx context.Context, stop <-chan struct{}) (err error) {
	log := log.FromContext(ctx).WithName("PCIDevicePlugin.Start")
	dpi.stop = stop
	dpi.done = make(chan struct{})
	dpi.deregistered = make(chan struct{})

	log.Info("Start Device Plugin", logkeys.ResourceName, dpi.resourceName)
	err = dpi.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", dpi.socketPath)
	if err != nil {
		log.Error(err, "error creating GRPC server socket", logkeys.ResourceName, dpi.resourceName)
		return fmt.Errorf("error creating GRPC server socket for resource %s: %v", dpi.resourceName, err)

	}

	dpi.server = grpc.NewServer([]grpc.ServerOption{}...)
	defer dpi.stopDevicePlugin()

	pluginapi.RegisterDevicePluginServer(dpi.server, dpi)

	errChan := make(chan error, 2)

	go func() {
		errChan <- dpi.server.Serve(sock)
	}()

	err = waitForGRPCServer(ctx, dpi.socketPath, connectionTimeout)
	if err != nil {
		return fmt.Errorf("error starting the GRPC server for resource %s: %v", dpi.resourceName, err)
	}

	err = dpi.register(ctx)
	if err != nil {
		return fmt.Errorf("error encountered while registering the resource %s with device plugin manager: %v", dpi.resourceName, err)
	}

	go func() {
		errChan <- dpi.healthCheck(ctx)

	}()

	dpi.setInitialized(true)
	log.Info("device plugin started", logkeys.ResourceName, dpi.resourceName)
	err = <-errChan

	return err
}

func (dpi *PCIDevicePlugin) ListAndWatch(_ *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	ctx := context.TODO()
	log := log.FromContext(ctx).WithName("PCIDevicePlugin.ListAndWatch")
	emptyList := []*pluginapi.Device{}
	s.Send(&pluginapi.ListAndWatchResponse{Devices: emptyList})

	// Acquire lock before accessing devs
	dpi.lock.Lock()
	s.Send(&pluginapi.ListAndWatchResponse{Devices: dpi.devs})
	dpi.lock.Unlock()

	done := false
	for {
		select {
		case devHealth := <-dpi.health:
			// Acquire lock before accessing devs
			dpi.lock.Lock()
			for _, dev := range dpi.devs {
				if devHealth.DevId == dev.ID {
					dev.Health = devHealth.Health
					log.Info("Device health updated", logkeys.DeviceId, dev.ID, logkeys.DeviceHealth, dev.Health)
				}
			}
			if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: dpi.devs}); err != nil {
				log.Error(err, "Failed to send updated device list", logkeys.ResourceName, dpi.resourceName)
			}
			dpi.lock.Unlock()
		case <-dpi.stop:
			done = true
		case <-dpi.done:
			done = true
		}
		if done {
			break
		}
	}
	// Send empty list to increase the chance that the kubelet acts fast on stopped device plugins
	// There exists no explicit way to deregister devices
	if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: emptyList}); err != nil {
		log.Error(err, "device plugin failed to deregister", logkeys.ResourceName, dpi.resourceName)
	}
	close(dpi.deregistered)
	return nil
}

func (dpi *PCIDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	log := log.FromContext(ctx).WithName("PCIDevicePlugin.Allocate")
	log.Info("PCIDevicePlugin Allocate request", logkeys.Request, r.ContainerRequests)
	resourceNameEnvVar := ResourceNameToEnvVar(PCI_RESOURCE_PREFIX, dpi.resourceName)
	allocatedDevices := []string{}
	resp := new(pluginapi.AllocateResponse)
	containerResponse := new(pluginapi.ContainerAllocateResponse)

	for _, request := range r.ContainerRequests {
		deviceSpecs := make([]*pluginapi.DeviceSpec, 0)
		for _, devID := range request.DevicesIDs {
			// translate device's iommu group to its pci address
			devPCIAddress, exist := dpi.iommuToPCIMap[devID]
			if !exist {
				continue
			}
			allocatedDevices = append(allocatedDevices, devPCIAddress)
			deviceSpecs = append(deviceSpecs, FormatVFIODeviceSpecs(devID)...)
		}
		containerResponse.Devices = deviceSpecs
		envVar := make(map[string]string)
		envVar[resourceNameEnvVar] = strings.Join(allocatedDevices, ",")

		containerResponse.Envs = envVar
		resp.ContainerResponses = append(resp.ContainerResponses, containerResponse)
	}
	return resp, nil
}

func (dpi *PCIDevicePlugin) healthCheck(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("PCIDevicePlugin.healthCheck")
	monitoredDevices := make(map[string]string)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to creating a fsnotify watcher for resource %s: %v", dpi.resourceName, err)
	}
	defer watcher.Close()

	// This way we don't have to mount /dev from the node
	devicePath := filepath.Join(dpi.deviceRoot, dpi.devicePath)

	// Start watching the files before we check for their existence to avoid races
	dirName := filepath.Dir(devicePath)
	err = watcher.Add(dirName)
	if err != nil {
		return fmt.Errorf("failed to add the device root path to the watcher for resource %s: %v", dpi.resourceName, err)
	}

	_, err = os.Stat(devicePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("could not stat the device for resource %s: %v", dpi.resourceName, err)
		}
	}

	// probe all devices
	// Acquire lock before accessing devs
	dpi.lock.Lock()
	for _, dev := range dpi.devs {
		vfioDevice := filepath.Join(devicePath, dev.ID)
		err = watcher.Add(vfioDevice)
		if err != nil {
			dpi.lock.Unlock()
			return fmt.Errorf("failed to add the device %s to the watcher for resource %s: %v", vfioDevice, dpi.resourceName, err)
		}
		monitoredDevices[vfioDevice] = dev.ID
	}
	dpi.lock.Unlock()

	dirName = filepath.Dir(dpi.socketPath)
	err = watcher.Add(dirName)

	if err != nil {
		return fmt.Errorf("failed to add the device-plugin kubelet path to the watcher for resource %s: %v", dpi.resourceName, err)
	}
	_, err = os.Stat(dpi.socketPath)
	if err != nil {
		return fmt.Errorf("failed to stat the device-plugin socket for resource %s: %v", dpi.resourceName, err)
	}

	for {
		select {
		case <-dpi.stop:
			return nil
		case err := <-watcher.Errors:
			log.Error(err, "error watching devices and device plugin directory", logkeys.ResourceName, dpi.resourceName)
		case event := <-watcher.Events:
			log.V(4).Info("health Event", logkeys.Event, event)
			if monDevId, exist := monitoredDevices[event.Name]; exist {
				// Health in this case is if the device path actually exists
				if event.Op == fsnotify.Create {
					log.Info("monitored device appeared", logkeys.ResourceName, dpi.resourceName)
					dpi.health <- deviceHealth{
						DevId:  monDevId,
						Health: pluginapi.Healthy,
					}
				} else if (event.Op == fsnotify.Remove) || (event.Op == fsnotify.Rename) {
					log.Info("monitored device disappeared", logkeys.ResourceName, dpi.resourceName)
					dpi.health <- deviceHealth{
						DevId:  monDevId,
						Health: pluginapi.Unhealthy,
					}
				}
			} else if event.Name == dpi.socketPath && event.Op == fsnotify.Remove {
				log.Info("device socket file was removed, kubelet probably restarted", logkeys.ResourceName, dpi.resourceName)

				// stop the device plugin
				err := dpi.stopDevicePlugin()
				if err != nil {
					log.Error(err, "failed to stop device plugin", logkeys.ResourceName, dpi.resourceName)
				}

				// Periodically scan for devices until rediscovered
				defaultBackoffTime := []time.Duration{1 * time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second}
				newDevices := DiscoverPCIDevicesWithRetry(ctx, dpi.pciVendorSelector, dpi.resourceName, defaultBackoffTime)
				log.Info("Discovered PCI devices on the node after node restart", logkeys.ResourceName, dpi.resourceName, logkeys.NumOfPCIDevices, len(newDevices))

				// Acquire lock before updating devs
				dpi.lock.Lock()
				dpi.devs = ConstructDevices(newDevices, dpi.iommuToPCIMap)
				dpi.lock.Unlock()

				err = dpi.registerWithRetry(ctx)
				if err != nil {
					log.Error(err, "failed to re-register device plugin", logkeys.ResourceName, dpi.resourceName)
				}
				return nil
			}
		}
	}
}

func DiscoverPCIDevicesWithRetry(ctx context.Context, pciVendorId string, resourceName string, backoffTimes []time.Duration) []*PCIDevice {
	log := log.FromContext(ctx).WithName("PCIDevicePlugin.DiscoverPCIDevicesWithRetry")
	var pciDevices []*PCIDevice
	retries := 0

	for {
		pciDevices = DiscoverPCIDevices(ctx, pciVendorId, resourceName)
		if len(pciDevices) > 0 {
			return pciDevices
		}

		retries = int(math.Min(float64(retries+1), float64(len(backoffTimes)-1)))
		log.Info("No PCI devices found", logkeys.ResourceName, resourceName, logkeys.RetryInterval, backoffTimes[retries])

		select {
		case <-time.After(backoffTimes[retries]):
			continue
		case <-ctx.Done():
			log.Info("Context cancelled while discovering PCI devices", logkeys.ResourceName, resourceName)
			return nil
		}
	}
}

// Try to register the device plugin with kubelet every retryInterval until succeeded
func (dpi *PCIDevicePlugin) registerWithRetry(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("PCIDevicePlugin.registerWithRetry")
	retryInterval := 2 * time.Second
	for {
		err := dpi.register(ctx)
		if err == nil {
			return nil
		}
		log.Error(err, "Failed to register device plugin with kubelet", logkeys.ResourceName, dpi.resourceName, logkeys.RetryInterval, retryInterval)
		time.Sleep(retryInterval)
	}
}

// Stop stops the gRPC server
func (dpi *PCIDevicePlugin) stopDevicePlugin() error {
	defer func() {
		if !IsChanClosed(dpi.done) {
			close(dpi.done)
		}
	}()

	// Give the device plugin one second to properly deregister
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	select {
	case <-dpi.deregistered:
	case <-ticker.C:
	}

	dpi.server.Stop()
	dpi.setInitialized(false)
	return dpi.cleanup()
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (dpi *PCIDevicePlugin) register(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("PCIDevicePlugin.register")
	log.Info("Registering the device plugin", logkeys.ResourceName, dpi.resourceName)
	conn, err := gRPCConnect(ctx, pluginapi.KubeletSocket, connectionTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	empty := &pluginapi.Empty{}
	client := pluginapi.NewRegistrationClient(conn)
	devicePluginOptions, err := dpi.GetDevicePluginOptions(ctx, empty)
	if err != nil {
		return err
	}
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(dpi.socketPath),
		ResourceName: dpi.resourceName,
		Options:      devicePluginOptions,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

func (dpi *PCIDevicePlugin) cleanup() error {
	if err := os.Remove(dpi.socketPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (dpi *PCIDevicePlugin) GetDevicePluginOptions(_ context.Context, _ *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	options := &pluginapi.DevicePluginOptions{
		PreStartRequired: true,
	}
	return options, nil
}

func (dpi *PCIDevicePlugin) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	log := log.FromContext(ctx).WithName("PCIDevicePlugin.PreStartContainer")
	res := &pluginapi.PreStartContainerResponse{}

	testModeStr := os.Getenv("TEST_MODE")
	testMode := strings.ToLower(testModeStr) == "true"

	if testMode {
		log.Info("Running device plugin in Test Mode. Performing Function Level Reset on all the PCI devices.")

		for _, pciAddress := range dpi.pciAddresses {
			err := dpi.FunctionLevelReset(pciAddress)
			if err != nil {
				log.Error(err, "Test Mode: FunctionLevelReset Error")
				return nil, err
			}
			log.Info("Test Mode: successfully reset PCI device", logkeys.PCIAddress, pciAddress)
		}

		return res, nil
	}

	for _, deviceID := range req.DevicesIDs {
		pciAddress, exist := dpi.iommuToPCIMap[deviceID]
		if !exist {
			continue
		}
		// Perform Function Level Reset (FLR) on the PCI device
		err := dpi.FunctionLevelReset(pciAddress)
		if err != nil {
			log.Error(err, "FunctionLevelReset Error")
			return nil, err
		}
		log.Info("successfully reset PCI device", logkeys.PCIAddress, pciAddress)
	}

	return res, nil
}

func (dpi *PCIDevicePlugin) FunctionLevelReset(pciAddress string) error {
	file, err := os.OpenFile(fmt.Sprintf("/proc/1/root/sys/bus/pci/devices/%s/reset", pciAddress), os.O_WRONLY, 0200)
	if err != nil {
		return fmt.Errorf("failed to open reset file for PCI device %s: %v", pciAddress, err)
	}

	_, err = file.WriteString("1")
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to write 1 to the reset file of PCI device %s: %v", pciAddress, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close the reset file for PCI device %s: %v", pciAddress, err)
	}

	return nil
}

func DiscoverPCIDevices(ctx context.Context, pciVendorId string, resourceName string) []*PCIDevice {
	log := log.FromContext(ctx).WithName("DiscoverPCIDevices")
	pciDevices := make([]*PCIDevice, 0)
	err := filepath.Walk(pciBasePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		pciID, err := GetDevicePCIID(pciBasePath, info.Name())
		if err != nil {
			log.Error(err, "failed get vendor:device ID for device", logkeys.DeviceName, info.Name())
			return nil
		}
		if pciVendorId == pciID {
			// check device driver
			driver, err := GetDeviceDriver(ctx, pciBasePath, info.Name())
			if err != nil || driver != "vfio-pci" {
				return nil
			}

			pcidev := &PCIDevice{
				pciID:      pciID,
				pciAddress: info.Name(),
			}
			iommuGroup, err := GetDeviceIOMMUGroup(ctx, pciBasePath, info.Name())
			if err != nil {
				return nil
			}
			pcidev.iommuGroup = iommuGroup
			pcidev.driver = driver
			pcidev.numaNode = GetDeviceNumaNode(ctx, pciBasePath, info.Name())
			pciDevices = append(pciDevices, pcidev)
		}
		return nil
	})
	if err != nil {
		log.Error(err, "failed to discover host devices", logkeys.ResourceName, resourceName)
	}
	return pciDevices
}

func (dpi *PCIDevicePlugin) setInitialized(initialized bool) {
	dpi.lock.Lock()
	dpi.initialized = initialized
	dpi.lock.Unlock()
}
