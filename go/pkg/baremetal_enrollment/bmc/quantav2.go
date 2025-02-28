// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package bmc

import (
	"context"
	//"encoding/json"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"net/url"
	//"net/http"
	//helper "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ipmilan"
)

const (
	quantaGridD55Q2U  = `^QuantaGrid D55Q-2U`
	quantaKCSFunc     = 0x30
	quantaKCSCmd      = 0xb4
	quantaKCSCheckCmd = 0xb3
)

var _ Interface = (*QuantaBMC)(nil)

type QuantaV2BMC struct {
	BMC
}

func (c *QuantaV2BMC) GetHostMACAddress(ctx context.Context) (string, error) {
	return "", nil
}

func (c *QuantaV2BMC) GetHostBMCAddress() (string, error) {
	u, err := url.Parse(c.config.URL)
	if err != nil {
		return "", fmt.Errorf("failed to parse hostname from URL (%s): %v", c.config.URL, err)
	}
	address := fmt.Sprintf("ipmi://%s", u.Hostname())
	return address, nil
}

func (c *QuantaV2BMC) EnableKCS(ctx context.Context) (err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.EnableKCS")
	log.Info("Enable KCS")
	ipmiHelper, err := ipmilan.NewIpmiLanHelper(ctx, c.config.URL, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("unable to initialize IPMI helper: %v", err)
	}
	err = ipmiHelper.Connect(ctx)
	if err != nil {
		return fmt.Errorf("unable to Connect to IPMI: %v", err)
	}
	defer ipmiHelper.Close()
	byteData := []byte{byte(ipmilan.QuantaAllow)}
	err = ipmiHelper.RunRawCommand(ctx, byteData, quantaKCSFunc, quantaKCSCmd)
	if err != nil {
		return fmt.Errorf("unable to Enable KCS: %v", err)
	}
	err = ipmiHelper.VerifyKCSMode(ctx, ipmilan.QuantaAllow, quantaKCSFunc, quantaKCSCheckCmd)
	if err != nil {
		return fmt.Errorf("unable to verify Secure KCS is Enabled with IPMI over LAN: %v", err)
	}
	return nil
}

func (c *QuantaV2BMC) DisableKCS(ctx context.Context) (err error) {
	log := log.FromContext(ctx).WithName("BaremetalEnrollment.DisableKCS")
	log.Info("Disable KCS")
	ipmiHelper, err := ipmilan.NewIpmiLanHelper(ctx, c.config.URL, c.config.Username, c.config.Password)
	if err != nil {
		return fmt.Errorf("unable to initialize IPMI helper: %v", err)
	}
	err = ipmiHelper.Connect(ctx)
	if err != nil {
		return fmt.Errorf("unable to Connect to IPMI: %v", err)
	}
	defer ipmiHelper.Close()
	byteData := []byte{byte(ipmilan.QuantaDisable)}
	err = ipmiHelper.RunRawCommand(ctx, byteData, quantaKCSFunc, quantaKCSCmd)
	if err != nil {
		return fmt.Errorf("unable to Disable KCS: %v", err)
	}
	err = ipmiHelper.VerifyKCSMode(ctx, ipmilan.QuantaDisable, quantaKCSFunc, quantaKCSCheckCmd)
	if err != nil {
		return fmt.Errorf("unable to verify Secure KCS is Disabled with IPMI over LAN: %v", err)
	}
	return nil
}

func (c *QuantaV2BMC) PowerOffBMC(ctx context.Context) error {
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

func (c *QuantaV2BMC) PowerOnBMC(ctx context.Context) error {
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
