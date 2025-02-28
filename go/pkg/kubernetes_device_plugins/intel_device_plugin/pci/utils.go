// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubevirt v0.54.0 virt-handler (https://github.com/kubevirt/kubevirt/blob/v0.54.0/pkg/virt-handler/device-manager/common.go)

package pci

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pluginapi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_device_plugins/deviceplugin/v1beta1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
)

func GetDeviceIOMMUGroup(ctx context.Context, basepath string, pciAddress string) (string, error) {
	log := log.FromContext(ctx).WithName("GetDeviceIOMMUGroup")
	iommuLink := filepath.Join(basepath, pciAddress, "iommu_group")
	iommuPath, err := os.Readlink(iommuLink)
	if err != nil {
		log.Error(err, "failed to read iommu_group link", logkeys.PCIAddress, pciAddress, logkeys.IOMMULink, iommuLink)
		return "", err
	}
	_, iommuGroup := filepath.Split(iommuPath)
	return iommuGroup, nil
}

// gets device driver
func GetDeviceDriver(ctx context.Context, basepath string, pciAddress string) (string, error) {
	log := log.FromContext(ctx).WithName("GetDeviceDriver")
	driverLink := filepath.Join(basepath, pciAddress, "driver")
	driverPath, err := os.Readlink(driverLink)
	if err != nil {
		log.Error(err, "failed to read driver link", logkeys.PCIAddress, pciAddress, logkeys.DriverLink, driverLink)
		return "", err
	}
	_, driver := filepath.Split(driverPath)
	return driver, nil
}

func GetDeviceNumaNode(ctx context.Context, basepath string, pciAddress string) (numaNode int) {
	log := log.FromContext(ctx).WithName("GetDeviceNumaNode")
	numaNode = -1
	numaNodePath := filepath.Join(basepath, pciAddress, "numa_node")
	// #nosec No risk for path injection. Reading static path of NUMA node info
	numaNodeStr, err := os.ReadFile(numaNodePath)
	if err != nil {
		log.Error(err, "failed to read numa node", logkeys.PCIAddress, pciAddress, logkeys.NumaNodePath, numaNodePath)
		return
	}
	numaNodeStr = bytes.TrimSpace(numaNodeStr)
	numaNode, err = strconv.Atoi(string(numaNodeStr))
	if err != nil {
		log.Error(err, "failed to convert numa node value", logkeys.PCIAddress, pciAddress, logkeys.NumaNodeString, numaNodeStr)
		return
	}
	return numaNode
}

func GetDevicePCIID(basepath string, pciAddress string) (string, error) {
	// #nosec No risk for path injection. Reading static path of PCI data
	file, err := os.Open(filepath.Join(basepath, pciAddress, "uevent"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PCI_ID") {
			equal := strings.Index(line, "=")
			value := strings.TrimSpace(line[equal+1:])
			return strings.ToLower(value), nil
		}
	}
	return "", fmt.Errorf("no pci_id is found")
}

func ConstructDevices(pciDevices []*PCIDevice, iommuToPCIMap map[string]string) (devs []*pluginapi.Device) {
	for _, pciDevice := range pciDevices {
		iommuToPCIMap[pciDevice.iommuGroup] = pciDevice.pciAddress
		dpiDev := &pluginapi.Device{
			ID:     string(pciDevice.iommuGroup),
			Health: pluginapi.Healthy,
		}
		devs = append(devs, dpiDev)
	}
	return devs
}

func SocketPath(deviceName string) string {
	return filepath.Join(pluginapi.DevicePluginPath, fmt.Sprintf("kubevirt-%s.sock", deviceName))
}

func IsChanClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func FormatVFIODeviceSpecs(devID string) []*pluginapi.DeviceSpec {
	// always add /dev/vfio/vfio device as well
	devSpecs := make([]*pluginapi.DeviceSpec, 0)
	devSpecs = append(devSpecs, &pluginapi.DeviceSpec{
		HostPath:      vfioMount,
		ContainerPath: vfioMount,
		Permissions:   "mrw",
	})

	vfioDevice := filepath.Join(vfioDevicePath, devID)
	devSpecs = append(devSpecs, &pluginapi.DeviceSpec{
		HostPath:      vfioDevice,
		ContainerPath: vfioDevice,
		Permissions:   "mrw",
	})
	return devSpecs
}

func ResourceNameToEnvVar(prefix string, resourceName string) string {
	varName := strings.ToUpper(resourceName)
	varName = strings.Replace(varName, "/", "_", -1)
	varName = strings.Replace(varName, ".", "_", -1)
	return fmt.Sprintf("%s_%s", prefix, varName)
}

func waitForGRPCServer(ctx context.Context, socketPath string, timeout time.Duration) error {
	conn, err := gRPCConnect(ctx, socketPath, timeout)
	if err != nil {
		return err
	}
	if err := conn.Close(); err != nil {
		return err
	}
	return nil
}

// dial establishes the gRPC communication with the registered device plugin.
func gRPCConnect(ctx context.Context, socketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c, err := grpc.DialContext(ctx, socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			dialer := &net.Dialer{}
			return dialer.DialContext(ctx, "unix", addr)
		}),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}
