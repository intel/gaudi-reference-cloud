// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bmc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/stmcginnis/gofish/redfish"
)

type CoyotePFR struct {
	Name         string `json:"Name"`
	Model        string `json:"Model"`
	Manufacturer string `json:"Manufacturer"`
	Oem          struct {
		FirmwareProvisioning struct {
			Locked      string `json:"Locked"`
			Provisioned string `json:"Provisioned"`
		} `json:"FirmwareProvisioning"`
	} `json:"Oem"`
	OdataEtag string `json:"@odata.etag"`
}

type CoyoteOEM struct {
	Name         string `json:"Name"`
	Model        string `json:"Model"`
	Manufacturer string `json:"Manufacturer"`
	Oem          struct {
		IntelRackScale struct {
			OdataType                         string `json:"@odata.type"`
			ProcessorSockets                  int    `json:"ProcessorSockets"`
			MemorySockets                     int    `json:"MemorySockets"`
			UserModeEnabled                   bool   `json:"UserModeEnabled"`
			TrustedExecutionTechnologyEnabled bool   `json:"TrustedExecutionTechnologyEnabled"`
			PciDevices                        []struct {
				VendorID string `json:"VendorId"`
				DeviceID string `json:"DeviceId"`
			} `json:"PciDevices"`
			PCIeConnectionID []any `json:"PCIeConnectionId"`
			Metrics          struct {
				OdataID string `json:"@odata.id"`
			} `json:"Metrics"`
		} `json:"Intel_RackScale"`
		GenSystemDebugLogState struct {
			OdataType string `json:"@odata.type"`
			Status    string `json:"Status"`
		} `json:"GenSystemDebugLogState"`
		OemNcsi struct {
			OdataType string `json:"@odata.type"`
			Ncsi      struct {
				OdataID string `json:"@odata.id"`
			} `json:"Ncsi"`
		} `json:"OemNcsi"`
	} `json:"Oem"`
}

const (
	coyotePassModelRegex = `^M50CYP` // System ID:   120316
	CoyoteRiserCard      = `UEFI IPv4: Network 00 at Riser 01 Slot 02`
	CoyoteBaseboardCard  = `Network Boot`
)

var _ Interface = (*CoyotePassBMC)(nil)

// CoyotePassBMC provides APIs for Coyote Pass server board
type CoyotePassBMC struct {
	BMC
}

func (c *CoyotePassBMC) GetHostMACAddress(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("BMC.getEthernetInterface")
	log.Info("Getting Ethernet Interface")

	system, err := c.getSystem()
	if err != nil {
		return "", err
	}

	ethernetInterfaces, err := system.EthernetInterfaces()
	if err != nil {
		return "", fmt.Errorf("unable to get the ethernet interface: %v", err)
	}

	for _, eth := range ethernetInterfaces {
		if eth.LinkStatus == redfish.LinkUpLinkStatus {
			if strings.Contains(eth.ID, "Nic0") {
				c.BMC.pxeBootRegex = CoyoteBaseboardCard
			} else {
				c.BMC.pxeBootRegex = CoyoteRiserCard
			}
			return eth.MACAddress, nil
		}
	}

	return "", fmt.Errorf("no available ethernet interface")
}

func (c *CoyotePassBMC) GetHostBMCAddress() (string, error) {
	system, err := c.getSystem()
	if err != nil {
		return "", fmt.Errorf("unable to get the computing system: %v", err)
	}
	address := fmt.Sprintf("intel-coyote-redfish+%s%s", c.config.URL, system.ODataID())
	return address, nil
}

func (c *CoyotePassBMC) SanitizeBMCBootOrder(ctx context.Context) error {
	return c.sanitizeBMCBootOrder(ctx, c.BMC.pxeBootRegex)
}

func (c *CoyotePassBMC) VerifyPlatformFirmwareResilience(ctx context.Context) error {
	return c.verifyPlatformFirmwareResilience(ctx)
}

func (c *CoyotePassBMC) GPUDiscovery(ctx context.Context) (count int, gpuModel string, err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.GPUDiscovery")
	log.Info("Starting GPU Discovery")
	systems, err := c.GetClient().GetService().Systems()
	count = 0
	gpuModel = NoGpuType
	if err != nil {
		return count, "", fmt.Errorf("unable to get the computing system: %v", err)
	}

	if len(systems) < 1 {
		return count, "", fmt.Errorf("no system found for BMC under Services")
	}
	// Get
	system := systems[0]
	response, err := c.GetClient().Get(system.ODataID())
	if err != nil {
		return count, "", fmt.Errorf("failed to Run get System %v", err)
	}
	var coyoteOem CoyoteOEM
	defer response.Body.Close()
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return count, "", fmt.Errorf("failed to read BMC System response: %v", err)
	}
	err = json.Unmarshal(resBody, &coyoteOem)
	if err != nil {
		return count, "", fmt.Errorf("failed to unmarshal BMC System: %v", err)
	}
	//check for matching GPUs here
	if len(coyoteOem.Oem.IntelRackScale.PciDevices) > 1 {
		for _, pciDevice := range coyoteOem.Oem.IntelRackScale.PciDevices {
			deviceID := fmt.Sprint(pciDevice.VendorID + ":" + pciDevice.DeviceID)
			if v, found := pcieToGPUTable[deviceID]; found {
				gpuModel = v
				count += 1
			}
		}
	}
	return count, gpuModel, nil
}

func (c *CoyotePassBMC) verifyPlatformFirmwareResilience(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.VerifyPlatformFirmwareResilience")
	log.Info("Requesting Platform Firmware Resilience status")
	systems, err := c.GetClient().GetService().Systems()
	if err != nil {
		return fmt.Errorf("unable to get the computing system: %v", err)
	}

	if len(systems) < 1 {
		return fmt.Errorf("no system found for BMC under Services")
	}
	// Get
	system := systems[0]
	response, err := c.GetClient().Get(system.ODataID())
	if err != nil {
		// return nil
		// This was return nil if we got no response which implied it was okay?
		return fmt.Errorf("failed to get BMC System response: %v", err)
	}
	var coyotePFR CoyotePFR
	defer response.Body.Close()
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read BMC System response: %v", err)
	}
	err = json.Unmarshal(resBody, &coyotePFR)
	if err != nil {
		return fmt.Errorf("failed to unmarshal BMC System: %v", err)
	}
	//check if enabled and lock prior to enrollemnt as locking is manual process
	if reflect.DeepEqual(coyotePFR.Oem.FirmwareProvisioning.Provisioned, "true") &&
		reflect.DeepEqual(coyotePFR.Oem.FirmwareProvisioning.Locked, "true") {
		log.Info("Requesting Platform Firmware Resilience (PFR) status succeeded ")
		return nil
	} else {
		return fmt.Errorf("failed to verify Platform Firmware Resilience (PFR) status, "+
			"Provisioned, current state is %s expected value %s, "+
			"Locked, current state is %s expected value %s, ",
			coyotePFR.Oem.FirmwareProvisioning.Provisioned, "true",
			coyotePFR.Oem.FirmwareProvisioning.Locked, "true")
	}
}
