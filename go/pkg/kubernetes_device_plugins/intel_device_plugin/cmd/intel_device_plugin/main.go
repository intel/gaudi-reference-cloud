// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// This file is based on Kubevirt v0.54.0 virt-handler (https://github.com/kubevirt/kubevirt/blob/v0.54.0/pkg/virt-handler/device-manager/device_controller.go)

package main

import (
	"context"
	"math"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_device_plugins/intel_device_plugin/pci"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("main")

	vendorSelectorResourceNameMap := make(map[string]string)
	vendorSelectorResourceNameMap["1da3:1020"] = "habana.com/GAUDI2_AI_TRAINING_ACCELERATOR"
	vendorSelectorResourceNameMap["8086:0bda"] = "intel.com/PONTE_VECCHIO_XT_1_TILE_DATA_CENTER_GPU_MAX_1100"

	log.Info("Discovering pci devices and initializing and starting plugin")
	for pciVendorSelector, resourceName := range vendorSelectorResourceNameMap {
		DiscoverInitializeStartPCIDevices(ctx, pciVendorSelector, resourceName)
	}
	<-ctx.Done()
}

func DiscoverInitializeStartPCIDevices(ctx context.Context, pciVendorSelector string, resourceName string) {
	log := log.FromContext(ctx).WithName("DiscoverInitializeStartPCIDevices")

	stop := make(chan struct{})
	defaultBackoffTime := []time.Duration{1 * time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second}
	retries := 0

	// Discover PCI devices
	pciDevices := pci.DiscoverPCIDevices(ctx, pciVendorSelector, resourceName)
	log.Info("Discovered PCI devices on the node", logkeys.ResourceName, resourceName, logkeys.NumOfPCIDevices, len(pciDevices))

	// Initialize PCI plugin
	pciPlugin := pci.NewPCIDevicePlugin(pciDevices, resourceName, pciVendorSelector)

	go func() {
		for {
			err := pciPlugin.Start(ctx, stop)
			if err != nil {
				log.Error(err, "Error starting device plugin", logkeys.ResourceName, resourceName)
				retries = int(math.Min(float64(retries+1), float64(len(defaultBackoffTime)-1)))
			} else {
				retries = 0
			}

			select {
			case <-stop:
				return
			case <-time.After(defaultBackoffTime[retries]):
				// Wait using exponential backoff and attempt to re-register
				continue
			}
		}
	}()
}
