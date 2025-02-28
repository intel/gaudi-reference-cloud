// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bmc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ipmilan"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"

	helper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
)

var _ Interface = (*WiwynnBMC)(nil)

// Wiwynn provides APIs for Wiwynn server board
type WiwynnBMC struct {
	BMC
}

type WiWynnApiSession struct {
	Ok             int    `json:"ok"`
	Privilege      int    `json:"privilege"`
	UserID         int    `json:"user_id"`
	Extendedpriv   int    `json:"extendedpriv"`
	RacsessionID   int    `json:"racsession_id"`
	RemoteAddr     string `json:"remote_addr"`
	ServerName     string `json:"server_name"`
	ServerAddr     string `json:"server_addr"`
	HTTPSEnabled   int    `json:"HTTPSEnabled"`
	CSRFToken      string `json:"CSRFToken"`
	Channel        int    `json:"channel"`
	PasswordStatus int    `json:"passwordStatus"`
}

type WiWynnIPMIInterface struct {
	IpmiOverLAN string `json:"ipmi_over_LAN"`
	IpmiOverKcs string `json:"ipmi_over_kcs"`
}

func (c *WiwynnBMC) GetHostBMCAddress() (string, error) {
	u, err := url.Parse(c.config.URL)
	if err != nil {
		return "", fmt.Errorf("failed to parse hostname from URL (%s): %v", c.config.URL, err)
	}
	address := fmt.Sprintf("ipmi://%s", u.Hostname())
	return address, nil
}

func (c *WiwynnBMC) GetHostMACAddress(ctx context.Context) (string, error) {
	return "", nil
}

func (c *WiwynnBMC) GPUDiscovery(ctx context.Context) (count int, gpuModel string, err error) {
	// TODO
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.GPUDiscovery")
	log.Info("Starting GPU Discovery")
	return 8, pcieToGPUTable["0x1da3:0x1020"], nil
}

func (c *WiwynnBMC) EnableKCS(ctx context.Context) (err error) {
	return c.setKcs(ctx, true)
}

func (c *WiwynnBMC) DisableKCS(ctx context.Context) (err error) {
	return c.setKcs(ctx, false)
}

// GetHostCPU returns the current available CPU of the host
func (c *WiwynnBMC) GetHostCPU(ctx context.Context) (*CPUInfo, error) {
	log := log.FromContext(ctx).WithName("BMC.GetHostCPU")
	log.Info("Getting host's CPU information")
	cpuInfo := &CPUInfo{}
	// TODO from IPMI
	cpuInfo.Manufacturer = "Intel_corporation"
	cpuInfo.CPUID = "0x08380"
	cpuInfo.Cores = 80
	cpuInfo.Sockets = 2
	cpuInfo.Threads = 2
	return cpuInfo, nil
}

func (c *WiwynnBMC) UpdateAccount(ctx context.Context, newUserName, newPassword string) error {
	// TODO
	return nil
}

func (c *WiwynnBMC) PowerOffBMC(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.PowerOffBMC")
	log.Info("Forcing Power Off and monitoring BMC ")
	ipmiHelper, err := ipmilan.NewIpmiLanHelper(ctx, c.config.URL, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("unable to initialize IPMI helper: %v", err)
	}
	err = ipmiHelper.Connect(ctx)
	if err != nil {
		return fmt.Errorf("unable to Connect to IPMI: %v", err)
	}
	defer ipmiHelper.Close()
	err = ipmiHelper.PowerOffHost(ctx)
	if err != nil {
		return fmt.Errorf("unable to powerOff: %v", err)
	}
	return nil
}

func (c *WiwynnBMC) PowerOnBMC(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.PowerOnBMC")
	log.Info("Forcing Power on and monitoring BMC ")
	ipmiHelper, err := ipmilan.NewIpmiLanHelper(ctx, c.config.URL, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("unable to initialize IPMI helper: %v", err)
	}
	err = ipmiHelper.Connect(ctx)
	if err != nil {
		return fmt.Errorf("unable to Connect to IPMI: %v", err)
	}
	defer ipmiHelper.Close()
	err = ipmiHelper.PowerOnHost(ctx)
	if err != nil {
		return fmt.Errorf("unable to powerOn: %v", err)
	}
	return nil
}

func (c *WiwynnBMC) setKcs(ctx context.Context, enableKcs bool) error {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.setKcs")
	log.Info("setKcs")
	loginSession := helper.ApiSession{}
	cookies, err := helper.HttpLogin(ctx, &loginSession, c.config.Username, c.config.Password, c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to login %v", err)
	}
	defer helper.HttpLogout(ctx, loginSession.CSRFToken, cookies, c.config.URL)
	ipmiOverKcs := WiWynnIPMIInterface{}
	switch enableKcs {
	case true:
		ipmiOverKcs.IpmiOverKcs = "1"
	default:
		ipmiOverKcs.IpmiOverKcs = "0"
	}
	ipmiOverKcs.IpmiOverLAN = "1"
	// convert json payload to bytes
	httpBody, err := json.Marshal(ipmiOverKcs)
	if err != nil {
		return fmt.Errorf(" %v", err)
	}
	httpResult, _, err := helper.HttpRequest(ctx, "settings/ipmi_disable_interfaces", http.MethodPost, httpBody, loginSession.CSRFToken, cookies, c.config.URL)
	if err != nil {
		return fmt.Errorf("failed to send ipmi_disable_interfaces %v", err)
	}
	// Unmarshal response
	log.Info(string(httpResult))
	resultIpmiOverKcs := WiWynnIPMIInterface{}
	err = json.Unmarshal(httpResult, &resultIpmiOverKcs)
	if err != nil {
		return fmt.Errorf("failed unmarshall resultIpmiOverKcs: %v", err)
	}
	if resultIpmiOverKcs.IpmiOverKcs != ipmiOverKcs.IpmiOverKcs {
		return fmt.Errorf("failed to set KCS, values don't match expected %s vs result %s", ipmiOverKcs.IpmiOverKcs, resultIpmiOverKcs.IpmiOverKcs)
	}
	return nil
}
