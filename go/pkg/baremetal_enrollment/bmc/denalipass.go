// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bmc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ipmilan"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const (
	denaliPassModelRegex     = `^D50DNP` // System ID:   105461
	DenaliPassBootIntelRegex = `UEFI IPv4: Intel Network 00`
	DenaliPassBootRegex      = `UEFI IPv4: Network 00`
	DenaliRiserCard          = ` at Riser 01 Slot 02`
	DenaliBaseboardCard      = ` at Basebaoard `
	DenaliPFRNotLocked       = "ProvisionedButNotLocked"
	DenaliPFRAndLocked       = "ProvisionedAndLocked"
	DenaliKCSFunc            = 0x30
	DenaliKCSCmd             = 0xb4
	DenaliKCSCheckCmd        = 0xb3
)

var _ Interface = (*DenaliPassBMC)(nil)

// DenaliPassBMC provides APIs for Denali Pass server board
type DenaliPassBMC struct {
	BMC
}

type HostInterfacePayload struct {
	Oem Oem
}

type Oem struct {
	HostInterface HostInterface
}

type HostInterface struct {
	Enabled bool
}

func (c *DenaliPassBMC) GetHostBMCAddress() (string, error) {
	system, err := c.getSystem()
	if err != nil {
		return "", fmt.Errorf("unable to get the computing system: %v", err)
	}
	address := fmt.Sprintf("intel-denali-redfish+%s%s", c.config.URL, system.ODataID())
	return address, nil
}

func (c *DenaliPassBMC) VerifyPlatformFirmwareResilience(ctx context.Context) error {
	return c.verifyPlatformFirmwareResilience(ctx)
}

func (c *DenaliPassBMC) GetHostMACAddress(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("DenaliPassBMC.GetSystemMACAddress")
	log.Info("Getting network interfaces")

	system, err := c.getSystem()
	if err != nil {
		return "", err
	}

	networkInterfaces, err := system.NetworkInterfaces()
	if err != nil {
		return "", fmt.Errorf("unable to get the network interface: %v", err)
	}
	if len(networkInterfaces) == 0 {
		return "", fmt.Errorf("no network interface found for BMC under system")
	}

	for _, nic := range networkInterfaces {
		macAddress, err := c.getAvailablePortFromNIC(nic)
		if err != nil {
			continue
		}
		return macAddress, nil
	}

	return "", fmt.Errorf("no available ethernet interface")
}

func (c *DenaliPassBMC) SanitizeBMCBootOrder(ctx context.Context) error {
	return c.sanitizeBMCBootOrder(ctx, c.BMC.pxeBootRegex)
}

func (c *DenaliPassBMC) GPUDiscovery(ctx context.Context) (count int, gpuModel string, err error) {
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
	pcieDevices, err := systems[0].PCIeDevices()
	if err != nil {
		return 0, NoGpuType, fmt.Errorf("failed to read pcie devices %v", err)
	}
	for _, pcieDevice := range pcieDevices {
		//read function 0
		pcieFunctionLink := fmt.Sprint(pcieDevice.ODataID + "/PCIeFunctions/0")
		response, err := c.GetClient().Get(pcieFunctionLink)
		if err != nil {
			return count, "", fmt.Errorf("failed to Run get pcieFunction %v", err)
		}
		var pcieFunction OpenBMCPcieFunction
		defer response.Body.Close()
		resBody, err := io.ReadAll(response.Body)
		if err != nil {
			return count, "", fmt.Errorf("failed to read open BMC System response: %v", err)
		}
		err = json.Unmarshal(resBody, &pcieFunction)
		if err != nil {
			return count, "", fmt.Errorf("failed to unmarshal open BMC pcieFunction: %v", err)
		}
		deviceID := fmt.Sprint(pcieFunction.VendorID + ":" + pcieFunction.DeviceID)
		if v, found := pcieToGPUTable[deviceID]; found {
			gpuModel = v
			count += 1
		}
	}
	return count, gpuModel, nil
}

func (c *DenaliPassBMC) ConfigureNTP(ctx context.Context) (err error) {
	return c.openBMCConfigureNTP(ctx, "/redfish/v1/Managers/bmc/NetworkProtocol")
}

func (c *DenaliPassBMC) HBMDiscovery(ctx context.Context) (hbmMode string, err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.HMBDiscovery")
	log.Info("Starting GPU Discovery")
	systems, err := c.GetClient().GetService().Systems()
	if err != nil {
		return string(HBMNone), fmt.Errorf("unable to get the computing system: %v", err)
	}

	if len(systems) < 1 {
		return string(HBMNone), fmt.Errorf("no system found for BMC under Services")
	}
	dimm0, err := systems[0].Memory()
	if err != nil {
		return string(HBMNone), fmt.Errorf("error reading Dimms list %v", err)
	}
	ddr5Count := 0
	hbmCount := 0
	for _, dimm := range dimm0 {
		if MemoryDeviceType(dimm.MemoryDeviceType) == DDR5 {
			ddr5Count += 1
		}
		if MemoryDeviceType(dimm.MemoryDeviceType) == HBM {
			hbmCount += 1
		}

	}
	// results
	// We support Flat mode and hbm only mode for now
	// in case we have a combination of Memory Types => Flat
	// in case of All hbm => hbm only
	if hbmCount > 0 && ddr5Count > 0 {
		return string(HBMFlat), nil
	} else if hbmCount > 0 && ddr5Count == 0 {
		return string(HBMOnly), nil
	} else {
		return string(HBMNone), nil
	}
}

func (c *DenaliPassBMC) EnableKCS(ctx context.Context) (err error) {
	return c.ipmiEnableKCS(ctx)
}

func (c *DenaliPassBMC) verifyPlatformFirmwareResilience(ctx context.Context) error {
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
	var openBMCSystem OpenBMCSystem
	defer response.Body.Close()
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read open BMC System response: %v", err)
	}
	err = json.Unmarshal(resBody, &openBMCSystem)
	if err != nil {
		return fmt.Errorf("failed to unmarshal open BMC System: %v", err)
	}
	//check if enabled and lock prior to enrollemnt as locking is manual process
	if reflect.DeepEqual(openBMCSystem.Oem.OpenBmc.FirmwareProvisioning.ProvisioningStatus, DenaliPFRAndLocked) {
		log.Info("Requesting Platform Firmware Resilience (PFR) status succeeded ")
		return nil
	} else {
		return fmt.Errorf("failed to verify Platform Firmware Resilience (PFR) status, "+
			"current state is %s expected value %s",
			openBMCSystem.Oem.OpenBmc.FirmwareProvisioning.ProvisioningStatus, DenaliPFRAndLocked)
	}
}

func (c *DenaliPassBMC) ipmiEnableKCS(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("EnrollmentTask.ipmiEnableKCS")
	log.Info("Enabling KCS using IPMI")

	ipmiHelper, err := ipmilan.NewIpmiLanHelper(ctx, c.config.URL, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("unable to initialize IPMI helper: %v", err)
	}

	err = ipmiHelper.Connect(ctx)
	if err != nil {
		return fmt.Errorf("unable to Connect to IPMI: %v", err)
	}

	defer ipmiHelper.Close()
	// Check if we are already KCS Mode Enabled
	err = ipmiHelper.VerifyKCSMode(ctx, ipmilan.AllowAll, DenaliKCSFunc, DenaliKCSCheckCmd)
	if err != nil {
		byteData := []byte{byte(ipmilan.AllowAll)}
		err = ipmiHelper.RunRawCommand(ctx, byteData, DenaliKCSFunc, DenaliKCSCmd)
		if err != nil {
			return fmt.Errorf("failed to Enable KCS with IPMI over LAN: %v", err)
		}

		err = ipmiHelper.VerifyKCSMode(ctx, ipmilan.AllowAll, DenaliKCSFunc, DenaliKCSCheckCmd)
		if err != nil {
			return fmt.Errorf("unable to verify Secure KCS is Enabled with IPMI over LAN: %v", err)
		}
	}

	log.Info("KCS Policy Control Mode verified as Unsecured, Provisioning")

	return nil
}

func (c *DenaliPassBMC) DisableKCS(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BmInstance.ipmiSecureKCS")

	log.Info("Securing KCS using IPMI")

	ipmiHelper, err := ipmilan.NewIpmiLanHelper(ctx, c.config.URL, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("unable to initialize IPMI helper: %v", err)
	}

	err = ipmiHelper.Connect(ctx)
	if err != nil {
		return fmt.Errorf("unable to Connect to IPMI: %v", err)
	}

	defer ipmiHelper.Close()
	// Check if we are already set to the right KCS Mode
	err = ipmiHelper.VerifyKCSMode(ctx, ipmilan.DenyAll, DenaliKCSFunc, DenaliKCSCheckCmd)
	if err != nil {
		byteData := []byte{byte(ipmilan.DenyAll)}
		err = ipmiHelper.RunRawCommand(ctx, byteData, DenaliKCSFunc, DenaliKCSCmd)
		if err != nil {
			return fmt.Errorf("failed to Secure KCS with IPMI over LAN: %v", err)
		}

		err = ipmiHelper.VerifyKCSMode(ctx, ipmilan.DenyAll, DenaliKCSFunc, DenaliKCSCheckCmd)
		if err != nil {
			return fmt.Errorf("unable to verify Secure KCS is Disabled with IPMI over LAN: %v", err)
		}
	}

	log.Info("KCS Policy Control Mode verified as Provisioned")

	return nil
}

func (c *DenaliPassBMC) getAvailablePortFromNIC(nic *redfish.NetworkInterface) (string, error) {
	ports, err := nic.NetworkPorts()
	if err != nil {
		return "", fmt.Errorf("unable to get the network interface's port: %v", err)
	}
	if len(ports) == 0 {
		return "", fmt.Errorf("no port found under network interface %s", nic.Name)
	}

	// Denali ports reports only NIC Channel 0 even if there is 2 NIC ports on the same card
	for _, port := range ports {
		if port.Status.State == common.EnabledState &&
			port.Status.Health == common.OKHealth {
			response, err := c.GetClient().Get(port.ODataID)
			if err != nil {
				continue
			}
			var networkPortRaw NetworkPortRaw
			defer response.Body.Close()
			resBody, err := io.ReadAll(response.Body)
			err = json.Unmarshal(resBody, &networkPortRaw)
			if err != nil {
				continue
			}
			if networkPortRaw.Oem.OpenBmc.MediaState > 0 {
				fmt.Printf(networkPortRaw.VendorID)
				if strings.Contains(nic.ID, "Nic0") {
					c.BMC.pxeBootRegex = DenaliPassBootIntelRegex + DenaliBaseboardCard
				} else {
					if PCIEVendor(networkPortRaw.VendorID) == Mellanox {
						c.BMC.pxeBootRegex = DenaliPassBootRegex + DenaliRiserCard
					} else {
						c.BMC.pxeBootRegex = DenaliPassBootIntelRegex + DenaliRiserCard
					}
				}
				if len(networkPortRaw.AssociatedNetworkAddresses) > 0 {
					return networkPortRaw.AssociatedNetworkAddresses[0], nil
				}
			}
		}
	}
	return "", fmt.Errorf("no available network port")
}

func (c *DenaliPassBMC) EnableHCI(ctx context.Context) (err error) {
	return c.setHCI(ctx, true)
}

func (c *DenaliPassBMC) DisableHCI(ctx context.Context) (err error) {
	return c.setHCI(ctx, false)
}

func (c *DenaliPassBMC) setHCI(ctx context.Context, enableHCI bool) (err error) {
	log := log.FromContext(ctx).WithName("DenaliPassBMC.setHCI")
	log.Info("Set Host Interface")

	serversPayload := HostInterfacePayload{
		Oem: Oem{
			HostInterface: HostInterface{
				Enabled: enableHCI,
			},
		},
	}

	resp, err := c.GetClient().Patch("/redfish/v1/Managers/bmc", serversPayload)
	if err != nil {
		return fmt.Errorf("failed to setup HCI to %s %v", strconv.FormatBool(enableHCI), err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to setup HCI with status code %d", resp.StatusCode)
	}
	return nil
}
