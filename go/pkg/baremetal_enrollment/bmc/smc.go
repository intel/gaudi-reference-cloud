// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bmc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type GaudiPCIeDevices struct {
	OdataType string `json:"@odata.type"`
	OdataID   string `json:"@odata.id"`
	Name      string `json:"Name"`
	Members   []struct {
		OdataID string `json:"@odata.id"`
	} `json:"Members"`
	MembersOdataCount int    `json:"Members@odata.count"`
	OdataEtag         string `json:"@odata.etag"`
}

type GaudiPcieFunctionsMembers struct {
	OdataType string `json:"@odata.type"`
	OdataID   string `json:"@odata.id"`
	Name      string `json:"Name"`
	Members   []struct {
		OdataID string `json:"@odata.id"`
	} `json:"Members"`
	MembersOdataCount int    `json:"Members@odata.count"`
	OdataEtag         string `json:"@odata.etag"`
}

const (
	smcGaudi2PassModelRegex = `^SYS-820GH`
	smcSys521GeTNRT         = `^SYS-521GE-TNRT`
	smcSys821GVTNRT         = `SYS-821GV-TNR`
	smcSys621CTN12R         = `SYS-621C-TN12R`
	smcSys822GANGR3IN001    = `SYS-822GA-NGR3-IN001`
	FanFullSpeed            = "FullSpeed"
)

var _ Interface = (*SmcBMC)(nil)

type SmcBMC struct {
	BMC
}

func (c *SmcBMC) GetHostMACAddress(ctx context.Context) (string, error) {
	return "", nil
}

func (c *SmcBMC) GPUDiscovery(ctx context.Context) (count int, gpuModel string, err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.GPUDiscovery")
	log.Info("Starting GPU Discovery")
	if c.hwType == Smc521GeTNRT {
		//TODO once bug is fixed in the Discovery, remove this if statement
		return 8, pcieToGPUTable["0x8086:0x0bda"], nil
	}
	if c.hwType == Smc821GVTNRT {
		//TODO
		return 8, "GPU-Max-1550", nil
	}
	count = 0
	gpuModel = NoGpuType
	// Get PCIEdevices
	response, err := c.GetClient().Get("/redfish/v1/Chassis/1/PCIeDevices")
	if err != nil {
		return count, "", fmt.Errorf("failed to Run get PCIeDevices %v", err)
	}
	var pcieDevices GaudiPCIeDevices
	defer response.Body.Close()
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, NoGpuType, fmt.Errorf("failed to read pcie devices %v", err)
	}
	err = json.Unmarshal(resBody, &pcieDevices)
	if err != nil {
		return count, "", fmt.Errorf("failed to unmarshal open BMC pcieDevices: %v", err)
	}
	// Iterate through all the pcieDevices
	for _, pcieDevice := range pcieDevices.Members {
		//find all the PCIE functions for given device
		pcieFunctionLinks := fmt.Sprint(pcieDevice.OdataID + "/PCIeFunctions")
		response, err := c.GetClient().Get(pcieFunctionLinks)
		if err != nil {
			return count, "", fmt.Errorf("failed to Run get pcieFunction %v", err)
		}
		var pcieFunctions GaudiPcieFunctionsMembers
		defer response.Body.Close()
		resBody, err := io.ReadAll(response.Body)
		if err != nil {
			return count, "", fmt.Errorf("failed to read pcieFunctions response: %v", err)
		}
		err = json.Unmarshal(resBody, &pcieFunctions)
		if err != nil {
			return count, "", fmt.Errorf("failed to unmarshal PcieFunctionsMembers: %v", err)
		}
		// Check for any PCIE functions
		if len(pcieFunctions.Members) < 1 {
			continue
		}
		// Only check for GPUs
		if !strings.Contains(pcieFunctions.Members[0].OdataID, "GPU") {
			continue
		}
		response, err = c.GetClient().Get(pcieFunctions.Members[0].OdataID)
		if err != nil {
			return count, "", fmt.Errorf("failed to Run get pcieFunction %v", err)
		}
		var pcieFunction OpenBMCPcieFunction
		defer response.Body.Close()
		resBody, err = io.ReadAll(response.Body)
		if err != nil {
			return count, "", fmt.Errorf("failed to read open BMC System response: %v", err)
		}
		err = json.Unmarshal(resBody, &pcieFunction)
		if err != nil {
			return count, "", fmt.Errorf("failed to unmarshal pcieFunction: %v", err)
		}
		deviceID := strings.ToLower(fmt.Sprint(pcieFunction.VendorID + ":" + pcieFunction.DeviceID))
		if v, found := pcieToGPUTable[deviceID]; found {
			gpuModel = v
			count += 1
		}
	}
	return count, gpuModel, nil
}

func (c *SmcBMC) EnableKCS(ctx context.Context) (err error) {
	return c.setKcs(ctx, true)
}

func (c *SmcBMC) DisableKCS(ctx context.Context) (err error) {
	return c.setKcs(ctx, false)
}

func (c *SmcBMC) setKcs(ctx context.Context, enableKcs bool) (err error) {
	log := log.FromContext(ctx).WithName("SmcBMC.setKcs")
	log.Info("kcs management")
	var kcsFunction string
	switch enableKcs {
	case true:
		kcsFunction = "Administrator"
	default:
		kcsFunction = "Callback"
	}
	serversPayload := map[string]string{
		"Privilege": kcsFunction,
	}

	resp, err := c.GetClient().Patch("/redfish/v1/Managers/1/Oem/Supermicro/KCSInterface", serversPayload)
	if err != nil {
		return fmt.Errorf("failed to setup KCS to %s %v", kcsFunction, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to setup KCS with status code %d", resp.StatusCode)
	}
	return nil
}

func (c *SmcBMC) EnableHCI(ctx context.Context) (err error) {
	return c.setHCI(ctx, true)
}

func (c *SmcBMC) DisableHCI(ctx context.Context) (err error) {
	return c.setHCI(ctx, false)
}

func (c *SmcBMC) setHCI(ctx context.Context, enableHCI bool) (err error) {
	if c.hwType == Smc822GANGR3IN001 {
		return ErrHCINotSupported
	}
	log := log.FromContext(ctx).WithName("SmcBMC.setHCI")
	log.Info("set host interface")
	serversPayload := map[string]bool{
		"InterfaceEnabled": enableHCI,
	}

	resp, err := c.GetClient().Patch("/redfish/v1/Managers/1/HostInterfaces/1", serversPayload)
	if err != nil {
		return fmt.Errorf("failed to setup HCI to %s %v", strconv.FormatBool(enableHCI), err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to setup HCI with status code %d", resp.StatusCode)
	}
	return nil
}

func (c *SmcBMC) SetFanSpeed(ctx context.Context) (err error) {
	log := log.FromContext(ctx).WithName("BMC.SetFanSpeed")

	if c.hwType != Smc822GANGR3IN001 {
		log.Info("Setting fan speed is not supported", "platform", c.name)
		return nil
	}
	// Check if fan mode is already set to full speed
	fanMode, err := c.getFanSpeed()
	if err != nil {
		return fmt.Errorf("failed to get the SMC fan mode with error: %v", err)
	}
	if fanMode == FanFullSpeed {
		log.Info("fan speed is already set to the expected mode", "platform", c.name, "mode", FanFullSpeed)
		return nil
	}
	serversPayload := map[string]string{
		"Mode": FanFullSpeed,
	}

	resp, err := c.GetClient().Patch("/redfish/v1/Managers/1/Oem/Supermicro/FanMode", serversPayload)
	if err != nil {
		return fmt.Errorf("failed to set SMC fan mode to %s %v", FanFullSpeed, err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to patch SMC fan mode. response status code %d", resp.StatusCode)
	}
	// validate if the fan mode change is applied
	time.Sleep(1 * time.Second)

	fanMode, err = c.getFanSpeed()
	if err != nil {
		return fmt.Errorf("failed to get the SMC fan mode after patching with error: %v", err)
	}
	if fanMode == FanFullSpeed {
		log.Info("fan speed is updated to the expected mode", "platform", c.name, "mode", fanMode)
		return nil
	}
	return nil
}

func (c *SmcBMC) getFanSpeed() (string, error) {
	resp, err := c.GetClient().Get("/redfish/v1/Managers/1/Oem/Supermicro/FanMode")
	if err != nil {
		return "", fmt.Errorf("failed to get SMC fan mode with err %v", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("invalid status code %d when getting SMC fan mode. expected status code is 200", resp.StatusCode)
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read SMC fan mode response: %v", err)
	}
	var FanModeResponse map[string]interface{}
	err = json.Unmarshal(resBody, &FanModeResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling SMC fan mode response: %v", err)
	}
	if fanMode, ok := FanModeResponse["Mode"].(string); ok {
		return fanMode, nil
	} else {
		return "", fmt.Errorf("'Mode' key not found in SMC fan mode response")
	}
}
