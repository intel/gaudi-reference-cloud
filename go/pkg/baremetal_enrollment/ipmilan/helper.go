// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Copyright © 2023 Intel Corporation
package ipmilan

import (
	"context"
	"fmt"
	"net/url"

	"sigs.k8s.io/controller-runtime/pkg/log"

	ipmi "github.com/bougou/go-ipmi"
)

const (
	defaultIpmiPort   = 623
	defaultIpmiCipher = ipmi.CipherSuiteID17
)

type ipmiLanHelper struct {
	hostname string
	port     int
	username string
	password string
	cipher   ipmi.CipherSuiteID
	client   *ipmi.Client
}

func NewIpmiLanHelper(ctx context.Context, bmcUrl, bmcUsername, bmcPassword string) (*ipmiLanHelper, error) {
	log := log.FromContext(ctx).WithName("IpmiLan.NewIpmiLanHelper")
	log.Info("Creating IPMI LAN Configuration")

	u, err := url.Parse(bmcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hostname from URL (%s): %v", bmcUrl, err)
	}

	helper := &ipmiLanHelper{
		hostname: u.Hostname(),
		port:     defaultIpmiPort,
		username: bmcUsername,
		password: bmcPassword,
		cipher:   defaultIpmiCipher,
	}

	client, err := ipmi.NewClient(helper.hostname, helper.port, helper.username, helper.password)
	// Support local mode client if runs directly on linux
	// client, err := ipmi.NewOpenClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create IPMI client: %v", err)
	}

	helper.client = client

	// Network interface
	client.WithInterface(ipmi.InterfaceLanplus)
	// you can optionally open debug switch
	//client.WithDebug(true)
	// Use Cipher Suite 17
	client.WithCipherSuiteID(helper.cipher)

	return helper, nil
}

func (h *ipmiLanHelper) Connect(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("IpmiLan.Connect")
	log.Info("Connecting to IPMI over LAN")

	// Connect will create an authenticated session for you.
	if err := h.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to IPMI client: %v", err)
	}
	return nil
}

func (h *ipmiLanHelper) Close() error {
	if err := h.client.Close(); err != nil {
		return fmt.Errorf("failed to close to IPMI client: %v", err)
	}
	return nil
}

func (h *ipmiLanHelper) GetManufacturer(ctx context.Context) (string, error) {
	log := log.FromContext(ctx).WithName("IpmiLan.GetManufacturer")
	log.Info("Get Manufacturer From IPMI over LAN")

	fru, err := h.client.GetFRU(0, "")
	if err != nil {
		return "", fmt.Errorf("failed to Get Manufacturer using IPMI client: %v", err)
	}
	return string(fru.ProductInfoArea.Manufacturer), nil
}

func (h *ipmiLanHelper) PowerOffHost(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("IpmiLan.PowerOffHost")
	log.Info("Power Off Host")
	_, err := h.client.ChassisControl(ipmi.ChassisControlPowerDown)
	if err != nil {
		return fmt.Errorf("failed to Power Off: %v", err)
	}
	return nil
}

func (h *ipmiLanHelper) PowerOnHost(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("IpmiLan.PowerOnHost")
	log.Info("Power On Host")
	_, err := h.client.ChassisControl(ipmi.ChassisControlPowerUp)
	if err != nil {
		return fmt.Errorf("failed to Power On: %v", err)
	}
	return nil
}

func (h *ipmiLanHelper) RunRawCommand(ctx context.Context, byteData []byte, netFn ipmi.NetFn, cmd uint8) error {
	log := log.FromContext(ctx).WithName("IpmiLan.RawCommand")
	log.Info("Run Raw command")
	rawResults, err := h.client.RawCommand(netFn, cmd, byteData, "Intel General Application: Run Raw command")
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Running Raw command Results: '%s'", rawResults.Format()))

	return nil
}
